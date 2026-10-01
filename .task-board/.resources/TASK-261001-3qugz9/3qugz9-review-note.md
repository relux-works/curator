# Review note — TASK-261001-3qugz9 research (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review the research results (op2-product second-operator readiness answers) for truthfulness before they go to tb-keeper and to another team.

Spot-check, with real exit codes:
1. **R2 claim.** In a temp non-git dir with a throwaway HOME, does `curator install` on current main skip local skills with exit 0? Use the attached probe.py, or a minimal repro. Confirm the gitignore gate cited (C6) is the cause.
2. **R1 claim.** Stable v0.14.0 already has profile/env/run entry points (`git show v0.14.0:cmd/curator/main.go`, around :73). Launcher v0.1.0 is tag-only, with no release assets.
3. **R6 claim.** Credentials are excluded from profiles, and the Codex `file` link points at the operator's own auth.json, with citations to managed.go.
4. **Citations.** Pick 3 random file:line citations and confirm they say what is claimed.
5. **Draft notice.** It must not overclaim readiness; it must say "remain on current setup".
6. **Hygiene.** No secrets and no employer name in any attached file.

accept_cr if accurate; otherwise changes requested with exact corrections.
