# TASK-260720-2284br — rework cycle 4 (R5) implementation notes

Worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1zntv0/worktree`
(base `origin/main` 17804ce; nothing staged, committed, or published).
Platform: darwin/arm64, Go 1.25.5.

Input: `TASK-260720-2284br_review-verdict-cycle-4.md` (R5, high, acceptance
blocking) and `TASK-260720-2284br_review-evidence-cycle-4.tar.gz`.

## R5 — the defect, and where it actually lived

All four declaration observations used the same three-step relation:

1. digest the path, outside the home lock;
2. read and parse the path as a *separate* operation;
3. under the home lock, digest the path again and compare only with step 1.

That catches A → B. It does not catch A → B → A around step 2: the closure is
resolved from the transient B while steps 1 and 3 both sample A, so the recheck
reports nothing and the run commits context, shims, adapters, and consumer state
for a declaration set that exists in neither the recorded nor the current state.

The window is not hypothetical on this codebase. Every supported writer rewrites
its manifest **in place** with `os.WriteFile` — `manifest.writeJSON`
(`curator add|remove`, `curator global add|remove`),
`scopes.AddHybridDecl`/`RemoveHybridDecl` (`curator hybrid add|rm`), and a hand
edit of `Skillfile.dev.json`. In-place truncate-and-rewrite is visible to a
reader that already holds the file open, so a parse genuinely consumes B.

Diagnosis accepted as the reviewer stated it. Nothing about the under-lock
ordering was wrong; the recorded generation simply did not belong to the parsed
bytes.

## Fix: one read owns both the bytes and the generation

New `internal/install/generation.go`. `readDocument` opens the path exactly once,
reads the payload from that handle, and returns the payload together with the
generation **of those bytes**. The parser is then handed the returned payload, so
"the generation this run revalidates" and "the bytes this closure was built from"
are the same object and cannot drift apart.

The reviewer offered a before/read/after protocol as an alternative. It was
rejected on inspection: before=A, after=A, parsed=B is exactly the ABA case, so a
path-sampling protocol cannot close it no matter how many samples it takes. Only
binding to the parsed bytes does.

Three consequences worth stating explicitly:

- **A torn read is safe.** Whatever bytes came back are what the parse saw and
  what the generation covers, so the settled file differs, the recheck reports a
  change, and the run restarts. It cannot commit.
- **A mode change no longer restarts.** A declaration generation covers bytes
  only; the old path digest included the permission bits. `chmod` selects nothing
  installed, so this removes a spurious restart rather than losing a signal.
  Covered by `TestDeclarationGenerationIsContentAddressed`.
- **A symlinked manifest is now revalidated by content.** `os.Open` follows a
  link exactly like the `os.ReadFile` the loaders always used, whereas
  `transaction.DigestPath` refused a link and pinned the observation to an
  `unreadable:` marker that compared equal to itself forever — i.e. a symlinked
  manifest used to be *not revalidated at all*. Behaviour change, in the safe
  direction, called out here because it is not in the verdict.

### Byte-level entry points, so no loader reads a path twice

Each owning package gained the byte-level parser its own `Load` is now built
from — additive, and `Load` keeps its exact contract for the read-only CLI
callers (`cmd/curator` hybrid list/status, `skill check`, `status`):

- `manifest.ParseBytes(payload, filePath)`
- `devsub.ParseBytes(payload, projectRoot)` — the project root stays a parameter
  because a relative substitution path resolves against it
- `scopes.ParseHybridDecls(payload, path)`

`internal/install` wraps each in a reader that returns `(value, generation, err)`
and reproduces its loader's absent semantics exactly (`readManifestDocument`,
`readSubstitutionsDocument`, `readHybridDocument`).

### The invariant is structural, not conventional

- `observations` is now **one record per key** — `generation`, `path`, and the
  reader that produced it — instead of parallel maps. An observation cannot exist
  without the location and the reader its recheck needs.
- `observation.current()` is the single recheck door and re-reads through the
  reader that recorded the entry. Two readers that disagreed about a path would
  restart every attempt until `MaxRestarts`, so this is a correctness rule, not a
  tidiness one.
- The four declaration keys have their own type, `documentKey`. A document can
  therefore only enter the set through `observeDocument`, which takes a
  generation; it cannot be passed to `observe`, which takes a path. Reintroducing
  the ABA-prone form is now a compile error rather than a review question.

### Call sites (four, all read once and observe the parsed generation)

| Input | Key | Site |
| --- | --- | --- |
| project manifest | `manifest/project` | `install.go:180,185` |
| project substitutions | `substitutions/project` | `install.go:220,225` |
| hybrid activation | `activation/hybrid-manifest` | `install.go:257,262` |
| machine-wide manifest | `manifest/global` | `global.go:81,86` |

Control flow is unchanged: `recheck` already mapped any generation difference to
`restartClosure`, and `runCommit` already calls it under the home lock after
journal recovery and before both cache publication and target staging.

## Tests

New permanent cases. `internal/install/generation_test.go` locks the binding at
the read; `internal/install/aba_test.go` drives the full A → B → A sequence
through each document's real writer.

The transient generation is applied inside the single read, through a one-shot
`afterDocumentOpen` seam (nil in production, same convention as
`godriver.controlSeamFault`). The seam fails the test if the window never opened,
so a case whose mutation missed the read cannot pass silently. The
byte-identical restoration lands at `OnStaged` — the documented pre-home-lock
boundary, after every private build succeeded and before the commit phase takes
the home lock — and each case asserts the restoration was byte-identical, so it
cannot pass for the one-way reason the R4 cases already cover.

| Case | Proves |
| --- | --- |
| `TestReadDocumentBindsGenerationToBytesRewrittenInPlace` | in-place rewrite mid-read: the generation follows the bytes the parser gets |
| `TestReadDocumentBindsGenerationToBytesReplacedByRename` | rename mid-read: payload stays on the old inode, generation follows it, and the recheck sees the replacement |
| `TestDeclarationGenerationIsContentAddressed` | absent is explicit; different bytes differ; byte-identical restoration is indistinguishable — which is *why* binding is required; chmod does not restart |
| `TestDocumentGenerationIsStableWhenUnreadable` | an unreadable document observes stably instead of restarting to `MaxRestarts` |
| `TestObservationsRecheckDocumentsAgainstTheParsedBytes` | recheck compares a document by its reader and a marker by its path digest |
| `TestProjectManifestABAAroundTheReadRestartsClosure` | `curator add` → parse → `curator remove`: restart, no `skill-b` context/shim/adapter, `skill-a` intact, no journal |
| `TestGlobalManifestABAAroundTheReadRestartsClosure` | same through `curator global add|remove`, machine-wide entries |
| `TestSubstitutionsABAAroundTheReadRestartsClosure` | transient hand-edited substitution: restart, no `SUBSTITUTION` message, and the committed marker records no `substituted` |
| `TestHybridActivationABAAroundTheReadRestartsClosure` | transient activation target through `curator hybrid add`: restart, no hybrid context, no adapter mirror |
| `TestByteIdenticalRewriteDuringTheReadDoesNotRestart` | control: a rewrite that changes no byte must commit, not loop |

Preserved unchanged: the six R4 one-way revalidation cases, the four R3
activation cases, and the R1/R2 transaction/staging/adapter suites.
`assertClosureRestart` took a `documentKey` instead of a `string`; no assertion
changed.

### Negative control — both directions

Restoring the pre-fix relation is a two-line change to `readDocument` (digest a
separately-read path instead of the payload). Diffs and logs in
`.temp/TASK-260720-2284br/negative-control-r5/`.

Variant 1, digest **before** the read (the shipped defect), exit 1:

- all four install-level ABA cases FAIL — and fail by *committing the transient
  declaration*: `skill-b tag v1 ... installed`, `SUBSTITUTED (path ...)`,
  `skill-h ... (hybrid) installed`, each with no restart message
- `TestReadDocumentBindsGenerationToBytesRewrittenInPlace` FAILS
- `TestByteIdenticalRewriteDuringTheReadDoesNotRestart` still PASSES, so the
  cases are not merely restart-happy

Variant 2, digest **after** the read, exit 1:

- `TestReadDocumentBindsGenerationToBytesReplacedByRename` FAILS
- the in-place case passes

So the two binding cases cover both orderings, and neither ordering can
reintroduce the defect without turning a test red. The implementation was
restored and re-verified before the gate run.

## Prior-cycle reviewer repros

- Cycle 2 (hybrid activation) and cycle 3 (project + global manifest) repros are
  additive test-only overlays. Re-run **unmodified** against the R5 tree: exit 0,
  all three pass — `gate-reviewer-overlays.log`.
- The cycle-4 ABA overlay is **not** re-runnable and was not claimed: it replaces
  `internal/install/install.go` with a pinned pre-R5 copy, so running it would
  exercise the old code rather than the fix, and its two hook points
  (`afterProjectManifestObserve`, `afterProjectManifestLoad`) name a window this
  fix deletes. `TestProjectManifestABAAroundTheReadRestartsClosure` is that
  scenario as a permanent test — same writers, same byte-identical assertion,
  same absence assertion — and the negative control above shows it reproduces the
  reviewer's exact failure.

## Gates

All re-run first-hand on the final tree by
`.temp/TASK-260720-2284br/gates-cycle5/run-gates.sh` (19:50:05 → 20:08:03); every
gate is a standalone process whose real exit code is written to a durable `.exit`
file as its last action, so a missing `.exit` means killed or running, never
passed. An earlier partial run of the same driver was killed on purpose before a
cosmetic `commit.go` edit, and its incomplete `gate-test-install.log` was
discarded rather than inherited — every number below postdates the final source.

| gate | exit | seconds |
| --- | --- | --- |
| gofmt -l . | 0 | 0 |
| git diff --check | 0 | 0 |
| go build ./... | 0 | 0 |
| go vet ./... | 0 | 1 |
| test-r5 (10 new cases) | 0 | 15 |
| test-revalidation (6 R4 cases) | 0 | 30 |
| test-loaders (manifest, devsub, scopes) | 0 | 1 |
| reviewer-overlays (cycle-2 + cycle-3 repros, unmodified) | 0 | 10 |
| test-install (whole package) | 0 | 190 |
| test-atomicity (whole package) | 0 | 406 |
| test-rest (remaining 37 packages) | 0 | 29 |
| race-r5 | 0 | 75 |
| race-revalidation | 0 | 108 |
| race-concurrency | 0 | 59 |
| race-activation | 0 | 98 |
| race-loaders | 0 | 3 |
| golangci-lint v2.4.0 run ./... | 0 | 2 |
| golangci-lint v2.4.0 run ./... after `cache clean` | 0 | — |

Zero data races reported by any race gate. Lint is honestly clean repo-wide, and
re-confirmed on a cold cache so the 0-issue result is not a cache artifact.

**Red gate, reported as red: `test-godriver` exited 1 (51s).** One subtest,
`TestBuildStopsBeforeBuildForEveryPreflightRejectionClass/escaped_embed`, hit the
sibling-owned 15.02s wall-clock probe deadline (`go-v1 process_timeout`). There is
no assertion failure anywhere in the log, and the machine was at load average 12+
from unrelated work (another agent session, iTerm2, suggestd, gramdrive). An
isolated rerun of the same tree exits 0 in 28.0s
(`gate-test-godriver-rerun1`). `internal/godriver` is owned by
TASK-260720-6i3cya, is untouched by this task, and its probe bound is
security-relevant — raising it so a test passes would be a forced fit against an
ownership boundary, so it is deliberately not patched here. The recommendation
from cycle 3 stands: a separate task for its owner to make the probe deadline
contention-tolerant.

## Boundaries, not claimed

- Darwin/arm64 only. No Windows runtime evidence.
- Hardened fail-closed containment (STORY-260728-327soo) is separate and
  non-gating; nothing here claims kernel or container isolation.
- `internal/godriver` is sibling-owned (TASK-260720-6i3cya) and untouched. Its
  wall-clock probe cap can still time out under heavy unrelated machine load; it
  is run as its own invocation with nothing else in flight, and its real exit is
  recorded either way.
- The user configuration remains a documented boundary rather than a freshness
  proof: it is loaded once by the entrypoint and passed through as one immutable
  value, so a mid-run edit is picked up by the next run, not this one.
