package models

import "runtime"

// BenchmarkConfig holds the configuration for benchmarking
type BenchmarkConfig struct {
	// Database
	DBName   string
	Username string
	Host     string
	Port     string

	// Benchmark-Init options
	FillFactor  int
	ScaleFactor int

	// Benchmark-Run options
	Mode       BenchmarkMode // Mode is the benchmarking mode
	NumThreads int           // NumThreads is the number of threads to use
	NumClients int           // NumClients is the number of clients to use
}

func (c *BenchmarkConfig) Sanitize() {
	// Database
	if c.Host == "" {
		c.Host = "localhost"
	}
	if c.Port == "" {
		c.Port = "5432"
	}
	if c.DBName == "" {
		c.DBName = "postgres"
	}
	if c.Username == "" {
		c.Username = "postgres"
	}

	// Benchmark-Init options
	if c.FillFactor < 10 {
		c.FillFactor = 100
	}
	if c.FillFactor > 100 {
		c.FillFactor = 100
	}

	if c.ScaleFactor < 1 {
		c.ScaleFactor = 1
	}
	if c.ScaleFactor > 1000 {
		c.ScaleFactor = 1000
	}

	// Benchmark-Run options
	if c.Mode == "" {
		c.Mode = ModeSimple
	}

	if c.NumThreads == -1 {
		// Use the number of CPU cores
		c.NumThreads = runtime.NumCPU()
	}
	if c.NumThreads < 1 {
		c.NumThreads = 1
	}

	if c.NumClients < 1 {
		c.NumClients = 1
	}
}
