# L15-1 slice 1: revision 2 developer evidence

Scope: TASK-261016-2en3yu — implement-rc7-credential-store-policy. This rework answers the P2 `dotted-profile-selector-path` finding in TASK-261016-2en3yu_review-verdict-rev1.md. The deciding review recorded `repeat-of: none`; a named regression and narrowing mutant were still added in this leaf.

## Candidate and scope

Base and unchanged Story HEAD: `dccac3199aa5c543c589c129f5ef2bc437523a2a`. Signed detached gate snapshot: `f2a79188823008f44e7e0a45b59beb044b00f115`, tree `ddde39b56122ce8b78ecb49ab410914ab8693171`. The 16 changed files match that snapshot; the Story index is unchanged and no producer commit was made on the Story branch. `git rev-list --count HEAD..main` returned 0 before verification. Source identities are supplied separately using relative paths and SHA-256 digests.

Normative curator-spec main was freshly read through the repository API and remains `15d970111277b1f68b3f7a4ddd63bc339b2c0a3f`. Protocol login-readiness, environments and profile-codex-settings bytes match the preceding reviewed pin. The common identifier schema permits internal dots, including repeated internal dots, but excludes leading/trailing dots and reserved names. No new normative conflict was found within the explicit slice-1 scope decision.

Slice 1 is configuration grammar, CLI, typed transport into resolve and fail-closed unavailable selections. Effective backend application, successor fragments, login/readiness and vendor runtime qualification remain owned by TASK-261016-hzbo83 (L15-1b). Nothing here claims that work.

## Rework and regression witness

`config.SplitEnvKnob` now parses a full credential-store selector from its two boundaries: `credential_store` and the final environment operand. The middle bytes form one validated profile identifier. This preserves `team.prod`, `Team..prod`, and `codex.prod` without changing isolation or other knob parsing. Alias normalization subsequently touches only the environment operand. The config CLI still edits only the user machine file.

- `TestReviewDottedCredentialProfile` and `TestReviewDottedCredentialProfileExisting` are now in the candidate under their original names. Both pass, exercising the reviewer's previously failing set/show/unset paths.
- `TestCredentialStoreCLIDottedProfileRouting` covers 4 profiles × 4 environment spellings (Codex/Claude aliases and canonical IDs). Every case drives set/show/unset, checks canonical persistence and exact profile identity, tests invalid-value/unknown-environment/missing-unset refusals with unchanged file bytes, and verifies the neighboring selector survives.
- `TestCredentialStoreCLIInvalidProfileRouting` covers 5 invalid profile classes × 3 commands. Each refuses with usage exit 2, emits no success output and leaves persisted policy unchanged. Valid internal-dot cases are the neighboring positive controls.
- The initial regression command ran against the unfixed candidate and exited 1 in the dotted-profile cases while plain profiles passed. That expected-red result is not a passing gate.

The rework adds 9 parser lines to internal/config/environments.go, focused CLI tests and 2 documentation lines relative to revision 1. No previously committed test was removed or renamed. All rework paths are in the leaf's declared modules. The original candidate's README.md addition remains the only extra path, justified by the root requirement to document exact tool commands; this rework does not alter README.

## Measured acceptance coverage

**6 of 6 slice AC groups driven** through production entry points; **3 of 3 selector command forms driven**. These are bounded slice groups, not a claim of full original-leaf conformance.

| Slice AC group | Production call site | Named tests |
|---|---|---|
| Successor grammar and frozen envelopes | config.Load → parseEnvironmentsVersion → parseCredentialStore | TestCredentialStoreConfigGrammar; TestCredentialStoreConfigVersionBoundary |
| Machine ownership and independent isolation locks | config.Load → system merge validation | TestCredentialStoreConfigMachineOwnership; TestSystemV2LockableSubsetIsClassWide |
| CLI aliases and complete profile identity | cli.run → cmdEnvConfigSet/Show/Unset → SplitEnvKnob/WriteRaw | TestCredentialStoreCLIConfigRoundTrip; TestCredentialStoreCLIConfigRefusesMalformed; TestCredentialStoreCLIDottedProfileRouting; TestCredentialStoreCLIInvalidProfileRouting; both TestReviewDottedCredentialProfile tests |
| Atomic successor promotion, unrelated posture preserved | cli.run → cmdEnvConfigSet → WriteRaw | TestCredentialStoreCLIPromotion |
| Typed transport and unavailable selection refusal | machineFromConfig → Resolve → admitCredentialStore; checkIsolatedCredentialStore → admitCredentialStore | TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery; TestCredentialStorePolicyActivation |
| Native read failure distinguished from absence; state preserved | Resolve → admitCredentialStore → codexCredentialStoreWith | TestCredentialStorePolicyNativeInspection; CLI resolve/round-trip state assertions |

Config validation checks valid selectors beside malformed/null/unknown/alias-persisted values. System configuration cannot set or lock this selector; isolation locks still bind. Failed CLI edits leave bytes unchanged. Explicit unsupported file/keychain mappings refuse without provisioning or emitting a fragment, and native Codex still produces a real fragment. Permission/parse/unknown-native-selector failures remain distinct from an absent key. These are synthetic-home tests; the permission error uses the existing I/O seam and is not physical platform qualification.

## Coverage map

Brief gap retained: the original brief has adapter and paired-test tables but no catalog surface table. The supplemental three-row map from revision 1 is preserved and updated for the finding.

| Surface row | Attacking tests | Killed narrowing mutants | Out-of-contract inputs (AC clause) |
|---|---|---|---|
| text grammar | TestCredentialStoreConfigGrammar, TestCredentialStoreConfigVersionBoundary | old-v3 → TestCredentialStoreConfigVersionBoundary; unknown-auto, persist-alias → TestCredentialStoreConfigGrammar | Unknown/malformed/null entries and old carriers must refuse under leaf grammar; none silently excluded. |
| selector paths | TestCredentialStoreConfigMachineOwnership, TestCredentialStoreCLIConfigRoundTrip, TestCredentialStoreCLIDottedProfileRouting, TestCredentialStoreCLIInvalidProfileRouting, TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery, both reviewer regressions | system-store → TestCredentialStoreConfigMachineOwnership; profile-normalization → TestCredentialStoreCLIConfigRoundTrip; dotted-profile-only → TestCredentialStoreCLIDottedProfileRouting; reserved-profile-show → TestCredentialStoreCLIInvalidProfileRouting/CON.prod/show; drop-codex-file → TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery | Invalid profile identifiers refuse; valid dotted identifiers are in contract. |
| error surface | TestCredentialStorePolicyNativeInspection | permission-as-absent → TestCredentialStorePolicyNativeInspection/permission | No failed read is excluded as out of contract. |

## Narrowing controls

Each command is a standalone behavioral `go test -overlay … <package> -run <named-test> -count=1` process with output redirected directly to a file, without a pipeline. The portable definitions specify exact relative source paths/replacements and test expressions. Each process exited **1** in the named test; all nine controls are killed, with no surviving executed mutant and no setup/build failure counted as a kill.

| Mutant | What the weakened gate admits | Named failing test | Result / survivor bound |
|---|---|---|---|
| old-v3 | Only version 3 may carry the new selector | TestCredentialStoreConfigVersionBoundary | exit 1, killed |
| unknown-auto | Only the extra machine selector auto | TestCredentialStoreConfigGrammar/unknown_selector | exit 1, killed |
| persist-alias | Alias accepted as a persisted environment key | TestCredentialStoreConfigGrammar/alias_id | exit 1, killed |
| system-store | Only credential_store bypasses the system subset gate | TestCredentialStoreConfigMachineOwnership | exit 1, killed |
| profile-normalization | Profile operand normalized as an environment alias | TestCredentialStoreCLIConfigRoundTrip | exit 1, killed |
| drop-codex-file | Only explicit Codex file dropped from CLI-to-resolve transport | TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery | exit 1, killed on forbidden native fragment |
| permission-as-absent | Only native permission read failure becomes absent | TestCredentialStorePolicyNativeInspection/permission | exit 1, killed on wrong diagnostic |
| dotted-profile-only | Profile regrouping retained only for plain selectors; dotted routing breaks | TestCredentialStoreCLIDottedProfileRouting/team.prod/codex (and neighboring dotted cases) | exit 1, killed; plain-profile controls pass |
| reserved-profile-show | Only reserved profile CON.prod bypasses profile validation | TestCredentialStoreCLIInvalidProfileRouting/CON.prod/show | exit 1, killed on forbidden native-default success |

No source-text checker is introduced; the token-preserving static-checker obligation is inapplicable. The original hosted-override-drop plant is not run because slice 1 emits no hosted override. The CLI transport plant proves the slice's real composition boundary, not deferred hosted application.

## Out-of-slice rows and stated bounds

The controlling AC clause is the convener's explicit narrowing to “config grammar, CLI, typed transport into resolve, and fail-closed unsupported selections.” Under that clause the following original paired-table portions belong to L15-1b:

| Original paired row | Slice coverage and deferred portion |
|---|---|
| Old/system/package/overlay/malformed selector | Machine grammar, system ownership and malformed values covered; new package/settings carrier contribution controls deferred. No package vocabulary added. |
| Shared explicit file with stale auth | Explicit selections refuse here; qualified shared-file/stale-auth admission deferred. |
| Native keyring/auto isolated and explicit keychain per-home | Existing native isolation suites retained; new explicit application refuses; qualification and isolated-file enablement deferred. |
| Native read failure versus absence | Covered here, with the I/O seam limitation stated above. |
| Muse file literal/default and readiness | Effective-file mapping, four XDG paths plus literal, native compatibility and readiness deferred. Existing native behavior is not claimed as rc.7 conformance. |
| Selector/repair ownership migration | Config edits and refusal preserve state; successful backend migration matrix deferred. |
| Alias persistence/profile normalization | Covered here, now including the dotted-profile class. |
| Adopted node/helper/broker source and forced override | Generic explicit refusal covered; source-specific negative/positive matrix deferred. |

The original whole-leaf paired table remains **2 of 8 fully driven rows**. Full selector paths use the leaf's `<profile>.<env-id>` spelling; a dotted profile subtree-only shorthand is not defined by that grammar (whole-map JSON edits remain available).

## Validation ledger

All commands below were executed in this rework, against the frozen candidate unless the expected-red row says otherwise. Local platform: darwin/arm64; Go 1.25.5; golangci-lint 2.12.2. No prior full-gate green is reused after this source/test identity change.

| Command | Exit | Evidence |
|---|---:|---|
| go test ./cmd/curator -run '^(TestReviewDottedCredentialProfile\|TestCredentialStoreCLIDottedProfileRouting)' -count=1 (before parser fix) | 1 | regression-red-01.log; expected-red dotted routing witness |
| go test ./internal/config ./internal/envregistry ./internal/envprofile ./cmd/curator -run '^(TestCredentialStore\|TestReviewDottedCredentialProfile)' -count=1 | 0 | focused-04.log; includes retained reviewer names and new negative/positive controls |
| go test ./internal/config ./internal/envregistry -count=1 | 0 | config-registry-07.log |
| go build ./... | 0 | build-06.log |
| go vet ./internal/config ./internal/envregistry ./internal/envprofile ./cmd/curator | 0 | vet-08.log |
| GOLANGCI_LINT_CACHE=<task-temp>/lint-cache golangci-lint run | 0 | lint-10.log; 0 issues |
| git diff --check | 0 | diff-check-09.log |
| git diff --cached --exit-code | 0 | index-check.log; index unchanged |
| Signed copy of sh scripts/remote-gate.sh | 0 | remote-gate-05.log; https://github.com/relux-works/curator/actions/runs/38869761755 |

Gate invocation: `source <configured-agent-git-env>; sh .temp/review-l151/signed-remote-gate.sh`. The repository script creates an unsigned snapshot under an invented author, contrary to the supplied signing rules. This run reuses the reviewer's existing scratch copy: only snapshot creation changes to `git commit-tree -S` with configured human identity and `git verify-commit` before publication. Workflow, test selection, polling, retries and cleanup are unchanged. The Story branch and index are never changed. The accompanying signing-only diff documents the difference; `diff -u` exits 1 as expected for that comparison, not a failed product gate. No unrelated reusable-tool fix is included in this leaf.

The full hosted suite reached completed/success: 20 successful jobs and two conditional skips. The gate process returned exit 0; a separate terminal API readback confirmed the same exact signed head. All repository test/race, build/vet, lint and gate jobs selected by this workflow passed. No local full-suite rerun is claimed; the local commands above and the full hosted suite were run in this rework. The optional additional self-hosted job and workflow-dispatch candidate job were skipped by unchanged workflow conditions.

Slice 1 is ready for review; review acceptance and integration remain the next role’s work.

## Logbook

- Revision 1 failed on a valid dotted profile, a selector identity defect. This rework stays in the leaf and carries the review witness into repository tests under both original names.
- The fixed parser handles repeated internal dots as required by the identifier schema. Invalid profile paths refuse atomically; no compensating runtime or vendor behavior is added.
- Seven existing narrowing controls were rerun against current candidate bytes, plus the new dotted-only and reserved-profile-show controls. All nine failed in named assertions. This is fresh evidence, not inference from prior kills.
- Runtime qualification, effective store activation, hosted override delivery and L15-1b remain explicitly unclaimed.

## Terminal hosted evidence

```json
{
  "status": "completed",
  "conclusion": "success",
  "url": "https://github.com/relux-works/curator/actions/runs/38869761755",
  "headSha": "f2a79188823008f44e7e0a45b59beb044b00f115",
  "tree": "ddde39b56122ce8b78ecb49ab410914ab8693171",
  "gate_exit_code": 0,
  "signature_verified": true,
  "successful_jobs": 20,
  "conditional_skips": 2,
  "jobs": [
    {
      "name": "Lint",
      "status": "completed",
      "conclusion": "success"
    },
    {
      "name": "Race (macos-latest)",
      "status": "completed",
      "conclusion": "success"
    },
    {
      "name": "Test (ubuntu-latest)",
      "status": "completed",
      "conclusion": "success"
    },
    {
      "name": "Gate self-test (ubuntu-latest)",
      "status": "completed",
      "conclusion": "success"
    },
    {
      "name": "Race (ubuntu-latest)",
      "status": "completed",
      "conclusion": "success"
    },
    {
      "name": "Test (macos-latest)",
      "status": "completed",
      "conclusion": "success"
    },
    {
      "name": "Go 1.25.5 driver (macos-latest)",
      "status": "completed",
      "conclusion": "success"
    },
    {
      "name": "Gate self-test (macos-latest)",
      "status": "completed",
      "conclusion": "success"
    },
    {
      "name": "Test (windows-latest)",
      "status": "completed",
      "conclusion": "success"
    },
    {
      "name": "Gate self-test (windows-latest)",
      "status": "completed",
      "conclusion": "success"
    },
    {
      "name": "Go 1.27.1 driver (macos-latest)",
      "status": "completed",
      "conclusion": "success"
    },
    {
      "name": "Interop conformance gate",
      "status": "completed",
      "conclusion": "success"
    },
    {
      "name": "Go 1.27.1 driver (ubuntu-latest)",
      "status": "completed",
      "conclusion": "success"
    },
    {
      "name": "Go 1.26.0 driver (ubuntu-latest)",
      "status": "completed",
      "conclusion": "success"
    },
    {
      "name": "Go 1.26.0 driver (windows-latest)",
      "status": "completed",
      "conclusion": "success"
    },
    {
      "name": "Go 1.25.5 driver (windows-latest)",
      "status": "completed",
      "conclusion": "success"
    },
    {
      "name": "Naming gate",
      "status": "completed",
      "conclusion": "success"
    },
    {
      "name": "Go 1.25.5 driver (ubuntu-latest)",
      "status": "completed",
      "conclusion": "success"
    },
    {
      "name": "Go 1.26.0 driver (macos-latest)",
      "status": "completed",
      "conclusion": "success"
    },
    {
      "name": "Go 1.27.1 driver (windows-latest)",
      "status": "completed",
      "conclusion": "success"
    },
    {
      "name": "Candidate suite (${{ matrix.os }})",
      "status": "completed",
      "conclusion": "skipped"
    },
    {
      "name": "Optional self-hosted test job",
      "status": "completed",
      "conclusion": "skipped"
    }
  ]
}
```

## Gate snapshot-signing delta

Only commit construction differs from the repository script. The original identity literals are redacted here; neither identity nor signing changes a test command.

```diff
--- scripts/remote-gate.sh	2026-10-16 08:25:17
+++ .temp/review-l151/signed-remote-gate.sh	2026-10-16 10:59:06
@@ -43,9 +43,8 @@
   commit=$head
 else
   commit=$(printf 'remote gate snapshot of %s\n\nworking tree on top of %s\n' "$name" "$head" \
-    | GIT_COMMITTER_NAME="remote-gate" GIT_COMMITTER_EMAIL="<snapshot-address>" \
-      GIT_AUTHOR_NAME="remote-gate" GIT_AUTHOR_EMAIL="<snapshot-address>" \
-      git -c commit.gpgsign=false commit-tree "$tree" -p "$head")
+    | git commit-tree -S "$tree" -p "$head")
+  git verify-commit "$commit"
 fi
 
 ATTEMPTS="${REMOTE_GATE_ATTEMPTS:-3}"

```

## Delivery logbook

- The terminal full remote verdict is completed/success on the exact candidate tree. The source and test identity has remained frozen since dispatch.
- All three rework paths are inside the leaf’s declared modules. The candidate stays uncommitted with the Story HEAD and index unchanged.
- This evidence supersedes the pending hosted status recorded while local/control evidence was first attached. No runtime qualification or L15-1b implementation is claimed.
