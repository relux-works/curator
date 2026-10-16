//go:build darwin || linux || freebsd || openbsd || netbsd || dragonfly

package envsnapshot

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Opt-in bounded controls copy this package into task-local scratch, weaken one
// production gate, and run the named behavioral test. A compiler/setup failure
// is NOT a kill; the named test must report Action=fail. Originals are untouched.
func TestNarrowingControls(t *testing.T) {
	if os.Getenv("CURATOR_SNAPSHOT_CONTROLS") != "1" {
		t.Skip("run explicitly with CURATOR_SNAPSHOT_CONTROLS=1")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	corpusPath, err := filepath.Abs(corpus)
	if err != nil {
		t.Fatal(err)
	}
	mutants := []struct{ name, file, old, new, test string }{
		{"M1-one-extra-LF", "codec.go", "payload := raw[:len(raw)-1]", "payload := raw[:len(raw)-1]; if bytes.HasSuffix(payload,[]byte{10}) { payload=payload[:len(payload)-1] }", "TestMetadataFramingAndHashDomains"},
		{"M2-one-unknown-member", "codec.go", "payload := raw[:len(raw)-1]", "payload := raw[:len(raw)-1]; payload=bytes.Replace(payload,[]byte(`\"_unknown\":true,`),nil,1)", "TestMetadataFramingAndHashDomains"},
		{"M3-launch-hash-without-LF", "metadata.go", "m.LaunchSHA256 != Digest(launch)", "m.LaunchSHA256 != Digest(launch) && m.LaunchSHA256 != Digest(launch[:len(launch)-1])", "TestLaunchHashIncludesLF"},
		{"M4-component-NFC-alias", "metadata.go", "old != prefix", "old != prefix && !strings.Contains(prefix,\"café\")", "TestMetadataNeighbors/NFD-components"},
		{"M5-mode", "verify.go", "!modeMatches(before, want.Mode)", "!(want.Mode == \"0700\" || modeMatches(before, want.Mode))", "TestActualTreeRefusals/execute-mode"},
		{"M5-extra", "verify.go", "if !ok {\n\t\t\treturn refuse(\"entry-set\")", "if !ok && p==\"home/extra\" { continue }; if !ok {\n\t\t\treturn refuse(\"entry-set\")", "TestActualTreeRefusals/extra-file"},
		{"M6-saved-manifest", "verify.go", "if err := verifyEntries(dir, \"\", dirs, files); err != nil {", "if expected != \"\" { return &Verified{bytes.Clone(manifest),bytes.Clone(launch)},nil }; if err := verifyEntries(dir, \"\", dirs, files); err != nil {", "TestBaselineNeverRecaptured"},
		{"M7-wrong-fixed-layer", "metadata.go", "d.MCP.Path != \"home/\"+c.Name+\".config.toml\"", "d.MCP.Path != \"home/\"+c.Name+\".config.toml\" && d.MCP.Path != \"home/AGENTS.md\"", "TestMetadataNeighbors/wrong-fixed-layer"},
		{"M8-portable-version", "verify.go", "am.FragmentSHA256 = \"\"", "am.FragmentSHA256 = \"\"; am.CuratorVersion=bm.CuratorVersion", "TestPortableIdentity/version"},
		{"M9-two-hardlinks", "fs_unix.go", "stat.Nlink == 1", "(stat.Nlink == 1 || stat.Nlink == 2)", "TestActualTreeRefusals/hardlink-outside"},
	}
	for _, m := range mutants {
		t.Run(m.name, func(t *testing.T) {
			scratch, err := os.MkdirTemp(filepath.Join(root, ".temp"), "snapshot-control-")
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := os.RemoveAll(scratch); err != nil {
					t.Error(err)
				}
			}()
			if err := filepath.WalkDir(".", func(p string, d fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if d.IsDir() {
					if p != "." && strings.HasPrefix(d.Name(), ".") {
						return filepath.SkipDir
					}
					return os.MkdirAll(filepath.Join(scratch, p), 0700)
				}
				if !strings.HasSuffix(p, ".go") && !strings.HasPrefix(p, "schemas/") {
					return nil
				}
				raw, err := os.ReadFile(p)
				if err != nil {
					return err
				}
				if p == "codec_test.go" {
					raw = bytes.ReplaceAll(raw, []byte(strconv.Quote(corpus)), []byte(strconv.Quote(corpusPath)))
				}
				if p == m.file {
					if bytes.Count(raw, []byte(m.old)) != 1 {
						t.Fatalf("control target count %d", bytes.Count(raw, []byte(m.old)))
					}
					raw = bytes.Replace(raw, []byte(m.old), []byte(m.new), 1)
					// The mode control keeps all other guards and permits only a 0700 asset
					// observed as non-executable at both pre- and post-read mode checks.
					if m.name == "M5-mode" {
						raw = bytes.Replace(raw, []byte("!modeMatches(after, want.Mode)"), []byte("!(want.Mode == \"0700\" || modeMatches(after, want.Mode))"), 1)
					}
				}
				return os.WriteFile(filepath.Join(scratch, p), raw, 0600)
			}); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("go", "test", "-count=1", "-json", "-run", "^"+m.test+"$", scratch)
			cmd.Dir = root
			// The child runs only the killer and cannot recurse into this control suite.
			output, runErr := cmd.CombinedOutput()
			killed := false
			for _, line := range bytes.Split(output, []byte{10}) {
				var event struct{ Action, Test string }
				if json.Unmarshal(line, &event) == nil && event.Action == "fail" && (event.Test == m.test || strings.HasPrefix(event.Test, m.test+"/")) {
					killed = true
				}
			}
			if runErr == nil || !killed {
				sanitized := strings.ReplaceAll(strings.ReplaceAll(string(output), scratch, "<control>"), root, "<repository>")
				t.Fatalf("survivor or invalid control: %v\n%s", runErr, sanitized)
			}
			t.Logf("%s: killed by %s; child exit 1", m.name, m.test)
		})
	}
}
