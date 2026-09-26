# TASK-260922-1t551d — revision 1 review verdict

Verdict: **changes_requested**, route to **to-dev**. No acceptance recorded.

## Candidate identity

Reviewed CR-TASK-260922-1t551d-1 revision 1, base 09b25ef6629b41455d91dcb252ab4e4034e12750, tree 2b179a7af518b9b73c5b0566d828676146977eca. All seven changed workspace files were byte-compared with the candidate objects. Adversarial probes and mutations ran only in `.temp/review-1t551d`, extracted from that tree; candidate production files were restored and compared with candidate objects after mutation. No producer code was edited.

GitHub gate https://github.com/relux-works/curator/actions/runs/35678477038 reports success for commit 9110bfa729688c6d3541ef39406a6ffc77ea5da3. `git rev-parse <commit>^{tree}` returns the exact candidate tree above. Full landing gate evidence is reused, not replayed. Normative checkout HEAD is 05053cd70bb64d68b3aa12e1af286203575d974c.

## Required rework

### R1 — dangling expected-target credential links are still silently current (high)

Location: `internal/envprofile/managed.go:1552-1559`; contrary regression assertion in `internal/envprofile/credential_link_test.go:354-366`.

`checkPassthrough` uses Lstat and Readlink and compares the stored path string, but never establishes that the native target exists. With a recorded Pi link pointing to the expected `.../pi/agent/auth.json` and no native file, `StatusOf` returns `Current:true`, `Findings:[]`; `Resolve` returns success with no stale reasons. This violates binding R2 and environments §7.4: dangling OR mis-targeted credential links must be reported detached with environment_credential_conflict-class wording. The operator wrong-target row only proves target-string mismatch detection, not dangling detection. The producer's own Pi provisioning test requires the contrary behavior, and its results incorrectly defer remaining dangling coverage to F-C3.

Reproduction: attached `TestReviewerDanglingExpectedTarget`, run through production `StatusOf` and `Resolve` on temporary stores. Both expectations fail on the unmodified candidate. Correct liveness checking without reading credential bytes, update the contrary test, and add a narrowing mutant distinguishing expected-target dangling from wrong-target links. Preserve the distinction between target absence and an inspection failure.

### R2 — valid TOML literal selectors bypass the Codex refusal gate (high)

Location: `internal/envprofile/managed.go:552-564`.

The regexp recognizes only an unquoted key followed by a double-quoted value. A valid native TOML value such as `cli_auth_credentials_store = 'keyring'` is treated as an absent key and defaults to file. The production isolated Resolve path succeeds for single-quoted `keyring`, `auto`, and `ephemeral`. Expected: environment_isolated_unsupported for the first two, environment_credential_unsupported for the last. This admits isolation under the native operator-global store, the exact R3 hazard.

Reproduction: attached `TestReviewerLiteralCodexSelector`, three subtests; 3/3 unexpectedly return nil errors. Parse the native TOML accurately, distinguish true absence from invalid/unreadable input, and cover supported TOML spelling forms at the production entry. Add narrowing mutants for the resulting refusal classes.

## Independent validation and gate attacks

All commands ran with real exit codes via zsh/Python subprocess on darwin. Eight new candidate test functions passed with `go test ./internal/envprofile -run 'Test(SharedToIsolatedRemovesStaleLink|CredentialLinkRegularFileRefuses|StaleCredentialLinkRefusals|DanglingPiLinkReportedDetached|CodexIsolatedAdmission|CodexUnknownStoreSharedRefuses|PiProvisionTargetsAgentRoot|CredentialLinkDirectory)$' -count=1` (8.712s, exit 0).

| Independent mutant | Production site | Result |
|---|---|---|
| Skip recorded-link removal | removeStaleCredentialLinks | killed: stale link survives shared to isolated |
| Restore unconditional Remove before ensuring wanted link | finalizeMarker | killed: both empty and nonempty regular-file rows admit instead of refusing |
| Compare only target basenames | checkPassthrough | killed: operator Pi status becomes current |

Mutants killed: 3/3 attempted, each an assertion failure rather than compilation failure. This does not certify the producer's other 11 mutants, which were not independently replayed. Additional reviewer negative shapes: 4/4 scenarios expose failures (one dangling target, three literal selector values). See attached raw evidence and reproduction test source. `git diff --check` passed.

Independent broader rerun: `go test ./internal/envprofile ./internal/envregistry -count=1` exited 0; envprofile 340.255s, envregistry 1.885s. This ran on the unchanged candidate workspace, whose seven changed files were byte-verified against the candidate tree. The green suite does not catch either added reviewer probe.

## Architecture and scope bounds

Recorded stale-link cleanup and refusal paths address the two original hazards; marker schema files are unchanged and the record construction still uses path+strategy only. The production diff adds no credential-content reads or copies; the Codex reader reads native config.toml. Pi credential target now reaches agent/auth.json. CHANGELOG and troubleshooting exist, but their broad dangling claim is not yet implemented.

These rows drive the Go production APIs used by CLI `cmd/curator/env.go:59` and `:95`; a CLI subprocess was not independently driven here. Producer CLI smoke evidence and the exact-tree landing gate were inspected/reused. Other operating systems and Windows symlink privileges are unverified by this review. Explicit credential migration remains F-C2, extended marker records remain F-S1. The two findings are ordinary implementation/coverage rework, not human-only decisions or external blockers.

Run goal query: no active goal (not goal-bound). No operator directives were present. Findings are also recorded in TASK-260922-1t551d_review-logbook-rev1.md; control-root LOGBOOK.md was not edited.
