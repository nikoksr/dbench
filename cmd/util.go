package cmd

import (
	"errors"
	"fmt"
	"github.com/nikoksr/dbench/internal/ui/styles"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nikoksr/dbench/internal/ui"
)

var (
	errNoPassword = errors.New(`no password provided

You can either enter a password or set the PGPASSWORD environment variable:

	# Example
	export PGPASSWORD=supersecret

For more information, see the official documentation:
https://www.postgresql.org/docs/current/libpq-envars.html
`)

	errPgbenchNotInstalled = errors.New(`pgbench is required to run the application. It can be installed with the following command:

	# Arch
	sudo pacman -S postgresql

	# Debian / Ubuntu
	sudo apt install postgresql-client

	# macOS
	brew install postgresql

For more information, see the official documentation:
https://www.postgresql.org/docs/current/pgbench.html
`)

	gnuPlotNotInstalledErr = fmt.Errorf(`gnuplot is required to run the application. It can be installed with the following command:

	# Arch
	sudo pacman -S gnuplot

	# Debian / Ubuntu
	sudo apt install gnuplot

	# macOS
	brew install gnuplot

For more information, see the official documentation:
http://www.gnuplot.info/
`)
)

func prepareDirectory(dir string) error {
	// Clean directory path
	dir = filepath.Clean(dir)

	// Create output directory if it doesn't exist
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create output directory: %w", err)
		}
	}

	return nil
}

func getDBPassword() (string, bool, error) {
	// Check if PGPASSWORD is set
	passwd := os.Getenv("PGPASSWORD")

	if passwd != "" {
		fmt.Printf("%s\n\n", styles.Hint.Render("Detected PGPASSWORD - leave the following prompt empty to use it."))
	}

	// Prompt for password
	prompt := ui.NewPrompt("Enter database password:", "Password", true)
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
