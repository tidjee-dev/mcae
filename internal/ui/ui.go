// Package ui renders styled terminal output with graceful plain-text fallback.
package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	ltable "github.com/charmbracelet/lipgloss/table"
	"github.com/mattn/go-isatty"
	"github.com/muesli/termenv"
)

var (
	noColor bool

	// Shared Lipgloss styles; all renderers honor SetNoColor.
	TitleStyle   = lipgloss.NewStyle().Bold(true)
	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	ErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	MutedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	HeaderStyle  = lipgloss.NewStyle().Bold(true).Underline(true)
	PathStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
)

// AssetOrder and DataOrder define stable row ordering for summaries.
var (
	AssetOrder = []string{"lang", "models/block", "models/item", "textures/block", "textures/entity", "textures/item"}
	DataOrder  = []string{"recipes", "loot_tables", "tags/blocks", "tags/items"}
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

// IsTTY reports whether stdout is an interactive terminal.
func IsTTY() bool {
	return isatty.IsTerminal(os.Stdout.Fd())
}

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

// ErrorLine renders a single "Error: msg" line for stderr.
func ErrorLine(msg string) string {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return ""
	}
	if !strings.HasPrefix(strings.ToLower(msg), "error") {
		msg = "Error: " + msg
	}
	if noColor {
		return msg
	}
	return ErrorStyle.Render(msg)
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

func tableStyle() (*lipgloss.Style, func(row, col int) lipgloss.Style) {
	header := lipgloss.NewStyle().Bold(true)
	if noColor {
		header = lipgloss.NewStyle()
	}
	styleFn := func(row, col int) lipgloss.Style {
		if row == ltable.HeaderRow && !noColor {
			return header
		}
		return lipgloss.NewStyle()
	}
	return &header, styleFn
}

// InspectTable renders asset/data counts as a single table.
// Either map may be nil; a combined section column is used.
func InspectTable(assets, data map[string]int) string {
	var rows [][]string
	for _, k := range AssetOrder {
		if v, ok := assets[k]; ok && v > 0 {
			rows = append(rows, []string{"assets", k, fmt.Sprintf("%d", v)})
		}
	}
	for _, k := range DataOrder {
		if v, ok := data[k]; ok && v > 0 {
			rows = append(rows, []string{"data", k, fmt.Sprintf("%d", v)})
		}
	}
	if len(rows) == 0 {
		return Muted("  (no retained assets found)")
	}

	_, styleFn := tableStyle()
	t := ltable.New().
		Border(lipgloss.NormalBorder()).
		Headers("Section", "Category", "Files").
		Rows(rows...).
		StyleFunc(styleFn)
	return t.Render()
}

// ModRow is one row of the modpack breakdown.
type ModRow struct {
	Name       string
	Namespaces int
	Assets     int
	Data       int
}

// ModpackTable renders a per-mod breakdown table (kept for compatibility).
func ModpackTable(names []string, assets, data []int) string {
	rows := make([]ModRow, len(names))
	for i, n := range names {
		a, d := 0, 0
		if i < len(assets) {
			a = assets[i]
		}
		if i < len(data) {
			d = data[i]
		}
		rows[i] = ModRow{Name: n, Assets: a, Data: d}
	}
	return RenderModpackTable(rows)
}

// RenderModpackTable renders mod rows with a totals footer row.
func RenderModpackTable(rows []ModRow) string {
	if len(rows) == 0 {
		return Muted("  (no mods)")
	}

	data := make([][]string, 0, len(rows)+1)
	totalA, totalD := 0, 0
	for _, r := range rows {
		name := r.Name
		data = append(data, []string{name, fmt.Sprintf("%d", r.Namespaces), fmt.Sprintf("%d", r.Assets), fmt.Sprintf("%d", r.Data)})
		totalA += r.Assets
		totalD += r.Data
	}
	data = append(data, []string{"Total", "", fmt.Sprintf("%d", totalA), fmt.Sprintf("%d", totalD)})

	_, styleFn := tableStyle()
	t := ltable.New().
		Border(lipgloss.NormalBorder()).
		Headers("Mod", "Namespaces", "Assets", "Data").
		Rows(data...).
		StyleFunc(styleFn)
	return t.Render()
}

// RenderSummary renders a full single-mod summary: title, path,
// one combined table with totals, and the namespace list.
func RenderSummary(path string, namespaces []string, assets, data map[string]int) string {
	var b strings.Builder
	b.WriteString(Title("Minecraft Asset Summary"))
	b.WriteString("\n\nPath: ")
	b.WriteString(path)
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("Namespaces: %d\n", len(namespaces)))
	b.WriteString(InspectTable(assets, data))
	totalA, totalD := 0, 0
	for _, v := range assets {
		totalA += v
	}
	for _, v := range data {
		totalD += v
	}
	b.WriteString(fmt.Sprintf("\nTotal: %d assets, %d data files\n", totalA, totalD))
	b.WriteString("\nNamespaces\n")
	if len(namespaces) == 0 {
		b.WriteString("  (none)")
		return b.String()
	}
	for _, n := range namespaces {
		b.WriteString("  - ")
		b.WriteString(n)
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// FormatBytes renders a byte count for download progress.
func FormatBytes(n int64) string {
	if n < 1024 {
		return fmt.Sprintf("%d B", n)
	}
	if n < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	}
	if n < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
	}
	return fmt.Sprintf("%.2f GB", float64(n)/(1024*1024*1024))
}
