package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/nikoksr/dbench/internal/build"
)

var dbenchDSN = fmt.Sprintf("file:%s.db?cache=shared&_fk=1", build.AppName)

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   build.AppName + " [COMMAND]",
		Short:                 "A nifty wrapper around pgbench that comes with plotting and result management.",
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		Args:                  cobra.NoArgs,
		ValidArgsFunction:     cobra.NoFileCompletions,
		Version:               build.Version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
		},
	}

	// Print the version number without the app name
	cmd.SetVersionTemplate("v{{.Version}}\n")

	cmd.AddGroup(&cobra.Group{
		ID:    "commands",
		Title: "Commands",
	})

	// Subcommands
	cmd.AddCommand(
		// Benchmarks
		newInitCommand(),
		newRunCommand(),
		newListCommand(),
		newExportCommand(),
		newRemoveCommand(),
		// Plotting
		newPlotCommand(),
		// Misc
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
