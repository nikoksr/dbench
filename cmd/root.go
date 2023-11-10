package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/nikoksr/dbench/pkg/buildinfo"
)

var resultsDSN = fmt.Sprintf("file:%s.db?cache=shared&_fk=1", buildinfo.AppName)

// requiredEnv holds the environment variables that must be set for the application to run. Each key is mapped to a
// helpful error message that will be returned if the environment variable is not set.
var requiredEnv = map[string]string{
	"PGPASSWORD": `Environment variable PGPASSWORD must be set to the password of the specified PostgreSQL user.

	# Example
	export PGPASSWORD=supersecret

For more information, see the official documentation:
https://www.postgresql.org/docs/current/libpq-envars.html
`,
}

// requiredTools holds the tools that must be installed for the application to run. Each key is mapped to a helpful
// error message that will be returned if the tool is not installed.
var requiredTools = map[string]string{
	"pgbench": `pgbench is required to run the application. It can be installed with the following command:

	# Arch
	sudo pacman -S postgresql

	# Debian / Ubuntu
	sudo apt install postgresql-client

	# macOS
	brew install postgresql

For more information, see the official documentation:
https://www.postgresql.org/docs/current/pgbench.html
`,
	"gnuplot": `gnuplot is required to run the application. It can be installed with the following command:

	# Arch
	sudo pacman -S gnuplot

	# Debian / Ubuntu
	sudo apt install gnuplot

	# macOS
	brew install gnuplot

For more information, see the official documentation:
http://www.gnuplot.info/
`,
}

func validateRequirements() error {
	isInstalled := func(tool string) bool {
		_, err := exec.LookPath(tool)
		return err == nil
	}

	for tool, msg := range requiredTools {
		if !isInstalled(tool) {
			return fmt.Errorf(msg)
		}
	}

	for env, msg := range requiredEnv {
		if os.Getenv(env) == "" {
			return fmt.Errorf(msg)
		}
	}

	return nil
}

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               buildinfo.AppName + " [command]",
		Short:             "A nifty wrapper around pgbench that comes with plotting and result management.",
		SilenceUsage:      true,
		SilenceErrors:     true,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		Version:           buildinfo.Version,
	}

	// Print the version number without the app name
	cmd.SetVersionTemplate("{{.Version}}\n")

	// Subcommands
	cmd.AddCommand(
		newBenchCommand(),
		newPlotCommand(),
	)

	return cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd := newRootCommand()

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
