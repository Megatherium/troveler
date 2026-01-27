package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type TableConfig struct {
	Headers    []string
	Rows       [][]string
	HeaderFunc func(string, int) string
	RowFunc    func(string, int, int) string
	ShowHeader bool
}

func DefaultHeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FF00"))
}

func DefaultRowStyle(rowIdx, colIdx int) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(GetGradientColorSimple(rowIdx)))
}

func RenderTable(config TableConfig) string {
	if len(config.Headers) == 0 && len(config.Rows) == 0 {
		return ""
	}

	if len(config.Rows) == 0 {
		return ""
	}

	colWidths := calculateColumnWidths(config.Headers, config.Rows)

	var b strings.Builder

	b.WriteString(renderTopBorder(colWidths))

	if config.ShowHeader && len(config.Headers) > 0 {
		renderHeader(&b, config.Headers, colWidths, config.HeaderFunc)
		b.WriteString(renderMidBorder(colWidths))
	}

	renderRows(&b, config.Rows, colWidths, config.RowFunc)

	b.WriteString(renderBotBorder(colWidths))

	return b.String()
}

func calculateColumnWidths(headers []string, rows [][]string) []int {
	numCols := len(headers)
	if len(rows) > 0 && numCols == 0 {
		numCols = len(rows[0])
	}

	colWidths := make([]int, numCols)
	for i, h := range headers {
		colWidths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}
	return colWidths
}

func renderTopBorder(colWidths []int) string {
	topBorder := "┌"
	joinChar := "┬"
	rightEnd := "┐"

	var b strings.Builder
	b.WriteString(topBorder)
	for i, w := range colWidths {
		b.WriteString(strings.Repeat("─", w+2))
		if i < len(colWidths)-1 {
			b.WriteString(joinChar)
		}
	}
	b.WriteString(rightEnd + "\n")
	return b.String()
}

func renderMidBorder(colWidths []int) string {
	midBorder := "├"
	joinMid := "┼"
	rightMid := "┤"

	var b strings.Builder
	b.WriteString(midBorder)
	for i, w := range colWidths {
		b.WriteString(strings.Repeat("─", w+2))
		if i < len(colWidths)-1 {
			b.WriteString(joinMid)
		}
	}
	b.WriteString(rightMid + "\n")
	return b.String()
}

func renderBotBorder(colWidths []int) string {
	botBorder := "└"
	joinBot := "┴"
	rightBot := "┘"

	var b strings.Builder
	b.WriteString(botBorder)
	for i, w := range colWidths {
		b.WriteString(strings.Repeat("─", w+2))
		if i < len(colWidths)-1 {
			b.WriteString(joinBot)
		}
	}
	b.WriteString(rightBot)
	return b.String()
}

func renderHeader(b *strings.Builder, headers []string, colWidths []int, styleFunc func(string, int) string) {
	borderChar := "│"

	if styleFunc == nil {
		defaultStyle := DefaultHeaderStyle()
		styleFunc = func(s string, _ int) string {
			return defaultStyle.Render(s)
		}
	}

	b.WriteString(borderChar)
	for i, h := range headers {
		pad := colWidths[i] - len(h)
		b.WriteString(" ")
		b.WriteString(styleFunc(h, i))
		b.WriteString(strings.Repeat(" ", pad+1))
		b.WriteString(borderChar)
	}
	b.WriteString("\n")
}

func renderRows(b *strings.Builder, rows [][]string, colWidths []int, styleFunc func(string, int, int) string) {
	borderChar := "│"

	if styleFunc == nil {
		styleFunc = func(s string, rowIdx, colIdx int) string {
			return DefaultRowStyle(rowIdx, colIdx).Render(s)
		}
	}

	for rowIdx, row := range rows {
		b.WriteString(borderChar)
		for i, cell := range row {
			if i >= len(colWidths) {
				break
			}
			pad := colWidths[i] - len(cell)
			b.WriteString(" ")
			b.WriteString(styleFunc(cell, rowIdx, i))
			b.WriteString(strings.Repeat(" ", pad+1))
			b.WriteString(borderChar)
		}
		b.WriteString("\n")
	}
}

func RenderKeyValueTable(rows [][]string) string {
	if len(rows) == 0 {
		return ""
	}

	colWidths := []int{0, 0}
	for _, row := range rows {
		if len(row) >= 2 {
			if len(row[0]) > colWidths[0] {
				colWidths[0] = len(row[0])
			}
			if len(row[1]) > colWidths[1] {
				colWidths[1] = len(row[1])
			}
		}
	}

	var b strings.Builder
	borderChar := "│"
	topBorder := "┌"
	botBorder := "└"
	joinChar := "┬"
	joinBot := "┴"
	rightEnd := "┐"
	rightBot := "┘"

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FF00"))

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC"))

	b.WriteString(topBorder)
	for i, w := range colWidths {
		b.WriteString(strings.Repeat("─", w+2))
		if i < len(colWidths)-1 {
			b.WriteString(joinChar)
		}
	}
	b.WriteString(rightEnd + "\n")

	for _, row := range rows {
		if len(row) < 2 {
			continue
		}
		b.WriteString(borderChar)
		b.WriteString(" ")
		b.WriteString(labelStyle.Render(row[0]))
		pad := colWidths[0] - len(row[0])
		b.WriteString(strings.Repeat(" ", pad+1))
		b.WriteString(borderChar)
		b.WriteString(" ")
		b.WriteString(valueStyle.Render(row[1]))
		pad = colWidths[1] - len(row[1])
		b.WriteString(strings.Repeat(" ", pad+1))
		b.WriteString(borderChar + "\n")
	}

	b.WriteString(botBorder)
	for i, w := range colWidths {
		b.WriteString(strings.Repeat("─", w+2))
		if i < len(colWidths)-1 {
			b.WriteString(joinBot)
		}
	}
	b.WriteString(rightBot)

	return b.String()
}
