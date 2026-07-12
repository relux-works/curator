# Curator v0.1 conformance review

## Scope

The review compares the current main branch with docs/implementation-plan.md, Spec §17.1, the normative sections referenced there, and observable behavior of the independent protocol implementation.

## Baseline

- `go test ./... -count=1` passes on macOS.
- Combined statement coverage is 71.5 percent.
- `internal/shell` and `internal/skillcheck` have zero direct package coverage.
- Recent commits contain valid SSH signatures for Ivan Oparin <oparin@me.com>.

## Findings

### F-001: phase 8 does not validate every closure node

Severity: high

Spec §4.1 and Spec §8.1 phase 8 require every skill to pass the same validation as `skill check`. The project installer relies on manifest parsing and defers the `SKILL.md` check to context materialization. A runtime-only provider never materializes context, so a package without the mandatory `SKILL.md` can install successfully.

Required correction: run skill validation for every closure node before collision, dependency, audit, registry, dry-run, or materialization phases. Add a runtime-only regression test.

### F-002: marker wire values are not object-compatible for empty fields

Severity: high

Spec §8.5 requires interoperable marker objects. Mandatory empty list fields can serialize as `null`, and an unselected locale is omitted. The independent implementation emits empty arrays and a present `locale: null`. This breaks parsed-object golden equality even though both readers tolerate part of the difference.

Required correction: normalize mandatory arrays and emit a nullable locale. Add a normalized marker golden fixture.

### F-003: documented CLI argument order is rejected

Severity: high

The standard Go flag parser stops at the first positional argument. Documented forms such as `add <name> --tag <ref>`, `install <path> --dry-run`, `status <path> --check`, and `skill check <dir> --json` therefore do not parse their trailing flags.

Required correction: support interspersed flags for every command and add parser regressions.

### F-004: the Phase 9 CLI surface is incomplete

Severity: high

The implementation plan requires the complete informative Spec §15 surface. Missing behavior includes bootstrap, project add and resolve, multi-project `--all`, install audit overrides, global status/update/upgrade, hybrid status, and standalone audit execution. Project aliases are also treated as filesystem paths by install and status.

Required correction: implement the missing dispatch and target resolution behavior, then cover it through CLI tests.

### F-005: status check misses installed content tampering

Severity: high

`statusDrift` compares only the direct declaration commit with the marker commit. It does not validate marker schema or recompute Spec §8.5 `content_sha256`, so `status --check` reports a locally modified installation as up to date.

Required correction: validate marker schema and content hash in read-only status checks.

### F-006: global installation omits required gates and locale semantics

Severity: high

Global installation ignores the global Skillfile locale, source audit gate, legacy skill dependency validation and warning, and moved-tag detection. The corresponding project path implements these behaviors, and Spec §9.2 uses the same Skillfile and installation model.

Required correction: apply the shared validation, dependency, audit, and moved-tag gates before global materialization.

### F-007: registry HTTP failures can be treated as successful responses

Severity: high

The snapshot and record HTTP readers do not reject non-2xx status codes. A failing records endpoint that returns a JSON object without records can suppress offline-cache fallback and be interpreted as a successful unknown result.

Required correction: reject non-2xx responses and malformed `records` shapes. Add cache fallback regressions.

### F-008: the interoperability gate is incomplete

Severity: high

P10 requires golden markers and a documented regeneration script. The current golden corpus contains hashes, a context list, and registry objects, but no marker fixture and no generator. The README claim that the gate covers all planned artifacts is not proven by the repository.

Required correction: add a normalized marker golden and a standalone deterministic generator that does not import Curator implementation packages.

### F-009: board completion state is not supported by task history

Severity: medium

Before this review the board contained no TASK directories, while every phase was marked done. Late-phase epics also lack STORY cards. This contradicts the board-first workflow but cannot be repaired retroactively without inventing history.

Required correction: keep this review task authoritative for corrective work and record evidence here. Do not fabricate historical task records.

## Deferred specification questions

- Snapshot objects are described with six required fields in Spec §13.4, but both implementations accept a correctly signed object that omits fields other than `version` and `created_at`. This needs a specification decision before changing compatibility behavior.
- PowerShell hook prose says activation runs on every prompt, while the observable implementation only invokes activation when the hook is loaded. This also needs a specification decision.

## Resolution evidence

| Finding | Status | Evidence |
|---|---|---|
| F-001 | corrected | `validateNodes` runs before phase 9; `TestRuntimeOnlyProviderStillRequiresSkillMd` proves runtime-only rejection before writes |
| F-002 | corrected | marker serialization normalizes mandatory arrays and nullable locale; marker unit test and `TestGoldenMarkerObject` prove the parsed wire object |
| F-003 | corrected | `parseInterspersed` supports flags around positional arguments; CLI parser tests cover trailing and optional audit flags |
| F-004 | corrected | bootstrap, project, alias, all-project, global, hybrid, standalone audit, and scope status paths are dispatched and covered by CLI/config tests |
| F-005 | corrected | status recomputes the marker content hash; CLI e2e proves clean success and tamper failure |
| F-006 | corrected | global install applies manifest locale, full validation, legacy dependency checks, audit, and scope-correct moved-tag detection |
| F-007 | corrected | registry snapshot and record fetches reject non-2xx responses; offline cache fallback has a regression test |
| F-008 | corrected | `marker.json` and the standalone `testdata/golden/regenerate` generator now complete the P10 artifact set |
| F-009 | corrected prospectively | this review and every corrective change are tracked by TASK-260713-7a9c1e; no historical records were fabricated |

## Spec §17.1 audit evidence

1. Package formats, whitelist, localization, schemas 1 through 5, and rejection rules are covered by `internal/skillspec`, `internal/whitelist`, `internal/locale`, `internal/skillcheck`, and their tests.
2. Skillfile schema 1 and development substitutions are covered by `internal/manifest`, `internal/devsub`, and their tests.
3. Canonical allowlist identities, closure unification, conflicts, cycles, ordering, and activation are covered by `internal/identity`, `internal/closure`, and graph tests.
4. Marker objects, byte-exact hashes, runtime layout, shims, adapters, tamper detection, and end-to-end install are covered by install tests and the independent golden gate.
5. Installation invokes only the fixed git executable and filesystem operations. No skill-provided command is launched during install.
6. MCP configuration surfaces and any/all read-only checks are covered by `internal/mcp` tests and install gating tests.
7. Registry canonical bytes, Ed25519, matching, deny-wins, snapshot monotonicity and staleness, TTL, offline grace, install policy, marker attestation, publication, and enforced configuration are covered by registry, install, config, and golden tests.
8. Curator does not provide a registry service, so the conditional service requirement does not apply.

Audit decision semantics and canary blocking are covered by `internal/audit` decision-table, cache, revocation, pin, and canary tests.

## Local verification

- `go test ./... -count=1`
- `go test -race ./... -count=1`
- `go vet ./...`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run`
- `go run ./testdata/golden/regenerate -root .` followed by a clean fixture diff
- Linux and Windows test-binary compilation through `go test -exec=true ./...`
- Naming gate over tracked source and commit messages
- `git verify-commit` for every corrective commit
