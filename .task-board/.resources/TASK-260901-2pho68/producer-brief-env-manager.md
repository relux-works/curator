# Producer brief: manager-profile and CLI sections for environments

## Setup
- Base: fresh `origin/main` of `~/Developer/ReluxWorks/curator-spec` — MUST be `c3b29b1f7f37829fd4d0c50b2023efa2feb4c615` or later. Fetch and verify.
- Worktree `~/Developer/ReluxWorks/.worktrees/curator-spec-env-manager`, new branch `draft/environments-manager-profile`. Work only there; shell tooling; signed commits.

## Task
Extend the manager conformance class and informative CLI guide for the environments capability, consistent with `protocol/environments.md` (normative source of truth — cite it, never restate divergently):
1. `profiles/manager.md`: new sections covering the environment adapter registry behavior (four adapters, secondary fixed-home targets with auto|off|explicit probing, shadowing-path warnings), materialization modes and mode defaults, profile lifecycle (install/use/sync, scoped switching, revision-1 onboarding: detect / foreign-manager stop / replace notice / backup / takeover), credential passthrough per platform, `env resolve` verify-and-repair behavior, always-strict profile audit and the `context-secret-material` detector class in §7, and status/GC extensions (marker live roots). Match existing section style, diagnostics tables referencing environments.md codes, no duplication of normative byte rules — reference them.
2. `cli/curator.md`: extend the informative command table (`profile install|list|use|sync`, `env resolve`, `env status`, `global --profile/--all-profiles`, umbrella note for `curator run`/`curator session`), plus a short usage example block.
3. `make validate` (links) green.

## Deliverables
Signed commit(s); board resource `env-manager-notes.md` (sections added, any normative gap discovered — file it, do not fix protocol prose); handoff to-review.

## Do not
Touch schemas/, conformance/, protocol/ (sibling story), CHANGELOG; push; tag; mark done.
