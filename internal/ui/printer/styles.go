package printer

import (
	"github.com/nikoksr/dbench/internal/pointer"
	"github.com/nikoksr/dbench/internal/ui/styles"
)

func newTitleStyle() *printStyle {
	return &printStyle{
		baseStyle: styles.Title(),
		newLine:   true, // titles always end with a newline
	}
}

func newSubTitleStyle() *printStyle {
	return &printStyle{
		baseStyle: styles.SubTitle(),
		newLine:   true, // subtitles always end with a newline
	}
}

func newTextStyle() *printStyle {
	return &printStyle{
		baseStyle: styles.Text(),
	}
}

func newHintStyle() *printStyle {
	return &printStyle{
		baseStyle: styles.Hint(),
	}
}

func newInfoStyle() *printStyle {
	return &printStyle{
		baseStyle: styles.Info(),
	}
}

func newHighlightStyle() *printStyle {
	return &printStyle{
		baseStyle: styles.Highlight(),
	}
}

func newSuccessStyle() *printStyle {
	return &printStyle{
		baseStyle:    styles.Info(),                // Style for "✓ Success [message]"
		prefixStyle:  pointer.To(styles.Success()), // Style for "[✓ Success] message"
		prefixSymbol: "✓ ",
		prefixText:   "Success ",
	}
}

func newWarningStyle() *printStyle {
	return &printStyle{
		baseStyle:    styles.Info(),             // Style for "⚠ Warning [message]"
		prefixStyle:  pointer.To(styles.Warn()), // Style for "[⚠ Warning] message"
		prefixSymbol: "⚠ ",
		prefixText:   "Warning ",
	}
}

func newErrorStyle() *printStyle {
	return &printStyle{
		baseStyle:    styles.Info(),              // Style for "✗ Error [message]"
		prefixStyle:  pointer.To(styles.Error()), // Style for "[✗ Error] message"
		prefixSymbol: "✗ ",
		prefixText:   "Error ",
	}
}
