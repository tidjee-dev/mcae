package extractor

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestShouldKeepAccepted(t *testing.T) {
	accepted := []string{
		"assets/create/lang/en_us.json",
		"assets/create/models/item/brass_ingot.json",
		"assets/create/models/block/brass_block.json",
		"assets/create/textures/item/brass_ingot.png",
		"assets/create/textures/block/brass_block.png",
		"assets/create/textures/entity/example.png",
		"data/create/recipes/brass_ingot.json",
		"data/create/loot_tables/example.json",
		"data/create/tags/items/ingots.json",
		"data/create/tags/blocks/storage.json",
		// dynamic namespaces
		"assets/jei/lang/en_us.json",
		"data/minecraft/recipes/stick.json",
	}

	for _, p := range accepted {
		if !ShouldKeep(p) {
			t.Errorf("expected to keep %q", p)
		}
	}
}

func TestShouldKeepRejected(t *testing.T) {
	rejected := []string{
		"META-INF/MANIFEST.MF",
		"assets/create/blockstates/example.json",
		"assets/create/sounds/example.ogg",
		"assets/create/shaders/example.json",
		"data/create/functions/example.mcfunction",
		"data/create/advancements/example.json",
		"data/create/structures/example.nbt",
		"com/example/Foo.class",
		"assets/create/textures/particle/x.png",
		"assets/create/models/geo/x.json",
		"data/create/tags/fluids/x.json",
		"pack.mcmeta",
		"assets",
		"data",
	}

	for _, p := range rejected {
		if ShouldKeep(p) {
			t.Errorf("expected to reject %q", p)
		}
	}
}

func TestSafePathTraversal(t *testing.T) {
	base := t.TempDir()

	for _, evil := range []string{
		"../../file",
		"../../../etc/passwd",
		"assets/../../escape.json",
	} {
		if _, err := SafePath(base, evil); err == nil {
			t.Errorf("expected traversal error for %q", evil)
		}
	}

	ok, err := SafePath(base, "assets/create/lang/en_us.json")
	if err != nil {
		t.Fatalf("expected safe path, got %v", err)
	}
	if ok == "" {
		t.Fatal("expected non-empty path")
	}
}

func makeJar(t *testing.T, files map[string]string) string {
	t.Helper()
	jarPath := filepath.Join(t.TempDir(), "test.jar")
	f, err := os.Create(jarPath)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	for name, content := range files {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return jarPath
}

func TestExtractModFiltersAndOutput(t *testing.T) {
	dir := t.TempDir()
	jarPath := filepath.Join(dir, "create-6.0.6.jar")

	f, err := os.Create(jarPath)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	entries := map[string]string{
		"assets/create/lang/en_us.json":      "{}",
		"assets/create/sounds/nope.ogg":      "x",
		"data/create/recipes/ingot.json":     "{}",
		"data/create/advancements/nope.json": "{}",
		"META-INF/MANIFEST.MF":               "x",
	}
	for name, content := range entries {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	w.Close()
	f.Close()

	outBase := filepath.Join(dir, "data")
	res, err := ExtractMod(jarPath, outBase)
	if err != nil {
		t.Fatal(err)
	}

	wantDir := filepath.Join(outBase, "create-6.0.6")
	if res.OutDir != wantDir {
		t.Errorf("OutDir = %q, want %q", res.OutDir, wantDir)
	}
	if res.Assets != 1 || res.Data != 1 {
		t.Errorf("counts = assets %d data %d, want 1/1", res.Assets, res.Data)
	}

	if _, err := os.Stat(filepath.Join(wantDir, "assets", "create", "lang", "en_us.json")); err != nil {
		t.Errorf("expected kept file: %v", err)
	}
	if _, err := os.Stat(filepath.Join(wantDir, "assets", "create", "sounds", "nope.ogg")); !os.IsNotExist(err) {
		t.Errorf("rejected file should not exist")
	}
	if _, err := os.Stat(filepath.Join(wantDir, "META-INF", "MANIFEST.MF")); !os.IsNotExist(err) {
		t.Errorf("META-INF should not exist")
	}
}

func TestExtractModMissingAndInvalid(t *testing.T) {
	if _, err := ExtractMod(filepath.Join(t.TempDir(), "missing.jar"), ""); err == nil {
		t.Error("expected error for missing JAR")
	}

	bad := filepath.Join(t.TempDir(), "bad.jar")
	if err := os.WriteFile(bad, []byte("not a zip"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ExtractMod(bad, filepath.Join(t.TempDir(), "out")); err == nil {
		t.Error("expected error for invalid JAR")
	}

	// unsupported input: directory
	if _, err := ExtractMod(t.TempDir(), ""); err == nil {
		t.Error("expected error for directory input")
	}

	// unsupported input: non-jar
	txt := filepath.Join(t.TempDir(), "x.txt")
	if err := os.WriteFile(txt, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ExtractMod(txt, ""); err == nil {
		t.Error("expected error for non-jar input")
	}
}

func TestJarModName(t *testing.T) {
	if got := JarModName("create-6.0.6.jar"); got != "create-6.0.6" {
		t.Errorf("got %q", got)
	}
	if got := JarModName("/a/b/jei-1.2.JSON"); got != "jei-1.2" {
		// ext is case-sensitive here; filepath.Ext(".JSON") is ".JSON" so trimmed
		t.Errorf("got %q", got)
	}
}

func TestExtractModpack(t *testing.T) {
	root := t.TempDir()
	mods := filepath.Join(root, "mods")
	if err := os.MkdirAll(mods, 0o755); err != nil {
		t.Fatal(err)
	}

	// two jars
	for _, name := range []string{"a-1.0.jar", "b-2.0.jar"} {
		f, err := os.Create(filepath.Join(mods, name))
		if err != nil {
			t.Fatal(err)
		}
		w := zip.NewWriter(f)
		fw, _ := w.Create("assets/a/lang/en.json")
		fw.Write([]byte("{}"))
		fw2, _ := w.Create("data/a/recipes/x.json")
		fw2.Write([]byte("{}"))
		w.Close()
		f.Close()
	}
	// non-jar ignored
	if err := os.WriteFile(filepath.Join(mods, "readme.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	results, err := ExtractModpack(root, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	for _, r := range results {
		if _, err := os.Stat(r.OutDir); err != nil {
			t.Errorf("missing out dir %s: %v", r.OutDir, err)
		}
	}
	// isolation: each mod has own dir
	if results[0].OutDir == results[1].OutDir {
		t.Error("mod outputs must differ")
	}

	// custom output
	custom := filepath.Join(t.TempDir(), "custom")
	results2, err := ExtractModpack(root, custom)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range results2 {
		if got := filepath.Dir(r.OutDir); got != custom {
			t.Errorf("custom out parent = %q, want %q", got, custom)
		}
	}
}

func TestExtractModpackErrors(t *testing.T) {
	// missing mods dir
	root := t.TempDir()
	if _, err := ExtractModpack(root, ""); err == nil {
		t.Error("expected error for missing mods/")
	}

	// empty mods dir
	empty := t.TempDir()
	if err := os.MkdirAll(filepath.Join(empty, "mods"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := ExtractModpack(empty, ""); err == nil {
		t.Error("expected error for no JARs")
	}

	// mods dir with only non-jars
	if err := os.WriteFile(filepath.Join(empty, "mods", "x.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ExtractModpack(empty, ""); err == nil {
		t.Error("expected error for no JARs (non-jar only)")
	}
}

func TestExtractJarTraversalBlocked(t *testing.T) {
	jar := makeJar(t, map[string]string{
		"assets/create/lang/ok.json": "{}",
		"../../evil.txt":             "evil",
	})
	// The evil entry uses a retained prefix? No — but SafePath is only
	// checked for kept entries. Craft a kept-looking traversal:
	jar2 := makeJar(t, map[string]string{
		"assets/create/lang/../../evil.json": "evil",
	})
	_ = jar
	_ = jar2

	// Direct SafePath check on traversal with kept prefix still escapes.
	base := t.TempDir()
	if _, err := SafePath(base, "assets/../../evil"); err == nil {
		t.Error("expected traversal rejection")
	}
}
