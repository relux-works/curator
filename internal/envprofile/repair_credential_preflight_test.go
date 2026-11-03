package envprofile

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
)

// A native selector contaminated by the managed target must refuse before any
// public file, link or marker changes. The clean native selector must provision
// normally. All homes and credential-shaped inputs here are isolated fixtures.
func TestRepairManagedNativeHomeRefusesBeforePublicWrites(t *testing.T) {
	requireLinkCapability(t)
	for _, existing := range []bool{false, true} {
		name := "fresh"
		if existing {
			name = "existing"
		}
		t.Run(name, func(t *testing.T) {
			fx := writeManagedFixture(t, "sample")
			seedLiveNativeCredentials(t, fx)
			home := ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI)
			if existing {
				provision(t, fx, envregistry.CodexCLI, envregistry.DefaultMachineConfig())
				// A stale public context forces the actual repair path, without
				// changing the credential entry or its native target.
				path := filepath.Join(home, "AGENTS.md")
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte("user-preserved stale context\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			before := repairPublicFootprint(t, home)
			req := fx.request(envregistry.CodexCLI)
			req.Repair = true
			clean := req.NativeHomeOf
			req.NativeHomeOf = func(id string) (string, error) {
				if id == envregistry.CodexCLI {
					return home, nil
				}
				return clean(id)
			}
			_, err := Resolve(req)
			if err == nil || !strings.Contains(err.Error(), envregistry.DiagCredentialConflict) {
				t.Errorf("managed native selector must refuse with %s, got %v", envregistry.DiagCredentialConflict, err)
			}
			if after := repairPublicFootprint(t, home); !reflect.DeepEqual(before, after) {
				t.Errorf("credential refusal changed public home: before=%v after=%v", before, after)
			}
		})
	}
	t.Run("clean-control", func(t *testing.T) {
		fx := writeManagedFixture(t, "sample")
		seedLiveNativeCredentials(t, fx)
		req := fx.request(envregistry.CodexCLI)
		req.Repair = true
		if _, err := Resolve(req); err != nil {
			t.Fatalf("clean native selector must provision: %v", err)
		}
		readManagedMarker(t, fx, envregistry.CodexCLI)
	})
}

// The footprint records public bytes and link text without following links or
// reading credential payloads. Directories alone are not public publication.
func repairPublicFootprint(t *testing.T, home string) map[string]string {
	t.Helper()
	result := map[string]string{}
	err := filepath.WalkDir(home, func(path string, entry fs.DirEntry, err error) error {
		if os.IsNotExist(err) && path == home {
			return nil
		}
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(home, path)
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			result[rel] = "link:" + target
			return nil
		}
		if rel == "auth.json" || rel == ".credentials.json" {
			result[rel] = "credential-entry-present"
			return nil
		}
		payload, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		result[rel] = "file:" + string(payload)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

// Existing credential conflict forms keep the same refusal and leave public
// context/marker/skill state unchanged, including on first provisioning.
func TestRepairCredentialRefusalBeforePublicWrites(t *testing.T) {
	requireLinkCapability(t)
	for _, existing := range []bool{false, true} {
		for _, form := range []string{"regular-file", "foreign-link"} {
			name := "fresh/" + form
			if existing {
				name = "existing/" + form
			}
			t.Run(name, func(t *testing.T) {
				fx := writeManagedFixture(t, "sample")
				seedLiveNativeCredentials(t, fx)
				home := ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI)
				if existing {
					provision(t, fx, envregistry.CodexCLI, envregistry.DefaultMachineConfig())
					context := filepath.Join(home, "AGENTS.md")
					if err := os.Remove(context); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(context, []byte("preserved context\n"), 0o600); err != nil {
						t.Fatal(err)
					}
					if err := os.Remove(filepath.Join(home, "auth.json")); err != nil {
						t.Fatal(err)
					}
				} else if err := os.MkdirAll(home, 0o700); err != nil {
					t.Fatal(err)
				}
				auth := filepath.Join(home, "auth.json")
				if form == "regular-file" {
					if err := os.WriteFile(auth, []byte("synthetic fixture\n"), 0o600); err != nil {
						t.Fatal(err)
					}
				} else {
					foreign := filepath.Join(t.TempDir(), "auth.json")
					if err := os.WriteFile(foreign, []byte("synthetic foreign fixture\n"), 0o600); err != nil {
						t.Fatal(err)
					}
					if err := os.Symlink(foreign, auth); err != nil {
						t.Fatal(err)
					}
				}
				before := repairPublicFootprint(t, home)
				req := fx.request(envregistry.CodexCLI)
				req.Repair = true
				_, err := Resolve(req)
				if err == nil || !strings.Contains(err.Error(), envregistry.DiagCredentialConflict) {
					t.Errorf("wrong credential refusal: %v", err)
				}
				if after := repairPublicFootprint(t, home); !reflect.DeepEqual(before, after) {
					t.Errorf("credential refusal changed public home: before=%v after=%v", before, after)
				}
			})
		}
	}
}

// Physical inside aliases refuse; a real external alias, missing ordinary
// external home and sibling sharing the root's text prefix remain legitimate.
func TestRepairNativeHomeLocationCompatibility(t *testing.T) {
	requireLinkCapability(t)
	for _, form := range []string{"inside-alias", "broken-inside-alias", "outside-alias", "outside-missing", "prefix-neighbor"} {
		t.Run(form, func(t *testing.T) {
			fx := writeManagedFixture(t, "sample")
			seedLiveNativeCredentials(t, fx)
			home := ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI)
			native := fx.native[envregistry.CodexCLI]
			refuse := strings.Contains(form, "inside")
			switch form {
			case "inside-alias", "broken-inside-alias":
				if form == "inside-alias" {
					if err := os.MkdirAll(home, 0o700); err != nil {
						t.Fatal(err)
					}
				}
				native = filepath.Join(t.TempDir(), "native")
				if err := os.Symlink(home, native); err != nil {
					t.Fatal(err)
				}
			case "outside-alias":
				native = filepath.Join(t.TempDir(), "native")
				if err := os.Symlink(fx.native[envregistry.CodexCLI], native); err != nil {
					t.Fatal(err)
				}
			case "outside-missing":
				native = filepath.Join(t.TempDir(), "missing", "native")
			case "prefix-neighbor":
				native = EnvRoot(fx.home) + "-native"
				if err := os.MkdirAll(native, 0o700); err != nil {
					t.Fatal(err)
				}
			}
			before := repairPublicFootprint(t, home)
			req := fx.request(envregistry.CodexCLI)
			req.Repair = true
			original := req.NativeHomeOf
			req.NativeHomeOf = func(id string) (string, error) {
				if id == envregistry.CodexCLI {
					return native, nil
				}
				return original(id)
			}
			_, err := Resolve(req)
			if refuse {
				if err == nil || !strings.Contains(err.Error(), envregistry.DiagCredentialConflict) {
					t.Errorf("wrong native location refusal: %v", err)
				}
				if after := repairPublicFootprint(t, home); !reflect.DeepEqual(before, after) {
					t.Errorf("native location refusal changed public home: before=%v after=%v", before, after)
				}
			} else {
				if err != nil {
					t.Fatalf("legitimate native location refused: %v", err)
				}
				readManagedMarker(t, fx, envregistry.CodexCLI)
			}
		})
	}
}
