package converter

import (
	"strconv"

	"github.com/nikoksr/dbench/internal/models"
	"github.com/nikoksr/dbench/internal/ui/text"
)

// BenchmarkToCSV converts a benchmark to a CSV-exportable type.
func BenchmarkToCSV(b *models.Benchmark) *models.BenchmarkCSV {
	// Avoid nil pointer dereference. Also makes the code below more readable.
	system := new(models.SystemConfig)
	if b.Edges.System != nil {
		system = b.Edges.System
	}

	systemMetric := new(models.SystemMetric)
	if b.Edges.SystemMetric != nil {
		systemMetric = b.Edges.SystemMetric
	}

	result := new(models.BenchmarkResult)
	if b.Edges.Result != nil {
		result = b.Edges.Result
	}

	// Convert to CSV-exportable type
	return &models.BenchmarkCSV{
		// Config

		ID:              b.ID.String(),
		GroupID:         b.GroupID.String(),
		Comment:         text.ValueOrNA(b.Comment),
		Version:         b.Version,
		Command:         b.Command,
		TransactionType: b.TransactionType,
		QueryMode:       b.QueryMode,
		ScalingFactor:   strconv.FormatFloat(b.ScalingFactor, 'f', 2, 64),
		Clients:         strconv.Itoa(b.Clients),
		Threads:         strconv.Itoa(b.Threads),

		// System config

		MachineID:      text.ValueOrNA(system.MachineID),
		OsName:         text.ValueOrNA(system.OsName),
		OsArch:         text.ValueOrNA(system.OsArch),
		CPUVendor:      text.ValueOrNA(system.CPUVendor),
		CPUModel:       text.ValueOrNA(system.CPUModel),
		CPUCount:       text.ValueOrNA(system.CPUCount),
		CPUCores:       text.ValueOrNA(system.CPUCores),
		CPUThreads:     text.ValueOrNA(system.CPUThreads),
		RAMPhysical:    text.ValueOrNA(system.RAMPhysical),
		RAMUsable:      text.ValueOrNA(system.RAMUsable),
		DiskCount:      text.ValueOrNA(system.DiskCount),
		DiskSpaceTotal: text.ValueOrNA(system.DiskSpaceTotal),

		// Benchmark result

		Transactions:          strconv.Itoa(result.Transactions),
		TransactionsPerSecond: strconv.FormatFloat(result.TransactionsPerSecond, 'f', 2, 64),
		FailedTransactions:    strconv.Itoa(result.FailedTransactions),
		AverageLatency:        result.AverageLatency.String(),
		ConnectionTime:        result.ConnectionTime.String(),
		TotalRuntime:          result.TotalRuntime.String(),

		// CPU

		CPUMinLoad:     strconv.FormatFloat(systemMetric.CPUMinLoad, 'f', 2, 64),
		CPUMaxLoad:     strconv.FormatFloat(systemMetric.CPUMaxLoad, 'f', 2, 64),
		CPUAverageLoad: strconv.FormatFloat(systemMetric.CPUAverageLoad, 'f', 2, 64),
		CPU50thLoad:    strconv.FormatFloat(systemMetric.CPU50thLoad, 'f', 2, 64),
		CPU75thLoad:    strconv.FormatFloat(systemMetric.CPU75thLoad, 'f', 2, 64),
		CPU90thLoad:    strconv.FormatFloat(systemMetric.CPU90thLoad, 'f', 2, 64),
		CPU95thLoad:    strconv.FormatFloat(systemMetric.CPU95thLoad, 'f', 2, 64),
		CPU99thLoad:    strconv.FormatFloat(systemMetric.CPU99thLoad, 'f', 2, 64),

		// Memory

		MemoryMinLoad:     strconv.FormatFloat(systemMetric.MemoryMinLoad, 'f', 2, 64),
		MemoryMaxLoad:     strconv.FormatFloat(systemMetric.MemoryMaxLoad, 'f', 2, 64),
		MemoryAverageLoad: strconv.FormatFloat(systemMetric.MemoryAverageLoad, 'f', 2, 64),
		Memory50thLoad:    strconv.FormatFloat(systemMetric.Memory50thLoad, 'f', 2, 64),
		Memory75thLoad:    strconv.FormatFloat(systemMetric.Memory75thLoad, 'f', 2, 64),
		Memory90thLoad:    strconv.FormatFloat(systemMetric.Memory90thLoad, 'f', 2, 64),
		Memory95thLoad:    strconv.FormatFloat(systemMetric.Memory95thLoad, 'f', 2, 64),
		Memory99thLoad:    strconv.FormatFloat(systemMetric.Memory99thLoad, 'f', 2, 64),

		// Misc

		RecordedAt: text.PrettyTime(b.RecordedAt),
	}
}

// BenchmarksToCSV converts a slice of benchmarks to a CSV-exportable type.
func BenchmarksToCSV(benchmarks []*models.Benchmark) []*models.BenchmarkCSV {
	csvBenchmarks := make([]*models.BenchmarkCSV, 0, len(benchmarks))
	for _, b := range benchmarks {
		csvBenchmarks = append(csvBenchmarks, BenchmarkToCSV(b))
	}
	return csvBenchmarks
}
