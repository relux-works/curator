# Brief L151B: L15-1 remainder (TASK-261016-hzbo83)

Implement the rest of the attached plan leaf `rc7-leaf-credential-store-policy.md`, on top of slice 1 (task TASK-261016-2en3yu, which must already be on curator main when you start; if it is not, stop and report). Read rc7-code-common.md and rc7-decisions.md first.

The remainder:
- the unified effective source/store/isolation policy projection;
- the Muse file literal and shared admission;
- successor fragment composition per curator-spec 15d9701;
- the projection consumed by readiness and login;
- qualified application rows.

The leaf is the contract. Stay inside its paths. If it is still too large for one reviewable change, implement the next coherent part, narrow honestly and report. Leave the change uncommitted for the Change Request flow.
