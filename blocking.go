package svc

import (
	"context"
	"sync"
)

// WrapBlocking transforms a blocking start func (like http.ListenAndServe)
// and returns a Service
func WrapBlocking(start LifecycleFn, stop LifecycleFn) Service {
	b := &blocking{start: start, stop: stop, ch: make(chan op)}
	go b.init()

	return b
}

type op struct {
	ctx context.Context
	op  string
	ret chan error
	fn  func(string)
}

type blocking struct {
	start LifecycleFn
	stop  LifecycleFn
	mu    sync.Mutex
	ch    chan op
	state int
}

func (b *blocking) init() {
	var (
		state    = StateEmpty
		startCh  = make(chan error)
		startErr error
		stopErr  error
		tonotify = make([]chan error, 0)
	)

	for {
		select {
		case err := <-startCh:
			startErr = err
			state = StateStopped
			for _, ch := range tonotify {
				ch <- startErr
			}
			tonotify = make([]chan error, 0)
		case op := <-b.ch:
			switch op.op {
			case "do":
				op.fn(state)
				op.ret <- nil
			case "start":
				switch state {
				case StateEmpty:
					state = StateStarted
					if b.start != nil {
						go func(ctx context.Context) { startCh <- b.start(ctx) }(op.ctx)
					}

					op.ret <- nil
				case StateStarted:
					op.ret <- nil
				default:
					op.ret <- ErrStartInvalid
				}
			case "stop":
				switch state {
				case StateStarted:
					state = StateStopped
					if b.stop != nil {
						stopErr = b.stop(op.ctx)
					}
					op.ret <- stopErr
					for _, ch := range tonotify {
						ch <- stopErr
					}
					tonotify = make([]chan error, 0)
				case StateStopped:
					op.ret <- stopErr
				default:
					op.ret <- ErrStopInvalid
				}
			case "wait":
				switch state {
				case StateStarted:
					tonotify = append(tonotify, op.ret)
				default:
					op.ret <- ErrWaitInvalid
				}
			}
		}
	}
}

func (b *blocking) Start(ctx context.Context) error {
	ch := make(chan error)
	b.ch <- op{op: "start", ret: ch, ctx: ctx}
	return <-ch
}

func (b *blocking) Stop(ctx context.Context) error {
	ch := make(chan error)
	b.ch <- op{op: "stop", ret: ch, ctx: ctx}
	return <-ch
}

func (b *blocking) Wait(ctx context.Context) error {
	ch := make(chan error)
	b.ch <- op{op: "wait", ret: ch, ctx: ctx}
	return <-ch
}

func (b *blocking) WithStatus(fn func(string)) {
	ch := make(chan error)
	b.ch <- op{op: "do", ret: ch, fn: fn, ctx: nil}
	<-ch
}
