package system

import (
	"fmt"
	"runtime"

	"github.com/jaypipes/ghw"
	"github.com/panta/machineid"

	"github.com/nikoksr/dbench/internal/build"
	"github.com/nikoksr/dbench/internal/models"
	"github.com/nikoksr/dbench/internal/pointer"
)

// GetConfig collects information about the system and returns it as a SystemConfig struct. If an error occurs, the
// method will not stop but continue to collect information. The errors will be returned as a slice. We do this since
// the system config are not essential for dbench to run and should not prevent the user from running a benchmark.
func GetConfig() (*models.SystemConfig, []error) {
	var errs []error
	systemconfig := new(models.SystemConfig)

	// Generate a unique machine ID
	id, err := machineid.ProtectedID(build.AppName)
	if err != nil {
		errs = append(errs, fmt.Errorf("generate machine id: %w", err))
	} else {
		systemconfig.MachineID = pointer.To(id)
	}

	// Collect information about the OS
	setOSConfig(systemconfig)

	// Collect information about the CPU
	cpu, err := ghw.CPU()
	if err != nil {
		errs = append(errs, fmt.Errorf("get cpu info: %w", err))
	} else {
		setCPUConfig(cpu, systemconfig)
	}

	// Collect information about the RAM
	memory, err := ghw.Memory()
	if err != nil {
		errs = append(errs, fmt.Errorf("get memory info: %w", err))
	} else {
		setRAMConfig(memory, systemconfig)
	}

	// Collect information about the disks
	block, err := ghw.Block()
	if err != nil {
		errs = append(errs, fmt.Errorf("get block info: %w", err))
	} else {
		setDiskConfig(block, systemconfig)
	}

	return systemconfig, errs
}

func setOSConfig(systemconfig *models.SystemConfig) {
	systemconfig.OsName = pointer.To(runtime.GOOS)
	systemconfig.OsArch = pointer.To(runtime.GOARCH)
}

func setCPUConfig(cpu *ghw.CPUInfo, systemconfig *models.SystemConfig) {
	systemconfig.CPUVendor = &cpu.Processors[0].Vendor
	systemconfig.CPUModel = &cpu.Processors[0].Model
	systemconfig.CPUCount = pointer.To(uint32(len(cpu.Processors)))
	systemconfig.CPUCores = &cpu.TotalCores
	systemconfig.CPUThreads = &cpu.TotalThreads
}

func setRAMConfig(memory *ghw.MemoryInfo, systemconfig *models.SystemConfig) {
	if memory.TotalPhysicalBytes > 0 {
		systemconfig.RAMPhysical = pointer.To(uint64(memory.TotalPhysicalBytes))
	}
	if memory.TotalUsableBytes > 0 {
		systemconfig.RAMUsable = pointer.To(uint64(memory.TotalUsableBytes))
	}
}

func setDiskConfig(block *ghw.BlockInfo, systemconfig *models.SystemConfig) {
	systemconfig.DiskCount = pointer.To(uint32(len(block.Disks)))
	systemconfig.DiskSpaceTotal = pointer.To(block.TotalPhysicalBytes)
}
