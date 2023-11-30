package cmd

import (
	"context"
	"fmt"
	"github.com/nikoksr/dbench/internal/fs"
	"os"

	"github.com/spf13/cobra"

	"github.com/nikoksr/dbench/cmd/cobrax"
	"github.com/nikoksr/dbench/internal/database"
	"github.com/nikoksr/dbench/internal/export"
	"github.com/nikoksr/dbench/internal/plot"
	"github.com/nikoksr/dbench/internal/ui/styles"
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
			// Connect to database
			db, err := connectToDB(cmd.Context(), opts.dataDir, fs.OSFileSystem{})
			if err != nil {
				return fmt.Errorf("connect to database: %w", err)
			}

			// Get benchmark-group IDs
			benchmarkGroupIDs := args

			// Cleanup output directory
			if opts.cleanOutputDir {
				fmt.Printf("%s\n", styles.Title.Render("Cleanup"))
				msg := fmt.Sprintf("Cleaning output directory (%s)", opts.outputDir)
				fmt.Printf("%s\t", styles.Info.Render(msg))
				if err := os.RemoveAll(opts.outputDir); err != nil {
					fmt.Printf("%s\n", styles.Error.Render("✗ Failed"))
					return fmt.Errorf("cleanup output directory: %w", err)
				}
				fmt.Printf("%s\n", styles.Success.Render("✓ Success"))
			}

			// Prepare output directory
			if err := prepareDirectory(opts.outputDir); err != nil {
				return fmt.Errorf("prepare output directory: %w", err)
			}

			// Plot benchmarks
			fmt.Printf("%s\n", styles.Title.Render("Plotting"))
			fmt.Printf("%s\n\n", styles.Text.Render("Plotting benchmark-groups"))
			for _, id := range benchmarkGroupIDs {
				fmt.Printf("  %s\t", styles.Info.Render("Plotting "+id))
				if err := plotBenchmarks(cmd.Context(), db, id, opts.outputDir); err != nil {
					fmt.Printf("  %s\n", styles.Error.Render("✗ Failed"))
					return fmt.Errorf("plot benchmark-group %q: %w", id, err)
				}
				fmt.Printf("  %s\n", styles.Success.Render("✓ Success"))
			}

			title := styles.Title.Render("Results")
			message := fmt.Sprintf("%s:\n\n $ cd %s", styles.Text.Render("Plotting done! Check out the results in the output directory"), opts.outputDir)
			fmt.Printf("%s\n%s\n\n", title, message)

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.outputDir, "output", "o", "dbench/plots", "Output directory for plots")
	cmd.Flags().BoolVarP(&opts.cleanOutputDir, "clean", "c", false, "Cleanup output directory before plotting")

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

	dataFile, err := export.ToCSV(benchmarks, "")
	if err != nil {
		return fmt.Errorf("export benchmarks to CSV: %w", err)
	}

	// Generate plots using gnuplot
	if err := plot.Plot(ctx, dataFile, outputDir); err != nil {
		return fmt.Errorf("plot benchmarks: %w", err)
	}

	return nil
}
