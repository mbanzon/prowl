package main

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
)

func TestStartCpuUsageReportingWith(t *testing.T) {
	ctx, cancel, wg := testContextWithWaitGroup()
	defer cancel()

	afterCh := make(chan time.Time, 1)
	ch := startCpuUsageReportingWith(ctx, func(time.Duration, bool) ([]float64, error) {
		return []float64{10.5, 20.5}, nil
	}, func(time.Duration) <-chan time.Time {
		return afterCh
	})

	afterCh <- time.Now()
	select {
	case got := <-ch:
		if got.Usage != 10.5 {
			t.Fatalf("expected usage 10.5, got %v", got.Usage)
		}
		if len(got.Cores) != 2 {
			t.Fatalf("expected 2 cores, got %d", len(got.Cores))
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("timed out waiting for cpu usage")
	}

	cancel()
	waitForWaitGroup(t, wg)
}

func TestStartLoadAverageReportingWith(t *testing.T) {
	ctx, cancel, wg := testContextWithWaitGroup()
	defer cancel()

	afterCh := make(chan time.Time, 1)
	ch := startLoadAverageReportingWith(ctx, func() (*load.AvgStat, error) {
		return &load.AvgStat{Load1: 1, Load5: 2, Load15: 3}, nil
	}, func(time.Duration) <-chan time.Time {
		return afterCh
	})

	afterCh <- time.Now()
	select {
	case got := <-ch:
		if got.Load1 != 1 || got.Load5 != 2 || got.Load15 != 3 {
			t.Fatalf("unexpected load averages: %+v", got)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("timed out waiting for load average")
	}

	cancel()
	waitForWaitGroup(t, wg)
}

func TestStartMemoryUsageReportingWith(t *testing.T) {
	ctx, cancel, wg := testContextWithWaitGroup()
	defer cancel()

	afterCh := make(chan time.Time, 1)
	ch := startMemoryUsageReportingWith(ctx, func() (*mem.VirtualMemoryStat, error) {
		return &mem.VirtualMemoryStat{Total: 100, Used: 50, Free: 50, UsedPercent: 50}, nil
	}, func(time.Duration) <-chan time.Time {
		return afterCh
	})

	afterCh <- time.Now()
	select {
	case got := <-ch:
		if got.Total != 100 || got.Used != 50 || got.Free != 50 || got.UsedPercent != 50 {
			t.Fatalf("unexpected memory stats: %+v", got)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("timed out waiting for memory usage")
	}

	cancel()
	waitForWaitGroup(t, wg)
}

func TestStartSwapUsageReportingWith(t *testing.T) {
	ctx, cancel, wg := testContextWithWaitGroup()
	defer cancel()

	afterCh := make(chan time.Time, 1)
	ch := startSwapUsageReportingWith(ctx, func() (*mem.SwapMemoryStat, error) {
		return &mem.SwapMemoryStat{Total: 200, Used: 60, Free: 140, UsedPercent: 30}, nil
	}, func(time.Duration) <-chan time.Time {
		return afterCh
	})

	afterCh <- time.Now()
	select {
	case got := <-ch:
		if got.Total != 200 || got.Used != 60 || got.Free != 140 || got.UsedPercent != 30 {
			t.Fatalf("unexpected swap stats: %+v", got)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("timed out waiting for swap usage")
	}

	cancel()
	waitForWaitGroup(t, wg)
}

func TestStartDiskUsageReportingWith(t *testing.T) {
	ctx, cancel, wg := testContextWithWaitGroup()
	defer cancel()

	afterCh := make(chan time.Time, 1)
	partitionsFn := func(bool) ([]disk.PartitionStat, error) {
		return []disk.PartitionStat{{Device: "/dev/sda1", Mountpoint: "/"}}, nil
	}
	usageFn := func(string) (*disk.UsageStat, error) {
		return &disk.UsageStat{Total: 10, Used: 4, Free: 6}, nil
	}

	ch := startDiskUsageReportingWith(ctx, partitionsFn, usageFn, func(time.Duration) <-chan time.Time {
		return afterCh
	})

	select {
	case got := <-ch:
		if len(got) != 1 || got[0].Device != "/dev/sda1" {
			t.Fatalf("unexpected disk stats: %+v", got)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("timed out waiting for initial disk usage")
	}

	afterCh <- time.Now()
	select {
	case got := <-ch:
		if len(got) != 1 || got[0].Used != 4 {
			t.Fatalf("unexpected disk stats after tick: %+v", got)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("timed out waiting for disk usage tick")
	}

	cancel()
	waitForWaitGroup(t, wg)
}

func testContextWithWaitGroup() (context.Context, context.CancelFunc, *sync.WaitGroup) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	ctx = context.WithValue(ctx, wgKey, wg)
	return ctx, cancel, wg
}

func waitForWaitGroup(t *testing.T, wg *sync.WaitGroup) {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("timed out waiting for goroutine to stop")
	}
}
