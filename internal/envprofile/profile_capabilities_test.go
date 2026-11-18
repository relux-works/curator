package envprofile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
)

func profileSkillRequest(caps string) string {
	if caps == "" {
		return `{"git":%s,"range":"*"}`
	}
	return `{"git":%s,"range":"*","capabilities":` + caps + `}`
}

func writeSkillProfile(t *testing.T, profile, name, request string, schema int) {
	t.Helper()
	writeManifestPackage(t, profile, `{"schema_version":`+strconv.Itoa(schema)+`,"name":"`+name+`","version":"1.0.0","requires":{"skills":{"guide":`+request+`}}}`, nil)
}

func TestProfileCapabilitiesAuditAndFreeze(t *testing.T) {
	home := gitFixtureHome(t)
	pinHomes(t)
	ids := newGitIdentities(t)
	skill := gitRepo(t, map[string]string{
		"SKILL.md": "---\nname: guide\ndescription: Profile capability fixture\n---\nRead https://api.example.net/guide and https://docs.example.net/guide\n",
	}, "v1.0.0")
	operand := ids.serve(skill, "https://source.example.org/skills/guide")
	profile := t.TempDir()
	request := strings.Replace(profileSkillRequest(`{"network":["api.example.net","docs.example.net"]}`), "%s", `"`+operand+`"`, 1)
	writeSkillProfile(t, profile, "full-profile", request, 4)
	info, _, _, err := Install(home, InstallOptions{Operand: profile})
	if err != nil {
		t.Fatalf("complete profile declaration refused: %v", err)
	}
	if info.Lock.SchemaVersion != contextlock.SkillIntegritySchemaVersion || len(info.Lock.SkillIntegrity) != 1 {
		t.Fatalf("profile declaration did not create lock-6 evidence: %+v", info.Lock)
	}
	var synthetic map[string]any
	if err := json.Unmarshal(info.Lock.SkillIntegrity[0].Manifest.Declaration, &synthetic); err != nil {
		t.Fatal(err)
	}
	declared, ok := synthetic["capabilities"].(map[string]any)
	if !ok || len(declared["network"].([]any)) != 2 {
		t.Fatalf("profile capabilities missing from frozen declaration: %#v", synthetic)
	}
	if err := os.Remove(filepath.Join(profile, "agent-context.json")); err != nil {
		t.Fatalf("remove current profile source for frozen-evidence control: %v", err)
	}
	checks, err := startupChecks(home, info.Lock)
	if err != nil {
		t.Fatalf("frozen profile evidence did not verify: %v", err)
	}
	verified := 0
	for _, check := range checks {
		if check.Code == "skill_integrity_verified" {
			if check.Status != "passed" {
				t.Fatalf("frozen profile evidence did not pass: %+v", check)
			}
			verified++
		}
	}
	if verified != 1 {
		t.Fatalf("startup did not consume frozen evidence: %+v", checks)
	}
}

// TestProfileWithoutCapabilitiesKeepsLegacyInstall proves a schema-3 package
// with no profile capability field still installs a manifestless skill using
// the original synthesized declaration and lock-6 evidence.
func TestProfileWithoutCapabilitiesKeepsLegacyInstall(t *testing.T) {
	home := gitFixtureHome(t)
	pinHomes(t)
	ids := newGitIdentities(t)
	skill := gitRepo(t, map[string]string{
		"SKILL.md": "---\nname: guide\ndescription: Legacy synthetic fixture\n---\nRead this guide.\n",
	}, "v1.0.0")
	operand := ids.serve(skill, "https://source.example.org/skills/guide")
	profile := t.TempDir()
	request := strings.Replace(profileSkillRequest(""), "%s", `"`+operand+`"`, 1)
	writeSkillProfile(t, profile, "legacy-profile", request, 3)
	info, _, _, err := Install(home, InstallOptions{Operand: profile})
	if err != nil {
		t.Fatalf("schema-3 package without capabilities stopped installing: %v", err)
	}
	if info.Lock.SchemaVersion != contextlock.SkillIntegritySchemaVersion || len(info.Lock.SkillIntegrity) != 1 {
		t.Fatalf("legacy install lock evidence changed: %+v", info.Lock)
	}
	entry := info.Lock.SkillIntegrity[0]
	if entry.Manifest.ManifestKind != "synthesized" || string(entry.Manifest.Declaration) != `{"schema_version":3,"runtime_roots":[],"commands":{},"capabilities":{"network":"none","filesystem":"none","exec":"none","secrets":"none","env_read":"none"}}` {
		t.Fatalf("legacy synthetic declaration changed: kind=%q declaration=%s", entry.Manifest.ManifestKind, entry.Manifest.Declaration)
	}
}

// TestMalformedProfileCapabilitiesDoNotPublishPartialState proves malformed
// declarations fail before publishing either the installed source or its lock.
func TestMalformedProfileCapabilitiesDoNotPublishPartialState(t *testing.T) {
	for _, raw := range []string{`{"unknown":[]}`, `{"network":"all"}`, `{"network":["api.example.net","api.example.net"]}`} {
		t.Run(raw, func(t *testing.T) {
			home := gitFixtureHome(t)
			pinHomes(t)
			ids := newGitIdentities(t)
			skill := gitRepo(t, map[string]string{
				"SKILL.md": "---\nname: guide\ndescription: Malformed profile capability fixture\n---\nRead this guide.\n",
			}, "v1.0.0")
			operand := ids.serve(skill, "https://source.example.org/skills/guide")

			prior := t.TempDir()
			priorRequest := strings.Replace(profileSkillRequest(""), "%s", `"`+operand+`"`, 1)
			writeSkillProfile(t, prior, "malformed-profile", priorRequest, 3)
			if _, _, _, err := Install(home, InstallOptions{Operand: prior}); err != nil {
				t.Fatalf("valid prior profile failed to install: %v", err)
			}
			priorSource, err := os.ReadFile(sourcePath(home, "malformed-profile"))
			if err != nil {
				t.Fatal(err)
			}
			priorLock, err := os.ReadFile(lockPath(home, "malformed-profile"))
			if err != nil {
				t.Fatal(err)
			}

			malformed := t.TempDir()
			malformedRequest := strings.Replace(profileSkillRequest(raw), "%s", `"`+operand+`"`, 1)
			writeSkillProfile(t, malformed, "malformed-profile", malformedRequest, 4)
			if _, _, _, err := Install(home, InstallOptions{Operand: malformed}); err == nil || !strings.Contains(err.Error(), "profile_skill_capability_invalid") {
				t.Fatalf("malformed declaration was not refused with its diagnostic: %v", err)
			}
			for path, before := range map[string][]byte{
				sourcePath(home, "malformed-profile"): priorSource,
				lockPath(home, "malformed-profile"):   priorLock,
			} {
				after, err := os.ReadFile(path)
				if err != nil || string(after) != string(before) {
					t.Fatalf("malformed declaration changed prior %s: %v", filepath.Base(path), err)
				}
			}
		})
	}
}

// TestManifestlessProfileWithoutCapabilitiesStillRejectsNetwork proves the
// schema-3 no-field path retains the existing strict uncovered-host refusal.
func TestManifestlessProfileWithoutCapabilitiesStillRejectsNetwork(t *testing.T) {
	home := gitFixtureHome(t)
	pinHomes(t)
	ids := newGitIdentities(t)
	skill := gitRepo(t, map[string]string{
		"SKILL.md": "---\nname: guide\ndescription: No declaration fixture\n---\nRead https://api.example.net/guide\n",
	}, "v1.0.0")
	operand := ids.serve(skill, "https://source.example.org/skills/guide")
	profile := t.TempDir()
	request := strings.Replace(profileSkillRequest(""), "%s", `"`+operand+`"`, 1)
	writeSkillProfile(t, profile, "no-capabilities-profile", request, 3)
	if _, _, _, err := Install(home, InstallOptions{Operand: profile}); err == nil ||
		!strings.Contains(err.Error(), "audit.capability.network-undeclared") || !strings.Contains(err.Error(), "api.example.net") {
		t.Fatalf("schema-3 profile without capabilities hid an uncovered host: %v", err)
	}
	for _, path := range []string{sourcePath(home, "no-capabilities-profile"), lockPath(home, "no-capabilities-profile")} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("uncovered-host refusal published %s: %v", filepath.Base(path), err)
		}
	}
}

// TestProfileCapabilityGlobMatchesNativeSemantics proves profile host globs
// admit covered hosts and leave an uncovered host as a strict audit refusal.
func TestProfileCapabilityGlobMatchesNativeSemantics(t *testing.T) {
	home := gitFixtureHome(t)
	pinHomes(t)
	ids := newGitIdentities(t)
	skill := gitRepo(t, map[string]string{
		"SKILL.md": "---\nname: guide\ndescription: Glob fixture\n---\nRead https://api.example.org/guide\n",
	}, "v1.0.0")
	operand := ids.serve(skill, "https://source.example.org/skills/guide")

	covered := t.TempDir()
	coveredRequest := strings.Replace(profileSkillRequest(`{"network":["*.example.org"]}`), "%s", `"`+operand+`"`, 1)
	writeSkillProfile(t, covered, "covered-profile", coveredRequest, 4)
	if _, _, _, err := Install(home, InstallOptions{Operand: covered}); err != nil {
		t.Fatalf("native-compatible host glob refused a covered host: %v", err)
	}

	uncovered := t.TempDir()
	uncoveredRequest := strings.Replace(profileSkillRequest(`{"network":["*.example.net"]}`), "%s", `"`+operand+`"`, 1)
	writeSkillProfile(t, uncovered, "uncovered-profile", uncoveredRequest, 4)
	if _, _, _, err := Install(home, InstallOptions{Operand: uncovered}); err == nil ||
		!strings.Contains(err.Error(), "audit.capability.network-undeclared") || !strings.Contains(err.Error(), "api.example.org") {
		t.Fatalf("profile host glob admitted an uncovered host: %v", err)
	}
}

// TestProfileCapabilityPartialAllowlistDoesNotReuseCachedAllowance proves a
// broad cached audit cannot hide a newly uncovered host or publish partial state.
func TestProfileCapabilityPartialAllowlistDoesNotReuseCachedAllowance(t *testing.T) {
	home := gitFixtureHome(t)
	pinHomes(t)
	ids := newGitIdentities(t)
	skill := gitRepo(t, map[string]string{
		"SKILL.md": "---\nname: guide\ndescription: Cache capability fixture\n---\nRead https://api.example.net/guide and https://docs.example.net/guide\n",
	}, "v1.0.0")
	operand := ids.serve(skill, "https://source.example.org/skills/guide")

	full := t.TempDir()
	fullRequest := strings.Replace(profileSkillRequest(`{"network":["api.example.net","docs.example.net"]}`), "%s", `"`+operand+`"`, 1)
	writeSkillProfile(t, full, "full-profile", fullRequest, 4)
	if _, _, _, err := Install(home, InstallOptions{Operand: full}); err != nil {
		t.Fatalf("full allowlist positive control refused: %v", err)
	}
	fullSource, err := os.ReadFile(sourcePath(home, "full-profile"))
	if err != nil {
		t.Fatal(err)
	}
	fullLock, err := os.ReadFile(lockPath(home, "full-profile"))
	if err != nil {
		t.Fatal(err)
	}

	narrow := t.TempDir()
	narrowRequest := strings.Replace(profileSkillRequest(`{"network":["api.example.net"]}`), "%s", `"`+operand+`"`, 1)
	writeSkillProfile(t, narrow, "narrow-profile", narrowRequest, 4)
	if _, _, _, err := Install(home, InstallOptions{Operand: narrow}); err == nil || !strings.Contains(err.Error(), "audit.capability.network-undeclared") || !strings.Contains(err.Error(), "docs.example.net") {
		t.Fatalf("narrow declaration reused broad allowance: %v", err)
	}
	for _, path := range []string{sourcePath(home, "narrow-profile"), lockPath(home, "narrow-profile")} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("refused narrow install published %s: %v", filepath.Base(path), err)
		}
	}
	for path, before := range map[string][]byte{sourcePath(home, "full-profile"): fullSource, lockPath(home, "full-profile"): fullLock} {
		after, err := os.ReadFile(path)
		if err != nil || string(after) != string(before) {
			t.Fatalf("narrow refusal changed prior %s: %v", filepath.Base(path), err)
		}
	}
}

// TestProfileCapabilitiesDoNotReplaceNativeManifest proves profile caps refuse
// for every supported native manifest, including an empty profile declaration,
// while the unchanged native path remains valid.
func TestProfileCapabilitiesDoNotReplaceNativeManifest(t *testing.T) {
	for _, native := range []struct {
		name, path, declaration string
	}{
		{name: "canonical", path: "agent-skill.json", declaration: `{"schema_version":3,"capabilities":{}}`},
		{name: "legacy", path: "csk-skill.json", declaration: `{"schema_version":3,"capabilities":{}}`},
		{name: "runtime fallback", path: "agents/runtime.json", declaration: `{"commands":{}}`},
	} {
		t.Run(native.name, func(t *testing.T) {
			home := gitFixtureHome(t)
			pinHomes(t)
			ids := newGitIdentities(t)
			skill := gitRepo(t, map[string]string{
				"SKILL.md":  "---\nname: guide\ndescription: Native precedence fixture\n---\nLocal content.\n",
				native.path: native.declaration,
			}, "v1.0.0")
			operand := ids.serve(skill, "https://source.example.org/skills/guide")

			for _, raw := range []string{`{"network":["api.example.net"]}`, `{}`} {
				profile := t.TempDir()
				withCaps := strings.Replace(profileSkillRequest(raw), "%s", `"`+operand+`"`, 1)
				writeSkillProfile(t, profile, "native-profile", withCaps, 4)
				if _, _, _, err := Install(home, InstallOptions{Operand: profile}); err == nil || !strings.Contains(err.Error(), "profile_skill_capability_native_manifest") {
					t.Fatalf("profile declaration %s replaced %s: %v", raw, native.path, err)
				}
				for _, path := range []string{sourcePath(home, "native-profile"), lockPath(home, "native-profile")} {
					if _, err := os.Stat(path); !os.IsNotExist(err) {
						t.Fatalf("native precedence refusal published %s: %v", filepath.Base(path), err)
					}
				}
			}

			profileWithout := t.TempDir()
			withoutCaps := strings.Replace(profileSkillRequest(""), "%s", `"`+operand+`"`, 1)
			writeSkillProfile(t, profileWithout, "native-profile", withoutCaps, 3)
			if _, _, _, err := Install(home, InstallOptions{Operand: profileWithout}); err != nil {
				t.Fatalf("%s without profile declaration stopped working: %v", native.path, err)
			}
		})
	}
}

// TestSharedProfileSkillCapabilitiesAgree exercises two independent context
// edges: different or omitted declarations refuse; semantic equality installs.
func TestSharedProfileSkillCapabilitiesAgree(t *testing.T) {
	for _, tc := range []struct {
		name, leftCaps, rightCaps, want string
	}{
		{name: "different hosts", leftCaps: `{"network":["api.example.net"]}`, rightCaps: `{"network":["docs.example.net"]}`, want: "profile_skill_capability_conflict"},
		{name: "one omitted", leftCaps: `{"network":["api.example.net"]}`, rightCaps: "", want: "profile_skill_capability_conflict"},
		{name: "equal reordered", leftCaps: `{"network":["api.example.net","docs.example.net"]}`, rightCaps: `{"network":["docs.example.net","api.example.net"]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home := gitFixtureHome(t)
			pinHomes(t)
			ids := newGitIdentities(t)
			skill := gitRepo(t, map[string]string{"SKILL.md": "---\nname: guide\ndescription: Shared edge fixture\n---\nShared skill.\n"}, "v1.0.0")
			skillOperand := ids.serve(skill, "https://source.example.org/skills/guide")
			left := gitRepo(t, map[string]string{"agent-context.json": contextRequiringSkill(t, "left", skillOperand, tc.leftCaps)}, "v1.0.0")
			right := gitRepo(t, map[string]string{"agent-context.json": contextRequiringSkill(t, "right", skillOperand, tc.rightCaps)}, "v1.0.0")
			leftOperand := ids.serve(left, "https://source.example.org/contexts/left")
			rightOperand := ids.serve(right, "https://source.example.org/contexts/right")
			profile := t.TempDir()
			rootManifest := `{"schema_version":3,"name":"shared-profile","version":"1.0.0","requires":{"contexts":{"left":{"git":"` + leftOperand + `","range":"*"},"right":{"git":"` + rightOperand + `","range":"*"}}}}`
			writeManifestPackage(t, profile, rootManifest, nil)
			_, _, _, err := Install(home, InstallOptions{Operand: profile})
			if tc.want != "" {
				if err == nil || !strings.Contains(err.Error(), tc.want) {
					t.Fatalf("shared conflict accepted: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("equal shared declarations refused: %v", err)
			}
		})
	}
}

func contextRequiringSkill(t *testing.T, name, skillOperand, caps string) string {
	t.Helper()
	request := strings.Replace(profileSkillRequest(caps), "%s", `"`+skillOperand+`"`, 1)
	return `{"schema_version":4,"name":"` + name + `","version":"1.0.0","requires":{"skills":{"guide":` + request + `}}}`
}
