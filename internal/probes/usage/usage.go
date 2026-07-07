package usage

import (
	"context"
	"os"

	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"

	"github.com/martinbhatta/ctrl/internal/probes"
)

type ResourceUsage struct {
	Name       string
	UsedBytes  uint64
	FreeBytes  uint64
	TotalBytes uint64
	UsedPct    float64
	Detail     string
	Level      probes.Level
	Err        error
}

func Check(ctx context.Context) []ResourceUsage {
	return []ResourceUsage{
		checkMemory(ctx),
		checkDisk(ctx),
	}
}

func checkMemory(ctx context.Context) ResourceUsage {
	stat, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return errorUsage("RAM", err)
	}

	return ResourceUsage{
		Name:       "RAM",
		UsedBytes:  stat.Used,
		FreeBytes:  stat.Available,
		TotalBytes: stat.Total,
		UsedPct:    stat.UsedPercent,
		Detail:     "available memory",
		Level:      levelForUsedPct(stat.UsedPercent),
	}
}

func checkDisk(ctx context.Context) ResourceUsage {
	path, err := diskPath()
	if err != nil {
		return errorUsage("Disk", err)
	}

	stat, err := disk.UsageWithContext(ctx, path)
	if err != nil {
		return errorUsage("Disk", err)
	}

	return ResourceUsage{
		Name:       "Disk",
		UsedBytes:  stat.Used,
		FreeBytes:  stat.Free,
		TotalBytes: stat.Total,
		UsedPct:    stat.UsedPercent,
		Detail:     path,
		Level:      levelForUsedPct(stat.UsedPercent),
	}
}

func diskPath() (string, error) {
	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		return home, nil
	}

	return os.Getwd()
}

func errorUsage(name string, err error) ResourceUsage {
	return ResourceUsage{
		Name:   name,
		Detail: err.Error(),
		Level:  probes.LevelError,
		Err:    err,
	}
}

func levelForUsedPct(usedPct float64) probes.Level {
	switch {
	case usedPct >= 90:
		return probes.LevelError
	case usedPct >= 75:
		return probes.LevelWarning
	default:
		return probes.LevelOK
	}
}
