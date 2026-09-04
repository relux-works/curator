# Lens A: threat model and security review of the environments capability

## Corpus (read ALL, landed main)
- `~/Developer/ReluxWorks/curator-spec`: decisions/0010-agent-environment-profiles.md, protocol/environments.md, profiles/manager.md §12 (+§4/§5/§6/§10 context), cli/curator.md, schemas/v1/{profilefile,context-manifest,agent-environment-marker,launch-env-fragment}-v1.schema.json, conformance/v1 environments vectors.
- `~/Developer/ReluxWorks/curator-agent-launcher`: SPEC.md (0.1.2-draft), README.md, cmd/curator-run.
- `~/Developer/ReluxWorks/agent-session-manager-spec`: SPEC.md v0.5.0 (+ the open PR #1 delta: `gh pr diff 1 --repo relux-works/agent-session-manager-spec`).
- skill-agents-management SKILL.md (`gh api repos/relux-works/skill-agents-management/contents/SKILL.md --jq .content | base64 -d`).
- Reference implementation `~/Developer/ReluxWorks/curator` (read-only) for feasibility grounding.
Read-only everywhere. Do not edit, commit, or push. Verify claims against binaries on this machine (claude, codex, pi, gemini installed) where relevant.

## Output contract
Board resource named as instructed, structured: severity (critical|high|medium|low), area, exact quote/section, why it is a weakness or gap, concrete strengthening proposal (spec text direction, not vague), and whether it MUST be resolved before implementation starts or can follow. Rank by cost-of-being-wrong before expensive implementation. Do not re-litigate accepted decisions unless you find a concrete failure; do not pad with style nits. Mark task to-review when done; do not mark done.

## Your lens
Adversarial security review of the WHOLE design as it will be implemented: profile repositories as an attack vector (malicious/compromised company or personal profile repo, overlays composing hostile instructions, skill-set composition shadowing mandated skills, prompt-injection surface of root context and system modules, the strict-audit + `context-secret-material` detector as both control and DoS/false-positive hazard with no pin escape), environment-variable injection boundary (fragment names/values, umbrella PATH dispatch, XDG seeding exposing `~/.config` links inside managed parents), credential passthrough by symlink (atomic-rename pitfalls turning links into files, isolated mode feasibility on macOS Keychain for claude_code), managed homes as new writable state roots (who can write, TOCTOU between resolve-repair and a running session), secondary fixed-home targets writing into another application's support directory, backups holding sensitive prior context forever (never GC'd), launcher as the one component executing plans from three sources, ax extension keys as a provenance channel, enterprise lockability (manager §1 system-config locked keys vs. composition/profile settings). Produce a threat table (asset, threat, existing control, gap, proposal). Name at least the top 5 things that MUST change before implementation.
