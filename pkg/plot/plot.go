package plot

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const basicPlot = `
set terminal pngcairo size 800,600 enhanced font 'Verdana,10'
set output '{{ .OutputPath }}'
set xlabel 'Number of Clients'
set ylabel 'Transactions per Second'
set y2label 'Average Latency / Connection Time (ms)'
set ytics nomirror
set y2tics
set key outside top center
set tmargin 5
set format y2 "%.0fms"
set autoscale y   # Enable autoscaling for the primary y-axis
set autoscale y2  # Enable autoscaling for the secondary y-axis
plot '{{ .DataPath }}' using 1:2 with linespoints title 'Transactions per Second', \
     '{{ .DataPath }}' using 1:3 axes x1y2 with linespoints title 'Average Latency', \
     '{{ .DataPath }}' using 1:4 axes x1y2 with linespoints title 'Connection Time'
`

type plotData struct {
	DataPath   string
	OutputPath string
}

// Function to execute the template with dynamic data
func generatePlotScript(dataPath, outputPath string) (string, error) {
	// Create an instance of PlotData with the desired output name
	data := plotData{
		DataPath:   dataPath,
		OutputPath: outputPath,
	}

	// Parse the template
	tmpl, err := template.New("plot").Parse(basicPlot)
	if err != nil {
		return "", err
	}

	// Execute the template with the PlotData struct
	var scriptBuilder strings.Builder
	err = tmpl.Execute(&scriptBuilder, data)
	if err != nil {
		return "", err
	}

	// Return the executed script
	return scriptBuilder.String(), nil
}

func Plot(dataPath, outputName string) (string, error) {
	// Sanitize the output name
	// Remove the file extension (could be any) if it exists
	outputName = strings.TrimSuffix(outputName, filepath.Ext(outputName))
	outputPath := outputName + ".png"

	// Generate the gnuplot script
	script, err := generatePlotScript(dataPath, outputPath)
	if err != nil {
		return "", fmt.Errorf("generate gnuplot script: %w", err)
	}

	// Create a gnuplot command
	cmd := exec.Command("gnuplot")

	cmd.Stderr = os.Stderr

	// Create a pipe to the standard input of the command
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return "", err
	}

	// Write the script to the standard input of the gnuplot command
	_, err = stdinPipe.Write([]byte(script))
	if err != nil {
		return "", err
	}
	stdinPipe.Close()

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		return "", err
	}

	return outputPath, nil
}
