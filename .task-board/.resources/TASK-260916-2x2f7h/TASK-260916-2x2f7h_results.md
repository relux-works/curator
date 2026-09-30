# TASK-260916-2x2f7h results

## Changes

- Amended the S5 protected-state remediation note in `protocol/environments.md` §10.1. It now explains that repair replays store bytes when a stale home needs repair, the persistence this permits for bytes admitted by the active pin, and the operator-owned store/lock authority bound accepted by S5 (§4). A store-only byte change that fails the pin still follows the existing rebuild-or-failure path.
- Added a one-row Claude/Codex asymmetry table in §7.8: Claude's strict channel excludes other MCP configuration when the manager supplies it; Codex's profile layer merges over its base, whose contents remain under the manager's seed policy.
- Added an Unreleased changelog entry. No MUST changed and no vector changed.

## Validation

| Command | Exit | Result |
|---|---:|---|
| `python3 tools/validate.py` with the initial interpreter | 1 | Could not import `jsonschema`; the dependency was absent. |
| `./.temp-validation-venv/bin/python tools/validate.py` after installing `requirements-dev.txt` in the worktree-local venv | 0 | Validated 64 schemas and 1,170 vector files. |
| `make validate` with the worktree-local venv on `PATH` | 130 | Its validator subcommand passed. The unittest phase was still running at the 10-minute shell-call limit and was interrupted in `DotfileManagersVectorTests.test_substituted_scenario_rejected_through_main`; the aggregate target did not reach its Go subcommand. |
| `go test ./tools/...` | 0 | Passed. |
| `make regenerate-check` | 0 | Generated vector corpus matched the checked-in corpus; no vector changes. |
| Relative-venv `python -B -m unittest test_validate.StoreBoundaryVectorTests test_validate.CodexSeedVectorTests` | 0 | 57 tests passed; Python emitted a `sys.prefix` warning because the executable path was relative. |
| Absolute-venv `python -B -m unittest test_validate.StoreBoundaryVectorTests test_validate.CodexSeedVectorTests` | 0 | 57 tests passed without warnings. |
| `git diff --check` | 0 | No whitespace errors. |

The full `make validate` unit-test suite remains incomplete because it exceeded the single-shell time bound. The direct protocol validator, relevant S5/E3 test classes, Go tooling tests, and vector regeneration check passed. The temporary venv was removed.

## Instructions and base evidence

- Read the attached `2x2f7h-brief.md` and the repository README validation guidance.
- The requested `remediation-spec-producer-rules.md`, task README, and Story README were not present in the worktree or exposed as task/Story board resources.
- The selected Story base is `4ad8042bde7f180319aaf15b363f44b4889c861c` on configured `main`; fresh advertisement, exact fetch, workspace `selected_base_oid`, and worktree tip all matched that OID.
- No `LOGBOOK.md` was created, per the task brief. Validation and instruction-availability findings are recorded here.
