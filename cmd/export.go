package cmd

import (
	"fmt"
	"github.com/nikoksr/dbench/internal/build"
	"github.com/nikoksr/dbench/internal/database"
	"github.com/nikoksr/dbench/internal/export"
	"github.com/nikoksr/dbench/internal/styles"
	"github.com/spf13/cobra"
	"time"
)

func newExportCommand() *cobra.Command {
	var format string

	generateExportFileName := func(fileFormat string) string {
		localTime := time.Now().Local().Format("2006-01-02_15-04-05")
		return fmt.Sprintf("%s-export_%s.%s", build.AppName, localTime, fileFormat)
	}

	cmd := &cobra.Command{
		Use:                   "export [OPTIONS]",
		Aliases:               []string{"e"},
		GroupID:               "commands",
		Short:                 "Export all data to a format of your choice",
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		Args:                  cobra.NoArgs,
		ValidArgsFunction:     cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Open database connection
			ctx := cmd.Context()
			dbenchDB, err := database.NewEntDatabase(ctx, dbenchDSN)
			if err != nil {
				return fmt.Errorf("create dbench database: %w", err)
			}
			defer dbenchDB.Close()

			// Query benchmarks
			benchmarks, err := dbenchDB.FetchBenchmarks(ctx)
			if err != nil {
				return fmt.Errorf("fetch benchmarks: %w", err)
			}

			if len(benchmarks) == 0 {
				return fmt.Errorf("no benchmarks found")
			}

			// Export benchmarks
			var dataFile string
			switch format {
			case "csv":
				dataFile, err = export.ToCSV(benchmarks, generateExportFileName("csv"))
			case "json":
				dataFile, err = export.ToJSON(benchmarks, generateExportFileName("json"))
			default:
				return fmt.Errorf("unknown export format: %s", format)
			}

			if err != nil {
				return fmt.Errorf("export benchmarks: %w", err)
			}

			fmt.Printf("\nExported data to %s\n\n", styles.Success.Render(dataFile))

			return nil
		},
	}

	// Flags
	cmd.Flags().StringVar(&format, "format", "csv", "Format to export benchmarks to (csv, json)")

	return cmd
}
