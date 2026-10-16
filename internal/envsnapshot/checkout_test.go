package envsnapshot

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// A real autocrlf checkout retains the published schema bytes with our actual
// attributes; an unprotected checkout converts them. No commit or source write
// is needed, and expected hashes come from the independent published pins.
func TestPinnedSchemaCheckoutBytes(t *testing.T) {
	attributes, err := os.ReadFile(".gitattributes")
	if err != nil {
		t.Fatal(err)
	}
	var pins struct {
		Files []struct{ Path, SHA256 string }
	}
	if err := json.Unmarshal(read(t, "manifest.json"), &pins); err != nil {
		t.Fatal(err)
	}
	for _, protected := range []bool{true, false} {
		name := "protected"
		if !protected {
			name = "unprotected-witness"
		}
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			git := func(input []byte, args ...string) []byte {
				t.Helper()
				options := []string{"-c", "core.autocrlf=true", "-c", "core.safecrlf=false", "-c", "core.attributesFile=" + os.DevNull, "-C", root}
				cmd := exec.Command("git", append(options, args...)...)
				for _, entry := range os.Environ() {
					if !strings.HasPrefix(entry, "GIT_") {
						cmd.Env = append(cmd.Env, entry)
					}
				}
				cmd.Env = append(cmd.Env, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_ATTR_NOSYSTEM=1")
				cmd.Stdin = bytes.NewReader(input)
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("git %s: %v\n%s", args[0], err, out)
				}
				return out
			}
			git(nil, "init", "--quiet")
			if protected {
				if err := os.WriteFile(filepath.Join(root, ".gitattributes"), attributes, 0600); err != nil {
					t.Fatal(err)
				}
			}
			count := 0
			for _, pin := range pins.Files {
				if !strings.HasPrefix(pin.Path, "schemas/managed-home-snapshot-v1/") {
					continue
				}
				raw, err := schemas.ReadFile(pin.Path)
				if err != nil {
					t.Fatal(err)
				}
				oid := strings.TrimSpace(string(git(raw, "hash-object", "-w", "--no-filters", "--stdin")))
				git(nil, "update-index", "--add", "--cacheinfo", "100644,"+oid+","+pin.Path)
				count++
			}
			if count != 3 {
				t.Fatalf("schema pin inventory = %d, want 3", count)
			}
			git(nil, "checkout-index", "--all")
			for _, pin := range pins.Files {
				if !strings.HasPrefix(pin.Path, "schemas/managed-home-snapshot-v1/") {
					continue
				}
				actual, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(pin.Path)))
				if err != nil {
					t.Fatal(err)
				}
				equal := "sha256:"+Digest(actual) == pin.SHA256
				if protected && !equal {
					t.Fatalf("checkout changed pinned bytes: %s", pin.Path)
				}
				if !protected && (equal || !bytes.Contains(actual, []byte("\r\n"))) {
					t.Fatalf("unprotected checkout did not demonstrate CRLF conversion: %s", pin.Path)
				}
			}
		})
	}
}
