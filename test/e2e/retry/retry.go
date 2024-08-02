package retry

import (
	"fmt"
	"time"
)

var (
	defaultMaxRetries   int64         = 10
	defaultInitialDelay time.Duration = 250 * time.Millisecond
	defaultMaxDelay     time.Duration = time.Duration(defaultMaxRetries) * defaultInitialDelay
)

type Retrier interface {
	Retry(func() error) error
}

type LinearBackoffRetry struct {
	MaxRetries   int64
	InitialDelay time.Duration
	MaxDelay     time.Duration
}

var _ Retrier = (*LinearBackoffRetry)(nil)

func NewLinearBackoffRetry() *LinearBackoffRetry {
	return &LinearBackoffRetry{
		MaxRetries:   defaultMaxRetries,
		InitialDelay: defaultInitialDelay,
		MaxDelay:     defaultMaxDelay,
	}
}

func (r *LinearBackoffRetry) Retry(fn func() error) error {
	retries := int64(0)

	for {
		err := fn()
		if err == nil {
			return nil
		}

		retries++
		if retries > r.MaxRetries {
			return fmt.Errorf("max retries exceeded, last error: %w", err)
		}

		delay := r.InitialDelay * time.Duration(retries)
		if delay > r.MaxDelay {
			delay = r.MaxDelay
		}
		time.Sleep(delay)
	}
}
