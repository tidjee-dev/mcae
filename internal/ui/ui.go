package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var (
	noColor bool

	TitleStyle   = lipgloss.NewStyle().Bold(true)
	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	ErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	MutedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	HeaderStyle  = lipgloss.NewStyle().Bold(true).Underline(true)
	PathStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
)

// SetNoColor disables ANSI styling (for --no-color, NO_COLOR, piped output, tests).
func SetNoColor(v bool) {
	noColor = v
	if v {
		lipgloss.SetColorProfile(termenv.Ascii)
	}
}

// IsNoColor reports whether styling is disabled.
func IsNoColor() bool { return noColor }

// Title renders a bold heading.
func Title(s string) string {
	if noColor {
		return s
	}
	return TitleStyle.Render(s)
}

// Success renders a green check-style message.
func Success(s string) string {
	if noColor {
		return s
	}
	return SuccessStyle.Render(s)
}

// Failure renders a red message.
func Failure(s string) string {
	if noColor {
		return s
	}
	return ErrorStyle.Render(s)
}

// Muted renders dim text.
func Muted(s string) string {
	if noColor {
		return s
	}
	return MutedStyle.Render(s)
}

// ExtractLine renders "Extracting <jar>...".
func ExtractLine(jar string) string {
	if noColor {
		return fmt.Sprintf("Extracting %s...", jar)
	}
	return fmt.Sprintf("%s %s...",
		MutedStyle.Render("Extracting"),
		PathStyle.Render(jar),
	)
}

// ExtractDone renders the per-mod result footer.
func ExtractDone(assets, data int, outDir string) string {
	if noColor {
		return fmt.Sprintf("  assets: %d files\n  data:   %d files\nDone: %s", assets, data, outDir)
	}
	return fmt.Sprintf("  %s %d files\n  %s %d files\n%s %s",
		MutedStyle.Render("assets:"), assets,
		MutedStyle.Render("data:  "), data,
		SuccessStyle.Render("Done:"), PathStyle.Render(outDir),
	)
}

// ModpackHeader renders "Found N mod JARs".
func ModpackHeader(n int) string {
	s := fmt.Sprintf("Found %d mod JARs", n)
	if noColor {
		return s
	}
	return TitleStyle.Render(s)
}

// ModpackStep renders "[i/n] name".
func ModpackStep(i, n int, name string) string {
	s := fmt.Sprintf("[%d/%d] %s", i, n, name)
	if noColor {
		return s
	}
	return MutedStyle.Render(fmt.Sprintf("[%d/%d]", i, n)) + " " + PathStyle.Render(name)
}

// StaticProgress renders a bubbles progress bar at percent (0..1) without
// requiring a running Tea program. Width is fixed for stable CLI output.
func StaticProgress(percent float64) string {
	if noColor {
		width := 20
		filled := int(percent * float64(width))
		if filled < 0 {
			filled = 0
		}
		if filled > width {
			filled = width
		}
		return "[" + strings.Repeat("=", filled) + strings.Repeat(" ", width-filled) + "]"
	}
	p := progress.New(progress.WithDefaultGradient(), progress.WithWidth(20))
	return p.ViewAs(percent)
}

// SpinnerFrame returns a spinner frame view for the given frame index.
// It exercises bubbles/spinner without needing a live Tea program.
func SpinnerFrame(frame int) string {
	sp := spinner.New(spinner.WithSpinner(spinner.Dot))
	frames := sp.Spinner.Frames
	if len(frames) == 0 {
		return ""
	}
	return frames[frame%len(frames)]
}

// InspectTable renders asset/data counts as a bubbles table (static view).
func InspectTable(assets, data map[string]int) string {
	cols := []table.Column{
		{Title: "Category", Width: 16},
		{Title: "Files", Width: 8},
	}
	var rows []table.Row
	order := []string{"lang", "models/block", "models/item", "textures/block", "textures/entity", "textures/item"}
	for _, k := range order {
		if v, ok := assets[k]; ok && v > 0 {
			rows = append(rows, table.Row{k, fmt.Sprintf("%d", v)})
		}
	}
	dorder := []string{"recipes", "loot_tables", "tags/blocks", "tags/items"}
	for _, k := range dorder {
		if v, ok := data[k]; ok && v > 0 {
			rows = append(rows, table.Row{k, fmt.Sprintf("%d", v)})
		}
	}
	if len(rows) == 0 {
		return Muted("  (no retained assets found)")
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithHeight(len(rows)+1),
	)
	style := table.DefaultStyles()
	style.Header = style.Header.Bold(true)
	if noColor {
		style.Header = lipgloss.NewStyle()
		style.Selected = lipgloss.NewStyle()
	}
	t.SetStyles(style)
	return t.View()
}

// ModpackTable renders a per-mod breakdown table.
func ModpackTable(names []string, assets, data []int) string {
	cols := []table.Column{
		{Title: "Mod", Width: 30},
		{Title: "Assets", Width: 8},
		{Title: "Data", Width: 8},
	}
	var rows []table.Row
	for i, n := range names {
		a, d := 0, 0
		if i < len(assets) {
			a = assets[i]
		}
		if i < len(data) {
			d = data[i]
		}
		if len(n) > 30 {
			n = n[:27] + "..."
		}
		rows = append(rows, table.Row{n, fmt.Sprintf("%d", a), fmt.Sprintf("%d", d)})
	}
	if len(rows) == 0 {
		return Muted("  (no mods)")
	}
	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithHeight(len(rows)+1),
	)
	style := table.DefaultStyles()
	style.Header = style.Header.Bold(true)
	if noColor {
		style.Header = lipgloss.NewStyle()
		style.Selected = lipgloss.NewStyle()
	}
	t.SetStyles(style)
	return t.View()
}
