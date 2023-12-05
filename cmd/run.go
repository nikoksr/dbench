package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"go.jetpack.io/typeid"

	"github.com/nikoksr/dbench/cmd/cobrax"
	"github.com/nikoksr/dbench/ent/schema/duration"
	"github.com/nikoksr/dbench/ent/schema/pulid"
	"github.com/nikoksr/dbench/internal/benchmark"
	"github.com/nikoksr/dbench/internal/build"
	"github.com/nikoksr/dbench/internal/events"
	"github.com/nikoksr/dbench/internal/fs"
	"github.com/nikoksr/dbench/internal/models"
	"github.com/nikoksr/dbench/internal/system"
	"github.com/nikoksr/dbench/internal/ui/printer"
	"github.com/nikoksr/dbench/internal/ui/text"
)

type runOptions struct {
	*globalOptions

	benchConfig         models.BenchmarkConfig
	clients             []int
	collectSystemConfig bool
}

func newRunCommand(globalOpts *globalOptions, connectToDB dbConnector) *cobra.Command {
	opts := &runOptions{
		globalOptions: globalOpts,
	}

	cmd := &cobra.Command{
		Use:                   "run [OPTIONS]",
		Aliases:               []string{"r"},
		GroupID:               "commands",
		Short:                 "Run an array of benchmarks against a PostgreSQL database",
		Long:                  runLongDesc,
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		ValidArgsFunction:     cobra.NoFileCompletions,
		PreRunE:               cobrax.HooksE(pgbenchInstalledHook()),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := connectToDB(cmd.Context(), opts.dataDir, opts.noMigration, fs.OSFileSystem{})
			if err != nil {
				return fmt.Errorf("connect to database: %w", err)
			}

			// Print header
			p := printer.NewPrinter(cmd.OutOrStdout(), 50)
			p.PrintlnTitle("Benchmarking")

			// Prompt for database password
			p.PrintlnSubTitle("Authentication")

			password, canceled, err := getDBPassword(p)
			if err != nil {
				return fmt.Errorf("get database password: %w", err)
			}
			if canceled {
				return nil
			}

			opts.benchConfig.Password = password

			// Generate a new benchmark group id. This is mostly helpful for the analysis of the benchmarks later on.
			benchmarkGroupID, err := typeid.WithPrefix("bmkgrp")
			if err != nil {
				return fmt.Errorf("create benchmark group id: %w", err)
			}

			// Collect system config if opted-in
			var systemConfig *models.SystemConfig
			ctx := cmd.Context()

			if opts.collectSystemConfig {
				p.Spacer(1)
				p.PrintlnSubTitle("System info")
				p.PrintlnHint("You've opted in to collect system information for benchmark analysis. This data collection is entirely optional and is designed to enhance the context and accuracy of performance metrics. No data is transmitted externally, ensuring your privacy and data security.", printer.WithParagraphMode())

				// Get system config
				p.Spacer(1)
				p.PrintInfo(" Collecting system info ... ", printer.WithIndent())

				var errs []error
				systemConfig, errs = system.GetConfig()

				// Errors are not fatal, but we still want to inform the user
				if len(errs) > 0 {
					p.PrintlnError("")
					p.Spacer(1)
					p.PrintlnText(text.WarningsList(errs))
				} else {
					p.PrintlnSuccess("")
				}
			}

			// Subscribe to events of the benchmark run
			events.Subscribe(func(event events.Event) {
				switch event.Type {
				case benchmark.RunCommandRunning:
					p.PrintInfo(fmt.Sprintf(" Executing %s ... ", event.Message))
				}
			})

			// Calculate estimated runtime
			estimatedRuntime := time.Duration(len(opts.clients)) * time.Second * 5 // 5 seconds per set. Clean this up, works for now.

			p.Spacer(2)
			p.PrintlnSubTitle("Running benchmarks")
			p.PrintlnHint(fmt.Sprintf("  Starting! Estimated total runtime %s", estimatedRuntime.String()))
			p.Spacer(1)

			for _, numClients := range opts.clients {
				// Create benchmark configuration
				opts.benchConfig.NumClients = numClients

				// Run benchmark
				benchStart := time.Now()
				bench, err := benchmark.Run(ctx, &opts.benchConfig)
				benchRuntime := time.Since(benchStart)

				if err != nil {
					p.PrintlnError(err.Error())
					return fmt.Errorf("run benchmark: %w", err)
				}

				p.PrintlnSuccess("")

				// Set some meta benchmark values
				bench.Edges.System = systemConfig
				bench.Edges.Result.TotalRuntime = duration.Duration(benchRuntime)
				bench.GroupID = pulid.ID(benchmarkGroupID.String())
				bench.RecordedAt = time.Now().UTC()

				// Save benchmark to database
				_, err = db.Save(ctx, bench)
				if err != nil {
					return fmt.Errorf("save benchmark: %w", err)
				}
			}

			// Print benchmark complete message
			p.Spacer(2)
			p.PrintlnText(" Complete! Run the following command to plot the results:")
			p.Spacer(1)
			p.PrintInfo(fmt.Sprintf("   $ %s plot ", build.AppName))
			p.PrintlnHighlight(benchmarkGroupID.String())
			p.Spacer(2)

			return nil
		},
	}

	// Store flags
	cmd.Flags().StringVarP(&opts.benchConfig.Host, "db-host", "H", "localhost", "Host of the database")
	cmd.Flags().StringVarP(&opts.benchConfig.Port, "db-port", "p", "5432", "Port of the database")
	cmd.Flags().StringVarP(&opts.benchConfig.Username, "db-user", "U", "postgres", "Username for connecting to the database")
	cmd.Flags().StringVarP(&opts.benchConfig.DBName, "db-name", "d", "postgres", "Name of the database")

	// Benchmark flags
	cmd.Flags().StringVar(&opts.benchConfig.Mode, "mode", models.ModeSimple, "Benchmarking mode (simple, thorough)")
	cmd.Flags().StringVarP(&opts.benchConfig.Comment, "comment", "c", "", "Comment to add some optional information to the benchmark")
	cmd.Flags().IntVar(&opts.benchConfig.NumThreads, "threads", 1, "Number of threads to use")
	cmd.Flags().IntSliceVar(&opts.clients, "clients", []int{1, 2, 4, 8, 16, 32, 64, 128, 256}, "List of number of clients to benchmark with")
	cmd.Flags().BoolVar(&opts.collectSystemConfig, "collect-sysinfo", false, "Opt-in to collect detailed system specifications (CPU, RAM, etc.) for benchmark analysis. See help for more")

	cmd.Flags().SortFlags = false

	return cmd
}

var runLongDesc = `Run a series of benchmarks against a PostgreSQL database. This
tool provides various options to customize the benchmarking process, including
client count, threading, and optional comments.

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
options listed below.`
