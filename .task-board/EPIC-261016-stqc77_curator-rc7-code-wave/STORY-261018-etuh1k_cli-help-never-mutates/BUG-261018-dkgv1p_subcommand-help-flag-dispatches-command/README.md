# BUG-261018-dkgv1p: subcommand-help-flag-dispatches-command

## Description
Repro: empty git repo as cwd, isolated HOME/XDG_CONFIG_HOME/CURATOR_HOME/TMPDIR; `curator init -h` exits 0 and creates Skillfile.json and .gitignore. Controls: `curator --version`, `curator -h` and `curator help init` write nothing; `help init` prints only top-level usage. Expected: `-h`/`--help` after any subcommand prints that subcommand's help and exits 0 with zero filesystem writes; `help <subcommand>` prints the same dedicated help.

## Scope
CLI argument dispatch for all subcommands; help text for subcommands; tests.

## Acceptance Criteria
1) For every subcommand, `-h` and `--help` (any position before `--`) print dedicated help and exit 0 with no writes to cwd, HOME, CURATOR_HOME or TMPDIR (asserted by a before/after tree snapshot). 2) `help <subcommand>` prints the same dedicated help. 3) A table-driven test enumerates every registered subcommand so a new mutating command without the guard fails the test. 4) An argument after `--` that looks like `-h` is passed through unchanged where the command takes passthrough args. 5) Mutant control: removing the guard for one subcommand makes the table test fail for that subcommand. 6) Full test suite green; signed commit.
