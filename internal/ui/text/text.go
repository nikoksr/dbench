package text

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/nikoksr/dbench/internal/ui/styles"
)

const naChar = "-"

// ValueOrNA returns the value of the given pointer or "N/A" if the pointer is nil.
func ValueOrNA[T any](v *T) string {
	if v == nil {
		return naChar
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

// Obfuscate returns a string with all but the first two and last four characters replaced by asterisks. The asterisks
// in the middle get shrunken to only a few characters.
func Obfuscate(s string) string {
	if len(s) < 6 {
		return s
	}

	var obfuscated strings.Builder
	obfuscated.WriteString(s[:6])
	obfuscated.WriteString(strings.Repeat("*", 3))
	obfuscated.WriteString(s[len(s)-4:])

	return obfuscated.String()
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
	bl.WriteString(style.Inline(true).Render(header))

	// Add items
	itemStyle := styles.Text().Inline(true)
	for idx, err := range items {
		if idx == 0 {
			bl.WriteString(" ") // First item has single space after header
		} else {
			bl.WriteString(indent) // Fully indent all other items
		}

		bl.WriteString(itemStyle.Render("• "))
		bl.WriteString(itemStyle.Render(err.Error()))
		bl.WriteString("\n")
	}

	// Final newline for readability
	bl.WriteString("\n")

	return bl.String()
}

// WarningsList returns a bullet list of the given items.
func WarningsList(items []error) string {
	return bulletList(" Warnings", styles.Warn(), items)
}

// BulletList returns a bullet list of the given items.
func BulletList(header string, style lipgloss.Style, items []error) string {
	return bulletList(header, style, items)
}

// PrettyTime returns a human-readable string for the given time.
func PrettyTime(t time.Time) string {
	return t.Local().Format("2006-01-02 15:04:05")
}
