# TASK-260927-4pv4au review verdict — rev2 (tree 1390c2f5, base 6bd98d49): CHANGES REQUESTED

Verified: worktree temp-index tree == 1390c2f5. Local (zsh, pipefail): `go test ./internal/config -run 'Posture|Security|Hardened|Permissive'` rc=0; `go test ./cmd/curator -run 'SecurityPosture|TestStatusJSONKeepsTheLegacyShape'` rc=0.

Accepted parts: one revision switch (internal/config/security_posture.go:8), permissive default, loadConfig warns on c.stderr once; `curator status --json` gaining `security_posture_rows` is spec-required (manager.md §10 L1459-1492: curator status carries header + 4 rows on schema-1) so status_test.go:1785 change is justified.

## F1 (blocking) — permissive warning still emitted on the launch path; gate fix hid it by changing tests
- cmd/curator/main.go:214-216 (runEnforcedShim) prints `warning: security_posture_permissive…` to os.Stderr of the enforced launcher, i.e. into the stream of the launched script/tool. 4pv4au-gatefix-1.md is explicit: never into "script/launcher/worker protocol streams or the output of a launched command; the launch/exec path … does not emit it unless the spec names that surface". manager.md §7.1 (L1169) says "every operation" and names reporting via status (section 10); it does not name the script launcher shim.
- Instead of removing the emission, rev2 changed the three failing tests' fixtures to `schema_version: 2, security_posture: hardened` (internal/scriptworker/derive_test.go:1138, internal/install/scriptpolicy_test.go:671) so the warning is not produced. Every schema-1 machine (permissive by spec) and every rev-A default machine still gets the warning injected into each launched script's stream — exactly the failure the gate caught.
Required: drop the emission from runEnforcedShim (restore the "streams belong to the caller's pipeline" rule), restore the original schema-1/permissive fixtures in both tests (keep a hardened variant if wanted), and add a negative test that a permissive launch produces no `security_posture_permissive` bytes on the launched command's stdout/stderr. Mutant: re-adding the emission must fail that test.

## Not re-verified this round (bounded)
Full mutant set and vector ratio not re-run by me; re-review after F1.
