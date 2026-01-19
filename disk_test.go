package main

import (
	"errors"
	"testing"

	"github.com/shirou/gopsutil/disk"
)

func TestReportDiskUsageHandlesNilUsage(t *testing.T) {
	originalPartitions := diskPartitions
	originalUsage := diskUsage
	diskPartitions = func(all bool) ([]disk.PartitionStat, error) {
		return []disk.PartitionStat{
			{Device: "dev0", Mountpoint: "/mnt"},
		}, nil
	}
	diskUsage = func(path string) (*disk.UsageStat, error) {
		return nil, errors.New("boom")
	}
	t.Cleanup(func() {
		diskPartitions = originalPartitions
		diskUsage = originalUsage
	})

	ch := make(chan []diskInfo, 1)
	reportDiskUsage(ch)

	got := <-ch
	if len(got) != 1 {
		t.Fatalf("expected one disk entry, got %d", len(got))
	}
	if got[0].Total != 0 || got[0].Free != 0 || got[0].Used != 0 {
		t.Fatalf("expected zeroed disk usage for nil stats, got %+v", got[0])
	}
}

func TestReportDiskUsageHandlesPartitionError(t *testing.T) {
	originalPartitions := diskPartitions
	diskPartitions = func(all bool) ([]disk.PartitionStat, error) {
		return nil, errors.New("boom")
	}
	t.Cleanup(func() {
		diskPartitions = originalPartitions
	})

	ch := make(chan []diskInfo, 1)
	reportDiskUsage(ch)

	got := <-ch
	if len(got) != 0 {
		t.Fatalf("expected no disks when partitions fail, got %d", len(got))
	}
}

func TestReportDiskUsageUsesUsageStats(t *testing.T) {
	originalPartitions := diskPartitions
	originalUsage := diskUsage
	diskPartitions = func(all bool) ([]disk.PartitionStat, error) {
		return []disk.PartitionStat{
			{Device: "dev1", Mountpoint: "/data"},
		}, nil
	}
	diskUsage = func(path string) (*disk.UsageStat, error) {
		return &disk.UsageStat{Total: 10, Free: 4, Used: 6}, nil
	}
	t.Cleanup(func() {
		diskPartitions = originalPartitions
		diskUsage = originalUsage
	})

	ch := make(chan []diskInfo, 1)
	reportDiskUsage(ch)

	got := <-ch
	if len(got) != 1 {
		t.Fatalf("expected one disk entry, got %d", len(got))
	}
	if got[0].Total != 10 || got[0].Free != 4 || got[0].Used != 6 {
		t.Fatalf("unexpected disk usage values: %+v", got[0])
	}
}
