package system

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/nikoksr/dbench/internal/models"
)

// CollectMetrics monitors the system and sends samples to the sample channel.
func CollectMetrics(interval time.Duration, stopChan <-chan struct{}, sampleChan chan<- models.SystemSample) error {
	for {
		select {
		case <-stopChan:
			close(sampleChan)
			return nil
		case <-time.After(interval):
			// Get system metrics
			sample, err := getMetrics()
			if err != nil {
				return fmt.Errorf("get system metrics: %w", err)
			}

			// Send sample to the channel
			sampleChan <- sample
		}
	}
}

func getMetrics() (models.SystemSample, error) {
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return models.SystemSample{}, fmt.Errorf("get cpu usage: %w", err)
	}

	memUsage, err := mem.VirtualMemory()
	if err != nil {
		return models.SystemSample{}, fmt.Errorf("get memory usage: %w", err)
	}

	return models.SystemSample{
		CPULoad:    cpuPercent[0],
		MemoryLoad: memUsage.UsedPercent,
	}, nil
}
