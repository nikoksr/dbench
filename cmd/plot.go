package cmd

import (
	"context"
	"fmt"
	"github.com/nikoksr/dbench/internal/ui/styles"
	"os"

	"github.com/spf13/cobra"

	"github.com/nikoksr/dbench/internal/export"
	"github.com/nikoksr/dbench/internal/plot"
	"github.com/nikoksr/dbench/internal/store"
)

func newPlotCommand() *cobra.Command {
	var (
		outputDir      string
		cleanOutputDir bool
	)

	cmd := &cobra.Command{
		Use:                   "plot [OPTIONS] BENCHMARK-GROUP-ID [BENCHMARK-GROUP-ID...]",
		Aliases:               []string{"p"},
		GroupID:               "commands",
		Short:                 "Plot benchmark results by benchmark-groups",
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		Args:                  cobra.MinimumNArgs(1),
		ValidArgsFunction:     cobra.NoFileCompletions,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Check if gnuplot is installed
			if !isToolInPath("gnuplot") {
				return gnuPlotNotInstalledErr
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get benchmark-group IDs
			benchmarkGroupIDs := args

			// Cleanup output directory
			if cleanOutputDir {
				fmt.Printf("%s\n", styles.Title.Render("Cleanup"))
				msg := fmt.Sprintf("Cleaning output directory (%s)", outputDir)
				fmt.Printf("%s\t", styles.Info.Render(msg))
				if err := os.RemoveAll(outputDir); err != nil {
					fmt.Printf("%s\n", styles.Error.Render("✗ Failed"))
					return fmt.Errorf("cleanup output directory: %w", err)
				}
				fmt.Printf("%s\n", styles.Success.Render("✓ Success"))
			}

			// Prepare output directory
			if err := prepareDirectory(outputDir); err != nil {
				return fmt.Errorf("prepare output directory: %w", err)
			}

			// Plot benchmarks
			fmt.Printf("%s\n", styles.Title.Render("Plotting"))
			fmt.Printf("%s\n\n", styles.Text.Render("Plotting benchmark-groups"))
			for _, id := range benchmarkGroupIDs {
				fmt.Printf("  %s\t", styles.Info.Render("Plotting "+id))
				if err := plotBenchmarks(cmd.Context(), id, outputDir); err != nil {
					fmt.Printf("  %s\n", styles.Error.Render("✗ Failed"))
					return fmt.Errorf("plot benchmark-group %q: %w", id, err)
				}
				fmt.Printf("  %s\n", styles.Success.Render("✓ Success"))
			}

			title := styles.Title.Render("Results")
			message := fmt.Sprintf("%s:\n\n $ cd %s", styles.Text.Render("Plotting done! Check out the results in the output directory"), outputDir)
			fmt.Printf("%s\n%s\n\n", title, message)

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "dbench/plots", "Output directory for plots")
	cmd.Flags().BoolVarP(&cleanOutputDir, "clean", "c", false, "Cleanup output directory before plotting")

	return cmd
}

func plotBenchmarks(ctx context.Context, id, outputDir string) error {
	dbenchDB, err := store.New(ctx, dbenchDSN)
	if err != nil {
		return fmt.Errorf("create dbench database: %w", err)
	}
	defer dbenchDB.Close()

	benchmarks, err := dbenchDB.FetchByGroupIDs(ctx, []string{id})
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
