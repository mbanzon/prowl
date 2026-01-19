package main

import (
	"errors"
	"testing"

	"github.com/shirou/gopsutil/mem"
)

func TestReportMemoryUsageHandlesNilStat(t *testing.T) {
	originalVirtualMemory := virtualMemory
	virtualMemory = func() (*mem.VirtualMemoryStat, error) {
		return nil, errors.New("boom")
	}
	t.Cleanup(func() {
		virtualMemory = originalVirtualMemory
	})

	ch := make(chan memoryInfo, 1)
	reportMemoryUsage(ch)

	got := <-ch
	if got.Total != 0 || got.Used != 0 || got.Free != 0 || got.UsedPercent != 0 {
		t.Fatalf("expected zeroed memory info for nil stats, got %+v", got)
	}
}

func TestReportSwapUsageHandlesNilStat(t *testing.T) {
	originalSwapMemory := swapMemory
	swapMemory = func() (*mem.SwapMemoryStat, error) {
		return nil, errors.New("boom")
	}
	t.Cleanup(func() {
		swapMemory = originalSwapMemory
	})

	ch := make(chan memoryInfo, 1)
	reportSwapUsage(ch)

	got := <-ch
	if got.Total != 0 || got.Used != 0 || got.Free != 0 || got.UsedPercent != 0 {
		t.Fatalf("expected zeroed swap info for nil stats, got %+v", got)
	}
}
