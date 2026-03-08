package ui

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

// Confirm asks the user to confirm an action by typing the expected text.
func Confirm(message, expected string) (bool, error) {
	var input string
	err := huh.NewInput().
		Title(message).
		Placeholder(expected).
		Value(&input).
		Run()
	if err != nil {
		return false, fmt.Errorf("confirmation prompt: %w", err)
	}
	return input == expected, nil
}

// ConfirmYesNo asks a simple yes/no question.
func ConfirmYesNo(message string) (bool, error) {
	var confirmed bool
	err := huh.NewConfirm().
		Title(message).
		Value(&confirmed).
		Run()
	if err != nil {
		return false, fmt.Errorf("confirmation prompt: %w", err)
	}
	return confirmed, nil
}

// SelectString presents a selection list.
func SelectString(title string, options []string) (string, error) {
	var selected string
	opts := make([]huh.Option[string], len(options))
	for i, opt := range options {
		opts[i] = huh.NewOption(opt, opt)
	}
	err := huh.NewSelect[string]().
		Title(title).
		Options(opts...).
		Value(&selected).
		Run()
	if err != nil {
		return "", fmt.Errorf("selection prompt: %w", err)
	}
	return selected, nil
}

// InputString asks for a string input.
func InputString(title, placeholder string) (string, error) {
	var value string
	err := huh.NewInput().
		Title(title).
		Placeholder(placeholder).
		Value(&value).
		Run()
	if err != nil {
		return "", fmt.Errorf("input prompt: %w", err)
	}
	return value, nil
}

// MultiSelect presents a multi-select list.
func MultiSelect(title string, options []string) ([]string, error) {
	var selected []string
	opts := make([]huh.Option[string], len(options))
	for i, opt := range options {
		opts[i] = huh.NewOption(opt, opt)
	}
	err := huh.NewMultiSelect[string]().
		Title(title).
		Options(opts...).
		Value(&selected).
		Run()
	if err != nil {
		return nil, fmt.Errorf("multi-select prompt: %w", err)
	}
	return selected, nil
}
