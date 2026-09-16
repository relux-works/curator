# Evidence — TASK-260916-dzbi8j: CLI aliases claude/codex normalized to claude_code/codex_cli

## What changed (2 files, +22/−0)

- `profiles/manager.md` §12.1 (Environment adapter registry): one alias-rule
  paragraph. The command-line environment operand (`env resolve <env-id>`,
  `profile use --env <env-id>`, `env unmanage --env <env-id>`) accepts
  `claude` and `codex` as aliases of the canonical `claude_code` and
  `codex_cli` registry ids; the manager normalizes an alias to the canonical
  id before any validation or lookup; every output, diagnostic, marker,
  fragment, configuration record, and lock carries the canonical id and an
  alias is never persisted; any other unknown spelling keeps the existing
  `environment_unknown` refusal. Final sentence is the launcher SPEC pointer:
  the launcher's `curator run <env-id>` follows the same rule under the
  launcher SPEC (curator-agent-launcher README), which owns that operand.
- `CHANGELOG.md`: Unreleased → Added entry describing the rule and stating
  no schema/vector/wire change.

No schema, vector, wire, or conformance change. `git diff --name-only`
lists only `CHANGELOG.md` and `profiles/manager.md`; frozen v1 bytes
(`schemas/v1/*`, `conformance/v1/*`, `protocol/*`) untouched.

## Why no launcher-SPEC edit in this repo

The launcher SPEC proper (launcher §3/§4.2/§5…) lives in the
curator-agent-launcher README, not in curator-spec; this repo holds only
decision records about it (0013 §6 "Launcher specification 0.2.0-draft",
0018 referencing "launcher SPEC §3"). Per dzbi8j-alias-brief.md, the
manager.md rule therefore carries a pointer and the launcher-side operand
is an implementation follow-up, not part of this leaf.

## Why no new tests

Task class is docs and the brief forbids schema/vector/conformance changes,
so there is no test surface to extend (a new case would itself be a
conformance change). Verification is the repo's own `make validate` gates,
run directly below.

## Validation — `make validate` constituent gates (repo venv `.temp/venv`, gitignored)

| Gate (exact Makefile line) | Exit |
|---|---|
| `python3 tools/validate.py` → `validated 60 schemas and 1047 vector files` | 0 |
| `python3 -B -m unittest discover -s tools -p 'test_*.py'` → `Ran 227 tests … OK` | 0 |
| `go test ./tools/...` → `ok …/tools/generate-vectors` | 0 |

Each gate ran as a standalone process; exit codes are real (`PIPESTATUS`
where tailed). Run split across sequential bounded shell calls per headless
constraints; the three commands are exactly the `make validate` recipe, so
`make validate` is green (exit 0).

## Worktree precondition (finding)

The Story worktree arrived dirty: 312 tracked modifications + 6 untracked
conformance artifacts, all from cancelled run RUN-260916-0c46ba's voided
wire rename (operator HOLD 11:10Z, retarget decision 11:40Z: "NO wire
rename; CLI aliases only"; sibling TASK-260916-11lwua is backlog and owns
no worktree changes). HEAD was the clean checkpoint 871d11b. I reverted all
tracked edits (`git checkout -- .`) and removed the 6 untracked artifacts,
restoring `git status` empty before applying this change; also recorded in
board notes. Control-root LOGBOOK writes are forbidden by campaign rules,
so this file + board notes are the record.

## Consumer follow-ups (not part of this leaf)

curator `envregistry`, `curator-run` launcher alias handling under its
SPEC, relux-root-context `validate.sh` — implementation tasks.
