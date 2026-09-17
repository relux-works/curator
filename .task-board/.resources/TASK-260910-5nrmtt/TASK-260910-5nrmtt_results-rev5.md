# TASK-260910-5nrmtt results — rework 3 (rev5)

Producer: developer. Shell for all commands below: bash, no pipelines
masking exit status except where noted (every quoted exit code is the
gate's real status; mutant-B and restore checks used file redirect,
not pipes).

## Revision

- Rework brief: `5nrmtt-rework-3.md` (close the diagnostic grammar;
  rev4 verdict `TASK-260910-5nrmtt_review-verdict-rev4.md`, reproducers
  `TASK-260910-5nrmtt_review-reproducers-rev4.patch`).
- Spec: curator-spec checkout `07e2b41`
  (`protocol/repository-transport.md` rev 1 §§2–3; revision 2 not in
  scope); worktree `task-board/story/STORY-260910-1bhj0g` at
  `38e62cb8f4b0a427625e17dc6a02cf825f4407a4` (same base as rev4).
- Tree: `git status --short` shows exactly the 12 rev2 paths (4
  modified: `internal/buildrepo/admission.go`,
  `internal/buildrepo/httpsbroker_test.go`, `internal/gitcred/gitcred.go`,
  `internal/gitcred/provider_test.go`; 8 new:
  `docs/draft-transport-resolution.md`,
  `internal/buildrepo/{process_unix,process_windows,sshbroker,sshbroker_test,transport,transport_test}.go`,
  `internal/gitcred/provider.go`). Only `transport.go`,
  `transport_test.go`, and `docs/draft-transport-resolution.md` differ
  from rev4; Windows refusal, admission, brokers, and process-graph
  files are byte-identical to the rev4-accepted tree.

## P1 — closed line table (`internal/buildrepo/transport.go`)

Replaced the substring/prefix classifier (reason signals, standalone
signals with open tails, the unknown-facility ignore rule, the bare
resolver-fragment shape) with one closed table,
`transportLineSignals`: every entry is anchored `^…$` and built only
from fixed text plus bounded tokens — dotted RFC-1123 hostname
(`transportHost`), optional `user@` prefix on ssh lines, `[0-9]{1,5}`
port, three-digit HTTP status, single-quoted endpoint span
(`'[^']+'`), numeric `[0-9]{1,20}`/`[0-9]{1,3}` counters, enumerated
reason phrases (resolver reports, refused/timed-out/unreachable/no-route
connections, `Operation timed out after … milliseconds with … bytes
received`, 401/403 and 502/503/504 wrappers, `RPC failed; HTTP …`,
headless `could not read Username|Password …` prompts, forge
rejections, the ssh daemon method list over
`publickey|password|keyboard-interactive|hostbased`, TLS sentences,
host-key sentences, ref/audit/integrity shapes). There is no `.*` and
no `.+` anywhere in the file (verified by fixed-string grep, empty
output). A line matching no entry — unknown facility, future-git
sentence, secret echo, warning, progress line, any unconsumed tail —
poisons the whole output to unknown: one fetch, lane diagnostic, no
alternate traffic. Only git's four fixed ssh/rpc framing trailers are
skipped as non-evidence (framing-glue restore kept for those two
`fatal:` sentences only). Empty output and the truncation flag still
fail closed as before.

Deliberate boundary, kept from the committed contract: `Recv failure:
Connection reset by peer` in the HTTPS wrapper stays unknown
(`reset-is-not-refused`), even though reset-by-peer reads like a
transport phrase — the committed test demands fail-closed and the
reviewer never contested it.

## Tests (`internal/buildrepo/transport_test.go`)

- `TestReviewRev4ClosedGrammar` — the three rev4 reproducer probes
  applied as the regression (one gofmt-mandated blank-line removal
  inside the composite literal; names, messages, fetch-count and error
  assertions byte-identical). All three refuse after 1 fetch.
- `TestResolvedTransportClosedGrammarTable` (new) — 14 accepted
  fallback-eligible line shapes (11 HTTPS-first, 3 SSH-first), each
  driven at `AcquireNetworkResolved`: positive proves fallback (2
  fetches, locked commit, first record carries availability/auth);
  three one-character mutations per shape (appended `;` tail, removed
  quote or broken `:` facility separator, extra `audit: policy
  denied` line) must refuse (1 fetch, lane diagnostic, one unknown
  record). 56 entry subtests.
- `TestClassifyFetchOutput` — corrected four cases that encoded the
  deleted ignore rule (`secret-tail-poisons-to-unknown`,
  `unknown-facility-poisons-dns`, `hint-line-poisons-dns`,
  `ssh-warning-poisons-auth` now expect unknown); added poison
  twins of the rev4 probes plus enumerated reason variants
  (ssh-resolve ×2, ssh-connect ×2, https-connect ×2, password/tty
  prompts, bare TLS). 80 subtests.
- Flood helper now repeats a closed-table availability line
  (`ssh: connect to host fixture.test port 22: Connection refused`)
  instead of the torn bare fragment no strict grammar can cover; the
  truncation test still proves the flag (not the content) closes the
  plan: retained filler reads availability at unit level, production
  refuses after 1 fetch with an unknown record.
- `TestResolvedTransportLeaksNothingIntoErrors/exhaustion` no longer
  echoes the live secret into stderr (any echo now poisons to unknown
  by construction); it still binds the live broker secret into both
  attempts and asserts the closed-vocabulary exhaustion and records
  are secret-free. Echo sanitization is covered by the `fail-closed`
  subtest (TLS + echo, asserts no leak).
- `TestResolvedTransportEnvironmentalNoiseStillFallsBack` rewritten
  as `TestResolvedTransportClosedTableFramingStillFallsBack`: RPC 503
  plus the hung-up framing trailer (both table entries) still falls
  back and proves the locked commit. All rev1/rev2 probes kept
  verbatim and green.

## Evidence (real exit codes)

- Unit grammar/admission/gate:
  `go test -p 1 -count=1 ./internal/buildrepo/ -run
  'TestClassifyFetchOutput|TestClassifyAdmissionCode|TestSemanticFallbackGate'`:
  exit 0 (80 grammar subtests + parent PASS, 0 FAIL).
- Rev4 probes: `-run 'TestReviewRev4ClosedGrammar'`: exit 0, 3/3
  PASS (each ~0.4 s, 1 fetch).
- New table: `-run 'TestResolvedTransportClosedGrammarTable'`:
  exit 0, `ok … 31.539 s`, 0 FAIL.
- Related entry set (rev1 ambiguous ×2, rev2 mixed, unknown-tail,
  framing positive, truncation, leaks, admission retention,
  platform gate): all PASS, exit 0.
- Full `./internal/buildrepo` + `./internal/gitcred` suites:
  exit 0 (`ok … 97.256 s`, `ok … 4.340 s`).
- `go vet ./internal/buildrepo/ ./internal/gitcred/`: exit 0.
- `GOOS=windows go build` + `GOOS=windows go vet
  ./internal/buildrepo`: exit 0.
- `gofmt -l internal/buildrepo internal/gitcred`: no output.
  `git diff --check`: exit 0.
- Full module suite deliberately NOT run (host rule; gate at handoff).

## Narrowing mutants (bytes restored, sha256-identical)

- M-A (P1 every-line-must-match): poison return
  `FailureUnknown, true` → `false` (unmatched lines ignored).
  `-run 'TestResolvedTransportUnknownTail…|TestReviewRev4ClosedGrammar'`:
  FAIL — rev4 audit-facility probe fell back (2 fetches, err=nil).
  KILLED. Restored (`sha256sum` identical before/after).
- M-B (P1 closed tail): HTTP-50x entry loosened with `.*`.
  `-run 'TestReviewRev4ClosedGrammar|TestResolvedTransportClosedGrammarTable'`:
  exit 1 — rev4 503-tail probe and `https-503/tail` fell back to a
  ready alternate and succeeded. KILLED. Restored (identical sha256,
  zero `MUTANT` markers, no `.*`/`.+` in the file).
- Post-restore check (rev4 + `https-503` table subset): exit 0.

## Findings / bounds

- A secret-echoing git now fails closed (unknown) instead of riding
  through to exhaustion: strictly stronger posture; sanitization
  holds on both paths by construction (stderr never interpolated).
- First-contact ssh `Warning:` lines likewise fail closed; only the
  four fixed framing trailers are skippable besides diagnostics.
- Windows refusal, POSIX process-group bounds, strict admission,
  broker binding, lock verification, and sanitized errors are
  unchanged from the rev4-accepted tree (not re-proved here beyond
  the green suites). No caller wiring, no live credentials, no
  runtime-home changes. Hosted gate verdicts for this tree come from
  the rev5 gate run at handoff. Checklist on the task stays ticked.
