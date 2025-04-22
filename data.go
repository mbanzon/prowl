package main

import (
	"context"
	"log"
	"sync"
	"time"
)

type output struct {
	Cpu    cpuInfo    `json:"cpu"`
	Load   loadInfo   `json:"load"`
	Memory memoryInfo `json:"memory"`
	Swap   memoryInfo `json:"swap"`
	Disks  []diskInfo `json:"disks"`
	Time   int64      `json:"time"`
}

type cpuInfo struct {
	Usage float64   `json:"usage"`
	Cores []float64 `json:"cores"`
}

type loadInfo struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

type memoryInfo struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

type diskInfo struct {
	Device     string `json:"device"`
	Mountpoint string `json:"mountpoint"`
	Total      uint64 `json:"total"`
	Free       uint64 `json:"free"`
	Used       uint64 `json:"used"`
}

func handleData(ctx context.Context) chan output {
	out := make(chan output)
	data := output{}
	wg := ctx.Value(wgKey).(*sync.WaitGroup)
	wg.Add(1)

	cpuIn := startCpuUsageReporting(ctx)
	loadIn := startLoadAverageReporting(ctx)
	memoryIn, swapIn := startMemoryUsageReporting(ctx)
	diskIn := startDiskUsageReporting(ctx)

	go func() {
		log.Println("Data handling started")
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				close(out)
				log.Println("Data handling stopped")
				return
			case cpuData := <-cpuIn:
				data.Cpu = cpuData
			case loadData := <-loadIn:
				data.Load = loadData
			case memoryData := <-memoryIn:
				data.Memory = memoryData
			case swapData := <-swapIn:
				data.Swap = swapData
			case diskData := <-diskIn:
				data.Disks = diskData
			case <-time.After(1 * time.Second):
				data.Time = time.Now().Unix()
				out <- data
			}
		}
	}()

	return out
}
