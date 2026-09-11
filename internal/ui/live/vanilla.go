package live

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	vanilla "github.com/tidjee-dev/mcae/internal/minecraft/vanilla"
	"github.com/tidjee-dev/mcae/internal/ui"
)

// Vanilla messages driven by the worker goroutine.
type vanillaStageMsg struct {
	stage string
	info  string
}

type vanillaProgressMsg struct {
	written int64
	total   int64
}

type vanillaDoneMsg struct {
	res vanilla.ExtractResult
	err error
}

// VanillaModel renders resolve → download → extract stages.
type VanillaModel struct {
	version  string
	stage    string
	info     string
	written  int64
	total    int64
	prog     progress.Model
	spin     spinner.Model
	start    time.Time
	finished bool
	failed   error
	result   *vanilla.ExtractResult
}

// NewVanillaModel creates a model for a Minecraft version.
func NewVanillaModel(version string) VanillaModel {
	p := progress.New(progress.WithDefaultGradient(), progress.WithWidth(30))
	s := spinner.New(spinner.WithSpinner(spinner.Dot))
	return VanillaModel{
		version: version,
		stage:   "resolving",
		prog:    p,
		spin:    s,
		start:   time.Now(),
	}
}

func (m VanillaModel) Init() tea.Cmd {
	return m.spin.Tick
}

func (m VanillaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			m.failed = fmt.Errorf("interrupted")
			m.finished = true
			return m, tea.Quit
		}
		return m, nil

	case vanillaStageMsg:
		m.stage = msg.stage
		m.info = msg.info
		return m, nil

	case vanillaProgressMsg:
		m.stage = "downloading"
		m.written = msg.written
		m.total = msg.total
		return m, nil

	case vanillaDoneMsg:
		m.finished = true
		if msg.err != nil {
			m.failed = msg.err
		} else {
			m.result = &msg.res
		}
		return m, tea.Quit

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		pm, cmd := m.prog.Update(msg)
		m.prog = pm.(progress.Model)
		return m, cmd
	}

	return m, nil
}

func (m VanillaModel) percent() float64 {
	if m.total <= 0 {
		return 0
	}
	p := float64(m.written) / float64(m.total)
	if p < 0 {
		return 0
	}
	if p > 1 {
		return 1
	}
	return p
}

func (m VanillaModel) View() string {
	var b strings.Builder
	if m.finished && m.failed == nil && m.result != nil {
		b.WriteString(ui.Success(fmt.Sprintf("Done: %s", m.result.OutDir)))
		b.WriteString(fmt.Sprintf("  (%d assets, %d data, in %s)\n",
			m.result.Assets, m.result.Data, time.Since(m.start).Round(time.Millisecond)))
		return b.String()
	}
	if m.finished && m.failed != nil {
		b.WriteString(ui.Failure("failed: " + m.failed.Error()))
		b.WriteString("\n")
		return b.String()
	}

	switch m.stage {
	case "resolving":
		b.WriteString(fmt.Sprintf("%s Resolving Minecraft %s…\n", m.spin.View(), m.version))
	case "downloading":
		if m.total > 0 {
			b.WriteString(fmt.Sprintf("%s Downloading client %s (%s / %s)\n",
				m.spin.View(), m.version, ui.FormatBytes(m.written), ui.FormatBytes(m.total)))
			b.WriteString(m.prog.ViewAs(m.percent()))
			b.WriteString("\n")
		} else {
			b.WriteString(fmt.Sprintf("%s Downloading client %s (%s)…\n",
				m.spin.View(), m.version, ui.FormatBytes(m.written)))
		}
	case "extracting":
		b.WriteString(fmt.Sprintf("%s Extracting vanilla %s…\n", m.spin.View(), m.version))
		if m.info != "" {
			b.WriteString(ui.Muted("  " + m.info))
			b.WriteString("\n")
		}
	default:
		b.WriteString(fmt.Sprintf("%s Working…\n", m.spin.View()))
	}
	return b.String()
}

// RunVanilla resolves, downloads and extracts with live progress to out.
func RunVanilla(versionID, outputBase string, force bool, manifestURL string, out io.Writer) (vanilla.ExtractResult, error) {
	if manifestURL == "" {
		manifestURL = vanilla.DefaultManifestURL
	}

	model := NewVanillaModel(versionID)
	p := tea.NewProgram(model)

	resultCh := make(chan vanilla.ExtractResult, 1)
	errCh := make(chan error, 1)

	go func() {
		p.Send(vanillaStageMsg{stage: "resolving"})
		clientURL, err := vanilla.ResolveClientURL(nil, manifestURL, versionID)
		if err != nil {
			p.Send(vanillaDoneMsg{err: err})
			errCh <- err
			return
		}

		jarPath, cleanup, err := vanilla.DownloadClient(nil, clientURL, func(written, total int64) {
			p.Send(vanillaProgressMsg{written: written, total: total})
		})
		if err != nil {
			p.Send(vanillaDoneMsg{err: err})
			errCh <- err
			return
		}
		defer cleanup()

		p.Send(vanillaStageMsg{stage: "extracting", info: "filtering assets…"})
		res, err := vanilla.ExtractDownloadedJar(versionID, outputBase, force, jarPath)
		if err != nil {
			p.Send(vanillaDoneMsg{err: err})
			errCh <- err
			return
		}
		resultCh <- res
		p.Send(vanillaDoneMsg{res: res})
	}()

	final, err := p.Run()
	if err != nil {
		return vanilla.ExtractResult{}, err
	}
	if fm, ok := final.(VanillaModel); ok {
		if fm.failed != nil {
			return vanilla.ExtractResult{}, fm.failed
		}
		if fm.result != nil {
			return *fm.result, nil
		}
	}
	select {
	case res := <-resultCh:
		return res, nil
	case err := <-errCh:
		return vanilla.ExtractResult{}, err
	default:
		return vanilla.ExtractResult{}, fmt.Errorf("vanilla extraction did not complete")
	}
}
