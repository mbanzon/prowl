package main

import (
	"errors"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func TestStartServerWithRetriesAndStops(t *testing.T) {
	in := make(chan output, 1)
	in <- output{Time: 1}
	close(in)

	var attempts atomic.Int32
	listen := func(*http.Server) error {
		tries := attempts.Add(1)
		if tries == 1 {
			return errors.New("listen failed")
		}
		return http.ErrServerClosed
	}

	sleepCalled := make(chan struct{}, 1)
	sleepFn := func(time.Duration) {
		sleepCalled <- struct{}{}
	}

	server := startServerWith(0, "", false, true, in, listen, sleepFn)
	if server == nil || server.Addr != ":0" {
		t.Fatalf("expected server with :0 address, got %+v", server)
	}

	select {
	case <-sleepCalled:
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("expected sleep to be called on retry")
	}

	deadline := time.Now().Add(200 * time.Millisecond)
	for attempts.Load() < 2 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	if attempts.Load() < 2 {
		t.Fatalf("expected listen to be called twice, got %d", attempts.Load())
	}
}
