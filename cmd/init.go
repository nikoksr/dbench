package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nikoksr/dbench/cmd/cobrax"
	"github.com/nikoksr/dbench/internal/benchmark"
	"github.com/nikoksr/dbench/internal/models"
	"github.com/nikoksr/dbench/internal/ui/styles"
)

type initOptions struct {
	*globalOptions

	benchConfig models.BenchmarkConfig
}

func newInitCommand(globalOpts *globalOptions) *cobra.Command {
	opts := &initOptions{
		globalOptions: globalOpts,
	}

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
		PreRunE:               cobrax.HooksE(pgbenchInstalledHook()),
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

			opts.benchConfig.Password = password

			// Initialize database
			fmt.Printf("%s\n", styles.Title.Render("Initialization"))

			return benchmark.Init(&opts.benchConfig)
		},
	}

	// Database flags
	cmd.Flags().StringVarP(&opts.benchConfig.DBName, "db-name", "d", "postgres", "Name of the database")
	cmd.Flags().StringVarP(&opts.benchConfig.Username, "db-user", "U", "postgres", "Username for connecting to the database")
	cmd.Flags().StringVarP(&opts.benchConfig.Host, "db-host", "H", "localhost", "Host of the database")
	cmd.Flags().StringVarP(&opts.benchConfig.Port, "db-port", "p", "5432", "Port of the database")

	// Init flags
	cmd.Flags().IntVar(&opts.benchConfig.FillFactor, "fill", 100, "Fill factor for the database (10-100)")
	cmd.Flags().IntVar(&opts.benchConfig.ScaleFactor, "scale", 1, "Scale factor for the database (1-1000)")

	return cmd
}
