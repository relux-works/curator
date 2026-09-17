# TASK-260916-y4sa6s review verdict — revision 3

Verdict: **accepted**. Accept CR-TASK-260916-y4sa6s-3 revision 3 and route to **integrating**, not done. R1's remaining conformance gap is closed; R2–R7 remain closed. No candidate files or original Git index modified; no commit or commit_ack supplied.

## Candidate identity, scope, and empty curator delta

Reviewed the curator-spec Story worktree `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-ioemse/worktree`, the campaign and producer/rework/review briefs, producer evidence including Revision 3, revision-2 and revision-3 patches, prior verdict, and audit E1 / Appendix B. Applied project-management and its negative-evidence reference. Architecture diagrams are not needed for this review.

`git diff origin/main | git patch-id --stable` and the attached revision-3 patch both yield `af964a62a8b2998fb54fff35226ecaa00ddf1063`. Their SHA-256 is `77336b0d29a22ed83453d622dc3333a57104b24d0576c6d5ebe3f92e9699cd85`. New paths already carry intent-to-add entries; no index mutation was needed. `git diff --check` exits 0.

Comparing every patch section of revision 2 with revision 3 shows exactly five changed files: `conformance/v1/vectors/environments-source-signers.json`, `conformance/v1/manifest.json`, `release/1.0.0-rc.9.json`, `tools/validate.py`, `tools/test_validate.py`. All other candidate file sections are byte-identical. This preserves the prior normative, schema, CLI, generator and R2–R7 closures. Both frozen `context-lock-v1` and `agent-mcp-v1` schemas remain unchanged relative to story HEAD. Of 30 existing vector files, 29 are byte-identical; only the explicitly permitted regenerated manager-config-v2 vector differs. Existing config schema-case regeneration is the accepted exception established in revision 2.

The exact curator CR base `aa46ecd80ad0b83853586454723ea99fb76977a9` to candidate `45aae0e7ba8fdc58a42eb553ded0726f3c68b8cf` has an empty diff. Current curator `git status --short` is empty. **No curator repository change is the correct result:** this leaf delivers normative work in the separate curator-spec Story worktree and attached spec patch. Manager implementation belongs to TASK-260916-1zgucp. No LOGBOOK was modified; this resource and board notes hold the review record under the campaign prohibition.

Integration note retained from revision 2: story HEAD is `23dafa798fa80fc2591ddb287c1c6345e2715b3b`, while current origin/main is `dced9b8317e0e8af79edf2d0539b32bd22b6c85b`. The requested patch against origin/main includes upstream reversals; integration must preserve newer main changes and integrate the task delta relative to its recorded story base. Acceptance does not authorize overwriting main with this old worktree snapshot.

## Revision-3 closure and independent attack

| Requirement | Exact evidence | Verdict |
|---|---|---|
| Real declaration snapshots, optional presence retained | tools/validate.py:5268: “presence-preserving declaration object”; :5321 constructs the real manifest; :5327 calls `mcp_validator.iter_errors(manifest)`; :5466 constructs `Draft202012Validator` with the schema registry. | Pass |
| Per-transport consistency retained | tools/validate.py:5297 “a stdio declaration carries no url”; :5306 “an http declaration carries no command or args”. Null padding is not admitted. | Pass |
| Complete canonical comparison unchanged | tools/validate.py:5352 `e1_mcp_changed` compares `ccj1_bytes(old) != ccj1_bytes(new)`; protocol/environments.md:1841 “no field is narrowed out” and “absent ... differs from present”. Neither comparator nor normative prose changed in revision 3. | Pass |
| Eight optional-presence cases | vector:1095, :1139, :1183, :1227 and their confirmed partners: absent→present selector, present→absent selector, absent→empty env_names and reverse. All 8/8 carry A migration hint; 4/4 unflagged refuse under B and 4/4 flagged proceed. | Pass |
| Prior URL, selector, order coverage retained | `changed-mcp-url`, `changed-mcp-selector`, confirmed partners and `mcp-env-names-reorder-triggers` remain in the vector inventory and pass recomputation. | Pass |
| Narrowing regression | tools/test_validate.py:2406 `test_absent_optional_padding_narrowing_fails_on_new_cases`; independent reviewer monkeypatch replaces the comparator itself, then drives `validate_environments_source_signers_vectors`. | Pass: narrowed gate rejected |
| Real entry wiring | tools/validate.py:6405 registers the E1 gate in main; all three snapshot call sites (delta, --all, reinstall) pass the real validator. | Pass |

Independent replay observed: the published gate passes 65/65 cases; replacing only the comparator with one that inserts absent optional arrays as `[]` is rejected on `absent-env-names-to-empty` with “expected is not the section 9.2 delta rule”. This directly replays the revision-2 attack, without modifying candidate bytes. Direct schema validation replay passed 36/36 MCP snapshots from top-level delta/reinstall fixtures; the full gate additionally traverses nested --all profiles. An invalid empty selector inserted into a fixture is rejected specifically by agent-mcp-v1 schema validation. Four transition pairs / eight cases close the named absence blind spot.

## Standing requirement assessment

The following unchanged normative/schema references and their exact quotations were cross-checked against the current candidate. Prior R2–R7 closure evidence is retained because those bytes are unchanged; the full suite is independently rerun below. Validator/test line numbers in this standing table refer to revision 2 where the revision-3 insertion shifts later lines; the new closure table above uses current lines.


All file references below are relative to the spec worktree; quotes are exact excerpts.

| Requirement | Evidence and assessment |
|---|---|
| Dated Decision amendment, retained history, §8 pointer | decisions/0012-context-packages-and-semver-locks.md:69 “Amendment (2026-09-17, E1)”; :100 “(Decision 8)”; :604 amendment marker. Pass. |
| Signer rule in amendment | Decision:92 “before the candidate enters the lock — either suffices”. Pass. |
| latest residual / strict tags | Decision:122 “`latest` stays `*`”; environments.md:254 “policy covers a *moved* tag, not a *new* one”. Pass. |
| §1.3 lock invariant | environments.md:194 “The lock is a record, not a signature”; resolution pointer added, context-lock-v1 schema diff empty. Pass. |
| §1.4 verification | environments.md:328 “before the candidate enters the lock”; :331 “signature of the commit it peels to MUST verify”; :332 “only the commit signature can satisfy the check”. Pass. |
| Fail closed / path / no downgrade | environments.md:335 “MUST leave the old lock in place”; :348 “MUST NOT silently select a lower candidate”; :351 “are never verified”. Pass. |
| Three resolution diagnostics | environments.md:109–111 and :339–345 spell `context_source_unsigned`, `context_source_signer_rejected`, `context_source_signers_missing`, matching validator constants and vectors. Pass. |
| Delta before publish, closed line forms | environments.md:1787 “resolved-version delta of the candidate lock against the old lock”; :1813–1815 added/removed/moved grammar includes version and pin. Pass. |
| System inventory trigger | environments.md:1832 “(`path`, `environments` selector, bytes)”; “admission under section 3 does not narrow the trigger”. Pass. |
| R1 canonical MCP trigger | environments.md:1837 “the CCJ-1 bytes”; :1841 “Any byte difference”; tools/validate.py:5325 uses `ccj1_bytes`. Pass. Absence coverage is now closed by the revision-3 replay below. |
| R1 URL / selector / reorder | vector:985/1007 `changed-mcp-url` and confirmed; :1029/1051 selector and confirmed; :1073 `mcp-env-names-reorder-triggers`. Negative tests test_validate.py:2291–2322. Pass these branches. |
| Two rollout revisions + migration hint | environments.md:1848 “Revision A (warning release)”; :1853 “Revision B (flip release)”; :1861 “MUST ship revision A before revision B”. A hint names B refusal and flag; e1_revision_outcomes:5380 pins it. Pass. |
| R2 reinstall flag | environments.md:1744 “`profile install` accepts”; :1869 “identical per-invocation semantics”; cli/curator.md:30 includes `[--confirm-system-delta]`; vector:1371/1395 flag/no-flag reinstall cases. Pass. |
| Per-run / --all / no config consent | environments.md:1862 “no configuration knob may pre-confirm it”; :1863 “confirms every profile of the run”; two --all vectors. Pass. |
| Update diagnostic table | environments.md:2147–2148 has `profile_update_system_delta` with migration hint and `profile_update_confirmation_required`. Pass. |
| §12 signer posture | environments.md:2565 “signer-verification posture per lock member's source”; three states and machine require value; :2599 “`unknown` when it cannot”. Pass, including settled unknown behavior. |
| R7 confirmation posture | environments.md:2606 “`A-warning`”; :2608 “`B-flip`”; behavior names warn/proceed versus refusal unless flagged, no preconfirmation. Two vector rows :1420/1427. Pass; B-flip consistently derives from the rollout's “flip release” label. |
| §12.1 knobs/defaults | environments.md:2657/2658 `source_signers.<source>` default `{}` and `require_source_signers` default `false`; schemas and generated defaults match. Pass. |
| Closed entry shape | environments.md:2671 “`type` exactly `ssh` or `gpg`”; :2678 prohibits cross-fields; manager schema:352 closed union. Pass. |
| R3 fingerprint | manager schema:372–374 minLength/maxLength 40 plus uppercase pattern; system schema:58 references it. New newline invalid cases in both configurations, real-entry regression test test_validate.py:3311. Pass. |
| R6 SSH identity | environments.md:2679 “key type plus its base64 key material”; :2680 “trailing comment is not part”; oracle:5024 compares first two tokens; vector:419/447 accepts changed comment and refuses different material. Pass. |
| §12.2 fleet lock/direction | environments.md:2699 admits both keys; :2713 “only to `true`”; :2717 machine list “is ignored” for system-named source; unnamed source takes machine list. profiles/manager.md:57/67/73 agrees. Pass. |
| Both config schemas and lockable set | manager schema:602/:616 knobs; system schema:26/:27 locked names and :58/:59 properties (`enum: [true]`). No extra configuration preconfirmation. Pass. |
| R4 revision forgery | validator:5094 forbids tag signature for revision; test_validate.py:2354 replays forgery. Reviewer separately replayed it through published gate and observed rejection. Pass. |
| R5 empty curator delta | Exact CR object diff and current curator status both empty, no LOGBOOK change. Pass. |
| §13 conformance + manifest | environments.md:2780 names `vectors/environments-source-signers.json`; :2802 requires warning/migration hint; manifest registers file and new schema cases. Pass; R1 coverage now closed. |
| Vector branches | Published gate recomputes 65/65 E1 cases: 19 verification, 6 merge, 5 posture, 29 delta, 2 --all, 2 reinstall, 2 confirmation posture. Eight new cases cover four optional-presence transitions with and without confirmation. |
| Schema cases retain purpose | 101/101 changed preexisting config cases are identical semantically after removing only the two E1 knobs/lock entries. Pass under explicit regeneration exception. |
| Existing vectors | 29/30 existing files byte-identical against story HEAD; sole changed file manager-config-v2.json adds knobs/new cases under exception. Other apparent origin/main reversals are upstream baseline advancement. |
| CHANGELOG | CHANGELOG.md:10 “E1: per-source signer allowlist”; :32–43 names both revisions, canonical trigger, reinstall, identity, residual. Pass. |
| Outcome resources | Producer evidence and full revision-3 patch downloaded through resource CLI; this task-scoped verdict is attached before routing. |


## Validation and bounds

Spec tooling is in scope; no manager implementation was changed. These vectors model signature observations, not real Git cryptographic verification; downstream manager enforcement remains outside this specification leaf. No new blocking finding.

Independent probe command (zsh, requested venv on PATH, `PYTHONDONTWRITEBYTECODE=1`): `python3 /tmp/e1-review-r3/probe.py`, exit 0. Comparator replacement used `unittest.mock.patch.object(validate, 'e1_mcp_changed', padded)` where `padded` deep-copies both declaration objects, inserts `env_names=[]` and `environments=[]` only when absent, and compares `ccj1_bytes`. It then calls the published complete E1 gate. This mutates only in-process Python state.

```text
Published E1 gate: PASS; cases: 65
Padding comparator REJECTED: source-signers case absent-env-names-to-empty: expected is not the section 9.2 delta rule
Direct snapshot schema replay: 36 / 36
absent-selector-to-present {'lines': ['lock-delta moved mcp figma-devmode 1.2.0 → 1.3.0 commit:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee → commit:ffffffffffffffffffffffffffffffffffffffff'], 'trigger': ['figma-devmode'], 'revision_a': {'diagnostic': 'profile_update_system_delta', 'hint': 'revision B refuses with profile_update_confirmation_required unless --confirm-system-delta is given', 'proceeds': True, 'lock_published': True}, 'revision_b': {'diagnostic': 'profile_update_confirmation_required', 'proceeds': False, 'lock_published': False}}
absent-selector-to-present-confirmed {'lines': ['lock-delta moved mcp figma-devmode 1.2.0 → 1.3.0 commit:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee → commit:ffffffffffffffffffffffffffffffffffffffff'], 'trigger': ['figma-devmode'], 'revision_a': {'diagnostic': 'profile_update_system_delta', 'hint': 'revision B refuses with profile_update_confirmation_required unless --confirm-system-delta is given', 'proceeds': True, 'lock_published': True}, 'revision_b': {'diagnostic': None, 'proceeds': True, 'lock_published': True}}
present-selector-to-absent {'lines': ['lock-delta moved mcp figma-devmode 1.2.0 → 1.3.0 commit:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee → commit:ffffffffffffffffffffffffffffffffffffffff'], 'trigger': ['figma-devmode'], 'revision_a': {'diagnostic': 'profile_update_system_delta', 'hint': 'revision B refuses with profile_update_confirmation_required unless --confirm-system-delta is given', 'proceeds': True, 'lock_published': True}, 'revision_b': {'diagnostic': 'profile_update_confirmation_required', 'proceeds': False, 'lock_published': False}}
present-selector-to-absent-confirmed {'lines': ['lock-delta moved mcp figma-devmode 1.2.0 → 1.3.0 commit:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee → commit:ffffffffffffffffffffffffffffffffffffffff'], 'trigger': ['figma-devmode'], 'revision_a': {'diagnostic': 'profile_update_system_delta', 'hint': 'revision B refuses with profile_update_confirmation_required unless --confirm-system-delta is given', 'proceeds': True, 'lock_published': True}, 'revision_b': {'diagnostic': None, 'proceeds': True, 'lock_published': True}}
absent-env-names-to-empty {'lines': ['lock-delta moved mcp figma-devmode 1.2.0 → 1.3.0 commit:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee → commit:ffffffffffffffffffffffffffffffffffffffff'], 'trigger': ['figma-devmode'], 'revision_a': {'diagnostic': 'profile_update_system_delta', 'hint': 'revision B refuses with profile_update_confirmation_required unless --confirm-system-delta is given', 'proceeds': True, 'lock_published': True}, 'revision_b': {'diagnostic': 'profile_update_confirmation_required', 'proceeds': False, 'lock_published': False}}
absent-env-names-to-empty-confirmed {'lines': ['lock-delta moved mcp figma-devmode 1.2.0 → 1.3.0 commit:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee → commit:ffffffffffffffffffffffffffffffffffffffff'], 'trigger': ['figma-devmode'], 'revision_a': {'diagnostic': 'profile_update_system_delta', 'hint': 'revision B refuses with profile_update_confirmation_required unless --confirm-system-delta is given', 'proceeds': True, 'lock_published': True}, 'revision_b': {'diagnostic': None, 'proceeds': True, 'lock_published': True}}
empty-env-names-to-absent {'lines': ['lock-delta moved mcp figma-devmode 1.2.0 → 1.3.0 commit:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee → commit:ffffffffffffffffffffffffffffffffffffffff'], 'trigger': ['figma-devmode'], 'revision_a': {'diagnostic': 'profile_update_system_delta', 'hint': 'revision B refuses with profile_update_confirmation_required unless --confirm-system-delta is given', 'proceeds': True, 'lock_published': True}, 'revision_b': {'diagnostic': 'profile_update_confirmation_required', 'proceeds': False, 'lock_published': False}}
empty-env-names-to-absent-confirmed {'lines': ['lock-delta moved mcp figma-devmode 1.2.0 → 1.3.0 commit:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee → commit:ffffffffffffffffffffffffffffffffffffffff'], 'trigger': ['figma-devmode'], 'revision_a': {'diagnostic': 'profile_update_system_delta', 'hint': 'revision B refuses with profile_update_confirmation_required unless --confirm-system-delta is given', 'proceeds': True, 'lock_published': True}, 'revision_b': {'diagnostic': None, 'proceeds': True, 'lock_published': True}}
Invalid selector schema probe REJECTED: source-signers case absent-selector-to-present: an mcp snapshot is a declaration valid under agent-mcp-v1 ({'transport': 'stdio', 'command': 'npx', 'args': ['-y', 'figma-developer-mcp', '--stdio'], 'env_names': ['FIGMA_API_KEY'], 'environments': []} is not valid under any of the given schemas)
```

Generator proof: copied all tracked candidate files (including intent-to-add paths) into `/tmp/e1-review-r3/regen`, initialized a disposable Git index and staged that candidate as the comparison baseline, then ran literal `make regenerate-check`. No candidate index or files were written, no commit made. This is the requested equivalent proof for an uncommitted candidate; comparing original working changes against HEAD is not a regeneration test.

```text
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
EXIT_CODE=0
```

The 101/101 preexisting modified config schema cases were independently compared with story HEAD after removing only the E1 knobs and lock entries: all remaining content is identical, preserving the named original test purposes. Existing vector byte identity was likewise checked against story HEAD, not the advanced upstream baseline.

## Completed independent validation

Shell `/bin/zsh`, spec Story worktree, exact command:

```sh
set -o pipefail
PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" PYTHONDONTWRITEBYTECODE=1 make validate > /tmp/e1-review-r3/validate.log 2>&1
```

```text
python3 tools/validate.py
validated 60 schemas and 1089 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
....................................................................................................................................................................................................................................................................................................................................................
----------------------------------------------------------------------
Ran 340 tests in 543.934s

OK
go test ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	6.542s
EXIT_CODE=0
```

All three gates were independently rerun to completion, not accepted from the producer transcript. No verification process remains running. Before recording the verdict, `task-board spawn goal "$TASK_BOARD_RUN_ID"` reported `Active Goal: none (run is not goal-bound)`. The live checklist is fully checked. Reviewer leaves landing/done to the authorized producer integration lifecycle.

Final candidate diff SHA-256 rechecked after validation: `77336b0d29a22ed83453d622dc3333a57104b24d0576c6d5ebe3f92e9699cd85`; unchanged.
