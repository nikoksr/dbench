package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/nikoksr/dbench/pkg/buildinfo"
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

func newPlotCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "plot [benchmark-group ID]",
		Aliases:           []string{"p"},
		GroupID:           "commands",
		Short:             "Plot benchmarks of a benchmark benchmark-group",
		SilenceUsage:      true,
		SilenceErrors:     true,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cobra.NoFileCompletions,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Check if gnuplot is installed
			if _, err := exec.LookPath("gnuplot"); err != nil {
				return gnuPlotNotInstalledErr
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get benchmark-group ID
			benchmarkGroupID := args[0]

			// Open database connection
			ctx := cmd.Context()
			dbenchDB, err := database.NewEntDatabase(ctx, dbenchDSN)
			if err != nil {
				return fmt.Errorf("create dbench database: %w", err)
			}
			defer dbenchDB.Close()

			benchmarks, err := dbenchDB.FetchBenchmarksByGroupIDs(ctx, []string{benchmarkGroupID})
			if err != nil {
				return fmt.Errorf("fetch benchmarks by benchmark-group ID: %w", err)
			}

			// Open data file for gnu plot
			f, err := os.CreateTemp("", buildinfo.AppName+".*.txt")
			if err != nil {
				return fmt.Errorf("create data file: %w", err)
			}
			defer func() { _ = os.Remove(f.Name()) }()

			dataFile := f.Name()
			_ = f.Close()

			if err := export.ToGnuplotBasic(benchmarks, dataFile); err != nil {
				return fmt.Errorf("export benchmarks to gnuplot: %w", err)
			}

			// Generate plots using gnuplot
			plotOutputName := fmt.Sprintf("plot_%s", benchmarkGroupID)
			outputPath, err := plot.Plot(f.Name(), plotOutputName)
			if err != nil {
				return fmt.Errorf("plot benchmarks: %w", err)
			}

			fmt.Printf("Plot saved to %s\n", outputPath)

			return nil
		},
	}

	return cmd
}
