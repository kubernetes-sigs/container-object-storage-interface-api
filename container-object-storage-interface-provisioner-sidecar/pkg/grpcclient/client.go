/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package grpcclient

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc"

	"k8s.io/klog"
)

type options struct {
	reconnect func() bool
}

// Option is the type of all optional parameters for Connect.
type Option func(o *options)

type GRPCClient struct {
	serverAddress string
	dialOptions   []grpc.DialOption
}

// NewGRPCClient creates a new GRPCClient
func NewGRPCClient(address string, dialOptions []grpc.DialOption, connectOptions []Option) (*GRPCClient, error) {
	var o options
	for _, option := range connectOptions {
		option(&o)
	}

	dialOptions = append(dialOptions,
		grpc.WithInsecure(),                   // Don't use TLS, it's usually local Unix domain socket in a container.
		grpc.WithBackoffMaxDelay(time.Second), // Retry every second after failure.
		grpc.WithBlock(),                      // Block until connection succeeds.
	)

	unixPrefix := "unix://"
	if strings.HasPrefix(address, "tcp://") {
		address = address[6:]
	}
	if strings.HasPrefix(address, "/") {
		// It looks like filesystem path.
		address = unixPrefix + address
	}

	if strings.HasPrefix(address, unixPrefix) {
		// state variables for the custom dialer
		haveConnected := false
		lostConnection := false
		reconnect := true

		dialOptions = append(dialOptions, grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			if haveConnected && !lostConnection {
				// We have detected a loss of connection for the first time. Decide what to do...
				// Record this once. TODO (?): log at regular time intervals.
				klog.Errorf("Lost connection to %s.", address)
				// Inform caller and let it decide? Default is to reconnect.
				if o.reconnect != nil {
					reconnect = o.reconnect()
				}
				lostConnection = true
			}
			if !reconnect {
				return nil, errors.New("connection lost, reconnecting disabled")
			}
			conn, err := net.DialTimeout("unix", address[len(unixPrefix):], timeout)
			if err == nil {
				// Connection restablished.
				haveConnected = true
				lostConnection = false
			}
			return conn, err
		}))
	} else if o.reconnect != nil {
		return nil, errors.New("OnConnectionLoss callback only supported for unix:// addresses")
	}

	return &GRPCClient{serverAddress: address, dialOptions: dialOptions}, nil
}

// Connect connects to the grpc server
func (c *GRPCClient) ConnectWithLogging(interval time.Duration) (*grpc.ClientConn, error) {
	klog.Infof("Connecting to %s", c.serverAddress)

	grpcLogFunc := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		klog.V(5).Infof("GRPC call: %s", method)
		klog.V(5).Infof("GRPC request: %s", req)
		err := invoker(ctx, method, req, reply, cc, opts...)
		klog.V(5).Infof("GRPC response: %s", reply)
		klog.V(5).Infof("GRPC error: %v", err)
		return err
	}

	// Log all messages
	c.dialOptions = append(c.dialOptions, grpc.WithChainUnaryInterceptor(grpcLogFunc))

	// Connect in background.
	var conn *grpc.ClientConn
	var err error
	ready := make(chan bool)
	go func() {
		conn, err = grpc.Dial(c.serverAddress, c.dialOptions...)
		close(ready)
	}()

	// Log error every connectionLoggingInterval
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Wait until Dial() succeeds.
	for {
		select {
		case <-ticker.C:
			klog.Warningf("Still connecting to %s", c.serverAddress)

		case <-ready:
			return conn, err
		}
	}
}
