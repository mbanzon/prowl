package main

import (
	"log"
	"time"
)

type output struct {
	Cpu cpuInfo `json:"cpu"`
}

type cpuInfo struct {
	Usage float64   `json:"usage"`
	Cores []float64 `json:"cores"`
}

func handleData(cpuIn chan cpuInfo, stopChan chan bool) chan output {
	out := make(chan output)
	data := output{}

	go func() {
		for {
			select {
			case <-stopChan:
				close(out)
				log.Println("Data handling stopped")
				return
			case cpuData := <-cpuIn:
				data.Cpu = cpuData
			case <-time.After(1 * time.Second):
				out <- data
			}
		}
	}()

	return out
}
