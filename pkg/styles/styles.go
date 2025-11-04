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

func PrintFS(style string, text string) {
	switch style {
	case "error":
		fmt.Println(errorStyle.Render(text))
	case "success":
		fmt.Println(successStyle.Render(text))
	case "info":
		fmt.Println(infoStyle.Render(text))
	default:
		fmt.Println(defaultStyle.Render(text))
	}
}
