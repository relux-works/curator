# Revision 3 independent local evidence

Candidate tree: da51dd68abb173fc894d8369e28662d6a17886d7.
Signed detached hosted snapshot: 7a2cc26eb4c2553aca352aec0ff5b9240538ca44; signature G.
Toolchain: Go 1.25.5, darwin/arm64.

Commands (all exit 0):
- go test ./internal/envprofile ./internal/envfragment -run 'TestMuse|TestMaterializeCaptureAndStatusNeverProbeVendor|TestMaterializeCaptureCredentialPolicyRelease' -count=1 -timeout=5m
- CURATOR_CONFORMANCE_ROOT=<worktree>/.temp/TASK-261016-wh6v9a/ci-spec/conformance/v1 go test ./cmd/curator -run 'TestMuse(CredentialPolicyNarrowingMutants|CLI|Fragment|RootContext)' -count=1 -v -timeout=7m
- go test -overlay <review-overlay.json> ./internal/envprofile -run '^TestReviewMuseCaptureRefusalNoProbe$' -count=1 -v -timeout=3m
- git merge-tree --write-tree a17e58156ac30a1629e3ff36efa259bab1555a15 be87653a94f23f39acbaa3e965a8f424cf769b60
- git diff --check

Conformance pin reused: 43bf0a2506d5c354a73bbc3ea4623d4653db10c7.
Normative scope inherited from accepted revision 1 and reviewed revision 2; no product/spec change in revision 3. Fresh Git fetch confirms current main equals base.
Hosted CI is pending and is NOT claimed passed in this local artifact.
No product code, real index, or Story branch was modified by review.


## comparison.json

{
  "candidate_tree": "da51dd68abb173fc894d8369e28662d6a17886d7",
  "previous_tree": "9bad791482638be26d3ab306653fd279791f449b",
  "clean_merge_tree": "4f88746642b992950ba314473747e321218eba00",
  "changed_candidate_paths_since_revision2": [
    "LOGBOOK.md"
  ],
  "owned_logbook_blocks_preserved": true
}

## sweep.md

# Revision 3 review sweep

Candidate: da51dd68abb173fc894d8369e28662d6a17886d7.
Revision-3 scope is convergence plus logbook relocation and fresh required CI.
Coverage checked before source inspection: 7 of 8 narrowed producer AC rows driven; row 8 has named boundary/schema tests and explicit external-consumer bound. No new AC row is omitted. Revision-2 interaction: 1 of 1 driven.
Clean merge independently recomputed with git merge-tree: 4f88746642b992950ba314473747e321218eba00. Only LOGBOOK.md differs. All 87 other candidate paths are byte-identical to revision 2. Both log blocks retain their content.

| Surface | Result | Fresh attack evidence / limits |
|---|---|---|
| observation boundary / cache | held | focused.log: TestMaterializeCaptureAndStatusNeverProbeVendor and CredentialPolicyRelease tests; real vendor sentinel, native/file x shared/isolated. CLI plant restores only Muse release observation and is killed. |
| destination conflicts | held | focused.log: TestMusePolicySwitchPreservesCredentials and TestMusePolicyExplicitMigration; refusal versus explicit migration controls through Resolve/ApplyMigration. |
| read boundary | held | focused.log: TestMuseAuthParentBoundaryAtResolve and TestMuseAuthRepairAdditionalStates; bad links/types against Resolve. |
| error surface | held | focused.log: TestMusePublishedLinkStates and TestMuseMissingNativeFilePolicy; failed/missing shared source refuses without placeholder. |
| selector paths | held | cli.log: TestMuseCLIStorePolicy; native/file positive controls and keychain refusal via real CLI; admission narrowing plant killed. |
| text grammar | held | cli.log: published v3/v4 corpus (pinned CI schemas) and boundary mutants admitting HOME, absent literal, false root are killed; external consumer ingestion remains unclaimed. |
| identity | held | cli.log: TestMuseRootContextVectors/Drift/Status and seven root narrowing plants, including header-only context; real Resolve/StatusOf entry points. |
| ordering | held | cli.log: TestMuseCLIStorePolicy; renderer plant removes only backend while preserving source tokens, named behavioral test fails. |

## Free hunt

free-hunt.log: TestReviewMuseCaptureRefusalNoProbe drives CaptureMaterialize and StatusOf through valid shared file, removed file, then restored file. Refusal is environment_home_stale, callback never runs, state stays unchanged, vendor trap never fires. Exit 0. No finding reproduced.

Fresh local controls: 16 Muse/root narrowing plants killed by named behavioral failures (9 policy/fragment and 7 root). Twelve standalone SNAP-LOCK overlay controls are inherited from revision-2 review, not rerun or represented as fresh evidence. Product/test definitions are unchanged; changed upstream scriptworker behavior is exercised by the fresh whole gate.

## focused.log

ok  	github.com/relux-works/curator/internal/envprofile	26.113s
ok  	github.com/relux-works/curator/internal/envfragment	0.270s
exit=0

## cli.log

=== RUN   TestMuseCredentialPolicyNarrowingMutants
=== RUN   TestMuseCredentialPolicyNarrowingMutants/fragment-reobserves-muse-release
    muse_credential_mutants_test.go:77: fragment-reobserves-muse-release: expected-red go test exit 1; killed by TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared
=== RUN   TestMuseCredentialPolicyNarrowingMutants/muse-keychain-admitted
    muse_credential_mutants_test.go:77: muse-keychain-admitted: expected-red go test exit 1; killed by TestMuseCredentialPolicyMatrix/keychain/shared
=== RUN   TestMuseCredentialPolicyNarrowingMutants/missing-shared-source-admitted
    muse_credential_mutants_test.go:77: missing-shared-source-admitted: expected-red go test exit 1; killed by TestMuseMissingNativeFilePolicy/shared
=== RUN   TestMuseCredentialPolicyNarrowingMutants/isolated-muse-native-link
    muse_credential_mutants_test.go:77: isolated-muse-native-link: expected-red go test exit 1; killed by TestMuseCredentialPolicyMatrix/file/isolated
=== RUN   TestMuseCredentialPolicyNarrowingMutants/renderer-drops-only-backend
    muse_credential_mutants_test.go:77: renderer-drops-only-backend: expected-red go test exit 1; killed by TestMuseCLIStorePolicy
=== RUN   TestMuseCredentialPolicyNarrowingMutants/header-only-root-fact-dropped
    muse_credential_mutants_test.go:77: header-only-root-fact-dropped: expected-red go test exit 1; killed by TestMuseRootContextVectors/zero-applicable
=== RUN   TestMuseCredentialPolicyNarrowingMutants/boundary-admits-HOME
    muse_credential_mutants_test.go:77: boundary-admits-HOME: expected-red go test exit 1; killed by TestMuseFragmentBoundary/HOME
=== RUN   TestMuseCredentialPolicyNarrowingMutants/boundary-admits-missing-literal
    muse_credential_mutants_test.go:77: boundary-admits-missing-literal: expected-red go test exit 1; killed by TestMuseFragmentBoundary/missing-literal
=== RUN   TestMuseCredentialPolicyNarrowingMutants/boundary-admits-false-root
    muse_credential_mutants_test.go:77: boundary-admits-false-root: expected-red go test exit 1; killed by TestMuseFragmentBoundary/false-root
--- PASS: TestMuseCredentialPolicyNarrowingMutants (18.56s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/fragment-reobserves-muse-release (3.32s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/muse-keychain-admitted (2.34s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/missing-shared-source-admitted (2.33s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/isolated-muse-native-link (2.50s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/renderer-drops-only-backend (3.36s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/header-only-root-fact-dropped (3.20s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/boundary-admits-HOME (0.50s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/boundary-admits-missing-literal (0.52s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/boundary-admits-false-root (0.49s)
=== RUN   TestMuseRootContextVectors
=== RUN   TestMuseRootContextVectors/monolithic
=== RUN   TestMuseRootContextVectors/zero-applicable
=== RUN   TestMuseRootContextVectors/no-context
=== RUN   TestMuseRootContextVectors/referenced-refused
--- PASS: TestMuseRootContextVectors (3.95s)
    --- PASS: TestMuseRootContextVectors/monolithic (1.14s)
    --- PASS: TestMuseRootContextVectors/zero-applicable (0.95s)
    --- PASS: TestMuseRootContextVectors/no-context (1.04s)
    --- PASS: TestMuseRootContextVectors/referenced-refused (0.81s)
=== RUN   TestMuseRootContextDrift
=== RUN   TestMuseRootContextDrift/changed
=== RUN   TestMuseRootContextDrift/removed
--- PASS: TestMuseRootContextDrift (2.23s)
    --- PASS: TestMuseRootContextDrift/changed (1.10s)
    --- PASS: TestMuseRootContextDrift/removed (1.13s)
=== RUN   TestMuseRootContextStatus
--- PASS: TestMuseRootContextStatus (0.57s)
=== RUN   TestMuseRootContextNarrowingMutants
=== RUN   TestMuseRootContextNarrowingMutants/configured-form-ignored
    muse_root_context_test.go:291: configured-form-ignored: expected-red go test exit 1; killed by TestMuseRootContextVectors/referenced-refused
=== RUN   TestMuseRootContextNarrowingMutants/referenced-fallback
    muse_root_context_test.go:291: referenced-fallback: expected-red go test exit 1; killed by TestMuseRootContextVectors/referenced-refused
=== RUN   TestMuseRootContextNarrowingMutants/no-context-admitted
    muse_root_context_test.go:291: no-context-admitted: expected-red go test exit 1; killed by TestMuseRootContextVectors/no-context
=== RUN   TestMuseRootContextNarrowingMutants/wrong-target
    muse_root_context_test.go:291: wrong-target: expected-red go test exit 1; killed by TestMuseRootContextVectors/monolithic
=== RUN   TestMuseRootContextNarrowingMutants/wrong-bytes
    muse_root_context_test.go:291: wrong-bytes: expected-red go test exit 1; killed by TestMuseRootContextVectors/zero-applicable
=== RUN   TestMuseRootContextNarrowingMutants/changed-root-admitted
    muse_root_context_test.go:291: changed-root-admitted: expected-red go test exit 1; killed by TestMuseRootContextDrift/changed
=== RUN   TestMuseRootContextNarrowingMutants/removed-root-admitted
    muse_root_context_test.go:291: removed-root-admitted: expected-red go test exit 1; killed by TestMuseRootContextDrift/removed
--- PASS: TestMuseRootContextNarrowingMutants (22.85s)
    --- PASS: TestMuseRootContextNarrowingMutants/configured-form-ignored (2.93s)
    --- PASS: TestMuseRootContextNarrowingMutants/referenced-fallback (3.46s)
    --- PASS: TestMuseRootContextNarrowingMutants/no-context-admitted (3.17s)
    --- PASS: TestMuseRootContextNarrowingMutants/wrong-target (3.32s)
    --- PASS: TestMuseRootContextNarrowingMutants/wrong-bytes (3.34s)
    --- PASS: TestMuseRootContextNarrowingMutants/changed-root-admitted (3.36s)
    --- PASS: TestMuseRootContextNarrowingMutants/removed-root-admitted (3.28s)
=== RUN   TestMuseCLIFragmentAndFakeLaunch
--- PASS: TestMuseCLIFragmentAndFakeLaunch (1.04s)
=== RUN   TestMuseFragmentV3PublishedCases
=== RUN   TestMuseFragmentV3PublishedCases/valid.json
    muse_test.go:275: production Resolve + CLI emission; schema/reader oracle also checked
=== RUN   TestMuseFragmentV3PublishedCases/invalid.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/valid-permissions-yolo-unlocked.json
    muse_test.go:275: production Resolve + CLI emission; schema/reader oracle also checked
=== RUN   TestMuseFragmentV3PublishedCases/valid-permissions-native-locked.json
    muse_test.go:275: production Resolve + CLI emission; schema/reader oracle also checked
=== RUN   TestMuseFragmentV3PublishedCases/valid-permissions-native-silent.json
    muse_test.go:275: production Resolve + CLI emission; schema/reader oracle also checked
=== RUN   TestMuseFragmentV3PublishedCases/invalid-permissions-absent.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-permissions-unknown-mode.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-permissions-yolo-locked.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-permissions-yolo-silent.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-permissions-locked-profile-source.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-permissions-global-unlocked.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-permissions-unknown-source.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-permissions-unknown-field.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-permissions-missing-mode.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/valid-claude_code.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-extra-parent-claude_code.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/valid-codex_cli.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-extra-parent-codex_cli.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/valid-opencode.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-extra-parent-opencode.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/valid-pi.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-extra-parent-pi.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-missing-XDG_CONFIG_HOME.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-wrong-parent-XDG_CONFIG_HOME.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-missing-XDG_DATA_HOME.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-wrong-parent-XDG_DATA_HOME.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-missing-XDG_STATE_HOME.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-wrong-parent-XDG_STATE_HOME.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-missing-XDG_CACHE_HOME.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-wrong-parent-XDG_CACHE_HOME.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-HOME-added.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-HOME-substitution.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-unverified-system-prompt.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-unverified-mcp.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-v2-token.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-unknown-environment.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== NAME  TestMuseFragmentV3PublishedCases
    coverage.go:236: published cases launch-env-fragment-v3/schema-cases: 36 driven, 0 known-gap, 0 bound, 0 skipped, 36 total
--- PASS: TestMuseFragmentV3PublishedCases (1.11s)
    --- PASS: TestMuseFragmentV3PublishedCases/valid.json (0.41s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/valid-permissions-yolo-unlocked.json (0.04s)
    --- PASS: TestMuseFragmentV3PublishedCases/valid-permissions-native-locked.json (0.04s)
    --- PASS: TestMuseFragmentV3PublishedCases/valid-permissions-native-silent.json (0.04s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-permissions-absent.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-permissions-unknown-mode.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-permissions-yolo-locked.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-permissions-yolo-silent.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-permissions-locked-profile-source.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-permissions-global-unlocked.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-permissions-unknown-source.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-permissions-unknown-field.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-permissions-missing-mode.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/valid-claude_code.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-extra-parent-claude_code.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/valid-codex_cli.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-extra-parent-codex_cli.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/valid-opencode.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-extra-parent-opencode.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/valid-pi.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-extra-parent-pi.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-missing-XDG_CONFIG_HOME.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-wrong-parent-XDG_CONFIG_HOME.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-missing-XDG_DATA_HOME.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-wrong-parent-XDG_DATA_HOME.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-missing-XDG_STATE_HOME.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-wrong-parent-XDG_STATE_HOME.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-missing-XDG_CACHE_HOME.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-wrong-parent-XDG_CACHE_HOME.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-HOME-added.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-HOME-substitution.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-unverified-system-prompt.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-unverified-mcp.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-v2-token.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-unknown-environment.json (0.00s)
=== RUN   TestMuseCLIAuthParentBoundary
=== RUN   TestMuseCLIAuthParentBoundary/resolve
    muse_test.go:467: production CLI refusal exit=1; no fragment emitted
=== RUN   TestMuseCLIAuthParentBoundary/resolve-repair
    muse_test.go:467: production CLI refusal exit=1; no fragment emitted
--- PASS: TestMuseCLIAuthParentBoundary (0.96s)
    --- PASS: TestMuseCLIAuthParentBoundary/resolve (0.01s)
    --- PASS: TestMuseCLIAuthParentBoundary/resolve-repair (0.02s)
=== RUN   TestMuseFragmentV4PublishedCases
=== RUN   TestMuseFragmentV4PublishedCases/valid-muse-root.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-muse-no-root.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-missing-XDG_CONFIG_HOME.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-missing-XDG_DATA_HOME.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-missing-XDG_STATE_HOME.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-missing-XDG_CACHE_HOME.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-missing-TBH_CREDENTIAL_BACKEND.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-root-null.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-root-empty.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-root-boolean.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-root-false.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-root-string.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-root-extra.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-backend-keychain.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-backend-absolute.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-backend-empty.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-store-native.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-store-keychain.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-store-null.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-store-empty.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-store-extra.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-store-override.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-store-missing.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-HOME.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-foreign-parent.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-wrong-suffix.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-relative-path.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-unknown-top.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-unknown-revision.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-system-prompt.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-mcp.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-claude_code.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-claude_code-root.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-claude_code-backend.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-claude_code-native.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-claude_code-file.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-codex_cli.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-codex_cli-root.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-codex_cli-backend.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-codex_cli-native.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-codex-file.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-codex-file-disagreement.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-codex-file-missing-override.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-codex-keychain.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-codex-keychain-disagreement.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-codex-keychain-missing-override.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-codex-native-override.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-codex-unknown-override.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-opencode.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-opencode-root.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-opencode-backend.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-opencode-native.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-opencode-file.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-pi.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-pi-root.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-pi-backend.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-pi-native.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-pi-file.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-v4-on-v1.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-additions-on-v1.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-v4-on-v2.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-additions-on-v2.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-v4-on-v3.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-additions-on-v3.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-v3-root_context.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-v3-credential_store.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-v3-backend.json
--- PASS: TestMuseFragmentV4PublishedCases (0.04s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-muse-root.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-muse-no-root.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-missing-XDG_CONFIG_HOME.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-missing-XDG_DATA_HOME.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-missing-XDG_STATE_HOME.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-missing-XDG_CACHE_HOME.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-missing-TBH_CREDENTIAL_BACKEND.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-root-null.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-root-empty.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-root-boolean.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-root-false.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-root-string.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-root-extra.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-backend-keychain.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-backend-absolute.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-backend-empty.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-store-native.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-store-keychain.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-store-null.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-store-empty.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-store-extra.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-store-override.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-store-missing.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-HOME.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-foreign-parent.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-wrong-suffix.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-relative-path.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-unknown-top.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-unknown-revision.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-system-prompt.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-mcp.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-claude_code.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-claude_code-root.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-claude_code-backend.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-claude_code-native.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-claude_code-file.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-codex_cli.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-codex_cli-root.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-codex_cli-backend.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-codex_cli-native.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-codex-file.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-codex-file-disagreement.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-codex-file-missing-override.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-codex-keychain.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-codex-keychain-disagreement.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-codex-keychain-missing-override.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-codex-native-override.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-codex-unknown-override.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-opencode.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-opencode-root.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-opencode-backend.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-opencode-native.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-opencode-file.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-pi.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-pi-root.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-pi-backend.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-pi-native.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-pi-file.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-v4-on-v1.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-additions-on-v1.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-v4-on-v2.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-additions-on-v2.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-v4-on-v3.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-additions-on-v3.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-v3-root_context.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-v3-credential_store.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-v3-backend.json (0.00s)
=== RUN   TestMuseCLIStorePolicy
--- PASS: TestMuseCLIStorePolicy (1.13s)
PASS
ok  	github.com/relux-works/curator/cmd/curator	52.813s
exit=0

## free-hunt.log

=== RUN   TestReviewMuseCaptureRefusalNoProbe
--- PASS: TestReviewMuseCaptureRefusalNoProbe (1.08s)
PASS
ok  	github.com/relux-works/curator/internal/envprofile	1.407s
exit=0
