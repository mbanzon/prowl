package main

import (
	"testing"
	"time"
)

func TestShutdownContextHasDeadline(t *testing.T) {
	ctx, cancel := shutdownContext()
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatalf("expected shutdown context to have a deadline")
	}

	remaining := time.Until(deadline)
	if remaining <= 0 {
		t.Fatalf("expected shutdown deadline in the future, got %v", remaining)
	}
	if remaining < shutdownTimeout-time.Second || remaining > shutdownTimeout {
		t.Fatalf("expected shutdown deadline near %v, got %v", shutdownTimeout, remaining)
	}
}
