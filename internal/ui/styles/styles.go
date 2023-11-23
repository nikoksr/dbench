package styles

import "github.com/charmbracelet/lipgloss"

var (
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#f8f8f2")).
		Border(lipgloss.RoundedBorder(), false, false, true, false).
		Margin(2, 0, 1, 0)

	SubTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#f8f8f2"))

	Text = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#f8f8f2"))

	Hint = lipgloss.NewStyle().
		Faint(true)

	Success = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#50fa7b"))

	Error = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ff5555"))

	Info = lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.Color("#abb2bf"))
)
