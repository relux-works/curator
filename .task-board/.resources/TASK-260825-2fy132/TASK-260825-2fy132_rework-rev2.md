# HTTPS credential documentation rework (revision 2)

## Changes

- Corrected `docs/build-https.md` to describe the shipped interactive
  prefetch candidate flow: terminal runs offer presence-only discovered host
  credentials or token entry, allow a persisted scope or a this-run-only
  choice, and stop when the operator aborts. Headless, non-terminal, and
  dry-run runs remain anonymous when no source covers a repository.
- Made the `list` transcript use real TAB separators.
- Kept the README overview's host-pinned HTTPS askpass-broker statement while
  linking both SSH and HTTPS operator pages.
- Added a correction to LOGBOOK entry 0057; the earlier false prompt claim is
  preserved as history and explicitly superseded.

## Verification

- `go run ./cmd/curator config build-https --help` in the delivery checkout:
  exit 0.
- Isolated delivery CLI `add` with `--token-env` and its `--keyring`
  replacement: both exit 0; output matches the documented transcripts.
- Delivery `list` inspected with `cat -vet`: exit 0 and emitted `^I` for a
  real TAB separator.
- `make lint` in the delivery checkout: exit 0 (`0 issues`).
- `lychee README.md CHANGELOG.md docs/build-https.md docs/build-ssh.md` in
  this worktree: exit 0 (25 links OK, 0 errors).
- `git diff --check` in this worktree: exit 0.
- `go test ./internal/config -run 'TestBuildHTTPS' -count=1` in the delivery
  checkout: exit 0 (`ok .../internal/config 0.354s`), covering HTTPS config
  grammar, strict parsing, token sources, and scope matching.

## Test note

The broad `go test ./cmd/curator ./internal/install` command was started in
the delivery checkout but was terminated by this run after about seven minutes
of active compilation and produced no result; it is not claimed as green. This
documentation-only rework changes no production or test code. The focused
HTTPS configuration test, command and link/lint checks above were rerun
directly for this rework; the board's accepted sibling implementation evidence
also covers the interactive resolver tests.
