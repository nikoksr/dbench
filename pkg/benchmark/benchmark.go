package benchmark

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/nikoksr/dbench/ent/schema/duration"
	"github.com/nikoksr/dbench/pkg/models"
)

// parseDuration converts a string like "5.359 ms" or "1.2 s" to a time.Duration.
func parseDuration(duration, unit string) time.Duration {
	if duration == "" {
		return 0
	}

	value, err := strconv.ParseFloat(duration, 64)
	if err != nil {
		return 0 // or log an error, or return an error
	}

	switch unit {
	case "us": // microsecond
		return time.Duration(value * float64(time.Microsecond))
	case "ms":
		return time.Duration(value * float64(time.Millisecond))
	case "s":
		return time.Duration(value * float64(time.Second))
	default:
		return 0 // or log an error, or return an error
	}
}

// ParseOutput parses the output of the pgbench command (always gets run with -M extended)
func ParseOutput(output string) (*models.Result, error) {
	lines := strings.Split(output, "\n")
	result := &models.Result{}

	for _, line := range lines {
		fields := strings.Fields(line) // Splits the line into words
		if len(fields) == 0 {
			continue
		}

		switch {
		case strings.HasPrefix(line, "pgbench ("):
			result.Version = strings.Join(fields[1:], " ")
			result.Version = strings.TrimPrefix(result.Version, "(")
			result.Version = strings.TrimSuffix(result.Version, ")")
		case strings.HasPrefix(line, "transaction type:"):
			result.TransactionType = strings.Join(fields[2:], " ")
			result.TransactionType = strings.TrimPrefix(result.TransactionType, "<")
			result.TransactionType = strings.TrimSuffix(result.TransactionType, ">")
		case strings.HasPrefix(line, "scaling factor:"):
			sf, err := strconv.ParseFloat(fields[2], 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse scaling factor: %w", err)
			}
			result.ScalingFactor = sf
		case strings.HasPrefix(line, "query mode:"):
			result.QueryMode = fields[2]
		case strings.HasPrefix(line, "number of clients:"):
			n, err := strconv.Atoi(fields[3])
			if err != nil {
				return nil, fmt.Errorf("failed to parse number of clients: %w", err)
			}
			result.Clients = n

		case strings.HasPrefix(line, "number of threads:"):
			n, err := strconv.Atoi(fields[3])
			if err != nil {
				return nil, fmt.Errorf("failed to parse number of threads: %w", err)
			}
			result.Threads = n

		case strings.HasPrefix(line, "number of transactions actually processed:"):
			n, err := strconv.Atoi(fields[5])
			if err != nil {
				return nil, fmt.Errorf("failed to parse number of transactions actually processed: %w", err)
			}
			result.Transactions = n

		case strings.HasPrefix(line, "number of failed transactions:"):
			fields := strings.Fields(line)
			n, err := strconv.Atoi(fields[4])
			if err != nil {
				return nil, fmt.Errorf("failed to parse number of failed transactions: %w", err)
			}
			result.FailedTransactions = n

		case strings.HasPrefix(line, "latency average ="):
			latency := parseDuration(fields[3], fields[4])
			result.AverageLatency = duration.Duration(latency)

		case strings.HasPrefix(line, "initial connection time ="):
			connTime := parseDuration(fields[4], fields[5])
			result.InitialConnectionTime = duration.Duration(connTime)

		case strings.HasPrefix(line, "tps ="):
			tps, err := strconv.ParseFloat(fields[2], 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse tps: %w", err)
			}
			result.TransactionsPerSecond = tps
		}
	}

	return result, nil
}

// Run executes the pgbench command with the provided configuration
func Run(config *models.BenchmarkConfig) (*models.Result, error) {
	var cmd *exec.Cmd

	// Fill the config with default values
	config.Sanitize()

	clients := strconv.Itoa(config.NumClients)

	// Construct the pgbench command
	switch config.Mode {
	case models.ModeSimple:
		cmd = exec.Command(
			"pgbench",
			// Connection settings
			"-U", config.Username,
			"-p", config.Port,
			"-h", config.Host,
			// Benchmark settings
			"-M", "extended",
			"-j", strconv.Itoa(config.NumThreads),
			"-c", clients,
			"-T", "5", // 5 seconds
			// Database settings
			config.DBName,
		)
	case models.ModeThorough:
		cmd = exec.Command(
			"pgbench",
			// Connection settings
			"-U", config.Username,
			"-p", config.Port,
			"-h", config.Host,
			// Benchmark settings
			"-M", "extended",
			"-j", strconv.Itoa(config.NumThreads),
			"-c", clients,
			"-T", "600", // 10 minutes
			// Database settings
			config.DBName,
		)
	default:
		return nil, fmt.Errorf("unknown benchmark mode: %q", config.Mode)
	}

	// Create buffers to capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute the command
	fmt.Printf("  > Running: %s\n", cmd.String())

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, stderr.String())
	}

	// Parse the pgbench output
	output := stdout.String()
	result, err := ParseOutput(output)
	if err != nil {
		return nil, fmt.Errorf("parse pgbench output: %w", err)
	}

	// Before returning, set the command that produced the result
	result.Command = cmd.String()

	return result, nil
}

// Init initializes a target database using pgbench
func Init(config *models.BenchmarkConfig) error {
	var cmd *exec.Cmd

	// Fill the config with default values
	config.Sanitize()

	// Construct the pgbench command
	cmd = exec.Command(
		"pgbench",
		// Connection settings
		"-U", config.Username,
		"-p", config.Port,
		"-h", config.Host,
		// Benchmark settings
		"-i",
		"-s", strconv.Itoa(config.ScaleFactor),
		"-F", strconv.Itoa(config.FillFactor),
		// Database settings
		config.DBName,
	)

	// Create buffers to capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute the command
	fmt.Printf("Running command: %s\n", cmd.String())

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%w: %s", err, stderr.String())
	}

	return nil
}
