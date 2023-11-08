package cmd

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/spf13/cobra"
	typeid "go.jetpack.io/typeid/typed"

	"github.com/nikoksr/go-pgbench/ent"
	"github.com/nikoksr/go-pgbench/ent/schema/duration"
	"github.com/nikoksr/go-pgbench/ent/schema/pulid"
	"github.com/nikoksr/go-pgbench/pkg/benchmark"
	"github.com/nikoksr/go-pgbench/pkg/buildinfo"
	"github.com/nikoksr/go-pgbench/pkg/database"
	"github.com/nikoksr/go-pgbench/pkg/export"
	"github.com/nikoksr/go-pgbench/pkg/models"
	"github.com/nikoksr/go-pgbench/pkg/ui"
)

func newBenchCommand() *cobra.Command {
	// Configuration for the benchmark command
	var benchConfig models.BenchmarkConfig

	cmd := &cobra.Command{
		Use:               "bench [command]",
		Short:             "Manage and run your database benchmarks.",
		SilenceUsage:      true,
		SilenceErrors:     true,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
	}

	// Define flags for database connection parameters
	cmd.PersistentFlags().StringVarP(&benchConfig.DBName, "dbname", "d", "postgres", "Name of the database")
	cmd.PersistentFlags().StringVarP(&benchConfig.Username, "username", "U", "postgres", "Username for connecting to the database")
	cmd.PersistentFlags().StringVarP(&benchConfig.Host, "host", "H", "localhost", "Host of the database")
	cmd.PersistentFlags().StringVarP(&benchConfig.Port, "port", "p", "5432", "Port of the database")

	// Bench subcommands
	cmd.AddCommand(
		newBenchInitCommand(&benchConfig),
		newBenchRunCommand(&benchConfig),
		newBenchListCommand(&benchConfig),
		newBenchExportCommand(&benchConfig),
	)

	return cmd
}

func printBenchStarting(numSets int) {
	// Calculate estimated runtime
	estimatedRuntime := time.Duration(numSets) * time.Second * 5 // 5 seconds per set. Clean this up, works for now.

	title := fmt.Sprintf(" Starting benchmarks! Estimated total runtime: %s", estimatedRuntime)

	fmt.Printf("%s\n\n", ui.TextBox(title))
}

func printBenchComplete(groupID string) {
	title := fmt.Sprintf(" Benchmarks complete! Run the following command to plot the results:\n\n\t%s plot %s", buildinfo.AppName, groupID)

	fmt.Printf("\n%s\n\n", ui.TextBox(title))
}

func newBenchRunCommand(benchConfig *models.BenchmarkConfig) *cobra.Command {
	var clients []int

	cmd := &cobra.Command{
		Use:               "run",
		Short:             "Run an array of benchmarks against a PostgreSQL database",
		SilenceUsage:      true,
		SilenceErrors:     true,
		ValidArgsFunction: cobra.NoFileCompletions,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return validateRequirements()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Open database connection
			resultsDB, err := database.NewEntDatabase(ctx, resultsDSN)
			if err != nil {
				return fmt.Errorf("create results database: %w", err)
			}
			defer resultsDB.Close()

			// Generate a new result group id. This is mostly helpful for the analysis of the results.
			resultGroupID, err := typeid.New[models.ResultGroupID]()
			if err != nil {
				return fmt.Errorf("create result group id: %w", err)
			}

			// Run benchmark for different client counts
			printBenchStarting(len(clients))

			for _, numClients := range clients {
				// Create benchmark configuration
				benchConfig.NumClients = numClients

				// Run benchmark
				benchStart := time.Now()
				result, err := benchmark.Run(benchConfig)
				benchRuntime := time.Since(benchStart)

				if err != nil {
					return fmt.Errorf("run benchmark: %w", err)
				}

				// Set some meta result values
				result.TotalRuntime = duration.Duration(benchRuntime)
				result.GroupID = pulid.ID(resultGroupID.String())

				// Save result to database
				if err := resultsDB.SaveResult(ctx, result); err != nil {
					return fmt.Errorf("save result: %w", err)
				}
			}

			// Print benchmark complete message
			printBenchComplete(resultGroupID.String())

			return nil
		},
	}

	cmd.Flags().StringVar(&benchConfig.Mode, "mode", models.ModeSimple, "Benchmarking mode (simple, thorough)")
	cmd.Flags().IntSliceVar(&clients, "clients", []int{1, 2, 4, 8, 16, 32, 64, 128, 256}, "List of number of clients to benchmark with")
	cmd.Flags().IntVar(&benchConfig.NumThreads, "threads", 1, "Number of threads to use")

	return cmd
}

func newBenchInitCommand(benchConfig *models.BenchmarkConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a target database using pgbench",
		Long: `This command initializes a target database using pgbench. You can either use this command to create a new
database or run pgbench manually. If you want to run pgbench manually, this is what an example command would look like:

	pgbench -i -s 1 -F 100 -U postgres -h localhost -p 5432 postgres

For consistency reasons, it is HIGHLY recommended to use this command instead of running pgbench manually.

For more information, see the official documentation:
https://www.postgresql.org/docs/current/pgbench.html
`,
		SilenceUsage:      true,
		SilenceErrors:     true,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			return benchmark.Init(benchConfig)
		},
	}

	cmd.Flags().IntVar(&benchConfig.FillFactor, "fill", 100, "Fill factor for the database (10-100)")
	cmd.Flags().IntVar(&benchConfig.ScaleFactor, "scale", 1, "Scale factor for the database (1-1000)")

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

func newBenchListCommand(benchConfig *models.BenchmarkConfig) *cobra.Command {
	var sort []string

	cmd := &cobra.Command{
		Use:               "list",
		Aliases:           []string{"ls"},
		Short:             "List all benchmark results",
		SilenceUsage:      true,
		SilenceErrors:     true,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Open database connection
			ctx := cmd.Context()
			resultsDB, err := database.NewEntDatabase(ctx, resultsDSN)
			if err != nil {
				return fmt.Errorf("create results database: %w", err)
			}
			defer resultsDB.Close()

			// Query results
			orderByFunc := func(query *ent.ResultQuery) *ent.ResultQuery {
				for _, order := range sort {
					orderFunc := parseOrderBy(order)
					query = query.Order(orderFunc)
				}

				return query
			}

			results, err := resultsDB.FetchResults(ctx, database.WithOrderBy(orderByFunc))
			if err != nil {
				return fmt.Errorf("fetch results: %w", err)
			}

			// Render results
			renderer := ui.NewResultTableRenderer()
			tableStr := renderer.Render(results)
			fmt.Println(tableStr)

			return nil
		},
	}

	// Flags
	cmd.Flags().StringSliceVar(&sort, "sort", []string{"id"}, "Sort results columns (+/- prefix for ascending/descending)")

	return cmd
}

func newBenchExportCommand(benchConfig *models.BenchmarkConfig) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:               "export",
		Short:             "Export all benchmark results to a format of your choice",
		SilenceUsage:      true,
		SilenceErrors:     true,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Open database connection
			ctx := cmd.Context()
			resultsDB, err := database.NewEntDatabase(ctx, resultsDSN)
			if err != nil {
				return fmt.Errorf("create results database: %w", err)
			}
			defer resultsDB.Close()

			// Query results
			results, err := resultsDB.FetchResults(ctx)
			if err != nil {
				return fmt.Errorf("fetch results: %w", err)
			}

			// Export results
			switch format {
			case "csv":
				err = export.ToCSV(results, "results.csv")
			case "json":
				err = export.ToJSON(results, "results.json")
			case "gnuplot":
				err = export.ToGnuplot(results, "results.dat")
			default:
				return fmt.Errorf("unknown export format: %s", format)
			}

			if err != nil {
				return fmt.Errorf("export results: %w", err)
			}

			return nil
		},
	}

	// Flags
	cmd.Flags().StringVar(&format, "format", "csv", "Format to export results to")

	return cmd
}
