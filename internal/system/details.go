package system

import (
	"fmt"
	"runtime"

	"github.com/jaypipes/ghw"
	"github.com/panta/machineid"

	"github.com/nikoksr/dbench/internal/build"
	"github.com/nikoksr/dbench/internal/models"
)

func toPointer[T any](v T) *T {
	return &v
}

// GetDetails collects information about the system and returns it as a SystemDetails struct. If an error occurs, the
// method will not stop but continue to collect information. The errors will be returned as a slice. We do this since
// the system details are not essential for dbench to run and should not prevent the user from running a benchmark.
func GetDetails() (*models.SystemDetails, []error) {
	var errs []error
	details := new(models.SystemDetails)

	// Generate a unique machine ID
	id, err := machineid.ProtectedID(build.AppName)
	if err != nil {
		errs = append(errs, fmt.Errorf("generate machine id: %w", err))
	} else {
		details.MachineID = toPointer(id)
	}

	// Collect information about the OS
	setOSDetails(details)

	// Collect information about the CPU
	cpu, err := ghw.CPU()
	if err != nil {
		errs = append(errs, fmt.Errorf("get cpu info: %w", err))
	} else {
		setCPUDetails(cpu, details)
	}

	// Collect information about the RAM
	memory, err := ghw.Memory()
	if err != nil {
		errs = append(errs, fmt.Errorf("get memory info: %w", err))
	} else {
		setRAMDetails(memory, details)
	}

	// Collect information about the disks
	block, err := ghw.Block()
	if err != nil {
		errs = append(errs, fmt.Errorf("get block info: %w", err))
	} else {
		setDiskDetails(block, details)
	}

	return details, errs
}

func setOSDetails(details *models.SystemDetails) {
	details.OsName = toPointer(runtime.GOOS)
	details.OsArch = toPointer(runtime.GOARCH)
}

func setCPUDetails(cpu *ghw.CPUInfo, details *models.SystemDetails) {
	details.CPUVendor = &cpu.Processors[0].Vendor
	details.CPUModel = &cpu.Processors[0].Model
	details.CPUCount = toPointer(uint32(len(cpu.Processors)))
	details.CPUCores = &cpu.TotalCores
	details.CPUThreads = &cpu.TotalThreads
}

func setRAMDetails(memory *ghw.MemoryInfo, details *models.SystemDetails) {
	if memory.TotalPhysicalBytes > 0 {
		details.RAMPhysical = toPointer(uint64(memory.TotalPhysicalBytes))
	}
	if memory.TotalUsableBytes > 0 {
		details.RAMUsable = toPointer(uint64(memory.TotalUsableBytes))
	}
}

func setDiskDetails(block *ghw.BlockInfo, details *models.SystemDetails) {
	details.DiskCount = toPointer(uint32(len(block.Disks)))
	details.DiskSpaceTotal = toPointer(block.TotalPhysicalBytes)
}
