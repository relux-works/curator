# TASK-260910-stbg4d — expose source workflow and actionable diagnostics (handoff evidence)

Candidate: uncommitted working tree of `.temp/STORY-260910-1cnwwp/worktree`
on top of checkpoint `97a161d` (trunk with STORY-20sx61 + STORY-1s75e1).
Shell: bash; every exit code below was captured from the binary
(`echo $?` / `${PIPESTATUS[0]}`), never through an unverified pipe.
No commits on the Story branch.

## Changed paths

```
M  README.md                       (draft section, labelled opt-in)
M  cmd/curator/main.go             (+10/-4: remediation wiring, gated help, project -h)
M  cmd/curator/project_resolve.go  (sanitized fetch diagnostics + remediation wiring)
M  docs/cli.md                     (+209: refresh section, draft workflow, examples, policy setup)
M  docs/troubleshooting.md         (+166: stable-class diagnostics section)
M  internal/buildrepo/transport.go (+7: exported FailureClass.Reason accessor, additive)
?? cmd/curator/draft_diagnostics.go      (159 lines: remediation table, sanitized clauses)
?? cmd/curator/draft_help.go            (91 lines: workflow help, gated usage)
?? cmd/curator/draft_diagnostics_test.go (1019 lines: 14 CLI-entry tests)
```

No new command or flag. No spec edit. No production change outside
cmd/curator except the additive Reason() accessor.

## What it implements (scope line + acceptance)

Existing cmd entry points keep their verbs; this leaf adds the
presentation layer the landed lanes lacked:

- `cmd/curator/draft_diagnostics.go` (`withDraftRemediation`):
  sanitized remediation for the 9 skillfile-sources §5 classes, the 4
  repository-transport §§2/6 classes, and the 2 source-audit-v1
  outcomes. Static prose only — nothing from the failing input is
  interpolated, so no secret or endpoint can leak. Entries whose
  underlying message already guides the operator (stale, unavailable,
  changed, exhaustion, draft-lane refusal) are left byte-identical
  instead of repeated. Wired into `printFailures` (install/upgrade/
  status/global paths), `cmdProjectResolve` (resolve/refresh), and the
  draft status drift refusal. Messages without a stable class —
  every frozen v1 message — return unchanged.
- `cmd/curator/project_resolve.go` (fetch path): warnings and the
  exhaustion error now render the revision-2 shape — endpoint order,
  transport, opaque provider, closed class + reason, portable identity,
  shared remediation — via `draftAttemptClause`/`endpointExhaustion`/
  `fetchExhaustion`. Raw git output (which embeds the full URL, e.g.
  `git clone failed for https://...`) feeds classification only and
  never reaches stderr. Previously the raw error was printed verbatim.
- `cmd/curator/draft_help.go`: `project resolve|refresh -h|--help|help`
  prints the workflow golden (previously `-h` was misparsed as a
  project path); `install -h`/`status -h` append the workflow section
  only while `CURATOR_DRAFT_SOURCES_V1=1`, byte-identical otherwise.
  Top-level `usage` is untouched by decision (see bounds).
- Docs: `project refresh` section (was missing entirely), "Draft
  Skillfile sources" in docs/cli.md (workflow, local/absolute/Git/
  collection examples, machine policy setup with endpoints/pin/
  fallback, providers pointer, root_inputs), stable-class section in
  troubleshooting, labelled README section. All draft support is
  labelled "opt-in, unreleased"; released v1 behavior is unchanged.

## Acceptance clauses — evidence

| clause | evidence (all at the CLI production entry `run()`) |
|---|---|
| coherent existing-command UX, local/absolute/Git/collection examples | `TestDraftDocumentedLocalShapesThroughCLI` (5 rows: relative verbatim from docs, absolute, list, star, star+exclude — resolve, dry-run, install, status each); `TestDraftDocumentedGitRevisionThroughCLI` (revision pin); tag/branch/collection Git shapes covered by existing `TestProjectResolveGitSelectionThroughCLI`, `TestProjectRefreshGitBranchMembershipChangeThroughCLI`, `TestProjectResolveGitCollectionPinnedToTagThroughCLI` (all green in the sibling rerun); `TestDraftDocsPinExamples` fails if docs drift from the exercised shapes |
| documented machine policy setup | `TestDraftMachinePolicySetupThroughCLI`: logical `repository` declaration resolves through `source-policy.json` beside the test config (pinned via `CURATOR_CONFIG`), lock binds the policy-selected commit, install + status green; bad-pin refresh fails closed with remediation and the prior lock byte-identical |
| stable errors carry sanitized remediation | `TestDraftRemediationThroughCLI` (15 rows, one per class incl. both audit outcomes): nonzero exit, class + remediation on stderr, no duplicate guidance; `TestEndpointExhaustionCarriesRemediation` (helper seam); `TestWithDraftRemediationTable` (unit: idempotence, already-guided, legacy untouched) |
| no secrets / endpoint provenance in user-facing text | `TestDraftFetchFailureSanitizedThroughCLI`: unroutable host → class + `endpoint 1 (https)` + remediation present; full URL, scheme, `.git` suffix, raw tool output absent; no lock published. `TestDraftAttemptClauseShape`: clause never contains the URL |
| launch never rescans live inputs | `TestDraftLaunchConsumesFrozenRuntimeThroughCLI`: shim prints v1, live mutation to v2 → shim still v1, status up-to-date, reinstall-without-refresh still v1, refresh+install → v2; tampered frozen runtime → `source_snapshot_changed` instead of launching |
| repair | `TestDraftRepairRestoresDriftedContentThroughCLI`: drift → `content-drift` → install restores bytes, lock identical, status up-to-date |
| draft labelled accurately | help golden + docs pins assert "opt-in, unreleased" and the switch name; `-h` sections only describe the lane as draft |
| released v1 retained | switch-off `-h` outputs byte-identical to flag defaults (gating test); legacy suites green (below); `usage` const untouched; underlying error strings untouched (remediation appended at print) |

## Narrowing mutants (each restored; residue grep clean; sibling trees diff-clean)

| id | mutant | killer | result |
|---|---|---|---|
| M1 | `withDraftRemediation` returns message unchanged | `TestDraftRemediationThroughCLI/alias-unknown` + unit table | KILLED exit 1 |
| M2 | raw `cloneErr` (with URL) restored in fetch warning | `TestDraftFetchFailureSanitizedThroughCLI` | KILLED exit 1 (output shows the exact leak the fix removes) |
| M3 | `endpointExhaustion` drops remediation | CLI row: SURVIVED exit 0 — the presentation wrapper re-adds it (redundant layer by design); `TestEndpointExhaustionCarriesRemediation` (helper seam, added for this) | KILLED exit 1 |
| M4 | draft `-h` section always appended | `TestDraftInstallStatusHelpGated` (switch-off half) | KILLED exit 1 |
| M-L1a | bogus shim target in `WriteBinShim` | launch test | SURVIVED exit 0 — wrong seam (project shims stage via `StageShimTransition`, not `WriteBinShim`); mutant-placement error, documented |
| M-L1 | bogus shim target in `UnixShimContent` | `TestDraftLaunchConsumesFrozenRuntimeThroughCLI` | KILLED exit 1 (shim exits 126) — proves execution is real |
| M-L3 | `authenticateLocalSnapshot` no-op | launch test tamper step | KILLED exit 1 (tampered store adopted, install exits 0) — proves verified-frozen consumption |

Bound, not a gap: no single-point mutant expresses "rescan a live
input at launch" because live paths are unreachable from frozen
consumption by construction (`OpenDraftFrozen` takes only
home/lock/opts). The two-phase v1→v1→v2 shim assertion discriminates
frozen-vs-live directly: a rescanning launcher would print v2 at
phase 2.

## Independently executed checks (real exit codes, macOS)

New tests (`-p 1`, `-count=1`, exit from the binary):

- `TestWithDraftRemediationTable|TestDraftAttemptTransportTable|TestDraftAttemptClauseShape|TestDraftDocsPinExamples`: exit 0 (1.4s)
- `TestDraftProjectResolveHelp|TestDraftInstallStatusHelpGated`: exit 0 (1.0s)
- `TestDraftDocumentedLocalShapesThroughCLI` (5 subtests): exit 0 (39.2s)
- `TestDraftRemediationThroughCLI` (15 subtests): exit 0 (5.1s in the full run; 17.9s standalone)
- `TestDraftMachinePolicySetupThroughCLI|TestDraftDocumentedGitRevisionThroughCLI`: exit 0 (19.4s)
- `TestDraftFetchFailureSanitizedThroughCLI`: exit 0 (3.2s; 0.1s warm)
- `TestDraftLaunchConsumesFrozenRuntimeThroughCLI|TestDraftRepairRestoresDriftedContentThroughCLI`: exit 0 (34.7s)
- Full new set, 14 tests one mask: exit 0 (82.4s, all `--- PASS`)
- `TestEndpointExhaustionCarriesRemediation`: exit 0 (0.5s)

Legacy/sibling regression (this session, unmodified masks):

- `TestDraftStatus|TestProjectResolve|TestClassifyDraftMember`: exit 0 (193.1s)
- `TestStatus|TestRun|TestUsage|TestInstall`: exit 0 (130.3s)
- buildrepo `TestExhaustionErrorIsClosedVocabulary|TestResolvedTransportExhaustionListsSanitizedClasses|TestResolvedTransportLeaksNothingIntoErrors|TestClassifyFetchOutput|TestClassifyAdmissionCode`: exit 0 (4.5s)

Static gates:

- `go build ./cmd/curator/ ./internal/buildrepo/`: exit 0
- `go vet ./cmd/curator/ ./internal/buildrepo/`: exit 0
- `gofmt -l cmd internal`: no output; `git diff --check`: clean
- `golangci-lint run ./cmd/curator/... ./internal/buildrepo/...`: exit 0, 0 issues (11 revive unused-parameter findings fixed during the run)

## Bounds (not findings)

- Whole-repository suite, race lanes, and Windows/Ubuntu rows are left
  to the hosted gate on the exact candidate tree (narrow packages only
  per the wave note). POSIX-only rows use declared reasons verbatim:
  `test transport wrapper is POSIX-only` (git/policy), `executes POSIX
  skill commands` (launch); `git is not available` follows existing
  precedent. No new skip class.
- Top-level `usage` is intentionally untouched: the draft workflow is
  discoverable via `project resolve|refresh -h`, switch-gated
  `install|status -h`, docs/cli.md, README, and troubleshooting. A
  reviewer preferring a top-level pointer can request it without
  touching behavior.
- `docs/draft-*.md` still say "not an enabled path" in places; they
  are sibling-owned internal notes, out of this scope line, left
  untouched.
- The draft-alias fetch path (`gitops.Clone/Fetch`) does not bind
  policy authentication providers (ambient git credentials apply);
  docs describe endpoint selection only and point providers at the
  manager's existing operator mechanism, without overclaiming.
- One early combined run (cold build cache + a concurrent `go run`
  smoke probe) hit the 400s `go test` timeout environmentally; the
  rerun of the same mask was green in 1.4s. The smoke probe itself
  exited 0.
- LOGBOOK.md was not edited per campaign worker policy (control-root
  writes are the orchestrator's); findings live in this outcome.

---

# Revision 2 — rework after CHANGES_REQUESTED (rev1 verdict, claude-opus-5)

Same uncommitted working tree on checkpoint `97a161d`; no commits on
the Story branch. Scope is exactly stbg4d-rework-1.md items 1–4 (F1,
F2, F3, nits); every surface the verdict accepted as-is is untouched.
Shell: bash; every exit code below is the binary's own
(`echo $?` / `${PIPESTATUS[0]}`).

## Delta vs revision 1

```
M  cmd/curator/main.go             (help intercept gated on the switch, bare help dropped)
M  cmd/curator/project_resolve.go  (stripCloneFraming + regexp import + one-line call-site change)
M  cmd/curator/draft_help.go       (comment: the -h gating rule)
M  cmd/curator/draft_diagnostics_test.go  (help test rewritten, fetch test rewritten,
                                   +TestDraftAvailabilityFallbackThroughCLI,
                                   +TestDraftStripCloneFramingTable, +fake-git arms helper,
                                   repository-alias-unknown rename, +2 docs pins)
M  docs/cli.md                     (F2 rule at resolve+refresh, F3 providers sentence)
```

No other path touched: the fetch-failure branch, `gitFailureDetail`,
the clone-loop retry structure, the transport closed table, README,
and troubleshooting are byte-identical to revision 1.

## F1 — clone framing no longer poisons classification

`stripCloneFraming` (cmd/curator/project_resolve.go) drops git's exact
`Cloning into '<dir>'...` progress line from clone stderr before
`ClassifyFetchOutput`, in the clone loop only. Only exact matches are
dropped, so classification stays fail-closed. A throwaway probe
(written, run, removed) confirmed the verdict's diagnosis precisely:
`gitFailureDetail` already returns the clean two-line stderr — there
is no double-wrap problem (the clone wrapper says "failed for", so
`Index` finds the inner marker; the alternative hypothesis was tested
and discarded) — and the `Cloning into` line alone flips the class to
`unknown`.

- (a) `TestDraftFetchFailureSanitizedThroughCLI` (rewritten): a
  fake-git arm emits the production two-line DNS stderr (progress
  line + `fatal: ... Could not resolve host: invalid.invalid`, exit
  128); `project resolve` fails with `endpoint 1 (https):
  availability: endpoint unavailable (...)`, remediation present, no
  URL/scheme/raw-output/`unknown` leak, no lock published. PASS
  (0.4s). No network: the former live DNS lookup is gone.
- (b) `TestDraftAvailabilityFallbackThroughCLI` (new): two-endpoint
  `availability-auth` policy (arm 1 DNS-failing, arm 2 an scp-like
  URL on the same canonical host rewritten to the bare fixture) →
  exit 0, lock binds the alternate commit with identity
  `fixture.test/kit`, stderr carries only the endpoint-1
  availability clause (no `endpoint 2: not attempted`, no leaks).
  PASS (7.7s). Fixture notes: policy validation requires
  `authentication` on every endpoint and every URL to canonicalize
  to the entry key — both confirmed by fast failures while shaping
  the test.
- (c) Narrowing mutant (call-site strip removed, framing
  re-poisoned): the sanitized test FAILS (`stderr misses "endpoint 1
  (https): availability: endpoint unavailable"`, stderr shows
  `unknown: unclassified failure` — the exact revision-1 symptom)
  and the fallback test FAILS (exit 1, fallback never advances).
  KILLED, exit 1. Residue grep clean. The unit table passes under
  this mutant (it pins the helper, which is intact) — expected and
  documented here.
- `TestDraftStripCloneFramingTable` (new, renamed into the
  `TestDraft` mask): framing+DNS → availability, DNS alone →
  availability, framing alone → unknown, near-match kept → unknown,
  empty → unknown, diagnostic-only input byte-identical. PASS.

## F2 — switch-gated help, no bare-help hijack

`cmdProject` (cmd/curator/main.go) now intercepts `-h`/`--help` for
`project resolve|refresh` only when
`install.DraftSourcesEnabled(os.Getenv)` is true — the same rule as
`appendDraftUsage`. With the switch off the spellings fall through
to the frozen v1 path untouched; the bare word `help` is never
intercepted and resolves as alias/path. Rule documented in docs/cli.md
at both `resolve -h` and `refresh -h` (plus the bare-`help` sentence).

- Committed `TestDraftProjectResolveHelp` (rewritten, both halves):
  switch off → `-h`/`--help` byte-identical (code+stdout+stderr) to
  the bare verb for resolve and refresh, no draft markers; alias
  `help` resolves to the exact v1 report with the switch off AND
  on; switch on → `-h`/`--help` print the workflow golden. PASS.
- Manual base-vs-candidate (switch unset, v1 fixture, binaries from
  a HEAD archive vs the worktree): 15/15 byte-identical on
  exit+stdout+stderr — the reviewer's 12 invocations (including the
  three revision-1 differers `project resolve -h|--help|help`) plus
  the three `project refresh` counterparts. Script kept as scratch
  (`/tmp/stbg4d/basediff.sh`, not committed). Exit 0.
- Switch-on binary sanity: `project resolve -h` prints the golden
  head (exit 0); `project add help` + `project resolve help`
  prints the v1 report for alias `help` (exit 0).

## F3 + nits

- docs/cli.md "Machine policy setup" now states that named
  authentication providers (`source-providers.json`) apply to the
  external build lane under `CURATOR_DRAFT_TRANSPORT_RESOLUTION`,
  while Skillfile sources use the operator's ambient Git
  credentials and the `authentication` identifier is validated but
  unused on this lane. Pinned by `TestDraftDocsPinExamples`
  (`ambient Git credentials`, `CURATOR_DRAFT_TRANSPORT_RESOLUTION`).
- Second `alias-unknown` row renamed to `repository-alias-unknown`
  (15/15 remediation rows pass, no `#01` suffix). README
  "unreleased" pin kept (README untouched).

## Checks (macOS, bash, real exit codes)

The required single-mask invocation (`-run
'TestDraft|TestProjectResolve'`, `-timeout=400s`) does not fit one
go-timeout on this host: it recorded 48 PASS and 0 FAIL before the
400s alarm fired inside sibling `TestDraftTransportProviderAdmission`.
Per the headless-run rule the mask was split into bounded sequential
calls; every chunk below completed with its own exit code:

- L1 leaf fast (8 tests: both remediation tables, both clause/shape
  tables, exhaustion seam, strip table, both help-gating tests, docs
  pins): exit 0 (2.2s)
- L2 `TestDraftDocumentedLocalShapesThroughCLI` +
  `TestDraftRemediationThroughCLI`: exit 0 (61.6s)
- L3 git/launch (machine-policy, git-revision, fetch-sanitized,
  availability-fallback, frozen-launch, repair): exit 0 (72.2s)
- S1 `TestProjectResolve` (18 sibling): exit 0 (159.9s)
- S2 `TestDraftStatus` (14 sibling): exit 0 (61.4s)
- S3 `TestDraftTransport` (3 sibling): exit 0 (72.7s)
- Full required mask (49 tests) + 2 helper tables: all green, zero
  failures observed in any run including the timed-out one.
- `internal/buildrepo -run 'Classify|Transport'`: exit 0 (160.4s
  with `-timeout=300s`; the suggested 120s budget does not fit this
  host uncontended, so the bound was raised and is recorded here).
- `go build` + `go vet` on `cmd/curator` + `internal/buildrepo`:
  exit 0; `gofmt -l cmd internal` empty; `git diff --check` clean;
  `golangci-lint run ./cmd/curator/... ./internal/buildrepo/...`:
  0 issues, exit 0.
- Base-vs-candidate 15-case switch-off diff: exit 0 (all identical).

## Bounds (revision 2)

- One transient: the first fallback-test run hit the documented host
  stall on a fresh test binary (6m alarm); the immediate rerun was
  green in 1.7s and the test has been stable across every chunk
  since. A standalone /tmp reproduction of the fake-git script
  likewise hung once environmentally, then ran instantly twice —
  script logic verified correct via `sh -x` trace.
- New fake-git tests skip on Windows with the declared reason
  verbatim (`test transport wrapper is POSIX-only`); hosted gate
  covers the rest. No new skip class.
- Whole-repository suite, race lanes, and hosted
  ubuntu/macos/windows rows are left to the gate on the exact
  candidate tree. LOGBOOK.md untouched per campaign worker policy.
