package ui

import (
	"strings"
	"testing"
)

func TestErrorLine(t *testing.T) {
	SetNoColor(true)
	defer SetNoColor(false)

	if got := ErrorLine("boom"); got != "Error: boom" {
		t.Errorf("got %q", got)
	}
	if got := ErrorLine("Error: boom"); got != "Error: boom" {
		t.Errorf("got %q", got)
	}
	if got := ErrorLine(""); got != "" {
		t.Errorf("got %q", got)
	}
}

func TestRenderSummaryTotals(t *testing.T) {
	SetNoColor(true)
	defer SetNoColor(false)

	got := RenderSummary("/tmp/x", []string{"create", "jei"},
		map[string]int{"lang": 2, "models/item": 3},
		map[string]int{"recipes": 4})
	for _, want := range []string{"Minecraft Asset Summary", "Total: 5 assets, 4 data files", "create", "jei", "recipes"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestRenderModpackTableTotals(t *testing.T) {
	SetNoColor(true)
	defer SetNoColor(false)

	got := RenderModpackTable([]ModRow{
		{Name: "a.jar", Namespaces: 1, Assets: 3, Data: 1},
		{Name: "b.jar", Namespaces: 2, Assets: 5, Data: 7},
	})
	for _, want := range []string{"a.jar", "b.jar", "Total", "8", "8"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	if FormatBytes(500) != "500 B" {
		t.Errorf("got %q", FormatBytes(500))
	}
	if !strings.Contains(FormatBytes(2048), "KB") {
		t.Errorf("got %q", FormatBytes(2048))
	}
	if !strings.Contains(FormatBytes(5*1024*1024), "MB") {
		t.Errorf("got %q", FormatBytes(5*1024*1024))
	}
}
