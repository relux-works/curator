# Launcher bootstrap notes — TASK-260901-32j97g

Bootstrap of `~/Developer/ReluxWorks/curator-agent-launcher`: compiling Go
skeleton plus the in-repo SPEC.md draft (`0.1.0-draft`). Local only, on
`main`, two signed commits, no remote configured, nothing pushed or
tagged — the orchestrator lands after review.

## File map

| Path | What it is |
|---|---|
| `SPEC.md` | The specification draft — the substance of this task. Scope/non-goals, CLI surface, five-step composition algorithm, ax handoff, system-prompt opt-in + warnings, diagnostics, versioning, open items. |
| `README.md` | Execution-plane summary, status (spec draft, not implemented), `curator-<name>` PATH-discovery install note. |
| `cmd/curator-run/main.go` | Stub: `--help`/`-h`/`--version` print name + spec version and exit 0; every other argv shape refuses with usage on stderr and exit 2. No composition logic; agents-management deliberately not imported (recorded in SPEC §7 as the planned dependency). |
| `cmd/curator-run/main_test.go` | Tests of `run()` (the production dispatch site — `main` exits with its value). Negative cases drive the exit-2 gate with argv shapes that must be refused (empty argv, a real env-id, flag+trailing args, spec-declared-but-unimplemented flags, `--`-prefixed native args); they fail if the gate admits any of them. |
| `go.mod` | `github.com/relux-works/curator-agent-launcher`, go 1.23, zero dependencies. |
| `Makefile` | `build` / `vet` / `test` / `check` / `clean`. |
| `LICENSE` | Apache-2.0, byte-copy of curator's (sha1 2b8b8152… identical). |
| `NOTICE` | Curator's shape, named for this project. |
| `.gitignore` | Go artifacts, `go.work*`, `.temp/`, editor/OS noise. |

## Commits (both signed, verified `G`)

- `d3400fb` Bootstrap the launcher repository skeleton
- `dae0c35` Add the launcher specification draft, 0.1.0-draft

## Validation (all run in the repo root, real exit codes)

- `go build ./...` — exit 0
- `go vet ./...` — exit 0
- `go test ./... -count=1` — exit 0 (`ok cmd/curator-run`)
- `gofmt -l .` — no output, exit 0
- Built-binary smoke: `--version` → `curator-run 0.1.0-draft`, exit 0;
  `curator-run bogus` → exit 2. (Note: `go run` masks the child's exit
  code as 1; the smoke used a compiled binary.)

## Spec decisions taken (reviewer attention here)

1. **`--system-prompt` takes a required value `append|replace`** — the
   opt-in must state its semantics; the launcher never defaults to
   `replace` (the destructive channel). Decision 0010 names only the
   replace-class channels, but §7.3 declares append channels for
   claude_code/pi and defaulting either way seemed worse than requiring
   the choice.
2. **Opt-in with no matching channel is an error** (`sysprompt_channel_unavailable`),
   not a silent no-op — an operator who asked for a customized run must
   not get an uncustomized one without noticing.
3. **Closed env-id → agentic-system mapping owned by the launcher**
   (§4.2): claude_code→claude-code, codex_cli→codex, pi→pi;
   `opencode` has no agents-management system plugin in revision 1 →
   `env_unsupported`. Neither spec text mandated who owns this mapping;
   the launcher is the only component that speaks both vocabularies.
4. **Env merge rule** (§4.4): process env < plan env < fragment env,
   fragment winning on exactly its closed adapter-registry names; SHOULD
   warn on an actual plan/fragment overlap. Justified by the §10.3
   profile-influence boundary (fragment can only re-aim the home).
5. **No untracked fallback on ax failure** (`ax_handoff_failed` is
   terminal): a machine configured for tracking has declared untracked
   sessions the failure mode, not the fallback. Same shape for resolve
   failures — a malformed fragment read is a read failure, never treated
   as absence (`resolve_fragment_invalid`, no degraded exec).
6. **Exit-code discipline**: 2 for usage, 1 for every operational
   failure, one stable diagnostic code line on stderr; the code, not the
   prose, is the contract.
7. **Native args MUST follow `--`**; a second bare operand before `--` is
   a usage error, and unknown launcher-side flags are never forwarded.

## Open items for reviewer

- SPEC §9 lists the intentional gaps: opencode spawn mapping, gemini
  variable-class channel, exact codex_cli config-override spelling and
  flag spellings (pinned-release verification before conformance vectors
  freeze), and the ax handoff argv contract (awaits the Decision 10 PR
  against agent-session-manager-spec).
- Stub's `--version` reports spec version only; build version reporting
  is specified (§8) but deferred to the implementation.
- go.mod pins `go 1.23` per the brief; toolchain on this machine is
  1.25.5 — no toolchain directive added.
