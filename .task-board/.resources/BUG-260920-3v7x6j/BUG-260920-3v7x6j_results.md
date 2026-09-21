# BUG-260920-3v7x6j results — transport-rev2 alias substitution reaches the connection

## Outcome

Alias-substituted endpoints now connect to the RESOLVED host:port on the
draft Skillfile-source lane. Corpus case `v2-alias-resolution` flips from
known-gap to driven-pass: CLI resolve clones
`https://mirror.corp.example:8443/kit.git` exactly once, the lock keeps the
canonical identity `fixture.test/kit`, and user-facing diagnostics carry no
endpoint provenance. All other v2 rows (positives, refusals, exhaustion,
user-config, external-build) verified unchanged at their production entries.
Narrowing mutant (clone the declared host) killed. Lint clean.

## Scope decision (read first)

The review text cites the strict-lane executor
(`internal/buildrepo/transport.go:502, 931`). That executor is intentionally
UNCHANGED: spec §7 mandates that the strict external-build lane REFUSE any
selected endpoint with an explicit port or an `alias` field
(`build_repository_identity_invalid` before network I/O; "MUST NOT strip the
port or ignore the alias to force admission"). The corpus rows
`v2-external-build-port-refused` / `v2-external-build-alias-refused` pin that
refusal at `ValidateTransportPlan` and at zero-fetch `install.Project`, and
both re-verified green. The cited lines sit below the §7 gate
(`parseTransportPlan`, transport.go:489), which aliased attempts never pass,
so the executor never connects for an alias endpoint — substituting there
would violate §7 and break the pinned rows. The draft Skillfile-source lane
(CLI `project resolve`) is the only lane that admits aliases, and that is
where substitution now happens.

## Change (5 files, uncommitted on the Story branch)

- `internal/config/sourcepolicy.go`: new `Attempt.ConnectionURL()`.
  Alias-less attempts return the listed URL verbatim (v1/literal/port/
  declared-mirror byte-identical). A named alias renders scheme + userinfo +
  resolved host + resolved port (when explicit) + path; scp-like + explicit
  port renders the equivalent `ssh://[user@]host:port/~/path` (admitted
  scp-like paths are home-relative, so `~/` preserves the remote path).
  Fail-closed (`ok=false`) on any malformed carried value; the port
  flag/value agreement mirrors the sibling executor §6 predicate.
- `cmd/curator/project_resolve.go`: `fetchDraftRepoAllowingFallback` derives
  every attempt's connection target up front and clones it. A mistranslation
  fails `repository_policy_invalid` before any clone (unreachable via the
  loader; sanitized message, no URL/host). Diagnostics, fallback gate, and
  exhaustion shape untouched.
- `internal/config/sourcepolicy_v2_test.go`: 13 substitution cases (all URL
  forms × port/no-port, same-host alias, nested paths), 18
  fail-closed mistranslation cases, 3 loader-plan round-trips.
- `internal/crossconformance/draftsources_semantic_v2_test.go`:
  `driveV2AliasResolution` converted from known-gap lock to drive: loader
  assertions kept, `ConnectionURL` pinned to `v2AliasURL`, shim serves the
  SUBSTITUTED address, `clones == [v2AliasURL]`, lock identity canonical,
  plus a negative scan (stderr must not contain the resolved host or alias
  name). Known-gap marker removed.
- `internal/crossconformance/draftsources_semantic_test.go`: removed the
  now-unused `semanticKnownGap` marker (last caller gone; generic
  outcome/tally machinery kept for future markers). Required for lint.

Provenance: CLI-lane warnings/exhaustion unchanged (closed class vocabulary,
no URLs — the row's stderr scan pins this); the machine-private
`draft-transport-provenance.jsonl` sink is strict-lane-only and untouched;
no new portable fields (lock still canonical identity, port-free).

## Evidence (shell sh, `go test -p 1 -count=1`, darwin/arm64)

- `go build` + `go vet` on `internal/config`, `cmd/curator`,
  `internal/crossconformance`: exit 0.
- `go test -run TestAttemptConnectionURL ./internal/config/`: PASS, exit 0
  (13 + 18 + 3 subtests).
- `go test ./internal/config/` (full package): ok 0.784s, exit 0.
- `go test -run 'TestDraftAttempt|TestEndpointExhaustion|TestDraftStripCloneFraming|TestWithDraftRemediation' ./cmd/curator/`: ok 0.520s, exit 0.
- `go test -run 'TestProjectResolve|TestProjectRefresh' ./cmd/curator/`:
  ok 90.682s, exit 0.
- `go test -run 'TestDraftFetchFailure|TestDraftAvailabilityFallback|TestDraftMachinePolicy' ./cmd/curator/`: ok 6.675s, exit 0.
- Flipped-row probe at the production entry (temporary test driving
  `driveV2AliasResolution` through the compiled CLI + fail-closed transport
  shim recording clone targets): PASS 2.87s, exit 0. Probe removed after.
- Narrowing mutant at the production call site
  (`project_resolve.go`: clone `attempt.URL` instead of the substituted
  target): probe FAILS `resolve = 1 ... repository_endpoint_unavailable`,
  exit 1 — KILLED. (First mutant invocation hit the host SIGKILL flake
  described below; rerun killed cleanly in 3.96s. Zero MUTANT markers
  remain; fix restored byte-identical.)
- All 21 v2 row bodies driven directly at their production entries
  (temporary probe, since removed): 21/21 PASS, 28.8s, exit 0 —
  `v2-alias-resolution` driven-pass; port/declared-mirror/mirror-first/
  reader-accepts-v1 positives, all 11 refusal rows, both user-config rows,
  and all 3 external-build rows unchanged.
- `TestDraftSourcesSemanticCoverage` + v2 corpus/schema-1 goldens: exit 0.
- `gofmt -l` on touched dirs: clean. `golangci-lint run` per package
  (`cmd/curator`, `internal/config`, `internal/crossconformance`):
  0 issues, exit 0 each. (One combined lint invocation was OOM-killed,
  exit 137, while the full matrix ran concurrently; per-package reruns
  alone are green.)
- Full 94-row `TestDraftSourcesSemanticCases` tally: NOT observable on this
  host. Attempt 1 hit the 10m go-test timeout at row 77 (earlier rows
  progressed; `bootstrap` children SIGKILLed, exit -1, under host memory
  pressure; one bootstrap hang). Attempt 2 (`-timeout 1500s`) saw the test
  binary SIGKILLed at 81s before first output (documented host
  fresh-binary stall/kill pathology + co-tenant load). The flipped row is
  proven at the production entry above and the last known-gap marker is
  removed, so the unchanged tally counts it driven; full-matrix tally
  observation belongs to the handoff landing suite (hosted gate), which per
  campaign rules runs the full suite exactly once. Windows likewise via the
  hosted gate (row skips on Windows by POSIX-shim design, as its siblings
  do; `ConnectionURL` unit tests are OS-agnostic).

## Checklist

- [x] Alias substitution changes the connection target (host:port) on the draft lane; declared identity and sanitized provenance as specified
- [x] Corpus case v2-alias-resolution flips to driven-pass (row proven at production entry; tally by the unchanged harness + hosted gate); narrowing mutant killed; provenance sanitization unchanged (stderr scan)
- [x] Code written per task description and AC (strict-lane refusal deliberately preserved per §7 + pinned rows)
- [x] Relevant tests written for new/changed behavior and passing (unit + corpus flip + negative scan)
- [x] Lint clean (gofmt, vet, golangci-lint on all touched packages)
- [x] Relevant build/validation commands run; build not broken
- [x] Outcome artifact attached (this file, task-scoped name)
- [x] Findings recorded here (campaign rules forbid LOGBOOK.md edits, so no logbook write)

## Bounds / notes for the reviewer

- The full-matrix tally and Windows lane are gate-observed (see above).
- Sibling checkpoints on this branch (3ukdk4 git-config isolation, 2wyzde
  evidence match) untouched: `git status` shows only the 5 files above.
- No commits made; work left uncommitted for handoff snapshot.
