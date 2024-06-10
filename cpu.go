package main

import (
	"log"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
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

func startLoadAverageReporting(stopChan chan bool) chan loadInfo {
	loadChannel := make(chan loadInfo)

	go func() {
		for {
			select {
			case <-stopChan:
				log.Println("Load average reporting stopped")
				return
			case <-time.After(1 * time.Second):
				reportLoadAverage(loadChannel)
			}
		}
	}()

	return loadChannel
}

func reportLoadAverage(loadChannel chan loadInfo) {
	load, err := load.Avg()
	if err != nil {
		log.Println("error getting load average:", err)
	}

	loadChannel <- loadInfo{
		Load1:  load.Load1,
		Load5:  load.Load5,
		Load15: load.Load15,
	}
}
