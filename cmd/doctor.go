package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"runtime"

	"github.com/nikoksr/dbench/internal/build"
	"github.com/nikoksr/dbench/internal/database"
	"github.com/nikoksr/dbench/internal/styles"
)

func newDoctorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "doctor",
		Aliases:               []string{"d", "doc"},
		Short:                 "Check if dbench is ready to run",
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		Args:                  cobra.NoArgs,
		ValidArgsFunction:     cobra.NoFileCompletions,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()

			fmt.Printf("%s\n", styles.Title.Render(fmt.Sprintf("%s %s", build.AppName, build.Version)))

			// System information
			fmt.Printf("%s\n", styles.SubTitle.Render("System Information"))
			fmt.Printf("  OS: %s\n", runtime.GOOS)
			fmt.Printf("  Architecture: %s\n", runtime.GOARCH)

			// Check dbench database
			fmt.Printf("\n%s\n", styles.SubTitle.Render("Checking dbench database..."))
			fmt.Print(styles.Info.Render("  Connecting ... "))

			dbenchDB, err := database.NewEntDatabase(ctx, dbenchDSN)
			_ = dbenchDB.Close()

			if err != nil {
				fmt.Printf("%s %s\n", styles.Error.Render("✗ Error:"), err)
				return
			} else {
				fmt.Println(styles.Success.Render("✓ Success"))
			}

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
		},
	}

	return cmd
}
