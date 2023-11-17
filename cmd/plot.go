package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/nikoksr/dbench/pkg/database"
	"github.com/nikoksr/dbench/pkg/export"
	"github.com/nikoksr/dbench/pkg/plot"
)

var gnuPlotNotInstalledErr = fmt.Errorf(`gnuplot is required to run the application. It can be installed with the following command:

	# Arch
	sudo pacman -S gnuplot

	# Debian / Ubuntu
	sudo apt install gnuplot

	# macOS
	brew install gnuplot

For more information, see the official documentation:
http://www.gnuplot.info/
`)

func prepareDirectory(dir string) error {
	// Clean directory path
	dir = filepath.Clean(dir)

	// Create output directory if it doesn't exist
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create output directory: %w", err)
		}
	}

	return nil
}

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
			if _, err := exec.LookPath("gnuplot"); err != nil {
				return gnuPlotNotInstalledErr
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Get benchmark-group IDs
			benchmarkGroupIDs := args

			// Cleanup output directory
			if cleanOutputDir {
				if err := os.RemoveAll(outputDir); err != nil {
					return fmt.Errorf("cleanup output directory: %w", err)
				}
			}

			// Prepare output directory
			if err := prepareDirectory(outputDir); err != nil {
				return fmt.Errorf("prepare output directory: %w", err)
			}

			// Plot benchmarks
			for _, id := range benchmarkGroupIDs {
				if err := plotBenchmarks(ctx, id, outputDir); err != nil {
					return fmt.Errorf("plot benchmark-group %q: %w", id, err)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "dbench/plots", "Output directory for plots")
	cmd.Flags().BoolVarP(&cleanOutputDir, "clean", "c", false, "Cleanup output directory before plotting")

	return cmd
}

func plotBenchmarks(ctx context.Context, id, outputDir string) error {
	dbenchDB, err := database.NewEntDatabase(ctx, dbenchDSN)
	if err != nil {
		return fmt.Errorf("create dbench database: %w", err)
	}
	defer dbenchDB.Close()

	benchmarks, err := dbenchDB.FetchBenchmarksByGroupIDs(ctx, []string{id})
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

	fmt.Printf("Plots successfully generated! Check the output directory %q.\n", outputDir)

	return nil
}
