package cmd

import (
	"fmt"
	"github.com/nikoksr/dbench/ent/schema/duration"
	"github.com/nikoksr/dbench/ent/schema/pulid"
	"github.com/nikoksr/dbench/internal/benchmark"
	"github.com/nikoksr/dbench/internal/build"
	"github.com/nikoksr/dbench/internal/database"
	"github.com/nikoksr/dbench/internal/models"
	"github.com/nikoksr/dbench/internal/styles"
	"github.com/spf13/cobra"
	"go.jetpack.io/typeid"
	"time"
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
	benchConfig := new(models.BenchmarkConfig)

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
