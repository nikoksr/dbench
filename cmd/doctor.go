package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nikoksr/dbench/pkg/buildinfo"
	"github.com/nikoksr/dbench/pkg/database"
	"github.com/nikoksr/dbench/pkg/styles"
)

func getToolVersion(tool string) (string, error) {
	cmd := exec.Command(tool, "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	// Remove all trailling spaces and newlines
	version := string(out)
	version = strings.Trim(version, "\n")
	version = strings.TrimSpace(version)

	if tool == "pgbench" {
		version = strings.TrimPrefix(version, "pgbench (PostgreSQL) ")
	} else if tool == "gnuplot" {
		version = strings.TrimPrefix(version, "gnuplot ")
	}

	return version, nil
}

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

			fmt.Printf("%s\n", styles.Title.Render(fmt.Sprintf("%s %s", buildinfo.AppName, buildinfo.Version)))

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
