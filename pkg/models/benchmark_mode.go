package models

// BenchmarkMode is a type for benchmark mode
type BenchmarkMode = string

const (
	// ModeSimple is a simple benchmark mode
	ModeSimple BenchmarkMode = "simple"
	// ModeThorough is a thorough benchmark mode
	ModeThorough BenchmarkMode = "thorough"
)
