package models

import (
	typeid "go.jetpack.io/typeid/typed"

	"github.com/nikoksr/dbench/ent"
)

type (
	// Benchmark represents a benchmark.
	Benchmark = ent.Benchmark

	// BenchmarkResult represents the result of a benchmark run.
	BenchmarkResult = ent.BenchmarkResult

	benchmarkGroupPrefix struct{}
	BenchmarkGroupID     struct {
		typeid.TypeID[benchmarkGroupPrefix]
	}

	// SystemMetric represents a system metric.
	SystemMetric = ent.SystemMetric

	SystemSample struct {
		CPULoad    float64
		MemoryLoad float64
	}

	// BenchmarkMode is a type for benchmark mode
	BenchmarkMode = string
)

const (
	ModeSimple   BenchmarkMode = "simple"   // ModeSimple is a simple benchmark mode
	ModeThorough BenchmarkMode = "thorough" // ModeThorough is a thorough benchmark mode
)

// Type returns the type of the result group ID.
func (benchmarkGroupPrefix) Type() string { return "bmkgrp" }
