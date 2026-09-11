package vanilla

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractDownloadedJar(t *testing.T) {
	dir := t.TempDir()
	jar := filepath.Join(dir, "client.jar")
	f, err := os.Create(jar)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	for _, n := range []string{"assets/m/lang/en.json", "data/m/recipes/r.json", "META-INF/X"} {
		fw, err := w.Create(n)
		if err != nil {
			t.Fatal(err)
		}
		fw.Write([]byte("{}"))
	}
	w.Close()
	f.Close()

	out := filepath.Join(dir, "out")
	res, err := ExtractDownloadedJar("1.21.8", out, false, jar)
	if err != nil {
		t.Fatal(err)
	}
	if res.Assets != 1 || res.Data != 1 {
		t.Errorf("counts = %d/%d", res.Assets, res.Data)
	}
	if res.OutDir != filepath.Join(out, "vanilla", "1.21.8") {
		t.Errorf("outdir = %q", res.OutDir)
	}

	// force cleans stale files
	stale := filepath.Join(res.OutDir, "stale.txt")
	if err := os.WriteFile(stale, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ExtractDownloadedJar("1.21.8", out, true, jar); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Error("expected stale file removed with force")
	}
}
