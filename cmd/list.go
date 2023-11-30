package cmd

import (
	"fmt"
	"github.com/nikoksr/dbench/internal/fs"
	"github.com/nikoksr/dbench/internal/ui/styles"
	"github.com/nikoksr/dbench/internal/ui/text"
	"math"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/spf13/cobra"

	"github.com/nikoksr/dbench/ent"
	"github.com/nikoksr/dbench/internal/database"
	"github.com/nikoksr/dbench/internal/ui"
)

const (
	defaultPage    = 1
	defaultPerPage = 10
)

type listOptions struct {
	*globalOptions

	sort                      []string
	page, perPage, totalPages int
	showAll                   bool
}

func (opts *listOptions) validate(recordsCount int) []error {
	var errs []error

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
		errs = append(errs, fmt.Errorf("page number out of range -> showing page %d of %d", opts.page, opts.totalPages))
	}

	// Check if per page number is out of range
	if opts.perPage > recordsCount {
		opts.page = defaultPage
		opts.perPage = recordsCount
		errs = append(errs, fmt.Errorf("per page number out of range -> showing page %d of %d", opts.page, opts.totalPages))
	}

	return errs
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
			// Connect to database
			db, err := connectToDB(cmd.Context(), opts.dataDir, fs.OSFileSystem{})
			if err != nil {
				return fmt.Errorf("connect to database: %w", err)
			}

			// Get the total benchmark count and calculate the total number of pages for pagination
			benchmarksCount, err := db.Count(cmd.Context())
			if err != nil {
				return fmt.Errorf("count benchmarks: %w", err)
			}

			// Validate options
			errs := opts.validate(benchmarksCount)
			if len(errs) > 0 {
				fmt.Println(text.WarningsList(errs))
			}

			// Query benchmarks
			orderByFunc := func(query *ent.BenchmarkQuery) *ent.BenchmarkQuery {
				for _, order := range opts.sort {
					orderFunc := parseOrderBy(order)
					query = query.Order(orderFunc)
				}

				return query
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
			fmt.Print(tableStr)

			// Render pagination info
			fmt.Println(styles.Hint.Render(fmt.Sprintf(" Showing page %d of %d\n", opts.page, opts.totalPages)))

			return nil
		},
	}

	// Flags
	cmd.Flags().StringSliceVar(&opts.sort, "sort", []string{"id"}, "Sort benchmarks columns (+/- prefix for ascending/descending)")
	cmd.Flags().IntVar(&opts.page, "page", defaultPage, "Page number")
	cmd.Flags().IntVar(&opts.perPage, "per-page", defaultPerPage, "Number of benchmarks per page")
	cmd.Flags().BoolVar(&opts.showAll, "all", false, "Show all benchmarks")

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
