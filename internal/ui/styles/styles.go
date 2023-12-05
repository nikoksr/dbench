package styles

import "github.com/charmbracelet/lipgloss"

func Title() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#f8f8f2")).
		Border(lipgloss.RoundedBorder(), false, false, true, false).
		Margin(2, 0, 1, 0)
}

func SubTitle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#f8f8f2"))
}

func Text() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#f8f8f2"))
}

func Hint() lipgloss.Style {
	return lipgloss.NewStyle().
		// Light gray
		Foreground(lipgloss.Color("#abb2bf")).
		Faint(true)
}

func Highlight() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true)
}

func Success() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#50fa7b"))
}

func Error() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ff5555"))
}

func Warn() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFCC66"))
}

func Info() lipgloss.Style {
	return lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.Color("#abb2bf"))
}
