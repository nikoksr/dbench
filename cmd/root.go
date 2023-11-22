package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/nikoksr/dbench/pkg/buildinfo"
)

var dbenchDSN = fmt.Sprintf("file:%s.db?cache=shared&_fk=1", buildinfo.AppName)

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   buildinfo.AppName + " [COMMAND]",
		Short:                 "A nifty wrapper around pgbench that comes with plotting and result management.",
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		Args:                  cobra.NoArgs,
		ValidArgsFunction:     cobra.NoFileCompletions,
		Version:               buildinfo.Version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
		},
	}

	// Print the version number without the app name
	cmd.SetVersionTemplate("{{.Version}}\n")

	cmd.AddGroup(&cobra.Group{
		ID:    "commands",
		Title: "Commands",
	})

	// Subcommands
	cmd.AddCommand(
		newBenchRunCommand(nil),
		newBenchInitCommand(nil),
		newBenchListCommand(),
		newBenchExportCommand(),
		newPlotCommand(),
		newDoctorCommand(),
	)

	return cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd := newRootCommand()

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "\n\nError: %s\n", err)
		os.Exit(1)
	}
}

var (
	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#f8f8f2")) // Light gray

	styleSubTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#8be9fd")) // Cyan

	styleSuccess = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#50fa7b"))

	styleError = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ff5555"))

	styleInfo = lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("#abb2bf")) // Lighter gray
)

func isToolInPath(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}
