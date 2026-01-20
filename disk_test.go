package main

import (
	"testing"

	"github.com/shirou/gopsutil/v4/disk"
)

func TestReportDiskUsageWith(t *testing.T) {
	ch := make(chan []diskInfo, 1)

	partitionsFn := func(_ bool) ([]disk.PartitionStat, error) {
		return []disk.PartitionStat{
			{Device: "/dev/sda1", Mountpoint: "/"},
			{Device: "/dev/sdb1", Mountpoint: "/data"},
		}, nil
	}

	usageFn := func(mountpoint string) (*disk.UsageStat, error) {
		switch mountpoint {
		case "/":
			return &disk.UsageStat{Total: 100, Free: 40, Used: 60}, nil
		case "/data":
			return &disk.UsageStat{Total: 200, Free: 50, Used: 150}, nil
		default:
			return &disk.UsageStat{}, nil
		}
	}

	reportDiskUsageWith(ch, partitionsFn, usageFn)

	got := <-ch
	if len(got) != 2 {
		t.Fatalf("expected 2 disks, got %d", len(got))
	}

	for _, info := range got {
		switch info.Mountpoint {
		case "/":
			if info.Device != "/dev/sda1" || info.Total != 100 || info.Free != 40 || info.Used != 60 {
				t.Fatalf("unexpected root disk data: %+v", info)
			}
		case "/data":
			if info.Device != "/dev/sdb1" || info.Total != 200 || info.Free != 50 || info.Used != 150 {
				t.Fatalf("unexpected data disk data: %+v", info)
			}
		default:
			t.Fatalf("unexpected mountpoint: %s", info.Mountpoint)
		}
	}
}
