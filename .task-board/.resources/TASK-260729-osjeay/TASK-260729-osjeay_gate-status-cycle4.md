# TASK-260729-osjeay — gate status, rework cycle 4 (revision 5)

**Date:** 2026-07-29
**Scope:** read-only execution-map revision. No product, spec, CI, `Makefile`, pin, or
`TASK-260720-1pvfj5` field was modified.

## Evidence-honesty ledger for this cycle

| Command | Standalone? | Real exit | Status |
|---|---|---|---|
| `sh .temp/TASK-260729-osjeay/verify-recipes.sh` | yes, no pipe | **0** | **green** — `ALL 21 EXPECTATIONS MET`, reproduced twice |
| `shasum -a 256 manifest.json` (rc.5 root) | yes | **0** | green — `b6f56aac…04c`, unchanged |
| rc.5 tree pipeline → inventory → `shasum -a 256` | yes | **0** | green — `e6a13215…2fae`, **448** files |
| `git status --short --untracked-files=all -- conformance/` | yes | **0** | green — **3 modified, 354 untracked** |
| `git show origin/main:.github/workflows/ci.yml \| grep -n 'ref:'` | grep is the consumer; the read is `git show` | **0** | green — `00b1688a…` at lines 28 and 81 |
| `task-board q 'get(…){id status}'` × 4 | yes | **0** | green — `1pvfj5` backlog, `jrrgw9` development, `2qqq0w` done, `1skseh` backlog |
| `gh api` × 5 (releases, release bodies, tag refs, announcements) | yes | **0** | green — see §2.3 and §10 rows 44–47, 49 |
| `curl -sS raw.githubusercontent.com/actions/runner-images/main/README.md` | yes, `-w '%{http_code}'` | **0** (http **200**) | green — §10 row 48 |

**No Go command of any kind was executed.** No `go`, `go test`, `go vet`, `go build`, `go list`,
`go version`, `gofmt`, `golangci-lint`. No install, no download, no dependency fetch, no host
mutation.

## Honesty correction carried in revision 5

Revisions 1–4 each asserted "no network fetch to GitHub". **This cycle made six read-only network
reads** (five `gh api`, one `curl`), disclosed in the map header and itemised as §10 rows 44–49.
They were required to source-verify F4, which cannot be answered from anything on disk. Nothing was
pulled, installed or downloaded as a dependency.

## Reviewer findings — disposition

| # | Finding | Status | Executable evidence |
|---|---|---|---|
| F1 | `require-toolchain` did not enforce the matching-root / no-auto-toolchain invariant | **fixed** | rev-4 recipe **accepts** cross-root `GOFMT` (case C, exit 0), unapproved `GOROOT` (D, 0), `GOTOOLCHAIN=auto` (E, 0). Corrected recipe **rejects** those plus five more (I–Q, exit 2) |
| F2 | Windows lane printed critical values; batch runner did not `exit /b %RC%` | **fixed in specification; not executable from here** | `ssh win` unreachable (exit 255). Every W1/W4 value now compared with `[ … ] \|\| exit 1`; runner ends `endlocal & exit /b %RC%` and persists the code; W8 reconciles three signals. Four producer-time negative injections W-N1…W-N4 are mandatory before any Windows exit 0 is trusted |
| F3 | source staging reintroduced an unchecked `tar\|tar` pipeline | **fixed** | rev-4 pipeline **exits 0 with 1 of 3 files** (case R). Corrected staging exits 1 on the same producer (S), on a missing file (T), on an extra file (U) |
| F4 | action-major and runner-label drift not dispositioned | **fixed, and the verdict's figures corrected** | §2.3 ledger. Current majors are **checkout v7** and **setup-go v7** (not v6); `golangci-lint-action v9` was right. New decision **D8**: `setup-go@v5` does **not** force `GOTOOLCHAIN=local` — v6.0.0 introduced that — so hosted jobs run `auto` today |

## Checklist reconciliation

- Item 13 `Tests green` — remains **unchecked**. The board owner reconciled it as not-applicable for
  this read-only, no-Go scope in cycle 3; the CLI still exposes only
  `add_checklist_item` / `check_item` / `uncheck_item`, so it cannot be removed or reworded by any
  agent. Unchanged this cycle.
- Item 15 (independent reviewer verification) — remains **unchecked**. Only the reviewer may check it.
- Item 16 (executable no-Go stub-harness validation) — **checked**, backed by real exit 0 and now by
  21 cases instead of 7.

## Artifact digests

| Artifact | SHA-256 |
|---|---|
| `.research/260729_final-curator-ci-execution-map.md` (revision 5, 2544 lines) | `1948f03811c54f59b2a5c1a1d32e01b43609a0cea8b0ffcb5ae6213400ff0d96` |
| `verify-recipes.sh` (21 cases) | `65a02fbee0bffe0f5dfefbe64f89f3537b5a32185c024ce0ed2a25e90c774e5a` |
| `verify-recipes-cycle4.log` | `020e916ce3a12eb9acc5da80f38d78465c9de03c605b8a8aa5d9fc1cc4cb8a62` |

## Preserved invariants (re-verified this cycle, all unchanged)

- rc.5 manifest `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`
- rc.5 tree `e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae` over **448** files
- conformance root **3 modified, 354 untracked**
- committed pin `00b1688a9b2457ca397a0bb550acf47cad8ee967`, unmoved, at `ci.yml:28` and `ci.yml:81`
- Linux native validation **non-gating** pending its named prerequisite
- every command in §5–§9 stated as a **future producer gate**; no green CI result claimed anywhere
