package ui

import "strings"

const (
	separatorLength   = 80
	separatorChar     = "="
	separatorCharThin = "-"
)

func HorizontalSeparator(thin bool) string {
	if thin {
		return strings.Repeat(separatorCharThin, separatorLength)
	}

	return strings.Repeat(separatorChar, separatorLength)
}
