package main

import (
	"log"
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

func handleData(cpuIn chan cpuInfo, loadIn chan loadInfo, stopChan chan bool) chan output {
	out := make(chan output)
	data := output{}

	go func() {
		log.Println("Data handling started")

		for {
			select {
			case <-stopChan:
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
