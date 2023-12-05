package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/nikoksr/dbench/cmd/cobrax"
	"github.com/nikoksr/dbench/internal/archive"
	"github.com/nikoksr/dbench/internal/build"
	"github.com/nikoksr/dbench/internal/database"
	"github.com/nikoksr/dbench/internal/fs"
	"github.com/nikoksr/dbench/internal/models"
	"github.com/nikoksr/dbench/internal/portability/importer"
	"github.com/nikoksr/dbench/internal/ui/printer"
)

func generateExtractDirPath() string {
	return filepath.Join(os.TempDir(), build.AppName, "extract")
}

type importOptions struct {
	*globalOptions

	importPaths []string
	lax         bool
}

func newImportCommand(globalOpts *globalOptions, connectToDB dbConnector) *cobra.Command {
	opts := &importOptions{
		globalOptions: globalOpts,
	}

	cmd := &cobra.Command{
		Use:                   "import [OPTIONS] PATH [PATH...]",
		Aliases:               []string{"i"},
		GroupID:               "commands",
		Short:                 "Import dbench data exports",
		Long:                  importLongDesc,
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		Args:                  cobra.MinimumNArgs(1),
		PreRunE:               cobrax.HooksE(pgbenchInstalledHook()),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := connectToDB(cmd.Context(), opts.dataDir, opts.noMigration, fs.OSFileSystem{})
			if err != nil {
				return fmt.Errorf("connect to database: %w", err)
			}

			// Print header
			p := printer.NewPrinter(cmd.OutOrStdout(), 50)
			p.PrintlnTitle("Import")
			p.PrintlnSubTitle("Preparation")

			// Set import paths
			opts.importPaths = args

			// In case we got any archive files
			extractDir := generateExtractDirPath()

			// Cleanup extracted directory
			if err := os.RemoveAll(extractDir); err != nil {
				return fmt.Errorf("remove extracted directory: %w", err)
			}

			// Make sure we can extract archives
			if err := os.MkdirAll(extractDir, 0o755); err != nil {
				return fmt.Errorf("prepare directory: %w", err)
			}

			defer func() {
				// Cleanup extracted directory
				_ = os.RemoveAll(extractDir)
			}()

			// Collect all files to import
			p.PrintInfo(" Collecting files to import ...", printer.WithIndent())

			files := make([]string, 0)
			for _, path := range opts.importPaths {
				// Check if path is a directory
				if !isPathADirectory(path) {

					// It's a file, check if it's an archive
					if !archive.IsArchive(path) {
						// It's a regular file, add it to import list and continue
						files = append(files, path)
						continue
					}

					// It's an archive, extract it
					if err := extractArchive(cmd.Context(), path, extractDir); err != nil {
						p.PrintlnError(err.Error())
						return fmt.Errorf("extract archive: %w", err)
					}

					// We fall through here to add all files in the extracted directory to the import list
					path = extractDir
				}

				// Get all files in directory
				filesInDir, err := os.ReadDir(path)
				if err != nil {
					p.PrintlnError(err.Error())
					return fmt.Errorf("read directory: %w", err)
				}

				// Add all files to import list
				for _, file := range filesInDir {
					files = append(files, filepath.Join(path, file.Name()))
				}
			}

			p.PrintlnSuccess("")

			// Import files
			p.Spacer(2)
			p.PrintlnSubTitle("Importing")

			start := time.Now()
			benchmarksImported := 0

			// Only initialize map if we're not in lax mode
			var importedFiles map[string]struct{}
			if !opts.lax {
				importedFiles = make(map[string]struct{}, len(files))
			}

			for _, path := range files {
				p.PrintInfo(fmt.Sprintf(" %s ... ", filepath.Base(path)), printer.WithIndent())

				// Check if file is a JSON file
				if !strings.EqualFold(filepath.Ext(path), ".json") {
					p.PrintlnWarning("skipping; not a JSON file")
					continue
				}

				// Check if file was already imported
				if _, found := importedFiles[path]; found && !opts.lax {
					p.PrintlnWarning("skipping; already imported")
					continue
				}

				// Import benchmarks from file
				benchmarkCount, err := importBenchmarksFromFile(cmd.Context(), db, path)
				if err != nil {
					p.PrintlnError(err.Error())
					return err
				}

				// Increment benchmark counter
				benchmarksImported += benchmarkCount

				// Add file to imported files. The lax flag check here isn't really necessary but helps with memory
				// usage in case we import a lot of files.
				if !opts.lax {
					importedFiles[path] = struct{}{}
				}

				p.PrintlnSuccess(fmt.Sprintf("imported %d benchmarks", benchmarkCount))
			}

			// Print benchmark complete message

			timeToImport := time.Since(start)

			p.Spacer(2)
			p.PrintText(" Complete! Imported ")
			p.PrintlnHighlight(fmt.Sprintf("%d benchmarks in %s", benchmarksImported, timeToImport.Round(time.Millisecond)))
			p.Spacer(2)

			return nil
		},
	}

	cmd.Flags().BoolVar(&opts.lax, "lax", false, "Less strict imports; allows to import duplicate data sources")

	cmd.Flags().SortFlags = false

	return cmd
}

func importBenchmarksFromFile(ctx context.Context, db database.Store, path string) (int, error) {
	// Open file
	file, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	data, err := importer.FromJSON[[]*models.Benchmark](file)
	if err != nil {
		return 0, fmt.Errorf("read data file: %w", err)
	}

	// Get number of records and check if we have any
	records := len(data)
	if records == 0 {
		return 0, fmt.Errorf("no benchmarks found in file")
	}

	// Save benchmarks
	_, err = db.SaveMany(ctx, data)
	if err != nil {
		return 0, fmt.Errorf("save benchmarks: %w", err)
	}

	return records, nil
}

func extractArchive(ctx context.Context, sourcePath, targetPath string) error {
	// Open file
	file, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer file.Close()

	// Extract archive
	return archive.Extract(ctx, file, targetPath)
}

var importLongDesc = `
Import a dbench data directory or file into the database. This command will import
all JSON files in the given directory or file.

Note: We only support JSON as it's far more consistent and robust to import from.
We only support the CSV exports for the sake of compatibility with other tools.`
