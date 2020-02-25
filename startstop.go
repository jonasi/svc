package svc

import (
	"context"
)

// WrapStartStop wraps a start function that returns immediately
// and returns a Service
func WrapStartStop(start LifecycleFn, stop LifecycleFn) Service {
	var (
		ch      = make(chan struct{})
		stopErr error
	)

	nstart := func(ctx context.Context) error {
		if start != nil {
			if err := start(ctx); err != nil {
				return err
			}
		}

		<-ch
		return stopErr
	}

	nstop := func(ctx context.Context) error {
		if stop != nil {
			stopErr = stop(ctx)
		}

		close(ch)
		return stopErr
	}

	return WrapBlocking(nstart, nstop)
}
