package converter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nikoksr/dbench/ent"
	"github.com/nikoksr/dbench/ent/schema/duration"
	"github.com/nikoksr/dbench/ent/schema/pulid"
	"github.com/nikoksr/dbench/internal/models"
	"github.com/nikoksr/dbench/internal/pointer"
	"github.com/nikoksr/dbench/internal/ui/text"
)

func TestBenchmarkToCSV(t *testing.T) {
	t.Parallel()

	staticTime := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

	b := &models.Benchmark{
		ID:              pulid.ID("1"),
		GroupID:         pulid.ID("1"),
		Comment:         pointer.To("Test Comment"),
		Version:         "1.0",
		Command:         "Test Command",
		TransactionType: "Test TransactionType",
		QueryMode:       "Test QueryMode",
		ScalingFactor:   1.0,
		Clients:         1,
		Threads:         1,
		Edges: ent.BenchmarkEdges{
			System: &models.SystemConfig{
				MachineID:      pointer.To("Test MachineID"),
				OsName:         pointer.To("Test OsName"),
				OsArch:         pointer.To("Test OsArch"),
				CPUVendor:      pointer.To("Test CPUVendor"),
				CPUModel:       pointer.To("Test CPUModel"),
				CPUCount:       pointer.To(uint32(1)),
				CPUCores:       pointer.To(uint32(1)),
				CPUThreads:     pointer.To(uint32(1)),
				RAMPhysical:    pointer.To(uint64(1)),
				RAMUsable:      pointer.To(uint64(1)),
				DiskCount:      pointer.To(uint32(1)),
				DiskSpaceTotal: pointer.To(uint64(1)),
			},
			SystemMetric: &models.SystemMetric{
				CPUMinLoad:        1.0,
				CPUMaxLoad:        1.0,
				CPUAverageLoad:    1.0,
				CPU50thLoad:       1.0,
				CPU75thLoad:       1.0,
				CPU90thLoad:       1.0,
				CPU95thLoad:       1.0,
				CPU99thLoad:       1.0,
				MemoryMinLoad:     1.0,
				MemoryMaxLoad:     1.0,
				MemoryAverageLoad: 1.0,
				Memory50thLoad:    1.0,
				Memory75thLoad:    1.0,
				Memory90thLoad:    1.0,
				Memory95thLoad:    1.0,
				Memory99thLoad:    1.0,
			},
			Result: &models.BenchmarkResult{
				Transactions:          1,
				TransactionsPerSecond: 1.0,
				FailedTransactions:    1,
				AverageLatency:        duration.Duration(1),
				ConnectionTime:        duration.Duration(1),
				TotalRuntime:          duration.Duration(1),
			},
		},
		RecordedAt: staticTime,
	}

	csv := BenchmarkToCSV(b)

	assert.Equal(t, "1", csv.ID)
	assert.Equal(t, "1", csv.GroupID)
	assert.Equal(t, "Test Comment", csv.Comment)
	assert.Equal(t, "1.0", csv.Version)
	assert.Equal(t, "Test Command", csv.Command)
	assert.Equal(t, "Test TransactionType", csv.TransactionType)
	assert.Equal(t, "Test QueryMode", csv.QueryMode)
	assert.Equal(t, "1.00", csv.ScalingFactor)
	assert.Equal(t, "1", csv.Clients)
	assert.Equal(t, "1", csv.Threads)
	assert.Equal(t, "Test MachineID", csv.MachineID)
	assert.Equal(t, "Test OsName", csv.OsName)
	assert.Equal(t, "Test OsArch", csv.OsArch)
	assert.Equal(t, "Test CPUVendor", csv.CPUVendor)
	assert.Equal(t, "Test CPUModel", csv.CPUModel)
	assert.Equal(t, "1", csv.CPUCount)
	assert.Equal(t, "1", csv.CPUCores)
	assert.Equal(t, "1", csv.CPUThreads)
	assert.Equal(t, "1", csv.RAMPhysical)
	assert.Equal(t, "1", csv.RAMUsable)
	assert.Equal(t, "1", csv.DiskCount)
	assert.Equal(t, "1", csv.DiskSpaceTotal)
	assert.Equal(t, "1", csv.Transactions)
	assert.Equal(t, "1.00", csv.TransactionsPerSecond)
	assert.Equal(t, "1", csv.FailedTransactions)
	assert.Equal(t, "1ns", csv.AverageLatency)
	assert.Equal(t, "1ns", csv.ConnectionTime)
	assert.Equal(t, "1ns", csv.TotalRuntime)
	assert.Equal(t, "1.00", csv.CPUMinLoad)
	assert.Equal(t, "1.00", csv.CPUMaxLoad)
	assert.Equal(t, "1.00", csv.CPUAverageLoad)
	assert.Equal(t, "1.00", csv.CPU50thLoad)
	assert.Equal(t, "1.00", csv.CPU75thLoad)
	assert.Equal(t, "1.00", csv.CPU90thLoad)
	assert.Equal(t, "1.00", csv.CPU95thLoad)
	assert.Equal(t, "1.00", csv.CPU99thLoad)
	assert.Equal(t, "1.00", csv.MemoryMinLoad)
	assert.Equal(t, "1.00", csv.MemoryMaxLoad)
	assert.Equal(t, "1.00", csv.MemoryAverageLoad)
	assert.Equal(t, "1.00", csv.Memory50thLoad)
	assert.Equal(t, "1.00", csv.Memory75thLoad)
	assert.Equal(t, "1.00", csv.Memory90thLoad)
	assert.Equal(t, "1.00", csv.Memory95thLoad)
	assert.Equal(t, "1.00", csv.Memory99thLoad)
	assert.Equal(t, text.PrettyTime(staticTime), csv.RecordedAt)
}

func TestBenchmarksToCSV(t *testing.T) {
	t.Parallel()

	staticTime := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

	benchmarks := []*models.Benchmark{
		{
			ID: pulid.ID("1"),
			Edges: ent.BenchmarkEdges{
				System: &models.SystemConfig{
					MachineID: pointer.To("Test MachineID"),
				},
				SystemMetric: &models.SystemMetric{
					CPUMinLoad: 1.0,
				},
				Result: &models.BenchmarkResult{
					Transactions: 1,
				},
			},
			RecordedAt: staticTime,
		},
		{
			ID: pulid.ID("2"),
			Edges: ent.BenchmarkEdges{
				System: &models.SystemConfig{
					MachineID: pointer.To("Test MachineID 2"),
				},
				SystemMetric: &models.SystemMetric{
					CPUMinLoad: 2.0,
				},
				Result: &models.BenchmarkResult{
					Transactions: 2,
				},
			},
			RecordedAt: staticTime,
		},
	}

	csvBenchmarks := BenchmarksToCSV(benchmarks)

	assert.Equal(t, "1", csvBenchmarks[0].ID)
	assert.Equal(t, "Test MachineID", csvBenchmarks[0].MachineID)
	assert.Equal(t, "1.00", csvBenchmarks[0].CPUMinLoad)
	assert.Equal(t, "1", csvBenchmarks[0].Transactions)
	assert.Equal(t, text.PrettyTime(staticTime), csvBenchmarks[0].RecordedAt)

	assert.Equal(t, "2", csvBenchmarks[1].ID)
	assert.Equal(t, "Test MachineID 2", csvBenchmarks[1].MachineID)
	assert.Equal(t, "2.00", csvBenchmarks[1].CPUMinLoad)
	assert.Equal(t, "2", csvBenchmarks[1].Transactions)
	assert.Equal(t, text.PrettyTime(staticTime), csvBenchmarks[1].RecordedAt)
}
