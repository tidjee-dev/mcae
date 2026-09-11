package extractor

import (
	"archive/zip"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func writeTestJar(t *testing.T, path string, names []string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	for _, n := range names {
		fw, err := w.Create(n)
		if err != nil {
			t.Fatal(err)
		}
		fw.Write([]byte("{}"))
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestListModJars(t *testing.T) {
	root := t.TempDir()
	if _, _, _, err := ListModJars(root, ""); err == nil {
		t.Error("expected error for missing mods/")
	}
	if err := os.MkdirAll(filepath.Join(root, "mods"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := ListModJars(root, ""); !errors.Is(err, ErrNoJars) {
		t.Errorf("expected ErrNoJars, got %v", err)
	}
	writeTestJar(t, filepath.Join(root, "mods", "a.jar"), []string{"assets/a/lang/en.json"})
	if _, _, jars, err := ListModJars(root, ""); err != nil || len(jars) != 1 {
		t.Errorf("jars=%v err=%v", jars, err)
	}
}

func TestExtractModForce(t *testing.T) {
	dir := t.TempDir()
	jar := filepath.Join(dir, "m.jar")
	writeTestJar(t, jar, []string{"assets/m/lang/en.json"})

	out := filepath.Join(dir, "out")
	res, err := ExtractModWithForce(jar, out, false)
	if err != nil {
		t.Fatal(err)
	}
	stale := filepath.Join(res.OutDir, "stale.txt")
	if err := os.WriteFile(stale, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Without force the stale file survives.
	if _, err := ExtractModWithForce(jar, out, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); err != nil {
		t.Errorf("expected stale file to survive without --force: %v", err)
	}

	// With force it is cleaned.
	if _, err := ExtractModWithForce(jar, out, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Error("expected stale file to be removed with --force")
	}
}

func TestExtractModpackStreamCallback(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "mods"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestJar(t, filepath.Join(root, "mods", "a.jar"), []string{"assets/a/lang/en.json"})
	writeTestJar(t, filepath.Join(root, "mods", "b.jar"), []string{"data/b/recipes/x.json"})

	var calls []string
	results, err := ExtractModpackWithForce(root, "", false, func(i, total int, res Result) {
		calls = append(calls, res.JarName)
		if total != 2 {
			t.Errorf("total = %d, want 2", total)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 2 || len(results) != 2 {
		t.Errorf("calls=%v results=%d", calls, len(results))
	}
}

func TestSentinelErrors(t *testing.T) {
	if _, err := ExtractMod(filepath.Join(t.TempDir(), "nope.jar"), ""); !errors.Is(err, ErrModNotFound) {
		t.Errorf("expected ErrModNotFound, got %v", err)
	}
	root := t.TempDir()
	if _, err := ExtractModpack(root, ""); !errors.Is(err, ErrInvalidModpack) {
		t.Errorf("expected ErrInvalidModpack, got %v", err)
	}
	empty := t.TempDir()
	if err := os.MkdirAll(filepath.Join(empty, "mods"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := ExtractModpack(empty, ""); !errors.Is(err, ErrNoJars) {
		t.Errorf("expected ErrNoJars, got %v", err)
	}
}
