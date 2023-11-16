// ui/table.go

package ui

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/nikoksr/dbench/pkg/models"
)

// Renderer defines the interface for rendering tables.
type Renderer interface {
	Render([]*models.Benchmark) string
}

// BenchmarksTableRenderer implements Renderer for the Benchmark model.
type BenchmarksTableRenderer struct{}

// NewBenchmarksTableRenderer creates a new instance of BenchmarksTableRenderer.
func NewBenchmarksTableRenderer() *BenchmarksTableRenderer {
	return &BenchmarksTableRenderer{}
}

// Render renders the table for a slice of benchmarks.
func (r *BenchmarksTableRenderer) Render(benchmarks []*models.Benchmark) string {
	var sb strings.Builder
	t := table.NewWriter()
	t.SetOutputMirror(&sb)
	t.AppendHeader(table.Row{
		"ID",
		"Group ID",
		"Clients",
		"Threads",
		"Transactions",
		"TPS",
		"Avg. Latency",
		"Conn Time",
		"Created At",
	})
	for _, benchmark := range benchmarks {
		t.AppendRow(table.Row{
			benchmark.ID,
			benchmark.GroupID,
			benchmark.Clients,
			benchmark.Threads,
			benchmark.Edges.Result.Transactions,
			fmt.Sprintf("%.2f", benchmark.Edges.Result.TransactionsPerSecond),
			benchmark.Edges.Result.AverageLatency,
			benchmark.Edges.Result.ConnectionTime,
			benchmark.CreatedAt,
		})
	}

	// Set the style for the table.
	t.SetStyle(table.StyleLight)

	// Set the table index.
	t.SetAutoIndex(true)

	// Render the table.
	t.Render()

	return sb.String()
}
