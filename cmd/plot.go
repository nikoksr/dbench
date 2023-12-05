package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/nikoksr/dbench/cmd/cobrax"
	"github.com/nikoksr/dbench/internal/build"
	"github.com/nikoksr/dbench/internal/database"
	"github.com/nikoksr/dbench/internal/fs"
	"github.com/nikoksr/dbench/internal/plot"
	"github.com/nikoksr/dbench/internal/portability/converter"
	"github.com/nikoksr/dbench/internal/portability/exporter"
	"github.com/nikoksr/dbench/internal/ui/printer"
)

type plotOptions struct {
	*globalOptions

	outputDir      string
	cleanOutputDir bool
}

func newPlotCommand(globalOpts *globalOptions, connectToDB dbConnector) *cobra.Command {
	opts := &plotOptions{
		globalOptions: globalOpts,
	}

	cmd := &cobra.Command{
		Use:                   "plot [OPTIONS] BENCHMARK-GROUP-ID [BENCHMARK-GROUP-ID...]",
		Aliases:               []string{"p"},
		GroupID:               "commands",
		Short:                 "Plot benchmark results by benchmark-groups",
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		Args:                  cobra.MinimumNArgs(1),
		PreRunE:               cobrax.HooksE(gnuplotInstalledHook()),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := connectToDB(cmd.Context(), opts.dataDir, opts.noMigration, fs.OSFileSystem{})
			if err != nil {
				return fmt.Errorf("connect to database: %w", err)
			}

			// Print header
			p := printer.NewPrinter(cmd.OutOrStdout(), 50)
			p.PrintlnTitle("Plotting")
			p.PrintlnSubTitle("Preparation")

			// Get benchmark-group IDs
			benchmarkGroupIDs := args

			// Cleanup output directory
			if opts.cleanOutputDir {
				p.PrintInfo(" Cleaning output directory ... ", printer.WithIndent())
				if err := os.RemoveAll(opts.outputDir); err != nil {
					p.PrintlnError(err.Error())
					return fmt.Errorf("cleanup output directory: %w", err)
				}
				p.PrintlnSuccess("")
			}

			// Prepare output directory
			p.PrintInfo(" Creating plots directory ... ", printer.WithIndent())
			if err := os.MkdirAll(opts.outputDir, 0o755); err != nil {
				p.PrintlnError(err.Error())
				return fmt.Errorf("create plots directory: %w", err)
			}
			p.PrintlnSuccess("")

			// Plot benchmarks
			p.Spacer(1)
			p.PrintlnSubTitle("Plotting")
			for _, id := range benchmarkGroupIDs {
				p.PrintInfo(" Plotting "+id+" ... ", printer.WithIndent())
				if err := plotBenchmarks(cmd.Context(), db, id, opts.outputDir); err != nil {
					p.PrintlnError(err.Error())
					return fmt.Errorf("plot benchmark-group %q: %w", id, err)
				}
				p.PrintlnSuccess("")
			}

			p.Spacer(2)
			p.PrintText(" Complete! Saved plots to ")
			p.PrintlnHighlight(opts.outputDir)
			p.Spacer(2)

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.outputDir, "output", "o", "dbench/plots", "Output directory for plots")
	cmd.Flags().BoolVarP(&opts.cleanOutputDir, "clean", "c", false, "Cleanup output directory before plotting")

	cmd.Flags().SortFlags = false

	return cmd
}

func plotBenchmarks(ctx context.Context, db database.Store, id, outputDir string) error {
	benchmarks, err := db.FetchByGroupIDs(ctx, []string{id})
	if err != nil {
		return fmt.Errorf("fetch benchmarks by benchmark-group ID: %w", err)
	}

	if len(benchmarks) == 0 {
		return fmt.Errorf("no benchmarks found for benchmark-group %q", id)
	}

	// Create temp file for CSV data
	file, err := os.CreateTemp("", build.AppName+"-*.csv")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	// Save temp file name for plotting and cleanup
	fileName := file.Name()
	defer func() {
		_ = file.Close()
		_ = os.Remove(fileName)
	}()

	// Convert benchmarks to CSV export format
	exportable := converter.BenchmarksToCSV(benchmarks)

	// Export benchmarks to CSV
	if err := exporter.ToCSV(file, exportable); err != nil {
		return fmt.Errorf("export benchmarks to CSV: %w", err)
	}

	// Generate plots using gnuplot
	if err := plot.Plot(ctx, fileName, outputDir); err != nil {
		return fmt.Errorf("plot benchmarks: %w", err)
	}

	return nil
}
