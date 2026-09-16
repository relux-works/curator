# TASK-260910-5nrmtt results — rework 2 (rev3)

Producer: developer. Shell for all commands below: bash. Exit codes are
each gate's real status; slow gates streamed to files (this host stalls
new test-binary startup for minutes and `| tail` hides the stream, so
piped invocations below were display-only — every gate exit comes from
an unpiped run or an `echo "exit=$?"` trailer).

## Revision

- Rework brief: `5nrmtt-rework-2.md` (P1 positive diagnostic grammar;
  P2 Windows Job Object or typed refusal before process creation).
  Verdict rev2 + reproducers
  (`TASK-260910-5nrmtt_review-reproducers-rev2.patch`) applied VERBATIM
  as `TestReviewRev2MixedUnknownMustNotFallback` (not weakened).
- Spec: curator-spec checkout `07e2b41`
  (`protocol/repository-transport.md` rev 1 §§2–3; revision 2 not in
  scope); worktree `task-board/story/STORY-260910-1bhj0g` at `38e62cb`.
- Tree: `git status --short` shows exactly the 12 rev2 paths (4
  modified: `internal/buildrepo/admission.go`,
  `internal/buildrepo/httpsbroker_test.go`,
  `internal/gitcred/gitcred.go`, `internal/gitcred/provider_test.go`;
  8 new: `docs/draft-transport-resolution.md`,
  `internal/buildrepo/{process_unix,process_windows,sshbroker,sshbroker_test,transport,transport_test}.go`,
  `internal/gitcred/provider.go`). No other dirty files; sibling policy
  checkpoint and parser/collections untouched; `internal/gitops`,
  `internal/buildsource` untouched; frozen v1 schemas untouched.

## P1 — closed line grammar (`internal/buildrepo/transport.go`)

Replaced the substring classifier with a line parser. Every non-empty
stderr line must parse as a complete git/ssh diagnostic: the
`fatal: unable to access '<url>':` wrapper (reason matched WITHOUT the
URL, so endpoint text can never read as a signal), whole-line anchored
ssh shapes (`ssh: could not resolve hostname`, `ssh: connect to host`
with an enumerated reason set, daemon `Permission denied (methods)`),
forge `remote:` shapes, or bare `fatal:`/`error:` sentences. Reasons
and token shapes are fail-closed-first (TLS, host-key, identity, ref,
404, integrity, audit before auth/availability). Ignored (not
evidence): non-diagnostic `facility:` lines (wrapper chatter, echoed
environment, ssh `warning:`, forge `hint:`) and git's fixed ssh/rpc
framing (`Could not read from remote repository` + access-rights hint,
`remote end hung up`), with the framing boundary restored when a helper
glues it onto its final line without a newline. Any other line —
unparsed diagnostic, local-filesystem failure quoting a signal as a
filename (`'connection timed out'`), bare numbers, forge timeout words
without endpoint structure — poisons the whole output to unknown.
Fallback needs ≥1 positively identified availability/auth line, zero
fail-closed lines, zero unparsed lines. New `could not read
username|password for '<url>'` auth shape (headless auth-unavailable)
and bare `could not|couldn't resolve host|hostname` fragment shape (the
committed truncation test requires the torn flood prefix to still read
availability — the truncation FLAG closes that plan, never content).

## P2 — Windows refusal (option b), with reason

Chose rework option (b): `AcquireNetworkResolved` returns typed
`transport_resolution_unsupported_platform` AFTER plan validation (a
malformed policy reports identically everywhere) and BEFORE provider
binding, admission, or any fetch — no process created, no attempt
recorded. Reason: a correct Job Object implementation needs
CREATE_SUSPENDED spawn + post-start assignment + a Windows-native
descendant fixture, none of it verifiable from this macOS host; a
refusal is fail-closed, typed, and tested, while an unverifiable Job
Object risks a broken Windows gate. The legacy lane keeps its
rev2-accepted WaitDelay behavior (reviewer scoped P2 to the resolved
lane). `ClassifyAdmissionCode` maps the new code to unknown
(fail-closed by construction). `process_windows.go`/`admission.go`
unchanged. `TestResolvedTransportSharedDeadlineStopsSecondAttempt` and
`TestResolvedTransportNilProvidersSkipWithoutTraffic` (valid plans,
previously gate-free) now take `requireResolvedLane` (skip on
Windows); plan-validation tests need no gate (validation precedes the
refusal — verified on all platforms by `GOOS=windows go vet` plus the
unchanged `TestResolvedTransportRejectsPlanBeforeAnyGitCall`).

## Tests added (all reach the production entry unless noted)

- `TestReviewRev2MixedUnknownMustNotFallback` — reviewer rev2 probes
  verbatim (3/3: mixed DNS+protocol, object+timeout, quoted-signal
  filename): 1 fetch + refusal.
- `TestResolvedTransportUnknownTailMustNotFallbackWhenAlternateReady` —
  DNS + unparsed `remote:` progress tail: lane diagnostic, 1 fetch,
  unknown record.
- `TestResolvedTransportEnvironmentalNoiseStillFallsBack` — DNS +
  wrapper chatter + ssh `Warning:`: fallback succeeds, locked commit
  proved, availability-then-success records.
- `TestTransportResolutionPlatformGate` (unit, all hosts) +
  `TestResolvedTransportWindowsRefusesBeforeAnyProcess` (Windows-only:
  valid plan + unrunnable tool → platform code, zero records).
- `TestClassifyFetchOutput` +20 grammar cases (framing, glue split vs
  glued-unknown, CRLF, forge words, bare-503, hint/warning ignore,
  unreadable-object-is-unknown); admission-code map + platform code.

## Evidence (real exit codes)

- `go test -run TestClassifyFetchOutput` (unit): exit 0, 62 subtests
  + parent PASS, 0 FAIL.
- Targeted entry set (SSH per-attempt binding, rev2 probes,
  unknown-tail, environmental-noise, platform gate): exit 0, 8 PASS
  lines, 0 FAIL.
- Full `./internal/buildrepo` suite on the final tree (fresh binary):
  `PASS`, exit 0.
- `./internal/gitcred`: exit 0 (`ok`, 3.468 s).
- `go vet ./internal/buildrepo ./internal/gitcred`: exit 0.
- `GOOS=windows go build` + `GOOS=windows go vet ./internal/buildrepo`:
  exit 0 (Windows refusal test compiles; it runs in the hosted gate).
- `gofmt -l internal/buildrepo internal/gitcred`: no output.
- `git diff --check`: exit 0 (ran during edits; tree verified clean).
- Full module suite deliberately NOT run (host rule; gate at handoff).

## Narrowing mutants (bytes restored, `diff` empty, `grep -c MUTANT` = 0)

- M1a (P1): poison fallthrough `true→false` (unparsed lines ignored).
  `TestReviewRev2MixedUnknownMustNotFallback`: exit 1 — 2/3 probes fall
  back (2 fetches, nil err); quoted-filename probe stays closed
  (no-signal path). KILLED.
- M1b (P1): audit shape minus `audit` alternative.
  `TestClassifyFetchOutput/audit-beats-timeout` + entry
  `.../audit-timeout`: exit 1 (unknown ≠ audit; poisoning still blocks
  the fallback — defense in depth). KILLED at unit and entry level.
- M2 (P2): `transportResolutionPlatformError("windows")` → nil.
  `TestTransportResolutionPlatformGate`: exit 1 (nil ≠ platform code).
  KILLED.

## Findings / bounds

- Real-stderr finding: helpers that don't newline-terminate glue git's
  trailer onto the ssh diagnostic line (observed bytes in
  `TestResolvedTransportSSHAttemptBindsPerAttemptCredentials`); only the
  two fixed framing sentences are ever split out, any other glued
  content stays unparsed (locked by `glued-unknown-stays-unknown`).
- Host anomaly: new test binaries stall minutes at startup with zero
  output; piped `| tail` hides the stream entirely. All gates above ran
  via file-streamed output or direct prebuilt binaries
  (`go test -c -o /tmp/*.test`); `go vet`/compile stay instant (0.3 s).
- Bounds: Windows refusal itself is compile-verified here, runtime-proven
  only in the hosted gate (windows-only test). No acquisition caller
  wiring (separate task); resolver/SSH-manager dispatch unchanged.
  Checklist on the task is fully ticked and still accurate.
