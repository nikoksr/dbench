package ui

type (
	// Component is a component that can be rendered to the terminal.
	Component interface {
		Render() error
	}

	errMsg error
)
