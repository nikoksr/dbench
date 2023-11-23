package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"go.jetpack.io/typeid"

	"github.com/nikoksr/dbench/ent/schema/duration"
	"github.com/nikoksr/dbench/ent/schema/pulid"
	"github.com/nikoksr/dbench/internal/benchmark"
	"github.com/nikoksr/dbench/internal/build"
	"github.com/nikoksr/dbench/internal/models"
	"github.com/nikoksr/dbench/internal/store"
	"github.com/nikoksr/dbench/internal/system"
	"github.com/nikoksr/dbench/internal/ui/styles"
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
	plotCmd := styles.Info.Render(fmt.Sprintf("$ %s plot %s", build.AppName, groupID))
	message := fmt.Sprintf("%s:\n\n  %s", benchCompleteMsg, plotCmd)
	fmt.Printf("%s\n%s\n\n", title, message)
}

func newRunCommand() *cobra.Command {
	var clients []int
	var collectSystemDetails bool
	benchConfig := new(models.BenchmarkConfig)

	cmd := &cobra.Command{
		Use:     "run [OPTIONS]",
		Aliases: []string{"r"},
		GroupID: "commands",
		Short:   "Run an array of benchmarks against a PostgreSQL database",
		Long: `Run a series of benchmarks against a PostgreSQL database. This tool provides
various options to customize the benchmarking process, including client count,
threading, and optional comments.

The '--collect-sysinfo' flag allows users to opt-in to collect detailed system
specifications such as CPU model, RAM size, etc., which are crucial for a
comprehensive analysis of the benchmark results. This data collection is entirely
optional and is designed to enhance the context and accuracy of performance metrics.
No data is transmitted externally, ensuring your privacy and data security.

By understanding the system's specifications, users can gain better insights into
how different hardware configurations impact database performance. This feature is
particularly useful for those looking to optimize database settings or evaluate
performance across various hardware setups.

Use this command to initiate the benchmarking process with options tailored to your
requirements. For detailed information on each flag and its usage, refer to the
options listed below.`,
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
			dbenchDB, err := store.New(ctx, dbenchDSN)
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

			// Collect system details if opted-in
			var systemID *pulid.ID
			if collectSystemDetails {
				fmt.Printf("%s\n", styles.Title.Render("System details"))
				fmt.Printf("%s\n\n", styles.Hint.Render("You've opted in to collect system information for benchmark analysis. To view a preview of the system data we'll collect, run 'dbench doctor --sysinfo'."))

				// Get system details
				fmt.Printf("%s", styles.Text.Render("Collecting system details ... "))

				systemDetails, errs := system.GetDetails()

				// Errors are not fatal, but we still want to inform the user
				if len(errs) > 0 {
					fmt.Printf("\n%s\n", styles.Error.Render("Some errors occurred while collecting system details"))
					for _, err := range errs {
						fmt.Printf("  %s\n", styles.Text.Render(err.Error()))
					}
					fmt.Println()
				} else {
					fmt.Printf("%s\n", styles.Success.Render("✓ Success"))
				}

				// Save system details to database
				systemDetails, err := dbenchDB.SaveSystemDetails(ctx, systemDetails)
				if err != nil {
					return fmt.Errorf("save system details: %w", err)
				}

				systemID = &systemDetails.ID
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
				bench.SystemID = systemID

				// Save benchmark to database
				bench, err = dbenchDB.Save(ctx, bench)
				if err != nil {
					return fmt.Errorf("save benchmark: %w", err)
				}
			}

			// Print benchmark complete message
			printBenchComplete(benchmarkGroupID.String())

			return nil
		},
	}

	// Store flags
	cmd.Flags().StringVarP(&benchConfig.DBName, "db-name", "d", "postgres", "Name of the database")
	cmd.Flags().StringVarP(&benchConfig.Username, "db-user", "U", "postgres", "Username for connecting to the database")
	cmd.Flags().StringVarP(&benchConfig.Host, "db-host", "H", "localhost", "Host of the database")
	cmd.Flags().StringVarP(&benchConfig.Port, "db-port", "p", "5432", "Port of the database")

	// Benchmark flags
	cmd.Flags().StringVar(&benchConfig.Mode, "mode", models.ModeSimple, "Benchmarking mode (simple, thorough)")
	cmd.Flags().IntSliceVar(&clients, "clients", []int{1, 2, 4, 8, 16, 32, 64, 128, 256}, "List of number of clients to benchmark with")
	cmd.Flags().IntVar(&benchConfig.NumThreads, "threads", 1, "Number of threads to use")
	cmd.Flags().StringVarP(&benchConfig.Comment, "comment", "c", "", "Comment to add some optional information to the benchmark")
	cmd.Flags().BoolVar(&collectSystemDetails, "collect-sysinfo", false, "Opt-in to collect detailed system specifications (CPU, RAM, etc.) for benchmark analysis. See help for more")

	return cmd
}
