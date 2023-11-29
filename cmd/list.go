package cmd

import (
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/spf13/cobra"

	"github.com/nikoksr/dbench/cmd/cobrax"
	"github.com/nikoksr/dbench/ent"
	"github.com/nikoksr/dbench/internal/database"
	"github.com/nikoksr/dbench/internal/ui"
)

type listOptions struct {
	*globalOptions

	sort []string
}

func newListCommand(globalOpts *globalOptions) *cobra.Command {
	opts := &listOptions{
		globalOptions: globalOpts,
	}

	db := new(database.Database)

	cmd := &cobra.Command{
		Use:                   "list [OPTIONS]",
		Aliases:               []string{"l", "ls"},
		GroupID:               "commands",
		Short:                 "List previously run benchmarks",
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		Args:                  cobra.NoArgs,
		ValidArgsFunction:     cobra.NoFileCompletions,
		PreRunE:               cobrax.HooksE(prepareDBHook(db, globalOpts.dataDir)),
		RunE: func(cmd *cobra.Command, args []string) error {
			if db == nil {
				return fmt.Errorf("database is nil")
			}

			// Query benchmarks
			orderByFunc := func(query *ent.BenchmarkQuery) *ent.BenchmarkQuery {
				for _, order := range opts.sort {
					orderFunc := parseOrderBy(order)
					query = query.Order(orderFunc)
				}

				return query
			}

			benchmarks, err := db.Fetch(cmd.Context(), database.WithOrderBy(orderByFunc))
			if err != nil {
				return fmt.Errorf("fetch benchmarks: %w", err)
			}

			// Render benchmarks
			renderer := ui.NewBenchmarksTableRenderer()
			tableStr := renderer.Render(benchmarks)
			fmt.Println(tableStr)

			return nil
		},
		PostRunE: cobrax.HooksE(closeDatabaseHook(db)),
	}

	// Flags
	cmd.Flags().StringSliceVar(&opts.sort, "sort", []string{"id"}, "Sort benchmarks columns (+/- prefix for ascending/descending)")

	return cmd
}

func parseOrderBy(orderBy string) func(*sql.Selector) {
	if orderBy == "" {
		return nil
	}

	// Check if order is ascending or descending
	prefix := "+"
	if strings.HasPrefix(orderBy, "-") {
		prefix = "-"
	}

	// Remove prefix
	orderBy = strings.TrimPrefix(orderBy, prefix)

	// Create order function
	orderFunc := ent.Asc(orderBy)
	if prefix == "-" {
		orderFunc = ent.Desc(orderBy)
	}

	return orderFunc
}
