# TASK-260905-2tvae4 review verdict: ACCEPT (Change Request CR-TASK-260905-2tvae4-1 rev 1)

Reviewer run: RUN-260905-21b82b. Subject: draft worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-env-manager`, branch `draft/environments-manager-cli-1-1`, head `f61ee9a75cd1861a9993b0ee9ad4ad32a5ef3c9f` (= remote head, = PR #41 head). Reviewed read-only; mutants ran on an rsync copy under `.temp/review-2tvae4/scratch`.

## Why an empty repository delta is the right outcome for this leaf
The Change Request snapshots the managed story worktree (`task-board/story/STORY-260905-2z9pw4`), which must stay at the cycle-1-accepted `9af8af8`. The producer brief explicitly directed the rework into the draft worktree / PR #41 branch and forbade touching the story worktree. The deliverable of this leaf is the amended draft head `f61ee9a` plus the drafting report, not a story-branch change. So `repository_delta=empty` is expected and correct; the substantive delta is `git diff 9af8af8 f61ee9a` in the draft worktree, reviewed below.

## Verified (all reproduced by me, not accepted from the report)
| Check | Result |
| --- | --- |
| `git log a68559b..f61ee9a` | exactly one commit; author Ivan Oparin <oparin@me.com>; `%G?`=U locally (no allowed_signers), GitHub verification `verified=true reason=valid` |
| `git diff a68559b -- conformance/v1/vectors/manager-config.json` | empty; manifest hash for the file is back to main's `e915c1f7...` |
| `git diff --stat 9af8af8 f61ee9a` | 9 files: CHANGELOG, conformance/README, manifest.json, manager-config-v2.json (new), manager-config.json, rc.9 pin, generate-vectors/main.go, validate.py, test_validate.py. Nothing else in 9af8af8 changed. |
| Case accounting | 9af8af8 file had 25 cases; now 15 (all schema 1) + 10 in v2 (9 schema-2 + `schema1-rejects-environments`); the union equals the old set case-for-case (JSON equality) |
| `make validate` (venv) | exit 0 (`validated 58 schemas and 819 vector files`) |
| `make regenerate-check` | exit 0 (generator output byte-stable) |
| `python3 -B -m unittest discover -s tools -p 'test_*.py'` | 170 tests OK |
| `implementation_coverage.py families --root conformance/v1` | 18 claims upheld; `.github/ci/implementation-coverage.tsv` untouched (no manager-config row exists; ledger lists consumed cases only) |
| Pinned Go manager a3abcf34 (`implementations.yml` ref), `CURATOR_CONFORMANCE_ROOT=<draft>/conformance/v1 go test -count=1 ./internal/interop` | `ok internal/interop 0.402s`, exit 0 |
| NEGATIVE: same pin against a `git archive 9af8af8` root, `-run TestManagerConfigVectors` | FAIL on schema2-minimal-defaults, schema2-empty-environments-defaults, schema2-partial-knobs-fill-defaults, schema2-every-knob — reproduces the hosted failure the split removes. `golden_test.go:290` reads only `vectors/manager-config.json`, so the v2 file is unread by the pin. |
| PR #41 checks at f61ee9a | Implementations pass on ubuntu/macos/windows (run 33969621585); Specification x3, Links, Formatting pass; Release target provenance skipping (as on main) |

## Gate attack (validator), via the production entry point `python3 tools/validate.py` AND the function directly
| Mutant | Function result | Entry-point result |
| --- | --- | --- |
| baseline / restored | GATE PASSED | exit 0 |
| M1 schema-2 case moved back into manager-config.json | `byte-frozen schema-1 family; it carries schema versions [1, 2]` | exit 1 (manifest digest mismatch fires first) |
| M2 forged `valid=true` on schema2-unknown-environments-field | `expected valid=True, got Additional properties are not allowed` | exit 1 |
| M3 drifted one default in `expected.environments` | `expected.environments is not defaults plus input` | exit 1 |
| M4 v2 file deleted | FileNotFoundError (fails closed, no vacuous pass) | `manifest inventory mismatch; extra=['vectors/manager-config-v2.json']` |
| M5 v2 with only the schema-1 rejection case | `manager-config-v2.json carries no schema-2 case` | exit 1 |
| M6 case name duplicated across families | `name missing or repeated: 'minimal-defaults'` | exit 1 |
No survivors. The mutants include forged evidence (M2), narrowing (M5), drift (M3), and absence (M4) — not delete-only. `validate_manager_config_vectors` is called from `validate.py` main (line 3534), so the gate is in force.

## Non-blocking notes
- Brief asked for an identity header on the v2 file; the producer kept the bare case-list shape of the existing family and said so. Consistent with the Go reader's shape; acceptable.
- `%G?`=U locally is a missing allowed_signers file, not a bad signature; GitHub reports valid.
- My `make validate` run left an untracked `tools/__pycache__/` in the draft worktree; I removed it. Nothing tracked changed there.

## Verdict
ACCEPT. Record via `accept_cr(TASK-260905-2tvae4, revision=1, evidence=TASK-260905-2tvae4_review-verdict.md)`. Not marking done; integration belongs to the producer side. repeat-of: none.
