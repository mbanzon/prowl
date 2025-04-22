package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/shirou/gopsutil/mem"
)

func startMemoryUsageReporting(ctx context.Context) (chan memoryInfo, chan memoryInfo) {
	memoryChannel := make(chan memoryInfo)
	swapChannel := make(chan memoryInfo)
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
			case <-time.After(1 * time.Second):
				reportMemoryUsage(memoryChannel)
				reportSwapUsage(swapChannel)
			}
		}
	}()

	return memoryChannel, swapChannel
}

func reportMemoryUsage(memoryChannel chan memoryInfo) {
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
}

func reportSwapUsage(swapChannel chan memoryInfo) {
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
}
