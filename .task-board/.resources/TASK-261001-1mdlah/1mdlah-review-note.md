# Review note — TASK-261001-1mdlah rev1 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

curator-agent-launcher CR, base ee66c107, 14 paths. Review it against `launcher-muse-int-brief.md`, including its two binding updates (v0.5.37; unlisted-release row). It is the last piece of curator#100: an interactive `curator run muse` root session.

Check, with real exit codes:
1. **Pin.** agents-management is pinned at v0.5.37 (go.mod/go.sum). README and SPEC name it.
2. **Root-entry rows**, driven by a fake curator (v3 fragment) and a fake muse that prints the exact upstream `--version` format:
   - 1.4.2 builds an INTERACTIVE plan;
   - native → no posture flag; yolo and the alias → exactly one `--yolo`;
   - env: the four XDG overrides; HOME inherited and never set;
   - the unlisted release (cite which version and the policy row) → refused, non-zero, no child started.

   Kill these mutants yourself: the HOME-set mutant and a double `--yolo`.
3. **Goldens.**
   - The Claude goldens add ONLY `CLAUDE_CODE_ENABLE_PROMPT_SUGGESTION=false` (curator#102, launcher half).
   - The new muse goldens match the asserted argv.
   - Nothing else changed versus v0.5.22, the base pin. Diff the goldens.
4. **Windows vet** is identical to baseline: compare the line multisets.
5. **Hygiene.** No stray files (in particular no `*_recovery-results.md`), no LOGBOOK, CHANGELOG under Unreleased. Never spell any employer name.

The host has syspolicyd exec stalls. Check `launchctl print system/com.apple.security.syspolicy | grep -E "state|successive"` before long commands and wait while it is down.

accept_cr, or changes requested with file:line.
