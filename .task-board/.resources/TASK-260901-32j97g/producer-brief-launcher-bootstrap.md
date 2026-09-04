# Producer brief: bootstrap curator-agent-launcher

## Setup
- Create a fresh local repository at `~/Developer/ReluxWorks/curator-agent-launcher` (`git init -b main`). It will be pushed to the already-created empty remote `https://github.com/relux-works/curator-agent-launcher` by the orchestrator AFTER review — you do NOT push, add remotes, or tag.
- Signed commits (`git commit -S`, verify each). Sensible commit granularity (skeleton, then spec is fine; one commit also fine).

## Normative inputs (read before writing)
- `~/Developer/ReluxWorks/curator-spec/decisions/0010-agent-environment-profiles.md` — Decision 6 (resolution primitive, umbrella subcommands, the launcher: boundary, curator-agent-launcher home, in-repo spec, ax always-when-configured, system-prompt application + warnings) and Decision 10 (four-plane map).
- `~/Developer/ReluxWorks/curator-spec/protocol/environments.md` — §10 launch-env-fragment-v1 (fields incl. system-prompt section), umbrella discovery convention, §7.3 system-prompt channels.
- skill-agents-management SKILL.md (`gh api repos/relux-works/skill-agents-management/contents/SKILL.md --jq .content | base64 -d`) — BuildPlan/BuildLaunch value contract (binary/argv/env/stdin), plugin graph, provider limits.

## Deliverables (repository skeleton + spec draft)
1. `README.md` — what the launcher is (execution plane composing spawn/context/session planes), status: specification draft, not yet implemented; install/discovery note (`curator run` via `curator-<name>` PATH discovery).
2. `SPEC.md` — the in-repo specification draft (this is the substance):
   - scope and non-goals (no session state — fire vs manage boundary; no plan rebuilding; no curator/ax imports — CLI contracts only, agents-management consumed as Go module);
   - CLI surface: `curator-run <env-id> [--profile P] [--system-prompt] [--model/--effort passthrough to spawn plane as declared] [--] <native args...>` — define exact flag set, precedence, and pass-through rules; keep flags minimal and justified;
   - composition algorithm: obtain agents-management launch plan (value, never rebuilt) -> obtain fragment via `curator env resolve <env> --profile P --format json` (subprocess; failure modes) -> merge env (fragment wins on its closed names; document the conflict rule) -> ax handoff: when the machine's ax integration is configured the launcher ALWAYS routes through ax instrumentation (configuration change, not a flag, to bypass) with the fragment recorded per environments.md §10 recommendations; without ax: direct exec, argv verbatim after `--`;
   - system-prompt application: engage only on explicit opt-in; per-channel behavior (flag-class claude_code/pi, config-key codex_cli, variable-class gemini when it lands); the mandatory warning text requirements (customized system prompt; replacement discards built-ins; cache/billing consequence) — reference environments.md rather than restating channel tables;
   - errors and diagnostics: stable code families for resolve failure, plan refusal (provider limits), missing provider binary, unsupported env, ax handoff failure;
   - versioning: spec version 0.1.0-draft; promotion note (sibling -spec repo at stabilization per Decision 0010).
3. Go skeleton that COMPILES: `go.mod` (module `github.com/relux-works/curator-agent-launcher`, go 1.23), `cmd/curator-run/main.go` — stub printing name/spec version/usage and exiting 2 on unknown args (no real logic; do NOT import agents-management yet — record it in SPEC as the planned dependency), `go build ./...` and `go vet ./...` green.
4. `LICENSE` (Apache-2.0, byte-copy from `~/Developer/ReluxWorks/curator/LICENSE`) + `NOTICE` (mirror curator's shape, name this project), `.gitignore` (Go + .temp/), `Makefile` (build/vet/test targets).
5. English throughout; house prose style of the curator-spec ecosystem for SPEC.md.

## Board
Attach resource `launcher-bootstrap-notes.md` (file map, spec decisions taken, open items for reviewer); handoff task to `to-review`.

## Do not
Push anywhere; add remotes; tag; touch other repos/worktrees; import agents-management in code; mark done.
