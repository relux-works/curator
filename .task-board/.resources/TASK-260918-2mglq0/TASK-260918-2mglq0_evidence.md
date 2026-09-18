# Evidence — TASK-260918-2mglq0: S1+S3 spec Revision 4 (rebase onto 1ca4b3d + codex-seed thirteenth posture row)

Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-2qmrb8/worktree`
(branch `task-board/story/STORY-260910-2qmrb8`, base `1ca4b3d`, own HEAD — NOT `origin/main`).
Role: doc-writer (normative spec text + vectors/validator only). No implementation code touched.
No LOGBOOK.md change anywhere. Curator worktree untouched (EMPTY delta, verified below).

Revision 3 was accepted on base `e8b53a0` (`TASK-260910-2qtiho_review-verdict-rev3.md`,
patch-id `d34f8329`); this revision is its rebase onto curator-spec main `1ca4b3d`
(E3 codex-seed + E6 path-kind admission landed) plus the substantive correction:
absorbing the E3 shipped-revision row into the closed posture inventory as the
thirteenth row. For the rev-1–rev-3 history see `TASK-260910-2qtiho_evidence.md`.

## Start-state verification (before any edit)

- `git rev-parse HEAD` → `1ca4b3df32baceff99f6be4b136f39addfa4da00` (`1ca4b3d`, E6).
- `git diff HEAD | git patch-id --stable` →
  `43f07deedbb0fed7270b0c81454b4fb94b4baedc`, identical to the attached
  `TASK-260910-2qtiho_spec-patch_rev3-rebased.patch` id and its
  sha256 `10ab0fbb…0daf` (229490 bytes). The union hunks the orchestrator
  resolved (E6 `mcp_declaration_path_source_refused` row, §12 status paragraph
  with E3 codex-seed/per-home rows + the twelve posture rows, E3 warnings,
  §13 vector enumeration, CHANGELOG order) are present in the tree as described.
- 24 files changed vs HEAD; 8 new files intent-to-added (` A` in porcelain);
  nothing staged.

## What changed vs the rebased rev-3 state (8 files; all else byte-identical)

Fixed position: `codex-seed` sits **directly after `update-confirmation`**
(the 10th of the thirteen posture rows), the other warn-first revision row.

- `profiles/manager.md`
  - §10 table: new `codex-seed` row — value `A` or `B` (environments §7.4),
    provenance `shipped` — directly after `update-confirmation`;
    "exactly these twelve" → "exactly these thirteen".
  - §12.7: "the twelve posture rows" → "thirteen"; prose inventory gains
    "codex seed" after "update confirmation".
- `protocol/environments.md` §12: "the twelve posture rows" → "thirteen";
  enumeration gains `` `codex-seed` `` after `` `update-confirmation` `` with
  the parenthetical "(the per-home `codex_seed_record` native-server rows
  above are not posture rows and stay unchanged)", in the style of the
  `source-signers` / `store-boundary` parentheticals — so "No other
  `env status` posture row exists" holds in the union.
- `conformance/v1/vectors/security-posture.json`
  - `status_gates`: `codex-seed` inserted after `update-confirmation` (13).
  - `scope_notes`: "twelve status_gates" → "thirteen status_gates".
  - Every case `shipped_revisions` gains `codex_seed` after
    `update_confirmation`: `A` in 16 cases, `B` in
    `posture-rows-flipped-revisions` — mirroring the `update_confirmation`
    `A-warning`/`B-flip` pinning (schema-1 case pins the input too, like
    `update_confirmation`, while its outputs stay header + four manager rows).
  - Every schema-2 `curator_status_rows` / `env_status_rows` gains
    `{"gate": "codex-seed", "value": "A"|"B", "source": "shipped"}` directly
    after the `update-confirmation` row (32 pinned rows).
- `tools/validate.py`
  - `SECURITY_POSTURE_STATUS_GATES` gains `codex-seed` after
    `update-confirmation`; "twelve" comments/message → "thirteen".
  - Shipped-revision closed set gains `codex_seed` ("closed four-gate set" →
    "closed five-gate set") with the closed-value check `("A", "B")`.
  - `SECURITY_POSTURE_SCENARIOS`: all 17 scenarios pin
    `shipped_revisions.codex_seed` (`A`, `B` for the flipped case) — rule 7.
  - `_posture_status_rows` derives the `codex-seed` row from
    `shipped["codex_seed"]` after `update-confirmation`.
- `tools/test_validate.py`: `SecurityPostureVectorTests` gains three negatives
  (docstring updated): `test_codex_seed_row_dropped_fails` (row omitted from
  both outputs), `test_codex_seed_row_misplaced_fails` (row swapped with
  `store-boundary`), `test_codex_seed_shipped_revision_rewritten_fails`
  (shipped `A`→`B` with rows re-derived — internally consistent under the
  same name, refused by the scenario pin; rule 7).
- `CHANGELOG.md`: S1/S3 entry now states thirteen rows and names the
  thirteenth as the E3 codex-seed shipped revision directly after
  `update-confirmation`, with the per-home `codex_seed_record` rows outside
  the inventory.
- Regenerated (`make regenerate`): `conformance/v1/manifest.json` (one-line
  digest update for `vectors/security-posture.json`) and
  `release/1.0.0-rc.9.json` (manifest pin twice). `schema-cases/index.json`
  and every other generated file are byte-identical — no schema-case change.

Byte-identity proof: `git archive HEAD` + `git apply` of the attached rebased
patch into `/tmp/rev3-base`, then `diff -r` against the worktree: only the 8
files above differ (plus the known `git archive` export-subst expansion of
`conformance/v1/fixtures/byte-exact/subst.txt`, which matches HEAD in the
worktree — the same artifact the rev-3 verdict records). Repo-wide grep finds
zero remaining "twelve" and "thirteen" exactly in the eight expected spots
(plus the CHANGELOG "thirteenth").

## E6 check (no new row)

E6 added no revision gate: the `path`-directory admission rule is an extension
of the §4 protected-boundary contract itself ("The contract extends to the
declared directory of every `path` source…", §4), always on, with failures
entry-class `environment_store_untrusted` reported in the §12 store-trust row.
The manager §10 `store-boundary` row reference — "`enforced` (environments
§4)" — still covers the §2.2/§4 extension, so the reference is unchanged and
no E6 row was added.

## Validation transcript (all green; shell `/bin/sh`, `set -o pipefail`)

Python is the repository venv
(`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin`,
prepended to `PATH`); system `python3` has no `jsonschema`, as in rev-1.

- `make regenerate` → `go run ./tools/generate-vectors -root .`, exit 0.
- `make regenerate-check` → exit 0 (after staging the generated paths, the
  rev-1 flow: `git add -- conformance/v1 release/1.0.0-rc.9.json`; the index
  was then restored with `git reset -q` + `git add -N` on the 8 new files and
  `git diff HEAD` verified hash-identical before/after). For the record, the
  unstaged run exits 2 by construction — `git diff --exit-code` compares the
  worktree against the index, and an uncommitted candidate is never staged —
  which is the expected shape, not a product failure.
- `python3 tools/validate.py` → `validated 62 schemas and 1112 vector files`,
  exit 0.
- `go test ./tools/...` → `ok …/tools/generate-vectors 0.740s`, exit 0.
- `python3 -B -m unittest` from `tools/` (same sys.path as
  `discover -s tools`), in bounded sequential groups, all exit 0:
  - `SecurityPostureVectorTests`: 24/24 OK in 0.9s (21 carried + 3 new).
  - `test_implementation_coverage test_release_gate
    test_verify_release_commit test_verify_release_merge_policy`: 78/78 OK.
  - `test_validate` groups: Wire+Assurance+RepoDesc+Lifecycle 31/31;
    BuildDriverGolden 13/13; SharedFixture+WorkflowRegen+Environment+
    EnvPassthrough 66/66; StoreBoundary+PathKind+SourceSigners+CodexSeed
    140/140; remaining 10 classes 147/147.
  - Total: 499/499 tests green (421 in `test_validate` incl. the 3 new).
  - (One mistyped invocation, `python -m unittest tools.test_validate…`
    from the repo root, failed with `ModuleNotFoundError: assurance` —
    wrong sys.path on my part, not a suite failure; rerun from `tools/`
    per the Makefile's discovery layout.)

## Deliberately out of scope

Implementation, tags/releases, proposals 0014–0018, any gate beyond the
codex-seed absorption, any reference change for E6 (checked, not needed).
No `manager-config-v2.json` vector cases, no `cli/curator.md` change — same
boundary as rev-3.

## Artifacts

- `TASK-260918-2mglq0_spec-patch_rev1.patch` = `git diff HEAD` on base
  `1ca4b3d` (new files intent-to-added): 237486 bytes,
  sha256 `562cc2ef00f08ffa10268bfc892b54a069876fdec88e4309eb7d362dcaf96f67`,
  patch-id `46aacb6df526ae462e3b430bec5876ff82ba34a8`. Same 24-file set as
  the rebased rev-3 patch; no new files added by this revision.
- Curator delta: EMPTY — curator worktree
  (`task-board/story/STORY-260910-2qmrb8` at `640a9df`) `git status
  --porcelain` clean.
- No stray files: spec worktree shows only the 24 intended paths, zero
  untracked entries; the `tools/__pycache__` left by the validate run was
  removed.

No logbook entry: nothing met the bar (no anomaly, no regression, no
decision beyond the brief — the regenerate-check staging step and the
export-subst fixture artifact are both already documented in the rev-1
evidence and rev-3 verdict respectively).
