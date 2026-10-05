# Review note — research/CIP draft review (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Same-provider review (operator rule, tb-R164). This is a research deliverable: a CIP draft plus evidence under .research/, written to the task brief and cip-template.md. Verify:
1. **Scope and operator decisions.** The brief's operator decisions are honoured (quoted in the brief). Every required question is answered, or listed as open with a recommended answer.
2. **Evidence.** Spot-check at least 8 file:line or spec citations against curator main and curator-spec rc.14. Every claim marked "verified" or "measured" has a reproducible probe in the evidence file. Nothing claims more than was measured.
3. **Safety.** No real credential was read, printed or copied. No login or logout was done on the operator account. No internal hostnames, personal paths or employer names appear; the board is public.
4. **Decision-readiness.** It has 2–4 real options with tradeoffs, one recommendation with a concrete control surface (config shape, defaults, precedence), a spec-changes section, implementation leaves and a test plan.
5. **Diff hygiene.** Only .research/ files changed. LOGBOOK.md is untouched. No product code changed.
accept_cr if all hold. Otherwise request changes, quoting the exact lines at fault.

## HOSTED-EVIDENCE MODE (desk #52, binding)
The Mac mini has exec stalls: do not run go build or go test locally. Verify citations by reading files and git, which is cheap. Keep board commands minimal: one outcome resource and the verdict.
