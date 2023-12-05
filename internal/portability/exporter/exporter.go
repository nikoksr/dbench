package exporter

import (
	"encoding/json"
	"io"

	"github.com/gocarina/gocsv"
)

// Exporter is a function that takes a writer and data and exports it to a file.
type Exporter func(w io.Writer, data any) error

// ToCSV exports a slice of Benchmark structs to a CSV file. An empty filename will create a temporary file.
func ToCSV(w io.Writer, data any) error {
	return gocsv.Marshal(data, w)
}

// ToJSON takes an interface and attempts to marshal it into JSON format, then write to a file. An empty filename will
// create a temporary file.
func ToJSON(w io.Writer, data any) error {
	return json.NewEncoder(w).Encode(data)
}
