package plot

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	gnuplot2 "github.com/nikoksr/dbench/internal/plot/gnuplot"
)

type plotData struct {
	DataPath   string
	OutputPath string
}

func executeScriptTemplate(name, text string, data plotData) (string, error) {
	tmpl, err := template.New(name).Parse(text)
	if err != nil {
		return "", fmt.Errorf("parse gnuplot script template: %w", err)
	}

	var scriptBuilder strings.Builder
	err = tmpl.Execute(&scriptBuilder, data)
	if err != nil {
		return "", err
	}

	return scriptBuilder.String(), nil
}

func Plot(ctx context.Context, dataFile, outputDir string) error {
	for name, template := range gnuplot2.ScriptTemplates {
		// outputPath is the template output dir + the template name + .png
		outputPath := filepath.Join(outputDir, name+".png")

		data := plotData{
			DataPath:   dataFile,
			OutputPath: outputPath,
		}

		// Execute the template and get the script
		script, err := executeScriptTemplate(name, template, data)
		if err != nil {
			return fmt.Errorf("execute gnuplot script template %q: %w", name, err)
		}

		if err := gnuplot2.ExecuteScript(ctx, script); err != nil {
			return fmt.Errorf("execute gnuplot script %q: %w", name, err)
		}
	}

	return nil
}
