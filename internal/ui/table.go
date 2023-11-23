// ui/table.go

package ui

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/nikoksr/dbench/internal/models"
	"github.com/nikoksr/dbench/internal/ui/text"
)

//
// FIXME: Move to bubbletea and implement Component interface.
//

var tableStyle table.Style

func init() {
	// Customize the table style.
	tableStyle = table.StyleLight
	tableStyle.Box = table.StyleBoxRounded
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
		"System ID",
		"Comment",
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
			text.ValueOrNA(benchmark.SystemID),
			benchmark.Comment,
			benchmark.Clients,
			benchmark.Threads,
			benchmark.Edges.Result.Transactions,
			fmt.Sprintf("%.2f", benchmark.Edges.Result.TransactionsPerSecond),
			benchmark.Edges.Result.AverageLatency,
			benchmark.Edges.Result.ConnectionTime,
			benchmark.CreatedAt.Local().Format("2006-01-02 15:04:05"),
		})
	}

	// Set the style for the table.
	t.SetStyle(tableStyle)

	// Set the table index.
	t.SetAutoIndex(true)

	t.SetColumnConfigs([]table.ColumnConfig{
		{
			Name:     "Comment",
			WidthMax: 30,
		},
	})

	// Render the table.
	t.Render()

	return sb.String()
}
