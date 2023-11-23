package text

import "fmt"

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
