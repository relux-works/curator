# TASK-261015-1yj9x7 — review revision 2

Verdict: accepted. Candidate tree: b6f8604b7d118365b8cdae1c97915493d46b64b0. Base: 3f5de6dff9aab4c037ba069323604b166bf0414e. Repository delta is present and limited to two test files. No production edits.

## Coverage and review surfaces

2 of 4 AC rows driven by named production-entry tests; the other two are structural checks. The portability row is bounded: macOS execution and hosted Linux execution are available, but the hosted system true/false symlink layout is not attested. No native uutils execution is claimed. Repository-wide test search finds no remaining hard-coded true/false paths, so that layout cannot affect the replaced fixture assignments.

The brief supplies no catalog surface table or wall-clock budget. This scoped sweep uses the four AC surfaces plus recovery and hygiene. Its unrelated design-content/Current-state/D1-D13 checklist references documents that are not attached to this test-only leaf; those are not fabricated requirements.

| Surface | Result | Attack and limit |
| --- | --- | --- |
| Broker symlink refusal | held | Replayed both field-specific Lstat-to-Stat narrowing overlays through TestGitToolRejectsSymlinkedBrokers and production ValidateGitTool; each fails its matching assertion with nil instead of CodeIdentityInvalid. Positive regular-file control passes. |
| Fixture portability | held | Repository-wide coreutils-path search is empty; changed helpers drive AcquireNetwork/AdmitLocal via TestNetworkAndLocalSHA1SHA256RawObjectParity and AcquireNetworkResolved via TestResolvedTransportFallsBackOnAvailabilityThenVerifies. Full buildrepo suite rerun. Native uutils layout remains unmeasured. |
| Production scope | held | Exact candidate diff permits only admission_test.go and transport_test.go; worktree bytes match candidate, production admission.go matches base. |
| Recovery skip classification | held | Executed production platform-case-gate with candidate synthetic Windows stream and actual skip-class table: exit 0. Original stream is independently replayed as negative control below. Synthetic classification is not native execution. |
| Public hygiene | held | Candidate additions and authored producer result inspected: no personal paths, identifiers or host names introduced into code. Existing system-generated logs are not new authored product text. |

## Spot checks

Eight producer statements checked directly: helper uses t.TempDir (admission_test.go:730); mode is 0755 (731); script exits 1 (731); HTTP fixture uses helper (813); transport AskPass uses helper (683); transport SSHWrapper uses helper (684); regular positive control precedes symlink cases (743–747); each field asserts identity refusal and field-specific message (760–762). Additionally, production validation uses Lstat (admission.go:202), and real callers exist in admission.go:317 and local.go:45. Skip classification matches skip-classes.tsv:57. The two current test hashes and production hash match the recovery report.

## Findings

```json
{"findings": [], "notes": [
  {"severity":"P3","location":"internal/buildrepo/admission_test.go:737","text":"Add a concise claim comment beside the new test, as requested by the repository test-comment convention; behavior and limits are currently documented in the outcome instead."},
  {"severity":"note","text":"Native uutils system-layout execution is unknown; do not describe the hosted Linux pass as proof of that particular layout."},
  {"severity":"note","text":"No supplied catalog table, rc7 design documents, or review budget. Review is bounded to the actual test-only brief and AC."}
]}
```

Free hunt: checked nil-error assertion safety (short-circuit prevents err.Error on nil), field isolation, shell stub failure semantics, and unchanged production callers. No blocking finding reproduced. This does not prove absence of defects.

## Validation provenance

Reviewer reruns and expected-red logs are attached as TASK-261015-1yj9x7_review-validation-rev2.log. Reused producer evidence: TASK-261015-1yj9x7_recovery-validation.log (lint 0 issues, build exit 0, hashes match exact current source). Hosted snapshot d246b20ebf84906292f8044905ca86c6b07dd273 has exactly the candidate tree; TASK-261015-1yj9x7_change-request_rev2-validation.log ends in exit 0 and reports successful Linux/macOS/Windows jobs and lint. No missing suffix is treated as success. Reviewer did not replay the whole hosted matrix.

Fresh fetch shows upstream differs from base only in three unrelated board metadata paths; neither changed test nor validation input overlaps. No branch, index, commit or tracked file was modified by review. Acceptance is not landing.
