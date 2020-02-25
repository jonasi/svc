package svc

import (
	"context"
	"errors"
)

// Valid state values
const (
	StateEmpty   = "empty"
	StateStarted = "started"
	StateStopped = "stopped"
)

// errors
var (
	ErrStartInvalid = errors.New("Attempting to start a service that has already been started")
	ErrStopInvalid  = errors.New("Attempting to stop a service that is not running")
	ErrWaitInvalid  = errors.New("Attempting to wait on a service that is not running")
)

// A Service is a long running process that can
// start, stop and wait.
type Service interface {
	Start(context.Context) error
	Stop(context.Context) error
	Wait(context.Context) error
}

// LifecycleFn is a helper type for describing
// the start, stop, wait methods
type LifecycleFn func(context.Context) error

// LifecycleFnNoCtx is a helper that turns a func() error into a LifecycleFn
func LifecycleFnNoCtx(fn func() error) LifecycleFn {
	return func(context.Context) error { return fn() }
}
