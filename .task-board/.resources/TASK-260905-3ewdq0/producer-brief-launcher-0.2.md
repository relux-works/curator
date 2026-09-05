# Producer brief: launcher SPEC `0.2.0-draft`

## Where and what

- Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-agent-launcher-spec-0.2`,
  branch `draft/spec-0.2`, base `6de42d8` (= main).
- Edit `SPEC.md` (the contract) and the two places that restate the version:
  `cmd/curator-run/main.go` `specVersion` (`0.1.0-draft` today — the stub reports the
  specification version; set it to `0.2.0-draft` and fix `main_test.go` if it pins the
  string) and `README.md` line ~24. `make check` must pass (record the tail). Nothing else.
- Authority: curator-spec Decision 0013 (`decisions/0013-execution-ownership-and-launch-plans.md`
  at `83de1a5`, main of `~/Developer/ReluxWorks/curator-spec`): Decisions 1, 2, 5,
  and above all **6.1–6.5**, plus 3.2/3.6/4 for the document the launcher composes and 7
  for the extension keys. Also Decision 0012 Decision 6 (MCP channel descriptors and the
  `env_names` union) and environments.md §10 (fragment; under 0012 the fragment's `profile`
  carries `lock_sha256`, `composition` is withdrawn, `precedence` carries two primitives,
  and an `mcp` section exists — the normative rewrite is pending as environments revision
  1.1, so cite Decision 0012 D8 for those members and say the reference is to the 1.1
  rewrite). Read the current SPEC.md fully first; keep its voice and density.
- Deliverable: one signed commit (`git commit -S`; paste the `git log --show-signature -1`
  line). Do not push, tag, or open a PR. Never write LOGBOOK.md or anything into the
  control root.

## What 0.2 must say (map each to a SPEC section; keep §5 and §6 unless named)

1. **§4 ordering (M7, D6.1)**: renumber so the fragment (old §4.3) is obtained first, then
   the environment-to-system mapping, then the plan; `LaunchRequest.Home` = the fragment's
   home-variable value for the environment; `WorkDir` = cwd. State why (limit state keyed
   by (provider, home)).
2. **§4.1 plan (M2, D5)**: the launcher requests `LaunchModeInteractive` by name; it never
   spells a provider flag; the plan's argv carries only model selection and effort
   transport; the request carries an empty `Composition` (the MCP channel is applied by the
   launcher from the fragment, not by the module). Update §1 non-goals accordingly ("no
   plan rebuilding" now reads: appending channel flags and native args to a plan value is
   composition, not rebuilding).
3. **§3 + new §4.x defaults (M8, D6.2)**: precedence `--model`/`--effort` flags → closed
   launcher machine-config mapping env-id → {model, effort} (name the file location and
   its closed schema: keys are env-ids of §4.2, values `{model, effort}`; lockable) →
   lineup fallback (`vendorplugin.Lineup` top admitted model for the mapped system,
   `Effort.Recommended`, no effort when `EffortSupportNone`); per-member resolution; the
   resolved pair and level printed on stderr every launch; refusal completion with
   `--effort` unchanged.
4. **§4.4 composition rule (D6.3)**: argv = plan argv ++ system-prompt channel flags
   (opt-in, §5) ++ MCP channel flags (fragment `mcp` descriptor: claude `--mcp-config <path>
   --strict-mcp-config`; codex `-p curator-mcp`; opencode none/variable) ++ native args;
   environment = inherited ⊕ plan Env ⊕ fragment env ⊕ engaged variable-kind channel;
   env_names = fragment `mcp.env_names`; stdin = plan stdin under D4 mapping; the
   literal-vs-lookup collision rule (a composed `env_literals` name also present in
   `mcp.env_names` is dropped from `env_names` with a warning — D6.3 as amended by review
   F5). Order-is-contract sentence plus "per-tool boundary verified against the pinned
   release before vectors freeze" in §9.
5. **§4.5 tracked mode (D6.4)**: the exact `ax start <name> --provider <id> --launch-plan -
   [--profile <ax-profile>] --workspace <cwd>` invocation with the document on ax's stdin;
   the document members (`argv_suffix` = composed argv minus element 0; `env_names`;
   `env_literals` = composer's own names only; `stdin`; `extensions` = the four
   `works.relux.curator.*` keys with their exact derivations, `profile-pin` =
   `sha256:` lock hash, `fragment-digest` over CCJ-1 of the parsed fragment,
   `system-modules` = presence of `system_prompt`); provider-id column added to the §4.2
   table (`claude_code`→`claude`, `codex_cli`→`codex`, `pi`→`pi`, opencode unsupported);
   a new launcher flag for the ax execution profile (name it, e.g. `--ax-profile
   standard|yolo`, default absent = ax default) — the only way `--profile yolo` reaches ax;
   session-name default `<env-id>-<utc-stamp>` (`YYYYMMDDTHHMMSSZ`), `--name` override,
   ax §2.1 grammar, >64 chars = `usage`; `ax_handoff_failed` terminal with ax's Structured
   Error passed through; untracked mode execs the composed plan. Add the new flags to the
   §3 table and parsing rules.
6. **§6 diagnostics**: add any new code the above needs (e.g. `defaults_unresolvable` when
   no level yields a model); keep the two invariants.
7. **§7 planned dependency**: the module version that carries `LaunchModeInteractive` is
   required; say the exact requirement is pinned when that release exists.
8. **§8 versioning + §8.1 row** for `0.2.0-draft`; **§9 open items**: replace the
   "ax handoff invocation shape is specified when the PR lands" item with the D6.4 shape
   and note the PR #1 revision is open; keep the per-tool boundary verification item;
   state that the behavioral contract is implementable only once the ax operation and the
   agents-management mode land (Decision 0013 Consequences).

## Board

Task `TASK-260905-3ewdq0` (curator board; `TASK_BOARD_DIR` set). Attach `TASK-260905-3ewdq0_drafting-report.md`
(commit + signature line, a decision-item → SPEC-section table, `make check` tail, anything
not verified labeled docs-confidence) as an outcome resource, then
`task-board handoff TASK-260905-3ewdq0 --role developer`.
