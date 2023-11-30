package cmd

import (
	"context"
	"fmt"
	"github.com/nikoksr/dbench/cmd/cobrax"
	"github.com/nikoksr/dbench/internal/database"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/nikoksr/dbench/internal/build"
	"github.com/nikoksr/dbench/internal/env"
	"github.com/nikoksr/dbench/internal/fs"
)

// dbConnector is a function type that returns a new database connection. It is used by all subcommands to get a
// database connection. We return the database.Store interface instead of the database.DB struct to avoid accidentally
// calling the Close() method on the database connection in a subcommand. The Close() method is only called once after
// all subcommands have finished by the root's PersistentPostRunE hook.
type dbConnector func(ctx context.Context, dataDir string, fs fs.FileSystem) (database.Store, error)

func newDBConnector(db *database.DB) dbConnector {
	connector := func(ctx context.Context, dataDir string, fs fs.FileSystem) (database.Store, error) {
		if dataDir == "" {
			return nil, fmt.Errorf("path to data directory is empty")
		}

		// Make sure the data directory exists and create it if not
		if err := fs.MkdirAll(dataDir, 0o755); err != nil {
			return nil, fmt.Errorf("create data directory: %w", err)
		}

		// Set database DSN
		dbenchDSN := buildDSN(dataDir)

		// Open database connection
		if _, err := db.Connect(ctx, dbenchDSN); err != nil {
			return nil, err
		}

		// Check if database is ready
		if err := db.IsReady(); err != nil {
			return nil, fmt.Errorf("database is not ready: %w", err)
		}

		return db, nil
	}

	return connector
}

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

	//noinspection GoBoolExpressions
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

	// Create the application wide database connector function
	db := database.New()
	dbConnector := newDBConnector(db)

	cmd := &cobra.Command{
		Use:                   build.AppName + " [COMMAND]",
		Short:                 "A nifty wrapper around pgbench that comes with plotting and result management.",
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		Args:                  cobra.NoArgs,
		ValidArgsFunction:     cobra.NoFileCompletions,
		Version:               build.Version,
		PersistentPostRunE:    cobrax.HooksE(closeDBHook(db)),
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
		newRunCommand(opts, dbConnector),
		newListCommand(opts, dbConnector),
		newExportCommand(opts, dbConnector),
		newRemoveCommand(opts, dbConnector),
		// Plotting
		newPlotCommand(opts, dbConnector),
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
