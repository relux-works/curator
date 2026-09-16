# TASK-260916-3na4vf evidence — umbrella requires.mcp directory + coordinated 1.0.1

Worktree: `.temp/STORY-260916-jfyr3f/worktree` (branch `task-board/story/STORY-260916-jfyr3f`).
Shell for all commands below: `bash`. No pipes; exit codes are the commands' own.

## 1. Gate outputs (all green, real exit codes)

### `bash scripts/validate.sh` — exit 0
```
validate: manifests, modules, weights, ranges: OK
sources: 14 module digests OK
validate: module bytes: OK
validate: PASS
```

### `python3 -m unittest discover -s tests -v` — exit 0 (15 tests, incl. new `test_mcp_directory`)
```
test_duplicate_row ... ok
test_extra_mcp ... ok
test_mcp_directory ... ok
test_mcp_inventory ... ok
test_mcp_range ... ok
test_mcp_source ... ok
test_missing_mcp ... ok
test_missing_row ... ok
test_missing_sources ... ok
test_source_drift ... ok
test_sources_drift ... ok
test_trailing_lf ... ok
test_unknown_env ... ok
test_valid ... ok
test_weight_drift ... ok
Ran 15 tests in 9.310s — OK
```

### `python3 tests/mutants.py` — exit 0 (all 10 narrowing mutants killed, incl. new `mcp-directory`)
```
mcp-inventory: behavioral suite exit 1; expected failing probe test_mcp_inventory
mcp-source: behavioral suite exit 1; expected failing probe test_mcp_source
mcp-directory: behavioral suite exit 1; expected failing probe test_mcp_directory
trailing-lf: behavioral suite exit 1; expected failing probe test_trailing_lf
unknown-env: behavioral suite exit 1; expected failing probe test_unknown_env
weight: behavioral suite exit 1; expected failing probe test_weight_drift
digest: behavioral suite exit 1; expected failing probe test_source_drift
sources-digest: behavioral suite exit 1; expected failing probe test_sources_drift
inventory: behavioral suite exit 1; expected failing probe test_missing_row
duplicate: behavioral suite exit 1; expected failing probe test_duplicate_row
all narrowing mutants killed
```

### Syntax sanity (no linter is configured in this repo: no Makefile/CI/lint config)
- `python3 -m py_compile tests/test_validate.py tests/mutants.py scripts/check_sources.py` — exit 0
- `bash -n scripts/validate.sh` — exit 0

## 2. AC verification

- Umbrella `requires.mcp.figma.directory = packages/figma`,
  `requires.mcp.safari.directory = packages/safari`; `git`/`range` unchanged;
  key order alphabetical (`directory`, `git`, `range`), matching `requires.contexts` style.
- All six `packages/*/agent-context.json` report version `1.0.1` (parsed, not grepped).
- Only remaining `1.0.0` string in `packages/ README.md scripts/ tests/ SOURCES.sha256`:
  the umbrella README sentence naming the satisfying **relux-mcp** repository-wide tag
  `v1.0.0` (intentional, per brief).
- Umbrella `requires.contexts` ranges stay `^1.0`, which admits `1.0.1`; `SOURCES.sha256`
  untouched (no module bytes changed).
- `git status --short`: 16 modified files, no untracked files, no commits made by this run.
- Changed files: 6 manifests, 6 package READMEs, root README, `scripts/validate.sh`,
  `tests/mutants.py`, `tests/test_validate.py`.

## 3. Required deviation note (reviewer attention)

`scripts/validate.sh` asserted exact equality
`entry != {"git": ..., "range": "^1.0"}` for the umbrella `requires.mcp` entries, so it
could not exit 0 once `directory` was added. The assertion now compares against a
per-entry `expected_mcp` map including each package directory. `tests/mutants.py`'s
`mcp-source` mutant string was updated to the new gate line, and a `test_mcp_directory`
negative probe plus an `mcp-directory` narrowing mutant were added per standing orders
(gate behavior changed). This is the minimal change set that keeps `validate.sh` exit 0;
no other gate logic was touched.

## Revision 2 (rework-1: root README `directory` sentence)

Sole change vs revision 1: root `README.md` tag/version paragraph gained the sentence
"A context or MCP requirement selects a package within the repository through its
`directory` field (for example `packages/relux-root-context-core`, `packages/figma`)."
Everything else is byte-identical to revision 1 (`git diff --stat`: 16 files,
39 insertions, 19 deletions; no untracked files; no commits on the Story branch).
Shell: `bash`. Gates run standalone (stdout redirected to a file, not piped);
exit codes below are the gates' own.

### `bash scripts/validate.sh` — exit 0
```
validate: manifests, modules, weights, ranges: OK
sources: 14 module digests OK
validate: module bytes: OK
validate: PASS
```

### `python3 -m unittest discover -s tests -v` — exit 0 (15 tests, all ok)
```
test_duplicate_row ... ok
test_extra_mcp ... ok
test_mcp_directory ... ok
test_mcp_inventory ... ok
test_mcp_range ... ok
test_mcp_source ... ok
test_missing_mcp ... ok
test_missing_row ... ok
test_missing_sources ... ok
test_source_drift ... ok
test_sources_drift ... ok
test_trailing_lf ... ok
test_unknown_env ... ok
test_valid ... ok
test_weight_drift ... ok
Ran 15 tests in 11.577s — OK
```

### `python3 tests/mutants.py` — exit 0 (all 10 narrowing mutants killed)
```
mcp-inventory: behavioral suite exit 1; expected failing probe test_mcp_inventory
mcp-source: behavioral suite exit 1; expected failing probe test_mcp_source
mcp-directory: behavioral suite exit 1; expected failing probe test_mcp_directory
trailing-lf: behavioral suite exit 1; expected failing probe test_trailing_lf
unknown-env: behavioral suite exit 1; expected failing probe test_unknown_env
weight: behavioral suite exit 1; expected failing probe test_weight_drift
digest: behavioral suite exit 1; expected failing probe test_source_drift
sources-digest: behavioral suite exit 1; expected failing probe test_sources_drift
inventory: behavioral suite exit 1; expected failing probe test_missing_row
duplicate: behavioral suite exit 1; expected failing probe test_duplicate_row
all narrowing mutants killed
```

### Syntax sanity — exit 0 each
- `python3 -m py_compile tests/test_validate.py tests/mutants.py scripts/check_sources.py` — exit 0
- `bash -n scripts/validate.sh` — exit 0
