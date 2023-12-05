package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/nikoksr/dbench/internal/ui/styles"
)

// Prompt is a component that renders a text input field with a prompt.
type Prompt struct {
	model *promptModel
}

func countLeadingChars(text string, char byte) int {
	count := 0
	for i := 0; i < len(text); i++ {
		if text[i] == char {
			count++
		} else {
			break
		}
	}

	return count
}

// NewPrompt returns a new Prompt component with the given text and placeholder.
func NewPrompt(text, placeholder string, sensitive bool) *Prompt {
	// Calculate the indent for the prompt
	indent := countLeadingChars(text, ' ')

	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 64
	ti.Width = 64
	ti.TextStyle = styles.Hint()
	ti.PromptStyle = styles.Info().Italic(false)
	ti.Focus()
	ti.Prompt = strings.Repeat(" ", indent) + "> "

	if sensitive {
		ti.EchoMode = textinput.EchoPassword
	}

	return &Prompt{
		model: &promptModel{
			textInput: ti,
			text:      styles.Info().Render(text),
		},
	}
}

// Render renders the component to the terminal.
func (p *Prompt) Render() error {
	_, err := tea.
		NewProgram(p.model).
		Run()

	return err
}

// WasCanceled returns true if the prompt was canceled.
func (p *Prompt) WasCanceled() bool {
	return p.model.wasCanceled
}

// Value returns the value of the prompt.
func (p *Prompt) Value() string {
	return p.model.textInput.Value()
}

var _ tea.Model = &promptModel{}

type promptModel struct {
	textInput   textinput.Model
	err         error
	wasCanceled bool

	text string
}

func (m *promptModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *promptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.wasCanceled = true
			return m, tea.Quit
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)

	return m, cmd
}

func (m *promptModel) View() string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		styles.Text().Render(m.text),
		m.textInput.View(),
		"",
	)
}
