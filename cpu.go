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
			case <-time.After(minimumPauseTime):
				requestedPauseTime := reportCpuUsage(cpuChannel, cpu.Percent)
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
			case <-time.After(minimumPauseTime):
				requestedPauseTime := reportLoadAverage(loadChannel)
				if requestedPauseTime > minimumPauseTime {
					minimumPauseTime = requestedPauseTime
				}
			}
		}
	}()

	return loadChannel
}

func reportLoadAverage(loadChannel chan loadInfo) (requestedPauseTime time.Duration) {
	startTime := time.Now()

	load, err := load.Avg()
	if err != nil {
		log.Println("error getting load average:", err)
	}

	loadChannel <- loadInfo{
		Load1:  load.Load1,
		Load5:  load.Load5,
		Load15: load.Load15,
	}

	return 2 * time.Since(startTime)
}
