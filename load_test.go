package main

import (
	"testing"

	"github.com/shirou/gopsutil/v4/load"
)

func TestReportLoadAverageWith(t *testing.T) {
	ch := make(chan loadInfo, 1)

	reportLoadAverageWith(ch, func() (*load.AvgStat, error) {
		return &load.AvgStat{
			Load1:  1.25,
			Load5:  2.5,
			Load15: 3.75,
		}, nil
	})

	got := <-ch
	if got.Load1 != 1.25 {
		t.Fatalf("expected Load1 to be 1.25, got %v", got.Load1)
	}
	if got.Load5 != 2.5 {
		t.Fatalf("expected Load5 to be 2.5, got %v", got.Load5)
	}
	if got.Load15 != 3.75 {
		t.Fatalf("expected Load15 to be 3.75, got %v", got.Load15)
	}
}
