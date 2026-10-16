# SNAP-LOCK — Non-recovering capture transaction

Owner repository: **curator**. Size: **3**. Required in rc.7.

Normative clauses: snapshots §§1 and 5, environments §§9.2/10.1 and existing read-failure rules; CIP-0014 consistency steps 1–5 and tests `MaterializeCLI_NoTornGeneration` / `MaterializeCLI_SourceRefusals`.

Attack surface: capture accidentally mutating its source, or publishing a mixture of generations while retaining the same fragment digest.

Acceptance criteria:

- Acquire the existing manager-home mutation lock with its bounded wait through a new non-recovering path. Do not call `beginOperation` or `beginOperationValidated` as-is: curator `0d00a109`, `internal/envprofile/lock.go:37–70`, both lead to `engine.Recover`.
- Read pending transaction state without running recovery or cleaning journals. A pending transaction returns `environment_materialize_recovery_required`; unreadable/corrupt evidence must fail closed with its appropriate existing refusal, not be treated as no transaction.
- Reselect profile and effective machine policy under the lock. Revalidate selected profile/policy, lock, marker, surfaces and every captured object's identity before release. Do not reuse selection made while waiting or recursively enter locking repair.
- Share a side-effect-free content verification core with metadata-only status work. Existing `verifyHome` invokes the vendor release probe (`managed.go:1957 → checkToolRelease → detected → detectRelease`); neither materialize form may execute it. Safe unavailable release observations stay unknown. A test-only Detect override is not a production implementation of this boundary.
- Bound capture; release the lock before publication transport or external service waits. Normal lock bookkeeping is distinct from forbidden source provisioning/repair/recovery. An absent manager/home must not be implicitly provisioned by the lock helper.

Mandatory tests with positive controls:

| Positive control | Negative case / required invariant |
| --- | --- |
| Current verified source, no pending journal, free lock yields one captured generation | Prepared/committing/cleanup/rollback journal refuses without journal, source or target mutation; instrumentation proves Recover was never called |
| Lock released within bound permits capture from the newly selected profile | Timeout returns `environment_lock_unavailable`; profile/policy changed while waiting cannot select the stale pre-lock inputs |
| Cooperating repair/use/global-sync completes before or after capture | Concurrent multi-file changes yield one verified generation or `environment_materialize_capture_drift`, never torn bytes |
| Stable fragment and stable verified objects succeed | Same-fragment ABA, external store-byte/mode/identity change or failed revalidation refuses; fragment equality alone is insufficient |
| Current source verifies and captures using file/metadata evidence | A vendor probe executable planted on the supported discovery path must never run, including the successful path and manifest-only; no provider/credential side effects |

Control plant: narrow revalidation to fragment/lock equality (M3) and verify a same-fragment mutation witness is rejected by the test. No implementation tests run in this collective.

Dependencies/paths: new `internal/envprofile/materialize_capture.go`; narrow shared edits to `internal/envprofile/lock.go` and a read-only pending-recovery observer in `internal/transaction/`. Existing `managerlock.AcquireHomeOnly` is a primitive, not proof of non-provisioning semantics (`manager.prepare` can create manager directories). Serialize shared lock/managed-profile edits with login/store and profile-prepare leaves. SNAP-CLOSURE supplies the content capture callback; SNAP-CLI integrates both.
