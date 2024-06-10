package main

import (
	"log"
	"time"

	"github.com/shirou/gopsutil/cpu"
)

func startCpuUsageReporting(stopChan chan bool) chan cpuInfo {
	cpuChannel := make(chan cpuInfo)

	go func() {
		for {
			select {
			case <-stopChan:
				log.Println("CPU reporting stopped")
				return
			case <-time.After(1 * time.Second):
				reportCpuUsage(cpuChannel)
			}
		}
	}()

	return cpuChannel
}

func reportCpuUsage(cpuChannel chan cpuInfo) {
	cpuUsages, err := cpu.Percent(0, true)
	if err != nil {
		log.Println("error getting CPU usage:", err)
	}

	combinedUsage, err := cpu.Percent(0, false)
	if err != nil {
		log.Println("error getting CPU usage:", err)
	}

	cpuChannel <- cpuInfo{
		Usage: combinedUsage[0],
		Cores: cpuUsages,
	}
}
