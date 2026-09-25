# Review note — BUG-260922-3v8k23 internal/install timeout budget (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review the newest revision against `3v8k23-brief.md` and the task DoD: measured baseline and ≥2x headroom per ubuntu/macOS lane
(elapsed/budget table from the handoff gate, ≤50 %); every lane's timeout expression pinned in gate-selftest.sh, including the Windows
candidate-suite line (kill a mutant that changes one lane's budget); a synthetic hanging package still fails within budget (negative row,
real exit code); per-package timeout semantics kept; no CHANGELOG edit (entry in results). Revision 1 failed its gate and a successor
published revision 2 — name what changed and why. Bounded runs of gate-selftest only (host memory is tight). accept_cr or changes
requested with file:line. No LOGBOOK.md.
