package svc

import (
	"context"
	"sync"

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
	var (
		wg    sync.WaitGroup
		l     sync.Mutex
		first bool
		err   error
	)

	for _, svc := range svcs {
		wg.Add(1)
		go func(svc Service) {
			defer wg.Done()
			serr := svc.Wait(ctx)
			l.Lock()
			defer l.Unlock()
			if serr != nil {
				err = multierr.Append(err, serr)
			}
			if !first {
				go Stop(ctx, svcs...)
			}
			first = true
		}(svc)
	}

	wg.Wait()
	return err
}
