package config

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// A successful edit preserves unrelated effective values and source mode.
// Invalid levels, cancellation and a competing writer leave source bytes
// untouched; the competing writer's update is never replaced.
func TestStartupEdit(t *testing.T) {
	for _, mode := range []os.FileMode{0o600, 0o644} {
		path := authEditFixture(t, `{"schema_version":7,"skills_root":"fixture","projects":{},"preferred_locale":"en","extensions":{"proxy":{"profile":"fixture"}}}`)
		if err := os.Chmod(path, mode); err != nil {
			t.Fatal(err)
		}
		if err := SetStartupVerification(context.Background(), path, 2); err != nil {
			t.Fatal(err)
		}
		cfg, err := Load(path, nil)
		if err != nil || cfg.StartupVerification.Level != 2 || cfg.PreferredLocale != "en" || cfg.Extensions.Proxy.Profile != "fixture" {
			t.Fatalf("edited config: %+v,%v", cfg, err)
		}
		info, _ := os.Stat(path)
		if info.Mode().Perm() != mode {
			t.Fatal("edit changed config mode")
		}
		before, _ := os.ReadFile(path)
		if err := SetStartupVerification(context.Background(), path, 0); err == nil {
			t.Fatal("accepted invalid level")
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if err := SetStartupVerification(ctx, path, 4); err != context.Canceled {
			t.Fatalf("canceled edit: %v", err)
		}
		after, _ := os.ReadFile(path)
		if !bytes.Equal(before, after) {
			t.Fatal("failed edit changed source")
		}
		competitor := append(bytes.TrimSpace(before), '\n', '\n')
		err = setStartupVerification(context.Background(), path, 4, func() {
			if err := os.WriteFile(path, competitor, mode); err != nil {
				t.Fatal(err)
			}
		})
		if err == nil || !strings.Contains(err.Error(), "startup_verification_config_changed") {
			t.Fatalf("competing edit: %v", err)
		}
		after, _ = os.ReadFile(path)
		if !bytes.Equal(after, competitor) {
			t.Fatal("lost competing writer update")
		}
		staged, _ := filepath.Glob(filepath.Join(filepath.Dir(path), ".startup-verification-*.tmp"))
		if len(staged) != 0 {
			t.Fatal("failed edit retained staging files")
		}
	}
}

// Present symlink/hardlink sources and locked policy are refusals, with no
// target mutation. Nearby owned regular-file controls are exercised above.
func TestStartupEditRefusals(t *testing.T) {
	for _, kind := range []string{"symlink", "hardlink", "writable"} {
		path := authEditFixture(t, `{"schema_version":7,"skills_root":"fixture","projects":{}}`)
		target := path
		before, _ := os.ReadFile(target)
		switch kind {
		case "symlink":
			path = filepath.Join(filepath.Dir(path), "link.json")
			if err := os.Symlink(target, path); err != nil {
				t.Fatal(err)
			}
		case "hardlink":
			if err := os.Link(path, path+".link"); err != nil {
				t.Fatal(err)
			}
		case "writable":
			if err := os.Chmod(path, 0o666); err != nil {
				t.Fatal(err)
			}
		}
		if err := SetStartupVerification(context.Background(), path, 1); err == nil {
			t.Fatalf("accepted %s source", kind)
		}
		after, _ := os.ReadFile(target)
		if !bytes.Equal(before, after) {
			t.Fatal("refusal changed target")
		}
	}
}

// Byte-backed reads reject non-integer spellings before the legacy config
// representation loses that distinction; ordinary integer files still load.
func TestStartupJSONNumericSpelling(t *testing.T) {
	for _, level := range []string{"3", "3.0", "3e0"} {
		path := authEditFixture(t, `{"schema_version":11,"skills_root":"fixture","projects":{},"startup_verification":{"level":`+level+`}}`)
		_, err := Load(path, nil)
		if (err == nil) != (level == "3") {
			t.Fatalf("level %s error=%v", level, err)
		}
	}
	// A subsequent established config editor must retain the new typed value.
	object, err := claudeAuthObject([]byte(`{"startup_verification":{"level":3}}`))
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(object)
	if validateStartupJSON(raw) != nil {
		t.Fatal("existing editor cannot carry valid policy")
	}
}

// Editing schema1 retains its permissive posture and environment defaults;
// updating auth on carrier11 must preserve both the carrier and startup policy.
func TestStartupEditMigrationPreservesPolicy(t *testing.T) {
	path := authEditFixture(t, `{"schema_version":1,"skills_root":"fixture","projects":{}}`)
	before, err := Load(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := SetStartupVerification(context.Background(), path, 2); err != nil {
		t.Fatal(err)
	}
	after, err := Load(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if before.SecurityPosture != after.SecurityPosture || !reflect.DeepEqual(before.Env, after.Env) {
		t.Fatal("startup migration changed existing security/environment defaults")
	}
	plan, err := PlanClaudeAuth(path, authEditSelection())
	if err != nil {
		t.Fatal(err)
	}
	var candidate map[string]any
	if err := json.Unmarshal(plan.CandidateBytes(), &candidate); err != nil {
		t.Fatal(err)
	}
	if candidate["schema_version"] != float64(StartupVerificationSchemaVersion) {
		t.Fatal("auth edit downgraded startup carrier")
	}
	level, _ := candidate["startup_verification"].(map[string]any)
	if level["level"] != float64(2) {
		t.Fatal("auth edit lost startup policy")
	}
}
