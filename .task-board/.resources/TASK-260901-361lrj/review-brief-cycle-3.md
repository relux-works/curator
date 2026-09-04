# Review brief: cycle 3 (operator-direction delta)

Everything in `review-brief-0010.md` applies with these updates.

## Subject
- Worktree: `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-draft-environments`, branch `draft/agent-environment-profiles`
- Baseline: `fe21fb0` (cycle-2 ACCEPT point). Delta under review: `f188173`, `5a782af`, `4365a7d`, `0289328` (head)
- Document: `decisions/0010-agent-environment-profiles.md` (~850 lines now)

## What changed since the ACCEPT (review whole document, focus the delta)
Materialization forms (monolithic/referenced); generation header; system-prompt module class with fragment channel description; profile composition (overlays, chapters, later-overrides-earlier default); scoped switching (--env/--target); first-run onboarding contract (rev1 steps 1-4, import deferred); marker renamed .agent-environment.json; umbrella subcommand convention (curator-<name> discovery, curator run / curator session); launcher extracted to curator-agent-launcher repo with in-repo spec; Decision 10 rewritten as four-plane map (context/spawn/session/execution); ax always-when-configured; new/renumbered open questions 1-7.

## Cycle-3 specific checks (in addition to the six standing dimensions)
1. **Numbering integrity** — the open-question list was renumbered twice and the onboarding list was damaged and repaired mid-flight. Verify: OQ list is exactly 1..7 with no gaps/duplicates; onboarding list is 1..6; EVERY "open question N" / "question N" cross-reference in the document points at the question whose content matches the citation context.
2. **New factual claims** — verify against local binaries/docs where possible: claude_code `@path` imports + `hasClaudeMdExternalIncludesApproved` in `.claude.json`/binary; `--system-prompt`/`--append-system-prompt` in claude 2.1.251 and pi 0.84.2 help; `model_instructions_file` in codex 0.151.0 strings; pi `APPEND_SYSTEM.md` unconditional application from agent dir (dist source); `GEMINI_SYSTEM_MD` is docs-confidence (should be presented as such); opencode `instructions` list is docs-confidence.
3. **Four-plane map vs reality** — read https://github.com/relux-works/skill-agents-management SKILL.md (via `gh api repos/relux-works/skill-agents-management/contents/SKILL.md --jq .content | base64 -d`) and confirm the decision's characterization of the spawn plane (builds plans, does not execute; knows systems/vendors/models/limits) is accurate and that the declared "only import edge" (launcher -> agents-management module) does not contradict it.
4. **Internal coherence of the extraction** — no stale references to a rich `curator launch` outside Rejected alternatives; Decision 2 system-prompt rules vs Decision 6 launcher application vs Security bullets vs phasing rows must tell one story (materialization=curator, application=launcher, ax always-when-configured, native homes never).
5. **Determinism story survives the header/composition/forms additions** — generation header deterministic (no timestamps), chapters deterministic, referenced form's output set well-defined enough for the promised byte-exact vectors, or flagged as normative-phase work.

## Verdict contract
Same as before: `review-findings-3.md` resource on TASK; blocking/major -> status development; otherwise ACCEPT explicitly. Note residue `tools/__pycache__/` is not subject matter.
