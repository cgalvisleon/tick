// Package ui renders status pills and percent bars using et/color, so tick's
// terminal output reuses the same ANSI helpers the rest of the workspace uses.
package ui

import (
	"strconv"
	"strings"

	"github.com/cgalvisleon/et/color"
)

/**
* StatusLabel: Returns the status value colored according to its meaning:
* pending=white, in_process=cyan, stop=red, await=yellow, done=green.
* @param status string
* @return string
**/
func StatusLabel(status string) string {
	switch status {
	case "in_process":
		return color.Cyan("in process")
	case "stop":
		return color.Red("stop")
	case "await":
		return color.Yellow("await")
	case "done":
		return color.Green("done")
	default: // pending
		return color.White("pending")
	}
}

/**
* PercentBar: Renders a 20-cell ASCII progress bar for percent (0-100), colored
* the same as the given status so percent and status read as one signal.
* @param status string, percent int
* @return string
**/
func PercentBar(status string, percent int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	const width = 20
	filled := percent * width / 100
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	paint := color.White
	switch status {
	case "in_process":
		paint = color.Cyan
	case "stop":
		paint = color.Red
	case "await":
		paint = color.Yellow
	case "done":
		paint = color.Green
	}

	return paint(bar) + " " + paint(padPercent(percent))
}

func padPercent(percent int) string {
	s := strconv.Itoa(percent) + "%"
	switch {
	case percent < 10:
		return "  " + s
	case percent < 100:
		return " " + s
	default:
		return s
	}
}
