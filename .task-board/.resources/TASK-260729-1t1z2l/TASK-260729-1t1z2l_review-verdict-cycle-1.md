# Review verdict cycle 1 — CHANGES REQUESTED

**Task:** `TASK-260729-1t1z2l`  
**Review date:** 2026-07-29  
**Verdict branch:** changes requested → `analysis`

The parity map is strong on 17-task coverage, Curator package/test counterparts,
the existing CocoaSkills dependency DAG, repository provenance, and the
no-publication/no-pin boundary. It is not acceptable yet because its central
rc.4-to-rc.5 delta and platform handoff are materially wrong. The corrections
are research/artifact rework, not product-code rework and not a stop-the-line
blocker.

## Accepted portions

1. The live CocoaSkills story contains exactly 17 pre-existing Go tasks plus
   this reconnaissance task. The 17 IDs and the 17 parity-table rows match
   exactly, with no omission or duplicate.
2. Every mapped Curator package/test path cited in the table exists in the
   accepted `TASK-260720-1ljev5` composite. Focused tests for `buildmeta`,
   `godriver`, `buildcache`, `buildsource`, `skillspec`, `skillcheck`,
   `whitelist`, `marker`, `managerlock`, `transaction`, `runtimestore`,
   `scopes`, and `staging` pass.
3. The task ordering matches the live board edges. The only two CocoaSkills
   roots are `TASK-260720-z9j4c9` and `TASK-260720-z2z795`, both blocked by
   `TASK-260720-1pvfj5`; the downstream joins in the map match the board.
4. Repository provenance was independently re-resolved:
   - Curator `origin/main`: `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`.
   - curator-spec `origin/main`: `57c1f56846d221ecc55786bd3c2467ec32f11730`;
     no rc.4 or rc.5 tag exists.
   - CocoaSkills local clean `main`: `edce8816dda44bb121d661b7c4dea942558ce408`,
     exactly two commits behind verified `origin/main`
     `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`.
5. The publication/pin boundary is correct: no working-tree candidate is a
   release, no manager pin may advance from candidate evidence, and accepted
   `TASK-260729-1kq1rd` recommends board-owner approval of rc.5 supersession
   followed by landing the exact accepted `TASK-260728-2kp3tv` snapshot.

## Required correction 1 — rc.5 changes local canonical identity

The map repeatedly says schema-6/local-driver metadata, receipt, marker, and
expected bytes remain frozen from rc.4. That is false at the protocol behavior
surface that CocoaSkills must implement.

Independent comparison:

- rc.4 `common.schema.json` does not require an execution policy in
  `goBuildPolicyV1`; its canonical local input has no `execution_policy`.
- accepted rc.5 adds required
  `"execution_policy": "manager-worker-v1"` to `goBuildPolicyV1`.
- rc.4 legacy canonical input/key:
  `sha256:3fcd714a40e8918eb67dbd35d435875dcce6c9047da811a1fa26626e5e57be48`.
- rc.5 portable canonical input/key:
  `sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b`.
- rc.4 receipt hash hard-coded in the CocoaSkills task:
  `sha256:750f5f7557ffe1dac1ec06b7d47cb5976e390a7dc7d57ea616c9edda91013c11`.
- accepted Curator rc.5 receipt hash:
  `sha256:919fbbad8e6ce95532219fd952c2309d0d7026f85209650508fd6834af4020cd`.
- rc.5 `go-host-execution-policy.json` explicitly marks the rc.4 input as
  `legacy_rc4_without_execution_policy`, `schema_valid: false`, and
  `aliases: false`.

Some top-level schema files are byte-identical between candidates, but they
reference the changed shared definitions. Byte identity of those files is not
semantic identity of canonical inputs, keys, receipts, or marker references.
The map must distinguish frozen schema-6 declaration shape/top-level artifacts
from the changed rc.5 execution-policy-bound local identity.

This directly changes the mapping and handoff for
`TASK-260720-2dnqw2`: its current acceptance criteria hard-code the obsolete
rc.4 key and receipt hash. The parity map must name the rc.5 values above,
require `execution_policy` in the canonical policy, and classify the rc.4 input
as a negative non-alias case.

## Required correction 2 — precise CocoaSkills brief/dependency retarget

The statement that stale rc.4 wording exists “across the 17 CocoaSkills
briefs” is not precise. A scoped board search finds literal rc.4 wording in two
of the 17 task briefs: `TASK-260720-12r55p` and
`TASK-260720-3s27te`. Other briefs are version-neutral but still encode
pre-amendment assumptions. The corrected map must enumerate at least:

| Task | Required retarget |
| --- | --- |
| `TASK-260720-2dnqw2` | Replace rc.4 key/receipt goldens with rc.5 execution-policy-bound input, key, receipt, and marker-reference expectations. |
| `TASK-260720-2g21eg` | Replace the direct source-aware Go process boundary with the hidden identity-verified `manager-worker-v1` re-execution, authenticated one-list/one-build session, native-control preflight, teardown, post-identity checks, and `capability-evidence-v1`. |
| `TASK-260720-12r55p` | Change literal rc.4 candidate consumption to the landed rc.5 root/digest and add `go-host-execution-policy.json`, native-control inventory, capability-evidence, and legacy non-alias cases. Keep its hard edge to `TASK-260720-3ag6pi` only if the owner retargets and re-reviews that task as the rc.5 verification gate; otherwise relink it to the replacement rc.5 verification task. |
| `TASK-260720-akf5kh` | Document the portable manager-worker boundary, honest capability evidence, macOS/Windows support claim, and deferred Linux boundary. |
| `TASK-260720-3pemm6` | Replace the current Linux/macOS/Windows claim with current rc.5 native macOS and Windows gates; route Linux to later qualification unless the protocol owner first extends the inventory and accepts a new contract. |
| `TASK-260720-3s27te` | Replace literal rc.4 root/hash and three-OS final gate with the landed rc.5 root/digest and current macOS/Windows qualification; keep Linux explicitly deferred. |

`TASK-260720-th0jdi` should also state that execution-policy mismatch is a
currentness/rebuild dimension, although its complete-key/receipt checks already
cover that transitively.

## Required correction 3 — Linux is later, not a current rc.5 gate

Accepted rc.5 `protocol/core.md` section 4.2.1 says every conforming manager
MUST implement `manager-worker-v1` on macOS and Windows. The native-control
inventory contains only those two platforms. Accepted Curator
`controls_other.go` rejects platforms outside the inventory.

The attached map instead requires a valid Ubuntu build in the current
CocoaSkills final gate. That contradicts the accepted contract and the review
directive's “later Linux qualification” boundary.

The revised artifact must:

1. treat local macOS as the primary current gate;
2. treat native Windows through SSH alias `win` as the second current gate;
3. remove Linux success/claim requirements from the current 17-task rc.5
   delivery chain;
4. identify later Linux ownership rather than silently dropping it:
   `TASK-260728-1e6811` is the deferred cross-manager Linux toolchain-preflight
   task, while `STORY-260728-1eye8p` owns later native Linux external-repository
   qualification. If local `go-v1` needs a separate full Linux lifecycle gate,
   the owner must create or designate that task; it cannot be inferred from the
   macOS/Windows rc.5 inventory.

## Required correction 4 — exact platform/toolchain readiness

- Primary host verified: macOS 26.5 arm64, Go `1.25.5`; this satisfies the
  currently accepted Curator family allowlist.
- `ssh win` is reachable and reports Windows NT `10.0.19045.0`, but `go` is
  **not on PATH**. No host install was performed. Before any Windows native
  driver/cache/transaction gate, the handoff must provide an approved Go-family
  installation or an exact operator-trusted absolute Go path plus its
  fingerprinting/tuning evidence.
- The 17-task Go chain owns Go-specific trusted identity
  (`TASK-260720-3j8pp5`). The later generic declarative toolchain-preflight
  implementation is separately owned by `TASK-260728-1j72zq`, which is
  downstream of `TASK-260720-3s27te`; the parity map must state this adaptation
  boundary so neither task absorbs the other's scope.

## Test and environment evidence

- The producer's exit 143/no-space event did not reproduce as a product
  failure. A first replay with 12 GiB free reached the suite but temporary Git
  commits inherited the host's mandatory SSH signing and failed on an encrypted
  key. That is host configuration contamination.
- A second replay disabled signing only through per-process Git config and
  supplied a disposable identity. Previously failing packages cleared. It was
  stopped after a separate task's concurrent full Go/race run reduced free
  space from 12 GiB to 2.6 GiB, to avoid manufacturing another storage failure.
  No no-space error appeared before the controlled stop.
- The focused Curator packages listed above pass. The repository-wide gate is
  still not claimed green because a clean serialized full replay was not
  completed.
- CocoaSkills baseline: the full run reported 482 passed, 17 protocol tests
  skipped, and one CLI failure caused solely by placing `TMPDIR` inside the
  Curator Git worktree; rerunning that test with `/private/tmp` passed. The
  official strict `python -m mypy` gate passes for 55 source files.
- Curator product/config surfaces and CocoaSkills remain unstaged and unchanged.

The revised parity artifact and its logbook summary must incorporate these
corrections, retain the verified 17-row map and DAG, and return for another
review cycle.
