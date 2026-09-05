# Review findings: launcher SPEC `0.2.0-draft` (cycle 1)

Subject: `curator-agent-launcher` worktree `.worktrees/curator-agent-launcher-spec-0.2`,
branch `draft/spec-0.2`, head `ffe9b68`, base `6de42d8`. Authority: curator-spec Decision 0013
at `83de1a5`, Decision 0012 D6/D8, environments.md, ax SPEC `28bf96d`, skill-agents-management `91bf945`.

## Verdict: ACCEPT (no blocking or major findings)

## Dimension 1: Decision 0013 D6 items

| Item | SPEC section | Result |
|---|---|---|
| D6.1 fragment-first, `LaunchRequest.Home` = home variable value, `WorkDir` = cwd, (provider, home) rationale | §4 preamble, §4.1, §4.4 | exact |
| D5 `LaunchModeInteractive` by name, no provider flag, argv = model + effort transport, empty `Composition`, `ErrCompositionNotInteractive` | §4.4, §1 | exact |
| D6.2 precedence flags → machine config → `Lineup` + `Effort.Recommended` / none under `EffortSupportNone`, per member, stderr print every launch, `--effort` completion, no retry | §4.3, §3 | exact; file location/schema/lock are drafting choices the decision left open |
| D6.3 argv order (plan, sysprompt opt-in, MCP, native), MCP spellings taken from fragment descriptor, four env layers with SHOULD-warn, `env_names` = `mcp.env_names` minus `env_literals` with stderr warning (F5), stdin under D4 | §4.5 | exact; order-is-contract sentence present; disjoint-by-construction rationale carried |
| D6.4 `ax start <name> --provider <id> --launch-plan - [--profile <ax-profile>] --workspace <cwd>`, document on ax stdin, `argv_suffix` minus element 0, `env_literals` composer-only, four extension keys with derivations, provider column, ax-profile flag as the only route for `yolo`, session name default + `--name` + §2.1 grammar + >64 = `usage`, `ax_handoff_failed` terminal with Structured Error, untracked exec | §4.6, §4.2, §3 | exact |
| D6.5 non-goals restated | §1 | exact |
| §5, §6 stand unless named | §5 diff = two cross-reference renumbers (§4.5→§4.6); §6 gains `defaults` family, both invariants kept | ok |
| §7 module release requirement | §7 | ok; `91bf945` LaunchMode set verified as exec/dry-run/managed-session |
| §8/§8.1/§9 | present; ax shape item replaced, PR #1 open, implementability per Consequences, per-tool boundary kept, open questions 2/5 noted | ok |

Provider flags spelled by the launcher itself: none. `--mcp-config … --strict-mcp-config` and
`-p curator-mcp` appear only as the fragment descriptor's own content (Decision 0012 D6 wording
matches verbatim).

## Dimension 2: internal consistency
- §3 table ↔ parsing rules ↔ §6 usage row ↔ §9: consistent (`--name`, `--ax-profile`, locked default).
- README line 24 and `specVersion` say `0.2.0-draft`; usage text lists the new flags.
- `make check` rerun by the reviewer at `ffe9b68`: `go build`, `go vet`, `go test ./... -count=1` → `ok cmd/curator-run 0.276s`, exit 0. `gofmt -l cmd/` empty; `git diff --check` clean.
- Negative evidence for the one gate (`TestSpecVersionPinned`): reviewer mutated a scratch copy
  (`git archive ffe9b68`) to `0.1.0-draft` and to `0.2.0-drafT`; both runs FAIL. The gate rejects
  what it must reject, not only a delete-mutant.

## Dimension 3: facts
- `skill-agents-management` `91bf945`: `LaunchRequest.Home` (`pkg/agentic/system.go:386`, `plan.go:41`),
  `vendorplugin.Lineup(models []Model) []RankedModel` (`lineup.go:52`), `Recommended` declaration
  (`vendor.go:448`), `ErrEffortMissing`, `ErrUnsupportedLaunchMode`, `EffortSupportNone` — all present.
- ax SPEC `28bf96d`: §2.1 grammar `[A-Za-z0-9][A-Za-z0-9._-]{0,63}`, 1–64 (line 363); §7.1 built-in
  ids include `codex`, `claude`, `pi`; §14.1 and §15.1 exist. `--launch-plan` itself is PR #1's,
  correctly labelled open in §9.
- environments.md §7.1 home variables: `CLAUDE_CONFIG_DIR`, `CODEX_HOME`, `PI_CODING_AGENT_DIR`,
  `XDG_CONFIG_HOME=<parent>` for opencode — SPEC §4.1 matches.

## Dimension 4: commit
One signed commit `ffe9b68`, `Good "git" signature` ED25519, author Ivan Oparin <ivan@relux.works>.
Files: SPEC.md, README.md, cmd/curator-run/main.go, cmd/curator-run/main_test.go. Nothing else.
Not pushed; no tag; no PR.

## Minor findings (non-blocking; fold into the next revision or 0.2.1)

1. **minor / §4.3 level 2, lock sentence** — quote: "the machine entry is used even against
   `--model`/`--effort` flags for that member — the flags are then a `usage` error". These two
   clauses contradict: if the flag is a usage error the launch fails and nothing is "used".
   Fix: drop "is used even against … flags" and keep only "a `--model`/`--effort` flag for a
   locked member is a `usage` error naming the locked member".
2. **minor / §4.3 level 2** — "the operator file overrides the machine file per env-id" is
   per-entry, while the whole precedence is per-member. Say which: an operator entry `{model}`
   over a machine entry `{model, effort}` either keeps the machine effort (per member) or drops
   it (per entry). Recommend per member, consistent with the rest of §4.3.
3. **minor / §3 `--ax-profile`, §6 usage row** — `usage` on an untracked machine requires the
   launcher to read its `ax`-integration configuration during argument handling; §3 says usage
   errors "resolve nothing", which still holds, but the spec nowhere says how "integration
   configured" is determined (a pre-existing 0.1.x gap now load-bearing for a usage error).
   Suggest one sentence in §4.6 naming the source of that fact.
4. **note / §4.2** — "a row is launchable only when both non-Curator columns are filled" is a
   drafting addition beyond D6.4; defensible (tracked/untracked parity) and recorded here as such.
5. **note** — docs-confidence items in the drafting report (defaults.json design, `--ax-profile`
   untracked = usage, `--name` untracked = ignored) are accurately labelled as drafting choices.

## On the Change Request's empty repository delta
`CR-TASK-260905-3ewdq0-1` shows `repository_delta=empty` against the curator-spec story worktree.
That is the correct outcome: the producer brief scoped every edit to the sibling repository
`curator-agent-launcher` (branch `draft/spec-0.2`) and forbade writing anything into the
control root or curator-spec. The deliverable lives at `ffe9b68` in that repository and is
verified above; curator-spec was not meant to change.
