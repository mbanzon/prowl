package main

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestHandleDataWithChannels(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	ctx = context.WithValue(ctx, wgKey, wg)

	cpuIn := make(chan cpuInfo, 1)
	loadIn := make(chan loadInfo, 1)
	memoryIn := make(chan memoryInfo, 1)
	swapIn := make(chan memoryInfo, 1)
	diskIn := make(chan []diskInfo, 1)

	out := handleDataWithChannels(ctx, cpuIn, loadIn, memoryIn, swapIn, diskIn)

	cpuIn <- cpuInfo{Usage: 55.5, Cores: []float64{11.1, 22.2}}
	got := readOutput(t, out)
	if got.Cpu.Usage != 55.5 || len(got.Cpu.Cores) != 2 {
		t.Fatalf("unexpected cpu output: %+v", got.Cpu)
	}
	if got.Time == 0 {
		t.Fatalf("expected timestamp to be set")
	}

	loadIn <- loadInfo{Load1: 1.1, Load5: 2.2, Load15: 3.3}
	got = readOutput(t, out)
	if got.Load.Load1 != 1.1 || got.Load.Load5 != 2.2 || got.Load.Load15 != 3.3 {
		t.Fatalf("unexpected load output: %+v", got.Load)
	}

	memoryIn <- memoryInfo{Total: 100, Used: 60, Free: 40, UsedPercent: 60}
	got = readOutput(t, out)
	if got.Memory.Total != 100 || got.Memory.UsedPercent != 60 {
		t.Fatalf("unexpected memory output: %+v", got.Memory)
	}

	swapIn <- memoryInfo{Total: 200, Used: 20, Free: 180, UsedPercent: 10}
	got = readOutput(t, out)
	if got.Swap.Total != 200 || got.Swap.UsedPercent != 10 {
		t.Fatalf("unexpected swap output: %+v", got.Swap)
	}

	diskIn <- []diskInfo{{Device: "/dev/sda1", Mountpoint: "/", Total: 500, Free: 100, Used: 400}}
	got = readOutput(t, out)
	if len(got.Disks) != 1 || got.Disks[0].Device != "/dev/sda1" {
		t.Fatalf("unexpected disk output: %+v", got.Disks)
	}

	cancel()
	waitForWaitGroup(t, wg)

	select {
	case _, ok := <-out:
		if ok {
			t.Fatalf("expected output channel to be closed")
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("timed out waiting for output channel to close")
	}
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
