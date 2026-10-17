# Producer brief: TASK-261017-yyg01d, install and takeover preservation triage

Read the task description: five field reports (A-E) from another operator's isolated sandbox. This is research: reproduce, map to the spec, classify, and write fix leaves. No product change in this task.

**Safety first.** Every reproduction runs with a private scratch `HOME` (mode 0700) and a scratch git project under this task's `.temp/`, a locally built `curator` from your Story workspace, `curator bootstrap -non-interactive` inside that HOME, and no network or real credentials. Never touch the real home, the real `~/.agents`, `~/.claude`, `~/.codex` or any real project. Check `HOME` before every command.

**Reproduce on current main:**
- A: a schema-2 Skillfile with a `path` source `./agents` and collection `skills/*`; put unmanaged entries beside the managed skill in `.agents/skills/` (an unmanaged skill folder with `SKILL.md`, a folder with a notes file, a loose file) and an unmanaged link in a native skills directory pointing at one of them; run `curator install .`; record what disappears, what is printed and the exit code.
- B: a skill package with no manifest: `curator skill check` versus `curator install .`; then the control with a minimal `agent-skill.json`.
- C: whether any takeover or activation path can remove hand-authored files under managed output (for example `.agents/.instructions/`) or under native entrypoints.
- D: a first `profile install` without `--use`: does it activate the profile and touch native homes?
- E: `profile use` across several runtimes when one refuses (an unmanaged `CLAUDE.md` or `AGENTS.md`): what ends switched, what not, and is there a rollback.

**Map each to curator-spec** (read core, environments and skillfile on main): which clause governs ownership of managed output, unmanaged content, takeover, activation and multi-runtime switching; give a verdict per item: implementation bug, spec gap, or specified behaviour (then: do diagnostics and docs say it plainly?).

**Write a fix leaf per bug:** required behaviour (for A the safe options are refuse naming the unmanaged entries, preserve them, or an explicit backup; choose what the spec supports), negative and positive controls, touched packages, risk, and size. Give a priority order (A first if it reproduces).

**Output:** attach `TASK-261017-yyg01d_triage.md` (commands, exit codes, observations, spec mapping, verdicts, leaves, priority) and the sandbox logs. Public repository rules: no host names, people or local paths in board resources; describe the source as "a field report".

**Handoff:** no repository change is expected; say so explicitly. Run `task-board handoff TASK-261017-yyg01d --role developer`; end your turn.
