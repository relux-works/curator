package config

import (
	"encoding/json"
	"strings"
	"testing"
)

// These tests drive the production entry points Parse and Load for the
// manager-config schema 2 and system-config schema 2 surfaces (environments
// §12.1, §12.2; manager §1, §12).

func loadText(t *testing.T, text string) *Config {
	t.Helper()
	path := writeConfig(t, t.TempDir(), "config.json", text)
	cfg, err := Load(path, nil)
	if err != nil {
		t.Fatalf("Load rejected: %v", err)
	}
	return cfg
}

func loadFails(t *testing.T, text, want string) {
	t.Helper()
	path := writeConfig(t, t.TempDir(), "config.json", text)
	_, err := Load(path, nil)
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("err = %v, want mention of %q", err, want)
	}
}

func loadWithSystem(t *testing.T, user, system string) (*Config, []string, error) {
	t.Helper()
	dir := t.TempDir()
	userPath := writeConfig(t, dir, "config.json", user)
	systemPath := writeConfig(t, dir, "system.json", system)
	t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)
	var warnings []string
	cfg, err := Load(userPath, func(message string) { warnings = append(warnings, message) })
	return cfg, warnings, err
}

func TestSchema2Defaults(t *testing.T) {
	cfg := loadText(t, `{"schema_version": 2, "skills_root": "/tmp/skills", "projects": {}}`)
	if cfg.Schema != 2 {
		t.Fatalf("Schema = %d, want 2", cfg.Schema)
	}
	env := cfg.Env
	if env.CurrentProfile != nil || env.RequireCurrent != nil {
		t.Fatalf("nullable knobs: %+v", env)
	}
	if env.OverlayDefaultWeight != 1000 || !env.OverlaysAllowed || env.BackupRetention != 5 {
		t.Fatalf("scalar defaults: %+v", env)
	}
	if env.Precedence != (Precedence{Winner: "higher-weight", Placement: "winner-last"}) {
		t.Fatalf("precedence default: %+v", env.Precedence)
	}
	if len(env.XDGSeedAllowlist) != 3 || env.XDGSeedAllowlist[0] != "git" {
		t.Fatalf("xdg default: %v", env.XDGSeedAllowlist)
	}
	if env.PassableEnvNames != nil {
		t.Fatalf("passable default must be null, got %v", env.PassableEnvNames)
	}
	for name, value := range map[string]int{
		"scoped": len(env.ScopedCurrent), "overlays": len(env.Overlays),
		"forms": len(env.Forms), "spf": len(env.SystemPromptFiles),
		"targets": len(env.Targets), "isolation": len(env.Isolation),
		"mcp": len(env.MCPPackageAllowlist), "shadow": len(env.ShadowAcknowledged),
		"waivers": len(env.SecretWaivers), "inplace": len(env.InPlaceMode),
	} {
		if value != 0 {
			t.Fatalf("%s default must be empty, got %d", name, value)
		}
	}
}

func TestSchema1StaysValid(t *testing.T) {
	cfg := loadText(t, minimal)
	if cfg.Schema != 1 {
		t.Fatalf("Schema = %d, want 1", cfg.Schema)
	}
	if cfg.Env.OverlayDefaultWeight != 1000 || !cfg.Env.OverlaysAllowed {
		t.Fatalf("schema-1 file must carry every default: %+v", cfg.Env)
	}
}

func TestSchema1WithEnvironmentsRejected(t *testing.T) {
	loadFails(t, `{"schema_version": 1, "skills_root": "x", "projects": {}, "environments": {}}`, "environments")
}

func TestSchema2NullEnvironmentsRejected(t *testing.T) {
	loadFails(t, `{"schema_version": 2, "skills_root": "x", "projects": {}, "environments": null}`, "environments")
}

func TestUnknownSchemaVersionsRejected(t *testing.T) {
	for _, version := range []string{"0", "3", "99"} {
		loadFails(t, `{"schema_version": `+version+`, "skills_root": "x", "projects": {}}`, "schema_version")
	}
}

func TestSchema2KnobRejections(t *testing.T) {
	base := `{"schema_version": 2, "skills_root": "x", "projects": {}, "environments": %s}`
	cases := []struct {
		name string
		env  string
		want string
	}{
		{"unknown knob", `{"nope": true}`, "environments"},
		{"current profile grammar", `{"current_profile": "-bad"}`, "current_profile"},
		{"current profile type", `{"current_profile": 3}`, "current_profile"},
		{"scoped value grammar", `{"scoped_current": {"pi": "-bad"}}`, "scoped_current"},
		{"scoped key grammar", `{"scoped_current": {"-bad": "personal"}}`, "scoped_current"},
		{"overlay profile grammar", `{"overlays": {"-bad": []}}`, "overlays"},
		{"overlay not list", `{"overlays": {"a": {}}}`, "overlays.a"},
		{"overlay empty source", `{"overlays": {"a": [{"range": "^1"}]}}`, "source"},
		{"overlay no form", `{"overlays": {"a": [{"source": "s"}]}}`, "exactly one"},
		{"overlay path no form", `{"overlays": {"a": [{"source": "/p"}]}}`, "exactly one"},
		{"overlay two forms", `{"overlays": {"a": [{"source": "s", "range": "^1", "tag": "v1"}]}}`, "exactly one"},
		{"overlay range grammar", `{"overlays": {"a": [{"source": "s", "range": "^1 (x)"}]}}`, "range"},
		{"overlay tag grammar", `{"overlays": {"a": [{"source": "s", "tag": "v1..x"}]}}`, "tag"},
		{"overlay revision grammar", `{"overlays": {"a": [{"source": "s", "revision": "ABC"}]}}`, "revision"},
		{"overlay directory traversal", `{"overlays": {"a": [{"source": "s", "range": "^1", "directory": "../x"}]}}`, "directory"},
		{"overlay negative weight", `{"overlays": {"a": [{"source": "s", "range": "^1", "weight": -1}]}}`, "weight"},
		{"overlay unknown field", `{"overlays": {"a": [{"source": "s", "range": "^1", "branch": "main"}]}}`, "unsupported"},
		{"overlay default weight negative", `{"overlay_default_weight": -1}`, "overlay_default_weight"},
		{"overlays allowed type", `{"overlays_allowed": "yes"}`, "overlays_allowed"},
		{"precedence winner", `{"precedence": {"winner": "heavier"}}`, "winner"},
		{"precedence placement", `{"precedence": {"placement": "middle"}}`, "placement"},
		{"precedence unknown", `{"precedence": {"tiebreak": "x"}}`, "unsupported"},
		{"form value", `{"forms": {"pi": "scroll"}}`, "forms.pi"},
		{"form key grammar", `{"forms": {"-bad": "monolithic"}}`, "forms"},
		{"in-place value", `{"in_place_mode": {"pi": "fax"}}`, "in_place_mode.pi"},
		{"system prompt env", `{"system_prompt_files": {"a": {"codex_cli": "append"}}}`, "unsupported"},
		{"system prompt value", `{"system_prompt_files": {"a": {"pi": "sometimes"}}}`, "system_prompt_files.a.pi"},
		{"target participation", `{"targets": {"t": {"participation": "maybe"}}}`, "participation"},
		{"target consented type", `{"targets": {"t": {"consented": "yes"}}}`, "consented"},
		{"target unknown", `{"targets": {"t": {"mood": "x"}}}`, "unsupported"},
		{"isolation value", `{"isolation": {"a": {"pi": "alone"}}}`, "isolation.a.pi"},
		{"xdg opencode", `{"xdg_seed_allowlist": ["opencode"]}`, "xdg_seed_allowlist"},
		{"xdg separator", `{"xdg_seed_allowlist": ["a/b"]}`, "xdg_seed_allowlist"},
		{"xdg duplicate", `{"xdg_seed_allowlist": ["git", "git"]}`, "xdg_seed_allowlist"},
		{"passable grammar", `{"passable_env_names": ["OK", "-bad"]}`, "passable_env_names"},
		{"mcp duplicate", `{"mcp_package_allowlist": ["s", "s"]}`, "mcp_package_allowlist"},
		{"mcp empty", `{"mcp_package_allowlist": [""]}`, "mcp_package_allowlist"},
		{"shadow missing path", `{"shadow_acknowledged": [{"env": "pi"}]}`, "shadow_acknowledged[0].path"},
		{"shadow env grammar", `{"shadow_acknowledged": [{"env": "-x", "path": "p"}]}`, "shadow_acknowledged[0].env"},
		{"shadow unknown", `{"shadow_acknowledged": [{"env": "pi", "path": "p", "why": "x"}]}`, "unsupported"},
		{"waiver pin grammar", `{"secret_material_waivers": [{"pin": "zz", "file": "f", "span": [1, 2], "reason": "r"}]}`, "secret_material_waivers[0].pin"},
		{"waiver span arity", `{"secret_material_waivers": [{"pin": "` + strings.Repeat("a", 40) + `", "file": "f", "span": [1], "reason": "r"}]}`, "span"},
		{"waiver file traversal", `{"secret_material_waivers": [{"pin": "` + strings.Repeat("a", 40) + `", "file": "../f", "span": [1, 2], "reason": "r"}]}`, "secret_material_waivers[0].file"},
		{"waiver unknown", `{"secret_material_waivers": [{"pin": "` + strings.Repeat("a", 40) + `", "file": "f", "span": [1, 2], "reason": "r", "extra": 1}]}`, "unsupported"},
		{"backup negative", `{"backup_retention": -1}`, "backup_retention"},
		{"require grammar", `{"require_current_profile": "-bad"}`, "require_current_profile"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			text := strings.Replace(base, "%s", tc.env, 1)
			loadFails(t, text, tc.want)
		})
	}
}

func TestSchema2EveryKnobParses(t *testing.T) {
	cfg := loadText(t, `{"schema_version": 2, "skills_root": "x", "projects": {}, "environments": {
		"current_profile": "companyA", "scoped_current": {"pi": "personal"},
		"overlays": {"companyA": [{"source": "https://example.com/p", "range": "^1", "weight": 7}]},
		"overlay_default_weight": 500, "overlays_allowed": false,
		"precedence": {"winner": "lower-weight", "placement": "winner-first"},
		"forms": {"claude_code": "referenced"}, "system_prompt_files": {"companyA": {"pi": "append"}},
		"targets": {"xcode-coding-assistant": {"participation": "enabled", "consented": true}},
		"isolation": {"companyA": {"pi": "isolated"}},
		"xdg_seed_allowlist": ["git"], "passable_env_names": ["A_B"],
		"mcp_package_allowlist": ["https://example.com/m"],
		"shadow_acknowledged": [{"env": "pi", "path": "AGENTS.override.md"}],
		"secret_material_waivers": [{"pin": "`+strings.Repeat("c", 40)+`", "file": "context/r.md", "span": [1, 2], "reason": "r"}],
		"backup_retention": 0, "require_current_profile": "companyA",
		"in_place_mode": {"pi": "copied"}}}`)
	env := cfg.Env
	if env.CurrentProfile == nil || *env.CurrentProfile != "companyA" {
		t.Fatalf("current_profile: %+v", env.CurrentProfile)
	}
	if env.ScopedCurrent["pi"] != "personal" || env.OverlayDefaultWeight != 500 || env.OverlaysAllowed {
		t.Fatalf("scalars: %+v", env)
	}
	decls := env.Overlays["companyA"]
	if len(decls) != 1 || decls[0].Range != "^1" || decls[0].Weight == nil || *decls[0].Weight != 7 {
		t.Fatalf("overlays: %+v", env.Overlays)
	}
	if env.Precedence != (Precedence{Winner: "lower-weight", Placement: "winner-first"}) {
		t.Fatalf("precedence: %+v", env.Precedence)
	}
	if env.Forms["claude_code"] != "referenced" || env.SystemPromptFiles["companyA"].Pi != "append" {
		t.Fatalf("forms/spf: %+v %+v", env.Forms, env.SystemPromptFiles)
	}
	if env.Targets["xcode-coding-assistant"] != (TargetConfig{Participation: "enabled", Consented: true}) {
		t.Fatalf("targets: %+v", env.Targets)
	}
	if env.Isolation["companyA"]["pi"] != "isolated" || env.BackupRetention != 0 {
		t.Fatalf("isolation/retention: %+v %d", env.Isolation, env.BackupRetention)
	}
	if len(env.PassableEnvNames) != 1 || env.RequireCurrent == nil || *env.RequireCurrent != "companyA" {
		t.Fatalf("passable/require: %+v %+v", env.PassableEnvNames, env.RequireCurrent)
	}
	if env.InPlaceMode["pi"] != "copied" || len(env.SecretWaivers) != 1 || len(env.ShadowAcknowledged) != 1 {
		t.Fatalf("tail: %+v", env)
	}
	// overlays_allowed false empties every overlay list (§12.2).
	if got := env.EffectiveOverlays("companyA"); len(got) != 0 {
		t.Fatalf("EffectiveOverlays with overlays_allowed=false = %v, want empty", got)
	}
}

// TestOverlaysAllowedTrueKeepsLists proves the other side of the §12.2
// composition policy: the gate empties lists only when the knob is false.
func TestOverlaysAllowedTrueKeepsLists(t *testing.T) {
	cfg := loadText(t, `{"schema_version": 2, "skills_root": "x", "projects": {},
		"environments": {"overlays": {"a": [{"source": "s", "range": "^1"}]}}}`)
	if got := cfg.Env.EffectiveOverlays("a"); len(got) != 1 {
		t.Fatalf("EffectiveOverlays = %v, want one declaration", got)
	}
}

// TestPathOverlayDeclarationParses proves the section 12.1 grammar: every
// overlay carries exactly one requirement form, git or path alike — the
// published family accepts a revision on a path source. The section 1
// refusal for a form on a path source fires at resolution, not at the
// reader.
func TestPathOverlayDeclarationParses(t *testing.T) {
	cfg := loadText(t, `{"schema_version": 2, "skills_root": "x", "projects": {},
		"environments": {"overlays": {"a": [{"source": "/srv/overlays/personal",
			"revision": "abababababababababababababababababababab", "weight": 3}]}}}`)
	decls := cfg.Env.Overlays["a"]
	if len(decls) != 1 || decls[0].Source != "/srv/overlays/personal" {
		t.Fatalf("overlays: %+v", cfg.Env.Overlays)
	}
	if decls[0].Revision != "abababababababababababababababababababab" {
		t.Fatalf("path overlay revision: %+v", decls[0])
	}
	if decls[0].Weight == nil || *decls[0].Weight != 3 {
		t.Fatalf("path overlay weight: %+v", decls[0].Weight)
	}
}

func TestSystemV2LockedKnobs(t *testing.T) {
	cfg, warnings, err := loadWithSystem(t,
		`{"schema_version": 2, "skills_root": "x", "projects": {},
		  "environments": {"precedence": {"winner": "lower-weight"}, "overlays_allowed": true}}`,
		`{"schema_version": 2, "locked": ["environments.precedence", "environments.overlays_allowed"],
		  "environments": {"precedence": {"winner": "higher-weight", "placement": "winner-first"}, "overlays_allowed": false}}`)
	if err != nil {
		t.Fatal(err)
	}
	// Locked keys win whole: precedence replaces both primitives even
	// though the user set only the winner.
	if cfg.Env.Precedence != (Precedence{Winner: "higher-weight", Placement: "winner-first"}) {
		t.Fatalf("locked precedence did not replace whole: %+v", cfg.Env.Precedence)
	}
	if cfg.Env.OverlaysAllowed {
		t.Fatalf("locked overlays_allowed did not win")
	}
	if len(warnings) != 2 {
		t.Fatalf("warnings = %v, want two locked-key warnings", warnings)
	}
	if !cfg.Locked["environments.precedence"] || !cfg.Locked["environments.overlays_allowed"] {
		t.Fatalf("locked set not recorded: %v", cfg.Locked)
	}
	if got := cfg.Env.EffectiveOverlays("any"); len(got) != 0 {
		t.Fatalf("locked overlays_allowed=false must empty every list")
	}
}

func TestSystemV2UnlockedEnvIsDefault(t *testing.T) {
	cfg, warnings, err := loadWithSystem(t,
		`{"schema_version": 2, "skills_root": "x", "projects": {},
		  "environments": {"precedence": {"winner": "lower-weight"}}}`,
		`{"schema_version": 2, "locked": [],
		  "environments": {"precedence": {"winner": "higher-weight"}, "overlays_allowed": false}}`)
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unlocked defaults must not warn: %v", warnings)
	}
	// The machine knob replaces the unlocked system knob whole.
	if cfg.Env.Precedence.Winner != "lower-weight" || cfg.Env.Precedence.Placement != "winner-last" {
		t.Fatalf("machine knob did not replace whole: %+v", cfg.Env.Precedence)
	}
	if cfg.Env.OverlaysAllowed {
		t.Fatalf("unlocked system default not applied: %+v", cfg.Env)
	}
}

func TestSystemV2Refusals(t *testing.T) {
	user := `{"schema_version": 2, "skills_root": "x", "projects": {}}`
	cases := []struct {
		name   string
		system string
		want   string
	}{
		{"unlockable knob carried", `{"schema_version": 2, "environments": {"current_profile": "a"}}`, "not lockable"},
		{"unlockable knob locked", `{"schema_version": 2, "locked": ["environments.current_profile"], "environments": {"current_profile": "a"}}`, "cannot lock"},
		{"bare environments locked", `{"schema_version": 2, "locked": ["environments"]}`, "cannot lock"},
		{"locked but unset", `{"schema_version": 2, "locked": ["environments.precedence"]}`, "locks"},
		{"isolated direction", `{"schema_version": 2, "locked": ["environments.isolation"], "environments": {"isolation": {"a": {"pi": "isolated"}}}}`, "only toward shared"},
		{"bad value grammar", `{"schema_version": 2, "environments": {"overlays_allowed": "yes"}}`, "overlays_allowed"},
		{"env under schema 1", `{"schema_version": 1, "locked": [], "environments": {"backup_retention": 1}}`, "environments"},
		{"env lock under schema 1", `{"schema_version": 1, "locked": ["environments.backup_retention"]}`, "schema_version 1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := loadWithSystem(t, user, tc.system)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want mention of %q", err, tc.want)
			}
		})
	}
}

// TestSystemV2DefaultsUnderSchema1User proves the manager §1 rule-3
// promotion: a schema-1 user file over system environments defaults keeps
// its on-disk version but the effective configuration carries the knobs.
func TestSystemV2DefaultsUnderSchema1User(t *testing.T) {
	cfg, _, err := loadWithSystem(t, minimal,
		`{"schema_version": 2, "locked": [], "environments": {"overlays_allowed": false}}`)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Schema != 2 || cfg.Env.OverlaysAllowed {
		t.Fatalf("effective config = schema %d overlays_allowed %v, want 2 and false", cfg.Schema, cfg.Env.OverlaysAllowed)
	}
}

// TestRequireCurrentProfileGate drives Load with a locked
// require_current_profile: the knob and its lock are carried on the loaded
// configuration, which PolicyFromConfig threads into the production refusal
// (Policy.CheckMachineUse at the useLocked seam, driven through run() by the
// profile-use and profile-install tests). The refusal itself lives there,
// not here.
func TestRequireCurrentProfileGate(t *testing.T) {
	cfg, _, err := loadWithSystem(t,
		`{"schema_version": 2, "skills_root": "x", "projects": {}}`,
		`{"schema_version": 2, "locked": ["environments.require_current_profile"],
		  "environments": {"require_current_profile": "companyA"}}`)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Env.RequireCurrent == nil || *cfg.Env.RequireCurrent != "companyA" {
		t.Fatalf("locked knob not carried: %+v", cfg.Env.RequireCurrent)
	}
	if !cfg.Locked["environments.require_current_profile"] {
		t.Fatalf("locked knob carries no lock: %+v", cfg.Locked)
	}
}

// TestRequireCurrentProfileUnlockedCarriesNoRefusal pins the bound: the
// refusal is a locked-key effect, so a user-set knob without the lock
// carries the value with no lock for the seam to refuse on.
func TestRequireCurrentProfileUnlockedCarriesNoRefusal(t *testing.T) {
	cfg := loadText(t, `{"schema_version": 2, "skills_root": "x", "projects": {},
		"environments": {"require_current_profile": "companyA"}}`)
	if cfg.Env.RequireCurrent == nil || *cfg.Env.RequireCurrent != "companyA" {
		t.Fatalf("unlocked knob not carried: %+v", cfg.Env.RequireCurrent)
	}
	if cfg.Locked["environments.require_current_profile"] {
		t.Fatalf("unlocked knob must carry no lock: %+v", cfg.Locked)
	}
}

// TestSystemV2LockableSubsetIsClassWide proves the system-file allowed-knob
// gate is class-wide, not per-knob-name. The knob list is EnvKnobNames (the
// §12.1 table the reader enforces) with a grammatically valid system-file
// payload per knob, so a knob added to the reader without a payload fails
// here. The admitted set is transcribed from environments §12.2 — the six
// keys the specification states, written out independently — and asserted
// set-equal with LockableEnvKeys (the map parseSystemEnvironments enforces
// via systemEnvKnobs), so a knob added to or removed from that map without
// §12.2 standing fails here instead of widening or narrowing the system
// file's reach silently. A narrowing mutant that admits exactly one
// non-lockable knob (for example secret_material_waivers,
// xdg_seed_allowlist or in_place_mode) must fail this test, and so must a
// widening mutant that adds one member to LockableEnvKeys.
func TestSystemV2LockableSubsetIsClassWide(t *testing.T) {
	user := `{"schema_version": 2, "skills_root": "x", "projects": {}}`
	payloads := map[string]string{
		"current_profile":         `{"current_profile": "a"}`,
		"scoped_current":          `{"scoped_current": {"codex_cli": "a"}}`,
		"overlays":                `{"overlays": {}}`,
		"overlay_default_weight":  `{"overlay_default_weight": 5}`,
		"overlays_allowed":        `{"overlays_allowed": false}`,
		"precedence":              `{"precedence": {"winner": "lower-weight"}}`,
		"forms":                   `{"forms": {"claude_code": "monolithic"}}`,
		"system_prompt_files":     `{"system_prompt_files": {"a": {"pi": "append"}}}`,
		"targets":                 `{"targets": {"codex_cli": {"participation": "off"}}}`,
		"isolation":               `{"isolation": {"a": {"pi": "shared"}}}`,
		"xdg_seed_allowlist":      `{"xdg_seed_allowlist": ["git"]}`,
		"passable_env_names":      `{"passable_env_names": []}`,
		"mcp_package_allowlist":   `{"mcp_package_allowlist": []}`,
		"shadow_acknowledged":     `{"shadow_acknowledged": [{"env": "codex_cli", "path": "some/path"}]}`,
		"secret_material_waivers": `{"secret_material_waivers": [{"pin": "0123456789abcdef0123456789abcdef01234567", "file": "context/a.md", "span": [1, 2], "reason": "org policy"}]}`,
		"backup_retention":        `{"backup_retention": 1}`,
		"require_current_profile": `{"require_current_profile": "acme"}`,
		"in_place_mode":           `{"in_place_mode": {"codex_cli": "linked"}}`,
	}
	if len(payloads) != len(EnvKnobNames) {
		t.Fatalf("payloads cover %d knobs, EnvKnobNames carries %d: keep the two in step", len(payloads), len(EnvKnobNames))
	}
	// The §12.2 transcription: exactly the six keys the specification
	// states ("overlays_allowed, precedence, mcp_package_allowlist,
	// passable_env_names, require_current_profile, and isolation"). This
	// literal is the spec sentence; LockableEnvKeys is the implementation.
	// Either drifting — a widening that would reach §9.1 secret material,
	// or a narrowing that would drop a fleet-policy knob — fails below.
	transcribed := map[string]bool{
		"environments.overlays_allowed":        true,
		"environments.precedence":              true,
		"environments.mcp_package_allowlist":   true,
		"environments.passable_env_names":      true,
		"environments.require_current_profile": true,
		"environments.isolation":               true,
	}
	if len(LockableEnvKeys) != len(transcribed) {
		t.Fatalf("LockableEnvKeys carries %d keys, §12.2 transcribes %d: a widening or narrowing without spec standing fails here", len(LockableEnvKeys), len(transcribed))
	}
	for key := range transcribed {
		if !LockableEnvKeys[key] {
			t.Fatalf("§12.2 key %q missing from LockableEnvKeys: the implementation narrowed the transcribed set", key)
		}
	}
	for key := range LockableEnvKeys {
		if !transcribed[key] {
			t.Fatalf("LockableEnvKeys key %q has no §12.2 standing: the implementation widened the transcribed set", key)
		}
	}
	for _, knob := range EnvKnobNames {
		body, ok := payloads[knob]
		if !ok {
			t.Fatalf("no system-file payload for knob %q: add one or the gate is untested", knob)
		}
		system := `{"schema_version": 2, "locked": [], "environments": ` + body + `}`
		if transcribed["environments."+knob] {
			t.Run("lockable/"+knob, func(t *testing.T) {
				if _, _, err := loadWithSystem(t, user, system); err != nil {
					t.Fatalf("lockable knob %q refused: %v", knob, err)
				}
			})
			continue
		}
		t.Run("non-lockable/"+knob, func(t *testing.T) {
			_, _, err := loadWithSystem(t, user, system)
			if err == nil || !strings.Contains(err.Error(), "not lockable") {
				t.Fatalf("non-lockable knob %q admitted: err = %v", knob, err)
			}
		})
	}
}

func TestEffectiveJSONShape(t *testing.T) {
	cfg := loadText(t, `{"schema_version": 2, "skills_root": "x", "projects": {}}`)
	payload, err := json.Marshal(cfg.EffectiveJSON())
	if err != nil {
		t.Fatal(err)
	}
	var rendered map[string]any
	if err := json.Unmarshal(payload, &rendered); err != nil {
		t.Fatal(err)
	}
	env, ok := rendered["environments"].(map[string]any)
	if !ok {
		t.Fatalf("no environments object: %v", rendered)
	}
	for _, knob := range []string{"current_profile", "scoped_current", "overlays",
		"overlay_default_weight", "overlays_allowed", "precedence", "forms",
		"system_prompt_files", "targets", "isolation", "xdg_seed_allowlist",
		"passable_env_names", "mcp_package_allowlist", "shadow_acknowledged",
		"secret_material_waivers", "backup_retention", "require_current_profile",
		"in_place_mode"} {
		if _, present := env[knob]; !present {
			t.Fatalf("rendered environments lack %q", knob)
		}
	}
	if env["overlay_default_weight"] != float64(1000) || env["overlays_allowed"] != true {
		t.Fatalf("scalar defaults: %v %v", env["overlay_default_weight"], env["overlays_allowed"])
	}
}
