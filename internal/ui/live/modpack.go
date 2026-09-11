package live

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tidjee-dev/mcae/internal/extractor"
	"github.com/tidjee-dev/mcae/internal/ui"
)

// modDoneMsg is sent by the worker after each JAR.
type modDoneMsg struct {
	index int
	res   extractor.Result
	err   error
}

// ModpackModel renders live modpack extraction progress.
type ModpackModel struct {
	total    int
	done     int
	current  string
	results  []extractor.Result
	prog     progress.Model
	spin     spinner.Model
	start    time.Time
	finished bool
	failed   error
	width    int
}

// NewModpackModel creates a model for total JARs.
func NewModpackModel(total int) ModpackModel {
	p := progress.New(progress.WithDefaultGradient(), progress.WithWidth(30))
	s := spinner.New(spinner.WithSpinner(spinner.Dot))
	return ModpackModel{
		total: total,
		prog:  p,
		spin:  s,
		start: time.Now(),
		width: 80,
	}
}

func (m ModpackModel) Init() tea.Cmd {
	return m.spin.Tick
}

func (m ModpackModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.prog.Width = min(msg.Width-20, 40)
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			m.failed = fmt.Errorf("interrupted")
			m.finished = true
			return m, tea.Quit
		}
		return m, nil

	case modDoneMsg:
		if msg.err != nil {
			m.failed = msg.err
			m.finished = true
			return m, tea.Quit
		}
		m.results = append(m.results, msg.res)
		m.done++
		m.current = ""
		if m.done >= m.total {
			m.finished = true
			return m, tea.Quit
		}
		return m, nil

	case modStartMsg:
		m.current = msg.name
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		pm, cmd := m.prog.Update(msg)
		m.prog = pm.(progress.Model)
		return m, cmd

	case quitMsg:
		m.finished = true
		if msg.err != nil {
			m.failed = msg.err
		}
		return m, tea.Quit
	}

	return m, nil
}

// modStartMsg announces the JAR currently extracting.
type modStartMsg struct {
	index int
	name  string
}

// quitMsg ends the program.
type quitMsg struct {
	err error
}

func (m ModpackModel) percent() float64 {
	if m.total == 0 {
		return 1
	}
	return float64(m.done) / float64(m.total)
}

func (m ModpackModel) View() string {
	var b strings.Builder
	b.WriteString(ui.ModpackHeader(m.total))
	b.WriteString("\n")
	if m.finished && m.failed == nil {
		b.WriteString(ui.StaticProgress(1))
		b.WriteString(fmt.Sprintf("  done in %s\n", time.Since(m.start).Round(time.Millisecond)))
		for _, r := range m.results {
			b.WriteString(fmt.Sprintf("  %s: %d assets, %d data\n", r.JarName, r.Assets, r.Data))
		}
		return b.String()
	}
	if m.failed != nil {
		b.WriteString(ui.Failure("failed: " + m.failed.Error()))
		b.WriteString("\n")
		return b.String()
	}
	if m.current != "" {
		b.WriteString(fmt.Sprintf("%s Extracting %s [%d/%d]\n", m.spin.View(), m.current, m.done+1, m.total))
	} else {
		b.WriteString(fmt.Sprintf("%s Working… [%d/%d]\n", m.spin.View(), m.done, m.total))
	}
	b.WriteString(m.prog.ViewAs(m.percent()))
	b.WriteString("\n")
	return b.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// RunModpack extracts jars live, rendering to out.
// jars are base names inside modsDir; dest for each is outputDir/<modname>.
func RunModpack(jars []string, modsDir, outputDir string, force bool, out io.Writer) ([]extractor.Result, error) {
	model := NewModpackModel(len(jars))
	p := tea.NewProgram(model, tea.WithOutput(out))

	results := make([]extractor.Result, 0, len(jars))
	var runErr error
	done := make(chan struct{})

	go func() {
		defer close(done)
		for i, name := range jars {
			if runErr != nil {
				return
			}
			p.Send(modStartMsg{index: i, name: name})
			dest := filepath.Join(outputDir, extractor.JarModName(name))
			var assets, data int
			var err error
			if force {
				// Best effort: remove then recreate via ExtractJar after MkdirAll.
				if rerr := os.RemoveAll(filepath.Clean(dest)); rerr != nil {
					p.Send(modDoneMsg{index: i, err: fmt.Errorf("clean output directory: %w", rerr)})
					return
				}
			}
			if err = os.MkdirAll(dest, 0o755); err != nil {
				p.Send(modDoneMsg{index: i, err: fmt.Errorf("create output directory: %w", err)})
				return
			}
			assets, data, err = extractor.ExtractJar(filepath.Join(modsDir, name), dest)
			if err != nil {
				p.Send(modDoneMsg{index: i, err: fmt.Errorf("extract %s: %w", name, err)})
				return
			}
			res := extractor.Result{JarName: name, OutDir: dest, Assets: assets, Data: data}
			results = append(results, res)
			p.Send(modDoneMsg{index: i, res: res})
		}
	}()

	final, err := p.Run()
	if err != nil {
		<-done
		return results, err
	}
	<-done
	if fm, ok := final.(ModpackModel); ok && fm.failed != nil {
		if len(results) > 0 {
			return results, fm.failed
		}
		return nil, fm.failed
	}
	if runErr != nil {
		return results, runErr
	}
	return results, nil
}
