package envprofile

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
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envfragment"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/toolprov"
)

func buildTestTarGz(t *testing.T, files map[string][]byte) ([]byte, string) {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, body := range files {
		hdr := &tar.Header{Name: name, Mode: 0o755, Size: int64(len(body)), Typeflag: tar.TypeReg}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(body); err != nil {
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

func TestProfileCommandRootScriptAndAquaToolAndFragmentV9(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX script/forwarder test")
	}
	home := gitFixtureHome(t)
	pinHomes(t)
	ids := newGitIdentities(t)

	toolTar, toolSum := buildTestTarGz(t, map[string][]byte{
		"demo-1.0.0/demo": []byte("#!/bin/sh\necho aqua-demo-ok\n"),
	})
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/pkgs/acme/demo/registry.yaml":
			_, _ = w.Write([]byte(`packages:
  - type: github_release
    repo_owner: acme
    repo_name: demo
    asset: demo-{{.OS}}-{{.Arch}}.tar.gz
    format: tar.gz
    files:
      - name: demo
        src: demo-1.0.0/demo
    checksum:
      type: github_release
      asset: checksums.txt
      algorithm: sha256
`))
		case "/repos/acme/demo/releases/latest":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"tag_name": "v1.0.0",
				"assets": []map[string]any{
					{"name": "demo-" + runtime.GOOS + "-" + runtime.GOARCH + ".tar.gz", "browser_download_url": srv.URL + "/dl/demo.tar.gz"},
					{"name": "checksums.txt", "browser_download_url": srv.URL + "/dl/checksums.txt"},
				},
			})
		case "/dl/demo.tar.gz":
			_, _ = w.Write(toolTar)
		case "/dl/checksums.txt":
			_, _ = w.Write([]byte(toolSum + "  demo-" + runtime.GOOS + "-" + runtime.GOARCH + ".tar.gz\n"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	skillRepo := gitRepo(t, map[string]string{
		"SKILL.md":         "---\nname: cmd-skill\ndescription: Command root test skill\n---\n# Cmd\n",
		"scripts/hi.sh":    "#!/bin/sh\necho script-hi-ok\n",
		"agent-skill.json": `{"schema_version":10,"runtime_roots":["scripts"],"commands":{"hi-cmd":{"type":"script","unix_path":"scripts/hi.sh"}},"capabilities":{"exec":["demo"]},"dependencies":{"tools":{"demo":{"id":"aqua/acme/demo","version":"latest"}}}}`,
	}, "v1.0.0")
	skillURL := ids.serve(skillRepo, "https://example.com/skills/cmd-skill")

	profileDir := t.TempDir()
	writeManifestPackage(t, profileDir, `{"schema_version":2,"name":"cmd-profile","version":"1.0.0","requires":{"skills":{"cmd-skill":{"git":"`+skillURL+`","range":"*"}}}}`, nil)

	info, _, _, err := Install(home, InstallOptions{Operand: profileDir, As: "cmd-profile"})
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}
	active, err := MaterializeCommandRoot(context.Background(), home, info.Name, Policy{}, CommandRootOptions{
		Catalog: toolprov.CatalogOptions{
			HTTPClient:          srv.Client(),
			AquaRegistryBaseURL: srv.URL,
			GitHubAPIBaseURL:    srv.URL,
		},
	})
	if err != nil {
		t.Fatalf("MaterializeCommandRoot failed: %v", err)
	}
	if active == nil {
		t.Fatal("expected non-nil active command root")
	}

	// Both script command and Aqua tool are directly executable from active.Root.
	outHi, err := exec.Command(filepath.Join(active.Root, "hi-cmd")).CombinedOutput()
	if err != nil || strings.TrimSpace(string(outHi)) != "script-hi-ok" {
		t.Fatalf("hi-cmd = (%q, %v), want script-hi-ok", string(outHi), err)
	}
	outDemo, err := exec.Command(filepath.Join(active.Root, "demo")).CombinedOutput()
	if err != nil || strings.TrimSpace(string(outDemo)) != "aqua-demo-ok" {
		t.Fatalf("demo = (%q, %v), want aqua-demo-ok", string(outDemo), err)
	}

	// Resolving managed home now emits launch-env-fragment-v9 with command_root + path_prepend.
	res, err := Resolve(ResolveRequest{
		Home:    home,
		Profile: "cmd-profile",
		EnvID:   envregistry.ClaudeCode,
		Repair:  true,
		Format:  "json",
	})
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(res.Document, &doc); err != nil {
		t.Fatalf("unmarshal fragment: %v", err)
	}
	if doc["fragment"] != envfragment.VersionV9 {
		t.Fatalf("fragment = %v, want %s", doc["fragment"], envfragment.VersionV9)
	}
	if doc["path_prepend"] != active.Root {
		t.Fatalf("path_prepend = %v, want %s", doc["path_prepend"], active.Root)
	}
	if _, ok := doc["command_root"].(map[string]any); !ok {
		t.Fatalf("missing command_root in v9 fragment: %v", doc)
	}

	// Tampering with an executable in bin/ must fail closed on next resolve.
	if err := os.Chmod(filepath.Join(active.Root, "demo"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(active.Root, "demo"), []byte("tampered"), 0o500); err != nil {
		t.Fatal(err)
	}
	if _, err := Resolve(ResolveRequest{
		Home:    home,
		Profile: "cmd-profile",
		EnvID:   envregistry.ClaudeCode,
		Format:  "json",
	}); err == nil || !strings.Contains(err.Error(), "command_root_invalid") {
		t.Fatalf("expected command_root_invalid after tampering, got %v", err)
	}
}
