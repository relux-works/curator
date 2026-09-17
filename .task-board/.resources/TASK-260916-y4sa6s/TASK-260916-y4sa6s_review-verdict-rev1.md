# TASK-260916-y4sa6s — review verdict, revision 1

Verdict: **changes_requested**. Route: **to-dev**. Review of CR-TASK-260916-y4sa6s-1, revision 1, plus the authoritative curator-spec candidate and its attached spec patch. No acceptance, commit, or commit acknowledgement is issued. No candidate files or original index were modified by this reviewer.

## Candidate and inputs

Read the campaign rules and producer brief, then the attached producer evidence and patch, then curator-spec `docs/security-audit-2026-09.md` E1 (lines 222–245) and Appendix B (line 423). Applied the project-management reviewer lifecycle and negative-evidence reference. The architecture-diagrams skill is not relevant to this normative-document review.

Spec worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-ioemse/worktree`, branch `task-board/story/STORY-260916-ioemse`, baseline `origin/main` 23dafa7. All file references below are relative to that worktree unless explicitly labelled curator.

The worktree diff and downloaded `TASK-260916-y4sa6s_spec-patch_rev1.patch` are byte-identical, not merely patch-id equivalent:

```
git diff origin/main | git patch-id --stable
 d348f5a78ceb5f99cbd48963c69aa479891e26a6
git patch-id --stable < /tmp/y4.patch
 d348f5a78ceb5f99cbd48963c69aa479891e26a6
SHA-256 of each patch:
2676cfb3e97c4b3fb90a49cc2a50691393b0f1cffafaf976988d0d5f9eaccc64
```

New files already had intent-to-add entries, so no `git add -N` was necessary. `git diff --check` passed. The separate curator CR was inspected by its exact base/candidate OIDs from the assignment; it contains seven added LOGBOOK.md lines, not the empty repository delta described by the review brief. See R5.

## Required corrections

### R1 — MCP trigger must compare the complete canonical declaration (high)

`protocol/environments.md:1834` still limits moved MCP declarations to “`command`, `args`, or `env_names` differ”; `tools/validate.py:5262` implements exactly that triple. This violates the settled orchestrator correction in the review brief. Independent probes passed URL-only and selector-only changes through the actual snapshot checker and delta recomputation; both returned an empty trigger.

Specify and recompute a difference in **CCJ-1 canonical bytes of the declaration object as lock/materialization reads it**, including HTTP `url` and `environments`. Add URL-change and selector-change vectors with A warning, B refusal, and confirmed acceptance. Update prose, CLI descriptions as needed, vector inventory, validator, and its negative tests together. Existing `mcp-env-names-reorder-silent` also needs reconciliation: CCJ-1 preserves array order, whereas the present oracle compares env_names as sets. Do not silently retain the old set comparison under a canonical-byte rule.

### R2 — Reinstall must accept the same per-run flag (medium)

`protocol/environments.md:1860` explicitly says “`profile install` takes no `--confirm-system-delta`”; `cli/curator.md:30` omits it. The review brief settles the opposite. Admit the flag on the reinstall path with identical per-invocation semantics, update §9.2 and the install CLI row, and pin the behavior with a vector or CLI row. Keep no-config-preconfirmation and `--all` semantics. Also make revision A's diagnostic carry the campaign-required migration hint for the forthcoming flag requirement; the current A paragraph only requires naming members and proceeding.

### R3 — Fingerprint schema admits a trailing newline (medium)

`schemas/v1/manager-config-v2.schema.json:372` uses `^[0-9A-F]{40}$` without an exact length constraint; the system schema references it at line 58. With the repository's actual Draft202012Validator and registry, both schemas accept 40 `A` characters followed by LF (41 characters). That contradicts `protocol/environments.md:2658`, “40 uppercase hex.”

Constrain the exact length/end of string and add generated negative schema cases for the newline form in both configurations; retain the valid 40-character control. Verify the real schema-validation entry rejects the malformed instance, rather than asserting only the regex spelling.

Independent output:

```
manager-config-v2.schema.json fingerprint chars 40 errors 0
manager-config-v2.schema.json fingerprint chars 41 errors 0
system-config-v2.schema.json fingerprint chars 40 errors 0
system-config-v2.schema.json fingerprint chars 41 errors 0
```

### R4 — Vector gate accepts a tag-only signature for an exact revision (medium)

`protocol/environments.md:332` says “A `revision` selection carries no tag, so only the commit signature can satisfy the check.” In `tools/validate.py:5061`, the fixture checker checks that `tag` is null for a revision, but does not prohibit `tag_signature`; the subsequent signature OR admits it.

Reproduction without changing files: load the published vector, locate `revision-selection-verifies-commit`, move its `commit_signature` object to `tag_signature`, set `commit_signature` to null, and call `validate_environments_source_signers_vectors(vector=vector)`. The complete vector gate accepts the unchanged expected acceptance, although the only signature is now on a nonexistent tag.

Require internally consistent signature evidence for the selection form and add a negative test/vector that cannot satisfy an exact revision with tag-only evidence. Preserve valid commit-signed revisions and tag-OR-commit semantics for tag selections.

```
FULL VECTOR GATE ACCEPTS: revision selection; tag=null; commit_signature=null;
fabricated tag_signature.valid=true; expected accepted
```

### R5 — Remove the out-of-scope curator LOGBOOK delta (delivery correction)

The exact curator CR delta `aa46ecd80ad0b83853586454723ea99fb76977a9..9d2c20c1a37838ff50ef936871860ea383e80d4f` adds a seven-line E1 entry at curator `LOGBOOK.md:6`. Campaign rules explicitly prohibit LOGBOOK.md edits and require work only in the curator-spec Story worktree. The review brief also expects an empty curator delta. Remove this task's LOGBOOK addition from the next candidate through the producer's normal rework flow; retain findings in task resources/notes. This reviewer did not alter it. No implementation code was changed.

## Deliverable-by-deliverable review

| Requirement | Evidence (exact excerpts and file:line) | Result |
| --- | --- | --- |
| Dated Decision 0012 amendment, history retained | `decisions/0012-context-packages-and-semver-locks.md:69`: “Amendment (2026-09-17, E1)”; original passages remain with amendment markers; original decision body otherwise preserved | Pass |
| Decision signer rule | Decision:92: “before the candidate enters the lock — either suffices” | Pass |
| Decision update rule and §8 cross-reference | Decision:100: “(Decision 8)”; amendment points to environments §9.2/§9.7; Decision 8 transaction text carries item-2 marker | Pass |
| Named latest residual, moved/new distinction | Decision:122: “`latest` stays `*`”; environments:251: “policy covers a *moved* tag, not a *new* one” | Pass |
| Lock stays record, check location specified | environments:194: “The lock is a record, not a signature”; “signature check lives in resolution (section 1.4)” | Pass |
| Verification before lock, tag OR commit, every source selection | environments:328: “before the candidate enters the lock”; :331: “signature of the commit it peels to MUST verify” | Prose pass; R4 oracle defect |
| Fail closed, old lock retained | environments:334: “resolution error and the operation fails closed”; :335: “MUST leave the old lock in place” | Pass |
| Three diagnostics admitted | environments:109–111 table and :338–345 list: `context_source_unsigned`, `context_source_signer_rejected`, `context_source_signers_missing` | Exact spellings agree with oracle/vector constants |
| No silent downgrade; path exemption | environments:347: “MUST NOT silently select a lower candidate”; :350: “are never verified” | Pass |
| Print delta before publication | environments:1786: “resolved-version delta of the candidate lock against the old lock”; :1804 onward defines added/removed/moved lines with pins | Pass |
| System inventory trigger | environments:1831: “(`path`, `environments` selector, bytes)” and “admission under section 3 does not narrow the trigger” | Pass |
| Complete MCP declaration trigger | environments:1834: “`command`, `args`, or `env_names` differ” | R1 |
| Two explicit rollout revisions | environments:1842: “Revision A (warning release)”; :1845: “Revision B (flip release)”; :1853: “MUST ship revision A before revision B” | Pass ordering; add migration hint with R2 |
| Flag per invocation, no persistent preconfirmation | environments:1854: “no configuration knob may pre-confirm it”; :1855: “confirms every profile of the run” | Pass update/--all; R2 reinstall |
| Update diagnostics table | environments:2137–2138: `profile_update_system_delta`, `profile_update_confirmation_required` | Pass |
| Posture and required-signers value | environments:2555: “signer-verification posture per lock member's source”; “`enforced`”, “`unconfigured`”, “`required-missing`” | Pass |
| Honest unknown from local verification | environments:2588: “`unknown` when it cannot” reproduce verification | Accepted as settled; no retention-rule request |
| Knob rows and defaults | environments:2637–2638: `source_signers.<source>` default `{}`, `require_source_signers` default `false` | Pass map/default representation |
| Closed entry shapes | environments:2652: “`type` exactly `ssh` or `gpg`”; :2658–2659 prohibit cross-fields; manager schema:352 closed sourceSigner union | Shapes pass; exact GPG grammar R3 |
| Lockable set, direction and per-source fleet precedence | environments:2676–2697: both keys admitted; “only to `true`”; “machine file's list for that source is ignored”; unnamed sources take machine list | Pass; manager profile §1 matches |
| Two config schemas | manager schema:600/614 defines knobs; system schema:26/27 admits locks, :58/59 properties and `enum: [true]` | Pass except R3 |
| context-lock-v1 unchanged | `git diff origin/main -- schemas/v1/context-lock-v1.schema.json` empty | Pass |
| Conformance surface and manifest | environments:2757 onward names `vectors/environments-source-signers.json`; manifest:4264 registers file | Pass presence; R1/R4 coverage corrections |
| Positive/negative signer, merge, posture cases | new vector:11/39 accepted SSH tag/GPG commit, :95 unsigned, :121 wrong signer, :177 optional, :201 required missing, :483 locked overlap | 17 verification + 6 merge + 5 posture cases inspected |
| Delta and --all vectors | new vector:710 onward, “plain-version-bump-no-confirmation”, “new-system-module”, “changed-mcp-args”, “confirmed-flag-proceeds”; two --all cases | 17 delta + 2 run cases inspected; R1/R2 missing branches |
| Schema cases and generation | schema-cases/index.json:4218 onward and manifest registration; Go generator extends manager/system config fixtures | Pass registration/idempotence; add R3 negatives |
| Existing cases retain purpose | 101/101 changed preexisting config schema cases equal old JSON after removing just the two added knobs/lock entries; 48/48 existing manager vector cases remain after removing new knobs | Pass under explicit review-brief exception |
| Existing vector byte identity | 29/30 old vector files byte-identical; sole changed old file is manager-config-v2.json | Pass under regeneration exception |
| CHANGELOG Unreleased | CHANGELOG:10 “E1: per-source signer allowlist”; :32 onwards both revisions and latest residual | Pass |
| Scope | Spec docs, schemas, vectors/manifest, generator/validator and generated rc.9 manifest pins; no manager implementation or proposals 0014–0018 changed | Spec scope pass; curator CR R5 |
| Attached outcomes | Board has task-scoped spec patch and producer evidence; this review supplies the required verdict | Pass |

## Validation and evidence bounds

All reviewer Python commands used the requested repository venv via PATH. The mandated full `make validate` was independently run from the exact spec worktree in `/bin/bash`, with `set -o pipefail`. No producer test result is substituted for a reviewer run. The completed three-gate transcript and exit status are appended below. All three gates passed; the independent adversarial findings above still require rework.

Regeneration writes files, so it was run on a disposable copy of all 1,372 tracked candidate files, preserving the candidate bytes and leaving the original worktree/index untouched. No branches or commits were created. Two results are deliberately distinguished:

1. Literal `make regenerate-check`, with the original index: exit **2** because its `git diff --exit-code` sees the uncommitted revision. Before/after generated-file SHA-256 comparison found **zero changed files**. This is a baseline-sensitive red command, not evidence of generator drift; do not describe that literal invocation as exit 0.
2. Same `make regenerate-check` in that copy, with `GIT_INDEX_FILE=/tmp/y4-review-candidate.index` populated with the exact candidate as its comparison baseline: exit **0**. Output:

```
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
```

The new vector gate is registered in the real spec validation entry at `tools/validate.py:6261`. It recomputes the expected policy results rather than accepting fixture expected fields alone. An in-memory narrowing of its moved-MCP comparator to command-only was rejected by `changed-mcp-args`; the same gate admits the R4 forged revision/tag fixture. Coverage inspected: **47/47 published E1 cases** (17+6+5+17+2), **0/2 requested URL/selector-only trigger branches**, and **0/1 reinstall flag branch** in this new vector. Tests use declared signature-validity observations, not cryptographic Git verification: this review establishes spec-policy/oracle behavior only, not downstream manager signature enforcement.

All findings are ordinary producer rework. No human-only decision or external blocker is needed. The review brief's accepted unknown-posture decision remains intact. Campaign rules prohibit LOGBOOK.md edits, so this review's institutional findings are persisted in this outcome and board notes instead.

## Completed independent make validate transcript

Shell: `/bin/bash`; workdir: the curator-spec Story worktree named above.

```bash
set -o pipefail
PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" make validate
```

```text
python3 tools/validate.py
validated 60 schemas and 1087 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
.....................................................................................................................................................................................................................................................................................................................................
----------------------------------------------------------------------
Ran 325 tests in 449.836s

OK
go test ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	0.690s
EXIT_CODE=0
```

The full invocation completed within one bounded run. No background process is left pending. Final spec diff SHA-256 remains `2676cfb3e97c4b3fb90a49cc2a50691393b0f1cffafaf976988d0d5f9eaccc64`, identical to the reviewed attachment. `task-board spawn goal "$TASK_BOARD_RUN_ID"` returned `Active Goal: none (run is not goal-bound)` before recording the verdict.

Next producer: correct R1–R5, regenerate fixtures/pins, attach the revised spec patch and evidence, and hand off for another independent reviewer cycle. The test pass does not satisfy the missing settled rules or the two reproduced gate/schema defects.
