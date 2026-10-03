# Review note — TASK-261002-2ipeqa rev2 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Same-provider review (operator rule, R138); release-critical. Curator CR, base 64345d71, 16 paths, hosted gate green. Review it against `rc14-lockstep-brief.md` and `rc14-lockstep-republish.md`.

Verify, with real exit codes and following host-rules (-work, syspolicyd backoff):
1. **The rc.14 digest.** 6f832d81… from curator-spec branch `candidate-rc14` (e3a88ced) is recognised with EXACT per-family counts and gap rows. Recompute at least 2 families independently from the candidate corpus.
2. **Fixed cases.** The seven cases fixed by 2afgyq and 9w4wy3 are NOT listed as rc.14 gaps.
3. **Older suites.** rc.13 and the previous candidate digests are still accepted unchanged.
4. **Switches.** `internal/hashing` EnableV2Writers is still false, and `.github/workflows/ci.yml` SPEC_PIN is unchanged.
5. **The spec's exact check.** Run the spec workflow's Implementations command against `candidate-rc14`, or confirm it from the hosted evidence.
6. **Hygiene.** No LOGBOOK; never spell any employer name.

The host has exec stalls. If local runs stall, rely on the hosted evidence for the full matrix, and say exactly what you ran yourself.

accept_cr, or changes requested with file:line.
