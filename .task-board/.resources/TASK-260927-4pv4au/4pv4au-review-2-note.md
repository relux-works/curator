# Re-review — TASK-260927-4pv4au rev3 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Your rev2 verdict: F1 — the permissive warning emitted on the enforced launch path (cmd/curator/main.go runEnforcedShim) and hidden by
switching fixtures to hardened. Rev3: tree c7e73fa2 (base 6bd98d49), gate green; vs rev2 it changes cmd/curator/main.go,
cmd/curator/security_posture_test.go, internal/install/scriptpolicy_test.go, internal/scriptworker/derive_test.go. Verify:
1. runEnforcedShim no longer emits `security_posture_permissive`; the ORIGINAL schema-1/permissive fixtures are restored in both tests
   (a hardened variant only added); a negative test proves a permissive and a schema-1 launch put no warning bytes on the launched command's
   stdout/stderr; mutant (emission re-added) fails it — real exit codes.
2. Then run the deferred rev2 checks: the remaining review-note items (hardened effective defaults, locked precedence, refusals, --check
   contradiction, vectors except the 1sapuy pair and B cases, gap attribution, mutants). accept_cr or changes requested with file:line.
No LOGBOOK.md.
