package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestEnvConfigShowPermissionsAtProductionEntry(t *testing.T) {
	for _, locked := range []bool{false, true} {
		name := "profile"
		if locked {
			name = "global"
		}
		t.Run(name, func(t *testing.T) {
			source, _ := profileHome(t)
			machine := `{"schema_version":2,"security_posture":"permissive","skills_root":"x","projects":{},"environments":{"permissions":{"acme":"yolo","uninstalled":"yolo"}}}`
			if err := os.WriteFile(source.path, []byte(machine), 0o600); err != nil {
				t.Fatal(err)
			}
			system := filepath.Join(t.TempDir(), "system.json")
			contents := `{"schema_version":2}`
			if locked {
				contents = `{"schema_version":2,"locked":["environments.permissions"],"environments":{"permissions":{}}}`
			}
			if err := os.WriteFile(system, []byte(contents), 0o600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("CURATOR_SYSTEM_CONFIG", system)
			for _, profile := range []string{"acme", "other"} {
				pkg := t.TempDir()
				writeContextPackage(t, pkg, profile, "1.0.0", "hello\n")
				if code, _, stderr := runFileConfigProfile(t, source.path, "profile", "install", pkg); code != exitOK {
					t.Fatalf("install = %d: %s", code, stderr)
				}
			}
			code, stdout, stderr := runFileConfigProfile(t, source.path, "env", "config", "show")
			if code != exitOK {
				t.Fatalf("show = %d: %s", code, stderr)
			}
			var full map[string]json.RawMessage
			if err := json.Unmarshal([]byte(stdout), &full); err != nil {
				t.Fatal(err)
			}
			var postures map[string]map[string]any
			if err := json.Unmarshal(full["permissions_posture"], &postures); err != nil {
				t.Fatal(err)
			}
			if len(postures) != 3 || postures["default"] == nil || postures["acme"] == nil || postures["other"] == nil || postures["uninstalled"] != nil {
				t.Fatalf("installed profiles = %v", postures)
			}
			for profile, got := range postures {
				mode, source := "native", "default"
				if profile == "acme" {
					mode, source = "yolo", "profile"
				}
				if locked {
					mode, source = "native", "global"
				}
				want := map[string]any{"mode": mode, "locked": locked, "source": source}
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("%s posture = %v, want %v", profile, got, want)
				}
			}
			// The existing permissions map keeps its original JSON shape.
			var raw map[string]string
			if err := json.Unmarshal(full["permissions"], &raw); err != nil {
				t.Fatal(err)
			}
			if !locked && raw["acme"] != "yolo" {
				t.Fatalf("raw permissions changed: %v", raw)
			}
			code, stdout, stderr = runFileConfigProfile(t, source.path, "env", "config", "show", "permissions")
			if code != exitOK {
				t.Fatalf("single show = %d: %s", code, stderr)
			}
			var single map[string]any
			if err := json.Unmarshal([]byte(stdout), &single); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(single, postures["acme"]) {
				t.Fatalf("current posture = %v, want %v", single, postures["acme"])
			}
		})
	}
}

func TestEnvConfigShowPermissionsRejectsUnreadableConfig(t *testing.T) {
	source, _ := profileHome(t)
	if err := os.WriteFile(source.path, []byte(`{"schema_version":2,"security_posture":"permissive","skills_root":"x","projects":{}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURATOR_SYSTEM_CONFIG", "")
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runFileConfigProfile(t, source.path, "profile", "install", pkg); code != exitOK {
		t.Fatalf("install = %d: %s", code, stderr)
	}
	for _, broken := range []string{"machine", "system"} {
		t.Run(broken, func(t *testing.T) {
			machine := source.path
			system := filepath.Join(t.TempDir(), "system.json")
			if err := os.WriteFile(system, []byte(`{"schema_version":2}`), 0o600); err != nil {
				t.Fatal(err)
			}
			if broken == "system" {
				system = t.TempDir()
			} else {
				machine = t.TempDir()
			}
			t.Setenv("CURATOR_SYSTEM_CONFIG", system)
			for _, args := range [][]string{{"env", "config", "show"}, {"env", "config", "show", "permissions"}} {
				code, stdout, stderr := runFileConfigProfile(t, machine, args...)
				if code != exitFail || stdout != "" || stderr == "" {
					t.Fatalf("unreadable config = %d, stdout=%q stderr=%q", code, stdout, stderr)
				}
			}
		})
	}
}

func TestEnvConfigShowPermissionsRejectsUnreadableInventoryAndCurrent(t *testing.T) {
	source, home := profileHome(t)
	// A regular file where a directory is required gives a portable read failure.
	if err := os.WriteFile(filepath.Join(home, "profiles"), []byte("unreadable inventory"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"env", "config", "show"}, {"env", "config", "show", "permissions"}} {
		code, stdout, stderr := runProfile(t, source, args...)
		if code != exitFail || stdout != "" || !strings.Contains(stderr, "permissions posture unreadable") {
			t.Fatalf("unreadable posture = %d, stdout=%q stderr=%q", code, stdout, stderr)
		}
	}
}

func TestEnvConfigShowPermissionsRejectsUnreadableInventoryAncestor(t *testing.T) {
	source, home := profileHome(t)
	blocker := filepath.Join(home, "blocker")
	if err := os.WriteFile(blocker, []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	// Keep config loading injected while driving real state reads beneath a
	// non-directory ancestor, including the inventory metadata inspection.
	source.cfg.Path = filepath.Join(blocker, "config.json")
	for _, args := range [][]string{{"env", "config", "show"}, {"env", "config", "show", "permissions"}} {
		code, stdout, stderr := runProfile(t, source, args...)
		if code != exitFail || stdout != "" || !strings.Contains(stderr, "permissions posture unreadable") {
			t.Fatalf("unreadable ancestor = %d, stdout=%q stderr=%q", code, stdout, stderr)
		}
	}
}

func TestEnvConfigShowPermissionsWithoutProfiles(t *testing.T) {
	source, home := profileHome(t)
	code, stdout, stderr := runProfile(t, source, "env", "config", "show")
	if code != exitOK {
		t.Fatalf("empty show = %d: %s", code, stderr)
	}
	var full map[string]json.RawMessage
	if err := json.Unmarshal([]byte(stdout), &full); err != nil {
		t.Fatal(err)
	}
	if string(full["permissions_posture"]) != "{}" {
		t.Fatalf("empty posture = %s", full["permissions_posture"])
	}
	if _, err := os.Stat(filepath.Join(home, "profiles")); !os.IsNotExist(err) {
		t.Fatalf("show changed profile state: %v", err)
	}
	code, stdout, stderr = runProfile(t, source, "env", "config", "show", "permissions")
	if code != exitFail || stdout != "" || !strings.Contains(stderr, "no current profile") {
		t.Fatalf("missing current = %d stdout=%q stderr=%q", code, stdout, stderr)
	}
}
