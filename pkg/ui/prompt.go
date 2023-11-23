package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/nikoksr/dbench/pkg/styles"
)

var _ Component = &Prompt{}

// Prompt is a component that renders a text input field with a prompt.
type Prompt struct {
	model *promptModel
}

// NewPrompt returns a new Prompt component with the given text and placeholder.
func NewPrompt(text, placeholder string, sensitive bool) *Prompt {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 256
	ti.Width = 50
	ti.PromptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#f8f8f2"))
	ti.Focus()

	if sensitive {
		ti.EchoMode = textinput.EchoPassword
	}

	return &Prompt{
		model: &promptModel{
			textInput: ti,
			text:      text,
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
		styles.Text.Render(m.text),
		m.textInput.View(),
		"",
	)
}
