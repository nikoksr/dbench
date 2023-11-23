package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.jetpack.io/typeid"

	"github.com/nikoksr/dbench/internal/store"
	"github.com/nikoksr/dbench/internal/ui/styles"
)

func newRemoveCommand() *cobra.Command {
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
			// Open database connection
			ctx := cmd.Context()
			dbenchDB, err := store.New(ctx, dbenchDSN)
			if err != nil {
				return fmt.Errorf("create dbench database: %w", err)
			}
			defer dbenchDB.Close()

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

			if len(ids) > 0 {
				fmt.Printf("%s\n", styles.Title.Render("Remove benchmarks"))
				msg := fmt.Sprintf("Removing %d benchmark(s)", len(ids))
				fmt.Printf("%s\t", styles.Text.Render(msg))

				if err := dbenchDB.RemoveByIDs(ctx, ids); err != nil {
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

				if err := dbenchDB.RemoveByGroupIDs(ctx, groupIDs); err != nil {
					fmt.Println(styles.Error.Render("✗ Failed\n"))
					return fmt.Errorf("remove benchmarks by group ids: %w", err)
				}

				fmt.Println(styles.Success.Render("✓ Success"))
			}

			fmt.Println()

			return nil
		},
	}

	return cmd
}
