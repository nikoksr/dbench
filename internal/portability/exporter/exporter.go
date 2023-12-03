package exporter

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"strconv"

	"github.com/nikoksr/dbench/internal/build"
	"github.com/nikoksr/dbench/internal/models"
	"github.com/nikoksr/dbench/internal/ui/text"
)

func createFile(name, extension string) (*os.File, error) {
	if name == "" {
		return os.CreateTemp("", build.AppName+"-*."+extension)
	}
	return os.Create(name)
}

// ToCSV exports a slice of Benchmark structs to a CSV file. An empty filename will create a temporary file.
func ToCSV(benchmarks []*models.Benchmark, filename string) (string, error) {
	file, err := createFile(filename, "csv")
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		// Config
		"ID",
		"GroupID",
		"Comment",
		"Version",
		"Command",
		"TransactionType",
		"QueryMode",
		"ScalingFactor",
		"Clients",
		"Threads",
		// System details
		"MachineID",
		"OsName",
		"OsArch",
		"CPUVendor",
		"CPUModel",
		"CPUCount",
		"CPUCores",
		"CPUThreads",
		"RAMPhysical",
		"RAMUsable",
		"DiskCount",
		"DiskSpaceTotal",
		// PGbench result
		"Transactions",
		"TransactionsPerSecond",
		"FailedTransactions",
		"AverageLatency",
		"ConnectionTime",
		"TotalRuntime",
		// CPU
		"CPUMinLoad",
		"CPUMaxLoad",
		"CPUAverageLoad",
		"CPU50thLoad",
		"CPU75thLoad",
		"CPU90thLoad",
		"CPU95thLoad",
		"CPU99thLoad",
		// Memory
		"MemoryMinLoad",
		"MemoryMaxLoad",
		"MemoryAverageLoad",
		"Memory50thLoad",
		"Memory75thLoad",
		"Memory90thLoad",
		"Memory95thLoad",
		"Memory99thLoad",
		// Misc
		"CreatedAt",
	}
	if err := writer.Write(header); err != nil {
		return "", err
	}

	// Write data rows
	// Write data rows
	for _, benchmark := range benchmarks {
		// Create a copy of the benchmark
		benchmarkCopy := *benchmark

		// If System is nil, initialize it with its zero value
		if benchmarkCopy.Edges.System == nil {
			benchmarkCopy.Edges.System = &models.SystemDetails{}
		}

		// Now use benchmarkCopy instead of benchmark
		record := []string{
			// Config
			benchmarkCopy.ID.String(),
			benchmarkCopy.GroupID.String(),
			benchmarkCopy.Comment,
			benchmarkCopy.Version,
			benchmarkCopy.Command,
			benchmarkCopy.TransactionType,
			benchmarkCopy.QueryMode,
			strconv.FormatFloat(benchmarkCopy.ScalingFactor, 'f', 2, 64),
			strconv.Itoa(benchmarkCopy.Clients),
			strconv.Itoa(benchmarkCopy.Threads),
			// System details
			text.ValueOrNA(benchmarkCopy.Edges.System.MachineID),
			text.ValueOrNA(benchmarkCopy.Edges.System.OsName),
			text.ValueOrNA(benchmarkCopy.Edges.System.OsArch),
			text.ValueOrNA(benchmarkCopy.Edges.System.CPUVendor),
			text.ValueOrNA(benchmarkCopy.Edges.System.CPUModel),
			text.ValueOrNA(benchmarkCopy.Edges.System.CPUCount),
			text.ValueOrNA(benchmarkCopy.Edges.System.CPUCores),
			text.ValueOrNA(benchmarkCopy.Edges.System.CPUThreads),
			text.ValueOrNA(benchmarkCopy.Edges.System.RAMPhysical),
			text.ValueOrNA(benchmarkCopy.Edges.System.RAMUsable),
			text.ValueOrNA(benchmarkCopy.Edges.System.DiskCount),
			text.ValueOrNA(benchmarkCopy.Edges.System.DiskSpaceTotal),
			// PGbench result
			strconv.Itoa(benchmarkCopy.Edges.Result.Transactions),
			strconv.FormatFloat(benchmarkCopy.Edges.Result.TransactionsPerSecond, 'f', 2, 64),
			strconv.Itoa(benchmarkCopy.Edges.Result.FailedTransactions),
			benchmarkCopy.Edges.Result.AverageLatency.String(),
			benchmarkCopy.Edges.Result.ConnectionTime.String(),
			benchmarkCopy.Edges.Result.TotalRuntime.String(),
			// CPU
			strconv.FormatFloat(benchmarkCopy.Edges.SystemMetric.CPUMinLoad, 'f', 2, 64),
			strconv.FormatFloat(benchmarkCopy.Edges.SystemMetric.CPUMaxLoad, 'f', 2, 64),
			strconv.FormatFloat(benchmarkCopy.Edges.SystemMetric.CPUAverageLoad, 'f', 2, 64),
			strconv.FormatFloat(benchmarkCopy.Edges.SystemMetric.CPU50thLoad, 'f', 2, 64),
			strconv.FormatFloat(benchmarkCopy.Edges.SystemMetric.CPU75thLoad, 'f', 2, 64),
			strconv.FormatFloat(benchmarkCopy.Edges.SystemMetric.CPU90thLoad, 'f', 2, 64),
			strconv.FormatFloat(benchmarkCopy.Edges.SystemMetric.CPU95thLoad, 'f', 2, 64),
			strconv.FormatFloat(benchmarkCopy.Edges.SystemMetric.CPU99thLoad, 'f', 2, 64),
			// Memory
			strconv.FormatFloat(benchmarkCopy.Edges.SystemMetric.MemoryMinLoad, 'f', 2, 64),
			strconv.FormatFloat(benchmarkCopy.Edges.SystemMetric.MemoryMaxLoad, 'f', 2, 64),
			strconv.FormatFloat(benchmarkCopy.Edges.SystemMetric.MemoryAverageLoad, 'f', 2, 64),
			strconv.FormatFloat(benchmarkCopy.Edges.SystemMetric.Memory50thLoad, 'f', 2, 64),
			strconv.FormatFloat(benchmarkCopy.Edges.SystemMetric.Memory75thLoad, 'f', 2, 64),
			strconv.FormatFloat(benchmarkCopy.Edges.SystemMetric.Memory90thLoad, 'f', 2, 64),
			strconv.FormatFloat(benchmarkCopy.Edges.SystemMetric.Memory95thLoad, 'f', 2, 64),
			strconv.FormatFloat(benchmarkCopy.Edges.SystemMetric.Memory99thLoad, 'f', 2, 64),
			// Misc
			benchmarkCopy.CreatedAt.Local().Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(record); err != nil {
			return "", err
		}
	}

	return file.Name(), nil
}

// ToJSON takes an interface and attempts to marshal it into JSON format, then write to a file. An empty filename will
// create a temporary file.
func ToJSON(data any, filename string) (string, error) {
	file, err := createFile(filename, "json")
	if err != nil {
		return "", err
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(data); err != nil {
		return "", err
	}

	return file.Name(), nil
}
