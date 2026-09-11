package inspect

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestInspectEmptyModpackErrors(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "mods"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Inspect(root); !errors.Is(err, ErrNoMods) {
		t.Errorf("expected ErrNoMods, got %v", err)
	}
}

func TestSummaryTotals(t *testing.T) {
	s := newSummary("x")
	s.add("assets/a/lang/en.json")
	s.add("assets/a/lang/de.json")
	s.add("data/a/recipes/r.json")
	if s.TotalAssets() != 2 || s.TotalData() != 1 {
		t.Errorf("totals = %d/%d", s.TotalAssets(), s.TotalData())
	}
	mp := &ModpackSummary{Mods: []Summary{*s, *s}}
	a, d := mp.Totals()
	if a != 4 || d != 2 {
		t.Errorf("modpack totals = %d/%d", a, d)
	}
}
