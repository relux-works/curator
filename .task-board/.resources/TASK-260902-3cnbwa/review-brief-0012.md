# Review brief: Decision 0012 draft (context packages, semver locks, launch-channel MCP)

## Subject
- Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-decision-0012`, branch `draft/decision-0012-context-packages`, base `4d55698` (= main).
- Document: `decisions/0012-context-packages-and-semver-locks.md`.
- Inputs to judge against (read-only): `decisions/0010-agent-environment-profiles.md`, `protocol/environments.md` (revision 1, landed), `protocol/core.md` §2, §4.4, §6.1–6.3, §7, §8, §10; `registry.md` §1 (CCJ-1); `profiles/manager.md` §1, §5, §6, §12; the pre-implementation review (board resource `pre-implementation-review-v3.md` on STORY-260901-zddtn8, 16 MUST items); the operator's settled answers recorded in the story description (semver yes; weight direction configurable; direct requirer + root final word; overlays via git or path; MCP in revision 1; strict semver 2.0 `v` tags, npm-style ranges + `latest`).

## Review dimensions
1. **Model soundness**: context package as a second kind on the skill closure engine; umbrella-as-convention; lock as identity; §7 one-name-one-commit preserved after ranges; joint resolution of root + overlays + skills + MCP; weights never resolving version constraints; effective-weight rule (default → direct requirer edge → root `weights` map) — attack for ambiguity, cycles, non-determinism, and whether any rule can be read two ways.
2. **Semver and range grammar**: strict SemVer 2.0 with mandatory `v`; caret/tilde/comparators/`||`/wildcards/`latest`; prerelease rule; total order; conflict diagnostics. Verify the stated npm semantics are correct (caret on 0.x, tilde, prerelease matching) — cite npm's semver rules; flag any deviation the text implies.
3. **Precedence primitives** (`winner`, `placement`): closed, orthogonal, default reproduces Decision 0010's `later-overrides-earlier`; tie-break rule complete; header states both.
4. **MCP launch-channel design**: declaration-only, no execution, values never present, allowlist lockable, managed homes only; channel descriptors per adapter — verify `claude --mcp-config`/`--strict-mcp-config` exist on the installed claude binary, `codex -p <name>` profile layering semantics from `codex --help`, `OPENCODE_CONFIG` from opencode docs; flag anything asserted beyond evidence.
5. **Impact accounting**: the list of environments.md sections rewritten vs unchanged is accurate (check each named section against the landed text); header type-line bump justified; `Profilefile.json`/`context.json` withdrawal safe (no implementation, no tag); core widening (ref.kind semver via skill schema 9 + Skillfile schema 2) is the minimal correct compatibility path under COMPATIBILITY.md.
6. **Consistency with Decision 0011 Option A** (curator-run composes, ax `--launch-plan`, SpawnPlan stdin) and with the pre-implementation review's MUST items — no regression, no new contradiction (e.g. M4 seeds, M6 `path_prepend`, M10 read-only resolve).
7. **House style**: matches decisions/0009 and 0010 (sections, tone, density); English; every § citation verified.

## Constraints
Read-only: no edits, commits, pushes. Shell tooling; both reference checkouts read-only (`~/Developer/ReluxWorks/curator-spec` main; installed binaries for verification).

## Verdict contract
`review-findings-0012-<n>.md` on the task: severity (blocking|major|minor|nit), section, quote, what is wrong, fix. Blocking/major → status development; else ACCEPT explicit, leave to-review. Do not mark done.
