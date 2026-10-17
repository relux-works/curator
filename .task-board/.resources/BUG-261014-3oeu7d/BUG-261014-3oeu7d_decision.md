# BUG-261014-3oeu7d — migration capability audit: decision required

Status: blocked on an audit API/scope decision. This is not a review handoff.

## Normative evidence

Read curator-spec main directly and verified its document bytes against commit
1059b94b4135089d842bf017070b8ab47d79a536:
https://github.com/relux-works/curator-spec/blob/1059b94b4135089d842bf017070b8ab47d79a536/protocol/environments.md

Section 9.1 requires a skill's own package manifest capabilities. Section 9.4
requires the network-undeclared finding, skill, host, file and network declaration
remedy even for a missing manifest or declaration. Other audit gates remain
unchanged. No spec change was made.

## Concrete constraint

The brief additionally requires the audit subject schema to match the skill's
actual manifest schema and confines implementation to internal/envprofile.
The existing audit.Gate -> decideWithPins in internal/audit/audit.go checks
SchemaVersion < 3 before applying the findings verdict. Passing the actual 0
(no manifest) or 2 therefore returns require_pin, with no host/file network
finding. Keeping the candidate's hardcoded schema 3 avoids this but violates
the brief's actual-schema requirement. It also changes the meaning of the audit
subject. Unconditionally forwarding schema 0 would additionally refuse clean
manifestless legacy skills previously admitted by profile audit.

An executable probe varied only subject schema with identical bytes, capability
set, hash version and strict policy. Schema 0 and 2 returned require_pin; schema
3 returned the network-undeclared finding. Full probe source and output are
attached below. This is an API ownership/policy distinction, not a missing test.

## Attempts and disposition

Reviewed the supplied two-file candidate critically; did not apply it verbatim.
Added a separate regression suite, retaining every existing test under its
original name, including TestMigratedGlobalSkillNetworkUndeclaredFollowUp.
A provisional implementation loaded skillspec only for skill members, refused
manifest errors, forwarded modern schema versions, retained the old floor for
legacy subjects, and added the required migration remedy. Targeted tests passed,
but this explicitly does not prove actual schema fidelity for legacy subjects.
The provisional product change was removed rather than shipping that mismatch.
Its patch is attached as a rejected approach for inspection, not a candidate.
The only repository delta is the uncommitted regression test file.

## Decision and options

Recommended: authorize a small change in internal/audit, outside this leaf's
stated module boundary, to expose the existing profile audit policy explicitly
while carrying the real skill schema. Keep the generic skill-pipeline Gate
policy unchanged; test both production callers. This makes the policy boundary
honest but needs scope/ownership authorization.

Alternative: approve retaining/documenting the existing profile schema floor 3
for pre-capability skills and amend the brief's actual-schema expectation.
Modern manifest schema versions can still be forwarded exactly. This is the
smaller change, but its subject schema is deliberately not the legacy manifest's
schema. No synthesized capability or operator trust pin is proposed.

Exact input needed: choose the audit-module scope extension or the documented
legacy schema-floor exception. No full gate or review handoff is claimed pending
that choice.

## Coverage map and bounds

Brief gap: no formal surface table from the attack-surface catalog was supplied.
The required behavioral controls are nevertheless mapped below. No input rows
are declared out of contract. No completed coverage-map claim is made without
killed narrowing mutants.

4 of 5 AC rows driven by named executable behavioral tests; AC5 (mutation and
full remote gate) is unrun. Driven means exercised, not satisfied: current
production fails the new required-behavior tests.

| AC | Production call site | Named tests / controls |
| --- | --- | --- |
| 1 | EnsureDefaultWithPolicy -> migrateGlobalSkills -> strictAuditMember; Install -> strictAuditMember | TestMigratedGlobalSkillNetworkCapabilities; TestInstalledMemberCapabilitiesStayKindScoped |
| 2 | EnsureDefaultWithPolicy -> ensureDefault -> migrateGlobalSkills -> strictAuditMember | TestMigratedGlobalSkillNetworkCapabilities (exact, glob, alias, schema 9, absent, none, missing network, other host, schema 2); TestMigratedGlobalSkillDeclarationKeepsOtherGates (neighbor, revoked, canary); TestMigratedGlobalSkillManifestErrorsBlock (malformed network/JSON, failed manifest read); every refusal asserts no source or lock publication |
| 3 | migrateGlobalSkills error return | TestMigratedGlobalSkillNetworkCapabilities and neighboring-host case assert skill, host, file, finding and remedy |
| 4 | Install -> strictAuditMember | TestInstalledMemberCapabilitiesStayKindScoped includes a declared skill, a clean non-skill with a malformed skill manifest, and non-skill network use with a covering skill manifest; existing tests unchanged |
| 5 | sh scripts/remote-gate.sh | Unrun pending decision; no accepted candidate |

Unreadable-manifest witness uses a directory at the manifest filename so the
read fails portably; permission-denied behavior is not proven. The diagnostic
probe calls audit.Gate directly, not the migration entry, and proves only the
schema/pin precedence constraint. Existing reviewer tests remain present, but
only the targeted selection was executed; whole-suite green is unverified.
All new files are within internal/envprofile; this is not review rework.

## Mutant evidence

| Mutant | Narrowed gate | Named failing test | Result / bound |
| --- | --- | --- | --- |
| ImplicitNone restoration | Unrun | None | Survivor/unexecuted: no mutation evidence claimed |
| Ignore manifest read/parse error | Unrun | None | Survivor/unexecuted: no mutation evidence claimed |

## Validation ledger

All gate commands ran directly with output redirected to files, without tee.
Source baseline: 26d8ddbf1278ac025a3333aee1053adbdb37f1bf (0 commits behind local main).
Toolchain: go1.25.5 darwin/arm64. No branch switch, producer commit, publication,
or control-root file edit was performed.

| Command | Exit | Evidence / interpretation |
| --- | ---: | --- |
| go test -count=1 -timeout=5m ./internal/envprofile -run 'Test(MigratedGlobalSkill(NetworkCapabilities|DeclarationKeepsOtherGates|ManifestErrorsBlock)|InstalledMemberCapabilitiesStayKindScoped)$' | 1 | baseline-red-01.log; expected-red regression witnesses on unchanged product |
| go test -count=1 -timeout=5m ./internal/envprofile -run 'Test(MigratedGlobalSkill|InstalledMemberCapabilitiesStayKindScoped|StrictAuditMemberRefusesV1NULBeforeHashing)' | 0 | targeted-01.log; provisional schema-floor implementation only, not current candidate |
| go test -count=1 -timeout=1m -v ./internal/envprofile -run '^TestMigrationCapabilitySchemaProbe$' | 0 | schema-probe-01.log; diagnostic distinction proven, not product acceptance |
| repeated baseline regression command after removing provisional product change | 1 | restored-baseline-red-01.log; expected-red, current production still has the bug |
| go test -run '^$' ./internal/envprofile | 0 | compile-01.log; compilation only, no tests executed |
| git diff --check | 0 | diff-check-01.log; does not validate behavior or untracked-file whitespace |
| sh scripts/remote-gate.sh | unrun | No candidate accepted; no hosted result claimed |
| lint / full property suites / full reviewer suites | unrun | No accepted candidate; no green claim |

## Logbook

The supplied candidate's manifest capability loading addresses the field report,
but retaining a fictional schema masks the distinction between generic strict
skill audit and the existing profile policy. The bounded schema probe establishes
that the conflict is real. Product edits were rolled back; regression tests and
this evidence remain for resumption. The board is blocked on the exact policy /
module-scope choice above, not on test availability or operational waiting.

### baseline-red-01.log

```text
--- FAIL: TestMigratedGlobalSkillNetworkCapabilities (2.80s)
    --- FAIL: TestMigratedGlobalSkillNetworkCapabilities/exact-host-declared (0.25s)
        global_migration_capabilities_test.go:42: covering declaration refused: profile_source_invalid: migrated global skill "binary-skill" audit blocked: audit blocked: binary-skill: high audit.capability.network-undeclared - contacts downloads.example.com, which no capabilities.network entry covers (file: scripts/fetch.sh)
    --- FAIL: TestMigratedGlobalSkillNetworkCapabilities/glob-host-declared (0.25s)
        global_migration_capabilities_test.go:42: covering declaration refused: profile_source_invalid: migrated global skill "binary-skill" audit blocked: audit blocked: binary-skill: high audit.capability.network-undeclared - contacts downloads.example.com, which no capabilities.network entry covers (file: scripts/fetch.sh)
    --- FAIL: TestMigratedGlobalSkillNetworkCapabilities/legacy-alias (0.32s)
        global_migration_capabilities_test.go:42: covering declaration refused: profile_source_invalid: migrated global skill "binary-skill" audit blocked: audit blocked: binary-skill: high audit.capability.network-undeclared - contacts downloads.example.com, which no capabilities.network entry covers (file: scripts/fetch.sh)
    --- FAIL: TestMigratedGlobalSkillNetworkCapabilities/schema-nine (0.25s)
        global_migration_capabilities_test.go:42: covering declaration refused: profile_source_invalid: migrated global skill "binary-skill" audit blocked: audit blocked: binary-skill: high audit.capability.network-undeclared - contacts downloads.example.com, which no capabilities.network entry covers (file: scripts/fetch.sh)
    --- FAIL: TestMigratedGlobalSkillNetworkCapabilities/no-manifest (0.44s)
        global_migration_capabilities_test.go:56: migration = profile_source_invalid: migrated global skill "binary-skill" audit blocked: audit blocked: binary-skill: high audit.capability.network-undeclared - contacts downloads.example.com, which no capabilities.network entry covers (file: scripts/fetch.sh), missing "agent-skill.json"
    --- FAIL: TestMigratedGlobalSkillNetworkCapabilities/network-none (0.25s)
        global_migration_capabilities_test.go:56: migration = profile_source_invalid: migrated global skill "binary-skill" audit blocked: audit blocked: binary-skill: high audit.capability.network-undeclared - contacts downloads.example.com, which no capabilities.network entry covers (file: scripts/fetch.sh), missing "agent-skill.json"
    --- FAIL: TestMigratedGlobalSkillNetworkCapabilities/network-absent (0.39s)
        global_migration_capabilities_test.go:56: migration = profile_source_invalid: migrated global skill "binary-skill" audit blocked: audit blocked: binary-skill: high audit.capability.network-undeclared - contacts downloads.example.com, which no capabilities.network entry covers (file: scripts/fetch.sh), missing "agent-skill.json"
    --- FAIL: TestMigratedGlobalSkillNetworkCapabilities/other-host-declared (0.32s)
        global_migration_capabilities_test.go:56: migration = profile_source_invalid: migrated global skill "binary-skill" audit blocked: audit blocked: binary-skill: high audit.capability.network-undeclared - contacts downloads.example.com, which no capabilities.network entry covers (file: scripts/fetch.sh), missing "agent-skill.json"
    --- FAIL: TestMigratedGlobalSkillNetworkCapabilities/pre-capabilities-schema (0.32s)
        global_migration_capabilities_test.go:56: migration = profile_source_invalid: migrated global skill "binary-skill" audit blocked: audit blocked: binary-skill: high audit.capability.network-undeclared - contacts downloads.example.com, which no capabilities.network entry covers (file: scripts/fetch.sh), missing "agent-skill.json"
--- FAIL: TestMigratedGlobalSkillDeclarationKeepsOtherGates (1.01s)
    --- FAIL: TestMigratedGlobalSkillDeclarationKeepsOtherGates/neighbouring-host (0.41s)
        global_migration_capabilities_test.go:98: migration = profile_source_invalid: migrated global skill "binary-skill" audit blocked: audit blocked: binary-skill: high audit.capability.network-undeclared - contacts downloads.example.com, which no capabilities.network entry covers (file: scripts/fetch.sh); audit blocked: binary-skill: high audit.capability.network-undeclared - contacts uploads.example.com, which no capabilities.network entry covers (file: scripts/fetch.sh), missing "agent-skill.json"
--- FAIL: TestMigratedGlobalSkillManifestErrorsBlock (1.35s)
    --- FAIL: TestMigratedGlobalSkillManifestErrorsBlock/malformed-network (0.49s)
        global_migration_capabilities_test.go:118: migration = <nil>, want manifest refusal naming "capabilities.network"
    --- FAIL: TestMigratedGlobalSkillManifestErrorsBlock/malformed-json (0.48s)
        global_migration_capabilities_test.go:118: migration = <nil>, want manifest refusal naming "agent-skill.json"
    --- FAIL: TestMigratedGlobalSkillManifestErrorsBlock/unreadable-manifest (0.39s)
        global_migration_capabilities_test.go:118: migration = <nil>, want manifest refusal naming "agent-skill.json"
--- FAIL: TestInstalledMemberCapabilitiesStayKindScoped (2.02s)
    --- FAIL: TestInstalledMemberCapabilitiesStayKindScoped/declared-skill (0.72s)
        global_migration_capabilities_test.go:157: profile_source_invalid: member skill:declared audit blocked: audit blocked: declared: high audit.capability.network-undeclared - contacts downloads.example.com, which no capabilities.network entry covers (file: scripts/fetch.sh)
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	7.811s
FAIL
```

### targeted-01.log

```text
ok  	github.com/relux-works/curator/internal/envprofile	14.304s
```

### schema-probe-01.log

```text
=== RUN   TestMigrationCapabilitySchemaProbe
    migcap_schema_probe_test.go:23: schema=0 diagnostic=audit requires pin: probe-skill: schema v0 has no capabilities; migrate to agent-skill.json schema v3 or pin the content hash sha256:3a9a3d3c57e33486d838d377889a0d310b8d38d058dd3f36f61b1f71afbb73e4 with a reason
    migcap_schema_probe_test.go:23: schema=2 diagnostic=audit requires pin: probe-skill: schema v2 has no capabilities; migrate to agent-skill.json schema v3 or pin the content hash sha256:3a9a3d3c57e33486d838d377889a0d310b8d38d058dd3f36f61b1f71afbb73e4 with a reason
    migcap_schema_probe_test.go:23: schema=3 diagnostic=audit blocked: probe-skill: high audit.capability.network-undeclared - contacts downloads.example.com, which no capabilities.network entry covers (file: scripts/fetch.sh)
--- PASS: TestMigrationCapabilitySchemaProbe (0.02s)
PASS
ok  	github.com/relux-works/curator/internal/envprofile	0.432s
```

### restored-baseline-red-01.log

```text
--- FAIL: TestMigratedGlobalSkillNetworkCapabilities (3.53s)
    --- FAIL: TestMigratedGlobalSkillNetworkCapabilities/exact-host-declared (0.39s)
        global_migration_capabilities_test.go:42: covering declaration refused: profile_source_invalid: migrated global skill "binary-skill" audit blocked: audit blocked: binary-skill: high audit.capability.network-undeclared - contacts downloads.example.com, which no capabilities.network entry covers (file: scripts/fetch.sh)
    --- FAIL: TestMigratedGlobalSkillNetworkCapabilities/glob-host-declared (0.28s)
        global_migration_capabilities_test.go:42: covering declaration refused: profile_source_invalid: migrated global skill "binary-skill" audit blocked: audit blocked: binary-skill: high audit.capability.network-undeclared - contacts downloads.example.com, which no capabilities.network entry covers (file: scripts/fetch.sh)
    --- FAIL: TestMigratedGlobalSkillNetworkCapabilities/legacy-alias (0.53s)
        global_migration_capabilities_test.go:42: covering declaration refused: profile_source_invalid: migrated global skill "binary-skill" audit blocked: audit blocked: binary-skill: high audit.capability.network-undeclared - contacts downloads.example.com, which no capabilities.network entry covers (file: scripts/fetch.sh)
    --- FAIL: TestMigratedGlobalSkillNetworkCapabilities/schema-nine (0.35s)
        global_migration_capabilities_test.go:42: covering declaration refused: profile_source_invalid: migrated global skill "binary-skill" audit blocked: audit blocked: binary-skill: high audit.capability.network-undeclared - contacts downloads.example.com, which no capabilities.network entry covers (file: scripts/fetch.sh)
    --- FAIL: TestMigratedGlobalSkillNetworkCapabilities/no-manifest (0.25s)
        global_migration_capabilities_test.go:56: migration = profile_source_invalid: migrated global skill "binary-skill" audit blocked: audit blocked: binary-skill: high audit.capability.network-undeclared - contacts downloads.example.com, which no capabilities.network entry covers (file: scripts/fetch.sh), missing "agent-skill.json"
    --- FAIL: TestMigratedGlobalSkillNetworkCapabilities/network-none (0.39s)
        global_migration_capabilities_test.go:56: migration = profile_source_invalid: migrated global skill "binary-skill" audit blocked: audit blocked: binary-skill: high audit.capability.network-undeclared - contacts downloads.example.com, which no capabilities.network entry covers (file: scripts/fetch.sh), missing "agent-skill.json"
    --- FAIL: TestMigratedGlobalSkillNetworkCapabilities/network-absent (0.58s)
        global_migration_capabilities_test.go:56: migration = profile_source_invalid: migrated global skill "binary-skill" audit blocked: audit blocked: binary-skill: high audit.capability.network-undeclared - contacts downloads.example.com, which no capabilities.network entry covers (file: scripts/fetch.sh), missing "agent-skill.json"
    --- FAIL: TestMigratedGlobalSkillNetworkCapabilities/other-host-declared (0.40s)
        global_migration_capabilities_test.go:56: migration = profile_source_invalid: migrated global skill "binary-skill" audit blocked: audit blocked: binary-skill: high audit.capability.network-undeclared - contacts downloads.example.com, which no capabilities.network entry covers (file: scripts/fetch.sh), missing "agent-skill.json"
    --- FAIL: TestMigratedGlobalSkillNetworkCapabilities/pre-capabilities-schema (0.36s)
        global_migration_capabilities_test.go:56: migration = profile_source_invalid: migrated global skill "binary-skill" audit blocked: audit blocked: binary-skill: high audit.capability.network-undeclared - contacts downloads.example.com, which no capabilities.network entry covers (file: scripts/fetch.sh), missing "agent-skill.json"
--- FAIL: TestMigratedGlobalSkillDeclarationKeepsOtherGates (5.68s)
    --- FAIL: TestMigratedGlobalSkillDeclarationKeepsOtherGates/neighbouring-host (0.33s)
        global_migration_capabilities_test.go:98: migration = profile_source_invalid: migrated global skill "binary-skill" audit blocked: audit blocked: binary-skill: high audit.capability.network-undeclared - contacts downloads.example.com, which no capabilities.network entry covers (file: scripts/fetch.sh); audit blocked: binary-skill: high audit.capability.network-undeclared - contacts uploads.example.com, which no capabilities.network entry covers (file: scripts/fetch.sh), missing "agent-skill.json"
--- FAIL: TestMigratedGlobalSkillManifestErrorsBlock (10.45s)
    --- FAIL: TestMigratedGlobalSkillManifestErrorsBlock/malformed-network (3.34s)
        global_migration_capabilities_test.go:118: migration = <nil>, want manifest refusal naming "capabilities.network"
    --- FAIL: TestMigratedGlobalSkillManifestErrorsBlock/malformed-json (3.85s)
        global_migration_capabilities_test.go:118: migration = <nil>, want manifest refusal naming "agent-skill.json"
    --- FAIL: TestMigratedGlobalSkillManifestErrorsBlock/unreadable-manifest (3.26s)
        global_migration_capabilities_test.go:118: migration = <nil>, want manifest refusal naming "agent-skill.json"
--- FAIL: TestInstalledMemberCapabilitiesStayKindScoped (11.18s)
    --- FAIL: TestInstalledMemberCapabilitiesStayKindScoped/declared-skill (4.64s)
        global_migration_capabilities_test.go:157: profile_source_invalid: member skill:declared audit blocked: audit blocked: declared: high audit.capability.network-undeclared - contacts downloads.example.com, which no capabilities.network entry covers (file: scripts/fetch.sh)
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	31.111s
FAIL
```

### compile-01.log

```text
ok  	github.com/relux-works/curator/internal/envprofile	0.511s [no tests to run]
```

### Diagnostic probe source

```go
package envprofile

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/audit"
	"github.com/relux-works/curator/internal/capabilities"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/hashing"
)

// Diagnostic probe: identical snapshots and policy, changing only the subject's
// schema, isolate the pre-capability pin rule from the network-declaration rule.
func TestMigrationCapabilitySchemaProbe(t *testing.T) {
	for _, schema := range []int{0, 2, 3} {
		root := t.TempDir()
		writeStoreTestFiles(t, root, map[string][]byte{"scripts/fetch.sh": []byte("curl https://downloads.example.com/a\n")})
		cfg := &config.Config{Path: filepath.Join(t.TempDir(), "config.json"), Audit: config.Audit{Enabled: true, Mode: "strict", FailOn: "high", Backend: "null"}}
		_, errs := audit.Gate(cfg, []audit.Subject{{Name: "probe-skill", Snapshot: root, SchemaVersion: schema, Capabilities: capabilities.ImplicitNone(), HashVersion: hashing.VersionV2}})
		diagnostic := strings.Join(errs, "; ")
		t.Logf("schema=%d diagnostic=%s", schema, diagnostic)
		want := "audit requires pin"
		if schema == 3 {
			want = "audit.capability.network-undeclared"
		}
		if !strings.Contains(diagnostic, want) {
			t.Fatalf("missing %s", want)
		}
	}
}
```
