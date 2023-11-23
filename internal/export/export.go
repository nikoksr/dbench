// Package export provides functionality to export data in various formats.
package export

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"strconv"

	"github.com/nikoksr/dbench/internal/build"
	"github.com/nikoksr/dbench/internal/models"
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
	for _, benchmark := range benchmarks {
		record := []string{
			// Config
			benchmark.ID.String(),
			benchmark.GroupID.String(),
			benchmark.Comment,
			benchmark.Version,
			benchmark.Command,
			benchmark.TransactionType,
			benchmark.QueryMode,
			strconv.FormatFloat(benchmark.ScalingFactor, 'f', 2, 64),
			strconv.Itoa(benchmark.Clients),
			strconv.Itoa(benchmark.Threads),
			// PGbench result
			strconv.Itoa(benchmark.Edges.Result.Transactions),
			strconv.FormatFloat(benchmark.Edges.Result.TransactionsPerSecond, 'f', 2, 64),
			strconv.Itoa(benchmark.Edges.Result.FailedTransactions),
			benchmark.Edges.Result.AverageLatency.String(),
			benchmark.Edges.Result.ConnectionTime.String(),
			benchmark.Edges.Result.TotalRuntime.String(),
			// CPU
			strconv.FormatFloat(benchmark.Edges.SystemMetric.CPUMinLoad, 'f', 2, 64),
			strconv.FormatFloat(benchmark.Edges.SystemMetric.CPUMaxLoad, 'f', 2, 64),
			strconv.FormatFloat(benchmark.Edges.SystemMetric.CPUAverageLoad, 'f', 2, 64),
			strconv.FormatFloat(benchmark.Edges.SystemMetric.CPU50thLoad, 'f', 2, 64),
			strconv.FormatFloat(benchmark.Edges.SystemMetric.CPU75thLoad, 'f', 2, 64),
			strconv.FormatFloat(benchmark.Edges.SystemMetric.CPU90thLoad, 'f', 2, 64),
			strconv.FormatFloat(benchmark.Edges.SystemMetric.CPU95thLoad, 'f', 2, 64),
			strconv.FormatFloat(benchmark.Edges.SystemMetric.CPU99thLoad, 'f', 2, 64),
			// Memory
			strconv.FormatFloat(benchmark.Edges.SystemMetric.MemoryMinLoad, 'f', 2, 64),
			strconv.FormatFloat(benchmark.Edges.SystemMetric.MemoryMaxLoad, 'f', 2, 64),
			strconv.FormatFloat(benchmark.Edges.SystemMetric.MemoryAverageLoad, 'f', 2, 64),
			strconv.FormatFloat(benchmark.Edges.SystemMetric.Memory50thLoad, 'f', 2, 64),
			strconv.FormatFloat(benchmark.Edges.SystemMetric.Memory75thLoad, 'f', 2, 64),
			strconv.FormatFloat(benchmark.Edges.SystemMetric.Memory90thLoad, 'f', 2, 64),
			strconv.FormatFloat(benchmark.Edges.SystemMetric.Memory95thLoad, 'f', 2, 64),
			strconv.FormatFloat(benchmark.Edges.SystemMetric.Memory99thLoad, 'f', 2, 64),
			// Misc
			benchmark.CreatedAt.Local().Format("2006-01-02 15:04:05"),
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
