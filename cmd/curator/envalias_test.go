package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/config"
)

// These tests drive the production run() entry point for the CLI alias
// rule (manager profile §12.1): claude and codex normalize to claude_code
// and codex_cli before validation, outputs keep the canonical id, aliases
// are never persisted, and unknown spellings are still refused.

// TestEnvResolveAcceptsAliases drives both spellings through env resolve:
// the fragment carries the canonical environment, the bare resolve emits
// bytes identical to the canonical spelling, and the env/shell formats
// name the canonical home variable.
func TestEnvResolveAcceptsAliases(t *testing.T) {
	source, _ := profileHome(t)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("install stderr:\n%s", stderr)
	}
	cases := []struct{ alias, canonical, variable string }{
		{"claude", "claude_code", "CLAUDE_CONFIG_DIR"},
		{"codex", "codex_cli", "CODEX_HOME"},
	}
	for _, tc := range cases {
		code, stdout, stderr := runProfile(t, source, "env", "resolve", tc.alias, "--repair")
		if code != exitOK {
			t.Fatalf("resolve %s --repair = %d\nstderr:\n%s", tc.alias, code, stderr)
		}
		var fragment map[string]any
		if err := json.Unmarshal([]byte(stdout), &fragment); err != nil {
			t.Fatalf("fragment is not JSON: %v\nstdout:\n%s", err, stdout)
		}
		if fragment["environment"] != tc.canonical {
			t.Fatalf("resolve %s fragment environment = %v, want %s", tc.alias, fragment["environment"], tc.canonical)
		}
		if code, again, _ := runProfile(t, source, "env", "resolve", tc.alias); code != exitOK || again != stdout {
			t.Fatalf("bare resolve %s = %d, identical bytes: %v", tc.alias, code, again == stdout)
		}
		if code, canonical, _ := runProfile(t, source, "env", "resolve", tc.canonical); code != exitOK || canonical != stdout {
			t.Fatalf("resolve %s = %d, alias/canonical bytes identical: %v", tc.canonical, code, canonical == stdout)
		}
		if code, envOut, _ := runProfile(t, source, "env", "resolve", tc.alias, "--format", "env"); code != exitOK || !strings.HasPrefix(envOut, tc.variable+"=") {
			t.Fatalf("env format %s = %d %q", tc.alias, code, envOut)
		}
		if code, shellOut, _ := runProfile(t, source, "env", "resolve", tc.alias, "--format", "shell"); code != exitOK || !strings.HasPrefix(shellOut, "export "+tc.variable+"='") {
			t.Fatalf("shell format %s = %d %q", tc.alias, code, shellOut)
		}
	}
}

// TestEnvResolveAliasAdjacentUnknownsRefused narrows the refusal around
// the aliases: near-miss spellings still fail with environment_unknown,
// never by normalizing.
func TestEnvResolveAliasAdjacentUnknownsRefused(t *testing.T) {
	source, _ := profileHome(t)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("install stderr:\n%s", stderr)
	}
	for _, id := range []string{"Claude", "Codex", "claude-code", "codexcli", "claud", "CLAUDE"} {
		code, _, stderr := runProfile(t, source, "env", "resolve", id)
		if code != exitFail || !strings.Contains(stderr, "environment_unknown") {
			t.Fatalf("resolve %s = %d\nstderr:\n%s", id, code, stderr)
		}
	}
}

// TestProfileUseEnvAliases drives both spellings through profile use
// --env: the switch reports the canonical adapter, the scoped record is
// filed under the canonical scope key, and the clear form accepts the
// alias too.
func TestProfileUseEnvAliases(t *testing.T) {
	source, home := profileHome(t)
	first := t.TempDir()
	writeContextPackage(t, first, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", first); code != exitOK {
		t.Fatalf("install stderr:\n%s", stderr)
	}
	second := t.TempDir()
	writeContextPackage(t, second, "other", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", second, "--as", "other"); code != exitOK {
		t.Fatalf("install other stderr:\n%s", stderr)
	}
	for alias, canonical := range map[string]string{"claude": "claude_code", "codex": "codex_cli"} {
		code, stdout, stderr := runProfile(t, source, "profile", "use", "other", "--env", alias)
		if code != exitOK {
			t.Fatalf("use --env %s = %d\nstdout:\n%s\nstderr:\n%s", alias, code, stdout, stderr)
		}
		if !strings.Contains(stdout, canonical+": switched") {
			t.Fatalf("use --env %s stdout:\n%s", alias, stdout)
		}
		if strings.Contains(stdout, alias+": switched") {
			t.Fatalf("use --env %s printed the alias:\n%s", alias, stdout)
		}
		// A scoped switch to another profile records a scoped current;
		// the record file must carry the canonical scope key.
		entries, err := os.ReadDir(filepath.Join(home, "profiles", "scoped"))
		if err != nil {
			t.Fatal(err)
		}
		var names []string
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		if want := "env%3A" + canonical; !containsName(names, want) {
			t.Fatalf("scoped records = %v, want %s", names, want)
		}
		if bad := "env%3A" + alias; containsName(names, bad) {
			t.Fatalf("scoped records persist the alias: %v", names)
		}
		if code, _, stderr := runProfile(t, source, "profile", "use", "--clear", "--env", alias); code != exitOK {
			t.Fatalf("clear --env %s = %d\nstderr:\n%s", alias, code, stderr)
		}
	}
}

// TestProfileUseUnknownEnvRefused narrows the --env gate: an unknown
// spelling fails with environment_unknown without moving any current.
func TestProfileUseUnknownEnvRefused(t *testing.T) {
	source, home := profileHome(t)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("install stderr:\n%s", stderr)
	}
	code, _, stderr := runProfile(t, source, "profile", "use", "acme", "--env", "cursor")
	if code != exitFail || !strings.Contains(stderr, "environment_unknown") {
		t.Fatalf("use --env cursor = %d\nstderr:\n%s", code, stderr)
	}
	if entries, err := os.ReadDir(filepath.Join(home, "profiles", "scoped")); err == nil && len(entries) != 0 {
		t.Fatalf("a refused switch recorded scoped state: %v", entries)
	} else if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
}

// TestEnvStatusPrintsCanonicalAfterAliasUse drives an alias switch and
// proves env status rows carry the canonical id only: the home rows, the
// scope rows, and the tool rows.
func TestEnvStatusPrintsCanonicalAfterAliasUse(t *testing.T) {
	source, _ := profileHome(t)
	first := t.TempDir()
	writeContextPackage(t, first, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", first); code != exitOK {
		t.Fatalf("install stderr:\n%s", stderr)
	}
	second := t.TempDir()
	writeContextPackage(t, second, "other", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", second, "--as", "other"); code != exitOK {
		t.Fatalf("install other stderr:\n%s", stderr)
	}
	if code, _, stderr := runProfile(t, source, "profile", "use", "other", "--env", "claude"); code != exitOK {
		t.Fatalf("use --env claude = %d\nstderr:\n%s", code, stderr)
	}
	if code, _, stderr := runProfile(t, source, "env", "resolve", "codex", "--repair"); code != exitOK {
		t.Fatalf("resolve codex --repair = %d\nstderr:\n%s", code, stderr)
	}
	code, stdout, stderr := runProfile(t, source, "env", "status")
	if code != exitOK {
		t.Fatalf("status = %d\nstderr:\n%s", code, stderr)
	}
	for _, want := range []string{"other claude_code:", "acme codex_cli:", "scope env:claude_code:", "tool claude_code:", "tool codex_cli:"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("status misses %q:\n%s", want, stdout)
		}
	}
	// The canonical spellings carry an underscore where the alias ends,
	// so a bare " claude:"/" codex:" row is an alias leak, not a prefix
	// of a canonical row.
	for _, leaked := range []string{" claude:", " codex:", "scope env:claude:", "scope env:codex:", "tool claude:", "tool codex:"} {
		if strings.Contains(stdout, leaked) {
			t.Fatalf("status leaks the alias in %q:\n%s", leaked, stdout)
		}
	}
}

// TestEnvConfigSetNormalizesEnvKnobs drives alias knob paths through env
// config set: the machine file persists the canonical id, show reads it
// back through either spelling, and unset accepts the alias too.
func TestEnvConfigSetNormalizesEnvKnobs(t *testing.T) {
	source := writeMachineConfig(t, `{"schema_version": 2, "skills_root": "x", "projects": {}}`)
	type knobCase struct {
		aliasKnob, canonicalKnob, value, want string
	}
	for _, tc := range []knobCase{
		{"forms.claude", "forms.claude_code", "referenced", `"referenced"`},
		{"isolation.acme.codex", "isolation.acme.codex_cli", "shared", `"shared"`},
		{"in_place_mode.claude", "in_place_mode.claude_code", "linked", `"linked"`},
		{"scoped_current.codex", "scoped_current.codex_cli", "acme", `"acme"`},
	} {
		if code, _, stderr := runProfile(t, source, "env", "config", "set", tc.aliasKnob, tc.value); code != exitOK {
			t.Fatalf("set %s = %d\nstderr:\n%s", tc.aliasKnob, code, stderr)
		}
		source = reloadSource(t, source)
		for _, spelling := range []string{tc.aliasKnob, tc.canonicalKnob} {
			code, stdout, stderr := runProfile(t, source, "env", "config", "show", spelling)
			if code != exitOK || strings.TrimSpace(stdout) != tc.want {
				t.Fatalf("show %s = %d %q, want %s\nstderr:\n%s", spelling, code, stdout, tc.want, stderr)
			}
		}
		raw, err := os.ReadFile(source.path)
		if err != nil {
			t.Fatal(err)
		}
		var object map[string]any
		if err := json.Unmarshal(raw, &object); err != nil {
			t.Fatal(err)
		}
		serialized, _ := json.Marshal(object["environments"])
		for _, leaked := range []string{`"claude":`, `"codex":`} {
			if strings.Contains(string(serialized), leaked) {
				t.Fatalf("machine file persists the alias after set %s: %s", tc.aliasKnob, serialized)
			}
		}
		if code, _, stderr := runProfile(t, source, "env", "config", "unset", tc.aliasKnob); code != exitOK {
			t.Fatalf("unset %s = %d\nstderr:\n%s", tc.aliasKnob, code, stderr)
		}
		source = reloadSource(t, source)
	}
}

// TestEnvConfigSetNormalizesWholesaleValues drives wholesale values that
// name environments: env-keyed map keys and shadow_acknowledged env
// members normalize, and an explicit canonical key wins over its alias.
func TestEnvConfigSetNormalizesWholesaleValues(t *testing.T) {
	source := writeMachineConfig(t, `{"schema_version": 2, "skills_root": "x", "projects": {}}`)
	if code, _, stderr := runProfile(t, source, "env", "config", "set", "forms", `{"codex": "monolithic"}`); code != exitOK {
		t.Fatalf("set forms = %d\nstderr:\n%s", code, stderr)
	}
	source = reloadSource(t, source)
	if got := source.cfg.Env.Forms["codex_cli"]; got != "monolithic" {
		t.Fatalf("forms = %+v, want codex_cli normalized", source.cfg.Env.Forms)
	}
	if _, leaked := source.cfg.Env.Forms["codex"]; leaked {
		t.Fatalf("forms persist the alias: %+v", source.cfg.Env.Forms)
	}
	if code, _, stderr := runProfile(t, source, "env", "config", "set", "isolation", `{"acme": {"claude": "isolated"}}`); code != exitOK {
		t.Fatalf("set isolation = %d\nstderr:\n%s", code, stderr)
	}
	source = reloadSource(t, source)
	if got := source.cfg.Env.Isolation["acme"]["claude_code"]; got != "isolated" {
		t.Fatalf("isolation = %+v, want claude_code normalized", source.cfg.Env.Isolation)
	}
	if code, _, stderr := runProfile(t, source, "env", "config", "set", "shadow_acknowledged", `[{"env": "codex", "path": "AGENTS.override.md"}]`); code != exitOK {
		t.Fatalf("set shadow_acknowledged = %d\nstderr:\n%s", code, stderr)
	}
	source = reloadSource(t, source)
	if len(source.cfg.Env.ShadowAcknowledged) != 1 || source.cfg.Env.ShadowAcknowledged[0].Env != "codex_cli" {
		t.Fatalf("shadow_acknowledged = %+v, want codex_cli normalized", source.cfg.Env.ShadowAcknowledged)
	}
	// Both spellings at once: the explicit canonical key wins, so the
	// rewrite is deterministic rather than map-ordered.
	if code, _, stderr := runProfile(t, source, "env", "config", "set", "forms", `{"claude": "monolithic", "claude_code": "referenced"}`); code != exitOK {
		t.Fatalf("set forms collision = %d\nstderr:\n%s", code, stderr)
	}
	source = reloadSource(t, source)
	if got := source.cfg.Env.Forms["claude_code"]; got != "referenced" {
		t.Fatalf("forms collision = %+v, want the canonical spelling to win", source.cfg.Env.Forms)
	}
}

// TestEnvConfigAliasOutputPrintsCanonicalKnob drives the set/unset output
// shapes through run(): a set via an alias prints the value JSON with no
// knob leak, an unset via an alias prints "unset <canonical>", and an
// unset of an absent alias knob refuses naming the canonical knob. The
// canonical spellings produce the identical bytes (controls), so the test
// proves the alias and canonical paths converge, not just that they exit.
func TestEnvConfigAliasOutputPrintsCanonicalKnob(t *testing.T) {
	cases := []struct{ aliasKnob, canonicalKnob, value, wantJSON string }{
		{"forms.claude", "forms.claude_code", "referenced", `"referenced"`},
		{"forms.codex", "forms.codex_cli", "monolithic", `"monolithic"`},
	}
	for _, tc := range cases {
		// Not-set refusal through the alias names the canonical knob.
		// The quoted alias ("forms.claude") is not a substring of the
		// quoted canonical ("forms.claude_code"), so the negative
		// assertion is exact, not a prefix accident.
		refusal := writeMachineConfig(t, `{"schema_version": 2, "skills_root": "x", "projects": {}}`)
		code, _, stderr := runProfile(t, refusal, "env", "config", "unset", tc.aliasKnob)
		if code != exitFail {
			t.Fatalf("unset %s = %d, want the not-set refusal", tc.aliasKnob, code)
		}
		if !strings.Contains(stderr, `"`+tc.canonicalKnob+`" is not set`) {
			t.Fatalf("unset %s stderr misses the canonical refusal:\n%s", tc.aliasKnob, stderr)
		}
		if strings.Contains(stderr, `"`+tc.aliasKnob+`"`) {
			t.Fatalf("unset %s stderr leaks the alias:\n%s", tc.aliasKnob, stderr)
		}
		code, _, controlStderr := runProfile(t, refusal, "env", "config", "unset", tc.canonicalKnob)
		if code != exitFail || controlStderr != stderr {
			t.Fatalf("unset %s = %d, refusal bytes identical to the alias: %v", tc.canonicalKnob, code, controlStderr == stderr)
		}
		// Set success through the alias prints the value only.
		source := writeMachineConfig(t, `{"schema_version": 2, "skills_root": "x", "projects": {}}`)
		code, stdout, stderr := runProfile(t, source, "env", "config", "set", tc.aliasKnob, tc.value)
		if code != exitOK {
			t.Fatalf("set %s = %d\nstderr:\n%s", tc.aliasKnob, code, stderr)
		}
		if strings.TrimSpace(stdout) != tc.wantJSON {
			t.Fatalf("set %s printed %q, want %s", tc.aliasKnob, stdout, tc.wantJSON)
		}
		if strings.Contains(stdout, "unset ") || strings.Contains(stdout, tc.canonicalKnob) {
			t.Fatalf("set %s stdout leaks a knob line:\n%s", tc.aliasKnob, stdout)
		}
		source = reloadSource(t, source)
		// Unset success through the alias prints the canonical knob.
		// The canonical spelling extends the alias (claude vs
		// claude_code), so only an exact match proves the render.
		code, stdout, stderr = runProfile(t, source, "env", "config", "unset", tc.aliasKnob)
		if code != exitOK {
			t.Fatalf("unset %s = %d\nstderr:\n%s", tc.aliasKnob, code, stderr)
		}
		if want := "unset " + tc.canonicalKnob + "\n"; stdout != want {
			t.Fatalf("unset %s printed %q, want %q", tc.aliasKnob, stdout, want)
		}
		// Canonical control: the same set/unset pair via the canonical
		// spelling prints the identical bytes.
		control := writeMachineConfig(t, `{"schema_version": 2, "skills_root": "x", "projects": {}}`)
		code, controlOut, stderr := runProfile(t, control, "env", "config", "set", tc.canonicalKnob, tc.value)
		if code != exitOK || controlOut != strings.TrimSpace(tc.wantJSON)+"\n" {
			t.Fatalf("set %s = %d %q, want the alias-identical value line\nstderr:\n%s", tc.canonicalKnob, code, controlOut, stderr)
		}
		control = reloadSource(t, control)
		code, controlOut, stderr = runProfile(t, control, "env", "config", "unset", tc.canonicalKnob)
		if code != exitOK || controlOut != "unset "+tc.canonicalKnob+"\n" {
			t.Fatalf("unset %s = %d %q\nstderr:\n%s", tc.canonicalKnob, code, controlOut, stderr)
		}
	}
}

// TestEnvConfigAliasLockRefusalPrintsCanonicalKnob drives the manager §1
// locked-key refusal through run() with environments.isolation locked (the
// one alias-bearing table that is lockable): set and unset via an alias
// refuse naming the canonical knob, the canonical spellings refuse with
// the identical bytes (controls), and no refusal writes the machine file.
func TestEnvConfigAliasLockRefusalPrintsCanonicalKnob(t *testing.T) {
	writeLocked := func(t *testing.T) stubConfigSource {
		t.Helper()
		source, _ := profileHome(t)
		home := t.TempDir()
		userPath := filepath.Join(home, "config.json")
		if err := os.WriteFile(userPath, []byte(`{"schema_version": 2, "skills_root": "x", "projects": {}}`), 0o600); err != nil {
			t.Fatal(err)
		}
		systemPath := filepath.Join(home, "system.json")
		if err := os.WriteFile(systemPath, []byte(`{"schema_version": 2, "locked": ["environments.isolation"],
			"environments": {"isolation": {"acme": {"claude_code": "shared", "codex_cli": "shared"}}}}`), 0o600); err != nil {
			t.Fatal(err)
		}
		t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)
		cfg, err := config.Load(userPath, nil)
		if err != nil {
			t.Fatalf("config.Load rejected the fixture: %v", err)
		}
		source.path = userPath
		source.cfg = cfg
		return source
	}
	cases := []struct{ aliasKnob, canonicalKnob string }{
		{"isolation.acme.claude", "isolation.acme.claude_code"},
		{"isolation.acme.codex", "isolation.acme.codex_cli"},
	}
	for _, tc := range cases {
		source := writeLocked(t)
		code, _, stderr := runProfile(t, source, "env", "config", "set", tc.aliasKnob, "isolated")
		if code != exitFail {
			t.Fatalf("set %s = %d, want the lock refusal", tc.aliasKnob, code)
		}
		if !strings.Contains(stderr, `"`+tc.canonicalKnob+`" is locked by`) {
			t.Fatalf("set %s stderr misses the canonical lock refusal:\n%s", tc.aliasKnob, stderr)
		}
		if strings.Contains(stderr, `"`+tc.aliasKnob+`"`) {
			t.Fatalf("set %s stderr leaks the alias:\n%s", tc.aliasKnob, stderr)
		}
		code, _, controlStderr := runProfile(t, source, "env", "config", "set", tc.canonicalKnob, "isolated")
		if code != exitFail || controlStderr != stderr {
			t.Fatalf("set %s = %d, refusal bytes identical to the alias: %v", tc.canonicalKnob, code, controlStderr == stderr)
		}
		code, _, stderr = runProfile(t, source, "env", "config", "unset", tc.aliasKnob)
		if code != exitFail {
			t.Fatalf("unset %s = %d, want the lock refusal", tc.aliasKnob, code)
		}
		if !strings.Contains(stderr, `"`+tc.canonicalKnob+`" is locked by`) {
			t.Fatalf("unset %s stderr misses the canonical lock refusal:\n%s", tc.aliasKnob, stderr)
		}
		if strings.Contains(stderr, `"`+tc.aliasKnob+`"`) {
			t.Fatalf("unset %s stderr leaks the alias:\n%s", tc.aliasKnob, stderr)
		}
		code, _, controlStderr = runProfile(t, source, "env", "config", "unset", tc.canonicalKnob)
		if code != exitFail || controlStderr != stderr {
			t.Fatalf("unset %s = %d, refusal bytes identical to the alias: %v", tc.canonicalKnob, code, controlStderr == stderr)
		}
		payload, err := os.ReadFile(source.path)
		if err != nil {
			t.Fatal(err)
		}
		var object map[string]any
		if err := json.Unmarshal(payload, &object); err != nil {
			t.Fatal(err)
		}
		if _, present := object["environments"]; present {
			t.Fatalf("locked set/unset must not write the machine file: %s", payload)
		}
	}
}

// TestRunDispatchNormalizesAliasOperand drives `run` through the umbrella
// dispatch with a scripted provider: a leading alias arrives canonical,
// while flag values and unknown spellings pass through verbatim for the
// launcher to classify and refuse.
func TestRunDispatchNormalizesAliasOperand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("no bash on this platform")
	}
	source, _ := profileHome(t)
	bin := t.TempDir()
	script := "#!/bin/sh\necho \"provider got: $@\"\n"
	if err := os.WriteFile(filepath.Join(bin, "curator-run"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	for alias, canonical := range map[string]string{"claude": "claude_code", "codex": "codex_cli"} {
		code, stdout, stderr := runProfile(t, source, "run", alias, "--beta")
		if code != exitOK {
			t.Fatalf("run %s = %d\nstderr:\n%s", alias, code, stderr)
		}
		if !strings.Contains(stdout, "provider got: "+canonical+" --beta") {
			t.Fatalf("run %s did not normalize:\n%s", alias, stdout)
		}
	}
	// A later alias is not in operand position: a profile literally named
	// claude travels as a flag value, untouched.
	code, stdout, stderr := runProfile(t, source, "run", "--profile", "claude")
	if code != exitOK {
		t.Fatalf("run --profile claude = %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stdout, "provider got: --profile claude") {
		t.Fatalf("flag values must pass through verbatim:\n%s", stdout)
	}
	// Unknown spellings pass through verbatim: the launcher owns the
	// refusal, the manager rewrites only known aliases.
	code, stdout, stderr = runProfile(t, source, "run", "cursor", "--beta")
	if code != exitOK {
		t.Fatalf("run cursor = %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stdout, "provider got: cursor --beta") {
		t.Fatalf("unknown spellings must pass through verbatim:\n%s", stdout)
	}
}

// TestNormalizeRunEnvOperand pins the run-dispatch rewrite decision
// without a provider fixture, so it runs on every platform including
// Windows, where the scripted-provider test above must skip: a leading
// alias rewrites to canonical, while flag values, unknown spellings,
// and non-run argv pass through untouched.
func TestNormalizeRunEnvOperand(t *testing.T) {
	for _, tc := range []struct{ args, want []string }{
		{[]string{"run", "claude", "--beta"}, []string{"run", "claude_code", "--beta"}},
		{[]string{"run", "codex"}, []string{"run", "codex_cli"}},
		{[]string{"run", "claude_code", "--beta"}, []string{"run", "claude_code", "--beta"}},
		{[]string{"run", "--profile", "claude"}, []string{"run", "--profile", "claude"}},
		{[]string{"run", "cursor", "--beta"}, []string{"run", "cursor", "--beta"}},
		{[]string{"run"}, []string{"run"}},
		{[]string{"status"}, []string{"status"}},
	} {
		got := normalizeRunEnvOperand(tc.args)
		if strings.Join(got, "\x00") != strings.Join(tc.want, "\x00") {
			t.Fatalf("normalizeRunEnvOperand(%q) = %q, want %q", tc.args, got, tc.want)
		}
	}
}

// TestUsageListsAliases proves the help rows name the alias rule at every
// environment-operand surface.
func TestUsageListsAliases(t *testing.T) {
	source, _ := profileHome(t)
	for _, args := range [][]string{
		{"env", "resolve"},
		{"profile", "use"},
		{"env", "config"},
		{"env", "config", "set", "forms.claude"},
	} {
		_, _, stderr := runProfile(t, source, args...)
		if !strings.Contains(stderr, "claude and codex as aliases of claude_code and codex_cli") {
			t.Fatalf("%v help misses the alias rule:\n%s", args, stderr)
		}
	}
	_, stdout, _ := runProfile(t, source, "help")
	// The top-level paragraph wraps across lines; compare folded.
	folded := strings.Join(strings.Fields(stdout), " ")
	if !strings.Contains(folded, "claude and codex as aliases of claude_code and codex_cli") {
		t.Fatalf("top-level help misses the alias rule:\n%s", stdout)
	}
}

// TestNormalizeEnvKnobPositions pins the knob table without a config
// file: only the documented environment positions rewrite, and target
// knobs, bare heads, and deeper paths pass through untouched.
func TestNormalizeEnvKnobPositions(t *testing.T) {
	for knob, want := range map[string]string{
		"forms.claude":            "forms.claude_code",
		"in_place_mode.codex":     "in_place_mode.codex_cli",
		"scoped_current.claude":   "scoped_current.claude_code",
		"isolation.acme.codex":    "isolation.acme.codex_cli",
		"forms.codex_cli":         "forms.codex_cli",
		"isolation.acme.pi":       "isolation.acme.pi",
		"targets.claude":          "targets.claude",
		"system_prompt_files.a.b": "system_prompt_files.a.b",
		"forms":                   "forms",
		"isolation.acme":          "isolation.acme",
	} {
		got := strings.Join(normalizeEnvKnob(strings.Split(knob, ".")), ".")
		if got != want {
			t.Fatalf("normalizeEnvKnob(%q) = %q, want %q", knob, got, want)
		}
	}
}

func containsName(names []string, want string) bool {
	for _, name := range names {
		if name == want {
			return true
		}
	}
	return false
}
