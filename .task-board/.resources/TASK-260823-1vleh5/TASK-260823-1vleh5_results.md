# TASK-260823-1vleh5 — schema-8 manifest parsing, first-party module roots

**Status:** ready for review (board `to-review`).
**Branch:** `task/TASK-260823-1vleh5-schema8-module-roots` — PR
[relux-works/curator#33](https://github.com/relux-works/curator/pull/33), branched from
`origin/main@e17b0f1`.
**Candidate suite:** `relux-works/curator-spec@6001dc3` (`candidate/schema-8-rc.9`),
`protocol/core.md` §4.2.3, `profiles/manager.md` module-root diagnostics table,
`schemas/v1/agent-skill-v8.schema.json`.

## What was implemented

### `internal/moduleroots` (new package)

Owns both halves of §4.2.3, deliberately split on the fixed `go list` because that
is where the section's failure boundary puts them.

| Function | Runs | Diagnostics it can raise |
| --- | --- | --- |
| `ValidateDeclaration` | before `go list` | `build_module_root_declaration_invalid`, `build_module_root_containment_invalid` |
| `EffectiveReplaceSet` | after `go list` | `build_module_root_directive_form_unsupported` |
| `ValidateBijection` | after `go list`, before `go build` | `build_module_root_directive_form_unsupported`, `build_module_root_directive_undeclared`, `build_module_root_declaration_unused` |

Declaration and containment: portable relative path other than `.`; a real,
link-free directory strictly inside the snapshot (every package-controlled
component inspected with `Lstat`, so no link redirects the check outward);
`go.mod` directly inside as a real regular file; unique and pairwise disjoint;
disjoint from every declared build root and every runtime root. Disjointness is
checked twice — on the exact protocol paths and on a platform-folded key
(`NFD ∘ fold ∘ NFD`, the same conservative model `internal/buildsource` already
uses for platform path collisions) — so two paths differing only by case are
rejected even where only one of them can exist on the host.

Effective replace set: read from `vendor/modules.txt` bytes only. Only a line
whose first two bytes are exactly `# ` and which contains exactly ` => ` is an
annotation; the split is at the first separator; both sides are
whitespace-tokenised and must carry one or two tokens. One-token-left
annotations are the effective directives. A two-token-left annotation must
match a one-token-left annotation exactly (same module path, same right side)
or it is a versioned-left directive and is rejected — this is what enforces the
no-version-on-the-left rule *without parsing `go.mod`*. A two-token right side
is a module-to-module redirect and is rejected. `go.mod` is never parsed and
`Module.Replace.Dir`/`Module.Replace.GoMod` are never accepted as evidence that
a path exists.

Bijection: each target is resolved against the build root; a target that leaves
the snapshot names no declared directory and is reported undeclared, exactly as
an in-snapshot target that was never declared is. Absolute and drive-qualified
targets are a form failure. Two directives resolving onto one declaration is
rejected — the second names no *distinct* declaration, and `undeclared` is the
only code in the closed set that fits; the reasoning is stated at the call site.

### `internal/skillspec`

- `SupportedSchemaVersions` gains 8.
- `modules` accepted only on a local `go-v1` build command. Absent and empty
  both mean the schema-6/7 single-module build root; `null` is neither.
- `execution_policy` + `interpreter` accepted only on a script command, bound in
  both directions (both or neither), against the single closed constant
  `script-worker-v1` and the closed set `{node-v1, python3-v1}`. Absence stays
  the one spelling of declared-only, so `null`, `"none"`, `manager-worker-v1`,
  `hardened-worker-v1` and `script-worker-v2` are all rejected.
- **Declared-only behaviour**: both script fields are parsed and carried on
  `Command`; no containment is derived, claimed or applied. Worker containment
  is decision-0008 work outside this task.
- Schemas 2–7 reject all three fields as unknown. Schema 1 keeps its deployed
  extension tolerance and **ignores** them — it never reads an unknown command
  field as an enforcement claim. Regression-tested.
- Parse time runs the declaration half only, per the failure boundary.

### `internal/marker` — install marker v4

Marker v4 is marker v3 with `schema_version` 4 and `skill_schema_version` 8 and
no other difference, so every marker-v3 build-record rule applies unchanged.
This was not optional: without it `Write` fails closed for any schema-8
installation (`skill_schema_version 8` is outside every existing band), so
admitting schema 8 in the parser alone would have landed a manifest curator can
read but cannot install. `internal/install` now selects the explicit
receipt-version/execution-policy build record for schema `>= 7` rather than
exactly 7.

### `.github/ci/root-artifacts.tsv`

`internal/moduleroots` → `vectors/module-roots.json`; `internal/skillspec` also
declares `schema-cases/agent-skill-v8`; `internal/marker` also declares
`schema-cases/install-marker-v4`. A root that does not publish them defers the
package (root unset) instead of going red on a missing file, which is the
table's whole contract.

## Suite consumption

| Family | Consumer | Result |
| --- | --- | --- |
| `schema-cases/agent-skill-v8` (67 cases) | `skillspec.TestReleasedSchemaCases` | all pass |
| `schema-cases/csk-skill-v8` (67 cases) | same | all pass |
| `schema-cases/agent-skill-v7`, `csk-skill-v7` | same (unchanged) | all pass |
| `schema-cases/install-marker-v4` (27 cases) | `marker.TestReadAuthoritativeMarkerV4SchemaCases` | 22 pass, 5 explicitly listed (see finding) |
| `vectors/module-roots.json` (10 cases) | `moduleroots.TestModuleRootVectors` | all 10 pass |

`TestReleasedSchemaCases` fails if a named family directory is absent or empty,
so the coverage cannot silently narrow. `TestModuleRootVectors` asserts each
vector against the *correct half* — a `fails_before: go-list` vector must be
rejected by `ValidateDeclaration` with the replace set never read, and a
`fails_before: go-build` vector must be admitted by the declaration half and
rejected by the replace half — so it proves the failure boundary, not just the
diagnostic. It hard-fails on a vector carrying `link_paths`, which it does not
materialise, rather than passing over it.

## Verification

Every command run standalone, unpiped, real exit code reported.

| Gate | Exit |
| --- | ---: |
| `gofmt -l cmd internal` | empty |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `golangci-lint run` (v2.12.2, CI's pin) | 0 — "0 issues." |
| `go test ./internal/... -count=1 -timeout 30m` | 0 |
| `go test ./cmd/... -count=1 -timeout 30m` | 0 (`cmd/curator` 276.287s) |
| `bash .github/ci/gate-selftest.sh` | 0 — 81 passed, 0 failed |
| `bash .github/ci/suite-plan.sh <candidate root>` | 0 — served=42 deferred=0 excluded=0 |
| `CURATOR_CONFORMANCE_ROOT=<6001dc3>/conformance/v1 CI_REQUIRE_FULL_ROOT=1 bash .github/ci/test-gate.sh` | 0 — go test exit=0, platform-case gate exit=0 |

The gate recorded 10 skips, all pre-existing host-capability/platform-control/
opt-in entries already in `skip-classes.tsv`. No new skip class was introduced;
the two new symlink tests use the existing `this host cannot create` reason and
ran on this host.

`GOOS=windows`/`GOOS=linux` native execution was not run from this session; the
three-OS matrix runs on the PR.

## Finding: pre-existing marker cross-field validation gap

Five `install-marker-v4` cases are accepted by `marker.Read` although the suite
marks them invalid:

- `invalid-external-declared-effective-mismatch.json`
- `invalid-marker-local-identity-kind-mismatch.json`
- `invalid-marker-network-identity-kind-mismatch.json`
- `invalid-marker-sha1-effective-revision-width.json`
- `invalid-marker-sha256-effective-revision-width.json`

All five are marker-**v3** external-repository cross-field rules `validV3Build`
has never implemented: declared/effective identity equality when unsubstituted,
effective identity kind against substitution type, and substitution revision
width against object format. Verified by running the identical check against
`schema-cases/install-marker-v3`, where the same five fail — no test consumes
that family today, which is why the gap has been invisible. Marker v4 inherits
it exactly; schema 8 neither introduces nor widens it.

Not fixed here: it is schema-7 external-repository validation, a different
surface, and tightening `Read` changes acceptance of already-installed markers
(a rejected marker reads as "not installed"), which needs its own change and
review. Encoded as `markerV4CasesThisReaderDoesNotModel`, asserted in both
directions — a case that starts being rejected fails the test and tells you to
delete its entry — so the allowance cannot go stale. Suggest routing to
`TASK-260823-2u5xov` or a dedicated bug.

## Second finding: `cmd/curator` treats every marker above v2 as unsupported

`cmd/curator/builds.go:365` compares `recorded.SchemaVersion != marker.SchemaVersion`
under a comment stating the intent as "a marker below schema 2 cannot describe a
compiled command". That `!=` was correct while 2 was the newest marker; since
marker v3 landed it also catches markers *above* 2, and `markerRefusal` reports
`stateUnsupportedMarker` for them. `status_test.go:605` pins this by tampering
`schema_version` to `marker.SchemaVersion + 1` and expecting exactly that.
Marker v4 inherits the same treatment; schema 8 introduces no new class of
breakage here. Untouched deliberately — it is the schema-7 external-repository
status surface and needs its own analysis.

## Handoff to TASK-260823-1wvgw8

`EffectiveReplaceSet` and `ValidateBijection` are ready to wire into the driver.
The driver still needs to: read `<build root>/vendor/modules.txt` (diagnosing an
unreadable one as `vendor_metadata_inconsistent`, as `graph.go:159` already
does), replace `validateModule`'s unconditional `item.Module.Replace != nil`
rejection at `internal/godriver/graph.go:306` with admission of exactly the
bijected set, extend the directive/cgo/assembly scan surface to the declared
directories and their vendor copies, and withhold the audited-vendor allowance
from any module carrying a replacement.
