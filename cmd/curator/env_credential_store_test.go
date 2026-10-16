package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func credentialStoreCLIHome(t *testing.T) stubConfigSource {
	t.Helper()
	source, _ := profileHome(t)
	payload, err := json.Marshal(map[string]any{"schema_version": 2, "security_posture": "permissive", "skills_root": filepath.Join(source.cfg.Home(), "skills"), "projects": map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source.path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	return reloadSource(t, source)
}

// CLI aliases canonicalize only environment positions; setting/unsetting changes no home state.
func TestCredentialStoreCLIConfigRoundTrip(t *testing.T) {
	source := credentialStoreCLIHome(t)
	installCLIEnvProfile(t, source)
	if code, _, stderr := runProfile(t, source, "env", "resolve", "codex", "--repair"); code != exitOK {
		t.Fatal(stderr)
	}
	markerPath, marker, _ := cliEnvMarker(t, source, "codex_cli")
	seedPath := filepath.Join(filepath.Dir(markerPath), "config.toml")
	// Write a vendor-owned seed and a detached credential after provisioning.
	seed := []byte("model = \"test\"\n")
	if err := os.WriteFile(seedPath, seed, 0o600); err != nil {
		t.Fatal(err)
	}
	authPath := filepath.Join(filepath.Dir(markerPath), "auth.json")
	if err := os.Remove(authPath); err != nil {
		t.Fatal(err)
	}
	auth := []byte("synthetic-private-credential")
	if err := os.WriteFile(authPath, auth, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, knob := range []string{"credential_store.codex.codex", "credential_store.codex.codex_cli"} {
		if code, stdout, stderr := runProfile(t, source, "env", "config", "set", knob, "file"); code != exitOK || strings.TrimSpace(stdout) != `"file"` {
			t.Fatalf("set %s = %d %s %s", knob, code, stdout, stderr)
		}
		source = reloadSource(t, source)
		if source.cfg.Schema != 5 {
			t.Fatalf("schema = %d", source.cfg.Schema)
		}
		for _, query := range []string{"credential_store.codex.codex", "credential_store.codex.codex_cli"} {
			if code, stdout, stderr := runProfile(t, source, "env", "config", "show", query); code != exitOK || strings.TrimSpace(stdout) != `"file"` {
				t.Fatalf("show = %d %s %s", code, stdout, stderr)
			}
		}
	}
	raw, err := os.ReadFile(source.path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(raw, []byte(`"codex_cli": "file"`)) || bytes.Contains(raw, []byte(`"codex_cli": {`)) || !bytes.Contains(raw, []byte(`"codex": {`)) {
		t.Fatalf("wrong profile/environment normalization: %s", raw)
	}
	if code, _, stderr := runProfile(t, source, "env", "config", "unset", "credential_store.codex.codex"); code != exitOK {
		t.Fatal(stderr)
	}
	source = reloadSource(t, source)
	if code, stdout, stderr := runProfile(t, source, "env", "config", "show", "credential_store.codex.codex"); code != exitOK || strings.TrimSpace(stdout) != `"native"` {
		t.Fatalf("default = %d %s %s", code, stdout, stderr)
	}
	for path, want := range map[string][]byte{markerPath: marker, seedPath: seed, authPath: auth} {
		got, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(got, want) {
			t.Fatalf("config edit changed vendor/managed state %s: %v", filepath.Base(path), err)
		}
	}
}

// Invalid selectors are rejected atomically beside valid canonical and alias-map controls.
func TestCredentialStoreCLIConfigRefusesMalformed(t *testing.T) {
	source := writeMachineConfig(t, `{"schema_version":5,"skills_root":"x","projects":{}}`)
	for _, tc := range []struct{ knob, value string }{
		{"credential_store", `null`},
		{"credential_store.p", `[]`},
		{"credential_store.p.codex", `auto`},
		{"credential_store.p.codex", `null`},
		{"credential_store.p.unknown", `file`},
	} {
		before, _ := os.ReadFile(source.path)
		code, stdout, stderr := runProfile(t, source, "env", "config", "set", tc.knob, tc.value)
		if code != exitFail || stdout != "" || !strings.Contains(stderr, "credential_store") {
			t.Fatalf("invalid set = %d %s %s", code, stdout, stderr)
		}
		after, _ := os.ReadFile(source.path)
		if !bytes.Equal(before, after) {
			t.Fatal("refused set changed config")
		}
	}
	if code, stdout, _ := runProfile(t, source, "env", "config", "show", "credential_store.p.unknown"); code != exitUsage || stdout != "" {
		t.Fatal("unknown environment show acquired a native default")
	}
	if code, _, stderr := runProfile(t, source, "env", "config", "set", "credential_store", `{"codex":{"codex":"file","codex_cli":"native","claude":"keychain"}}`); code != exitOK {
		t.Fatal(stderr)
	}
	source = reloadSource(t, source)
	if code, stdout, stderr := runProfile(t, source, "env", "config", "show", "credential_store.codex"); code != exitOK || strings.Contains(stdout, `"codex":`) || !strings.Contains(stdout, `"codex_cli": "native"`) || !strings.Contains(stdout, `"claude_code": "keychain"`) {
		t.Fatalf("collision/profile = %d %s %s", code, stdout, stderr)
	}
}

// Promoting a schema-1 file for this knob preserves its permissive posture;
// a failed edit does not promote it.
func TestCredentialStoreCLIPromotion(t *testing.T) {
	source := writeMachineConfig(t, `{"schema_version":1,"skills_root":"x","projects":{}}`)
	before, _ := os.ReadFile(source.path)
	if code, _, _ := runProfile(t, source, "env", "config", "set", "credential_store.p.codex", "auto"); code != exitFail {
		t.Fatal("invalid selector accepted")
	}
	after, _ := os.ReadFile(source.path)
	if !bytes.Equal(before, after) {
		t.Fatal("failed edit promoted the envelope")
	}
	if code, _, stderr := runProfile(t, source, "env", "config", "set", "credential_store.p.codex", "native"); code != exitOK {
		t.Fatal(stderr)
	}
	source = reloadSource(t, source)
	if source.cfg.Schema != 5 || source.cfg.SecurityPosture != "permissive" {
		t.Fatal("promotion changed unrelated posture")
	}
}

// The production CLI composition refuses unavailable explicit-store delivery without a fragment or writes.
// Native selection remains a positive control. This does not qualify any vendor runtime.
func TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery(t *testing.T) {
	source := credentialStoreCLIHome(t)
	installCLIEnvProfile(t, source)
	if code, _, stderr := runProfile(t, source, "env", "resolve", "codex", "--repair"); code != exitOK {
		t.Fatal(stderr)
	}
	markerPath, marker, _ := cliEnvMarker(t, source, "codex_cli")
	authPath := filepath.Join(filepath.Dir(markerPath), "auth.json")
	link, err := os.Readlink(authPath)
	if err != nil {
		t.Fatal(err)
	}
	native, err := os.ReadFile(link)
	if err != nil {
		t.Fatal(err)
	}
	for _, selection := range []string{"file", "keychain"} {
		if code, _, stderr := runProfile(t, source, "env", "config", "set", "credential_store.acme.codex", selection); code != exitOK {
			t.Fatal(stderr)
		}
		source = reloadSource(t, source)
		configBefore, _ := os.ReadFile(source.path)
		for _, repair := range [][]string{nil, {"--repair"}} {
			args := append([]string{"env", "resolve", "codex"}, repair...)
			code, stdout, stderr := runProfile(t, source, args...)
			want := "codex_settings_unsupported"
			if selection == "keychain" {
				want = "environment_credential_store_unsupported"
			}
			if code != exitFail || stdout != "" || !strings.Contains(stderr, want) {
				t.Fatalf("unqualified %s = %d %s %s", selection, code, stdout, stderr)
			}
		}
		for path, want := range map[string][]byte{markerPath: marker, link: native, source.path: configBefore} {
			got, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(got, want) {
				t.Fatalf("refusal changed %s: %v", filepath.Base(path), err)
			}
		}
		if got, err := os.Readlink(authPath); err != nil || got != link {
			t.Fatal("refusal changed native credential link")
		}
	}
	if code, _, stderr := runProfile(t, source, "env", "config", "unset", "credential_store.acme.codex"); code != exitOK {
		t.Fatal(stderr)
	}
	source = reloadSource(t, source)
	if code, stdout, stderr := runProfile(t, source, "env", "resolve", "codex"); code != exitOK || !strings.Contains(stdout, "launch-env-fragment") {
		t.Fatalf("native control = %d %s %s", code, stdout, stderr)
	}
}

// A valid dotted profile must remain one profile in the public selector CLI.
func TestReviewDottedCredentialProfile(t *testing.T) {
	for _, profile := range []string{"team", "team.prod"} {
		t.Run(profile, func(t *testing.T) {
			source := writeMachineConfig(t, `{"schema_version":5,"skills_root":"x","projects":{}}`)
			code, stdout, stderr := runProfile(t, source, "env", "config", "set", "credential_store."+profile+".codex", "file")
			if code != exitOK {
				t.Fatalf("valid profile refused: code=%d stdout=%s stderr=%s", code, stdout, stderr)
			}
			source = reloadSource(t, source)
			code, stdout, stderr = runProfile(t, source, "env", "config", "show", "credential_store."+profile+".codex")
			if code != exitOK || strings.TrimSpace(stdout) != `"file"` {
				t.Fatalf("show=%d %s %s", code, stdout, stderr)
			}
			if code, _, stderr := runProfile(t, source, "env", "config", "unset", "credential_store."+profile+".codex"); code != exitOK {
				t.Fatal(stderr)
			}
		})
	}
}

// Read/unset must address a previously persisted valid dotted profile.
func TestReviewDottedCredentialProfileExisting(t *testing.T) {
	for _, op := range []string{"show", "unset"} {
		t.Run(op, func(t *testing.T) {
			source := writeMachineConfig(t, `{"schema_version":5,"skills_root":"x","projects":{},"environments":{"credential_store":{"team.prod":{"codex_cli":"file"}}}}`)
			code, stdout, stderr := runProfile(t, source, "env", "config", op, "credential_store.team.prod.codex_cli")
			if code != exitOK || (op == "show" && strings.TrimSpace(stdout) != `"file"`) {
				t.Fatalf("existing dotted profile %s: code=%d stdout=%s stderr=%s", op, code, stdout, stderr)
			}
		})
	}
}

// Full selector paths preserve profile punctuation and case for both aliases and
// canonical IDs. Failed writes and missing unsets leave the existing map intact.
func TestCredentialStoreCLIDottedProfileRouting(t *testing.T) {
	for _, profile := range []string{"team", "team.prod", "Team..prod", "codex.prod"} {
		for _, environment := range []string{"codex", "codex_cli", "claude", "claude_code"} {
			t.Run(profile+"/"+environment, func(t *testing.T) {
				source := writeMachineConfig(t, `{"schema_version":5,"skills_root":"x","projects":{},"environments":{"credential_store":{"neighbor":{"codex_cli":"keychain"}}}}`)
				knob := "credential_store." + profile + "." + environment
				canonical := "codex_cli"
				if strings.HasPrefix(environment, "claude") {
					canonical = "claude_code"
				}
				if code, stdout, stderr := runProfile(t, source, "env", "config", "set", knob, "file"); code != exitOK || strings.TrimSpace(stdout) != `"file"` {
					t.Fatalf("set = %d %s %s", code, stdout, stderr)
				}
				source = reloadSource(t, source)
				if len(source.cfg.Env.CredentialStore) != 2 || source.cfg.Env.CredentialStore[profile][canonical] != "file" || len(source.cfg.Env.CredentialStore[profile]) != 1 {
					t.Fatalf("profile or environment identity changed: %v", source.cfg.Env.CredentialStore)
				}
				before, err := os.ReadFile(source.path)
				if err != nil {
					t.Fatal(err)
				}
				for _, args := range [][]string{
					{"set", knob, "auto"},
					{"set", "credential_store." + profile + ".unknown", "file"},
					{"unset", "credential_store." + profile + ".muse"},
				} {
					if code, stdout, stderr := runProfile(t, source, append([]string{"env", "config"}, args...)...); code != exitFail || stdout != "" || !strings.Contains(stderr, "credential_store") {
						t.Fatalf("refusal = %d %s %s", code, stdout, stderr)
					}
					after, err := os.ReadFile(source.path)
					if err != nil || !bytes.Equal(before, after) {
						t.Fatal("refused edit changed selector map")
					}
				}
				for _, query := range []string{knob, "credential_store." + profile + "." + canonical} {
					if code, stdout, stderr := runProfile(t, source, "env", "config", "show", query); code != exitOK || strings.TrimSpace(stdout) != `"file"` {
						t.Fatalf("show = %d %s %s", code, stdout, stderr)
					}
				}
				if code, _, stderr := runProfile(t, source, "env", "config", "unset", knob); code != exitOK {
					t.Fatal(stderr)
				}
				source = reloadSource(t, source)
				if code, stdout, stderr := runProfile(t, source, "env", "config", "show", knob); code != exitOK || strings.TrimSpace(stdout) != `"native"` {
					t.Fatalf("default = %d %s %s", code, stdout, stderr)
				}
				if source.cfg.Env.CredentialStore["neighbor"]["codex_cli"] != "keychain" || len(source.cfg.Env.CredentialStore[profile]) != 0 {
					t.Fatal("unset changed another selector")
				}
			})
		}
	}
}

// Invalid profile operands refuse at path parsing for every command, with no
// write; valid internal dots are covered by the neighboring routing controls.
func TestCredentialStoreCLIInvalidProfileRouting(t *testing.T) {
	source := writeMachineConfig(t, `{"schema_version":5,"skills_root":"x","projects":{},"environments":{"credential_store":{"team.prod":{"codex_cli":"file"}}}}`)
	before, err := os.ReadFile(source.path)
	if err != nil {
		t.Fatal(err)
	}
	for _, profile := range []string{".team", "team.", "team/prod", "CON.prod", strings.Repeat("p", 129)} {
		for _, op := range []string{"set", "show", "unset"} {
			t.Run(profile+"/"+op, func(t *testing.T) {
				args := []string{"env", "config", op, "credential_store." + profile + ".codex"}
				if op == "set" {
					args = append(args, "native")
				}
				if code, stdout, stderr := runProfile(t, source, args...); code != exitUsage || stdout != "" || !strings.Contains(stderr, "credential_store profile") {
					t.Fatalf("invalid profile = %d %s %s", code, stdout, stderr)
				}
				after, err := os.ReadFile(source.path)
				if err != nil || !bytes.Equal(before, after) {
					t.Fatal("invalid path changed existing policy")
				}
			})
		}
	}
}
