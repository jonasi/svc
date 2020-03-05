package svc

// WrapStartStop wraps a start function that returns immediately
// and returns a Service
func WrapStartStop(start LifecycleFn, stop LifecycleFn) Service {
	b := &base{start: start, startBlocking: false, stop: stop, ch: make(chan op)}
	go b.init()

	return b
}
