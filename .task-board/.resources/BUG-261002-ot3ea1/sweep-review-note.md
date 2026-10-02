# Review note — BUG-261002-ot3ea1 rev2 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Same-provider review (operator rule, R138). Review against `sweep-brief.md`, `sweep-evidence-mini.md` and `sweep-gate-note.md`. The hosted gate is green.

Verify, with real exit codes and using `-work` per host-rules:
1. **The build-cache sweep** (after `curator global upgrade` / gc) retains every cache build whose binary is the executable of a live process, on macOS (libproc / sysctl), Linux (/proc/<pid>/exe) and Windows (the process image path). Enumeration errors fail SAFE: keep the build and warn. Long-lived daemons are covered: the mini case had two tb-sessiond processes on swept builds.
2. **The /proc read path.** The state-read guard handles it with a precise, justified allowlist entry or a seam route. The gosec G304 fix builds the path only from a validated numeric pid; no blanket nolint.
3. **Tests.** Rows for in-use retained, unused swept, enumeration error retained with a warning, daemon rule unchanged. Kill at least one mutant yourself: drop the liveness check.
4. **Hygiene.** One CHANGELOG line; no LOGBOOK; never spell any employer name.

accept_cr, or changes requested with file:line.
