// Package vanilla resolves, downloads and extracts vanilla Minecraft client assets.
package vanilla

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/tidjee-dev/mcae/internal/extractor"
	"github.com/tidjee-dev/mcae/internal/version"
)

// DefaultManifestURL is Mojang's version manifest endpoint.
const DefaultManifestURL = "https://piston-meta.mojang.com/mc/game/version_manifest_v2.json"

type manifest struct {
	Versions []struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	} `json:"versions"`
}

type versionDoc struct {
	Downloads struct {
		Client struct {
			URL string `json:"url"`
		} `json:"client"`
	} `json:"downloads"`
}

// ProgressFunc reports download progress (written, total; total may be -1).
type ProgressFunc func(written, total int64)

// ExtractResult summarizes a vanilla extraction.
type ExtractResult struct {
	Version string
	OutDir  string
	Assets  int
	Data    int
}

func httpClient() *http.Client {
	return &http.Client{Timeout: 60 * time.Second}
}

func getJSON(client *http.Client, url string, v any) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", version.Name+"/"+version.Version)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request %s: unexpected status %s", url, resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

// ResolveVersionURL finds the version manifest URL for a Minecraft version.
func ResolveVersionURL(client *http.Client, manifestURL, versionID string) (string, error) {
	if client == nil {
		client = httpClient()
	}
	if manifestURL == "" {
		manifestURL = DefaultManifestURL
	}

	var m manifest
	if err := getJSON(client, manifestURL, &m); err != nil {
		return "", fmt.Errorf("fetch version manifest: %w", err)
	}

	for _, v := range m.Versions {
		if v.ID == versionID {
			return v.URL, nil
		}
	}

	return "", fmt.Errorf("minecraft version not found: %s", versionID)
}

// ResolveClientURL resolves the client JAR download URL for a version.
func ResolveClientURL(client *http.Client, manifestURL, versionID string) (string, error) {
	if client == nil {
		client = httpClient()
	}

	versionURL, err := ResolveVersionURL(client, manifestURL, versionID)
	if err != nil {
		return "", err
	}

	var doc versionDoc
	if err := getJSON(client, versionURL, &doc); err != nil {
		return "", fmt.Errorf("fetch version details: %w", err)
	}

	if doc.Downloads.Client.URL == "" {
		return "", fmt.Errorf("client download not found for version: %s", versionID)
	}

	return doc.Downloads.Client.URL, nil
}

// DownloadClient downloads the client JAR to a temp file.
func DownloadClient(client *http.Client, clientURL string, onProgress ProgressFunc) (string, func(), error) {
	if client == nil {
		client = httpClient()
	}

	req, err := http.NewRequest("GET", clientURL, nil)
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("User-Agent", version.Name+"/"+version.Version)

	resp, err := client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("download client JAR: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("download client JAR: unexpected status %s", resp.Status)
	}

	tmp, err := os.CreateTemp("", "mcae-vanilla-*.jar")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { os.Remove(tmp.Name()) }

	var written int64
	total := resp.ContentLength
	buf := make([]byte, 128*1024)
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := tmp.Write(buf[:n]); werr != nil {
				tmp.Close()
				cleanup()
				return "", nil, werr
			}
			written += int64(n)
			if onProgress != nil {
				onProgress(written, total)
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			tmp.Close()
			cleanup()
			return "", nil, fmt.Errorf("download client JAR: %w", rerr)
		}
	}

	if err := tmp.Close(); err != nil {
		cleanup()
		return "", nil, err
	}

	return tmp.Name(), cleanup, nil
}

// ExtractVanilla resolves, downloads and extracts a vanilla client JAR.
// outputBase is the base directory; the final dir is
// <outputBase>/vanilla/<version> (default base: ./extracted).
func ExtractVanilla(versionID, outputBase string, onProgress ProgressFunc) (ExtractResult, error) {
	return ExtractVanillaWithManifest(versionID, outputBase, DefaultManifestURL, onProgress)
}

// ExtractVanillaWithForce is ExtractVanilla with an explicit overwrite policy.
func ExtractVanillaWithForce(versionID, outputBase string, force bool, onProgress ProgressFunc) (ExtractResult, error) {
	return ExtractVanillaWithManifestAndForce(versionID, outputBase, DefaultManifestURL, force, onProgress)
}

// ExtractVanillaWithManifest is ExtractVanilla with an injectable manifest URL (for tests).
func ExtractVanillaWithManifest(versionID, outputBase, manifestURL string, onProgress ProgressFunc) (ExtractResult, error) {
	return ExtractVanillaWithManifestAndForce(versionID, outputBase, manifestURL, false, onProgress)
}

// ExtractVanillaWithManifestAndForce adds force + injectable manifest URL.
func ExtractVanillaWithManifestAndForce(versionID, outputBase, manifestURL string, force bool, onProgress ProgressFunc) (ExtractResult, error) {
	if versionID == "" {
		return ExtractResult{}, fmt.Errorf("minecraft version is required")
	}

	base := outputBase
	if base == "" {
		base = "extracted"
	}
	dest := filepath.Join(base, "vanilla", versionID)

	if force {
		if err := os.RemoveAll(filepath.Clean(dest)); err != nil {
			return ExtractResult{}, fmt.Errorf("clean output directory: %w", err)
		}
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return ExtractResult{}, fmt.Errorf("create output directory: %w", err)
	}

	client := httpClient()

	clientURL, err := ResolveClientURL(client, manifestURL, versionID)
	if err != nil {
		return ExtractResult{}, err
	}

	jarPath, cleanup, err := DownloadClient(client, clientURL, onProgress)
	if err != nil {
		return ExtractResult{}, err
	}
	defer cleanup()

	return ExtractDownloadedJarWithManifest(versionID, outputBase, force, jarPath)
}

// ExtractDownloadedJar extracts an already-downloaded client JAR.
func ExtractDownloadedJar(versionID, outputBase string, force bool, jarPath string) (ExtractResult, error) {
	return ExtractDownloadedJarWithManifest(versionID, outputBase, force, jarPath)
}

// ExtractDownloadedJarWithManifest extracts a local client JAR to the
// vanilla output directory (used by the live TUI worker).
func ExtractDownloadedJarWithManifest(versionID, outputBase string, force bool, jarPath string) (ExtractResult, error) {
	if versionID == "" {
		return ExtractResult{}, fmt.Errorf("minecraft version is required")
	}

	base := outputBase
	if base == "" {
		base = "extracted"
	}
	dest := filepath.Join(base, "vanilla", versionID)

	if force {
		if err := os.RemoveAll(filepath.Clean(dest)); err != nil {
			return ExtractResult{}, fmt.Errorf("clean output directory: %w", err)
		}
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return ExtractResult{}, fmt.Errorf("create output directory: %w", err)
	}

	assets, data, err := extractor.ExtractJar(jarPath, dest)
	if err != nil {
		return ExtractResult{}, fmt.Errorf("extract vanilla %s: %w", versionID, err)
	}

	return ExtractResult{
		Version: versionID,
		OutDir:  dest,
		Assets:  assets,
		Data:    data,
	}, nil
}
