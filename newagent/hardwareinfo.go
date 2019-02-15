package main

import (
	"github.com/jaypipes/ghw"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type HardwareInfo struct {
	Processor string `json:"processor"`
	Memory    uint64 `json:"memory"`
	DiskSize  uint64 `json:"disk_size"`
}

func NewHardwareInfo() (*HardwareInfo, error) {
	cpuinfo, err := cpu.Info()
	if err != nil {
		return nil, err
	}

	meminfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	block, err := ghw.Block()
	if err != nil {
		return nil, err
	}

	return &HardwareInfo{
		Processor: cpuinfo[0].ModelName,
		Memory:    meminfo.Total,
		DiskSize:  block.TotalPhysicalBytes,
	}, nil
}
