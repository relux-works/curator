# M1 gap analysis

Baseline: curator `34a89308a69d4863699887359cd122d64b1e716f` (contains landed L15-1 `3113bad`); clean Story checkpoint, zero commits behind local main. Normative curator-spec main read through GitHub API at `1059b94b4135089d842bf017070b8ab47d79a536`: CIP-0017 implementation leaf 5, environments §§5.4, 7.1, 7.4, 8.2–8.4.1, 10.1–10.3, 12, v4 schema and Muse v4 conformance notes. Attached owner decisions D1–D14 and the corrected M1 leaf govern scope. No normative conflict found: the conformance notes' older generic credential diagnostic is refined by environments §7.4's explicit `environment_credential_store_unsupported`.

## Acceptance mapping before edits

| AC | Existing production implementation | Existing named tests | Missing evidence or behavior |
|---|---|---|---|
| 1: v4, effective file, five values, HOME, renderer parity | `envprofile.Resolve` → `buildFragment`, operation-local `credentialPolicy`, `envfragment.Fragment` JSON/EnvFormat/ShellFormat | `TestMuseCredentialPolicyMatrix`, `TestMuseCLIFragmentAndFakeLaunch`, `TestMuseCLIStorePolicy`, `TestMuseFragmentBoundary`, `TestMuseFragmentV4PublishedCases` | CLI test checks fifth value and count but does not compare all five parsed projections; strengthen existing test. Omission mutant already exists. |
| 2: verified roots iff fact, no fragment on failed required root | `Resolve` → `verifyHome`/`checkSurfaces` → `buildFragment`; `homePlan.rootContext` retains declared header-only surfaces | `TestMuseRootContextVectors`, `TestMuseRootContextDrift`, `TestMuseRootContextNarrowingMutants`, header-only mutant | Add unrelated-file negative, denied required root and failed repair, copy-fallback/current fact coverage; strengthen unchanged-state assertions. |
| 3: native/unset mapping, keychain refusal, shared pre-override source, isolated no link | Common `computeCredentialPolicy` and `inspectMuseAuth`; CLI config storage | `TestMuseCredentialPolicyMatrix`, `TestMuseMissingNativeFilePolicy`, `TestMuseCLIStorePolicy`, `TestMusePolicySwitchPreservesCredentials` | Reuse implementation and tests. No second policy resolver. |
| 4: preserve conflicts, readability, seeds; status evidence distinctions | `inspectMuseAuth`, seed provisioning in `applyHome`, metadata-only `StatusOf` | `TestMusePublishedLinkStates`, `TestMuseForkRefusedRegression`, `TestMuseAuthRepairAdditionalStates`, `TestMuseCLIFragmentAndFakeLaunch`, `TestMuseRootContextStatus` | Status omits 1.4.8 backend/replay qualification, interactive exclusion and explicit rules/skills distinction; effective backend/isolation report missing. Strengthen additional auth-state preservation/refusal assertions. |

## Mandatory paired table mapping before edits

| Leaf row | Existing evidence | Planned delta |
|---|---|---|
| Missing/drifted/denied required root; failed repair / linked, copied, current, repaired | `TestMuseRootContextDrift`, `TestMuseRootContextVectors` | Add denied/failed-repair/copy controls; retain existing test names. |
| Unrelated AGENTS/header must not mint fact / real declared root | `TestMuseRootContextVectors/no-context`, `/monolithic`, `/zero-applicable` | Extend no-context vector with unrelated files. |
| Extra env, constant, suffix/parent, HOME, adapter / exact ordered object | `TestMuseFragmentBoundary`, `TestMuseFragmentV4PublishedCases` (frozen normative corpus) | Preserve schema cases; strengthen semantic boundary errors/immutability for neighboring shapes. |
| JSON-only backend fails / five parsed projections equal | `TestMuseCLIStorePolicy`, `TestMuseCredentialPolicyNarrowingMutants/renderer-drops-only-backend` | Compare full maps and independent declared order, retain bounded omission plant. |
| Fork/retarget/dangling/nonempty/metadata denied / live and relink/empty | `TestMusePublishedLinkStates`, `TestMuseForkRefusedRegression`, `TestMuseAuthRepairAdditionalStates`, `TestMuseMissingNativeFilePolicy` | Run vectors explicitly; strengthen additional-state diagnostic and full-scope preservation assertions. |
| Native stored unchanged, keychain refusal / effective file | `TestMuseCLIStorePolicy`, `TestMuseCredentialPolicyMatrix` | Reuse existing integration behavior and byte snapshots. |

The brief carries an attack-surface description and paired test table, but no catalog surface table. This is a brief gap, not an implicit waiver: the handoff will map all six paired rows and four AC rows with narrowing controls. Runtime Muse authentication, refresh, consumer launch/hosted merge and exact 1.4.8 P1–P4 qualification remain outside M1 per its dependency/overlap clause; no qualification is claimed. Existing reviewer test names will be retained. CLI integration test files are a justified seam beyond the leaf's literal path list: existing public resolve/status tests already live in `cmd/curator` and the leaf explicitly requires these entry points.

Initial focused baseline: `go test ./internal/envfragment ./internal/envprofile ./cmd/curator -run 'TestMuse(CLIStorePolicy|RootContextVectors|RootContextDrift|RootContextStatus|FragmentBoundary|CredentialPolicyMatrix|MissingNativeFilePolicy|ForkRefusedRegression|AuthRepairAdditionalStates)$' -count=1 -timeout=5m`, exit 0. Not full-suite or runtime qualification evidence.
