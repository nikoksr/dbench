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
		Use:               "plot [result-group ID]",
		Aliases:           []string{"p"},
		GroupID:           "commands",
		Short:             "Plot results produced by the bench command",
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
			// Get result-group ID
			resultGroupID := args[0]

			// Open database connection
			ctx := cmd.Context()
			resultsDB, err := database.NewEntDatabase(ctx, resultsDSN)
			if err != nil {
				return fmt.Errorf("create results database: %w", err)
			}
			defer resultsDB.Close()

			results, err := resultsDB.FetchResultsByGroupIDs(ctx, []string{resultGroupID})
			if err != nil {
				return fmt.Errorf("fetch results by result-group ID: %w", err)
			}

			// Open data file for gnu plot
			f, err := os.CreateTemp("", buildinfo.AppName+".*.txt")
			if err != nil {
				return fmt.Errorf("create data file: %w", err)
			}
			defer func() { _ = os.Remove(f.Name()) }()

			dataFile := f.Name()
			_ = f.Close()

			if err := export.ToGnuplotBasic(results, dataFile); err != nil {
				return fmt.Errorf("export results to gnuplot: %w", err)
			}

			// Generate plots using gnuplot
			plotOutputName := fmt.Sprintf("plot_%s", resultGroupID)
			outputPath, err := plot.Plot(f.Name(), plotOutputName)
			if err != nil {
				return fmt.Errorf("plot results: %w", err)
			}

			fmt.Printf("Plot saved to %s\n", outputPath)

			return nil
		},
	}

	return cmd
}
