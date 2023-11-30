package text

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/nikoksr/dbench/internal/ui/styles"
	"strings"
)

// ValueOrNA returns the value of the given pointer or "N/A" if the pointer is nil.
func ValueOrNA[T any](v *T) string {
	if v == nil {
		return "-"
	}

	return fmt.Sprintf("%v", *v)
}

// HumanizeBytes returns a human-readable string for the given bytes. If the given pointer is nil, "N/A" is returned.
func HumanizeBytes(bytes *uint64) string {
	if bytes == nil {
		return "-"
	}

	const unit = 1024
	if *bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := uint64(unit), 0
	for n := *bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %ciB", float64(*bytes)/float64(div), "KMGTPE"[exp])
}

func bulletList(header string, style lipgloss.Style, items []error) string {
	if len(items) == 0 {
		return ""
	}

	// Prepare header and calculate indentation
	if !strings.HasSuffix(header, ":") {
		header += ":"
	}

	indent := strings.Repeat(" ", len(header)+1)

	// Build bullet list
	var bl strings.Builder
	bl.WriteString("\n" + style.Render(header) + "\n")

	// Add items
	for _, item := range items {
		bl.WriteString(indent + fmt.Sprintf("- %s", item) + "\n")
	}

	// Final newline for readability
	bl.WriteString("\n")

	return bl.String()
}

// WarningsList returns a bullet list of the given items.
func WarningsList(items []error) string {
	return bulletList("Warnings", styles.Error, items)
}
