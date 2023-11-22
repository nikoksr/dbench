package ui

import "strings"

func TextBox(text string) string {
	separator := HorizontalSeparator(false)
	return strings.Join([]string{separator, text, separator}, "\n\n")
}
