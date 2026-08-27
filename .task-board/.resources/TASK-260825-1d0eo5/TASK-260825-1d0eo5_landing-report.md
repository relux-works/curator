# TASK-260825-1d0eo5 — landing report

Composite for `EPIC-260825-2p2kl3` (scoped HTTPS credentials for external build
repositories).

- Branch: `task/TASK-260825-1d0eo5-https-credentials-composite`
- Cut from: `origin/main` `e027667` (Merge pull request #42)
- Pull request: https://github.com/relux-works/curator/pull/43
- Worktree: `.temp/STORY-260825-39h6vz/worktree/.temp/TASK-260825-1d0eo5/composite`

## How the composite was assembled

The epic produced **no commit anywhere**. Every accepted task was reviewed
against uncommitted files, and those files were split across three working
trees that disagreed with each other. Authoritative state had to be
reconstructed from file mtimes:

| Tree | Contents | mtimes | Used |
| --- | --- | --- | --- |
| primary checkout `/Users/iv/Developer/ReluxWorks/curator` | source, older; **no** `buildhttpsprompt.go`, **no** SSH-prompt persistence fix; plus unrelated `.github/ci/*.sh` edits | 02:28–02:53 (sh: 2026-08-23) | logbook entries only |
| `.temp/STORY-260825-32bopo/worktree` | complete source superset | 03:37–04:23 | **source of record** |
| `.temp/STORY-260825-37cz7x/worktree` | clean at base commit | — | nothing |
| `.temp/STORY-260825-39h6vz/worktree` | `docs/build-https.md`, `CHANGELOG.md`, `README.md`, `LOGBOOK.md` | 05:16–06:22 | **docs of record** |

Assembling from the primary checkout — the tree three task notes point at —
would have shipped a documentation page describing a candidate prompt whose
implementation was not in that tree, and would have dropped the fix that stops
the SSH prompt persisting an answer the operator did not ask to save.

Source was carried onto `origin/main` as a three-way **patch**, not a file copy,
because `origin/main` had moved on and independently touched
`internal/config/config.go`, `CHANGELOG.md`, `README.md` and `LOGBOOK.md`.
All twelve tracked files applied cleanly. Two reconciliations were made by hand:

- **CHANGELOG.md** conflicted in `Unreleased / Added`. Both sides kept; a
  feature-level entry was added above the docs-only entry the docs task wrote.
- **LOGBOOK.md** — `origin/main` carries a 47-line file with no shared history
  with the 3000-line file the epic wrote into (see logbook `0627`). The epic's
  nine entries were inserted into the `origin/main` file under `2026-08-25` in
  descending order. Entries `0052` and `0057` are numbered as a sequence rather
  than the file's `HHMM` convention; preserved verbatim rather than renumbered.

**Scrubbed before commit:** logbook entry `0130` named the sibling
implementation and an absolute local path. The repository's `ci.yml` naming gate
allows that name on exactly one README line and nowhere else, so the branch
would have gone red. The note was rewritten to reference the Curator Protocol
specification and this repository only, per the epic's policy.

Deliberately **not** carried: the primary checkout's `.github/ci/gate-selftest.sh`
and `.github/ci/test-gate.sh` edits, which are dated 2026-08-23 and belong to
different work.

## Commits

Nine signed commits, in dependency order. Each was independently checked out in
a separate worktree and proved to `go build ./...` **and** `go vet ./...` at
exit 0, so the branch bisects.

| Commit | Subject |
| --- | --- |
| `1d34f71` | Add the build_https configuration field for operator HTTPS credentials |
| `3437510` | Read and store operator HTTPS credentials through git credential only |
| `02195ab` | Answer a private HTTPS fetch through a host-pinned askpass broker |
| `9bb13de` | Ask before persisting an SSH credential answered at the prompt |
| `02aa23e` | Resolve an HTTPS credential per repository before the closure is fetched |
| `681f3e9` | Offer HTTPS credential candidates before the first fetch |
| `2543715` | Add curator config build-https for managing token-source selections |
| `25339e1` | Document scoped HTTPS build-repository credentials |
| `6f1040f` | Record how this epic was delivered, and what nearly shipped because of it |

An earlier ordering placed the HTTPS resolution commit before the SSH-prompt
commit; commit 4 of that ordering failed `go build` on an undefined
`credentialScopeCovers`. The history was rebuilt rather than the failure noted,
and the rebuilt sequence was re-verified commit by commit.

## Gates run locally on this branch

Every command was run as a standalone process; the exit code below is its real
exit code.

| Gate | Command | Exit | Result |
| --- | --- | ---: | --- |
| gofmt | `gofmt -l cmd internal` | 0 | no files listed |
| build | `go build ./...` | 0 | — |
| vet | `go vet ./...` | 0 | — |
| lint | `golangci-lint run` | 0 | 0 issues |
| gate self-test | `bash .github/ci/gate-selftest.sh` | 0 | 81 passed, 0 failed |
| broad suppression | `bash .github/ci/no-broad-suppression.sh` | 0 | ok |
| ledger | `bash .github/ci/ledger-consistency.sh` | 0 | 80 rows across linux darwin windows |
| naming gate | the `ci.yml` grep, run verbatim | 0 | clean outside README, exactly 1 README line |
| full test gate | `bash .github/ci/test-gate.sh` with `CURATOR_CONFORMANCE_ROOT` at `SPEC_PIN` `0ed5c691` | 0 | plan served/deferred/excluded = 44/0/0; `go test` exit 0; platform-case gate exit 0, 11 skips recorded |

The test gate ran at `25339e1`. The only later commit, `6f1040f`, changes
`LOGBOOK.md` alone; nothing under `cmd/`, `internal/` or `.github/` reads that
file, verified by grep. gofmt and the naming gate were re-run at `6f1040f` and
are clean.

Per-commit `go build` / `go vet` verification used a scratch worktree at
`.temp/TASK-260825-1d0eo5/verify`; the conformance root was materialised at
`.temp/TASK-260825-1d0eo5/spec-pin` from `curator-spec` `0ed5c691`.

## What is not claimed

- The Windows and Linux lanes were not run locally — this host is macOS. They
  are covered by CI on the pull request.
- `make race` was not run locally. CI runs the race lane.
- The epic's per-task acceptance evidence is carried from the reviewer verdicts
  already on the board; this task re-ran the gate set over the assembled branch
  rather than re-reviewing each task.
