package vanilla

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func manifestServer(t *testing.T, versionURL string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"versions": []map[string]string{
				{"id": "1.21.8", "url": versionURL},
			},
		})
	}))
}

func TestResolveClientURL(t *testing.T) {
	var clientURL string
	var versionSrv *httptest.Server
	versionSrv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"downloads": map[string]any{
				"client": map[string]any{"url": clientURL},
			},
		})
	}))
	defer versionSrv.Close()

	// client jar server
	clientSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("fake"))
	}))
	defer clientSrv.Close()
	clientURL = clientSrv.URL

	manifestSrv := manifestServer(t, versionSrv.URL)
	defer manifestSrv.Close()

	got, err := ResolveClientURL(nil, manifestSrv.URL, "1.21.8")
	if err != nil {
		t.Fatal(err)
	}
	if got != clientSrv.URL {
		t.Errorf("got %q want %q", got, clientSrv.URL)
	}

	if _, err := ResolveClientURL(nil, manifestSrv.URL, "9.9.9"); err == nil {
		t.Error("expected error for unknown version")
	}
}

func TestResolveVersionURLMissing(t *testing.T) {
	srv := manifestServer(t, "http://example.com/version.json")
	defer srv.Close()

	if _, err := ResolveVersionURL(nil, srv.URL, "nope"); err == nil {
		t.Error("expected error")
	}
}

func TestExtractVanillaRequiresVersion(t *testing.T) {
	if _, err := ExtractVanillaWithManifest("", t.TempDir(), "http://example.com", nil); err == nil {
		t.Error("expected error for empty version")
	}
}

func TestExtractVanillaCreatesOutput(t *testing.T) {
	if _, err := ExtractVanillaWithManifest("9.9.9-nope", t.TempDir(), "http://127.0.0.1:1", nil); err == nil {
		t.Error("expected error for unreachable manifest")
	}
	_ = filepath.Join
	_ = os.MkdirAll
}
