package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Sims 1 bar colors: green filled over red empty
var (
	barFilledColor = lipgloss.Color("#30B020") // bright Sims green
	barEmptyColor  = lipgloss.Color("#8B1A1A") // Sims dark red
)

func BarColor(value int) lipgloss.Color {
	return barFilledColor
}

func EmptyBarColor(value int) lipgloss.Color {
	switch {
	case value >= 7:
		return lipgloss.Color("#1A6B10") // dark green
	case value >= 5:
		return lipgloss.Color("#B08B00") // amber/orange
	default:
		return barEmptyColor // dark red
	}
}

func RenderBar(value, width int) string {
	if width <= 0 {
		return ""
	}
	filled := 0
	if value > 0 {
		filled = (value * width) / 10
		if filled > width {
			filled = width
		}
	}
	empty := width - filled

	emptyColor := EmptyBarColor(value)
	filledStyle := lipgloss.NewStyle().Foreground(barFilledColor)
	emptyStyle := lipgloss.NewStyle().Foreground(emptyColor)

	return filledStyle.Render(strings.Repeat("█", filled)) + emptyStyle.Render(strings.Repeat("░", empty))
}

var panelBg = lipgloss.Color("#9099C0")
var panelFg = lipgloss.Color("#1A2260")

func textStyle() lipgloss.Style {
	return lipgloss.NewStyle().Background(panelBg).Foreground(panelFg).Bold(true)
}

func RenderLabeledBar(label string, value, barWidth, labelWidth int) string {
	paddedLabel := fmt.Sprintf("%-*s", labelWidth, label)
	score := fmt.Sprintf("%d/10", value)
	return textStyle().Render(paddedLabel+" ") + RenderBar(value, barWidth) + textStyle().Render(" "+score)
}

func RenderGrid(values [NeedCount]int, barWidth int) string {
	names := NeedNames()
	labelWidth := 0
	for _, n := range names {
		if len(n) > labelWidth {
			labelWidth = len(n)
		}
	}

	ts := textStyle()
	half := int(NeedCount) / 2
	lines := make([]string, half)
	for i := 0; i < half; i++ {
		left := RenderLabeledBar(names[i], values[i], barWidth, labelWidth)
		right := RenderLabeledBar(names[i+half], values[i+half], barWidth, labelWidth)
		lines[i] = left + ts.Render("    ") + right
	}
	return strings.Join(lines, "\n\n")
}

func SimsBlueBackground(w, h int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(w).
		Height(h).
		Bold(true).
		Background(lipgloss.Color("#9099C0")).
		Foreground(lipgloss.Color("#1A2260")).
		Padding(1, 2)
}
