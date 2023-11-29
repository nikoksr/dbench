package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nikoksr/dbench/internal/build"
	"github.com/nikoksr/dbench/internal/env"
	"github.com/nikoksr/dbench/internal/fs"
)

var (
	// envPrefix is the prefix for environment variables. E.g. DBENCH_DATA_DIR
	envPrefix  = strings.ToUpper(build.AppName)
	envDataDir = envPrefix + "_DATA_DIR"
)

// determineDefaultDataPath function to return data path based on operating system.
func determineDefaultDataPath(appName string, env env.Environment, fs fs.FileSystem) (string, error) {
	// Check if data directory is set via environment variable
	if envDataDir := env.Getenv(envDataDir); envDataDir != "" {
		return envDataDir, nil
	}

	// Get user home directory
	homeDir, err := fs.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get user home directory: %w", err)
	}

	// Config path for everything but Windows
	var dataDir string
	if runtime.GOOS != "windows" {
		dataDir = filepath.Join(homeDir, ".local", "share", appName)
	} else {
		dataDir = filepath.Join(homeDir, "AppData", "Local", appName)
	}

	return dataDir, nil
}

type globalOptions struct {
	dataDir string
}

func newRootCommand() *cobra.Command {
	opts := new(globalOptions)

	cmd := &cobra.Command{
		Use:                   build.AppName + " [COMMAND]",
		Short:                 "A nifty wrapper around pgbench that comes with plotting and result management.",
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		Args:                  cobra.NoArgs,
		ValidArgsFunction:     cobra.NoFileCompletions,
		Version:               build.Version,
	}

	// Print the version number without the app name
	cmd.SetVersionTemplate("v{{.Version}}\n")

	cmd.AddGroup(&cobra.Group{
		ID:    "commands",
		Title: "Commands",
	})

	// Get default data directory
	dataDir, err := determineDefaultDataPath(build.AppName, env.RealEnvironment{}, fs.OSFileSystem{})
	if err != nil {
		dataDir = "./dbench"
		fmt.Printf("Error: determine default data path: %s\n", err)
		fmt.Printf("Using fallback data directory: %s\n", dataDir)
	}

	// Flags
	cmd.PersistentFlags().StringVar(&opts.dataDir, "data-dir", dataDir, "Path to the data directory")

	// Subcommands
	cmd.AddCommand(
		// Benchmarks
		newInitCommand(opts),
		newRunCommand(opts),
		newListCommand(opts),
		newExportCommand(opts),
		newRemoveCommand(opts),
		// Plotting
		newPlotCommand(opts),
		// Misc
		newDoctorCommand(opts),
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
