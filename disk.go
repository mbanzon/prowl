package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/shirou/gopsutil/disk"
)

var (
	diskPartitions = disk.Partitions
	diskUsage      = disk.Usage
)

func startDiskUsageReporting(ctx context.Context) chan []diskInfo {
	minimumPauseTime := 5 * time.Second
	diskChannel := make(chan []diskInfo)
	wg := ctx.Value(wgKey).(*sync.WaitGroup)
	wg.Add(1)

	go func() {
		log.Println("Disk reporting started")
		defer wg.Done()

		reportDiskUsage(diskChannel)

		for {
			select {
			case <-ctx.Done():
				log.Println("Disk reporting stopped")
				return
			case <-time.After(minimumPauseTime):
				requestedPauseTime := reportDiskUsage(diskChannel)
				if requestedPauseTime > minimumPauseTime {
					minimumPauseTime = requestedPauseTime
				}
			}
		}
	}()

	return diskChannel
}

func reportDiskUsage(diskChannel chan []diskInfo) time.Duration {
	startTime := time.Now()
	partitions, err := diskPartitions(false)
	if err != nil {
		log.Println("error getting disk partitions:", err)
	}

	disks := []diskInfo{}

	for _, part := range partitions {
		usageStat, err := diskUsage(part.Mountpoint)
		if err != nil {
			log.Println("error getting disk usage:", err)
		}
		total := uint64(0)
		free := uint64(0)
		used := uint64(0)
		if usageStat != nil {
			total = usageStat.Total
			free = usageStat.Free
			used = usageStat.Used
		}

		disks = append(disks, diskInfo{
			Device:     part.Device,
			Mountpoint: part.Mountpoint,
			Total:      total,
			Free:       free,
			Used:       used,
		})
	}

	diskChannel <- disks

	return 2 * time.Since(startTime)
}
