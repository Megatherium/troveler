package commands

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"troveler/db"
	"troveler/internal/install"
	"troveler/pkg/ui"
)

func showAllInstalls(name string, installs []db.InstallInstruction) error {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FFFF")).
		Render(name + " - All Install Commands:"))
	fmt.Println(strings.Repeat("─", len(name)+len(" - All Install Commands:")))
	fmt.Println()

	virtuals := install.GenerateVirtualInstallInstructions(installs)

	headers := []string{"Platform", "Command"}

	totalRows := len(installs) + len(virtuals)
	rows := make([][]string, 0, totalRows)

	for _, inst := range installs {
		rows = append(rows, []string{inst.Platform, inst.Command})
	}

	for _, v := range virtuals {
		rows = append(rows, []string{v.Platform, v.Command})
	}

	tableConfig := ui.TableConfig{
		Headers: headers,
		Rows:    rows,
		RowFunc: func(cell string, rowIdx, colIdx int) string {
			var color string
			if colIdx == 0 {
				color = ui.GetGradientColorSimple(rowIdx)
			} else {
				color = ui.GetGradientColorSimple((rowIdx + len(rows)/2) % len(ui.GradientColors))
			}

			return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(cell)
		},
		ShowHeader: true,
	}

	fmt.Println(ui.RenderTable(tableConfig))

	return nil
}

func displayInstallCommands(platformID string, matched []db.InstallInstruction, miseEnabled bool) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FF00")).
		Render("Install command for " + platformID + ":"))
	fmt.Println()

	for _, inst := range matched {
		cmd := inst.Command
		if miseEnabled {
			cmd = install.TransformToMise(cmd)
		}
		fmt.Println(lipgloss.NewStyle().Bold(true).Render(cmd))
	}
	fmt.Println()
}

func displayLanguageFallback(toolLanguage string, langMatched []db.InstallInstruction) {
	fmt.Printf("Trying language (%s):\n", toolLanguage)
	for _, inst := range langMatched {
		fmt.Println(lipgloss.NewStyle().Bold(true).Render(inst.Command))
	}
	fmt.Println()
}

func displayNoInstallMethod(toolName string, platformID string, installs []db.InstallInstruction, miseEnabled bool) {
	virtuals := install.GenerateVirtualInstallInstructions(installs)

	if miseEnabled && len(virtuals) > 0 {
		fmt.Println(lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFF00")).
			Render(fmt.Sprintf("No install method matched for %s.", platformID)))
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF00")).
			Render("Suggested (mise):"))
		fmt.Println()

		for _, v := range virtuals {
			fmt.Println(lipgloss.NewStyle().Bold(true).Render(v.Command))
		}
		fmt.Println()
		fmt.Println("Or select a method manually with -o <platform>")
	} else {
		fmt.Println(lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF0000")).
			Render(fmt.Sprintf("No install method could be determined for %s.", platformID)))
		fmt.Println()
		fmt.Println("Select a method manually with -o <platform>, or use --all to see all options.")
	}
	fmt.Println()
}
