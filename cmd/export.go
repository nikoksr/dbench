package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/nikoksr/dbench/ent"
	"github.com/nikoksr/dbench/ent/benchmark"
	"github.com/nikoksr/dbench/internal/archive"
	"github.com/nikoksr/dbench/internal/build"
	"github.com/nikoksr/dbench/internal/database"
	"github.com/nikoksr/dbench/internal/fs"
	"github.com/nikoksr/dbench/internal/portability/converter"
	"github.com/nikoksr/dbench/internal/portability/exporter"
	"github.com/nikoksr/dbench/internal/ui/printer"
)

const (
	csvFileFormat  = "csv"
	jsonFileFormat = "json"
)

type exportOptions struct {
	*globalOptions

	// Flags
	format    string
	batchSize int
	archive   bool
	keep      bool

	// Internal
	targetDir string
}

func generateExportDirPath() string {
	localTime := time.Now().Local().Format("2006-01-02_15-04-05")
	return filepath.Join(build.AppName, "export", localTime)
}

func generateExportFilePath(fileFormat string) string {
	return fmt.Sprintf("export_%d.%s", time.Now().Local().UnixMilli(), fileFormat)
}

func generateArchiveFileName() string {
	localTime := time.Now().Local().Format("2006-01-02_15-04-05")
	return fmt.Sprintf("%s-export_%s", build.AppName, localTime)
}

func newExportCommand(globalOpts *globalOptions, connectToDB dbConnector) *cobra.Command {
	opts := &exportOptions{
		globalOptions: globalOpts,
	}

	cmd := &cobra.Command{
		Use:                   "export [OPTIONS]",
		Aliases:               []string{"e"},
		GroupID:               "commands",
		Short:                 "Export all data to a format of your choice",
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		Args:                  cobra.NoArgs,
		ValidArgsFunction:     cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := connectToDB(cmd.Context(), opts.dataDir, opts.noMigration, fs.OSFileSystem{})
			if err != nil {
				return fmt.Errorf("connect to database: %w", err)
			}

			// Print header
			p := printer.NewPrinter(cmd.OutOrStdout(), 40)
			p.PrintlnTitle("Export")
			p.PrintlnSubTitle("Preparation")

			// Check how many benchmarks are stored
			p.PrintInfo(" Checking database for benchmarks ... ", printer.WithIndent())

			count, err := db.Count(cmd.Context())
			if err != nil {
				p.PrintlnError(err.Error())
				return fmt.Errorf("count benchmarks: %w", err)
			}

			if count <= 0 {
				p.PrintlnWarning("no benchmarks found")
				p.Spacer(2)
				return nil
			}

			p.PrintlnSuccess("")

			// Generate export directory
			p.PrintInfo(" Creating export directory ... ", printer.WithIndent())

			if err := os.MkdirAll(opts.targetDir, 0o755); err != nil {
				p.PrintlnError(err.Error())
				return fmt.Errorf("create export directory: %w", err)
			}

			p.PrintlnSuccess("")
			p.Spacer(2)

			// Sanitize batch size
			p.PrintlnSubTitle("Exporting benchmarks")

			// Sanitize batch size
			opts.batchSize = sanitizeBatchSize(opts.batchSize)

			// We export in batches of 5_000 by default
			runs := int(count / uint64(opts.batchSize))
			remainder := int(count % uint64(opts.batchSize))
			if remainder > 0 {
				runs++
			}

			for i := 0; i < runs; i++ {
				offset := i * opts.batchSize
				limit := opts.batchSize
				if i == runs-1 && remainder > 0 {
					limit = remainder
				}

				if err := exportBenchmarksBatch(cmd.Context(), opts, p, db, offset, limit); err != nil {
					return err
				}
			}

			// Only print success message if we:
			// - don't archive
			// - archive and keep the export directory
			if !opts.archive || (opts.archive && opts.keep) {
				p.Spacer(2)
				p.PrintText(" Complete! Exported data to ")
				p.PrintlnHighlight(opts.targetDir)
			}

			// Archive export directory
			if !opts.archive {
				fmt.Println()
				return nil
			}

			p.PrintlnTitle("Archive")
			p.PrintInfo(" Archiving export directory ... ", printer.WithIndent())

			// Archive the export directory
			fileName := generateArchiveFileName()

			fileName, err = archive.Directory(cmd.Context(), opts.targetDir, fileName)
			if err != nil {
				p.PrintlnError(err.Error())
				return fmt.Errorf("archive export directory: %w", err)
			}

			p.PrintlnSuccess("")

			// Remove export directory
			if !opts.keep {
				p.PrintInfo(" Removing export directory ... ", printer.WithIndent())

				if err := os.RemoveAll(opts.targetDir); err != nil {
					p.PrintlnError(err.Error())
					return fmt.Errorf("remove export directory: %w", err)
				}

				p.PrintlnSuccess("")
			}

			p.Spacer(2)
			p.PrintText(" Complete! Archive saved to ")
			p.PrintlnHighlight(fileName)
			p.Spacer(2)

			return nil
		},
	}

	// Default values
	opts.targetDir = generateExportDirPath()

	// Flags
	cmd.Flags().BoolVarP(&opts.archive, "archive", "a", false, "Archive the export directory")
	cmd.Flags().BoolVarP(&opts.keep, "keep", "k", false, "Keep the export directory after archiving")
	cmd.Flags().StringVarP(&opts.format, "format", "f", jsonFileFormat, "Format to export benchmarks to (json, csv)")
	cmd.Flags().StringVarP(&opts.targetDir, "output-dir", "o", opts.targetDir, "Directory to export data to")
	cmd.Flags().IntVarP(&opts.batchSize, "batch-size", "b", defaultBatchSize, "Number of benchmarks to export per file")

	_ = cmd.MarkFlagDirname("output-dir")
	_ = cmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{jsonFileFormat, csvFileFormat}, cobra.ShellCompDirectiveNoFileComp
	})

	cmd.Flags().SortFlags = false

	return cmd
}

func exportBenchmarksBatch(ctx context.Context, opts *exportOptions, p *printer.Printer, db database.Store, offset, limit int) error {
	// Query benchmarks
	benchmarkRange := fmt.Sprintf("(%d-%d)", offset+1, offset+limit)
	p.PrintInfo(" Fetching ... ", printer.WithIndent())

	// The order function we're using here is the reverse of our default order as seen in [internal/database/database.go].
	// When we're, for example, running `dbench ls`, we want to see the latest benchmarks first. When we're exporting,
	// we want to export the oldest benchmarks first, so that when they're imported again, they're in the correct order.
	benchmarks, err := db.Fetch(ctx,
		database.WithOrderBy(func(query *ent.BenchmarkQuery) *ent.BenchmarkQuery {
			return query.Order(
				ent.Asc(benchmark.FieldGroupID),
				ent.Asc(benchmark.FieldID),
			)
		}),
		database.WithLimit(limit),
		database.WithOffset(offset),
	)
	if err != nil {
		p.PrintlnError(err.Error())
		return fmt.Errorf("fetch benchmarks: %w", err)
	}

	if len(benchmarks) == 0 {
		p.PrintlnWarning("no benchmarks found")
		return fmt.Errorf("no benchmarks found")
	}

	p.PrintlnSuccess(benchmarkRange)

	// Export benchmarks
	p.PrintInfo(" Writing ... ", printer.WithIndent())

	// Open file
	var (
		format  string
		expFunc exporter.Exporter
		expData any
	)

	switch opts.format {
	case csvFileFormat:
		format = csvFileFormat
		expFunc = exporter.ToCSV
		expData = converter.BenchmarksToCSV(benchmarks)
	case jsonFileFormat:
		format = jsonFileFormat
		expFunc = exporter.ToJSON
		expData = benchmarks // Doesn't need conversion
	default:
		p.PrintlnError("unknown export format")
		return fmt.Errorf("unknown export format: %s", opts.format)
	}

	// Create the file
	fileName := generateExportFilePath(format)
	path := filepath.Join(opts.targetDir, fileName)

	file, err := os.Create(path)
	if err != nil {
		p.PrintlnError(err.Error())
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	// Export data
	if err := expFunc(file, expData); err != nil {
		p.PrintlnError(err.Error())
		return fmt.Errorf("export benchmarks: %w", err)
	}

	p.PrintlnSuccess(benchmarkRange)

	return nil
}
