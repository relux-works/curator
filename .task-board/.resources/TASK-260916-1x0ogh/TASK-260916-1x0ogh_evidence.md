# Evidence — TASK-260916-1x0ogh: spec-provider-resolution-trust-roots (E4), rev2

Story `STORY-260916-2otjbn` (umbrella-provider-trust-roots), wave 1.
Role: doc-writer. Worktree:
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-2otjbn/worktree`
(branch `task-board/story/STORY-260916-2otjbn`, base `main` @ `07e2b41`).
No commits, no pushes, no branches created by this task.
Rev2 answers `TASK-260916-1x0ogh_review-verdict-rev1.md` (changes requested,
F1–F4); everything it marked passing stays as it is.

## Finding

`docs/security-audit-2026-09.md` E4 (confirmed, Appendix B:
`cmd/curator/umbrella.go:30-63` resolves `curator-<name>` with
`exec.LookPath` on the ambient `PATH`; only manager-published directories
refused), composed with S6 (project `.agents/env.sh` sourced on directory
change can prepend a project directory to `PATH` and plant `curator-run`).

## Rev2 closure of review findings F1–F4

- F2 (high; revision A keeps PATH selection): `protocol/environments.md`
  §11 rewritten — `§11:2230-2246` revision A searches the ambient `PATH`
  exactly as before (first executable directly inside a `PATH` entry),
  refuses a `PATH`-selected published/managed candidate, fails on an
  unreadable root, resolves silently when the selection lies directly
  inside a trust root, otherwise resolves with
  `subcommand_provider_outside_trust_roots` (path + roots + hint), and
  reports `subcommand_provider_missing` when no `PATH` entry holds the
  provider even when a trust root does. `§11:2247-2270` revision B
  searches only the trust roots, never `PATH`. Vectors fixed:
  `listed-directory-provider-resolved` revision_a now warns the `PATH`
  copy; `install-dir-provider-missing-then-resolved` (renamed),
  `install-dir-beats-listed-directory`, `listed-order-first-match-wins`
  revision_a now missing; new
  `path-selects-different-provider-while-trusted-exists` (A: `PATH` +
  warning, B: trusted) and `trusted-provider-on-path-resolves-silently`
  (both silent). `cli/curator.md:122-126` aligned with both profiles.
  `CHANGELOG.md:99-119` states A keeps selection, B never selects.
- F3 (medium; disjoint B outcomes): `§11:2247-2270` lists five mutually
  exclusive B outcomes (trusted dispatch; published/managed refusal;
  diagnostic-only `PATH`-probe refusal; true missing excluding `PATH`-only;
  unreadable failure). `§11.1:2275-2278` table rows are disjoint and name
  the refused path + consulted roots. B performs a diagnostic-only `PATH`
  probe (stated in `§11:2257-2261` and `§11:2185-2191`); vectors carry
  `untrusted_path` + `trust_roots_consulted` on every refusal and
  `trust_roots_consulted` on every missing. New `managed-directory-refused`
  covers the managed branch under both revisions.
- F4 (medium; unreadable is failure): new closed diagnostic
  `subcommand_provider_root_unreadable` admitted in `§11:2180-2184`
  (read failure, first unreadable root in order, never absence/fallback,
  §8.4 cited), `§11:2236-2237` (A fails instead of resolving),
  `§11:2264-2266` (B fails, no later/probe/absence fires),
  `§11.1:2278`, `§12:2313` (status names the directory) and
  `§12:2327-2333` (non-current, currency unknown). Vectors
  `unreadable-listed-directory-fails` and
  `unreadable-install-directory-fails` fail under both profiles with
  `unreadable_directory` + `trust_roots_consulted`; the old
  `contributes-nothing` case is gone.
- F1 (high; semantic enforcement): `tools/validate.py` gains the resolver
  model `_umbrella_expected` plus `_umbrella_path_search`,
  `_umbrella_trust_roots`, executable filtering and forbidden containment
  checks; `validate_umbrella_provider_vectors` now compares every declared
  `resolved`/`diagnostic` and detail field against the model for both
  profiles, and requires `trust_roots_consulted` on every warning/refusal/
  missing/unreadable outcome, `untrusted_path` on refusals,
  `migration_hint_directory` on warnings, `unreadable_directory` on
  failures. `tools/test_validate.py` `UmbrellaProviderVectorTests` grows
  from 6 to 18 tests, including the reviewer's mutant
  (`test_s6_revision_b_silent_resolution_mutant_fails`) which fails as
  required, plus narrowing mutants for order, executability, published
  dispatch, path/roots/hint/directory details, missing-vs-untrusted and
  unreadable-vs-absence.

## What changed per file (rev2 delta; rev1 carried)

- `protocol/environments.md`: §11 intro (profile-dependent search);
  trust-roots bullet (unchanged roots, unreadable sentence moved to its
  own fail-closed bullet); new unreadable bullet; `PATH` bullet (A
  selects, B probes only); published/managed bullet (per-profile scope,
  names path + roots); dispatch bullet (per-diagnostic details); S6
  paragraph (A still runs + warns, B refuses); trust-model paragraph
  (B roots, A verdict + `PATH` selection); rollout A/B rewritten as
  above; §11.1 four-row disjoint table; §12 status discovery sentence
  (active-revision search, unreadable directory) and currentness rule
  (refuse-or-fail non-current, A warning current, A `PATH`-absent
  missing); §13 vector inventory (A selection cases, published/managed,
  unreadable failures) and conformance sentence (warn/fail, refuse-or-fail).
- `cli/curator.md:122-126`: unknown-subcommand resolution cites §11 with
  both profiles (A `PATH` + warning, B roots only, never `PATH`).
- `CHANGELOG.md:99-119`: E4 entry names the trust roots, the new
  unreadable diagnostic, both rollout steps (A keeps selection, B never
  selects + probe), status posture, vector/schema coverage, 0016 block.
- `tools/generate-vectors/umbrella_provider.go` (untracked, new in rev1,
  rewritten in rev2): 14 cases with the §11 detail fields; profiles
  `warning release: PATH selects…` / `flip release: trust roots only…`.
- `conformance/v1/vectors/umbrella-provider-resolution.json` (untracked,
  regenerated): 14 cases × 2 profiles as above; manifest + `release/`
  pin regenerated. Schema files, `manager_config.go`, `system_config.go`,
  `main.go` untouched in rev2 (rev1's `provider_directories` knob,
  lockability, schema cases carried).
- `tools/validate.py`: `UMBRELLA_PROVIDER_DIAGNOSTICS` +4th diagnostic,
  `UMBRELLA_PROVIDER_CASES` 14 names, resolver model + detail checks +
  semantic comparison + unreadable-coverage gate.
- `tools/test_validate.py`: `UmbrellaProviderVectorTests` docstring +
  12 new semantic/narrowing tests (18 total).
- `profiles/manager.md`, schemas, existing schema cases: unchanged from
  rev1 (review's 88/88 meaning-preserved finding still holds; rev2
  touches no schema or schema-case generator).

## Validation transcript

Shell `/bin/zsh`, worktree root as cwd unless noted.
Python via the repo venv on `PATH`
(`PATH=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH`),
`set -o pipefail` for the make gates. Every command run directly; exit
codes are the gate's real status.

1. `gofmt -l tools/generate-vectors/` → no output, exit 0.
   `go vet ./tools/generate-vectors/` → no output, exit 0.
2. `go run ./tools/generate-vectors -root .` (via `make regenerate`) →
   exit 0.
3. `python3 tools/validate.py` → `validated 60 schemas and 1054 vector
   files`, exit 0.
4. `make validate` → validate.py line as above, then `Ran 245 tests in
   203.493s / OK` (233 rev1 + 12 new umbrella), then `ok
   github.com/relux-works/curator-spec/tools/generate-vectors (cached)`,
   exit 0.
5. Focused: `UmbrellaProviderVectorTests` 18 tests OK (1.724s);
   `ManagerConfigVectorTests` + `SystemConfigV2SchemaTests` 41 tests OK
   (76.828s); other `tools/test_*` files 78 tests OK (31.798s);
   `test_validate` alone 167 tests OK (117.505s); `go test ./tools/...`
   ok (1.040s).
6. `make regenerate-check` in the worktree → exit 2 (make Error 1 from
   `git diff --exit-code`). Expected-failure rationale: the gate diffs
   the tree against HEAD, so any uncommitted change-request tree fails
   it by construction; the diff is this change set, release hunk is the
   two manifest-pin lines. Generator stability proven separately: two
   consecutive `go run ./tools/generate-vectors -root .` runs are
   byte-identical over all 1060 files under `conformance/v1` + `release/`
   (`IDEMPOTENT: regenerate output stable`), and a disposable byte copy
   with a temporary git baseline of the exact candidate runs `make
   regenerate-check` exit 0 (`go run …` then `git diff --exit-code …`,
   no output).
7. Reviewer's mutant (unit): S6 `revision_b` set to `{"resolved":
   "/home/operator/work/acme/.bin/curator-run", "diagnostic": null}`
   fails `validate_umbrella_provider_vectors` with `§11 model expects`
   (`test_s6_revision_b_silent_resolution_mutant_fails`, part of the 18).

## Deliberately out of scope

- Implementation (`TASK-260916-3oh0u8` manager, `TASK-260916-16ys92`
  launcher), proposal 0016 itself, launcher SPEC.
- No writability-fallback rule (settled decision adopts trust roots).
- UNC provider spellings refused by the schema (POSIX/drive arms only);
  stated in §11 via "the two spellings the schema admits".
- `cli/curator.md` follow-up from rev1 is closed in rev2 (edited per the
  rework brief); no other CLI rows touched.

## Checklist mapping

1. Normative rule in §11/§11.1/§12/§13, RFC 2119 keywords, closed lists;
   `provider_directories`, `subcommand_provider_outside_trust_roots`,
   `subcommand_provider_root_unreadable` (+ `missing`/`untrusted`) spelled
   identically in text, tables, schemas, vectors, generator and validator.
   Done.
2. Warn-first as two labelled profiles (A keeps `PATH` selection + warning
   with `/usr/local/bin` hint example; B trust-only + diagnostic probe +
   refusal); `env status` posture + currentness incl. unreadable. Done.
3. `vectors/umbrella-provider-resolution.json` 14 cases (trusted
   positives, order, A-selection incl. S6, published/managed refusals,
   unreadable failures, missing) in the manifest via `make regenerate`;
   `provider_directories` schema cases carried; `make validate` exit 0
   quoted above. Done.
4. CHANGELOG Unreleased entry "E4: …" (finding id, both steps, 0016
   dependency); spec patch rev2 + this evidence attached as outcome
   resources; no curator/launcher implementation touched (spec-repo
   conformance tooling only). Done on attach.
5/6. Docs consistent: §11/§12/§13, `cli/curator.md`, schemas,
   `profiles/manager.md`, generator, validator and vectors agree
   (validator cross-checks text details ↔ vectors semantically).
7. Outcome resources: `TASK-260916-1x0ogh_spec-patch_rev2.patch` +
   this file attached to the task.
8. Findings recorded here and in task notes; no LOGBOOK.md edit
   (worktree rules forbid it).
9/10/11. Spec matches AC (trust-root rule + vectors), fits the spec-repo
   architecture (closed sets, generator + semantic gate precedent), tests
   green (`make validate` exit 0, 245 tests).

## Patch identity

`TASK-260916-1x0ogh_spec-patch_rev2.patch`: worktree diff against the
fork base `07e2b41` (= worktree `HEAD`) via a temporary git index
(`read-tree HEAD`, `add -N .`, `diff HEAD`; actual index untouched).
4758 lines, SHA-256
`2d66b496c5779a4afa982369690aac5cb81362005099db146bedc4f9f406a38a`,
`git patch-id --stable` `2ac6d1969dcbed6ec9168fe85a8818637a7b2a2b`.
Covers the full rev1+rev2 change set (104 modified + 8 new files via
intent-to-add = 112 paths, 112 diffs).

Base note: `origin/main` has advanced since the fork (`07e2b41` →
`90d50c64`, "Specify the shell-hook trust gate for project env files
(S6)", 8 files). A diff against current `origin/main` would spuriously
list `conformance/v1/vectors/shell-hook-trust.json` (plus S6 hunks in
CHANGELOG/manifest/etc.) as deleted/changed. The patch is therefore
taken against the fork base so it contains only this task's changes;
merging with the S6 landing is the orchestrator's integration step. The
candidate tree itself holds no S6 content and no unrelated edits.
