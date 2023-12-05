package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nikoksr/dbench/internal/ui"
	"github.com/nikoksr/dbench/internal/ui/printer"
)

var (
	errNoPassword = errors.New(`no password provided

You can either enter a password or set the PGPASSWORD environment variable:

	# Example
	export PGPASSWORD=supersecret

For more information, see the official documentation:
https://www.postgresql.org/docs/current/libpq-envars.html`)

	errPgbenchNotInstalled = errors.New(`pgbench is required to run the application.

For more information, see the official documentation:
https://www.postgresql.org/docs/current/pgbench.html`)

	errGNUPlotNotInstalled = fmt.Errorf(`gnuplot is required to run the application.

For more information, see the official documentation:
http://www.gnuplot.info/`)
)

func isPathADirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.IsDir()
}

func getDBPassword(p *printer.Printer) (string, bool, error) {
	// Check if PGPASSWORD is set
	passwd := os.Getenv("PGPASSWORD")

	if passwd != "" {
		p.PrintlnHint("Detected PGPASSWORD - leave the following prompt empty to use it.")
		p.Spacer(1)
	}

	// Prompt for password
	prompt := ui.NewPrompt(" Enter database password:", "Password", true)
	if err := prompt.Render(); err != nil {
		return "", false, err
	}

	// If the user canceled the prompt, return. Signal that we don't want to continue.
	if prompt.WasCanceled() {
		return "", true, nil
	}

	// If the user entered a password, return it
	if prompt.Value() != "" {
		return prompt.Value(), false, nil
	}

	// No password entered, if PGPASSWORD is set, return it
	if passwd != "" {
		return passwd, false, nil
	}

	// No password entered and PGPASSWORD is not set, return an error
	return "", false, errNoPassword
}

func isToolInPath(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}

func getToolVersion(tool string) (string, error) {
	cmd := exec.Command(tool, "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	// Remove all trailling spaces and newlines
	version := string(out)
	version = strings.Trim(version, "\n")
	version = strings.TrimSpace(version)

	if tool == "pgbench" {
		version = strings.TrimPrefix(version, "pgbench (PostgreSQL) ")
	} else if tool == "gnuplot" {
		version = strings.TrimPrefix(version, "gnuplot ")
	}

	return version, nil
}

const (
	defaultBatchSize = 5_000
	maxBatchSize     = 25_000
)

func sanitizeBatchSize(batchSize int) int {
	if batchSize <= 0 {
		return defaultBatchSize
	}

	if batchSize > maxBatchSize {
		return maxBatchSize
	}

	return batchSize
}
