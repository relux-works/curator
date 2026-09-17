# TASK-260916-2rnkei — review verdict, revision 1

Verdict: **changes_requested**. Route to **to-dev** for spec/conformance rework, followed by another independent review. No human-only decision or external blocker is involved. Reviewed 2026-09-17.

## Candidate and scope

Read campaign producer rules, producer brief, producer evidence and patch, and audit E3 / Appendix B. Applied the project-management skill as the assigned reviewer. No candidate code, index, branch or repository content was changed by this review.

Spec worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-1i1gfo/worktree`; HEAD `684c9f1324d46b4938b2e5943f20c89e27971ec8`.
Attached patch and `git diff HEAD` both have stable patch ID `45c23d6c8685009a1af62611e22fd0ffa7ef9862`. The complete patch bytes also match, SHA-256 `21be00ad86f4fbcdeb0f0be71b035e140dbe72774f574b37116d8a9b0be4f8d3`. New vector already had intent-to-add, so no index mutation was necessary. Do not compare against moving origin/main.

The curator CR delta from `0f0ae61766026cf2389ec4b91c09134cdb8aac62` to `3c30a7f627678f20a8c70ce0f776d7ef7ef7260f` is empty, and curator status is clean. This is the correct repository outcome for this leaf: its deliverable lives in curator-spec and the task carries the real spec patch as an outcome. Empty curator delta is neither the reason for rejection nor evidence that the spec is merged. No LOGBOOK.md edit is needed or allowed.

Changed spec paths are exactly CHANGELOG.md, protocol/environments.md, the marker schema, the new vector family, conformance manifest and the regenerated rc.9 pins. No product implementation, proposals 0014–0018, E5 or S5 edits. Validation tooling is absent too, which is a required omission to correct (F1), not a scope success.

## Findings requiring correction

### F1 — High: required semantic gate and replacement tests omitted

Campaign rule 7 requires `tools/validate.py` to pin each named scenario to its discriminating inputs, with replacement tests in `tools/test_validate.py`. The producer evidence §5 explicitly says there is no semantic gate and calls it optional follow-up. That contradicts the task's explicit requirement; a role boundary cannot waive it.

`tools/validate.py:6533` main registers the existing environment families but no codex-seed family. No codex-seed references exist anywhere under tools. The new family is digest-registered at `conformance/v1/manifest.json:4284`, but a digest is not semantic coverage.

Independent negative probe in a disposable copy: replace the contents of `b-strips-servers-keeps-rest` (`conformance/v1/vectors/environments-codex-seed.json:24`) with the internally consistent `b-without-servers-no-warning` case (`:50`), retaining the original name. Regenerate manifest/release pins using the repository generator, then invoke the real entry point `python3 tools/validate.py`. It incorrectly exits **0**, printing `validated 62 schemas and 1094 vector files`. Measured replacement rejection: **0/1**. No exhaustive mutation coverage is claimed.

Required correction: add a semantic validator reached from main, pin all required provisioning/posture scenarios to their distinguishing inputs, and add negative replacement tests that preserve names while collapsing branches. Validate expected names, retained TOML members/values, diagnostics, revision, marker shape and posture consistently. Add generated positive/negative marker schema cases for the added closed record; the existing schema cases contain no new-field coverage. Coordinate the appropriate producer role if necessary, but do not defer required conformance work to manager implementation. Re-run gates and attach a new patch/evidence revision.

### F2 — Medium: revision-A homes lose the required upgrade repair warning

The settled brief requires the seed revision to identify a home provisioned under an older revision and report `mcp_seed_unstripped` with re-provision guidance. Revision A deliberately inherits servers, and an upgrade to B deliberately leaves those home bytes intact.

Instead, `protocol/environments.md:1309` restricts the warning to a home “whose marker predates the rule”; §7.7 at `:1437` further says “marker lacks the seed record”, §8.2 at `:1631` binds it to absence, and §12 at `:2635` repeats the restriction. Thus an A record with inherited servers under a B manager does not get the stipulated repair warning. The 7 posture cases (`conformance/v1/vectors/environments-codex-seed.json:102`) contain A/A, B/B and missing-record combinations, but no B-manager/A-home case: migration-branch coverage **0/1**.

Required correction: explicitly distinguish manager-shipped revision from the home's recorded seed revision; retain and list the inherited names as ungoverned, emit `mcp_seed_unstripped` with the re-provision hint for an older unstripped A home under B, and preserve its bytes. Keep text, diagnostics, status and CHANGELOG synchronized. Add a pinned B-manager/A-home vector and replacement test. Clarify §7.8's “under revision B the base carries none” as applying to B-provisioned homes, not all homes used by a B manager.

## Per-deliverable assessment

Line numbers below refer to the candidate. `environments.md` means `protocol/environments.md`.

| Brief item | Evidence / exact quote | Assessment |
|---|---|---|
| §7.4 seed row and evidence confidence | environments.md:1273: “revision A copies it whole”; “revision B copies every top-level member except `mcp_servers`”; “remaining member shapes are **docs-confidence**” | Pass for the seed rule and evidence posture. |
| Two labelled revisions, A first | environments.md:1282: “a manager MUST ship revision A before revision B”; :1284 and :1291 label warning/flip releases | Pass. |
| A inherited names and migration | environments.md:1285: “MUST emit `mcp_native_servers_ungoverned`”; :1288: “declare the server in the profile's MCP set, or accept the loss” | Pass. |
| B strips all MCP subtree, reports once | environments.md:1292: “`mcp_servers` table and every `mcp_servers.*` sub-table removed”; :1295: “MUST report the stripped names once” | Pass for newly provisioned homes. |
| Names-only provisioning snapshot, empty names | environments.md:1304: “names only, never server commands or env values”; :1306: “a native file with no `mcp_servers` entries warns nothing” | Pass. |
| Existing homes retain bytes, old-rule repair | environments.md:1308: “an existing home keeps its bytes”; :1309 limits warning to a marker predating the rule | Partial; F2. |
| §7.8 all four adapters | :1460 “`--strict-mcp-config` disables every other MCP configuration”; :1461 “`-p curator-mcp` layers the profile set over the seeded base”; :1462 “project-level servers remain”; :1463 “none: no channel” | Four adapter rows present; distinguish B-provisioned homes as F2 requests. |
| §7.7 closed diagnostic list | environments.md:1435–1437 contains exactly `mcp_native_servers_ungoverned`, `mcp_native_servers_not_inherited`, `mcp_seed_unstripped` | Exact spellings consistent; F2 predicate correction required. |
| §8.2 closed marker record | environments.md:1625: “closed object `{ revision, native_mcp_servers }`”; :1628: “ascending-byte-order snapshot” | Shape is explicit; semantic enforcement/negative coverage missing (F1). |
| Marker schema and version discipline | schemas/v1/agent-environment-marker-v1.schema.json:89 requires both members, :91 enum A/B, :98 `additionalProperties: false`, :124 forbids non-managed usage | Additive change is within this unreleased environments schema; schemas/v1/README.md distinguishes the frozen earlier objects and history shows this marker was already amended in fcdb9ba. No demonstrated frozen-surface violation. New schema-case coverage still required. |
| §12 posture and currency | environments.md:2631 “active codex-seed revision”; :2633 per-home names; :2660–2662 all three warnings “never make a row non-current” | Present; missing migration combination F2. |
| §12.1 knobs / §12.2 lock set | environments.md:2685: “no configuration knob selects the revision” | No new knob or lock key requested or added; those tables appropriately unchanged. |
| §13 conformance | environments.md:2869 names new family; :2892–2893 “MUST NOT claim revision B while still inheriting native servers” | Text present; lacks required executable enforcement (F1). |
| Vectors and registration | New family :9–101 has 7 provisioning cases, :102–202 has 7 posture cases; manifest :4284 registers it | 7/7 provisioning inputs independently TOML-parsed and expectations checked; 7/7 posture cases agree with candidate text, but F1/F2 remain. |
| Existing vectors byte-identical | Compared every HEAD-tracked conformance/v1/vectors file to working bytes: **31/31** unchanged | Pass. |
| CHANGELOG and handoff | CHANGELOG.md:10 “E3”; :14–20 both revisions/diagnostics; :23 old-home warning; :33–35 manager/README follow-up TASK-260916-33abdk | Present; F2 predicate correction needed. Patch/evidence/logbook task resources exist. |
| Scope and integration | Six spec files; no curator changes or product implementation | Correct scope location. Acceptance criteria requiring merge are not yet met; this reviewer does not integrate. |

## Validation and reproducibility

Original candidate was not mutated. Independent gate command, shell bash, original spec worktree:

```sh
set -o pipefail
PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" make validate 2>&1 | tee /tmp/TASK-260916-2rnkei-validate.log
```

```text
python3 tools/validate.py
validated 62 schemas and 1094 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
.................................................................................................................................................................................................................................................................................................................................................................
----------------------------------------------------------------------
Ran 353 tests in 634.270s

OK
go test ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	4.939s
exit 0 (make validate and pipefail pipeline)
```

All three gates were independently rerun against the original candidate. The Python gate took 634.270 seconds; output was collected through bounded sequential session reads and the process was awaited to completion, not detached. Future runs should split this slow gate into package/test subsets to remain below the headless command budget.

For read-only regeneration verification, copied candidate files to `/tmp/TASK-260916-2rnkei-review` excluding .git/.temp/.venv/__pycache__, and created a temporary Git baseline of that candidate only. This avoids rewriting candidate files and makes the existing target compare generated output against the reviewed candidate rather than its pre-change HEAD.

```text
$ make regenerate-check
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
exit 0
```

Then performed F1's replacement only in that disposable copy, ran `make regenerate` (exit 0), and the real validator (exit 0, incorrectly accepted the mutant). This proves the specific missing rejection; it does not prove manager runtime behavior. The throwaway independent fixture checks are review evidence, not a substitute for committed validator tests.

Independent structural schema probes passed **7/7**: valid A, valid B with empty names, and rejection of revision C, missing names, duplicate names, extra command field, and empty server name. These exercise the actual Draft 2020-12 schema; they are not committed regression coverage.

`git diff --check` exited 0. No validation conclusions rely solely on producer evidence. No manager runtime implementation was tested; it is explicitly out of scope.

## Review logbook and routing

2026-09-17: recorded F1's surviving branch-collapse mutant and F2's missing A-to-B migration posture here as the review logbook entry; campaign rules prohibit repository LOGBOOK.md edits. Producer's task-scoped logbook records the earlier decision to omit validation, which this review rejects under campaign rule 7.

`task-board spawn goal "$TASK_BOARD_RUN_ID"` reports no active goal; run is not goal-bound. Attach this task-scoped verdict before routing to `to-dev`. Do not call accept_cr or supply commit_ack. Rework is autonomous; no approval or blocked status is needed.
