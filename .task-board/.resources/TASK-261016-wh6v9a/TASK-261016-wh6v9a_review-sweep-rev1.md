# Review sweep — revision 1

Candidate: ad5e8e7f45c56718d5889c9c675d4a33765b3b27.
No catalog surface table or time budget was supplied; this review maps the narrowed producer requirements to catalog surfaces. Scope is the producer slice, with consumer work explicitly deferred. Seven of eight AC rows are production-driven; row 8 has named boundary/schema tests and an explicit wire-consumer bound.

| Surface | Result | Executed attacks and bounds |
|---|---|---|
| destination conflicts | held | TestMusePolicySwitchPreservesCredentials and TestMusePolicyExplicitMigration: detached file and shared-to-isolated repair refusal versus explicit migration; native bytes preserved. |
| read boundary | held | TestMuseAuthParentBoundaryAtResolve and TestMuseAuthRepairAdditionalStates: parent symlink, retarget, relative link, directory target; existing valid route controls. |
| error surface | held | TestMusePublishedLinkStates and TestMuseAuthRepairAdditionalStates: metadata/readlink errors refuse, not absence; missing native source guidance and no placeholder in TestMuseMissingNativeFilePolicy. |
| selector paths | held | CLI config set/show/unset and malformed/alias/dotted-profile cases, TestMuseCLIStorePolicy; forced keychain refusal beside native/file controls. Unqualified Codex activation still refuses. |
| text grammar | held | Unknown/missing/conflicting Muse policy and v1-v3 tokens in 67 frozen schema cases; CheckBoundary negatives; malformed machine config through CLI. Wire-consumer ingestion is explicitly out of this producer slice. |
| identity | held | TestMuseRootContextVectors and TestMuseRootContextDrift: header-only, no-context, missing/drifted surface and repair paths; no successful fact from a stale surface. |
| ordering | held | TestMuseCLIStorePolicy and TestMuseFragmentBoundary/valid: JSON/env/shell keep all five values and env/shell registry order with backend last. Keys are a fixed registry set, not locale-derived names. |

## Free hunt

TestReviewMuseIsolatedCredentialSwitch (overlay-only review test) passed: native-to-file on an isolated home preserves its vendor credential, while isolated-to-shared bare resolve/repair emits no fragment and preserves all credential scope state. Repair gives environment_credential_conflict. No new product source file was written.

Eight candidate narrowing mutants ran in the focused CLI command, including renderer-only backend omission and header-only root omission. The harness requires exit 1 and the named behavioral test failure, rejecting setup/compile failures. No blocking findings reproduced.

The supporting remote-gate script was also attacked in an isolated scratch repository with its exact candidate bytes: override signingkey to a missing public key, invoke the real shell script and real Git signing, and instrument push reachability. It refused with gate exit 1 on the missing key, and no push was reached. The real signed hosted snapshot is the positive control. This is an executed signing failure, not a forged verification result.
