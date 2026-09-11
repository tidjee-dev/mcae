package inspect

import (
	"archive/zip"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tidjee-dev/mcae/internal/extractor"
)

// Asset categories shown in the summary.
const (
	CatLang           = "lang"
	CatModelsBlock    = "models/block"
	CatModelsItem     = "models/item"
	CatTexturesBlock  = "textures/block"
	CatTexturesEntity = "textures/entity"
	CatTexturesItem   = "textures/item"
	CatRecipes        = "recipes"
	CatLootTables     = "loot_tables"
	CatTagsBlocks     = "tags/blocks"
	CatTagsItems      = "tags/items"
)

// Summary is the per-mod (or per-directory) asset census.
type Summary struct {
	Path       string
	Namespaces []string
	Assets     map[string]int
	Data       map[string]int
}

// ModpackSummary is a per-mod breakdown for a modpack directory.
type ModpackSummary struct {
	Path string
	Mods []Summary
}

// Report is either a single summary or a modpack breakdown.
type Report struct {
	Single  *Summary
	Modpack *ModpackSummary
}

func newSummary(path string) *Summary {
	return &Summary{
		Path:   path,
		Assets: map[string]int{},
		Data:   map[string]int{},
	}
}

// Categorize maps a retained archive/relative path to (section, category).
// ok=false means the path is not retained.
func Categorize(name string) (section, category string, ok bool) {
	if !extractor.ShouldKeep(name) {
		return "", "", false
	}
	slash := filepath.ToSlash(name)
	parts := strings.Split(slash, "/")
	if len(parts) < 3 {
		return "", "", false
	}

	switch parts[0] {
	case "assets":
		switch parts[2] {
		case "lang":
			return "assets", CatLang, true
		case "models":
			if len(parts) >= 4 && parts[3] == "block" {
				return "assets", CatModelsBlock, true
			}
			return "assets", CatModelsItem, true
		case "textures":
			if len(parts) >= 4 {
				switch parts[3] {
				case "block":
					return "assets", CatTexturesBlock, true
				case "entity":
					return "assets", CatTexturesEntity, true
				default:
					return "assets", CatTexturesItem, true
				}
			}
		}
	case "data":
		switch parts[2] {
		case "recipes":
			return "data", CatRecipes, true
		case "loot_tables":
			return "data", CatLootTables, true
		case "tags":
			if len(parts) >= 4 && parts[3] == "blocks" {
				return "data", CatTagsBlocks, true
			}
			return "data", CatTagsItems, true
		}
	}

	return "", "", false
}

func (s *Summary) add(name string) {
	section, category, ok := Categorize(name)
	if !ok {
		return
	}
	parts := strings.Split(filepath.ToSlash(name), "/")
	ns := ""
	if len(parts) >= 2 {
		ns = parts[1]
	}
	if ns != "" {
		found := false
		for _, existing := range s.Namespaces {
			if existing == ns {
				found = true
				break
			}
		}
		if !found {
			s.Namespaces = append(s.Namespaces, ns)
		}
	}

	if section == "assets" {
		s.Assets[category]++
	} else {
		s.Data[category]++
	}
}

func (s *Summary) finalize() {
	sort.Strings(s.Namespaces)
}

func summarizeJar(jarPath string) (*Summary, error) {
	reader, err := zip.OpenReader(jarPath)
	if err != nil {
		return nil, fmt.Errorf("open JAR: %w", err)
	}
	defer reader.Close()

	s := newSummary(jarPath)
	for _, f := range reader.File {
		if f.FileInfo().IsDir() {
			continue
		}
		s.add(f.Name)
	}
	s.finalize()

	return s, nil
}

func summarizeDir(dir string) (*Summary, error) {
	s := newSummary(dir)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		s.add(filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.finalize()

	return s, nil
}

// Inspect accepts a .jar file, an extracted directory, or a modpack
// directory (containing mods/). It returns a single or modpack report.
func Inspect(path string) (Report, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Report{}, fmt.Errorf("unsupported input (path does not exist): %s", path)
		}
		return Report{}, err
	}

	if !info.IsDir() {
		if strings.ToLower(filepath.Ext(path)) != ".jar" {
			return Report{}, fmt.Errorf("unsupported input (expected a .jar file or directory): %s", path)
		}
		s, err := summarizeJar(path)
		if err != nil {
			return Report{}, err
		}
		return Report{Single: s}, nil
	}

	// Directory: modpack detection — <path>/mods exists.
	modsDir := filepath.Join(path, "mods")
	if st, err := os.Stat(modsDir); err == nil && st.IsDir() {
		entries, err := os.ReadDir(modsDir)
		if err != nil {
			return Report{}, err
		}
		mp := &ModpackSummary{Path: path}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if strings.ToLower(filepath.Ext(e.Name())) != ".jar" {
				continue
			}
			s, err := summarizeJar(filepath.Join(modsDir, e.Name()))
			if err != nil {
				return Report{}, fmt.Errorf("inspect %s: %w", e.Name(), err)
			}
			mp.Mods = append(mp.Mods, *s)
		}
		// Also consider already-extracted dirs: <path>/extracted/* for JARs
		// missing from mods/ is out of scope; keep it simple.
		sort.Slice(mp.Mods, func(i, j int) bool { return mp.Mods[i].Path < mp.Mods[j].Path })
		return Report{Modpack: mp}, nil
	}

	s, err := summarizeDir(path)
	if err != nil {
		return Report{}, err
	}
	return Report{Single: s}, nil
}
