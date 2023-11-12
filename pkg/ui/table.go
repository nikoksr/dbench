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
	Render([]*models.Result) string
}

// ResultTableRenderer implements Renderer for the Result model.
type ResultTableRenderer struct{}

// NewResultTableRenderer creates a new instance of ResultTableRenderer.
func NewResultTableRenderer() *ResultTableRenderer {
	return &ResultTableRenderer{}
}

// Render renders the table for a slice of Results.
func (r *ResultTableRenderer) Render(results []*models.Result) string {
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
	for _, result := range results {
		t.AppendRow(table.Row{
			result.ID,
			result.GroupID,
			result.Clients,
			result.Threads,
			result.Transactions,
			fmt.Sprintf("%.2f", result.TransactionsPerSecond),
			result.AverageLatency,
			result.InitialConnectionTime,
			result.CreatedAt,
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
