# TASK-260910-33j1hu — review revision 2

Verdict: ACCEPTED. F1 is closed; no remaining rework findings. Acceptance is of the reviewed curator-spec revision, pending producer integration, not a claim that it has already merged or that the registry implementation already implements it.

## Candidate and scope

Reviewed curator-spec Story worktree `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-35tbgb/worktree`, base `dced9b8317e0e8af79edf2d0539b32bd22b6c85b`. `git diff HEAD` is byte-identical to attached `TASK-260910-33j1hu_spec-patch_rev2.patch`; both stable patch IDs are `0bca4199e3dfa85c3c65be3e217fd37bebf81f5c`. Patch SHA256: `2f100e4da0f4bbe02b2e9649921ed9271a729f3d34b8df4bcac353e534147ece`.

The runtime CR has an empty **curator** repository delta. This is the correct outcome for this leaf: its authorized revision belongs to the separate **curator-spec** worktree and is delivered as the task-scoped spec patch. No curator implementation change is authorized; service implementation belongs to TASK-260910-1ny7yl. The empty runtime delta alone is not the acceptance evidence; the exact spec worktree, patch, and independent gates below are.

Compared revision 1 and revision 2 patch sections (excluding Git index metadata): only `tools/validate.py` and `tools/test_validate.py` changed. All other seven file patches are identical. Scope is exactly the nine expected paths: CHANGELOG, profile, protocol, generator, validator, validator tests, registry-service vectors, manifest, rc.9 pins. No schema, health-response, client-bootstrap, proposal, or product implementation edits. No candidate code edited during review. No LOGBOOK.md edit, per explicit campaign prohibition; review findings and F1 closure persist here on the board.

## Per-item review

Line references are relative to the reviewed curator-spec worktree.

| Requirement | Evidence and exact text | Result |
|---|---|---|
| Finding and enforcement point | Audit `docs/security-audit-2026-09.md:188`: “Restore-checkpoint enforcement point is ambiguous”. Profile `profiles/registry-service.md:167`: “The normative enforcement point for \"before the service becomes ready\" is the startup checkpoint comparison.” | Pass |
| Ordering and accepted keys | Profile `:170`: “AFTER the section 5 startup integrity verification and BEFORE it binds its listener or reports ready”; `:172`: “against its accepted signing keys (the staged rotation set)”. | Pass |
| Signed interchange | Profile `:160`: “The portable checkpoint interchange object is a signed `registry-snapshot-v1.schema.json` object”; complete body retained externally, including created_at. | Pass |
| Below, equal, above | Profile `:176`: “a live `version`/`log_size` below the checkpoint is refused”; `:177`: “an equal version with a different `head`, `merkle_root`, or `log_size`”; `:180`: “serves only when the live log reproduces the checkpoint boundary at its `log_size` (head and Merkle root at that prefix)”. | Pass |
| Refusal effects | Profile `:184`: “readiness fails, writes are disabled, and the service MUST NOT truncate or repair history automatically”. Missing verified suffix must be recovered or service remains unavailable. | Pass |
| No checkpoint and posture | Profile `:190`: “Without a configured checkpoint the service starts with no behavior change”; `:191`: “startup diagnostics and the audit log carry `checkpoint_not_configured`”; configured boundary and outcome at `:193`. | Pass |
| Offline procedure distinction | Profile `:195`: “An offline `verify-backup`-style comparison ... stays as the operator procedure”; `:197`: “it is not the readiness gate”. | Pass |
| Closed diagnostics | Profile table `:203`–`:210` lists exactly `restore_below_checkpoint`, `restore_inconsistent_with_checkpoint`, `checkpoint_signature_invalid`, `checkpoint_not_configured`; “No other restore-checkpoint diagnostic exists.” Identical spellings in prose, vector outcomes/posture, validator constants/oracle, and CHANGELOG `:10`. | Pass |
| Cross references 5/9/10/11 | Profile `:144` “This startup integrity verification runs BEFORE the section 6 startup checkpoint comparison”; `:251` refused comparison is non-ready; `:273` threat model names startup comparison; `:296` conformance includes “startup checkpoint comparison including the no-checkpoint posture (section 6)”. | Pass |
| Frozen health | Profile `:254`: “The `health-response-v1` schema is unchanged; a refusal is reported through readiness, not a new response member.” No schema diff. | Pass |
| Operator protocol pointer | `protocol/registry.md:180`: “The server-side counterpart is the registry-service profile's startup checkpoint comparison”; no bootstrap interchange redefinition. | Pass |
| Reporting / knobs / rollout | Profile `:260` names startup diagnostic/audit reporting. CHANGELOG `:34`: “Rollout is direct, not warn-first”. No new config/lock key; §12.1/12.2 knob rows and an env-status CLI row are not applicable to this service profile. Service posture is explicitly reported in §9. | Pass |
| Generated vectors | `conformance/v1/vectors/registry-service.json:26` has 7/7 required scenarios: above-consistent, equal-consistent, equal-inconsistent, below, above-prefix-mismatch, invalid signature, absent checkpoint. Generator `tools/generate-vectors/main.go:620` produces them. | Pass |
| Existing vectors and pins | All pre-existing JSON values equal HEAD; recovery_cases raw byte block identical. Manifest `:4261` pins registry-service SHA256 `1bacfdf34d0ed2a18e17bf6a0dd578d4e7a0da32f6023b6d61255eff1554e727`; rc.9 candidate and downstream pins updated together. Regeneration reproduces all 1197 files. | Pass |
| F1 scenario pins | `tools/validate.py:2591` adds `require_checkpoint_scenario`; production gate calls it at `:2718`, and `main()` registers the gate at `:6001`. Required names pin configuration, signature, version relation and body/prefix predicates. Four negative replacements at `tools/test_validate.py:3200` onward; positive controls retained; main-entry regression at `:3246`. Independent main-entry attack rejects 4/4 replacements. | Pass |
| CHANGELOG and artifacts | CHANGELOG Unreleased `:10` starts “R3/P2”, names story, all diagnostics, enforcement and direct rollout. Both revision patches and updated producer evidence are attached task outcomes. This verdict and independent probe/validation are additional outcomes. | Pass |

Coverage bound: these are abstract spec scenario vectors using signature/body/prefix predicates, not a running registry service or cryptographic checkpoint-file integration test. This review proves the published scenarios and validator gate, not downstream implementation conformance. The rework requested exactly this scope.

## Independent validation

Shell: zsh, with `set -o pipefail`; cwd was the curator-spec Story worktree. Command:

```
PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" make validate
```

The single `make validate` invocation was deliberately interrupted to respect the headless command time bound; its exit was **2 (SIGINT), not 0**. It had passed the first gate and emitted 53 successful test dots, with no failure, then stopped in `shutil.copytree` in release-test fixture setup. The remaining discovery suite was run in bounded sequential slices using the same Python executable and unchanged test loader. Slice 52:78 repeats one completed test conservatively; no test is omitted. This is an independently completed split execution of all Makefile gates, not a claimed one-shot make exit 0. No producer test result substitutes for an unrun test.

```
python3 tools/validate.py
validated 62 schemas and 1071 vector files
# first gate exit 0 (make advanced to unittest)

Initial unittest discovery: tests [0:53) passed before intentional interrupt.
Continuation [52:78): Ran 26 tests in 176.811s; OK; exit 0
Continuation [78:198): Ran 120 tests in 212.563s; OK; exit 0
Continuation [198:318): Ran 120 tests in 96.926s; OK; exit 0
Union: 318/318 discovered tests passed; one overlap.

go test ./tools/...
ok github.com/relux-works/curator-spec/tools/generate-vectors (cached)
exit 0
```

The full initial transcript, per-test continuation IDs/results, and Go output are attached as `TASK-260910-33j1hu_review-validation-rev2.log`; the slicing harness is attached for reproduction. Go's result is explicitly cached, not represented as a fresh uncached execution.

Regeneration ran independently in a disposable copy of the candidate excluding `.git`, `.venv`, and Python caches. A temporary Git index staged the candidate solely as the comparison baseline, without a commit. The real Makefile target ran:

```
make regenerate-check
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
REGENERATE_CHECK_EXIT 0
HASHED_FILES 1197 IDENTICAL True
```

An additional raw SHA256 comparison of every conformance/release file before and after regeneration was equal (1197/1197), independent of Git's line-ending normalization. The real Story worktree was not regenerated or staged.

Independent F1 probe: replaced each negative case wholesale with the published equal-consistent positive case, preserving only its required name, injected through the production JSON loader and called `validate.main()` with all checks active. Unmodified `main()` exit 0; all four replacements exit 1 with the appropriate scenario-pin refusal. Coverage: **4/4 rejected, 1/1 unmodified control accepted**. Probe script and complete output attached separately. It does not change any candidate file.

Final candidate SHA256 recheck equals the starting patch SHA256. `git diff HEAD --check` exits 0. `task-board spawn goal` reports this reviewer run is not goal-bound. Acceptance evidence is recorded before `accept_cr(TASK-260910-33j1hu, revision=2, evidence=TASK-260910-33j1hu_review-verdict-rev2.md)`; integration remains producer-owned. No commit acknowledgement supplied.
