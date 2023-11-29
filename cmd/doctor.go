package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nikoksr/dbench/internal/build"
	"github.com/nikoksr/dbench/internal/system"
	"github.com/nikoksr/dbench/internal/ui/styles"
	"github.com/nikoksr/dbench/internal/ui/text"
)

type doctorOptions struct {
	*globalOptions

	showSystemDetails bool
}

func newDoctorCommand(globalsOpts *globalOptions) *cobra.Command {
	opts := &doctorOptions{
		globalOptions: globalsOpts,
	}

	cmd := &cobra.Command{
		Use:     "doctor",
		Aliases: []string{"d", "doc"},
		Short:   "Check if dbench is ready to run",
		Long: `Check if dbench is ready to run.

Enabling the --system flag, dbench will show you the exact system details that would be collected during a benchmark if you execute 'dbench run' with the '--collect-sysinfo' flag.`,
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		Args:                  cobra.NoArgs,
		ValidArgsFunction:     cobra.NoFileCompletions,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s\n", styles.Title.Render(fmt.Sprintf("%s v%s", build.AppName, build.Version)))

			// Check dbench database
			fmt.Printf("%s\n", styles.SubTitle.Render("Checking dbench database..."))
			fmt.Print(styles.Info.Render("  Connecting ... "))

			// Check required tools
			fmt.Printf("\n%s\n", styles.SubTitle.Render("Checking required tools..."))

			tools := []string{"pgbench", "gnuplot"}
			for _, tool := range tools {
				fmt.Printf(styles.Info.Render("  %s ... "), tool)

				if !isToolInPath(tool) {
					fmt.Println(styles.Error.Render("✗ Not found"))
					continue
				}

				version, err := getToolVersion(tool)
				if err != nil {
					fmt.Printf("%s %s\n", styles.Error.Render("✗ Error:"), err)
					continue
				}

				fmt.Printf("%s (%s)\n", styles.Success.Render("✓ Found"), styles.Info.Render(version))
			}

			fmt.Println()

			// System information
			if !opts.showSystemDetails {
				return // Skip system details
			}

			fmt.Printf("%s\n", styles.SubTitle.Render("System Details"))
			systemDetails, errs := system.GetDetails()

			// Pretty print errors
			if len(errs) > 0 {
				fmt.Println() // Conditional newline
			}

			for _, err := range errs {
				fmt.Printf("  %s %v\n", styles.Error.Render("Warning:"), err)
			}

			if len(errs) > 0 {
				fmt.Println() // Conditional newline
			}

			// Print system details
			fmt.Printf("  %s: %s\n", styles.Info.Render("Machine ID"), text.ValueOrNA(systemDetails.MachineID))
			fmt.Printf("  %s: %s\n", styles.Info.Render("OS Name"), text.ValueOrNA(systemDetails.OsName))
			fmt.Printf("  %s: %s\n", styles.Info.Render("OS Architecture"), text.ValueOrNA(systemDetails.OsArch))
			fmt.Printf("  %s: %s\n", styles.Info.Render("CPUs Count"), text.ValueOrNA(systemDetails.CPUCount))
			fmt.Printf("  %s: %s\n", styles.Info.Render("CPU Vendor"), text.ValueOrNA(systemDetails.CPUVendor))
			fmt.Printf("  %s: %s\n", styles.Info.Render("CPU Model"), text.ValueOrNA(systemDetails.CPUModel))
			fmt.Printf("  %s: %s\n", styles.Info.Render("CPU Cores"), text.ValueOrNA(systemDetails.CPUCores))
			fmt.Printf("  %s: %s\n", styles.Info.Render("CPU Threads"), text.ValueOrNA(systemDetails.CPUThreads))
			fmt.Printf("  %s: %s\n", styles.Info.Render("RAM Physical"), text.HumanizeBytes(systemDetails.RAMPhysical))
			fmt.Printf("  %s: %s\n", styles.Info.Render("RAM Usable"), text.HumanizeBytes(systemDetails.RAMUsable))
			fmt.Printf("  %s: %s\n", styles.Info.Render("Disk Count"), text.ValueOrNA(systemDetails.DiskCount))
			fmt.Printf("  %s: %s\n", styles.Info.Render("Disk Total"), text.HumanizeBytes(systemDetails.DiskSpaceTotal))

			fmt.Println()
		},
	}

	cmd.Flags().BoolVarP(&opts.showSystemDetails, "sysinfo", "s", false, "Show detailed system information")

	return cmd
}
