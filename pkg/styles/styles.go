package styles

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

const fontSize = 16

var defaultStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#7D56F4"))

var errorStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#F45E6E"))

var successStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#6ef4a1ff"))

var infoStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#6EC4F4"))

func PrintFS(style string, text string, a ...interface{}) {
	text = fmt.Sprintf(text, a...)
	switch style {
	case "error":
		text = errorStyle.Render(text)
	case "success":
		text = successStyle.Render(text)
	case "info":
		text = infoStyle.Render(text)
	default:
		text = defaultStyle.Render(text)
	}
	fmt.Println(text)
}

func SprintfS(style string, format string, a ...interface{}) string {
	text := fmt.Sprintf(format, a...)
	switch style {
	case "error":
		text = errorStyle.Render(text)
	case "success":
		text = successStyle.Render(text)
	case "info":
		text = infoStyle.Render(text)
	default:
		text = defaultStyle.Render(text)
	}
	return text
}
