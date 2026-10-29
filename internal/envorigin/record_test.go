package envorigin

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixtureRecord(t *testing.T) (*Record, Binding, string) {
	t.Helper()
	root := t.TempDir()
	manager := filepath.Join(root, "manager")
	target := filepath.Join(manager, "environments", "p", "codex_cli")
	b := Binding{Profile: "p", LockSHA256: "sha256:" + strings.Repeat("a", 64), ManagedTarget: target}
	r := &Record{Record: Version, Adapter: "codex_cli", OperatorHome: root, Selector: Selector{Name: "CODEX_HOME", State: "absent"}, NativeRoot: filepath.Join(root, "native"), Capture: Capture{Method: "provisioning", Authority: "native-selection", Time: "2026-10-27T00:00:00Z"}, Binding: b}
	return r, b, manager
}

// Proves the closed carrier/presence/binding contract; not filesystem or account transport.
func TestProvisioningOriginBinding(t *testing.T) {
	r, b, m := fixtureRecord(t)
	raw, err := r.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Parse(raw, b, "codex_cli", m); err != nil {
		t.Fatal(err)
	}
	cases := map[string][]byte{
		"caseFoldRoot":     bytes.Replace(raw, []byte(`"record":`), []byte(`"Record":`), 1),
		"caseFoldCapture":  bytes.Replace(raw, []byte(`"authority":`), []byte(`"Authority":`), 1),
		"caseFoldBinding":  bytes.Replace(raw, []byte(`"profile":`), []byte(`"Profile":`), 1),
		"caseFoldSelector": bytes.Replace(raw, []byte(`"name":`), []byte(`"Name":`), 1),
		"duplicate":        bytes.Replace(raw, []byte(`"record":`), []byte(`"record":"provisioning-origin/1","record":`), 1),
		"unknown":          append([]byte(`{"unknown":true,`), raw[1:]...),
		"nullSelector":     bytes.Replace(raw, []byte(`"state":"absent"`), []byte(`"state":"absent","value":null`), 1),
		"unknownVersion":   bytes.Replace(raw, []byte(Version), []byte("provisioning-origin/2"), 1),
		"wrongProvenance":  bytes.Replace(raw, []byte(`"native-selection"`), []byte(`"explicit-operator-input"`), 1),
		"nonUTC":           bytes.Replace(raw, []byte(`2026-10-27T00:00:00Z`), []byte(`2026-10-27T00:00:00+00:00`), 1),
		"wrongLock":        bytes.Replace(raw, []byte(strings.Repeat("a", 64)), []byte(strings.Repeat("b", 64)), 1),
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse(payload, b, "codex_cli", m); err == nil || !strings.Contains(err.Error(), DiagInvalid) {
				t.Fatalf("want origin-invalid, got %v", err)
			}
		})
	}
	for _, state := range []string{"present-empty", "present-value"} {
		v := ""
		if state == "present-value" {
			v = filepath.Join(t.TempDir(), "native")
		}
		r.Selector = Selector{Name: "CODEX_HOME", State: state, Value: &v}
		raw, err = r.Marshal()
		if err != nil {
			t.Fatal(err)
		}
		if _, err = Parse(raw, b, "codex_cli", m); err != nil {
			t.Fatal(err)
		}
	}
}

// Proves an absent authority below a substituted parent is invalid, not legacy.
func TestProvisioningOriginAbsentRecordParent(t *testing.T) {
	_, b, manager := fixtureRecord(t)
	outside := t.TempDir()
	if err := os.MkdirAll(filepath.Dir(b.ManagedTarget), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, b.ManagedTarget); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Read(manager, b, "codex_cli"); err == nil || !strings.Contains(err.Error(), DiagInvalid) {
		t.Fatalf("unsafe absent authority must be invalid, got %v", err)
	}
}

// A missing lexical tail is legacy absence. Broken/internal links and regular
// file parents must remain invalid even if an underlying lookup says ENOENT.
// The readable private-file neighbor exercises the actual authority transport.
func TestProvisioningOriginAuthorityParentErrors(t *testing.T) {
	for _, mode := range []string{"absent-tail", "broken-link", "internal-link", "file-parent", "present"} {
		t.Run(mode, func(t *testing.T) {
			manager := t.TempDir()
			parent := filepath.Join(manager, "carrier")
			switch mode {
			case "broken-link", "internal-link":
				target := filepath.Join(manager, "destination")
				if mode == "internal-link" {
					if err := os.Mkdir(target, 0700); err != nil {
						t.Fatal(err)
					}
				}
				if err := os.Symlink(target, parent); err != nil {
					t.Fatal(err)
				}
			case "file-parent":
				if err := os.WriteFile(parent, []byte("wrong kind"), 0600); err != nil {
					t.Fatal(err)
				}
			case "present":
				if err := os.Mkdir(parent, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(parent, "authority.json"), []byte("fixture"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			raw, present, err := ReadPrivateAuthority(manager, filepath.Join(parent, "authority.json"), 1024)
			if mode == "absent-tail" {
				if err != nil || present || raw != nil {
					t.Fatalf("real absence: %q %t %v", raw, present, err)
				}
			} else if mode == "present" {
				if err != nil || !present || string(raw) != "fixture" {
					t.Fatalf("private positive: %q %t %v", raw, present, err)
				}
			} else if err == nil || !strings.Contains(err.Error(), DiagInvalid) || present || raw != nil {
				t.Fatalf("unsafe parent became absence: %q %t %v", raw, present, err)
			}
		})
	}
}

// Proves rebind changes only binding bytes, retaining original capture JSON byte slices.
func TestProvisioningOriginRebindPreservesCapture(t *testing.T) {
	r, b, m := fixtureRecord(t)
	raw, err := r.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	raw = bytes.Replace(raw, []byte(`"capture":{`), []byte("\"capture\": { \n"), 1)
	next := b
	next.LockSHA256 = "sha256:" + strings.Repeat("b", 64)
	rebound, err := Rebind(raw, b, next, "codex_cli", m)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := CaptureFacts(raw)
	after, _ := CaptureFacts(rebound)
	if !bytes.Equal(before, after) {
		t.Fatal("capture facts changed on rebind")
	}
	if _, err = Parse(rebound, next, "codex_cli", m); err != nil {
		t.Fatal(err)
	}
}

// Proves authority is a private no-follow single-link file; credential contents are never fixtures here.
func TestProvisioningOriginPrivateRecord(t *testing.T) {
	r, b, m := fixtureRecord(t)
	if err := os.MkdirAll(b.ManagedTarget, 0700); err != nil {
		t.Fatal(err)
	}
	raw, _ := r.Marshal()
	path := filepath.Join(b.ManagedTarget, Name)
	if _, _, err := Read(m, b, "codex_cli"); err == nil || !strings.Contains(err.Error(), DiagRequired) {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Read(m, b, "codex_cli"); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Read(m, b, "codex_cli"); err == nil {
		t.Fatal("public authority accepted")
	}
	_ = os.Chmod(path, 0600)
	other := path + ".other"
	if err := os.Link(path, other); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Read(m, b, "codex_cli"); err == nil {
		t.Fatal("hardlink authority accepted")
	}
	_ = os.Remove(other)
	if err := os.Rename(path, other); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(other, path); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Read(m, b, "codex_cli"); err == nil {
		t.Fatal("symlink authority accepted")
	}
}
