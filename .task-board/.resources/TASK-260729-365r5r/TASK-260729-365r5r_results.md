# TASK-260729-365r5r — prototype: saveJournal namespace validation, O(P²) → O(P)

Status of this document: **static verification complete, runtime gates
complete, lint blocker cleared.** Every runtime number below is read from a real
`.exit` file. Any gate without an `.exit` file is reported as not run, never as
passing.

**Headline for a reviewer in a hurry:**

- The decisive AC clause is **resolved**: focused non-race atomicity **66 s**
  and three race repetitions **84 / 76 / 75 s**, all exit 0, against a 480 s
  bar. Same-session baseline non-race atomicity is **306 s**.
- The lint blocker from the previous review is **cleared**.
  `golangci-lint run ./internal/transaction/...` exits **0**, `0 issues.`
  Full-repo `golangci-lint run` exits **1** with exactly one issue — the
  inherited `ineffassign` in `internal/godriver`, on a file byte-identical to
  the never-edited baseline twin. **Zero introduced findings.** See §10.
- The rework needed **four** renames, not the three the constraint named. The
  previous full-lint report was truncated by golangci-lint's default
  `max-same-issues: 3`. This is the one place this cycle went beyond its literal
  instruction; the reasoning is stated in §10 and `evidence/rework-lint.md` §1.

---

## 1. Baseline, explicitly recorded

| | |
| --- | --- |
| Source | `.temp/TASK-260729-rfrdfo/worktree` (the accepted rfrdfo prototype state) |
| Copy method | `rsync -a --exclude=.git --exclude=.temp --exclude=.task-board` |
| Prototype tree | `.temp/TASK-260729-365r5r/worktree` |
| Never-edited twin | `.temp/TASK-260729-365r5r/worktree-baseline` |
| Equivalence tree | `.temp/TASK-260729-365r5r/equivcheck` (baseline product + the behavioral tests) |
| Pre-manifest | `evidence/manifest-pre.txt`, 391 files, path-sorted SHA-256 |
| Post-manifest | `evidence/manifest-post.txt`, 392 files |

rfrdfo was chosen over the pristine `TASK-260720-jrrgw9` candidate because the
480 s bar is quoted against rfrdfo's own gate timings (atomicity race
593 / 561 / 564 s), and rfrdfo's delta is test-only under `internal/install`,
disjoint from the `internal/transaction` product edit made here.

### Immutability, re-verified this cycle

```
diff manifest-source-rfrdfo-pre.txt  manifest-source-rfrdfo-post.txt    exit 0
diff manifest-source-jrrgw9-pre.txt  manifest-source-jrrgw9-post.txt    exit 0
diff manifest-baseline-twin.txt      manifest-baseline-twin-post.txt    exit 0
```

The rfrdfo source, the main jrrgw9 candidate, and the baseline twin are all
byte-identical to their recorded pre state.

---

## 2. The literal delta

`evidence/manifest-pre-post.diff` — **1 modified, 1 added, 0 deleted**:

```
< 997d53df…  ./internal/transaction/namespace.go        (baseline)
> bb332038…  ./internal/transaction/namespace.go        (prototype)
> 3611f04f…  ./internal/transaction/namespace_pass_test.go   (new)
```

The test-file hash moved from `eec83a0a…` to `3611f04f…` in the
RUN-260729-d36102 lint rework (four parameter renames, §10).
`namespace.go` is `bb332038…` before and after — no product code moved in that
cycle. Pre-rework evidence is preserved verbatim as
`evidence/manifest-post-prerework.txt`,
`evidence/manifest-pre-post-prerework.diff` and
`evidence/prototype-prerework.patch` (still hashing `441b7677…`, the value the
previous review verdict recorded).

Product: `internal/transaction/namespace.go` only. `engine.go`, `journal.go`
and `staging.go` are byte-identical to baseline — `saveJournal`,
`validateJournal`, `loadJournal`, `buildJournal` and `type Engine` are
untouched. Full function-level allowlist: `evidence/allowlist.md`.

`evidence/prototype.patch` (`3dbbbfbd…`) applies to a pristine copy of the
baseline with **exit 0, no fuzz, no offset, no reject**, reproduces both files
byte-identically (`diff -q` exit 0 each), and leaves the baseline tree
unmutated (manifest diff exit 0 after the run). `git apply --check -p1` also
exits 0, both against a pristine copy and against `worktree-baseline` in place.

---

## 3. What the change actually is

The pairwise independence sweep still visits **every** O(P²) pair. What changed
is that each *path* is now resolved once per pass instead of once per pair.

`resolvedNamespacePath` carries, per declared path:

- `volume` / `parts` — the key split, previously recomputed by
  `namespaceComponents` for both sides of every pair;
- `volumeNFD` / `partsNFD` — the same split NFD-normalized per component,
  previously applied inside `namespaceComponentEqual` on every comparison;
- `identityRead` / `identityInfo` / `identityErr` — the one filesystem answer
  for "what object does this path name", previously an `os.Stat`/`os.Lstat`
  per surviving pair.

Two deliberate design points a reviewer should check:

1. **The identity read stays lazy.** `resolveNamespacePath` does *not* touch
   the filesystem. The baseline only asks for identity once a pair survives the
   containment test, so an eager read would surface an inspection failure for a
   path a containment overlap would have rejected first. Laziness preserves
   error precedence exactly.
2. **Case folding stayed where it was.** `normInsensitive` selects between the
   raw and NFD splits at comparison time and `caseInsensitive` still folds
   inside `namespaceComponentEqual`, because both are properties of the *pair*
   (`left.x || right.x`), not of the path. Only the per-component NFD transform
   — a pure function of one component — was hoisted.

### The one real tradeoff, stated plainly

Within a single pass the prototype is **strictly less likely** to notice a
filesystem mutation that lands mid-pass. Baseline re-read path A for every pair
A took part in, so a mutation between pair (A,B) and pair (A,C) could be caught
by the later pair. The prototype answers (A,C) from the snapshot taken at
(A,B), so that particular window closes.

This is a real reduction in mid-pass detection opportunities and a reviewer
should weigh it rather than wave it through. The argument that it is acceptable:

- Neither version is atomic. The baseline's pairwise re-reads happen at
  arbitrary, unsynchronized times; catching a mid-pass mutation was luck, not a
  guarantee, and the sweep has no lock over the target filesystem in either
  version.
- The window is bounded by one pass. Every `saveJournal` runs two fresh passes,
  and there are 23 call sites, so a graph is re-validated against live
  filesystem facts continuously through a transaction.
- The property that actually matters — a mutation landing *between* saves must
  fail closed before anything is written — is preserved and asserted
  behaviorally by
  `TestSaveJournalRejectsNamespaceAliasIntroducedBetweenSaves` and
  `TestRecoverRejectsDecodedTargetNamespacesAliasedWhileStopped`, both of which
  additionally check that the on-disk journal / live target was not mutated.

If the reviewer does not accept that tradeoff, the prototype should be rejected
on this ground specifically — it is inherent to per-pass reuse, not a bug that
can be patched out while keeping the O(P) property.

### There is no cross-save state

This was the explicitly rejected design from RUN-260729-bd5fd3, and it is
absent:

```
grep -rn 'namespaceGraphAccepted|acceptNamespaceGraph|forgetNamespaceGraph|
          namespaceChecked|namespaceGraph|namespaceMu' internal/     → 0 matches
grep -n '^var' internal/transaction/namespace.go                     → 0 matches
sed -n '/^type Engine struct/,/^}/p' internal/transaction/engine.go
    → mu, journalRoot, hooks, syncStagedParent   (4 fields, byte-identical)
```

Every `resolvedNamespacePath` lives only in the `paths` slice that
`validateIndependentTargetNamespaces` allocates locally. It is never returned,
never stored on `Engine`, and never escapes the call. When the function returns
the whole snapshot is garbage, so **every `saveJournal` call re-resolves and
re-reads the live filesystem from scratch.**

---

## 4. Fail-closed before mutation — static call path

Full proof with re-runnable greps: `evidence/call-path-proof.md`.

`saveJournal` calls `engine.validateJournal(journal)` as its **first
statement** (`journal.go:72`), so `ensureJournalRoot`, `CreateTemp`, `Write`,
`Sync`, `durableReplaceFile` and `durableRenameNoReplace` are all unreachable
for a graph that failed validation. The 23 call sites — `engine.go` (16) and
`staging.go` (7) — cannot bypass it, because the check lives in the callee.

All four ways a target graph enters the engine pass through it:

| origin of graph | entry | validated at |
| --- | --- | --- |
| new, built from a caller `Plan` | `Prepare` → `buildJournal` | `engine.go:234` |
| resumed by id | `Commit` → `loadJournal` | `journal.go:176` |
| recovered by journal-root sweep | `Recover` → `loadJournal` | `journal.go:176` |
| externally decoded bytes | any `loadJournal` | `journal.go:176`, after `Decode`, **before** the canonical-bytes check |

Both namespace sweeps (`journal.go:344` bare, `journal.go:354` with the
manager's journal root reserved) remain **unconditional**. No call site was
made conditional, skippable, or cached. Folding the two sweeps into one staged
pass is a further ~2×, deliberately left out: it moves error attribution
between two distinct journal errors and belongs to a separate change.

---

## 5. O(P) filesystem reads — measured

`namespaceIdentity` (the only `os.Stat`/`os.Lstat` in the sweep) has exactly
one caller, behind the `identityRead` guard:

```
grep -rn 'namespaceIdentity(' internal/
  namespace.go:81   the single call, inside (*resolvedNamespacePath).identity
  namespace.go:257  the definition
```

Benchmark `BenchmarkValidateIndependentTargetNamespaces`, one full validation
pass over a **disjoint** graph — the worst case, because a valid graph is
exactly the one where no pair short-circuits. Declared paths P = targets×7 + 1.

| targets | P | baseline | prototype | speedup |
| ---: | ---: | ---: | ---: | ---: |
| 8 | 57 | 11.00 ms | 1.82 ms | 6.1× |
| 16 | 113 | 42.27 ms | 4.24 ms | 10.0× |
| 32 | 225 | 162.63 ms | 10.57 ms | 15.4× |
| 64 | 449 | 638.27 ms | 31.06 ms | 20.6× |

The allocation counts are the cleaner proof, because they isolate the
resolution work from scheduler noise:

| P doubling | baseline allocs | prototype allocs |
| --- | ---: | ---: |
| 57 → 113 | ×3.47 | ×1.99 |
| 113 → 225 | ×3.70 | ×1.99 |
| 225 → 449 | ×3.84 | ×2.00 |

Baseline approaches ×4 per doubling — **quadratic**. The prototype is ×2.00 —
**linear**. Per-path allocations confirm it: baseline grows 362 → 2267 per
path, prototype is flat at 95.7 → 96.4.

The residual super-linearity in the *time* column is the in-memory pairwise
comparison, which is O(P²) by design and unchanged. The AC asks for O(P)
filesystem identity/resolution reads per pass, not an O(P) sweep.

The remaining per-pass filesystem work was already O(P) and is untouched: one
`canonicalNamespaceTargetPath` (`EvalSymlinks` walk), one
`namespaceCaseInsensitive` and one `namespaceNormalizationInsensitive` per
declared path, all in the build loop, none inside the pairwise sweep.

---

## 6. Negative coverage — `internal/transaction/namespace_pass_test.go` (new)

| test | cases | asserts |
| --- | ---: | --- |
| `…RejectsMalformedPaths` | 9 | relative / NUL / invalid-UTF-8 across live, staged, backup, rollback, entry-kind and reserved; all must fail with `path is not valid absolute filesystem text` |
| `…RejectsOverlappingPaths` | 7 | nested, exactly repeated, live-collides-with-another-target's-backup, live-collides-with-a-cleanup-tomb, **hard-link alias**, **symlink alias on the live parent**, target reaching into the reserved journal namespace |
| `…AcceptsDisjointPaths` | 1 (24 targets) | guards against becoming vacuously strict |
| `TestSaveJournalRejectsNamespaceAliasIntroducedBetweenSaves` | 2 | the between-save case: first save accepts a disjoint graph, filesystem then aliases two paths (hard link / symlink on parent), second save must fail with `errInvalidJournal` **and leave the on-disk journal byte-identical** |
| `TestRecoverRejectsDecodedTargetNamespacesAliasedWhileStopped` | 1 | externally decoded graph: valid when written, aliased while stopped, `Recover` must reject **and leave the live target unmutated** |
| `TestNamespaceIdentityIsReadOnceWithinOneValidationPass` | 1 | white-box: second read within a pass reuses the snapshot |
| `TestNamespaceIdentitySnapshotDoesNotOutliveItsPass` | 1 | white-box: a *new* pass sees the current filesystem |

The last two tests are the pair that pins the whole safety argument: reuse
*within* a pass, no reuse *across* passes. The two `saveJournal`/`Recover`
tests assert the same contract behaviorally, and additionally assert that the
rejection happened **before** mutation by comparing on-disk bytes.

No existing test file was edited. No assertion was weakened, deleted, skipped,
or retimed. No `-timeout` token, no journal schema change, no CI or conformance
fixture touched.

---

## 7. Evidence defects found and corrected this cycle

Static review of the inherited evidence found two defects in
`evidence/call-path-proof.md`, both now fixed:

1. **Dangling citation.** §6 cited `evidence/equivalence-check.md` as
   behavioral confirmation that the prototype neither adds nor removes a
   rejection. **That file does not exist and the check had never been run.**
   The `equivcheck/` tree itself is correctly built — its `namespace.go` is
   byte-identical to `worktree-baseline` (`diff -q` exit 0), and its test file
   drops exactly the two white-box tests that name prototype-only symbols and
   therefore cannot compile against baseline. §6 now says *prepared but not
   run*, and the check is wired as `gate-equivalence` in `bin/run-gates.sh`.
2. **Wrong grep annotation.** §1 annotated a `wc -l` as "23 call sites + 1
   definition" (24), but the command returns **25** on the prototype tree: 23
   call sites, the definition, and one prose occurrence in the new
   `resolvedNamespacePath` doc comment. Corrected, with explicit per-file
   call-site line numbers so the arithmetic is checkable.

Neither defect changes a conclusion, but both would have failed an independent
reviewer's re-run.

### Driver hardening

`run()` in `bin/run-gates.sh` previously wrote `99` and marched on to the next
gate if the process barrier was busy, so a foreign Go process merely *between*
two of its own gates could poison a whole run. It now retries barrier
acquisition for up to ~40 min before recording a refusal. This changes only
*when* a gate starts, never what its exit code means.

---

## 8. Gates

Protocol: every gate is a standalone process — no `tee`, no pipe chain — behind
a two-scan process barrier; `<name>.exit` is written **last**, so a missing
`.exit` means "killed or still running", never "passed".

### Inherited from RUN-260729-313095 (archived, `gates-partial-RUN-260729-313095/`)

| gate | exit | wall |
| --- | ---: | ---: |
| `gate-gofmt` | 0 | 0 s |
| `gate-vet` | 0 | 0 s |
| `gate-transaction` | 0 | 14 s |
| `gate-race-transaction` | 0 | 18 s |
| `gate-bench-baseline` | 0 | 6 s |
| `gate-bench-prototype` | 0 | 6 s |
| `gate-atomicity-structure` | **99** | — |

`99` is this driver's **barrier-refusal** code, not a test result. It means the
gate never ran. It is reported here as a non-result, not as a failure of the
code and certainly not as a pass.

### Prototype tree — complete run, `gates/`

The chain did what it was built to do: `CHAIN-START` 19:38:37 → foreign
`DRIVER-DONE` observed after 560 s → `DRIVER-START` 19:48:08 → `DRIVER-DONE`
19:55:44. Every gate took `BARRIER_OK` on two scans before starting, so no
timing below was measured under contention with `TASK-260729-2afulh`.

| gate | exit | wall | log |
| --- | ---: | ---: | --- |
| `gate-gofmt` | 0 | 0 s | empty — nothing unformatted |
| `gate-vet` | 0 | 0 s | empty |
| `gate-build` | 0 | 1 s | empty — `go build ./...` |
| `gate-lint` | **127** | 0 s | `golangci-lint: command not found` — see §10 |
| `gate-transaction` | 0 | 14 s | `ok … 13.319s` |
| `gate-race-transaction` | 0 | 18 s | `ok … 18.419s` |
| `gate-namespace-verbose` | 0 | 1 s | 25 PASS, **0 SKIP** |
| `gate-equivalence` | 0 | 1 s | `ok … 0.625s` (baseline product code) |
| `gate-bench-baseline` | 0 | 8 s | §5 left column |
| `gate-bench-prototype` | 0 | 6 s | §5 right column |
| `gate-atomicity-structure` | 0 | **66 s** | `ok … 66.238s` |
| `gate-race-atomicity-1` | 0 | **84 s** | `ok … 83.776s` |
| `gate-race-atomicity-2` | 0 | **76 s** | `ok … 76.016s` |
| `gate-race-atomicity-3` | 0 | **75 s** | `ok … 75.204s` |
| `gate-race-install-1` | 0 | 72 s | `ok … 70.145s` |

`gate-namespace-verbose` exists specifically to stop the alias coverage from
being vacuous: `namespace_pass_test.go` carries 5 `t.Skipf` capability guards
for hard links and symlinks. **None fired.** `grep -c SKIP` on the log returns
0, and the hard-link and symlink-on-parent subtests are visible as `PASS`. The
negative coverage in §6 is real, not skipped.

### Re-entry cycle RUN-260729-3819ca — one command, `gates/gate-lint-abs.*`

| gate | exit | wall |
| --- | ---: | ---: |
| `gate-lint-abs` | **1** | 3 s |

`evidence/lint-gate.md`. Barrier proven empty first (`SCAN 1 empty` /
`SCAN 2 empty` / `BARRIER_OK`, exit 0), plus a `ps -ax` at 20:04:17 showing no
driver alive.

### Lint rework cycle RUN-260729-d36102 — `gates/gate-rw-*` and `gates/gate-rw2-*`

Barrier proven empty before any Go command (`gates/gate-rw-preflight.barrier`,
exit 0) plus an independent `ps -ax` sweep at 20:19:40 finding zero Go or driver
processes. Every gate additionally took its own two-scan barrier.

Round 1, after the three renames the constraint named:

| gate | exit | wall |
| --- | ---: | ---: |
| `gate-rw-gofmt` | 0 | 0 s |
| `gate-rw-transaction` | 0 | 15 s |
| `gate-rw-race-transaction` | 0 | 18 s |
| `gate-rw-namespace-verbose` | 0 | 1 s |
| `gate-rw-lint-transaction` | **1** | 1 s |

Round 1's lint gate is **RED** and is reported as failing. It is kept because it
is the evidence that a fourth `unused-parameter` finding existed — hidden from
the previous cycle by `max-same-issues: 3` — and was found by running the
command rather than assumed away.

Round 2, after the fourth rename — the accepted set:

| gate | exit | wall | log |
| --- | ---: | ---: | --- |
| `gate-rw2-gofmt` | 0 | 0 s | empty |
| `gate-rw2-transaction` | 0 | 14 s | `ok … 13.662s` |
| `gate-rw2-race-transaction` | 0 | 19 s | `ok … 18.486s` |
| `gate-rw2-namespace-verbose` | 0 | 0 s | `ok … 0.413s`, 25 PASS, **0 SKIP** |
| `gate-rw2-lint-transaction` | **0** | 2 s | `0 issues.` |
| `gate-rw2-lint-full` | **1** | 0 s | 1 inherited `ineffassign`, §10 |
| `gate-rw2-equivalence` | **0** | 1 s | `ok … 0.741s`, baseline product code |

`gate-rw2-namespace-verbose` still reports **25 PASS and 0 SKIP**, identical to
the pre-rework count, so no capability guard started firing and no subtest was
silently neutered by the rename.

`gate-rw2-equivalence` was run under an explicit orchestrator directive received
at a checkpoint: apply the same renames to the adapted
`equivcheck/internal/transaction/namespace_pass_test.go` (`26dd7405…` →
`c86e3fbb…`, same four cases) and re-run that one gate behind the barrier.
`equivcheck`'s product `namespace.go` stays at the baseline `997d53df…`, so the
gate still compares prototype behavior against baseline product code.

Not re-run in this cycle, per the constraint: atomicity, install and benchmark
gates. `evidence/rework-lint.md` §4 gives the mechanical argument for why a
`_test.go` parameter rename cannot stale them — chiefly that `namespace.go` is
bit-identical across the cycle and that a `_test.go` file in
`package transaction` is not an input to any `internal/install` build.

### Baseline twin — `gates-baseline/`, partial and stated as partial

| gate | exit | wall |
| --- | ---: | ---: |
| `gate-transaction` | 0 | 15 s |
| `gate-atomicity-structure` | 0 | **306 s** |
| `gate-race-atomicity-*` | — | **not measured** |

The baseline driver was killed by orchestrator directive after the non-race
atomicity gate, and a 20:01:45 attempt to complete the race half was cancelled
before its first gate produced an `.exit`. Those artifacts are invalid and
explicitly excluded — `evidence/excluded-artifacts.md`.
`gates-baseline/DRIVER-DONE` is absent and is not claimed.

---

## 9. The atomicity margin — the decisive AC clause

The AC asks for focused non-race **and** race evidence of a defensible margin at
or below 480 s, or an explicit rejection. Both halves exist, both are green:

| measurement | prototype | bar | headroom |
| --- | ---: | ---: | ---: |
| non-race `internal/install/atomicity` | 66 s | 480 s | 7.3× |
| race, worst of 3 | 84 s | 480 s | **5.7×**, 396 s under |
| race, best of 3 | 75 s | 480 s | 6.4× |
| race `internal/install` | 72 s | — | — |

Same-session A/B on the non-race gate, the only axis where the never-edited
twin has a number:

| | baseline twin | prototype | ratio |
| --- | ---: | ---: | ---: |
| non-race atomicity | 306 s | 66 s | **4.6×** |

Cross-session context, for scale only: `TASK-260729-rfrdfo` recorded race
atomicity at 593 / 561 / 564 s on this host. The prototype's 75–84 s is the
same package under the same barrier protocol.

### The honest caveats on this number

1. **No same-session baseline race figure exists.** The race-to-race ratio is
   therefore *not* established on this machine in this session. What is
   established is (a) the prototype's absolute race times, which is what the
   480 s bar is actually stated against, and (b) a 4.6× same-session reduction
   on the non-race axis.
2. **The 306 s baseline is itself below 480 s.** The bar was quoted against
   rfrdfo's 561–593 s race timings, so a reviewer comparing 66 s to 306 s should
   not conclude the baseline was failing the bar on the non-race axis — it was
   not. The clause the prototype moves decisively is the race axis, where the
   only baseline reference is rfrdfo's cross-session 561–593 s.
3. `internal/install/atomicity` timing is not solely namespace validation. The
   4.6× is the end-to-end effect, consistent with §5's 20× on the isolated
   sweep, but the two numbers measure different things.

**Verdict on this clause: satisfied, not rejected.** The margin is defensible
and large. It is not the reason to withhold integration — §10 is.

---

## 10. The lint blocker, and how it was cleared

The previous cycle handed off with `golangci-lint run` at exit **1** and four
issues. Three were this prototype's. RUN-260729-d36102 fixed them — and found a
fourth of the same kind. Full detail: `evidence/rework-lint.md`.

### Final state

| command | exit | result |
| --- | ---: | --- |
| `golangci-lint run ./internal/transaction/...` | **0** | `0 issues.` |
| `golangci-lint run` (full repo) | **1** | 1 issue, inherited |

The one remaining issue:

```
internal/godriver/builddriver_positive_conformance_test.go:178:4:
    ineffectual assignment to environment (ineffassign)
```

**Inherited, not this prototype's.** The file is byte-identical to the
never-edited twin (`diff -q` exit 0) and absent from `manifest-pre-post.diff`.
The rfrdfo baseline lints red on it too. The rework constraint explicitly
permits retaining it; acceptance is stated as **zero introduced/transaction
findings**, which is met.

### The three-versus-four discrepancy, stated plainly

The rework constraint named **three** closures. **Four** needed the fix.

`.golangci.yml` sets neither `max-same-issues` nor `max-issues-per-linter`, so
golangci-lint's defaults apply — `max-same-issues: 3`. The previous cycle's
full-repo report was therefore **truncated at exactly three** `unused-parameter`
findings. After fixing those three, the scoped run went red on a fourth, at
line 144, the `"live path is another target's cleanup tomb"` case.

It was fixed, because the same constraint's acceptance bar is *zero introduced
findings* and honouring an enumeration derived from a capped report would have
missed that bar and guaranteed a fourth red cycle. This is the one place this
cycle exceeded its literal instruction, and a reviewer may overrule it. The
remaining three closures (lines 152, 165, 177) genuinely use `t` and keep the
named parameter; the scoped run reporting **1** issue under a cap of 3 proves
that count was complete rather than truncated a second time.

### What the reviewer should decide

1. **The mid-pass detection tradeoff (§3).** Inherent to per-pass reuse. If it
   is unacceptable, reject on that ground — it cannot be patched out while
   keeping the O(P) property. **This is now the only open design question.**
2. **Whether the missing same-session baseline race figure matters** for
   accepting the margin, given the prototype's absolute race times are 5.7×
   under the bar.
3. **Whether the fourth rename was in scope.** Mechanically identical to the
   three named ones, zero behavioral effect, argued above.

Independent review is required before integration, per the AC. This document is
a handoff, not an acceptance.

### What RUN-260729-d36102 changed

Inside the manifested prototype worktree: exactly one file, eight lines, four
parameter renames — verified after the gates ran:

```
$ shasum -a 256 internal/transaction/namespace.go \
                internal/transaction/namespace_pass_test.go
bb332038…  internal/transaction/namespace.go          == manifest-post.txt (UNCHANGED)
3611f04f…  internal/transaction/namespace_pass_test.go == manifest-post.txt
```

Across 392 manifested files this cycle moved **one** hash. Outside the manifest,
one further file changed: the adapted test copy
`equivcheck/internal/transaction/namespace_pass_test.go` (`26dd7405…` →
`c86e3fbb…`) carries **the same four closure-parameter renames**
(`t *testing.T` → `_ *testing.T`, lines 70/78/87/95 — the identical four cases),
applied under the explicit orchestrator directive recorded in §8, and
`gate-rw2-equivalence` **exited 0** (1 s, `ok … 0.741s`) after that adaptation.
`equivcheck`'s product `namespace.go` stayed at the baseline `997d53df…`, so the
gate still compares prototype behavior against baseline product code.

Otherwise: no product code, no journal schema, no timeout value, no CI, no
protocol behavior, no assertion, no main candidate. Nothing committed, staged or
published. The rfrdfo source, the jrrgw9 candidate and the baseline twin all
re-manifest byte-identical to their recorded state (`diff` exit 0, three trees).
