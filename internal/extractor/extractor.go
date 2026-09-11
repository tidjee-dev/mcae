package extractor

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Sentinel errors for stable handling and tests.
var (
	ErrModNotFound    = errors.New("mod file does not exist")
	ErrModsNotFound   = errors.New("mods directory not found")
	ErrNoJars         = errors.New("no mod JARs found")
	ErrInvalidInput   = errors.New("unsupported input")
	ErrInvalidModpack = errors.New("invalid modpack")
)

// Result summarizes a single JAR extraction.
type Result struct {
	JarName string
	OutDir  string
	Assets  int
	Data    int
}

// JarModName converts a JAR filename to an output directory name.
// Example: create-6.0.6.jar -> create-6.0.6
func JarModName(jarPath string) string {
	base := filepath.Base(jarPath)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

// ShouldKeep reports whether an archive entry should be extracted.
// Only the Minecraft resources listed in the spec are retained.
// The namespace (parts[1]) stays dynamic.
func ShouldKeep(name string) bool {
	name = filepath.ToSlash(name)
	parts := strings.Split(name, "/")

	if len(parts) < 3 {
		return false
	}

	switch parts[0] {
	case "assets":
		switch parts[2] {
		case "lang":
			return true

		case "models":
			return len(parts) >= 4 &&
				(parts[3] == "block" || parts[3] == "item")

		case "textures":
			return len(parts) >= 4 &&
				(parts[3] == "block" ||
					parts[3] == "entity" ||
					parts[3] == "item")
		}

	case "data":
		switch parts[2] {
		case "loot_tables", "recipes":
			return true

		case "tags":
			return len(parts) >= 4 &&
				(parts[3] == "blocks" || parts[3] == "items")
		}
	}

	return false
}

// SafePath joins base with an archive entry name and rejects
// paths that escape the destination (Zip Slip protection).
func SafePath(base, name string) (string, error) {
	baseAbs, err := filepath.Abs(base)
	if err != nil {
		return "", err
	}

	targetAbs, err := filepath.Abs(
		filepath.Join(baseAbs, filepath.FromSlash(name)),
	)
	if err != nil {
		return "", err
	}

	rel, err := filepath.Rel(baseAbs, targetAbs)
	if err != nil {
		return "", err
	}

	if rel == ".." ||
		strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("unsafe path in archive: %s", name)
	}

	return targetAbs, nil
}

// ExtractJar extracts a single JAR/ZIP file to dest, applying the
// retain filter. It returns the number of asset and data files written.
func ExtractJar(jarPath, dest string) (assets, data int, err error) {
	reader, err := zip.OpenReader(jarPath)
	if err != nil {
		return 0, 0, fmt.Errorf("open JAR: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		if !ShouldKeep(file.Name) {
			continue
		}

		target, err := SafePath(dest, file.Name)
		if err != nil {
			return 0, 0, err
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return 0, 0, err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return 0, 0, err
		}

		src, err := file.Open()
		if err != nil {
			return 0, 0, err
		}

		dst, err := os.Create(target)
		if err != nil {
			src.Close()
			return 0, 0, err
		}

		_, copyErr := io.Copy(dst, src)

		src.Close()
		dst.Close()

		if copyErr != nil {
			return 0, 0, copyErr
		}

		slash := filepath.ToSlash(file.Name)
		if strings.HasPrefix(slash, "assets/") {
			assets++
		} else if strings.HasPrefix(slash, "data/") {
			data++
		}
	}

	return assets, data, nil
}

// ExtractMod extracts a single mod JAR.
// If outputBase is empty, "./extracted/<modname>" is used,
// otherwise "<outputBase>/<modname>".
func ExtractMod(jarPath, outputBase string) (Result, error) {
	return ExtractModWithForce(jarPath, outputBase, false)
}

// ExtractModWithForce is ExtractMod with an explicit overwrite policy.
// With force=true the destination directory is removed first.
func ExtractModWithForce(jarPath, outputBase string, force bool) (Result, error) {
	info, err := os.Stat(jarPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Result{}, fmt.Errorf("%w: %s", ErrModNotFound, jarPath)
		}
		return Result{}, fmt.Errorf("stat mod file: %w", err)
	}
	if info.IsDir() {
		return Result{}, fmt.Errorf("%w (expected a .jar file): %s", ErrInvalidInput, jarPath)
	}
	if strings.ToLower(filepath.Ext(jarPath)) != ".jar" {
		return Result{}, fmt.Errorf("%w (expected a .jar file): %s", ErrInvalidInput, jarPath)
	}

	modName := JarModName(jarPath)
	var dest string
	if outputBase == "" {
		dest = filepath.Join("extracted", modName)
	} else {
		dest = filepath.Join(outputBase, modName)
	}

	if force {
		if err := resetDir(dest); err != nil {
			return Result{}, err
		}
	} else if err := os.MkdirAll(dest, 0o755); err != nil {
		return Result{}, fmt.Errorf("create output directory: %w", err)
	}

	assets, data, err := ExtractJar(jarPath, dest)
	if err != nil {
		return Result{}, fmt.Errorf("extract %s: %w", filepath.Base(jarPath), err)
	}

	return Result{
		JarName: filepath.Base(jarPath),
		OutDir:  dest,
		Assets:  assets,
		Data:    data,
	}, nil
}

// resetDir removes dest (if present) and recreates it.
// The path is cleaned and must be non-empty to avoid catastrophic deletes.
func resetDir(dest string) error {
	clean := filepath.Clean(dest)
	if clean == "" || clean == "." || clean == "/" {
		return fmt.Errorf("refuse to reset unsafe directory: %q", dest)
	}
	if err := os.RemoveAll(clean); err != nil {
		return fmt.Errorf("clean output directory: %w", err)
	}
	if err := os.MkdirAll(clean, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	return nil
}

// ExtractModpack extracts every .jar directly inside <modpack>/mods.
// If outputBase is empty, "<modpack>/extracted" is used.
func ExtractModpack(modpackDir, outputBase string) ([]Result, error) {
	return ExtractModpackWithForce(modpackDir, outputBase, false, nil)
}

// ModProgress is called after each JAR finishes (index is 0-based).
type ModProgress func(index, total int, res Result)

// ListModJars returns the .jar base names directly inside <modpack>/mods
// plus the resolved mods and output directories.
func ListModJars(modpackDir, outputBase string) (modsDir, outputDir string, jars []string, err error) {
	info, err := os.Stat(modpackDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", nil, fmt.Errorf("modpack does not exist: %s", modpackDir)
		}
		return "", "", nil, fmt.Errorf("stat modpack: %w", err)
	}
	if !info.IsDir() {
		return "", "", nil, fmt.Errorf("%w (not a directory): %s", ErrInvalidModpack, modpackDir)
	}

	modsDir = filepath.Join(modpackDir, "mods")
	modsInfo, err := os.Stat(modsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", nil, fmt.Errorf("%w: mods directory not found: %s", ErrInvalidModpack, modsDir)
		}
		return "", "", nil, fmt.Errorf("stat mods directory: %w", err)
	}
	if !modsInfo.IsDir() {
		return "", "", nil, fmt.Errorf("%w: mods directory not found: %s", ErrInvalidModpack, modsDir)
	}

	outputDir = outputBase
	if outputDir == "" {
		outputDir = filepath.Join(modpackDir, "extracted")
	}

	entries, err := os.ReadDir(modsDir)
	if err != nil {
		return "", "", nil, fmt.Errorf("mods directory does not exist: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.ToLower(filepath.Ext(entry.Name())) != ".jar" {
			continue
		}
		jars = append(jars, entry.Name())
	}
	if len(jars) == 0 {
		return "", "", nil, fmt.Errorf("%w in %s", ErrNoJars, modsDir)
	}
	return modsDir, outputDir, jars, nil
}

// ExtractModpackWithForce streams per-mod results via onMod.
// With force=true each mod destination is removed first.
func ExtractModpackWithForce(modpackDir, outputBase string, force bool, onMod ModProgress) ([]Result, error) {
	modsDir, outputDir, jars, err := ListModJars(modpackDir, outputBase)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, fmt.Errorf("create output directory: %w", err)
	}

	var results []Result
	for i, name := range jars {
		jarPath := filepath.Join(modsDir, name)
		dest := filepath.Join(outputDir, JarModName(name))

		if force {
			if err := resetDir(dest); err != nil {
				return nil, err
			}
		} else if err := os.MkdirAll(dest, 0o755); err != nil {
			return nil, fmt.Errorf("create output directory: %w", err)
		}

		assets, data, err := ExtractJar(jarPath, dest)
		if err != nil {
			return nil, fmt.Errorf("extract %s: %w", name, err)
		}

		res := Result{
			JarName: name,
			OutDir:  dest,
			Assets:  assets,
			Data:    data,
		}
		results = append(results, res)
		if onMod != nil {
			onMod(i, len(jars), res)
		}
	}

	return results, nil
}
