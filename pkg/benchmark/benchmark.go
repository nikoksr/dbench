package benchmark

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"

	"github.com/nikoksr/dbench/ent/schema/duration"
	"github.com/nikoksr/dbench/pkg/models"
	"github.com/nikoksr/dbench/pkg/probing"
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
func ParseOutput(output string) (*models.Benchmark, error) {
	lines := strings.Split(output, "\n")
	benchmark := &models.Benchmark{}
	result := &models.BenchmarkResult{}

	for _, line := range lines {
		fields := strings.Fields(line) // Splits the line into words
		if len(fields) == 0 {
			continue
		}

		switch {
		case strings.HasPrefix(line, "pgbench ("):
			benchmark.Version = strings.Join(fields[1:], " ")
			benchmark.Version = strings.TrimPrefix(benchmark.Version, "(")
			benchmark.Version = strings.TrimSuffix(benchmark.Version, ")")
		case strings.HasPrefix(line, "transaction type:"):
			benchmark.TransactionType = strings.Join(fields[2:], " ")
			benchmark.TransactionType = strings.TrimPrefix(benchmark.TransactionType, "<")
			benchmark.TransactionType = strings.TrimSuffix(benchmark.TransactionType, ">")
		case strings.HasPrefix(line, "scaling factor:"):
			sf, err := strconv.ParseFloat(fields[2], 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse scaling factor: %w", err)
			}
			benchmark.ScalingFactor = sf
		case strings.HasPrefix(line, "query mode:"):
			benchmark.QueryMode = fields[2]
		case strings.HasPrefix(line, "number of clients:"):
			n, err := strconv.Atoi(fields[3])
			if err != nil {
				return nil, fmt.Errorf("failed to parse number of clients: %w", err)
			}
			benchmark.Clients = n

		case strings.HasPrefix(line, "number of threads:"):
			n, err := strconv.Atoi(fields[3])
			if err != nil {
				return nil, fmt.Errorf("failed to parse number of threads: %w", err)
			}
			benchmark.Threads = n

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
			result.ConnectionTime = duration.Duration(connTime)

		case strings.HasPrefix(line, "tps ="):
			tps, err := strconv.ParseFloat(fields[2], 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse tps: %w", err)
			}
			result.TransactionsPerSecond = tps
		}
	}

	// Add the result to the benchmark, and finally return it
	benchmark.Edges.Result = result

	return benchmark, nil
}

func roundToTwoDecimals(f float64) float64 {
	return float64(int(f*100)) / 100
}

// Run executes the pgbench command with the provided configuration
func Run(ctx context.Context, config *models.BenchmarkConfig) (*models.Benchmark, error) {
	// Fill the config with default values
	config.Sanitize()

	// Evaluate benchmark config parameters based on the mode
	var (
		totalBenchmarkDuration string
		samplingRate           time.Duration
		predictedSamplesCount  int
	)

	switch config.Mode {
	case models.ModeSimple:
		// Simple mode runs for 5 seconds total, so we probe every second
		totalBenchmarkDuration = "5"
		samplingRate = 1 * time.Second

		// Predict the number of samples we will get
		// 5 seconds / 1 second = 5 samples
		predictedSamplesCount = 5

	case models.ModeThorough:
		// Thorough mode runs for 10 minutes total, so we probe every 30 seconds
		totalBenchmarkDuration = "600"
		samplingRate = 30 * time.Second

		// Predict the number of samples we will get
		// 10 minutes / 30 seconds = 20 samples
		predictedSamplesCount = 20
	default:
		return nil, fmt.Errorf("unknown benchmark mode: %q", config.Mode)
	}

	// Create errgoup to monitor the system while the benchmark is running
	eg, ctx := errgroup.WithContext(ctx)

	// Create pgbench command
	cmd := exec.CommandContext(ctx, "pgbench",
		// Database config
		"-U", config.Username,
		"-p", config.Port,
		"-h", config.Host,
		// Benchmark config
		"-M", "extended",
		"--vacuum-all",
		"-j", strconv.Itoa(config.NumThreads),
		"-c", strconv.Itoa(config.NumClients),
		"-T", totalBenchmarkDuration,
		// Database name is expected as the last argument
		config.DBName,
	)

	// Create buffers to capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Start system monitoring
	stopChan := make(chan struct{})
	systemSampleChan := make(chan models.SystemSample)

	// Start monitoring the system
	eg.Go(func() error {
		return probing.MonitorSystem(samplingRate, stopChan, systemSampleChan)
	})

	// Handle system metrics
	cpuLoad := make([]float64, 0, predictedSamplesCount)
	totalCPULoad := 0.0

	memoryUsage := make([]float64, 0, predictedSamplesCount)
	totalMemoryUsage := 0.0

	go func() {
		for sample := range systemSampleChan {
			cpuLoad = append(cpuLoad, sample.CPULoad)
			totalCPULoad += sample.CPULoad

			memoryUsage = append(memoryUsage, sample.MemoryLoad)
			totalMemoryUsage += sample.MemoryLoad
		}
	}()

	// Execute pgbench
	fmt.Printf("  > Running: %s\n", cmd.String())
	eg.Go(func() error {
		err := cmd.Run()
		close(stopChan) // Stop system monitoring
		return err
	})

	// Wait for the group to finish
	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("%w: %s", err, stderr.String())
	}

	// Parse the pgbench output
	output := stdout.String()
	benchmark, err := ParseOutput(output)
	if err != nil {
		return nil, fmt.Errorf("parse pgbench output: %w", err)
	}

	// Sort CPU and memory usage slices, so we can calculate the median and percentiles
	slices.Sort(cpuLoad)
	slices.Sort(memoryUsage)

	// Calculate the median and percentiles
	metrics := new(models.SystemMetric)

	totalCPUSamples := float64(len(cpuLoad))
	metrics.CPUMinLoad = roundToTwoDecimals(cpuLoad[0])
	metrics.CPUMaxLoad = roundToTwoDecimals(cpuLoad[len(cpuLoad)-1])
	metrics.CPUAverageLoad = roundToTwoDecimals(totalCPULoad / totalCPUSamples)
	metrics.CPU50thLoad = roundToTwoDecimals(cpuLoad[int(totalCPUSamples*0.50)])
	metrics.CPU75thLoad = roundToTwoDecimals(cpuLoad[int(totalCPUSamples*0.75)])
	metrics.CPU90thLoad = roundToTwoDecimals(cpuLoad[int(totalCPUSamples*0.90)])
	metrics.CPU95thLoad = roundToTwoDecimals(cpuLoad[int(totalCPUSamples*0.95)])
	metrics.CPU99thLoad = roundToTwoDecimals(cpuLoad[int(totalCPUSamples*0.99)])

	totalMemorySamples := float64(len(memoryUsage)) // We could use cpuSamplesCount here, but just to be safe
	metrics.MemoryMinLoad = roundToTwoDecimals(memoryUsage[0])
	metrics.MemoryMaxLoad = roundToTwoDecimals(memoryUsage[len(memoryUsage)-1])
	metrics.MemoryAverageLoad = roundToTwoDecimals(totalMemoryUsage / totalMemorySamples)
	metrics.Memory50thLoad = roundToTwoDecimals(memoryUsage[int(totalMemorySamples*0.50)])
	metrics.Memory75thLoad = roundToTwoDecimals(memoryUsage[int(totalMemorySamples*0.75)])
	metrics.Memory90thLoad = roundToTwoDecimals(memoryUsage[int(totalMemorySamples*0.90)])
	metrics.Memory95thLoad = roundToTwoDecimals(memoryUsage[int(totalMemorySamples*0.95)])
	metrics.Memory99thLoad = roundToTwoDecimals(memoryUsage[int(totalMemorySamples*0.99)])

	// Add the missing pieces to the benchmark

	// Store meta information
	benchmark.Comment = config.Comment
	benchmark.Command = cmd.String()

	// Store the system metrics
	benchmark.Edges.SystemMetric = metrics

	return benchmark, nil
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
