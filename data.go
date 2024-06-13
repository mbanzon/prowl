package main

import (
	"context"
	"log"
	"sync"
	"time"
)

type output struct {
	Cpu  cpuInfo  `json:"cpu"`
	Load loadInfo `json:"load"`
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

func handleData(ctx context.Context) chan output {
	out := make(chan output)
	data := output{}
	wg := ctx.Value(wgKey).(*sync.WaitGroup)
	wg.Add(1)

	cpuIn := startCpuUsageReporting(ctx)
	loadIn := startLoadAverageReporting(ctx)

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
			case <-time.After(1 * time.Second):
				out <- data
			}
		}
	}()

	return out
}
