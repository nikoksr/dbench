package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/spf13/cobra"
	"go.jetpack.io/typeid"

	"github.com/nikoksr/dbench/ent"
	"github.com/nikoksr/dbench/ent/schema/duration"
	"github.com/nikoksr/dbench/ent/schema/pulid"
	"github.com/nikoksr/dbench/pkg/benchmark"
	"github.com/nikoksr/dbench/pkg/buildinfo"
	"github.com/nikoksr/dbench/pkg/database"
	"github.com/nikoksr/dbench/pkg/export"
	"github.com/nikoksr/dbench/pkg/models"
	"github.com/nikoksr/dbench/pkg/styles"
	"github.com/nikoksr/dbench/pkg/ui"
)

var (
	errNoPassword = errors.New(`No password provided. You can either enter a password or set the PGPASSWORD environment variable:

	# Example
	export PGPASSWORD=supersecret

For more information, see the official documentation:
https://www.postgresql.org/docs/current/libpq-envars.html
`)

	errPgbenchNotInstalled = errors.New(`pgbench is required to run the application. It can be installed with the following command:

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

func printBenchStarting(numSets int) {
	// Calculate estimated runtime
	estimatedRuntime := time.Duration(numSets) * time.Second * 5 // 5 seconds per set. Clean this up, works for now.

	fmt.Printf("%s\n%s: %s\n\n",
		styles.Title.Render("Benchmark"),
		styles.Text.Render("Starting benchmarks! Estimated total runtime"),
		styles.Info.Render(estimatedRuntime.String()),
	)
}

func printBenchComplete(groupID string) {
	title := styles.Title.Render("Results")
	benchCompleteMsg := styles.Text.Render("Benchmarks complete! Run the following command to plot the results")
	plotCmd := styles.Info.Render(fmt.Sprintf("$ %s plot %s", buildinfo.AppName, groupID))
	message := fmt.Sprintf("%s:\n\n  %s", benchCompleteMsg, plotCmd)
	fmt.Printf("%s\n%s\n\n", title, message)
}

func getDBPassword() (string, bool, error) {
	// Check if PGPASSWORD is set
	passwd := os.Getenv("PGPASSWORD")

	if passwd != "" {
		fmt.Printf("%s\n\n", styles.Hint.Render("Detected PGPASSWORD - leave the following prompt empty to use it."))
	}

	// Prompt for password
	prompt := ui.NewPrompt("Enter database password:", "Password", true)
	if err := prompt.Render(); err != nil {
		return "", false, err
	}

	// If the user canceled the prompt, return. Signal that we don't want to continue.
	if prompt.WasCanceled() {
		return "", true, nil
	}

	// If the user entered a password, return it
	if prompt.Value() != "" {
		return prompt.Value(), false, nil
	}

	// No password entered, if PGPASSWORD is set, return it
	if passwd != "" {
		return passwd, false, nil
	}

	// No password entered and PGPASSWORD is not set, return an error
	return "", false, errNoPassword
}

func newBenchRunCommand(benchConfig *models.BenchmarkConfig) *cobra.Command {
	var clients []int

	if benchConfig == nil {
		benchConfig = &models.BenchmarkConfig{}
	}

	cmd := &cobra.Command{
		Use:                   "run [OPTIONS]",
		Aliases:               []string{"r"},
		GroupID:               "commands",
		Short:                 "Run an array of benchmarks against a PostgreSQL database",
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		ValidArgsFunction:     cobra.NoFileCompletions,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Check if pgbench is installed
			if !isToolInPath("pgbench") {
				return errPgbenchNotInstalled
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("%s\n", styles.Title.Render("Authentication"))

			// Open database connection
			ctx := cmd.Context()
			dbenchDB, err := database.NewEntDatabase(ctx, dbenchDSN)
			if err != nil {
				return fmt.Errorf("create dbench database: %w", err)
			}
			defer dbenchDB.Close()

			// Prompt for password
			password, canceled, err := getDBPassword()
			if err != nil {
				return fmt.Errorf("get database password: %w", err)
			}
			if canceled {
				return nil
			}

			benchConfig.Password = password

			// Generate a new benchmark group id. This is mostly helpful for the analysis of the benchmarks.
			benchmarkGroupID, err := typeid.WithPrefix("bmkgrp")
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
		Use:     "init [OPTIONS]",
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
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		ValidArgsFunction:     cobra.NoFileCompletions,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Check if pgbench is installed
			if !isToolInPath("pgbench") {
				return errPgbenchNotInstalled
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("%s\n", styles.Title.Render("Authentication"))

			// Prompt for password
			password, canceled, err := getDBPassword()
			if err != nil {
				return fmt.Errorf("get database password: %w", err)
			}
			if canceled {
				return nil
			}

			benchConfig.Password = password

			// Initialize database
			fmt.Printf("%s\n", styles.Title.Render("Initialization"))

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
		Use:                   "export [OPTIONS]",
		Aliases:               []string{"e"},
		GroupID:               "commands",
		Short:                 "Export all data to a format of your choice",
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		Args:                  cobra.NoArgs,
		ValidArgsFunction:     cobra.NoFileCompletions,
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
			var dataFile string
			switch format {
			case "csv":
				dataFile, err = export.ToCSV(benchmarks, generateExportFileName("csv"))
			case "json":
				dataFile, err = export.ToJSON(benchmarks, generateExportFileName("json"))
			default:
				return fmt.Errorf("unknown export format: %s", format)
			}

			if err != nil {
				return fmt.Errorf("export benchmarks: %w", err)
			}

			fmt.Printf("\nExported data to %s\n\n", styles.Success.Render(dataFile))

			return nil
		},
	}

	// Flags
	cmd.Flags().StringVar(&format, "format", "csv", "Format to export benchmarks to (csv, json)")

	return cmd
}

func newBenchRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "remove ID [ID...]",
		Aliases:               []string{"r", "rm"},
		GroupID:               "commands",
		Short:                 "Remove benchmarks from the database",
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		Args:                  cobra.MinimumNArgs(1),
		ValidArgsFunction:     cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Open database connection
			ctx := cmd.Context()
			dbenchDB, err := database.NewEntDatabase(ctx, dbenchDSN)
			if err != nil {
				return fmt.Errorf("create dbench database: %w", err)
			}
			defer dbenchDB.Close()

			// Convert and validate ids
			fmt.Printf("%s\n", styles.Title.Render("Validation"))
			fmt.Printf("%s\t", styles.Text.Render("Validating ids..."))

			var ids, groupIDs []string
			for _, arg := range args {
				// Try to convert id to typeid
				id, err := typeid.FromString(arg)
				if err != nil {
					fmt.Println(styles.Error.Render("✗ Failed\n"))
					return fmt.Errorf("convert id to typeid: %w", err)
				}

				if id.Prefix() == "bmkgrp" {
					groupIDs = append(groupIDs, id.String())
				} else {
					ids = append(ids, id.String())
				}
			}

			fmt.Println(styles.Success.Render("✓ Success"))

			// Remove benchmarks

			if len(ids) > 0 {
				fmt.Printf("%s\n", styles.Title.Render("Remove benchmarks"))
				msg := fmt.Sprintf("Removing %d benchmark(s)", len(ids))
				fmt.Printf("%s\t", styles.Text.Render(msg))

				if err := dbenchDB.RemoveBenchmarksByIDs(ctx, ids); err != nil {
					fmt.Println(styles.Error.Render("✗ Failed\n"))
					return fmt.Errorf("remove benchmarks by ids: %w", err)
				}

				fmt.Println(styles.Success.Render("✓ Success"))
			}

			// Remove benchmark groups

			if len(groupIDs) > 0 {
				fmt.Printf("%s\n", styles.Title.Render("Remove benchmark groups"))
				msg := fmt.Sprintf("Removing %d benchmark-group(s)", len(groupIDs))
				fmt.Printf("%s\t", styles.Text.Render(msg))

				if err := dbenchDB.RemoveBenchmarksByGroupIDs(ctx, groupIDs); err != nil {
					fmt.Println(styles.Error.Render("✗ Failed\n"))
					return fmt.Errorf("remove benchmarks by group ids: %w", err)
				}

				fmt.Println(styles.Success.Render("✓ Success"))
			}

			fmt.Println()

			return nil
		},
	}

	return cmd
}
