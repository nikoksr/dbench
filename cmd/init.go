package cmd

import (
	"fmt"
	"github.com/nikoksr/dbench/internal/ui/styles"

	"github.com/spf13/cobra"

	"github.com/nikoksr/dbench/internal/benchmark"
	"github.com/nikoksr/dbench/internal/models"
)

func newInitCommand() *cobra.Command {
	benchConfig := new(models.BenchmarkConfig)

	cmd := &cobra.Command{
		Use:     "init [OPTIONS]",
		Aliases: []string{"i"},
		GroupID: "commands",
		Short:   "Initialize a target database using pgbench",
		Long: `This command initializes a target database using pgbench. You can either use this command to create a new
database or run pgbench manually. If you want to run pgbench manually, this is what an example command would look like:

	pgbench -i -s 1 -F 100 -U postgres -h localhost -p 5432 postgres

For consistency reasons, it is HIGHLY recommended to use this command instead of running pgbench manually.

For more information, see the official documentation:
https://www.postgresql.org/docs/current/pgbench.html
`,
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		ValidArgsFunction:     cobra.NoFileCompletions,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Check if pgbench is installed
			if !isToolInPath("pgbench") {
				return errPgbenchNotInstalled
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("%s\n", styles.Title.Render("Authentication"))

			// Prompt for password
			password, canceled, err := getDBPassword()
			if err != nil {
				return fmt.Errorf("get database password: %w", err)
			}
			if canceled {
				return nil
			}

			benchConfig.Password = password

			// Initialize database
			fmt.Printf("%s\n", styles.Title.Render("Initialization"))

			return benchmark.Init(benchConfig)
		},
	}

	// Database flags
	cmd.Flags().StringVarP(&benchConfig.DBName, "dbname", "d", "postgres", "Name of the database")
	cmd.Flags().StringVarP(&benchConfig.Username, "username", "U", "postgres", "Username for connecting to the database")
	cmd.Flags().StringVarP(&benchConfig.Host, "host", "H", "localhost", "Host of the database")
	cmd.Flags().StringVarP(&benchConfig.Port, "port", "p", "5432", "Port of the database")

	// Init flags
	cmd.Flags().IntVar(&benchConfig.FillFactor, "fill", 100, "Fill factor for the database (10-100)")
	cmd.Flags().IntVar(&benchConfig.ScaleFactor, "scale", 1, "Scale factor for the database (1-1000)")

	return cmd
}
