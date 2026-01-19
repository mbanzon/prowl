package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/mem"
)

func startMemoryUsageReporting(ctx context.Context) chan memoryInfo {
	minimumPauseTime := 1 * time.Second
	memoryChannel := make(chan memoryInfo)
	wg := ctx.Value(wgKey).(*sync.WaitGroup)
	wg.Add(1)

	go func() {
		log.Println("Memory reporting started")
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				log.Println("Memory reporting stopped")
				return
			case <-time.After(minimumPauseTime):
				requestedPauseTime := reportMemoryUsage(memoryChannel)
				if requestedPauseTime > minimumPauseTime {
					minimumPauseTime = requestedPauseTime
				}
			}
		}
	}()

	return memoryChannel
}

func startSwapUsageReporting(ctx context.Context) chan memoryInfo {
	minimumPauseTime := 1 * time.Second
	swapChannel := make(chan memoryInfo)
	wg := ctx.Value(wgKey).(*sync.WaitGroup)
	wg.Add(1)

	go func() {
		log.Println("Swap reporting started")
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				log.Println("Swap reporting stopped")
				return
			case <-time.After(minimumPauseTime):
				requestedPauseTime := reportSwapUsage(swapChannel)
				if requestedPauseTime > minimumPauseTime {
					minimumPauseTime = requestedPauseTime
				}
			}
		}
	}()

	return swapChannel
}

func reportMemoryUsage(memoryChannel chan memoryInfo) time.Duration {
	startTime := time.Now()
	memory, err := mem.VirtualMemory()
	if err != nil {
		log.Println("error getting memory usage:", err)
	}

	memoryChannel <- memoryInfo{
		Total:       memory.Total,
		Used:        memory.Used,
		Free:        memory.Free,
		UsedPercent: memory.UsedPercent,
	}

	return 2 * time.Since(startTime)
}

func reportSwapUsage(swapChannel chan memoryInfo) time.Duration {
	startTime := time.Now()
	swap, err := mem.SwapMemory()
	if err != nil {
		log.Println("error getting swap usage:", err)
	}

	swapChannel <- memoryInfo{
		Total:       swap.Total,
		Used:        swap.Used,
		Free:        swap.Free,
		UsedPercent: swap.UsedPercent,
	}

	return 2 * time.Since(startTime)
}
