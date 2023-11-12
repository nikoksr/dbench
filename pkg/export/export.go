// Package export provides functionality to export data in various formats.
package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/nikoksr/dbench/pkg/models"
)

func openFile(filename string) (*os.File, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// ToCSV exports a slice of Result structs to a CSV file.
func ToCSV(results []*models.Result, filename string) error {
	file, err := openFile(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"ID",
		"GroupID",
		"TransactionType",
		"ScalingFactor",
		"QueryMode",
		"Clients",
		"Threads",
		"Transactions",
		"TransactionsPerSecond",
		"TransactionsPerClient",
		"FailedTransactions",
		"AverageLatency",
		"InitialConnectionTime",
		"TotalRuntime",
		"Version",
		"Command",
		"CreatedAt",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data rows
	for _, result := range results {
		record := []string{
			result.ID.String(),
			result.GroupID.String(),
			result.TransactionType,
			strconv.FormatFloat(result.ScalingFactor, 'f', 6, 64),
			result.QueryMode,
			strconv.Itoa(result.Clients),
			strconv.Itoa(result.Threads),
			strconv.Itoa(result.Transactions),
			strconv.FormatFloat(result.TransactionsPerSecond, 'f', 6, 64),
			strconv.Itoa(result.FailedTransactions),
			result.AverageLatency.String(),
			result.InitialConnectionTime.String(),
			result.TotalRuntime.String(),
			result.Version,
			result.Command,
			result.CreatedAt.String(),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// ToJSON takes an interface and attempts to marshal it into JSON format, then write to a file.
func ToJSON(data any, filename string) error {
	file, err := openFile(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(data)
}

// ToGnuplotBasic exports a slice of Result structs to a Gnuplot compatible .dat file.
func ToGnuplotBasic(results []*models.Result, filename string) error {
	file, err := openFile(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, result := range results {
		line := fmt.Sprintf("%d %f %d %d\n",
			result.Clients,
			result.TransactionsPerSecond,
			time.Duration(result.AverageLatency).Milliseconds(),
			time.Duration(result.InitialConnectionTime).Milliseconds(),
		)
		_, err := fmt.Fprint(file, line)
		if err != nil {
			return fmt.Errorf("write to file: %w", err)
		}
	}

	return nil
}

// ToGnuplot exports a slice of Result structs to a Gnuplot compatible .dat file.
func ToGnuplot(results []*models.Result, filename string) error {
	file, err := openFile(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, result := range results {
		line := fmt.Sprintf(
			"%s %s %s %f %s %d %d %d %f %d %d %s %s %s %s %s %s\n",
			result.ID,
			result.GroupID,
			result.TransactionType,
			result.ScalingFactor,
			result.QueryMode,
			result.Clients,
			result.Threads,
			result.Transactions,
			result.TransactionsPerSecond,
			result.FailedTransactions,
			result.AverageLatency.String(),
			result.InitialConnectionTime.String(),
			result.TotalRuntime.String(),
			result.Version,
			result.Command,
			result.CreatedAt.String(),
		)
		_, err := fmt.Fprint(file, line)
		if err != nil {
			return fmt.Errorf("write to file: %w", err)
		}
	}

	return nil
}
