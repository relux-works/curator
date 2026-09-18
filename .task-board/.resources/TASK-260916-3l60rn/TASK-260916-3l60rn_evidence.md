# TASK-260916-3l60rn evidence — E6 path-kind admission rule (spec revision 1)

Story `STORY-260916-wgt8vz`, wave 3 of the 2026-09 security-audit remediation.
Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-wgt8vz/worktree`
(branch `task-board/story/STORY-260916-wgt8vz`, base `e8b53a0` which carries S5 §4).
Role: doc-writer of normative spec text; no implementation.

Finding read first: `docs/security-audit-2026-09.md` E6 (Medium) and Appendix B
(E6 partially confirmed: dependencies are git-only so a `path`-kind MCP package
cannot enter a closure through `requires` — not applicable; a `path`-kind root or
overlay still carries `class: system` modules with no directory boundary — confirmed).

## What changed per file

- `decisions/0012-context-packages-and-semver-locks.md` — new `## Amendment
  (2026-09-18, E6)` in the E1-amendment shape (verbatim quotes, why-incomplete,
  evidence, added rule, normative home): item 1 MCP declarations git-only
  (quotes Decision 6 Policy), item 2 system modules from path need direct naming
  plus the §4 boundary contract (quotes Decision 5), item 3 the contract covers
  every path source directory incl. onboarding imports (quotes the §9.6
  compat row). Original passages marked `[Amendment 2026-09-18, item N]`.
- `protocol/environments.md`
  - §1 `path` rules: "never reads the source directory again" narrowed to
    "never adopts bytes … again" plus the §4 re-inspection sentence
    (`lstat`-class, metadata only, adopts no bytes) — the old sentence
    contradicted the new resolve-time verification.
  - §2.1 diagnostics table: new row `mcp_declaration_path_source_refused`
    (names the package and the declaration).
  - §2.2: new "Source kind: `git` only" rule — resolution MUST refuse a
    closure whose `mcp` member carries no canonical source (`state_sha256`
    pin), never admitted, never warned-through; rationale (no canonical
    identity, never verified per §1.4) and allowlist totality stated.
  - §3: new "`path` sources" clause — E2 direct-naming applies as is
    (root/overlay direct; a transitive path lock is unproducible since
    `requires` are git-only, §2); system modules additionally need the §4
    contract at every resolve and before any materialization; failure is
    `environment_store_untrusted`, not an admission verdict, with no rebuild;
    the check is on the directory, not the content class.
  - §4: new "`path` source directories" extension — the five boundary checks
    with the declared directory as root, same timing as the store contract
    (every resolve + under the mutation lock for mutating ops), no bytes
    adopted, no pin recomputed against the live directory (`state_sha256`
    stays the store-entry baseline), entry-class `environment_store_untrusted`
    (no fragment, non-current, posture names path + check), no rebuild
    (MUST NOT re-copy to clear the verdict). Rollout line added without
    touching S5's: direct (impact row "E6": under the hood).
  - §6: pointer paragraph (path overlays pass §4; MUST NOT carry MCP).
  - §8.5 + §10.4 tables: `environment_store_untrusted` condition cells
    extended with "a `path` source directory fails its §4 boundary checks".
  - §9.6: import is a `path` root for §4; MUST NOT carry MCP (reassembly
    emits no `requires.mcp`).
  - §10.1: store-trust verification object list gains the path source
    directory; repair parenthetical updated to "bytes are never adopted…".
  - §12: store-trust posture row gains the source directory, naming the
    failing check and the path.
  - §13: conformance-surfaces sentence for the new vector family.
  - §1.1/§3.1 tables unchanged: no new diagnostic belongs to them (the new
    diagnostic is §2.1's; §3 reuses `context_system_module_transitive` and
    `environment_store_untrusted`, already tabled). §12.1/§12.2 unchanged:
    no new knob, nothing lockable.
- `CHANGELOG.md` — Unreleased entry "E6: …" (finding id, rule, direct rollout,
  impact row "E6": under the hood, vector family).
- `conformance/v1/vectors/environments-path-kind-admission.json` — NEW,
  hand-authored (see below). 5 MCP-kind cases (git admitted; path
  overlay/root/import refused; 1 negative) + 11 boundary cases (admitted
  overlay/root/import; five single-check refusals; transitive E2 refusal;
  2 negatives).
- `conformance/v1/manifest.json`, `release/1.0.0-rc.9.json` — regenerated
  (`make regenerate` equivalent): manifest +4 lines (new file entry only),
  rc.9 pin updates only. All pre-existing vector files byte-identical
  (`git status` shows no other conformance/ file modified).
- `tools/validate.py` — new E6 gate
  `validate_environments_path_kind_admission_vectors` + pin tables, wired into
  `main()` (rule 7: exact inventories, per-name input pins, recomputed
  verdicts, negatives must still violate).
- `tools/test_validate.py` — new `PathKindAdmissionVectorTests` (21 tests):
  published-passes, both-direction recomputation, naming checks, wrong
  diagnostic, path-rewritten-as-git, healed negatives (3), five-to-one
  narrowing, transitive misdiagnosis/rewrite, pin-hash-comparison refusal,
  multi-failure, inventory-exact, capability identity.

## Why a new vector family (not environments.json)

`vectors/environments.json` is generated by `tools/generate-vectors`
(`writeEnvironmentVectors`); hand-editing it would be overwritten by
`make regenerate` and break `regenerate-check`. Every prior finding family
(S5/E1/S4/E5) ships as a hand-authored file with its own gate; E6 follows
that pattern. The brief's "transitive path package (if expressible)" is
covered as a lock-level admission case pinning E2 uniformity across pin
kinds; resolution cannot produce such a lock (`requires` are git-only, §2),
which §3 now states.

## Deliberately out of scope

Implementation (`TASK-260916-yvxbs1`), S5 itself (referenced), E1/E2 rules
(referenced only), E7, `profiles/manager.md` (no new knob/lockable/CLI row;
the posture row extends the existing store-trust row the manager text
already cites generically), schemas (no new knob or wire shape),
proposals 0014–0018. No LOGBOOK.md change. No implementation code touched.

## Validation transcript

Shell: `bash` in the spec worktree above. The worktree carries no Python venv
and CI installs `requirements-dev.txt` via pip; to avoid touching the system
interpreter, the two Python steps of `make validate` were run with the same
commands under a uv-managed interpreter pinned to `jsonschema==4.25.1` (the
exact `requirements-dev.txt` pin). `go run`/`go test` used the system Go
toolchain (go1.26.0 darwin/amd64).

- `go run ./tools/generate-vectors -root .` → exit 0 (regeneration proof;
  only manifest.json + rc.9.json regenerated, see above).
- `python3 tools/validate.py` → exit 0, output:
  `validated 62 schemas and 1096 vector files`
- `python3 -B -m unittest discover -s tools -p 'test_*.py'` → (full suite,
  transcript below)
- `go test ./tools/...` → exit 0, output:
  `ok  github.com/relux-works/curator-spec/tools/generate-vectors  1.685s`
- Focused: `PathKindAdmissionVectorTests` → 21 tests, OK.
- Subsets (bounded runs): `test_release_gate.py` 32 OK (162s),
  `test_verify_*.py` 10 OK, `test_implementation_coverage.py` 36 OK.

### Full-suite transcript

Command (the `make validate` Python test step, run to completion with exit 0
in this turn):

```
uv run --no-project --with 'jsonschema==4.25.1' python3 -B -m unittest discover -s tools -p 'test_*.py'
```

Tail (progress dots elided):

```
----------------------------------------------------------------------
Ran 428 tests in 874.937s

OK
EXIT: 0
```

## Closed-set spelling

`mcp_declaration_path_source_refused` spelled identically in 0012 amendment,
§2.1 row, §2.2 rule, §13, CHANGELOG, vectors (diagnostics pin + 3 cases),
`tools/validate.py` constant/pin/gate, and tests. Reused
`environment_store_untrusted` and `context_system_module_transitive` unchanged.
No new knob, key, or schema field.

---

# Revision 2 (rework after `TASK-260916-3l60rn_review-verdict-rev1.md`)

Role, worktree, base (`e8b53a0`), finding, and settled decisions unchanged.
Revision 1 was rejected with two corrections (R1, R2); everything else passed
and is kept byte-identical (verified: all 382 rev1 lines of
`environments-path-kind-admission.json` present in order in the rev2 file;
`decisions/0012-*.md`, `§1/§2.1/§2.2/§3/§6/§8.5/§9.6/§12` untouched in rev2).

## R1 — dry-run no-rebuild stated consistently (§4, §10.1, §10.4 + vectors)

- §4 "`path` source directories" extension gains the explicit exclusion: a
  `path` directory failure "is not in the entry-rebuild branch of (b)
  above: dry-run evaluation … reports `environment_store_untrusted` with no
  rebuild planned and mutates nothing — never
  `would-rebuild-untrusted-store`, which names only a store entry, lock, or
  marker file failure that a real operation would rebuild". S5's (b)
  paragraph untouched (its closed object list already excludes path
  directories); the operator-repair rule unchanged.
- §10.1 dry-run sentence now reads "an entry-class failure of a store
  entry, the lock file, or a marker file reports
  `would-rebuild-untrusted-store` …; dry-run evaluation of an
  enclosing-boundary failure, or of a `path` source directory failure,
  reports `environment_store_untrusted` with no rebuild planned".
- §10.4 table: the `would-rebuild-untrusted-store` row is scoped to "a
  store entry, the lock file, or a marker file"; the
  `environment_store_untrusted` no-rebuild row gains "or of a `path`
  source directory failure".
- Vectors: new `dry_run_cases` array in
  `environments-path-kind-admission.json` (new family, not the S5 file, so
  all pre-existing vector files stay byte-identical):
  `path-overlay-dry-run-untrusted-no-rebuild` (world-writable system
  overlay → diagnostic `environment_store_untrusted`, outcome null,
  mutated false, no rebuild, posture names path + `permissions`),
  `path-overlay-dry-run-intact-plans-nothing` (trusted control), and the
  negative `path-overlay-dry-run-reports-would-rebuild` (`conforming:
  false`, identical inputs, outcome `would-rebuild-untrusted-store`).
- Gate: the E6 dry-run loop recomputes diagnostic/outcome/mutated/rebuild/
  failing-check/naming from the five checks; expected outcome is always
  null ("a path directory dry-run never plans a rebuild"); the negative
  must report exactly `S5_OUTCOME_WOULD_REBUILD` (shared constant, so the
  spelling is identical by construction) with every other field correct
  ("only the reported outcome violates the §10.1 rule").
- Tests: 9 new dry-run tests — would-rebuild report refused, untrusted
  silence refused, intact-reports-untrusted refused (both directions),
  mutates refused, rebuild-planned refused, healed failure-branch refused
  (replacement), healed negative refused, wrong-violation-mode refused
  (healed outcome + mutated still refused: violation mode pinned), dry-run
  inventory exact.

## R2 — boundary check pinned regardless of content class (vectors + gate)

- New boundary cases: `path-overlay-no-system-modules-admitted` (trusted
  control: no-system overlay, all checks pass),
  `path-overlay-no-system-world-writable-untrusted` (`permissions` fails,
  same directory/package as the system refusals),
  `path-import-no-system-wrong-ownership-untrusted` (`ownership` fails;
  trusted control is the existing
  `path-import-no-system-modules-admitted`). Each is pinned to
  `carries_system_modules: false` plus its failing check, role, and origin.
- Gate: `E6_BOUNDARY_CASES`/`E6_BOUNDARY_PIN` extended; verdict logic
  unchanged (checks only, never content class).
- Tests: 6 new R2 tests — both refusals healed refused, both refusals
  rewritten-as-system refused (`carries` pin), the new control
  rewritten-as-system refused, and
  `test_no_system_narrowing_disagrees_with_corpus`, which replays the
  reviewer's narrowing as verdicts (`trusted if carries else True` +
  the E2 transitive branch) and asserts it contradicts the published
  observations on both new refusals. A manager that skips the directory
  check for no-system sources is now non-conforming against this corpus.
- Reviewer-narrowing replay (independent probe, this turn): the narrowed
  verdicts mismatch exactly the two new no-system refusals (plus no
  false positive on the transitive E2 case once its branch is accounted
  for) — the narrowing that survived rev1 fails against rev2.

## Rev2 per-file delta (on top of rev1)

- `protocol/environments.md`: R1 sentences in §4/§10.1/§10.4; §13
  enumeration extended (no-system overlay admitted, no-system refusals,
  dry-run case, would-rebuild-reporting negative). Nothing else touched.
- `CHANGELOG.md`: E6 entry's conformance summary updated (admitted root/
  overlay/import; boundary refusals "one per check plus no-system overlay
  and import refusals"; dry-run cases). Rule text, rollout, finding id
  unchanged.
- `conformance/v1/vectors/environments-path-kind-admission.json`: +3
  boundary cases, +3 dry-run cases (new `dry_run_cases` array); all rev1
  cases byte-preserved (382/382 lines in order; +160 lines).
- `conformance/v1/manifest.json`, `release/1.0.0-rc.9.json`: regenerated
  (manifest: same +4-line new-file entry with the rev2 digest; rc.9: the
  two pin lines only). All 1095 pre-existing vector files byte-identical
  (`git diff --name-only HEAD -- conformance/v1` lists only manifest.json
  plus the intent-to-add E6 file).
- `tools/validate.py`: `E6_BOUNDARY_CASES`/`E6_BOUNDARY_PIN` +3,
  `E6_DRY_RUN_CASES`/`E6_DRY_RUN_NEGATIVE_CASES`/`E6_DRY_RUN_PIN` new,
  dry-run gate loop new; gate docstring gains §10.1/§10.4. No other gate
  touched.
- `tools/test_validate.py`: `PathKindAdmissionVectorTests` 21 → 36 tests
  (+9 dry-run, +6 no-system); class docstring gains the dry-run verdict.
- Untouched in rev2: `decisions/0012-*.md` (amendment already states both
  rules), schemas, `profiles/manager.md`, proposals, LOGBOOK.md (campaign
  prohibition), implementation code. `git status --short` lists exactly
  the same 8 paths as rev1.

## Rev2 validation transcript

Shell: `bash` in the spec worktree. Python via the campaign venv
(`PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH"`,
`set -o pipefail` semantics: every gate command quoted below ran as a
standalone process, exit code is the command's own). `go run`/`go test`
used the system Go toolchain. The full Python suite was split into
bounded per-file / per-class runs (each under 10 minutes); every test
method ran exactly once across the batches (443 = 428 rev1 + 15 new).

- `go run ./tools/generate-vectors -root .` → exit 0 (no output).
  Regeneration idempotence: re-ran after all edits, exit 0; manifest
  (`997b8de4…`) + rc.9 (`3f1468c6…`) sha256 stable, `git status --short`
  still exactly the same 8 paths (see footprint above). (`make
  regenerate-check`'s `git diff --exit-code` cannot pass with intended
  uncommitted changes present; idempotent re-run is the candidate-relative
  proof, as in rev1.)
- `python3 tools/validate.py` → exit 0, output:
  `validated 62 schemas and 1096 vector files`
- `python3 -B -m unittest discover -s tools -p 'test_implementation_coverage.py'` → 36 tests, OK (3.776s), exit 0
- `python3 -B -m unittest discover -s tools -p 'test_verify_release_commit.py'` → 5 tests, OK, exit 0
- `python3 -B -m unittest discover -s tools -p 'test_verify_release_merge_policy.py'` → 5 tests, OK, exit 0
- `python3 -B -m unittest discover -s tools -p 'test_release_gate.py'` → 32 tests, OK (87.195s), exit 0
- `test_validate.py` batch 1 (`-k` per class: WireSemantic 18, AssuranceRelational 3, RepositoryDescriptor 6, ManagerLifecycle 4, BuildDriverGolden 13, SharedFixtureMarker 10) → 54 tests, all OK, exit 0
- `test_validate.py` batch 2 (WorkflowRegeneration 2, EnvironmentVector 27, EnvPassthrough 27, StoreBoundary 25) → 81 tests, all OK, exit 0
- `test_validate.py` batch 3 (PathKindAdmission 36, SourceSigners 47, ContextVersion 15, ContextDetector 9, SnapshotAcquisition 5) → 112 tests, all OK, exit 0
- `test_validate.py` batch 4 (ShellHookTrust 8, WriteNofollow 12 in 173.972s, UmbrellaProvider 18, ManagerConfig 26, SystemConfigV2 24, RegistryPageBoundary 13, RegistryCheckpoint 17) → 118 tests, all OK, exit 0
- `go test ./tools/...` → exit 0, output:
  `ok  github.com/relux-works/curator-spec/tools/generate-vectors  6.637s`
- Total: 443 tests (36+5+5+32+54+81+112+118 = 428 rev1 + 15 new), every batch exit 0. No test was narrowed, deselected, or skipped; each method ran exactly once.

## Rev2 closed-set spelling

No new diagnostic, knob, lock key, or schema field in rev2. Reused
spellings: `environment_store_untrusted` (dry-run diagnostic),
`would-rebuild-untrusted-store` (forbidden dry-run outcome; vector
negative, §4/§10.1/§10.4 text, and the shared `S5_OUTCOME_WOULD_REBUILD`
constant — identical by construction). All rev1 spellings unchanged.
