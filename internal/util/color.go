package util

import "github.com/fatih/color"

var (
	Cyan      = color.New(color.FgCyan, color.Bold).SprintFunc()
	Highlight = color.New(color.FgCyan, color.Bold).SprintFunc()

	Green   = color.New(color.FgGreen, color.Bold).SprintFunc()
	Success = color.New(color.FgGreen, color.Bold).SprintFunc()

	Red = color.New(color.FgRed, color.Bold).SprintFunc()
	Err = color.New(color.FgRed, color.Bold).SprintFunc()

	Yellow = color.New(color.FgYellow, color.Bold).SprintFunc()
	Info   = color.New(color.FgYellow, color.Bold).SprintFunc()
)
