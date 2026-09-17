# TASK-260910-5nrmtt results — rework 4 (rev6)

Producer: developer. Shell for all commands below: bash in the Story
worktree (`task-board/story/STORY-260910-1bhj0g`). Quoted exit codes are
real process results; `| tail` was display-only except where the exit
came from a prebuilt binary run or an `echo` trailer.

## Revision

- Rework brief: `5nrmtt-rework-4.md` (one P1 from verdict rev5
  `TASK-260910-5nrmtt_review-verdict-rev5.md`, reproducer
  `TASK-260910-5nrmtt_review-reproducers-rev5.patch`).
- Spec: curator-spec `protocol/repository-transport.md` rev 1 (rev 2
  not in scope); worktree base `38e62cb` (same as rev5).
- Tree: 4 modified (`internal/buildrepo/admission.go`,
  `internal/buildrepo/httpsbroker_test.go`,
  `internal/gitcred/gitcred.go`, `internal/gitcred/provider_test.go`)
  + 9 new (`docs/draft-transport-resolution.md`,
  `internal/buildrepo/{process_unix,process_windows,sshbroker,sshbroker_test,transport,transport_review_test,transport_test}.go`,
  `internal/gitcred/provider.go`). The extra new file vs rev5 is the
  split-out reviewer regression file required by this rework. Only
  `transport.go`, `transport_test.go`, `transport_review_test.go` and
  `docs/draft-transport-resolution.md` changed since rev5; closed
  table, Windows refusal, admission, brokers, process-graph files
  otherwise untouched.

## P1 — framingGlue repair deleted from classification

`ClassifyFetchOutput` (`internal/buildrepo/transport.go`) no longer
applies any replacement before splitting: it splits the captured
stderr only on real `\n` boundaries and requires every non-empty line
to match exactly one entry of the closed full-line table, with >=1
availability/auth diagnostic and zero fail-closed or unmatched lines.
A helper that glues `fatal: the remote end hung up unexpectedly` onto
its own final line without a newline now yields one original line
matching no table entry -> `FailureUnknown`, one fetch, lane refusal.
Verified: `grep -n "Replace\|Glue" transport.go` prints nothing (the
two `ReplaceAll` hits in the package are CRLF normalization in
`local.go`, outside the classifier); `grep -c '\.\*'`
`transport.go` = 0. The code comment on the split states the rule
("the classifier never invents boundaries") and the draft doc
(`docs/draft-transport-resolution.md:38-39`) records it.

## Tests

- `TestReviewRev5GluedLineRefuses`
  (`internal/buildrepo/transport_review_test.go`) — the rev5
  reproducer applied as the committed regression at the production
  entry `AcquireNetworkResolved` (glued `503`+trailer single line,
  ready SSH alternate): refuses with 1 fetch. PASS.
- Framing positive
  (`TestResolvedTransportClosedTableFramingStillFallsBack`) uses
  properly newline-terminated diagnostics (`RPC 503\n...hung up`):
  still falls back and proves the locked commit. PASS.
- All rev1-rev4 reviewer probes kept verbatim and green; closed-table
  unit/table tests unchanged.

## Evidence (real exit codes)

- `-run 'TestReviewRev5GluedLineRefuses'` on the final tree: PASS
  (2.08 s, prebuilt `/tmp/br260910.test` binary).
- `-run 'Review|ResolvedTransport|Classify'` (`./internal/buildrepo`,
  `-p 1 -count=1`): exit 0, `ok ... 82.670 s`.
- `./internal/gitcred` (`-p 1 -count=1`): exit 0, `ok ... 4.069 s`.
- `go vet ./internal/buildrepo/ ./internal/gitcred/`: exit 0.
- `gofmt -l internal/buildrepo internal/gitcred`: no output.
- `git diff --check`: exit 0.
- `GOOS=windows go build ./internal/buildrepo/`: exit 0.
- Full module suite deliberately NOT run (host rule; gate at handoff).

## Narrowing mutant (bytes restored, sha256-identical)

- M (P1 glue repair): inserted
  `stderr = strings.ReplaceAll(stderr, "503fatal:", "503\nfatal:")`
  before the split. `-run 'TestReviewRev5GluedLineRefuses'`: FAIL —
  `malformed full line produced 2 fetches, err=<nil>` (1.301 s),
  exactly the reviewer symptom. KILLED.
- Restored via backup copy; `sha256sum transport.go` =
  `fa4554cdf19f5200c55f27f5c9182b5117d3d7571effdaf5ff10e464af2c98ed`
  before the mutant and after the restore (identical).

## Findings / bounds

- Host anomaly (recorded, not a code finding): two post-restore `go
  test` invocations of the single regression hung past the 600 s
  default timeout while sibling workers ran concurrent `go test`
  suites on the same host (shared go-build cache/CPU contention; the
  same bytes passed in 1.3-2.1 s before and after via a prebuilt
  binary). Evidence above uses the prebuilt-binary run plus the full
  82 s set on the same bytes.
- Bounds unchanged from rev5: closed table covers only the enumerated
  shapes; a missing real-world diagnostic shape is a follow-up, not a
  refusal regression (per the rev5 review note). Windows refusal,
  POSIX process-group bounds, strict admission, broker binding, lock
  verification, sanitized errors unchanged. No caller wiring, no live
  credentials, no runtime-home changes. Checklist stays ticked.
