package cmd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/nikoksr/dbench/cmd/cobrax"
	"github.com/nikoksr/dbench/internal/build"
	"github.com/nikoksr/dbench/internal/database"
	"github.com/nikoksr/dbench/internal/fs"
)

func prepareDBHook(db *database.Database, dataDir string) cobrax.HookE {
	return func(cmd *cobra.Command, args []string) error {
		if db == nil {
			// DB is used in other callbacks, so we need to make sure it has been initialized
			// before. Initializing it here is too late, as the reference to the db is passed
			// to the other callbacks before this one is executed.
			return fmt.Errorf("database is nil")
		}

		var err error
		_db, err := prepareDatabase(cmd.Context(), dataDir, fs.OSFileSystem{})
		if err != nil {
			return fmt.Errorf("prepare storage: %w", err)
		}

		// Sanity check
		if _db == nil {
			return fmt.Errorf("dbench database is nil")
		}

		*db = *_db

		return nil
	}
}

func closeDatabaseHook(db *database.Database) cobrax.HookE {
	return func(cmd *cobra.Command, args []string) error {
		if err := db.Close(); err != nil {
			return fmt.Errorf("close database: %w", err)
		}

		return nil
	}
}

func pgbenchInstalledHook() cobrax.HookE {
	return func(cmd *cobra.Command, args []string) error {
		if !isToolInPath("pgbench") {
			return errPgbenchNotInstalled
		}

		return nil
	}
}

func gnuplotInstalledHook() cobrax.HookE {
	return func(cmd *cobra.Command, args []string) error {
		if !isToolInPath("gnuplot") {
			return errGNUPlotNotInstalled
		}

		return nil
	}
}

func buildDSN(dataDir string) string {
	path := filepath.Join(dataDir, build.AppName)
	return fmt.Sprintf("file:%s.db?cache=shared&_fk=1", path)
}

func prepareDatabase(ctx context.Context, dataDir string, fs fs.FileSystem) (*database.Database, error) {
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
	db, err := database.New(ctx, dbenchDSN)
	if err != nil {
		return nil, fmt.Errorf("open dbench database: %w", err)
	}

	return db, nil
}
