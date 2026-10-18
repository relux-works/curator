# Producer brief: BUG-261018-dkgv1p — a help flag must never run the command

Repository: `relux-works/curator` (public), your Story worktree. Code and tests.

Field report: `curator init -h` creates `Skillfile.json` and `.gitignore` in the current directory. `-h` after a subcommand is ignored, and the subcommand runs. `curator -h`, `curator --version` and `curator help init` write nothing; `help init` prints only the top-level usage.

**Fix at the dispatch level, not per command:**
- Detect `-h`/`--help` for every registered subcommand before the command runs. Print that subcommand's dedicated help and exit 0. Do not touch cwd, HOME, `CURATOR_HOME` or TMPDIR.
- Make `help <subcommand>` print the same dedicated help.
- Keep `--` semantics: after `--`, a literal `-h` belongs to the passthrough arguments of commands that take them.

**Tests (Swift Testing does not apply; this is Go — use the repository's test style):**
- A table-driven test enumerates every registered subcommand. For each it runs `-h`, `--help` and `help <sub>` in an isolated temporary cwd and HOME, and asserts exit 0, help text naming the subcommand, and an identical before/after file tree (paths, sizes, modes).
- A positive control shows `init` without `-h` still creates its files.
- A mutant control: disabling the guard for one subcommand makes the table test fail for that subcommand. Record the command and exit code.
- A comment beside each test states the claim it proves.

**Validation:** the repository's full test suite and linters. No host names, people or local paths in code, tests or commit text.

**Handoff:** signed commit; attach `BUG-261018-dkgv1p_validation.md` (the commands, exit codes and the mutant control); `task-board handoff BUG-261018-dkgv1p --role developer`; end your turn.
