package models

import (
	"github.com/nikoksr/dbench/ent"
)

// BenchmarkMode is a type for benchmark mode
type BenchmarkMode = string

const (
	ModeSimple   BenchmarkMode = "simple"   // ModeSimple is a simple benchmark mode
	ModeThorough BenchmarkMode = "thorough" // ModeThorough is a thorough benchmark mode
)

type (
	// Benchmark represents a benchmark.
	Benchmark = ent.Benchmark

	// BenchmarkResult represents the result of a benchmark run.
	BenchmarkResult = ent.BenchmarkResult

	// SystemConfig represents a system config.
	SystemConfig = ent.SystemConfig

	// SystemMetric represents a system metric.
	SystemMetric = ent.SystemMetric

	// SystemSample represents a system sample.
	SystemSample struct {
		CPULoad    float64
		MemoryLoad float64
	}

	// BenchmarkCSV is the CSV-exportable type for models.Benchmark.
	BenchmarkCSV struct {
		ID                    string `csv:"ID"`
		GroupID               string `csv:"GroupID"`
		Comment               string `csv:"Comment"`
		Version               string `csv:"Version"`
		Command               string `csv:"Command"`
		TransactionType       string `csv:"TransactionType"`
		QueryMode             string `csv:"QueryMode"`
		ScalingFactor         string `csv:"ScalingFactor"`
		Clients               string `csv:"Clients"`
		Threads               string `csv:"Threads"`
		MachineID             string `csv:"MachineID"`
		OsName                string `csv:"OsName"`
		OsArch                string `csv:"OsArch"`
		CPUVendor             string `csv:"CPUVendor"`
		CPUModel              string `csv:"CPUModel"`
		CPUCount              string `csv:"CPUCount"`
		CPUCores              string `csv:"CPUCores"`
		CPUThreads            string `csv:"CPUThreads"`
		RAMPhysical           string `csv:"RAMPhysical"`
		RAMUsable             string `csv:"RAMUsable"`
		DiskCount             string `csv:"DiskCount"`
		DiskSpaceTotal        string `csv:"DiskSpaceTotal"`
		Transactions          string `csv:"Transactions"`
		TransactionsPerSecond string `csv:"TransactionsPerSecond"`
		FailedTransactions    string `csv:"FailedTransactions"`
		AverageLatency        string `csv:"AverageLatency"`
		ConnectionTime        string `csv:"ConnectionTime"`
		TotalRuntime          string `csv:"TotalRuntime"`
		CPUMinLoad            string `csv:"CPUMinLoad"`
		CPUMaxLoad            string `csv:"CPUMaxLoad"`
		CPUAverageLoad        string `csv:"CPUAverageLoad"`
		CPU50thLoad           string `csv:"CPU50thLoad"`
		CPU75thLoad           string `csv:"CPU75thLoad"`
		CPU90thLoad           string `csv:"CPU90thLoad"`
		CPU95thLoad           string `csv:"CPU95thLoad"`
		CPU99thLoad           string `csv:"CPU99thLoad"`
		MemoryMinLoad         string `csv:"MemoryMinLoad"`
		MemoryMaxLoad         string `csv:"MemoryMaxLoad"`
		MemoryAverageLoad     string `csv:"MemoryAverageLoad"`
		Memory50thLoad        string `csv:"Memory50thLoad"`
		Memory75thLoad        string `csv:"Memory75thLoad"`
		Memory90thLoad        string `csv:"Memory90thLoad"`
		Memory95thLoad        string `csv:"Memory95thLoad"`
		Memory99thLoad        string `csv:"Memory99thLoad"`
		RecordedAt            string `csv:"RecordedAt"`
	}
)
