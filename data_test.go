package main

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestHandleDataCollectsUpdates(t *testing.T) {
	originalCpuReporter := cpuUsageReporter
	originalLoadReporter := loadAverageReporter
	originalMemoryReporter := memoryUsageReporter
	originalSwapReporter := swapUsageReporter
	originalDiskReporter := diskUsageReporter

	cpuCh := make(chan cpuInfo, 1)
	loadCh := make(chan loadInfo, 1)
	memoryCh := make(chan memoryInfo, 1)
	swapCh := make(chan memoryInfo, 1)
	diskCh := make(chan []diskInfo, 1)

	cpuUsageReporter = func(ctx context.Context) chan cpuInfo {
		return cpuCh
	}
	loadAverageReporter = func(ctx context.Context) chan loadInfo {
		return loadCh
	}
	memoryUsageReporter = func(ctx context.Context) chan memoryInfo {
		return memoryCh
	}
	swapUsageReporter = func(ctx context.Context) chan memoryInfo {
		return swapCh
	}
	diskUsageReporter = func(ctx context.Context) chan []diskInfo {
		return diskCh
	}

	t.Cleanup(func() {
		cpuUsageReporter = originalCpuReporter
		loadAverageReporter = originalLoadReporter
		memoryUsageReporter = originalMemoryReporter
		swapUsageReporter = originalSwapReporter
		diskUsageReporter = originalDiskReporter
	})

	wg := &sync.WaitGroup{}
	ctx := context.WithValue(context.Background(), wgKey, wg)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	out := handleData(ctx)

	cpuCh <- cpuInfo{Usage: 42.0, Cores: []float64{42.0}}
	select {
	case got := <-out:
		if got.Cpu.Usage != 42.0 {
			t.Fatalf("expected cpu usage 42, got %v", got.Cpu.Usage)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for cpu update")
	}

	memoryCh <- memoryInfo{Total: 10, Used: 3, Free: 7, UsedPercent: 30}
	select {
	case got := <-out:
		if got.Cpu.Usage != 42.0 {
			t.Fatalf("expected cpu usage to persist, got %v", got.Cpu.Usage)
		}
		if got.Memory.Total != 10 || got.Memory.Used != 3 || got.Memory.Free != 7 {
			t.Fatalf("unexpected memory update: %+v", got.Memory)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for memory update")
	}

	cancel()
	wg.Wait()

	if _, ok := <-out; ok {
		t.Fatal("expected output channel to be closed after cancel")
	}
}
