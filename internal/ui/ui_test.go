package ui

import (
	"strings"
	"testing"
)

func TestStaticProgressNoColor(t *testing.T) {
	SetNoColor(true)
	defer SetNoColor(false)

	if got := StaticProgress(0); !strings.HasPrefix(got, "[") {
		t.Errorf("unexpected %q", got)
	}
	if got := StaticProgress(1); !strings.Contains(got, "=") {
		t.Errorf("unexpected %q", got)
	}
}

func TestInspectTableEmpty(t *testing.T) {
	SetNoColor(true)
	defer SetNoColor(false)

	if got := InspectTable(nil, nil); !strings.Contains(got, "no retained") {
		t.Errorf("unexpected %q", got)
	}
}

func TestModpackTableRenders(t *testing.T) {
	SetNoColor(true)
	defer SetNoColor(false)

	got := ModpackTable([]string{"a-1.0.jar"}, []int{3}, []int{2})
	if !strings.Contains(got, "a-1.0.jar") {
		t.Errorf("missing mod name in %q", got)
	}
}

func TestSpinnerFrameCycles(t *testing.T) {
	a := SpinnerFrame(0)
	b := SpinnerFrame(1000)
	if a == "" || b == "" {
		t.Error("expected non-empty frames")
	}
}
