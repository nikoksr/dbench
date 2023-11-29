package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/nikoksr/dbench/cmd/cobrax"
	"github.com/nikoksr/dbench/internal/build"
	"github.com/nikoksr/dbench/internal/database"
	"github.com/nikoksr/dbench/internal/export"
	"github.com/nikoksr/dbench/internal/ui/styles"
)

type exportOptions struct {
	*globalOptions

	format string
}

func newExportCommand(globalOpts *globalOptions) *cobra.Command {
	opts := &exportOptions{
		globalOptions: globalOpts,
	}

	db := new(database.Database)

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
		PreRunE:               cobrax.HooksE(prepareDBHook(db, globalOpts.dataDir)),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Query benchmarks
			benchmarks, err := db.Fetch(cmd.Context())
			if err != nil {
				return fmt.Errorf("fetch benchmarks: %w", err)
			}

			if len(benchmarks) == 0 {
				return fmt.Errorf("no benchmarks found")
			}

			// Export benchmarks
			var dataFile string
			switch opts.format {
			case "csv":
				dataFile, err = export.ToCSV(benchmarks, generateExportFileName("csv"))
			case "json":
				dataFile, err = export.ToJSON(benchmarks, generateExportFileName("json"))
			default:
				return fmt.Errorf("unknown export format: %s", opts.format)
			}

			if err != nil {
				return fmt.Errorf("export benchmarks: %w", err)
			}

			fmt.Printf("\nExported data to %s\n\n", styles.Success.Render(dataFile))

			return nil
		},
		PostRunE: cobrax.HooksE(closeDatabaseHook(db)),
	}

	// Flags
	cmd.Flags().StringVar(&opts.format, "format", "csv", "Format to export benchmarks to (csv, json)")

	return cmd
}
