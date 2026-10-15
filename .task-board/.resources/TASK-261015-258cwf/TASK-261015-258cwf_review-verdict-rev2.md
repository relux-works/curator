# TASK-261015-258cwf review verdict, revision 2

Verdict: accepted. Candidate tree `a5c49ad3e15a43364c62dac70a642ce3fce77fc5`, base `3f5de6dff9aab4c037ba069323604b166bf0414e`. Repository delta is present and limited to the ten declared paths. Candidate blobs were compared byte-for-byte with the worktree, including untracked new files. No product files changed during review.

```json
{"verdict":"accepted","findings":[],"notes":["No producer surface table or review/free-hunt budget was supplied; the sweep below covers the eight behavioral AC rows and additional read faults.","The generic Review B references Required design content, Current state and D1-D13, but those sections and rc7 decision resources are not attached to this leaf. Reviewed the concrete four-item LOCK implementation brief; no claim of independently validating unavailable decisions.","Platform-wide tests and lint are reused from the exact-tree hosted gate; reviewer personally reran focused entry tests, five narrowing mutants, fragment parity, and additional read faults."]}
```

## AC coverage and sweep

**8 of 8 behavioral AC rows driven**, using the candidate tests named below. Entry is `run -> cmdEnv -> cmdEnvConfigShow`, then InstalledPermissions/Current/PermissionsForProfile; configuration errors are driven through the real file-backed config loader. There is no source-text gate. Expected-red mutants run Go behavioral tests with `-count=1`, not a static token checker. Findings and free-hunt findings are empty; held means only these attacks did not reproduce a defect.

| Row | Result | Driving test and attack |
| --- | --- | --- |
| Global lock | held | TestEnvConfigShowPermissionsAtProductionEntry/global; absent system-map profile cannot escape lock; global-lock narrowing killed |
| Profile override and additive JSON | held | TestEnvConfigShowPermissionsAtProductionEntry/profile; yolo remains profile/unlocked and raw map retained |
| Default and installed inventory | held | same /profile test; native/default and uninstalled exclusion |
| Current profile | held | same production-entry test, both lock cases; current-read narrowing killed |
| Config read failure | held | TestEnvConfigShowPermissionsRejectsUnreadableConfig, machine/system; system-read narrowing killed |
| Inventory type/current read | held | TestEnvConfigShowPermissionsRejectsUnreadableInventoryAndCurrent; inventory-type narrowing killed |
| Inventory ancestor metadata | held | TestEnvConfigShowPermissionsRejectsUnreadableInventoryAncestor; inventory-metadata narrowing killed |
| Empty inventory/missing current | held | TestEnvConfigShowPermissionsWithoutProfiles; empty map, no profile-state write, current unavailable |

The revision-specific metadata/type repair was attacked first with its existing regression and the metadata/type narrowing mutants before the remaining mutant sweep. Per-attack results were appended locally as each completed. No rows were left not-attacked. Free hunt: TestReviewLockReadFaults drives cyclic inventory and current symlinks and malformed machine JSON through production entry. All refuse with empty stdout; no finding. Scratch test is included in the evidence resource and used via Go overlay without changing candidate files.

## Eight source spot-checks

1. Full output is additive: cmd/curator/envconfig.go:41-50 retains EffectiveJSON and adds permissions_posture.
2. Existing and interactive output share JSON: cmd/curator/envconfig.go:166 printJSON; no separate human renderer introduced.
3. Current-profile form reads the pointer: cmd/curator/envconfig.go:53-64.
4. Missing current fails, not default/unlocked: cmd/curator/envconfig.go:59-61.
5. Global locks force native/global/true: internal/envprofile/permissions.go:14-16.
6. Profile/default modes remain fragment semantics: internal/envprofile/permissions.go:18-21; managed.go:2429 calls the same function.
7. Inventory metadata/type failures propagate: internal/envprofile/permissions.go:29-35; installedProfiles in migrate.go:342 propagates listing failures.
8. Both machine lock and effective permissions reach the resolver: cmd/curator/env.go:29-30; envfragment/envfragment.go:29-31 defines lowercase JSON keys.

Docs in docs/cli.md and docs/environment-config.md match these paths. CHANGELOG has the requested Unreleased Added line. README's lint entry is directly related validation documentation. Changed public text was inspected for prohibited identity/location data; none introduced. Brief items 1-4 addressed. Existing JSON-only behavior is the explicitly permitted human/JSON interpretation. Test fixtures cover deterministic filesystem failures, not every OS error or concurrent filesystem mutation.

## Validation and provenance

Hosted gate https://github.com/relux-works/curator/actions/runs/38686372579 independently queried as success at commit `bea713f832366a77edc0c6391703b167f9541952`; its Git tree equals the exact candidate tree. Attached revision-2 validation log ends with exit 0. Lint, macOS/Linux/Windows tests, race, interop, driver matrix and gate self-tests succeeded; optional candidate-suite and additional optional-host job were skipped, not counted as passes. No claim of a fresh full local suite.

Fresh remote main `207ef978667b086da50b6c53bcf53b549c14e945` matched advertised/fetched authority. Changes since CR base are board-only and do not overlap candidate paths. Landing is not part of this reviewer verdict.

## Logbook

The previous Windows regular-file inventory failure is addressed by an explicit metadata/type check. The existing regression passes on the hosted Windows job; targeted metadata/type mutants fail locally as intended. No new blocking mechanism found. The acceptance authorizes integration routing only, not a claim that landing occurred.

Raw reviewer logs and scratch reproduction: TASK-261015-258cwf_review-evidence-rev2.md.
