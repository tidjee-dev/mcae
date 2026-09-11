package inspect

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func writeJar(t *testing.T, path string, names []string) {
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

func TestCategorize(t *testing.T) {
	cases := map[string]string{
		"assets/create/lang/en_us.json":       "lang",
		"assets/create/models/block/x.json":   "models/block",
		"assets/create/models/item/x.json":    "models/item",
		"assets/create/textures/block/x.png":  "textures/block",
		"assets/create/textures/entity/x.png": "textures/entity",
		"assets/create/textures/item/x.png":   "textures/item",
		"data/create/recipes/x.json":          "recipes",
		"data/create/loot_tables/x.json":      "loot_tables",
		"data/create/tags/blocks/x.json":      "tags/blocks",
		"data/create/tags/items/x.json":       "tags/items",
	}
	for in, want := range cases {
		_, cat, ok := Categorize(in)
		if !ok || cat != want {
			t.Errorf("%s: got %q,%v want %q", in, cat, ok, want)
		}
	}

	for _, rej := range []string{
		"META-INF/MANIFEST.MF",
		"assets/create/sounds/x.ogg",
		"data/create/functions/x.mcfunction",
	} {
		if _, _, ok := Categorize(rej); ok {
			t.Errorf("expected reject %s", rej)
		}
	}
}

func TestInspectJar(t *testing.T) {
	dir := t.TempDir()
	jar := filepath.Join(dir, "m.jar")
	writeJar(t, jar, []string{
		"assets/create/lang/en_us.json",
		"assets/create/models/item/a.json",
		"assets/create/textures/item/a.png",
		"data/create/recipes/a.json",
		"data/create/recipes/b.json",
		"META-INF/MANIFEST.MF",
	})

	rep, err := Inspect(jar)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Single == nil {
		t.Fatal("expected single report")
	}
	if len(rep.Single.Namespaces) != 1 || rep.Single.Namespaces[0] != "create" {
		t.Errorf("namespaces = %v", rep.Single.Namespaces)
	}
	if rep.Single.Assets["lang"] != 1 {
		t.Errorf("lang = %v", rep.Single.Assets)
	}
	if rep.Single.Data["recipes"] != 2 {
		t.Errorf("recipes = %v", rep.Single.Data)
	}
}

func TestInspectDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "assets", "jei", "lang"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets", "jei", "lang", "en.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "data", "jei", "recipes"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "data", "jei", "recipes", "x.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	rep, err := Inspect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Single == nil || rep.Single.Assets["lang"] != 1 || rep.Single.Data["recipes"] != 1 {
		t.Errorf("unexpected summary: %+v", rep.Single)
	}
}

func TestInspectModpack(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "mods"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeJar(t, filepath.Join(root, "mods", "a.jar"), []string{"assets/a/lang/en.json"})
	writeJar(t, filepath.Join(root, "mods", "b.jar"), []string{"data/b/recipes/x.json"})

	rep, err := Inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Modpack == nil {
		t.Fatal("expected modpack report")
	}
	if len(rep.Modpack.Mods) != 2 {
		t.Fatalf("mods = %d, want 2", len(rep.Modpack.Mods))
	}
}

func TestInspectErrors(t *testing.T) {
	if _, err := Inspect(filepath.Join(t.TempDir(), "nope.jar")); err == nil {
		t.Error("expected error for missing path")
	}
	txt := filepath.Join(t.TempDir(), "x.txt")
	if err := os.WriteFile(txt, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Inspect(txt); err == nil {
		t.Error("expected error for non-jar file")
	}
}
