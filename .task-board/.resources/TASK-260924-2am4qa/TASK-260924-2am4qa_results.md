# TASK-260924-2am4qa results

## Before and after

Before, core §4.4 dependency entries selected a provider from its pinned Git
repository and ref, which meant the skill package was at the repository root.
Closure and package identity distinguished repository and commit, while lock,
marker and local audit identity had no selected dependency subdirectory.

After, opt-in draft manifest schema 9 allows optional `directory`. It uses the
same directory grammar and containment contract as the Skillfile schema-2
individual selector. Omission normalizes to `.` and preserves the existing
root selection. The selected package must be a real, link-free folder with a
matching `SKILL.md`. Package identity and closure now include canonical
repository, full commit and normalized directory; same-name requirements at
different directories conflict, while distinct skill names at different
directories of one repository can be separate closure nodes. Lock and marker
package identity, local audit records and audit-cache keys bind that directory.
Schema 9 uses draft marker v5. Schemas 1–8 and rc.9 artifacts remain unchanged.

## Schema and conformance coverage

- Added draft `agent-skill-v9` and `csk-skill-v9` schemas, using the shared
  `source-types-v1` directory definition also referenced by Skillfile
  selectors and lock members. Draft marker v5 admits skill schema version 9.
- Added valid subfolder, omitted-root and explicit-root schema instances, plus
  negatives for absolute, escaping, empty, backslash, parent-component and
  glob directories.
- Added semantic vectors for root normalization, valid subfolder package,
  missing folder, folder without `SKILL.md`, symlink escape, same-name/different
  directory conflict, and a diamond selecting frontend/backend folders from
  one repo while both select one shared folder. The diamond records one shared
  package node and one repository acquisition.
- Updated core §4.4, Skillfile source identity/audit rules, manager planning,
  audit and marker rules, schema README cross-references and Unreleased notes.

## Validation results

| Command | Result |
|---|---|
| `python3 tools/validate.py` | **0** — `validated 64 schemas and 1169 vector files`. The initial attempt returned 1 because the host Python lacked `jsonschema`; rerun used the pinned `requirements-dev.txt` dependency in an ignored worktree venv. |
| Full `python3 -B -m unittest discover -s tools -p 'test_*.py'` | **130 (interrupted twice)** — a single discovery process exceeded the approximately 10-minute process bound. The second attempt reached existing `ReadFailureVectorTests.test_substituted_scenario_rejected_through_main`, which invokes the full validator repeatedly. No assertion failure was emitted before interruption. The first attempt also copied the temporary venv into release-gate fixtures; the venv was then moved under ignored `__pycache__`. |
| Bounded Python unit chunks | **0 across all chunks — 615/615 tests passed.** Commands used `python3 -B -m unittest discover -s tools -p '<module>.py'` for `test_implementation_coverage.py` (36), `test_release_gate.py` (32), `test_verify_release_commit.py` (5) and `test_verify_release_merge_policy.py` (5). All 537 `test_validate.py` tests passed through bounded `python3 -B -m unittest test_validate.<Class>` runs: 61 (schema registry, draft directory, wire semantics, assurance, repository identity, lifecycle, build-driver, shared-marker and workflow classes); 79 (environment, env-passthrough and store-boundary classes); 115 (path-kind, source-signers and Codex-seed classes); 61 (context-version, context-detector, snapshot, shell-hook and security-posture classes); then individual groups of 12 (write-nofollow), 20 (dotfile managers), 18 (read failures), 102 (umbrella provider, manager config, system config and takeover text), 13 (registry page boundaries), 17 (registry checkpoints) and 39 (registry bootstrap). Each bounded process exited 0. |
| `PYTHONPATH=tools python3 -B -m unittest test_validate.ManifestDependencyDirectoryDraftTests` | **0** — 4 tests passed separately, including schema/vector validation and narrowing mutants for audit identity, parent escape and diamond provider order. |
| `go test ./tools/...` | **0** — Go vector generator tests passed. |
| `make regenerate-check` | **0** — generator output matches the frozen `conformance/v1` and rc.5–rc.9 files. |
| `git diff --check` | **0** — no whitespace errors. |

The full discovery invocation did not return before its per-process bound; all
615 Python test cases passed when run by bounded file/class chunks, as directed
for long gates. The repository has no separate lint target in its Makefile.

## Findings and handoff

Schema 9 is deliberately draft-only so this amendment does not widen frozen
schemas or rc.9. The full discovery process is too long for one bounded run;
the complete test set passed in file/class subsets. No LOGBOOK.md edit was made;
this task-scoped artifact records the implementation and validation findings.

## Revision 2 (cleanup)

Per review verdict rev1 (blocking item 1): deleted `TASK-260924-2am4qa_results.md` from the repository root of the Story worktree. It remains only as this board outcome resource. No other file changed.

Evidence (this run, pipefail shells, rc=0):
- Pre-removal `git status --short` showed the 14 tracked modifications plus 5 untracked paths including the stray results file; worktree copy sha256 `f469cfa78f3c27171313a4f9ff94fb83a08d60825263debfb5fb0deb4d611ef5` matched the board resource byte-for-byte (`diff` clean).
- `rm TASK-260924-2am4qa_results.md` rc=0; post-removal `git status --short` lists only the same 14 modified + 4 untracked code paths.
- Per-file sha256 of all 18 remaining content paths before (minus the removed file) vs after: `diff` clean — BYTE-IDENTICAL-ALL-OTHER-PATHS. Untracked schema-case dirs (22 files) untouched.
- No commit made; work left uncommitted for handoff snapshot. No `TASK-260924-2am4qa_results.md` recreated in the worktree (this edit applied only to `/tmp` copy via `task-board resource update`).
