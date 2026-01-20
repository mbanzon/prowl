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
	return reportMemoryUsageWith(memoryChannel, mem.VirtualMemory)
}

func reportMemoryUsageWith(memoryChannel chan memoryInfo, virtualMemory func() (*mem.VirtualMemoryStat, error)) time.Duration {
	startTime := time.Now()
	memory, err := virtualMemory()
	if err != nil {
		log.Println("error getting memory usage:", err)
	}
	if memory == nil {
		memory = &mem.VirtualMemoryStat{}
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
	return reportSwapUsageWith(swapChannel, mem.SwapMemory)
}

func reportSwapUsageWith(swapChannel chan memoryInfo, swapMemory func() (*mem.SwapMemoryStat, error)) time.Duration {
	startTime := time.Now()
	swap, err := swapMemory()
	if err != nil {
		log.Println("error getting swap usage:", err)
	}
	if swap == nil {
		swap = &mem.SwapMemoryStat{}
	}

	swapChannel <- memoryInfo{
		Total:       swap.Total,
		Used:        swap.Used,
		Free:        swap.Free,
		UsedPercent: swap.UsedPercent,
	}

	return 2 * time.Since(startTime)
}
