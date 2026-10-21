# Brief: final-form spec, exact-plan hosted launches on snapshots (curator-spec, Draft)

**Repository:** curator-spec. **Branch:** `exact-plan-final-form`, from `origin/main`. Status: **Draft**. The owner wants the final form written now, so that implementation can start right after the first CIP-0002 slice.

## Context

- CIP-0014 (managed-home snapshots, Accepted) and CIP-0016 (local hosted launches, Accepted) define the target: hosted sessions start from an immutable snapshot, with an exact plan bound to its digest, and the host never composes.
- On 2026-10-21 the owner accepted the interim `launch-delegated/1` route: the manager composes against the live home. See the CIP-0016 amendment on branch `cip16-delegated-launches` (in review) and the launcher's FRONT-DOOR §5.2 and §5.8 (board resources `front-door-appendix-A.md` and `front-door-5.8.md` on TASK-261021-yqt1qr). The interim route retires per runtime once the manager advertises the exact-plan capability for that runtime, **together with a detached form**.
- `launch-exact/2` today cannot carry a doorbell or return a session record, so detached launches and launches with a doorbell (`werkstatt up`) must use the delegated route. The final form has to close that gap.
- The first wave of runtimes is Claude, Codex and Muse.

## Write

A Draft CIP, or a Draft revision of CIP-0016, whichever the repository's process prefers for a successor capability. Say which, and why. It specifies:

1. The exact-plan successor capability (`launch-exact/3` or similar): the closed payload, the snapshot digest binding, the session record returned, and the **detached form and doorbell support**. With those, `werkstatt up` can run on exact plans.
2. Per-runtime qualification (Claude, Codex, Muse), plus how the joint three-runtime rule of CIP-0016 §1 relates to per-runtime retirement of the delegated route.
3. The retirement mechanics of `launch-delegated/1` per runtime: what the manager advertises, how curator-run selects, and the refusal codes.
4. How a project-context composition (CIP-0002: stable home, lease, admitted generation) becomes a snapshot input, so project context survives the move to exact plans. Cross-reference the CIP-0002 candidate text on branch `cip02-ready-for-acceptance` if it is pushed; otherwise state the interface you assume.
5. Resume, recycle and GC of snapshots referenced by live sessions.
6. Security considerations, compatibility and migration from delegated sessions, the conformance vectors to add, and the open owner questions with recommendations.

No wire schemas are published in this draft beyond the CIP text. The schemas belong to the launcher and manager owners and are listed as follow-ups.

## Hand-off

Results file with the summary, open questions and files changed. Branch pushed, signed.
