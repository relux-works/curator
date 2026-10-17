# Brief L151C: rc.7 credential store policy, remaining workstreams (TASK-261016-wh6v9a)

TASK-261016-hzbo83 landed the first part of the L15-1 remainder: one operation-local projection of the native managed-home credential-store policy, shared by Resolve, StatusOf and migration (`internal/envprofile/credential_policy.go`). Its scope note listed what it deferred. That deferred work is this task.

Implement the attached leaf `rc7-leaf-credential-store-policy.md` for these requirements, on top of the landed projection:
1. Muse file literal and shared/isolated successor admission.
2. Successor fragment composition, per curator-spec main.
3. Readiness and login consume the same projection. Never add a second policy derivation path.
4. Qualified application rows. Unqualified rows refuse fail-closed. Execute the hosted-override plant control.
5. Adopted environment/node/helper/broker source policy.

Read rc7-code-common.md and rc7-decisions.md (attached) first. The leaf is the contract. Read the cited normative sections on curator-spec main yourself; where the leaf and the normative text disagree, the normative text wins: stop and report.

If this is too large for one reviewable change, implement the next coherent parts in the order above, narrow honestly in the handoff, and say exactly what remains. Do not fake the rest. No runtime qualification is claimed unless you executed it. Leave the change uncommitted for the Change Request flow.
