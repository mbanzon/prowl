package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/load"
)

func startCpuUsageReporting(ctx context.Context) chan cpuInfo {
	return startCpuUsageReportingWith(ctx, cpu.Percent, time.After)
}

func startCpuUsageReportingWith(ctx context.Context, percent func(time.Duration, bool) ([]float64, error), after func(time.Duration) <-chan time.Time) chan cpuInfo {
	minimumPauseTime := 1 * time.Second
	cpuChannel := make(chan cpuInfo)
	wg := ctx.Value(wgKey).(*sync.WaitGroup)
	wg.Add(1)

	go func() {
		log.Println("CPU reporting started")
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				log.Println("CPU reporting stopped")
				return
			case <-after(minimumPauseTime):
				requestedPauseTime := reportCpuUsage(cpuChannel, percent)
				if requestedPauseTime > minimumPauseTime {
					minimumPauseTime = requestedPauseTime
				}
			}
		}
	}()

	return cpuChannel
}

func reportCpuUsage(cpuChannel chan cpuInfo, percent func(time.Duration, bool) ([]float64, error)) (requestedPauseTime time.Duration) {
	startTime := time.Now()

	cpuUsages, err := percent(0, true)
	if err != nil {
		log.Println("error getting CPU usage:", err)
	}

	combinedUsage, err := percent(0, false)
	if err != nil {
		log.Println("error getting CPU usage:", err)
	}

	usage := 0.0
	if len(combinedUsage) > 0 {
		usage = combinedUsage[0]
	} else {
		log.Println("error getting CPU usage: empty combined usage")
	}

	cpuChannel <- cpuInfo{
		Usage: usage,
		Cores: cpuUsages,
	}

	return 2 * time.Since(startTime)
}

func startLoadAverageReporting(ctx context.Context) chan loadInfo {
	return startLoadAverageReportingWith(ctx, load.Avg, time.After)
}

func startLoadAverageReportingWith(ctx context.Context, avg func() (*load.AvgStat, error), after func(time.Duration) <-chan time.Time) chan loadInfo {
	minimumPauseTime := 1 * time.Second
	loadChannel := make(chan loadInfo)
	wg := ctx.Value(wgKey).(*sync.WaitGroup)
	wg.Add(1)

	go func() {
		log.Println("Load average reporting started")
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				log.Println("Load average reporting stopped")
				return
			case <-after(minimumPauseTime):
				requestedPauseTime := reportLoadAverageWith(loadChannel, avg)
				if requestedPauseTime > minimumPauseTime {
					minimumPauseTime = requestedPauseTime
				}
			}
		}
	}()

	return loadChannel
}

func reportLoadAverage(loadChannel chan loadInfo) (requestedPauseTime time.Duration) {
	return reportLoadAverageWith(loadChannel, load.Avg)
}

func reportLoadAverageWith(loadChannel chan loadInfo, avg func() (*load.AvgStat, error)) (requestedPauseTime time.Duration) {
	startTime := time.Now()

	loadAvg, err := avg()
	if err != nil {
		log.Println("error getting load average:", err)
	}
	if loadAvg == nil {
		loadAvg = &load.AvgStat{}
	}

	loadChannel <- loadInfo{
		Load1:  loadAvg.Load1,
		Load5:  loadAvg.Load5,
		Load15: loadAvg.Load15,
	}

	return 2 * time.Since(startTime)
}
