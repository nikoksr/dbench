package cmd

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/spf13/cobra"

	"github.com/nikoksr/dbench/ent"
	"github.com/nikoksr/dbench/ent/benchmark"
	"github.com/nikoksr/dbench/ent/benchmarkresult"
	"github.com/nikoksr/dbench/ent/systemconfig"
	"github.com/nikoksr/dbench/internal/database"
	"github.com/nikoksr/dbench/internal/fs"
	"github.com/nikoksr/dbench/internal/ui"
	"github.com/nikoksr/dbench/internal/ui/printer"
	"github.com/nikoksr/dbench/internal/ui/text"
)

const (
	defaultPage    = 1
	defaultPerPage = 9 // 9 because a benchmark-group holds 9 benchmarks by default
)

type listOptions struct {
	*globalOptions

	sort                      []string
	page, perPage, totalPages int
}

func newListCommand(globalOpts *globalOptions, connectToDB dbConnector) *cobra.Command {
	opts := &listOptions{
		globalOptions: globalOpts,
	}

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
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := connectToDB(cmd.Context(), opts.dataDir, opts.noMigration, fs.OSFileSystem{})
			if err != nil {
				return fmt.Errorf("connect to database: %w", err)
			}

			p := printer.NewPrinter(cmd.OutOrStdout())

			// Get the total benchmark count and calculate the total number of pages for pagination
			benchmarksCount, err := db.Count(cmd.Context())
			if err != nil {
				return fmt.Errorf("count benchmarks: %w", err)
			}

			// Validate options
			errs := opts.validate(benchmarksCount)
			if len(errs) > 0 {
				p.Spacer(2)
				fmt.Println(text.WarningsList(errs))
				return nil
			}

			// Query benchmarks
			var orderByFunc func(*ent.BenchmarkQuery) *ent.BenchmarkQuery
			sortConfig := newListOrderConfig()

			if len(opts.sort) > 0 {
				orderByFunc = func(query *ent.BenchmarkQuery) *ent.BenchmarkQuery {
					for _, columnID := range opts.sort {
						if orderBy := parseOrderBy(sortConfig, columnID); orderBy != nil {
							query = query.Order(orderBy)
						}
					}

					return query
				}
			}

			benchmarks, err := db.Fetch(cmd.Context(),
				database.WithOrderBy(orderByFunc),
				database.WithLimit(opts.perPage),
				database.WithOffset((opts.page-1)*opts.perPage),
			)
			if err != nil {
				return fmt.Errorf("fetch benchmarks: %w", err)
			}

			// Render benchmarks
			renderer := ui.NewBenchmarksTableRenderer()
			tableStr := renderer.Render(benchmarks)

			p.Spacer(1)
			// p.PrintText(tableStr)
			fmt.Print(tableStr) // Using fmt.Print instead of p.PrintText because of the table formatting

			// Render pagination info
			p.PrintlnHint(fmt.Sprintf(" Showing page %d of %d", opts.page, opts.totalPages))
			p.Spacer(1)

			return nil
		},
	}

	// Flags
	cmd.Flags().IntVar(&opts.page, "page", defaultPage, "Page number")
	cmd.Flags().IntVar(&opts.perPage, "per-page", defaultPerPage, "Number of benchmarks per page")
	cmd.Flags().StringSliceVar(&opts.sort, "sort", nil, "Sort benchmarks columns (+/- for ascending/descending)")

	cmd.Flags().SortFlags = false

	return cmd
}

func (opts *listOptions) validate(recordsCount uint64) []error {
	var errs []error

	// Special case: no records
	if recordsCount == 0 {
		opts.perPage = 0
		opts.page = 0
		opts.totalPages = 0

		errs = append(errs, fmt.Errorf("no records found"))

		return errs
	}

	// Avoid negative page numbers
	if opts.page < 1 {
		opts.page = defaultPage
		errs = append(errs, fmt.Errorf("negative page number -> setting page to %d", opts.page))
	}
	if opts.perPage < 1 {
		opts.perPage = defaultPerPage
		errs = append(errs, fmt.Errorf("negative per page number -> setting per page to %d", opts.perPage))
	}

	// Calculate the total number of pages for pagination
	opts.totalPages = int(math.Ceil(float64(recordsCount) / float64(opts.perPage)))

	// Check if page number is out of range
	if opts.page > opts.totalPages && opts.totalPages > 0 {
		opts.page = opts.totalPages
		errs = append(errs, fmt.Errorf("page out of range -> showing page %d of %d", opts.page, opts.totalPages))
	}

	// Check if per page hits our upper boundary
	if opts.perPage > 250 {
		opts.perPage = 250
		errs = append(errs, fmt.Errorf("per page out of range -> setting per page to %d", opts.perPage))
	}

	// Check if per page number is out of range
	if uint64(opts.perPage) > recordsCount {
		opts.page = defaultPage
		opts.perPage = int(recordsCount) // This is safe because of the previous check
		errs = append(errs, fmt.Errorf("per page out of range -> showing page %d of %d", opts.page, opts.totalPages))
	}

	return errs
}

type (
	listOrderOptionFunc func(opts ...sql.OrderTermOption) benchmark.OrderOption

	listOrderColumnConfig struct {
		index      int
		optionFunc listOrderOptionFunc
	}

	listOrderConfig struct {
		columnToEntOption map[string]listOrderColumnConfig
		idxToColumn       map[int]string
	}
)

func newListOrderConfig() *listOrderConfig {
	cfg := new(listOrderConfig)

	cfg.columnToEntOption = map[string]listOrderColumnConfig{
		benchmark.FieldID:      {index: 1, optionFunc: benchmark.ByID},
		benchmark.FieldGroupID: {index: 2, optionFunc: benchmark.ByGroupID},
		systemconfig.FieldMachineID: {index: 3, optionFunc: func(opts ...sql.OrderTermOption) benchmark.OrderOption {
			return benchmark.BySystemField(systemconfig.FieldMachineID, opts...)
		}},
		benchmark.FieldComment: {index: 4, optionFunc: benchmark.ByComment},
		benchmark.FieldClients: {index: 5, optionFunc: benchmark.ByClients},
		benchmark.FieldThreads: {index: 6, optionFunc: benchmark.ByThreads},
		benchmarkresult.FieldTransactions: {index: 7, optionFunc: func(opts ...sql.OrderTermOption) benchmark.OrderOption {
			return benchmark.ByResultField(benchmarkresult.FieldTransactions, opts...)
		}},
		benchmarkresult.FieldTransactionsPerSecond: {index: 8, optionFunc: func(opts ...sql.OrderTermOption) benchmark.OrderOption {
			return benchmark.ByResultField(benchmarkresult.FieldTransactionsPerSecond, opts...)
		}},
		benchmarkresult.FieldAverageLatency: {index: 9, optionFunc: func(opts ...sql.OrderTermOption) benchmark.OrderOption {
			return benchmark.ByResultField(benchmarkresult.FieldAverageLatency, opts...)
		}},
		benchmarkresult.FieldConnectionTime: {index: 10, optionFunc: func(opts ...sql.OrderTermOption) benchmark.OrderOption {
			return benchmark.ByResultField(benchmarkresult.FieldConnectionTime, opts...)
		}},
		benchmark.FieldRecordedAt: {index: 11, optionFunc: benchmark.ByRecordedAt},
	}

	// Reverse assign column names to indexes
	cfg.idxToColumn = make(map[int]string, len(cfg.columnToEntOption))
	for columnName, columnConfig := range cfg.columnToEntOption {
		cfg.idxToColumn[columnConfig.index] = columnName
	}

	return cfg
}

func parseOrderByName(cfg *listOrderConfig, columnName string) benchmark.OrderOption {
	if columnName == "" {
		return nil
	}

	// Check for +/- prefix
	direction := sql.OrderAsc()
	if strings.HasPrefix(columnName, "+") {
		columnName = strings.TrimPrefix(columnName, "+")
	} else if strings.HasPrefix(columnName, "-") {
		direction = sql.OrderDesc()
		columnName = strings.TrimPrefix(columnName, "-")
	}

	// Check if column name is valid and if it is return the orderBy function
	if order, ok := cfg.columnToEntOption[columnName]; ok {
		return order.optionFunc(direction)
	}

	return nil
}

func parseOrderByIndex(cfg *listOrderConfig, columnID int) benchmark.OrderOption {
	if cfg == nil {
		return nil
	}

	// Negative columnID means descending order
	orderBy := sql.OrderAsc()
	if columnID < 0 {
		orderBy = sql.OrderDesc()
		columnID *= -1
	}

	// Get column name, if it exists return the orderBy function. Implicitly checks if columnID is valid.
	columnName, ok := cfg.idxToColumn[columnID]
	if !ok {
		return nil
	}

	// Column name is valid, return the orderBy function
	if order, ok := cfg.columnToEntOption[columnName]; ok {
		return order.optionFunc(orderBy)
	}

	return nil
}

func parseOrderBy(cfg *listOrderConfig, columnID string) benchmark.OrderOption {
	if cfg == nil {
		return nil
	}

	// Try to convert columnID to int
	if columnIDInt, err := strconv.Atoi(columnID); err == nil {
		return parseOrderByIndex(cfg, columnIDInt)
	}

	return parseOrderByName(cfg, columnID)
}
