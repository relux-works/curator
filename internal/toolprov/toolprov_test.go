package toolprov

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildTarGz(t *testing.T, files map[string]struct {
	body []byte
	mode int64
}) ([]byte, string) {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, f := range files {
		hdr := &tar.Header{
			Name:     name,
			Mode:     f.mode,
			Size:     int64(len(f.body)),
			Typeflag: tar.TypeReg,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(f.body); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(buf.Bytes())
	return buf.Bytes(), hex.EncodeToString(sum[:])
}

func TestEnsureGoToolchainProvisionAndReuse(t *testing.T) {
	t.Setenv("CURATOR_GO", "")
	t.Setenv("GOROOT", "")
	t.Setenv(EnvDisableAutoToolchain, "")

	tarBytes, sum := buildTarGz(t, map[string]struct {
		body []byte
		mode int64
	}{
		"go/bin/go": {body: []byte("#!/bin/sh\necho go1.25.5\n"), mode: 0o755},
	})
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if r.URL.Path != "/go1.25.5.test-plat.tar.gz" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(tarBytes)
	}))
	defer srv.Close()

	home := t.TempDir()
	opts := GoOptions{
		Platform:         "test-plat",
		BaseURL:          srv.URL,
		HTTPClient:       srv.Client(),
		SkipSystemSearch: true,
		Pins: map[string]goReleasePin{
			"test-plat": {Version: "1.25.5", Filename: "go1.25.5.test-plat.tar.gz", SHA256: sum},
		},
	}

	got, err := EnsureGoToolchain(context.Background(), home, opts)
	if err != nil {
		t.Fatalf("EnsureGoToolchain failed: %v", err)
	}
	if !strings.HasSuffix(got, filepath.Join("go", "1.25.5", "test-plat", "go", "bin", "go")) {
		t.Fatalf("unexpected go path: %s", got)
	}
	if hits != 1 {
		t.Fatalf("hits = %d, want 1", hits)
	}

	// Second call reuses verified sealed store without network hits.
	got2, err := EnsureGoToolchain(context.Background(), home, opts)
	if err != nil {
		t.Fatalf("second EnsureGoToolchain failed: %v", err)
	}
	if got2 != got || hits != 1 {
		t.Fatalf("expected cache reuse: got2=%s hits=%d", got2, hits)
	}
}

func TestEnsureGoToolchainExplicitOverrideAndRefusals(t *testing.T) {
	home := t.TempDir()

	// Explicit CURATOR_GO must bypass auto-provisioning.
	t.Setenv("CURATOR_GO", "/explicit/go")
	got, err := EnsureGoToolchain(context.Background(), home, GoOptions{SkipSystemSearch: true})
	if err != nil || got != "" {
		t.Fatalf("explicit CURATOR_GO returned (%q, %v), want (\"\", nil)", got, err)
	}
	t.Setenv("CURATOR_GO", "")
	t.Setenv("GOROOT", "")

	// Disabled auto-download refuses without network.
	t.Setenv(EnvDisableAutoToolchain, "1")
	if _, err := EnsureGoToolchain(context.Background(), home, GoOptions{SkipSystemSearch: true}); err == nil || !strings.Contains(err.Error(), "toolchain_provisioning_disabled") {
		t.Fatalf("expected disabled refusal, got %v", err)
	}
	t.Setenv(EnvDisableAutoToolchain, "")

	// SHA-256 mismatch refuses and leaves no store entry.
	tarBytes, _ := buildTarGz(t, map[string]struct {
		body []byte
		mode int64
	}{"go/bin/go": {body: []byte("fake"), mode: 0o755}})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(tarBytes)
	}))
	defer srv.Close()

	_, err = EnsureGoToolchain(context.Background(), home, GoOptions{
		Platform:         "test-plat",
		BaseURL:          srv.URL,
		HTTPClient:       srv.Client(),
		SkipSystemSearch: true,
		Pins: map[string]goReleasePin{
			"test-plat": {Version: "1.25.5", Filename: "go.tar.gz", SHA256: strings.Repeat("0", 64)},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "toolchain_provisioning_digest_mismatch") {
		t.Fatalf("expected digest mismatch refusal, got %v", err)
	}
}

func TestSafeExtractRejectsTraversal(t *testing.T) {
	tarBytes, _ := buildTarGz(t, map[string]struct {
		body []byte
		mode int64
	}{
		"../escape": {body: []byte("boom"), mode: 0o644},
	})
	archive := filepath.Join(t.TempDir(), "evil.tar.gz")
	if err := os.WriteFile(archive, tarBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := SafeExtractTarGz(archive, filepath.Join(t.TempDir(), "out")); err == nil || !strings.Contains(err.Error(), "unsafe archive member") {
		t.Fatalf("expected traversal refusal, got %v", err)
	}
}

func TestProvisionDynamicAquaToolAndRefusals(t *testing.T) {
	tarBytes, sum := buildTarGz(t, map[string]struct {
		body []byte
		mode int64
	}{
		"demo-1.2.3/demo": {body: []byte("#!/bin/sh\necho demo\n"), mode: 0o755},
	})

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/pkgs/acme/demo/registry.yaml":
			_, _ = w.Write([]byte(`packages:
  - type: github_release
    repo_owner: acme
    repo_name: demo
    asset: demo-{{trimV .Version}}-{{.OS}}-{{.Arch}}.{{.Format}}
    format: tar.gz
    files:
      - name: demo
        src: demo-{{trimV .Version}}/demo
    checksum:
      type: github_release
      asset: checksums.txt
      algorithm: sha256
`))
		case "/pkgs/acme/nochecksum/registry.yaml":
			_, _ = w.Write([]byte(`packages:
  - type: github_release
    repo_owner: acme
    repo_name: nochecksum
    asset: nc.tar.gz
`))
		case "/repos/acme/demo/releases/latest":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"tag_name": "v1.2.3",
				"assets": []map[string]any{
					{"name": "demo-1.2.3-linux-amd64.tar.gz", "browser_download_url": srv.URL + "/dl/demo.tar.gz"},
					{"name": "checksums.txt", "browser_download_url": srv.URL + "/dl/checksums.txt"},
				},
			})
		case "/dl/demo.tar.gz":
			_, _ = w.Write(tarBytes)
		case "/dl/checksums.txt":
			_, _ = w.Write([]byte(sum + "  demo-1.2.3-linux-amd64.tar.gz\n"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	home := t.TempDir()
	opts := CatalogOptions{
		Platform:            "linux-amd64",
		HTTPClient:          srv.Client(),
		AquaRegistryBaseURL: srv.URL,
		GitHubAPIBaseURL:    srv.URL,
	}

	got, err := ProvisionTool(context.Background(), home, "demo", "aqua/acme/demo", "latest", opts)
	if err != nil {
		t.Fatalf("ProvisionTool aqua/acme/demo failed: %v", err)
	}
	if got.Version != "1.2.3" || got.Backend != "aqua" || got.NeedsForwarder {
		t.Fatalf("unexpected provisioned tool: %+v", got)
	}
	if _, err := os.Stat(got.ExecutablePath); err != nil {
		t.Fatalf("provisioned executable missing: %v", err)
	}

	// Missing checksum in Aqua registry refuses closed.
	if _, err := ProvisionTool(context.Background(), home, "nc", "aqua/acme/nochecksum", "latest", opts); err == nil || !strings.Contains(err.Error(), "tool_provisioning_unverified") {
		t.Fatalf("expected unverified refusal, got %v", err)
	}

	// Unknown PyPI package refuses closed.
	if _, err := ProvisionTool(context.Background(), home, "evil", "pypi/not-in-allowlist", "1.0.0", opts); err == nil || !strings.Contains(err.Error(), "tool_catalog_unknown") {
		t.Fatalf("expected catalog unknown refusal, got %v", err)
	}
}

func TestLiveGoToolchainAutoDownload(t *testing.T) {
	if os.Getenv("CURATOR_LIVE_TOOL_TESTS") != "1" {
		t.Skip("set CURATOR_LIVE_TOOL_TESTS=1 to run live go.dev download verification")
	}
	t.Setenv("CURATOR_GO", "")
	t.Setenv("GOROOT", "")
	home := t.TempDir()
	goBin, err := EnsureGoToolchain(context.Background(), home, GoOptions{SkipSystemSearch: true})
	if err != nil {
		t.Fatalf("live EnsureGoToolchain failed: %v", err)
	}
	out, err := exec.Command(goBin, "version").CombinedOutput() // #nosec G204 -- verified freshly provisioned go binary
	if err != nil {
		t.Fatalf("provisioned go version failed: %v\n%s", err, string(out))
	}
	if !strings.Contains(string(out), "go1.25.5") {
		t.Fatalf("unexpected go version output: %s", string(out))
	}
	t.Logf("live provisioned Go verified: %s (%s)", goBin, strings.TrimSpace(string(out)))
}
