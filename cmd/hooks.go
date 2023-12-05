package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nikoksr/dbench/cmd/cobrax"
	"github.com/nikoksr/dbench/internal/database"
)

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
