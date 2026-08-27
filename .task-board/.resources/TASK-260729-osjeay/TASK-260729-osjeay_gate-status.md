# Gate status — TASK-260729-osjeay rework cycle 2

Supersedes the cycle-1 gate status. Same evidence-honesty posture; the command list below is what
**this** cycle actually ran.

## Commands actually executed this cycle

Every one ran as a standalone foreground process. None was piped through `tee` or an unguarded pipe.

| # | Command | Exit | Result |
|---|---|---|---|
| 1 | `git rev-parse HEAD`; `git rev-parse origin/main` | 0 | `c06aa1a…` / `17804cea…` |
| 2 | `git show origin/main:.github/workflows/ci.yml \| grep -n 'ref:'` | 0 | `00b1688a9b2457ca397a0bb550acf47cad8ee967` at lines 28 and 81 |
| 3 | `shasum -a 256 manifest.json` in the rc.5 root | **0** | `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c` — matches the immutable value named in the task scope brief |
| 4 | `find . -type f -print0 \| LC_ALL=C sort -z \| xargs -0 shasum -a 256 \| shasum -a 256` | **0** | `e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae` |
| 5 | `find . -type f \| wc -l` | 0 | `448` |
| 6 | `git status --short --untracked-files=all -- conformance/ \| awk '{print $1}' \| sort \| uniq -c` | 0 | **354 `??` + 3 `M` = 357 lines** — confirms the reviewer's 3-modified / 354-untracked count |
| 7 | `tar -cf broken.tar -C "$(dirname "$DST")" conformance` in a scratch dir | **1** | `tar: conformance: Cannot stat: No such file or directory` — **expected-red: this reproduces reviewer finding F1**, the revision-2 archive command is not executable |
| 8 | `COPYFILE_DISABLE=1 tar --no-mac-metadata --no-xattrs --no-acls --no-fflags -cf full.tar -C "$(dirname "$(dirname "$DST")")" conformance` | 0 | corrected form archives cleanly |
| 9 | `tar -tf full.tar \| grep -x 'conformance/v1/manifest.json'` | 0 | listing assertion works; archive lists `conformance/`, `conformance/v1/`, `conformance/v1/manifest.json` |
| 10 | `find . -type l`; `find . ! -type f ! -type d ! -type l`; `find . -type d -empty`; `LC_ALL=C grep '[^ -~]'` and `grep -c '[[:space:]]'` over the path list | 0 | 0 symlinks, 0 other non-regular files, 68 dirs, 0 empty dirs, 0 non-ASCII path bytes, 0 paths with whitespace |
| 11 | `find … \| xargs -0 shasum -a 256 > $INV`; `wc -l < $INV`; `shasum -a 256 $INV` | 0 | 448 lines; the inventory file's own digest **is** `e6a13215…2fae` — the basis of the new cross-platform verification |
| 12 | `cd <rc5 root> && shasum -a 256 -c --status $INV` | **0** | per-file verification of all 448 files |
| 13 | `grep -rn 'func TestProbeRejects…' internal/godriver/*_test.go` | 0 | `internal/godriver/worker_test.go:434` |
| 14 | `cat go.mod`; `wc -l go.sum`; `ls -d vendor` | 0 / 1 for `ls -d vendor` | 4 direct + 17 indirect deps; `go.sum` 45 lines; **no `vendor/` dir** (the `ls` non-zero is the expected "no such file") |
| 15 | `cat .gitmodules`; `git submodule status` | 0 | submodule `agents/skills/skill-go-testing-tools` pinned `21585d0e…` (`v1.0.1-4-g21585d0`), populated |
| 16 | `find . -type l -not -path './.git/*' …` in the candidate worktree | 0 | exactly 2 symlinks: `.claude/skills/…`, `.codex/skills/…` |
| 17 | `git show origin/main:.github/workflows/ci.yml`; `sed -n '1,60p' Makefile` | 0 | current CI/Make inventories re-read in full |
| 18 | `ssh -o BatchMode=yes -o ConnectTimeout=10 win 'echo ok'` ×2 | **255** | `connect to host 100.120.84.42 port 22: Operation timed out` — **expected-red / measured blocker**: the Windows host is unreachable, so W-P1 is unsatisfied |
| 19 | `ssh -o BatchMode=yes -o ConnectTimeout=10 lev 'echo ok'` ×2 | **255** | same for `100.67.190.45` — Linux native host unreachable; stays non-gating |
| 20 | `ssh relux 'echo reachable; command -v go \|\| echo no-go; command -v tar'` | 0 | reachable; `command -v go` empty; `/usr/bin/tar` present |
| 21 | `ssh relux` probe of five conventional Go paths + `uname -sm` | 0 | **`FOUND /usr/local/bin/go`**; `/usr/local/go` and `/opt/homebrew/opt/go` absent; `Darwin x86_64` |
| 22 | `grep -n "2kaopg" LOGBOOK.md`; `sed -n '8,50p' LOGBOOK.md`; `tail -30 LOGBOOK.md` | 0 | surfaced the stale D6 premise — see below |

Rows 7–9 and 11–12 ran inside `.temp/TASK-260729-osjeay/tarcheck/` against scratch files and a
read-only copy of the candidate inventory. No product path was written.

## Commands NOT executed, and why

**No Go command of any kind ran this cycle** — no `go`, `go test`, `go vet`, `go build`, `go list`,
`go version`, `gofmt`, `golangci-lint`. No install, download, GitHub fetch, or host mutation. The
task scope brief forbids it: *"do not pull/install/download, and do not run Go or heavy tests while
verifier3 is active."*

Two consequences worth stating plainly:

- `/usr/local/bin/go` on `ssh relux` was proved to **exist and be executable**. Its **version is
  unmeasured**, because `go version` is a Go command. It is recorded as producer preflight R-P1, not
  as a fact.
- The PowerShell verification script in §5.4 W5 of the execution map has **never been run** — the
  Windows host is unreachable and this audit runs no PowerShell. The map requires the producer to
  prove it fails closed (delete a file, add a file, flip a byte → exit 1 each time) before trusting
  an exit 0.

Therefore every command in §5, §6, §7, §8 and §9 of the execution map is a **future producer gate**,
not evidence. The red predictions in D3, D6, §4.3 and §8.2 are static source analysis or cited prior
measurements by other tasks; each carries a named confirmation command in §9.1.
**No green CI result is claimed anywhere in this task output.**

## Reviewer findings F1–F4 — disposition

| Finding | Disposition | Evidence |
|---|---|---|
| F1 archive transport not executable | **fixed and empirically reproduced** | rows 7–9 above; execution map §5.2 C3, §10 rows 24–25 |
| F2 candidate identity contradictory and under-enforced | **fixed** — adoption language withdrawn; `candidate-digest` now checks manifest + tree + count + every file; Windows gets an equivalent whole-tree check sealed by the inventory digest | rows 11–12; map §5.2 C2, §5.4 W5, §7, I2 |
| F3 Make/CI parity claim false | **fixed** — `require-pin-root` makes the root mandatory; `check-ci` mirrors `test`, new `check-ci-linux` mirrors `test-linux`; target→job table relabelled exact / equivalent / intentionally different | row 17; map §7, §7.2, §9.2 |
| F4 Windows transport still an option set | **fixed** — one path W1→W9, base64 fallback deleted; the lane now fails closed on named prerequisites W-P1 (reachability, measured failing) and W-P2 (approved Go root) | rows 18, 21; map §5.4 |

Also corrected beyond the four findings:

- the conformance-root status count was **re-measured**, not copied (row 6);
- three host facts changed since revision 2 and were re-measured rather than inherited (rows 18–21);
- **a stale D6 premise was self-caught** (row 22). Revision 2 claimed the composite is red at Go's
  10-minute default without `-timeout 30m`. The most recent verifier-3 measurement records the exact
  non-race command exiting **0 in 444 s** (`cmd/curator` 384.270 s). The gate that *is* red at the
  default alarm is the **race** lane — `internal/install` **603.306 s**, `internal/install/atomicity`
  **603.701 s**, no `DATA RACE` — which is exactly the clause the AC names. D6, §6.3, §8 row 4,
  §9.1 step 4 and invariant I7 are corrected; the recipe is unchanged, but what the producer may
  claim is not. The race job is now labelled **specified but unproven**, with the 918 s / 1121 s
  projections marked as projections and the live `TASK-260720-jrrgw9` / `TASK-260729-3dr6hw`
  dependency named.

**Artifact identity.** `TASK-260729-osjeay_final-ci-execution-map.md` revision 3, SHA-256
`d93e155edf2a4ddf2b23b353f5c411bb40223a9c4be8295973a03bc2990c7d93`.

## Checklist item 13 — left unchecked, deliberately

`13 [ ] Tests green` is a generic checklist item inherited by this role. It is tied to a command
this task is explicitly scoped out of running. Under the evidence-honesty contract a command-tied
item may be checked only after that exact command has run green with exit code 0; it did not run, so
the item stays unchecked.

`task-board handoff TASK-260729-osjeay --role researcher` therefore fails closed:

```
cannot hand off TASK-260729-osjeay: unchecked checklist items [13] (Tests green): handoff evidence missing
```

The role end status `to-review` is applied with an explicit `set_status` instead. This is the same
conflict `TASK-260729-2sxx7k` recorded on 2026-07-29 (LOGBOOK 0510): a read-only or preflight-limited
task cannot satisfy a stored generic checklist that requires tests, and resolution needs a
board-owner checklist reconciliation that preserves the source-only, non-gating limitation.

**Requested board-owner action:** reconcile the generic checklist for read-only audit roles, or mark
item 13 not-applicable for this task. Do not resolve it by checking the item.

## Mutation scope

No product, spec, CI, `Makefile`, pin, or `TASK-260720-1pvfj5` field was modified. Board writes were
limited to this task: its own status, notes, and outcome resources. The superseded revision-2 draft
was moved from `.research/` to `.temp/TASK-260729-osjeay/rev2-superseded.md`; the board still holds
its own copy of every prior revision.
