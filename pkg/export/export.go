// Package export provides functionality to export data in various formats.
package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/nikoksr/dbench/pkg/models"
)

func openFile(filename string) (*os.File, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// ToCSV exports a slice of Benchmark structs to a CSV file.
func ToCSV(benchmarks []*models.Benchmark, filename string) error {
	file, err := openFile(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		// Config
		"ID",
		"GroupID",
		"Version",
		"Command",
		"TransactionType",
		"ScalingFactor",
		"QueryMode",
		"Clients",
		"Threads",
		// PGbench result
		"Transactions",
		"TransactionsPerSecond",
		"TransactionsPerClient",
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
		"CPU999thLoad",
		"CPU9999thLoad",
		// Memory
		"MemoryMinLoad",
		"MemoryMaxLoad",
		"MemoryAverageLoad",
		"Memory50thLoad",
		"Memory75thLoad",
		"Memory90thLoad",
		"Memory95thLoad",
		"Memory99thLoad",
		"Memory999thLoad",
		"Memory9999thLoad",
		// Misc
		"CreatedAt",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data rows
	for _, benchmark := range benchmarks {
		record := []string{
			// Config
			benchmark.ID.String(),
			benchmark.GroupID.String(),
			benchmark.Version,
			benchmark.Command,
			benchmark.TransactionType,
			strconv.FormatFloat(benchmark.ScalingFactor, 'f', 6, 64),
			benchmark.QueryMode,
			strconv.Itoa(benchmark.Clients),
			strconv.Itoa(benchmark.Threads),
			// PGbench result
			strconv.Itoa(benchmark.Edges.Result.Transactions),
			strconv.FormatFloat(benchmark.Edges.Result.TransactionsPerSecond, 'f', 6, 64),
			strconv.Itoa(benchmark.Edges.Result.FailedTransactions),
			benchmark.Edges.Result.AverageLatency.String(),
			benchmark.Edges.Result.ConnectionTime.String(),
			benchmark.Edges.Result.TotalRuntime.String(),
			// CPU
			strconv.FormatFloat(benchmark.Edges.SystemMetric.CPUMinLoad, 'f', 6, 64),
			strconv.FormatFloat(benchmark.Edges.SystemMetric.CPUMaxLoad, 'f', 6, 64),
			strconv.FormatFloat(benchmark.Edges.SystemMetric.CPUAverageLoad, 'f', 6, 64),
			strconv.FormatFloat(benchmark.Edges.SystemMetric.CPU50thLoad, 'f', 6, 64),
			strconv.FormatFloat(benchmark.Edges.SystemMetric.CPU75thLoad, 'f', 6, 64),
			strconv.FormatFloat(benchmark.Edges.SystemMetric.CPU90thLoad, 'f', 6, 64),
			strconv.FormatFloat(benchmark.Edges.SystemMetric.CPU95thLoad, 'f', 6, 64),
			strconv.FormatFloat(benchmark.Edges.SystemMetric.CPU99thLoad, 'f', 6, 64),
			// Memory
			strconv.FormatFloat(benchmark.Edges.SystemMetric.MemoryMinLoad, 'f', 6, 64),
			strconv.FormatFloat(benchmark.Edges.SystemMetric.MemoryMaxLoad, 'f', 6, 64),
			strconv.FormatFloat(benchmark.Edges.SystemMetric.MemoryAverageLoad, 'f', 6, 64),
			strconv.FormatFloat(benchmark.Edges.SystemMetric.Memory50thLoad, 'f', 6, 64),
			strconv.FormatFloat(benchmark.Edges.SystemMetric.Memory75thLoad, 'f', 6, 64),
			strconv.FormatFloat(benchmark.Edges.SystemMetric.Memory90thLoad, 'f', 6, 64),
			strconv.FormatFloat(benchmark.Edges.SystemMetric.Memory95thLoad, 'f', 6, 64),
			strconv.FormatFloat(benchmark.Edges.SystemMetric.Memory99thLoad, 'f', 6, 64),
			// Misc
			benchmark.CreatedAt.String(),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// ToJSON takes an interface and attempts to marshal it into JSON format, then write to a file.
func ToJSON(data any, filename string) error {
	file, err := openFile(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(data)
}

// ToGnuplotBasic exports a slice of Benchmark structs to a Gnuplot compatible .dat file.
func ToGnuplotBasic(benchmarks []*models.Benchmark, filename string) error {
	file, err := openFile(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, benchmark := range benchmarks {
		line := fmt.Sprintf("%d %f %d %d\n",
			benchmark.Clients,
			benchmark.Edges.Result.TransactionsPerSecond,
			time.Duration(benchmark.Edges.Result.AverageLatency).Milliseconds(),
			time.Duration(benchmark.Edges.Result.ConnectionTime).Milliseconds(),
		)
		_, err := fmt.Fprint(file, line)
		if err != nil {
			return fmt.Errorf("write to file: %w", err)
		}
	}

	return nil
}
