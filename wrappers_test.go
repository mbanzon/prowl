package main

import (
	"testing"
	"time"
)

func TestHandleDataWrapper(t *testing.T) {
	ctx, cancel, wg := testContextWithWaitGroup()
	out := handleData(ctx)

	got := readOutput(t, out)
	if got.Time == 0 {
		t.Fatalf("expected timestamp to be set")
	}

	cancel()
	waitForWaitGroup(t, wg)
}

func TestStartReportingWrappersWithCanceledContext(t *testing.T) {
	ctx, cancel, wg := testContextWithWaitGroup()
	cancel()

	startCpuUsageReporting(ctx)
	startLoadAverageReporting(ctx)
	startMemoryUsageReporting(ctx)
	startSwapUsageReporting(ctx)

	waitForWaitGroup(t, wg)
}

func TestStartDiskUsageReportingWrapper(t *testing.T) {
	ctx, cancel, wg := testContextWithWaitGroup()
	defer cancel()

	ch := startDiskUsageReporting(ctx)

	select {
	case <-ch:
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timed out waiting for disk usage report")
	}

	cancel()
	waitForWaitGroup(t, wg)
}

func TestReportWrappers(t *testing.T) {
	loadCh := make(chan loadInfo, 1)
	reportLoadAverage(loadCh)
	<-loadCh

	memoryCh := make(chan memoryInfo, 1)
	reportMemoryUsage(memoryCh)
	<-memoryCh

	swapCh := make(chan memoryInfo, 1)
	reportSwapUsage(swapCh)
	<-swapCh

	diskCh := make(chan []diskInfo, 1)
	reportDiskUsage(diskCh)
	<-diskCh
}
