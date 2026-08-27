# TASK-260720-jrrgw9 — Windows verifier 5 results

Date: 2026-07-29  
Role: tester  
Candidate: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-jrrgw9/worktree`

## Result

The accepted verifier-4 macOS full and race gates remain green and were not
rerun. Native Windows qualification was retried exactly three times through
`ssh win`; every BatchMode attempt timed out connecting to
`100.120.84.42:22` and returned exit 255.

Windows therefore remains externally unqualified. It is not emulated and is
not claimed as passing. This external availability result does not block or
reinterpret the accepted verifier-4 macOS gates.

## SSH retry ledger

Each retry used:

`ssh -T -o BatchMode=yes -o ConnectTimeout=15 -o ConnectionAttempts=1 win 'exit 0'`

| Attempt | UTC start | UTC end | Real exit | Result |
| --- | --- | --- | ---: | --- |
| 1 | 2026-07-29T17:24:07Z | 2026-07-29T17:24:22Z | 255 | `ssh: connect to host 100.120.84.42 port 22: Operation timed out` |
| 2 | 2026-07-29T17:24:30Z | 2026-07-29T17:24:45Z | 255 | `ssh: connect to host 100.120.84.42 port 22: Operation timed out` |
| 3 | 2026-07-29T17:24:55Z | 2026-07-29T17:25:10Z | 255 | `ssh: connect to host 100.120.84.42 port 22: Operation timed out` |

There was a two-second pause before attempts 2 and 3.

## Scope and safety

- No local Go command was run.
- The macOS candidate was not edited, staged, committed, stashed, checked out,
  or published.
- Because no SSH connection succeeded, no Windows inventory command ran, no
  remote directory was created, no candidate archive was transferred, and no
  remote mutation or Windows Go test occurred.
- No local transfer archive was created; the verified archive count is zero.
- No remote cleanup was required because execution never reached remote setup.
- Local evidence consists only of this outcome and the three task-owned raw SSH
  logs.

## Candidate identity check

Read-only local hashes still match the accepted verifier-4 snapshot:

- `internal/transaction/namespace.go`:
  `bb332038c5bf41a4043f6c3f799ea3ab530b9beeac9b5688fed8d1ad0edc56be`
- `internal/transaction/namespace_pass_test.go`:
  `3611f04f296b6ed3b48efdeac969a6a79f262cbeba83a7786d5cf20ce94d4d63`
- `internal/install/atomicity/fixture_test.go`:
  `e0732e2e3df9adee95321ba28723a878699722747cb231a5309902a56f1f6120`

The candidate staged-path count is zero.

## Prior accepted executable evidence

Verifier 4 remains authoritative for macOS execution:

- `CURATOR_CONFORMANCE_ROOT=... go test -count=1 ./...`: exit 0.
- `CURATOR_CONFORMANCE_ROOT=... go test -count=1 -race ./...`: exit 0.
- Race `internal/install/atomicity`: 115.687 seconds, within the 480-second bar.

## Handoff

Return to review with Windows externally unqualified due to three connection
timeouts. No candidate rework is indicated by this retry.

## Board evidence note

The task-scoped outcome and combined raw ledger were attached successfully.
Three additional per-attempt raw attachments were also uploaded. A dry-run to
remove those redundant aliases succeeded, but the confirmed board deletion
returned exit 1 because resource-reference validation encountered an unrelated
legacy `todo` status value. No resource was deleted; all evidence remains
recoverable and attached. This board metadata anomaly does not affect the
candidate or verifier result.
