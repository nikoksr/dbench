package ui

import "strings"

const (
	separatorLength = 80
	separatorChar   = "="
)

func HorizontalSeparator() string {
	return strings.Repeat(separatorChar, separatorLength)
}
