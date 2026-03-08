package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	boldStyle    = lipgloss.NewStyle().Bold(true)
)

func Success(msg string) {
	fmt.Fprintln(os.Stderr, successStyle.Render("  "+msg))
}

func Error(msg string) {
	fmt.Fprintln(os.Stderr, errorStyle.Render("  "+msg))
}

func Warn(msg string) {
	fmt.Fprintln(os.Stderr, warnStyle.Render("  "+msg))
}

func Info(msg string) {
	fmt.Fprintln(os.Stderr, infoStyle.Render("  "+msg))
}

func Bold(msg string) string {
	return boldStyle.Render(msg)
}

func Step(msg string) {
	fmt.Fprintf(os.Stderr, "  %s\n", msg)
}
