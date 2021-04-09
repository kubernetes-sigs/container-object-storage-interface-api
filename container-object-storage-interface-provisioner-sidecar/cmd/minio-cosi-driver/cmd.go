// Copyright 2021 The Kubernetes Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"flag"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"sigs.k8s.io/container-object-storage-interface-provisioner-sidecar/cmd/minio-cosi-driver/internal"
	"sigs.k8s.io/container-object-storage-interface-provisioner-sidecar/pkg/provisioner"

	"k8s.io/klog/v2"
)

const provisionerName = "minio.objectstorage.k8s.io"

var (
	driverAddress = "unix:///var/lib/cosi/cosi.sock"

	minioAccessKey = ""
	minioSecretKey = ""
	minioHost      = ""
)

var cmd = &cobra.Command{
	Use:           "minio-cosi-driver",
	Short:         "K8s COSI driver for MinIO object storage",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return run(cmd.Context(), args)
	},
	DisableFlagsInUseLine: true,
}

func init() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	flag.Set("alsologtostderr", "true")
	kflags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(kflags)

	persistentFlags := cmd.PersistentFlags()
	persistentFlags.AddGoFlagSet(kflags)

	stringFlag := persistentFlags.StringVarP

	stringFlag(&driverAddress,
		"driver-addr",
		"d",
		driverAddress,
		"path to unix domain socket where driver should listen")

	stringFlag(&minioHost,
		"minio-host",
		"m",
		minioHost,
		"endpoint where minio server is listening")

	stringFlag(&minioAccessKey,
		"minio-access-key",
		"a",
		minioAccessKey,
		"access key for minio")

	stringFlag(&minioSecretKey,
		"minio-secret-key",
		"s",
		minioSecretKey,
		"secret key for minio")

	viper.BindPFlags(cmd.PersistentFlags())
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			cmd.PersistentFlags().Set(f.Name, viper.GetString(f.Name))
		}
	})
}

func run(ctx context.Context, args []string) error {
	identityServer, bucketProvisioner, err := internal.NewDriver(ctx,
		provisionerName,
		minioHost,
		minioAccessKey,
		minioSecretKey)
	if err != nil {
		return err
	}

	server, err := provisioner.NewDefaultCOSIProvisionerServer(driverAddress,
		identityServer,
		bucketProvisioner)
	if err != nil {
		return err
	}
	return server.Run(ctx)
}
