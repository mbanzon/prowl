package main

import (
	"testing"
	"time"
)

func TestReportCpuUsageGuardsEmptyCombinedUsage(t *testing.T) {
	stubPercent := func(d time.Duration, percpu bool) ([]float64, error) {
		if percpu {
			return []float64{12.5, 33.3}, nil
		}
		return []float64{}, nil
	}

	ch := make(chan cpuInfo, 1)
	reportCpuUsage(ch, stubPercent)

	got := <-ch
	if got.Usage != 0 {
		t.Fatalf("expected Usage to be 0 for empty combined usage, got %v", got.Usage)
	}
	if len(got.Cores) != 2 {
		t.Fatalf("expected 2 core entries, got %d", len(got.Cores))
	}
}
