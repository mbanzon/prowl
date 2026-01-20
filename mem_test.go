package main

import (
	"testing"

	"github.com/shirou/gopsutil/v4/mem"
)

func TestReportMemoryUsageWith(t *testing.T) {
	ch := make(chan memoryInfo, 1)

	reportMemoryUsageWith(ch, func() (*mem.VirtualMemoryStat, error) {
		return &mem.VirtualMemoryStat{
			Total:       1000,
			Used:        600,
			Free:        400,
			UsedPercent: 60.0,
		}, nil
	})

	got := <-ch
	if got.Total != 1000 {
		t.Fatalf("expected Total to be 1000, got %d", got.Total)
	}
	if got.Used != 600 {
		t.Fatalf("expected Used to be 600, got %d", got.Used)
	}
	if got.Free != 400 {
		t.Fatalf("expected Free to be 400, got %d", got.Free)
	}
	if got.UsedPercent != 60.0 {
		t.Fatalf("expected UsedPercent to be 60.0, got %v", got.UsedPercent)
	}
}

func TestReportSwapUsageWith(t *testing.T) {
	ch := make(chan memoryInfo, 1)

	reportSwapUsageWith(ch, func() (*mem.SwapMemoryStat, error) {
		return &mem.SwapMemoryStat{
			Total:       2000,
			Used:        500,
			Free:        1500,
			UsedPercent: 25.0,
		}, nil
	})

	got := <-ch
	if got.Total != 2000 {
		t.Fatalf("expected Total to be 2000, got %d", got.Total)
	}
	if got.Used != 500 {
		t.Fatalf("expected Used to be 500, got %d", got.Used)
	}
	if got.Free != 1500 {
		t.Fatalf("expected Free to be 1500, got %d", got.Free)
	}
	if got.UsedPercent != 25.0 {
		t.Fatalf("expected UsedPercent to be 25.0, got %v", got.UsedPercent)
	}
}
