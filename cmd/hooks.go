package cmd

import (
	"fmt"
	"github.com/nikoksr/dbench/internal/build"
	"github.com/nikoksr/dbench/internal/database"
	"github.com/spf13/cobra"
	"path/filepath"

	"github.com/nikoksr/dbench/cmd/cobrax"
)

func buildDSN(dataDir string) string {
	path := filepath.Join(dataDir, build.AppName) + ".db"
	return fmt.Sprintf("file:%s?cache=shared&_fk=1", path)
}

func closeDBHook(db *database.DB) cobrax.HookE {
	return func(cmd *cobra.Command, args []string) error {
		if db != nil {
			return db.Close()
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
