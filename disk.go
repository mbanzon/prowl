package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/disk"
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
	return reportDiskUsageWith(diskChannel, disk.Partitions, disk.Usage)
}

func reportDiskUsageWith(diskChannel chan []diskInfo, partitions func(bool) ([]disk.PartitionStat, error), usage func(string) (*disk.UsageStat, error)) time.Duration {
	startTime := time.Now()
	diskPartitions, err := partitions(false)
	if err != nil {
		log.Println("error getting disk partitions:", err)
	}

	disks := []diskInfo{}

	for _, part := range diskPartitions {
		usageStat, err := usage(part.Mountpoint)
		if err != nil {
			log.Println("error getting disk usage:", err)
		}
		if usageStat == nil {
			usageStat = &disk.UsageStat{}
		}

		disks = append(disks, diskInfo{
			Device:     part.Device,
			Mountpoint: part.Mountpoint,
			Total:      usageStat.Total,
			Free:       usageStat.Free,
			Used:       usageStat.Used,
		})
	}

	diskChannel <- disks

	return 2 * time.Since(startTime)
}
