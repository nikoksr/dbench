package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.jetpack.io/typeid"

	"github.com/nikoksr/dbench/cmd/cobrax"
	"github.com/nikoksr/dbench/internal/database"
	"github.com/nikoksr/dbench/internal/ui/styles"
)

type removeOptions struct {
	*globalOptions
}

func newRemoveCommand(globalOpts *globalOptions) *cobra.Command {
	opts := &removeOptions{
		globalOptions: globalOpts,
	}

	db := new(database.Database)

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
		PreRunE:               cobrax.HooksE(prepareDBHook(db, opts.dataDir)),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Convert and validate ids
			fmt.Printf("%s\n", styles.Title.Render("Validation"))
			fmt.Printf("%s\t", styles.Text.Render("Validating ids..."))

			var ids, groupIDs []string
			for _, arg := range args {
				// Try to convert id to typeid
				id, err := typeid.FromString(arg)
				if err != nil {
					fmt.Println(styles.Error.Render("✗ Failed\n"))
					return fmt.Errorf("convert id to typeid: %w", err)
				}

				if id.Prefix() == "bmkgrp" {
					groupIDs = append(groupIDs, id.String())
				} else {
					ids = append(ids, id.String())
				}
			}

			fmt.Println(styles.Success.Render("✓ Success"))

			// Remove benchmarks
			ctx := cmd.Context()

			if len(ids) > 0 {
				fmt.Printf("%s\n", styles.Title.Render("Remove benchmarks"))
				msg := fmt.Sprintf("Removing %d benchmark(s)", len(ids))
				fmt.Printf("%s\t", styles.Text.Render(msg))

				if err := db.RemoveByIDs(ctx, ids); err != nil {
					fmt.Println(styles.Error.Render("✗ Failed\n"))
					return fmt.Errorf("remove benchmarks by ids: %w", err)
				}

				fmt.Println(styles.Success.Render("✓ Success"))
			}

			// Remove benchmark groups

			if len(groupIDs) > 0 {
				fmt.Printf("%s\n", styles.Title.Render("Remove benchmark groups"))
				msg := fmt.Sprintf("Removing %d benchmark-group(s)", len(groupIDs))
				fmt.Printf("%s\t", styles.Text.Render(msg))

				if err := db.RemoveByGroupIDs(ctx, groupIDs); err != nil {
					fmt.Println(styles.Error.Render("✗ Failed\n"))
					return fmt.Errorf("remove benchmarks by group ids: %w", err)
				}

				fmt.Println(styles.Success.Render("✓ Success"))
			}

			fmt.Println()

			return nil
		},
		PostRunE: cobrax.HooksE(closeDatabaseHook(db)),
	}

	return cmd
}
