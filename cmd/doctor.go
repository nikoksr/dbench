package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/nikoksr/dbench/internal/build"
	"github.com/nikoksr/dbench/internal/fs"
	"github.com/nikoksr/dbench/internal/system"
	"github.com/nikoksr/dbench/internal/ui/printer"
	"github.com/nikoksr/dbench/internal/ui/text"
)

type doctorOptions struct {
	*globalOptions

	showSystemConfig bool
}

func newDoctorCommand(globalsOpts *globalOptions, connectToDB dbConnector) *cobra.Command {
	opts := &doctorOptions{
		globalOptions: globalsOpts,
	}

	cmd := &cobra.Command{
		Use:     "doctor",
		Aliases: []string{"d", "doc"},
		Short:   "Check if dbench is ready to run",
		Long: `Check if dbench is ready to run.

Enabling the --system flag, dbench will show you the exact system config that would be collected during a benchmark if you execute 'dbench run' with the '--collect-sysinfo' flag.`,
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		Args:                  cobra.NoArgs,
		ValidArgsFunction:     cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Open database connection early. We need to check for potential migrations before printing the doc's
			// output. This way we can show the user a meaningful error message if the database is not up-to-date.
			connStart := time.Now()
			db, err := connectToDB(cmd.Context(), opts.dataDir, opts.noMigration, fs.OSFileSystem{})
			connDuration := time.Since(connStart).Round(time.Millisecond)

			if err != nil {
				return fmt.Errorf("connect to database: %w", err)
			}

			// Print header
			p := printer.NewPrinter(cmd.OutOrStdout(), 30)
			p.PrintlnTitle(fmt.Sprintf("%s %s", build.AppName, build.Version))

			// Check dbench database
			p.PrintlnSubTitle("Database")
			p.PrintInfo(" Connecting ... ", printer.WithIndent())
			p.PrintlnSuccess(connDuration.String())
			p.PrintInfo(" Counting database records ... ", printer.WithIndent())
			count, err := db.CountAll(cmd.Context())
			if err != nil {
				p.PrintlnError(err.Error())
			} else {
				p.PrintlnSuccess(fmt.Sprintf("%d records", count))
			}

			// Check required tools
			p.Spacer(2)
			p.PrintlnSubTitle("Dependencies")

			tools := []string{"pgbench", "gnuplot"}
			for _, tool := range tools {
				p.PrintInfo(fmt.Sprintf("  %s ... ", tool), printer.WithIndent())

				if !isToolInPath(tool) {
					p.PrintlnError("not found")
					continue
				}

				version, err := getToolVersion(tool)
				if err != nil {
					p.PrintlnError(err.Error())
					continue
				}

				p.PrintlnSuccess(version)
			}

			p.Spacer(2)

			// System information
			if !opts.showSystemConfig {
				return nil // Skip system config
			}

			p.PrintlnSubTitle("Host system")
			systemConfig, errs := system.GetConfig()

			// Pretty print errors
			if len(errs) > 0 {
				fmt.Println() // Conditional newline
			}

			for _, err := range errs {
				p.PrintlnWarning(err.Error())
			}

			if len(errs) > 0 {
				fmt.Println() // Conditional newline
			}

			// Print system config

			p.PrintInfo("  Machine ID:", printer.WithIndent())
			p.PrintlnInfo(text.ValueOrNA(systemConfig.MachineID))

			p.PrintInfo("  OS Name:", printer.WithIndent())
			p.PrintlnInfo(text.ValueOrNA(systemConfig.OsName))

			p.PrintInfo("  OS Architecture:", printer.WithIndent())
			p.PrintlnInfo(text.ValueOrNA(systemConfig.OsArch))

			p.PrintInfo("  CPUs Count:", printer.WithIndent())
			p.PrintlnInfo(text.ValueOrNA(systemConfig.CPUCount))

			p.PrintInfo("  CPU Vendor:", printer.WithIndent())
			p.PrintlnInfo(text.ValueOrNA(systemConfig.CPUVendor))

			p.PrintInfo("  CPU Model:", printer.WithIndent())
			p.PrintlnInfo(text.ValueOrNA(systemConfig.CPUModel))

			p.PrintInfo("  CPU Cores:", printer.WithIndent())
			p.PrintlnInfo(text.ValueOrNA(systemConfig.CPUCores))

			p.PrintInfo("  CPU Threads:", printer.WithIndent())
			p.PrintlnInfo(text.ValueOrNA(systemConfig.CPUThreads))

			p.PrintInfo("  RAM Physical:", printer.WithIndent())
			p.PrintlnInfo(text.HumanizeBytes(systemConfig.RAMPhysical))

			p.PrintInfo("  RAM Usable:", printer.WithIndent())
			p.PrintlnInfo(text.HumanizeBytes(systemConfig.RAMUsable))

			p.PrintInfo("  Disk Count:", printer.WithIndent())
			p.PrintlnInfo(text.ValueOrNA(systemConfig.DiskCount))

			p.PrintInfo("  Disk Total:", printer.WithIndent())
			p.PrintlnInfo(text.HumanizeBytes(systemConfig.DiskSpaceTotal))

			p.Spacer(1)

			return nil
		},
	}

	cmd.Flags().BoolVar(&opts.showSystemConfig, "sysinfo", false, "Show detailed system information")

	cmd.Flags().SortFlags = false

	return cmd
}
