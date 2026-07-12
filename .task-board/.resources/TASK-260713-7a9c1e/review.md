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
