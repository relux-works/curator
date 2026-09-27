# TASK-260916-1zgucp — review verdict rev4 (CR-TASK-260916-1zgucp-4, tree 3a680a1e): CHANGES REQUESTED

Scope per delta-review note: the 7 intersecting paths. Method: synthetic commit (bd3c0f43 + tree 78bcb664) three-way merged with d41da0fb via `git merge-tree --write-tree` (tree 92fa4bbe, conflicts on 6 paths), then diffed against 3a680a1e with conflict markers filtered, plus per-revision grep of the introduced constructs.

## Clean
- internal/envprofile/status.go, internal/envprofile/surfacing_test.go, internal/config/config.go: both sides present, no drop/duplication (trunk's `status.Profiles, profileDiagnostics, err = profileStates(...)` kept, status.go:322).
- internal/config/environments.go knob lists / EnvLockKey / render: permissions + source_signers + require_source_signers merged correctly, no duplicate keys.
- environments_test.go: lockable set comment and stable-first-blocker test reconciled (now uses synthetic a_unknown/z_unknown, sound since both real fields are supported).
- conformance-gaps.tsv: 44 E1-owned rows removed + 11 STORY-260922-1cenbr rows removed (manager-config-v2 valid*, schema2-* vectors). The 1cenbr removals are justified: those cases pass on the candidate (below), and the gap ledger forbids rows for passing cases. No trunk row re-added.
- Tests on an archive of 3a680a1e (zsh, pipefail): `go test ./internal/config` → ok, EXIT=0; `go test ./internal/envprofile -run 'Status|Surfacing|Signer|Delta|Guarded'` → ok (144s), EXIT=0.

## Findings (content in neither rev3 nor trunk — not a re-apply)
F1. internal/config/environments.go:489-491 (in parseOpenSSHPublicKey) adds a `base64.RawStdEncoding` fallback. Count of `RawStdEncoding`: bd3c0f43=0, rev3 78bcb664=0, trunk d41da0fb=0, rev4=1. This WIDENS the source-signer key admission gate (unpadded key material now accepted) with no accepted spec citation and no negative test; it was not in the accepted rev3. Remove it, or justify against the spec (OpenSSH authorized_keys material is padded std base64) in a new revision with a test.
F2. internal/config/environments_conformance_test.go:414-440 rewrites trunk's `exactManagerEffectiveJSON` (trunk contract: "compares canonical JSON bytes for the full normalized object. In particular, extra actual keys are a mismatch too") to compare only the `environments` sub-map, and rewrites `TestManagerEffectiveJSONComparisonRejectsExtraKnobs` so an extra top-level key (`adapter_mode`) is no longer covered. `actualEnvironments` count: all of bd3c0f43/78bcb664/d41da0fb=0, rev4=2. This is a revert/weakening of a trunk gate. It is also unnecessary: restoring trunk's strict comparator in the candidate tree, `go test ./internal/config -run TestManagerConfigV2Vectors` → ok (EXIT 0). Restore trunk's function and test verbatim.

## Required for rev5
Drop F1 and F2 (restore trunk bytes for those hunks); keep everything else of rev4. Re-run `go test ./internal/config` and the envprofile focused run with real exit codes; hosted gate green.
