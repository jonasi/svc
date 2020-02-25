package svc

import (
	"context"

	"go.uber.org/multierr"
)

// Multi returns a new Service that provides a unified Service interface
// for all of the provided services
func Multi(svcs ...Service) Service {
	return &multi{svcs}
}

type multi struct {
	services []Service
}

func (m *multi) Start(ctx context.Context) error {
	return Start(ctx, m.services...)
}

func (m *multi) Stop(ctx context.Context) error {
	return Stop(ctx, m.services...)
}

func (m *multi) Wait(ctx context.Context) error {
	return Wait(ctx, m.services...)
}

// Start starts multiple services
func Start(ctx context.Context, svcs ...Service) error {
	var (
		err error
		j   int
	)
	for i, svc := range svcs {
		if err = svc.Start(ctx); err != nil {
			j = i
			break
		}
	}

	if err != nil && j > 0 {
		tostop := svcs[:j]
		err = multierr.Combine(err,
			Stop(ctx, tostop...),
		)
	}

	return err
}

// Stop stops multiple services
func Stop(ctx context.Context, svcs ...Service) error {
	var err error
	for _, svc := range svcs {
		err = multierr.Append(err, svc.Stop(ctx))
	}

	return err
}

// Wait waits on multiple services
func Wait(ctx context.Context, svcs ...Service) error {
	errs := make([]chan error, len(svcs))

	for i, svc := range svcs {
		errs[i] = make(chan error)
		ch := errs[i]
		go func(svc Service) {
			ch <- svc.Wait(ctx)
		}(svc)
	}

	var (
		first  = true
		retErr error
		j      = 0
	)

	for {
		for _, ch := range errs {
			select {
			case err := <-ch:
				if first {
					go Stop(ctx, svcs...)
				}

				first = false
				retErr = multierr.Append(retErr, err)

				j++
				if j == len(svcs)-1 {
					return retErr
				}
			default:
			}
		}
	}
}
