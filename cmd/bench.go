package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/spf13/cobra"
	typeid "go.jetpack.io/typeid/typed"

	"github.com/nikoksr/dbench/ent"
	"github.com/nikoksr/dbench/ent/schema/duration"
	"github.com/nikoksr/dbench/ent/schema/pulid"
	"github.com/nikoksr/dbench/pkg/benchmark"
	"github.com/nikoksr/dbench/pkg/buildinfo"
	"github.com/nikoksr/dbench/pkg/database"
	"github.com/nikoksr/dbench/pkg/export"
	"github.com/nikoksr/dbench/pkg/models"
	"github.com/nikoksr/dbench/pkg/ui"
)

var (
	noPasswordErr = fmt.Errorf(`Environment variable PGPASSWORD must be set to the password of the specified PostgreSQL user.

	# Example
	export PGPASSWORD=supersecret

For more information, see the official documentation:
https://www.postgresql.org/docs/current/libpq-envars.html
`)

	pgbenchNotInstalledErr = fmt.Errorf(`pgbench is required to run the application. It can be installed with the following command:

	# Arch
	sudo pacman -S postgresql

	# Debian / Ubuntu
	sudo apt install postgresql-client

	# macOS
	brew install postgresql

For more information, see the official documentation:
https://www.postgresql.org/docs/current/pgbench.html
`)
)

func newBenchCommand() *cobra.Command {
	// Configuration for the benchmark command
	var benchConfig models.BenchmarkConfig

	cmd := &cobra.Command{
		Use:               "bench [command]",
		Aliases:           []string{"b"},
		GroupID:           "commands",
		Short:             "Manage and run your database benchmarks",
		SilenceUsage:      true,
		SilenceErrors:     true,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		Hidden:            true,
		Deprecated:        "all subcommands have been moved to the root command. Use run, init, list and export directly instead.",
	}

	cmd.AddGroup(&cobra.Group{
		ID:    "commands",
		Title: "Commands",
	})

	// Bench subcommands
	cmd.AddCommand(
		newBenchInitCommand(&benchConfig),
		newBenchRunCommand(&benchConfig),
		newBenchListCommand(),
		newBenchExportCommand(),
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

	if benchConfig == nil {
		benchConfig = &models.BenchmarkConfig{}
	}

	cmd := &cobra.Command{
		Use:               "run",
		Aliases:           []string{"r"},
		GroupID:           "commands",
		Short:             "Run an array of benchmarks against a PostgreSQL database",
		SilenceUsage:      true,
		SilenceErrors:     true,
		ValidArgsFunction: cobra.NoFileCompletions,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Check if PGPASSWORD is set
			if os.Getenv("PGPASSWORD") == "" {
				return noPasswordErr
			}

			// Check if pgbench is installed
			if _, err := exec.LookPath("pgbench"); err != nil {
				return pgbenchNotInstalledErr
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Open database connection
			dbenchDB, err := database.NewEntDatabase(ctx, dbenchDSN)
			if err != nil {
				return fmt.Errorf("create dbench database: %w", err)
			}
			defer dbenchDB.Close()

			// Generate a new benchmark group id. This is mostly helpful for the analysis of the benchmarks.
			benchmarkGroupID, err := typeid.New[models.BenchmarkGroupID]()
			if err != nil {
				return fmt.Errorf("create benchmark group id: %w", err)
			}

			// Run benchmark for different client counts
			printBenchStarting(len(clients))

			for _, numClients := range clients {
				// Create benchmark configuration
				benchConfig.NumClients = numClients

				// Run benchmark
				benchStart := time.Now()
				bench, err := benchmark.Run(ctx, benchConfig)
				benchRuntime := time.Since(benchStart)

				if err != nil {
					return fmt.Errorf("run benchmark: %w", err)
				}

				// Set some meta benchmark values
				bench.Edges.Result.TotalRuntime = duration.Duration(benchRuntime)
				bench.GroupID = pulid.ID(benchmarkGroupID.String())

				// Save benchmark to database
				if err := dbenchDB.SaveBenchmark(ctx, bench); err != nil {
					return fmt.Errorf("save benchmark: %w", err)
				}
			}

			// Print benchmark complete message
			printBenchComplete(benchmarkGroupID.String())

			return nil
		},
	}

	// Database flags
	cmd.Flags().StringVarP(&benchConfig.DBName, "dbname", "d", "postgres", "Name of the database")
	cmd.Flags().StringVarP(&benchConfig.Username, "username", "U", "postgres", "Username for connecting to the database")
	cmd.Flags().StringVarP(&benchConfig.Host, "host", "H", "localhost", "Host of the database")
	cmd.Flags().StringVarP(&benchConfig.Port, "port", "p", "5432", "Port of the database")

	// Benchmark flags
	cmd.Flags().StringVar(&benchConfig.Mode, "mode", models.ModeSimple, "Benchmarking mode (simple, thorough)")
	cmd.Flags().IntSliceVar(&clients, "clients", []int{1, 2, 4, 8, 16, 32, 64, 128, 256}, "List of number of clients to benchmark with")
	cmd.Flags().IntVar(&benchConfig.NumThreads, "threads", 1, "Number of threads to use")
	cmd.Flags().StringVarP(&benchConfig.Comment, "comment", "c", "", "Comment to add some optional information to the benchmark")

	return cmd
}

func newBenchInitCommand(benchConfig *models.BenchmarkConfig) *cobra.Command {
	if benchConfig == nil {
		benchConfig = &models.BenchmarkConfig{}
	}

	cmd := &cobra.Command{
		Use:     "init",
		Aliases: []string{"i"},
		GroupID: "commands",
		Short:   "Initialize a target database using pgbench",
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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Check if PGPASSWORD is set
			if os.Getenv("PGPASSWORD") == "" {
				return noPasswordErr
			}

			// Check if pgbench is installed
			if _, err := exec.LookPath("pgbench"); err != nil {
				return pgbenchNotInstalledErr
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return benchmark.Init(benchConfig)
		},
	}

	// Database flags
	cmd.Flags().StringVarP(&benchConfig.DBName, "dbname", "d", "postgres", "Name of the database")
	cmd.Flags().StringVarP(&benchConfig.Username, "username", "U", "postgres", "Username for connecting to the database")
	cmd.Flags().StringVarP(&benchConfig.Host, "host", "H", "localhost", "Host of the database")
	cmd.Flags().StringVarP(&benchConfig.Port, "port", "p", "5432", "Port of the database")

	// Init flags
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

func newBenchListCommand() *cobra.Command {
	var sort []string

	cmd := &cobra.Command{
		Use:               "list",
		Aliases:           []string{"l", "ls"},
		GroupID:           "commands",
		Short:             "List previously run benchmarks",
		SilenceUsage:      true,
		SilenceErrors:     true,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Open database connection
			ctx := cmd.Context()
			dbenchDB, err := database.NewEntDatabase(ctx, dbenchDSN)
			if err != nil {
				return fmt.Errorf("create dbench database: %w", err)
			}
			defer dbenchDB.Close()

			// Query benchmarks
			orderByFunc := func(query *ent.BenchmarkQuery) *ent.BenchmarkQuery {
				for _, order := range sort {
					orderFunc := parseOrderBy(order)
					query = query.Order(orderFunc)
				}

				return query
			}

			benchmarks, err := dbenchDB.FetchBenchmarks(ctx, database.WithOrderBy(orderByFunc))
			if err != nil {
				return fmt.Errorf("fetch benchmarks: %w", err)
			}

			// Render benchmarks
			renderer := ui.NewBenchmarksTableRenderer()
			tableStr := renderer.Render(benchmarks)
			fmt.Println(tableStr)

			return nil
		},
	}

	// Flags
	cmd.Flags().StringSliceVar(&sort, "sort", []string{"id"}, "Sort benchmarks columns (+/- prefix for ascending/descending)")

	return cmd
}

func newBenchExportCommand() *cobra.Command {
	var format string

	generateExportFileName := func(fileFormat string) string {
		localTime := time.Now().Local().Format("2006-01-02_15-04-05")
		return fmt.Sprintf("%s-export_%s.%s", buildinfo.AppName, localTime, fileFormat)
	}

	cmd := &cobra.Command{
		Use:               "export",
		Aliases:           []string{"e"},
		GroupID:           "commands",
		Short:             "Export all benchmarks to a format of your choice",
		SilenceUsage:      true,
		SilenceErrors:     true,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Open database connection
			ctx := cmd.Context()
			dbenchDB, err := database.NewEntDatabase(ctx, dbenchDSN)
			if err != nil {
				return fmt.Errorf("create dbench database: %w", err)
			}
			defer dbenchDB.Close()

			// Query benchmarks
			benchmarks, err := dbenchDB.FetchBenchmarks(ctx)
			if err != nil {
				return fmt.Errorf("fetch benchmarks: %w", err)
			}

			if len(benchmarks) == 0 {
				return fmt.Errorf("no benchmarks found")
			}

			// Export benchmarks
			switch format {
			case "csv":
				err = export.ToCSV(benchmarks, generateExportFileName("csv"))
			case "json":
				err = export.ToJSON(benchmarks, generateExportFileName("json"))
			default:
				return fmt.Errorf("unknown export format: %s", format)
			}

			if err != nil {
				return fmt.Errorf("export benchmarks: %w", err)
			}

			return nil
		},
	}

	// Flags
	cmd.Flags().StringVar(&format, "format", "csv", "Format to export benchmarks to (csv, json)")

	return cmd
}
