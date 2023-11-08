package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/nikoksr/dbench/pkg/buildinfo"
	"github.com/nikoksr/dbench/pkg/database"
	"github.com/nikoksr/dbench/pkg/export"
	"github.com/nikoksr/dbench/pkg/plot"
)

func newPlotCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "plot [result-group ID]",
		Short:             "Plot results produced by the bench command",
		SilenceUsage:      true,
		SilenceErrors:     true,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cobra.NoFileCompletions,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return validateRequirements()
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
