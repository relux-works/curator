# Canonical CR2 execution re-review

TASK-260908-3ued5d, Codex gpt-6-astra medium. CR2 ready treeff61be4a8bd43fa4ffb179d31aa38e41891d4313/base84747c3. Runtime exact-tree make check has passed at23:07:47Z; do not treat producer's prefinalization pending wording as a missing result.

Review F1 repair against original canonical PTY reproduction and current process.go: separate child foreground group, original terminal restoration, parent-only relay, Wait4 stop/reap handling and Cmd.Wait cleanup. Independently drive realPTY counting/read/stop/resume cases bothmodes; check exits and relevant failure cleanup. No debounce or background-only workaround. Inspect focused/mutant evidence (F1,F2,E6) and original48matrix; avoid redundant fullsuite/allmutants absent concrete concern. Record any remaining terminal bounds honestly; actualforeground default launch must work.

CR2 carries unchanged upstream §5code/tests/harness from adf627607eb334e9839288cfffce63e1268ae688; verify those3blob hashes and README integration. Final difference against currentmainadf must contain only execution/config deliverable. That upstream code is not newimplementation. Mainwiring and actual§5callback connection remain explicitly TASK-260908-1o7i8y obligations; retain requiredcallbackgate. No realmodel/nativePi/axbackend claim. All helperenvs synthetic; no ambientvalues persisted.

Attach task-prefixed verdict before canonical accept_cr2 or precise changes_requested. No source edits/commits, installs/restarts/CI/tags/releases, actualmodels/ax/homes, LOGBOOK/private/control-root writes. Reviewer owns currentmergedchecklist; inapplicable clauses get truthfulbounds, not a spurious blocker. Parent owns exact signed delivery.
