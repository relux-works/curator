# Review verdict — TASK-260910-33j1hu revision 1

Verdict: changes_requested. Route to `to-dev`; do not accept CR revision 1.

## Candidate and scope

Reviewed curator-spec Story worktree `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-35tbgb/worktree`, HEAD `dced9b8317e0e8af79edf2d0539b32bd22b6c85b`, against HEAD as the round-specific brief requires (origin/main moved). Attached spec patch and `git diff HEAD` are byte-identical. Both stable patch IDs: `7170e7475592bfdf9e8390c90bb0d474838032d4`. Diff SHA256: `bbb11991fa639e516bbad0cb7d6d87acb04bd2cbbd4773ba2c5af54cfb4d955e`.

The curator CR has an empty repository delta. That is appropriate for this leaf's repository boundary: all nine changed files are in curator-spec, represented by the attached spec patch and evidence. It is neither an absent deliverable nor proof of acceptance. No product implementation, schema, client-bootstrap interchange, proposal, tag, or release publication was changed. Spec generation/validation tooling is explicitly authorized. Profile revision is not yet merged; acceptance would only authorize integration, not claim landing.

Read campaign rules, producer brief, producer evidence and patch, and audit P2 (`docs/security-audit-2026-09.md:188`). No candidate code edited. No index changes or commits made.

## Findings and required corrections

### F1 — medium: required checkpoint scenario coverage is not pinned by the validator

`tools/validate.py:2602` requires seven names, then `:2609` computes expectations solely from each case's current inputs. There is no assertion that a name still represents its mandatory scenario. At `:2547` the oracle can correctly recompute a passing answer after the negative scenario itself has disappeared.

Independent narrowing probe replaced each of the four negative cases, individually, with equal-consistent inputs and passing output while retaining its required name. All four survived both the checkpoint gate and the full `validate.main()` entry point: **0/4 branch-replacement mutants rejected; 4/4 accepted**. The published corpus currently represents 7/7 requested branches; this finding concerns the validator's inability to preserve that coverage as vectors evolve. Existing tests at `tools/test_validate.py:3095` onward reject incorrect output for the original input, but do not assert that the validator rejects a self-consistent replacement scenario. The full test suite may detect such replacements through its assumptions; this is not a claim that the entire suite admits them.

Correction: pin each required scenario's discriminating input predicates (configured state, signature state, version relationship, equal-body or prefix relation as applicable), or independently assert complete semantic branch coverage in addition to required names. Add negative tests replacing scenarios with internally consistent passing cases, keeping names, and require the production validator to reject them. Preserve the current positive controls. Republish the spec patch and evidence for another review.

Probe runs use an in-memory `load_json` substitution only for registry-service.json, leaving all repository bytes untouched; `validate.main()` still invokes its ordinary gates. This isolates semantic pinning from checksums, which a normal regeneration updates. This does not prove a stale manifest would be accepted. Reproduction script and output are attached as task outcomes.

## Per-item review

Paths below are relative to the curator-spec worktree.

| Requirement | Evidence / exact quote | Result |
|---|---|---|
| Startup enforcement and ordering | `profiles/registry-service.md:167`: “The normative enforcement point ... is the startup checkpoint comparison”; `:170`: “AFTER the section 5 startup integrity verification and BEFORE it binds its listener or reports ready” | Pass |
| Signed object and accepted keys | `:160`: “signed `registry-snapshot-v1.schema.json` object”; `:172`: “accepted signing keys (the staged rotation set)” | Pass |
| Signature-first refusal | `:173`: “Signature verification comes first”; `:174`: “`checkpoint_signature_invalid` without comparison” | Pass |
| Below, equal and above comparisons | `:176`: “live `version`/`log_size` below”; `:177`: “equal version with a different `head`, `merkle_root`, or `log_size`”; `:180`: “live log reproduces the checkpoint boundary at its `log_size` (head and Merkle root at that prefix)” | Pass text |
| Refusal consequences | `:185`: “MUST NOT serve the same canonical URL”; `:186`: “disabled ... MUST NOT truncate or repair history automatically” | Pass |
| No-checkpoint and configured posture | `:190`: “starts with no behavior change”; `:192`: “`checkpoint_not_configured`”; `:193`: “compared boundary (`version`, `log_size`, `head`) and the comparison outcome” | Pass |
| Offline procedure remains | `:195`: “offline `verify-backup`-style comparison”; `:197`: “not the readiness gate” | Pass |
| Exactly four diagnostics | `:205–208` table lists `restore_below_checkpoint`, `restore_inconsistent_with_checkpoint`, `checkpoint_signature_invalid`, `checkpoint_not_configured`; `:210`: “No other restore-checkpoint diagnostic exists.” Same spelling in prose, vectors, generator, validator and CHANGELOG | Pass |
| Cross references | `:144` startup integrity ordering; `:251` refused startup comparison non-ready; `:273` threat model startup comparison; `:296` conformance “startup checkpoint comparison including the no-checkpoint posture (section 6)” | Pass |
| Protocol pointer | `protocol/registry.md:180`: “server-side counterpart is the registry-service profile’s startup checkpoint comparison” | Pass; no client-bootstrap definition |
| Frozen health schema | No schema path in diff; `profiles/registry-service.md:254`: “`health-response-v1` schema is unchanged” | Pass |
| Vector branches and validator pins | `conformance/v1/vectors/registry-service.json:25` adds seven cases; generator `tools/generate-vectors/main.go:620`; validator invoked from main at `tools/validate.py:5891` | Seven current branches present; pinning fails F1 |
| Prior vectors unchanged | Removing the newly inserted checkpoint_cases byte span yields exactly HEAD's entire registry-service.json bytes, including recovery_cases | Pass |
| Manifest and rc.9 pins | `conformance/v1/manifest.json:4260` registers same vector path with new digest; two manifest digest pins updated in `release/1.0.0-rc.9.json` | Pass |
| CHANGELOG and rollout | `CHANGELOG.md:10`: “R3/P2”; `:35`: “Rollout is direct, not warn-first”; `:41`: “STORY-260910-35tbgb” | Pass |
| Knob/lock tables and reporting | No new knob or lock; this profile has no §12.1/§12.2. Startup/audit posture in §9 is the settled reporting surface; no manager env-status change requested | Not applicable / posture present |
| Patch and evidence outcomes | Retrieved named spec patch and producer evidence through task-board resource get | Pass |
| Architectural fit | Uses existing signed snapshot, startup integrity check, and staged keys; offline restore workflow retained | Pass |
| Docs/code consistency bounds | Spec generator and vector oracle reviewed. Product enforcement is intentionally deferred to TASK-260910-1ny7yl; this review makes no runtime enforcement claim | Bounded |

## Independent validation

Shell: zsh, from the exact Story worktree. Independently ran:

```sh
set -o pipefail
PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" make validate 2>&1 | tee /tmp/TASK-260910-33j1hu-review-validate.log
```

Observed pipeline exit **0**. All three unmodified Makefile gates passed; no re-sign or substitute binary required in this review. The process was kept alive and observed through bounded sequential tool calls until completion (no background abandonment). Test progress dots elided below:

```text
python3 tools/validate.py
validated 62 schemas and 1071 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
Ran 310 tests in 619.876s
OK
go test ./tools/...
ok github.com/relux-works/curator-spec/tools/generate-vectors 0.723s
EXIT=0
```

No producer test result was substituted for these independent results. Passing baseline gates do not resolve F1.

Full validator narrowing output:

```text
checkpoint-equal-inconsistent replaced with equal-consistent while preserving name: main exit 0 validated 62 schemas and 1071 vector files
live-below-checkpoint replaced with equal-consistent while preserving name: main exit 0 validated 62 schemas and 1071 vector files
live-above-prefix-mismatch replaced with equal-consistent while preserving name: main exit 0 validated 62 schemas and 1071 vector files
checkpoint-signature-invalid replaced with equal-consistent while preserving name: main exit 0 validated 62 schemas and 1071 vector files
```


Regeneration ran in a disposable copy of the exact candidate to preserve read-only review. `make regenerate-check` ran its ordinary `go run ./tools/generate-vectors -root .` successfully. Its subsequent Git comparison failed because the copy deliberately has no .git directory (make exit 2 / git exit 129). This is a review-harness limitation, not a candidate gate regression, and is not represented as a green regenerate-check. Independent before/after SHA256 comparison of all 1,077 conformance/release files found zero changes, establishing successful deterministic regeneration. No host signing workaround was needed for this generator run.

`git diff HEAD --check`: exit 0.

No LOGBOOK.md write: explicitly prohibited by campaign rules. Findings are persisted here and in board notes instead.
