# TASK-260720-1nvomm review verdict

## Verdict

Changes requested; route to `to-dev` for documentation rework and another reviewer cycle.

## Evidence checked

- Repository baseline is exact: `HEAD`, `origin/main`, and merge base are `57c1f56846d221ecc55786bd3c2467ec32f11730`.
- Accepted contract SHA-256 is exact: `6308d99d8bdad4445841bc1cbd230cadbed0020012d0e9d38d877b413348f681`.
- Intended implementation files are `protocol/core.md`, `SECURITY.md`, and new `decisions/0004-compile-only-build-drivers.md`. The untracked `.temp/TASK-260720-poa3ze/` tree is prerequisite scratch material, not part of this task's product diff.
- With `python3` resolved from a task-local virtual environment populated from pinned `requirements-dev.txt`, `python3 tools/validate.py` passed: 30 schemas and 93 vector files.
- `make validate` passed: the same validator, 8 Python tests, and `go test ./tools/...`.
- `git diff --check` passed.

## Required changes

1. Restore the accepted bootstrap isolation contract in `protocol/core.md` section 4.2. The accepted contract requires the three package-independent probes to run from a manager-owned empty CWD with private user/config/temp roots, `GOENV=off`, `GOTOOLCHAIN=local`, and no inherited `GOROOT` or target. Current lines 224-259 say only that bootstrap has no inherited `GOROOT` or target and defer `GOENV=off` and `GOTOOLCHAIN=local` until after the probes; they do not require the empty CWD. This weakens the closed command contract. State the bootstrap CWD and bootstrap environment explicitly, while retaining the post-probe addition of resolved `GOROOT`, target, and tuning.
2. Restore the accepted `go list` trust classification in `protocol/core.md` section 4.2. The accepted contract trusts a toolchain package only when the result has both `Standard` and `Goroot`; every other result and its module/source/embed inputs must remain below the build root. Current lines 261-270 refer only to “Standard packages,” omitting the `Goroot` conjunction. Require `Standard && Goroot` for the GOROOT branch and route every other result through the build-root/non-standard checks.

These are normative security/interop gaps, not test failures or an external blocker. No changes are requested to `SECURITY.md` or decision 0004 beyond any terminology alignment needed after the core corrections.
