package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.jetpack.io/typeid"

	"github.com/nikoksr/dbench/internal/fs"
	"github.com/nikoksr/dbench/internal/ui/printer"
)

type removeOptions struct {
	*globalOptions
}

func newRemoveCommand(globalOpts *globalOptions, connectToDB dbConnector) *cobra.Command {
	opts := &removeOptions{
		globalOptions: globalOpts,
	}

	cmd := &cobra.Command{
		Use:                   "remove ID [ID...]",
		Aliases:               []string{"r", "rm"},
		GroupID:               "commands",
		Short:                 "Remove benchmarks from the database",
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		Args:                  cobra.MinimumNArgs(1),
		ValidArgsFunction:     cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := connectToDB(cmd.Context(), opts.dataDir, opts.noMigration, fs.OSFileSystem{})
			if err != nil {
				return fmt.Errorf("connect to database: %w", err)
			}

			// Print header
			p := printer.NewPrinter(cmd.OutOrStdout(), 60)
			p.PrintlnTitle("Remove")

			// Convert and validate ids
			p.PrintlnSubTitle("Validation")

			var ids, groupIDs []string
			for _, arg := range args {
				// Try to convert id to typeid
				p.PrintInfo(fmt.Sprintf(" Validating ID %s ... ", arg), printer.WithIndent())

				id, err := typeid.FromString(arg)
				if err != nil {
					p.PrintlnError(err.Error())
					return fmt.Errorf("convert id to typeid: %w", err)
				}

				p.PrintlnSuccess("")

				if id.Prefix() == "bmkgrp" {
					groupIDs = append(groupIDs, id.String())
				} else {
					ids = append(ids, id.String())
				}
			}

			// Remove benchmark groups
			ctx := cmd.Context()

			p.Spacer(2)
			p.PrintlnSubTitle("Removing")

			if len(groupIDs) > 0 {
				p.PrintInfo(fmt.Sprintf(" Removing %d benchmark-group(s)", len(groupIDs)), printer.WithIndent())

				if err := db.RemoveByGroupIDs(ctx, groupIDs); err != nil {
					p.PrintlnError(err.Error())
					return fmt.Errorf("remove benchmarks by group ids: %w", err)
				}

				p.PrintlnSuccess("")
			}

			// Remove benchmarks

			if len(ids) > 0 {
				p.PrintInfo(fmt.Sprintf(" Removing %d benchmark(s)", len(ids)), printer.WithIndent())

				if err := db.RemoveByIDs(ctx, ids); err != nil {
					p.PrintlnError(err.Error())
					return fmt.Errorf("remove benchmarks by ids: %w", err)
				}

				p.PrintlnSuccess("")
			}

			p.Spacer(2)

			return nil
		},
	}

	return cmd
}
