package svc_test

import (
	"context"
	"testing"

	"github.com/jonasi/svc"
)

// test that you can call start twice
func TestBlocking_StartTwice(t *testing.T) {
	svc := svc.WrapBlocking(nil, nil)
	ctx := context.Background()
	err := svc.Start(ctx)
	if err != nil {
		t.Errorf("Received unexpected error: %s", err)
	}
	err = svc.Start(ctx)
	if err != nil {
		t.Errorf("Received unexpected error: %s", err)
	}
}

// TestBlocking_StartCancel tests that calling stop
// will send a cancel signal to the start function ctx
func TestBlocking_StartCancel(t *testing.T) {
	cancelled := false
	svc := svc.WrapBlocking(func(ctx context.Context) error {
		<-ctx.Done()
		cancelled = true
		return ctx.Err()
	}, nil)

	ctx := context.Background()
	err := svc.Start(ctx)
	if err != nil {
		t.Errorf("Received unexpected error: %s", err)
	}
	err = svc.Stop(ctx)
	if err != nil {
		t.Errorf("Received unexpected error: %s", err)
	}

	if !cancelled {
		t.Errorf("Expected cancelled to be true")
	}
}
