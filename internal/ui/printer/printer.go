package printer

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Printer is a struct that contains a writer and an indent level. It is used to print messages to the terminal.
type Printer struct {
	w io.Writer

	indent int
}

// NewPrinter creates a new Printer with the provided writer and optional indent level.
// If no indent level is provided, a default of 50 is used.
// If the provided indent level is negative, indentation is disabled.
func NewPrinter(w io.Writer, indent ...int) *Printer {
	p := &Printer{
		w:      w,
		indent: 50,
	}

	if len(indent) <= 0 {
		return p // No custom indent, use default
	}

	if indent[0] < 0 {
		p.indent = 0 // Disable indentation
	} else {
		p.indent = indent[0] // Use custom indent
	}

	return p
}

// Spacer prints a number of newlines to the printer's writer.
func (p *Printer) Spacer(n int) {
	for i := 0; i < n; i++ {
		_, _ = fmt.Fprintln(p.w)
	}
}

// printStyle is a struct that contains the style information for a print operation.
type printStyle struct {
	// General
	baseStyle lipgloss.Style

	indent int

	// Prefix
	prefixStyle  *lipgloss.Style // We could use a recursive printStyle here, probably overkill tho
	prefixSymbol string
	prefixText   string

	// Text
	text    string
	newLine bool
}

// PrintOption is a function that modifies a printStyle.
type PrintOption func(*printStyle)

// WithIndent returns a PrintOption that sets the indent level of a printStyle.
// If no level is provided, the printer's default is used.
// If the provided level is negative, indentation is disabled.
func WithIndent(level ...int) PrintOption {
	return func(style *printStyle) {
		if len(level) <= 0 {
			style.indent = -1 // Use printer default
			return
		}

		if level[0] < 0 {
			style.indent = 0 // Disable indentation
			return
		}

		style.indent = level[0] // Use custom indent
	}
}

// WithParagraphMode returns a PrintOption that sets the indent level of a printStyle to 80 and aligns the text to the
// left with a margin of 2.
func WithParagraphMode() PrintOption {
	return func(style *printStyle) {
		style.indent = 100 // Will be handled by "printer.printWithStyle"
		style.baseStyle = style.baseStyle.
			Align(lipgloss.Left).
			MarginLeft(2)
	}
}

// printWithStyle prints a message to the printer's writer with the provided style and options.
func (p *Printer) printWithStyle(style *printStyle, options ...PrintOption) {
	for _, option := range options {
		option(style)
	}

	// A negative indent means we should use the printer default
	if style.indent <= -1 {
		style.indent = p.indent
	}

	style.baseStyle = style.baseStyle.
		Width(style.indent)

	var msg strings.Builder

	if style.prefixSymbol != "" {
		msg.WriteString(style.prefixStyle.Render(style.prefixSymbol))
	}

	if style.prefixText != "" {
		msg.WriteString(style.prefixStyle.Render(style.prefixText))
	}

	msg.WriteString(style.baseStyle.Render(style.text))

	if style.newLine {
		_, _ = fmt.Fprintln(p.w, msg.String())
	} else {
		_, _ = fmt.Fprint(p.w, msg.String())
	}
}

// Title

// PrintlnTitle prints a title to the printer's writer with a newline at the end.
func (p *Printer) PrintlnTitle(title string, opts ...PrintOption) {
	style := newTitleStyle()
	style.text = title

	p.printWithStyle(style, opts...)
}

// PrintlnSubTitle prints a subtitle to the printer's writer with a newline at the end.
func (p *Printer) PrintlnSubTitle(title string, opts ...PrintOption) {
	style := newSubTitleStyle()
	style.text = title

	p.printWithStyle(style, opts...)
}

// General messages

// Text

// PrintText prints a text message to the printer's writer.
func (p *Printer) PrintText(msg string, opts ...PrintOption) {
	style := newTextStyle()
	style.text = msg

	p.printWithStyle(style, opts...)
}

// PrintlnText prints a text message to the printer's writer with a newline at the end.
func (p *Printer) PrintlnText(msg string, opts ...PrintOption) {
	style := newTextStyle()
	style.text = msg
	style.newLine = true

	p.printWithStyle(style, opts...)
}

// Hint

// PrintHint prints a hint message to the printer's writer.
func (p *Printer) PrintHint(msg string, opts ...PrintOption) {
	style := newHintStyle()
	style.text = msg

	p.printWithStyle(style, opts...)
}

func (p *Printer) PrintlnHint(msg string, opts ...PrintOption) {
	style := newHintStyle()
	style.text = msg
	style.newLine = true

	p.printWithStyle(style, opts...)
}

// Info

// PrintInfo prints an info message to the printer's writer.
func (p *Printer) PrintInfo(msg string, opts ...PrintOption) {
	style := newInfoStyle()
	style.text = msg

	p.printWithStyle(style, opts...)
}

// PrintlnInfo prints an info message to the printer's writer with a newline at the end.
func (p *Printer) PrintlnInfo(msg string, opts ...PrintOption) {
	style := newInfoStyle()
	style.text = msg
	style.newLine = true

	p.printWithStyle(style, opts...)
}

// Highlight

// PrintHighlight prints a highlighted message to the printer's writer.
func (p *Printer) PrintHighlight(msg string, opts ...PrintOption) {
	style := newHighlightStyle()
	style.text = msg

	p.printWithStyle(style, opts...)
}

// PrintlnHighlight prints a highlighted message to the printer's writer with a newline at the end.
func (p *Printer) PrintlnHighlight(msg string, opts ...PrintOption) {
	style := newHighlightStyle()
	style.text = msg
	style.newLine = true

	p.printWithStyle(style, opts...)
}

// Result-type messages; usually indented to the right by a previous message

// Success

// PrintlnSuccess prints a success message to the printer's writer with a newline at the end.
func (p *Printer) PrintlnSuccess(msg string, opts ...PrintOption) {
	style := newSuccessStyle()
	style.text = msg
	style.newLine = true

	p.printWithStyle(style, opts...)
}

// Warning

// PrintlnWarning prints a warning message to the printer's writer with a newline at the end.
func (p *Printer) PrintlnWarning(msg string, opts ...PrintOption) {
	style := newWarningStyle()
	style.text = msg
	style.newLine = true

	p.printWithStyle(style, opts...)
}

// Error

// PrintlnError prints an error message to the printer's writer with a newline at the end.
func (p *Printer) PrintlnError(msg string, opts ...PrintOption) {
	style := newErrorStyle()
	style.text = msg
	style.newLine = true

	p.printWithStyle(style, opts...)
}
