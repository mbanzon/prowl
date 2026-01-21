package main

import (
	"testing"
	"time"
)

func TestTeeOutput(t *testing.T) {
	ctx, cancel, wg := testContextWithWaitGroup()
	defer cancel()

	in := make(chan output, 1)
	outA, outB := teeOutput(ctx, in)

	in <- output{Time: 7}
	close(in)

	if got := <-outA; got.Time != 7 {
		t.Fatalf("expected output time 7, got %d", got.Time)
	}
	if got := <-outB; got.Time != 7 {
		t.Fatalf("expected output time 7, got %d", got.Time)
	}

	waitForWaitGroup(t, wg)
}

func readOutput(t *testing.T, out <-chan output) output {
	t.Helper()

	select {
	case got := <-out:
		return got
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timed out waiting for output")
	}

	return output{}
}
