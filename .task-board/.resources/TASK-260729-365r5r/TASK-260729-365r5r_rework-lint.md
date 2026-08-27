# RUN-260729-d36102 — lint-only rework

Scope: `TASK-260729-365r5r_lint-rework-constraint.md`, answering the single
blocking finding in `TASK-260729-365r5r_review-verdict.md`.

**Result: the prototype is lint-clean.** `golangci-lint run ./internal/transaction/...`
exits **0** with `0 issues.` Full-repo `golangci-lint run` exits **1** with
exactly **one** issue, the inherited `ineffassign` in
`internal/godriver/builddriver_positive_conformance_test.go`, on a file that is
byte-identical to the never-edited baseline twin. **Zero introduced findings.**

---

## 1. The edit, and the one place it went beyond the constraint's literal list

The constraint named three cases. **Four** were needed. This is the one
deviation in this cycle and it is stated up front rather than buried.

`gate-lint-abs` (previous cycle, full-repo run) reported three
`revive unused-parameter` findings, at lines 119, 127 and 136. That report was
**truncated**. `.golangci.yml` sets neither `max-same-issues` nor
`max-issues-per-linter`, so golangci-lint's defaults apply:

```
max-same-issues: 3          <-- this is why exactly three were shown
max-issues-per-linter: 50
```

After renaming those three, the scoped run `gate-rw-lint-transaction` exited
**1** and surfaced a fourth instance of the identical finding:

```
internal/transaction/namespace_pass_test.go:144:16:
    unused-parameter: parameter 't' seems to be unused (revive)
```

Line 144 is the `"live path is another target's cleanup tomb"` case — the fourth
of the seven `build` closures in
`TestValidateIndependentTargetNamespacesRejectsOverlappingPaths`, immediately
adjacent to the three named ones and identical in kind.

It was fixed. The reasoning, so a reviewer can overrule it:

- The constraint's own acceptance bar is *"acceptance requires zero
  introduced/transaction lint findings"*. Stopping at three would have violated
  the acceptance bar in order to honour an enumeration that was itself derived
  from a capped report.
- The enumeration was descriptive of what the linter printed, not a deliberate
  exclusion of a fourth case the constraint author had seen and chosen to keep.
- The fix is the same mechanical rename, in the same table, in the same file,
  with the same zero behavioral effect.
- Stopping short would have guaranteed a fourth red handoff cycle.

**The remaining three closures were checked and are correct as-is.** Lines 152,
165 and 177 genuinely use `t` (`mustWrite`, `mustMkdirAll`, `t.Skipf`), so they
keep the named parameter. The scoped lint run reporting exactly **1** issue
while the cap allows 3 proves that count was complete, not truncated a second
time.

### The complete delta of this cycle

```
$ diff -u <pre-rework namespace_pass_test.go> worktree/internal/transaction/namespace_pass_test.go
-			build: func(t *testing.T, root string) ([]TargetRecord, []targetNamespacePath) {   x4
+			build: func(_ *testing.T, root string) ([]TargetRecord, []targetNamespacePath) {   x4
```

Eight changed lines, four pairs, at the `nested live paths`, `repeated live
path`, `live path is another target's backup sidecar` and `live path is another
target's cleanup tomb` cases. Nothing else. The pre-rework file was
reconstructed from `evidence/prototype-prerework.patch` and independently
hashes to `eec83a0a…`, the value the reviewer recorded, so the diff above is
against the exact tree the reviewer read.

**`namespace.go` was not touched.** It hashes `bb332038…` before and after —
identical to `manifest-post-prerework.txt`. No product code moved in this cycle.

---

## 2. Manifest and patch evidence

Pre-rework artifacts are preserved verbatim, not overwritten:

| file | role | sha256 |
| --- | --- | --- |
| `manifest-post-prerework.txt` | reviewer-read post state, 392 files | — |
| `manifest-pre-post-prerework.diff` | reviewer-read delta | — |
| `prototype-prerework.patch` | reviewer-read patch | `441b7677…` (matches the verdict) |

That `prototype-prerework.patch` still hashes to the value in the review verdict
is the check that this cycle regenerated evidence with the *same* method the
reviewer validated, not a different one.

Regenerated:

| file | value |
| --- | --- |
| `prototype.patch` | `3dbbbfbd06d586442bd08166142934763b217efae42f85506d9c6a258c4c50d2` |
| `manifest-post.txt` | 392 files, path-sorted |
| `manifest-pre-post.diff` | 1 modified, 1 added, 0 deleted |

```
359c359
< 997d53df…  ./internal/transaction/namespace.go          (baseline)
---
> bb332038…  ./internal/transaction/namespace.go          (prototype, UNCHANGED by this cycle)
362a363
> 3611f04f…  ./internal/transaction/namespace_pass_test.go  (new; was eec83a0a… pre-rework)
```

Across 392 files, this cycle changed **exactly one hash**.

### Patch verification, real exit codes

```
git apply --check -p1 evidence/prototype.patch   (pristine baseline copy)   0
git apply --check -p1 evidence/prototype.patch   (worktree-baseline)        0
patch -p1 < evidence/prototype.patch             (pristine baseline copy)   0
```

No fuzz, no offset, zero `.rej`/`.orig` files. The applied result is
byte-identical to the prototype tree (`diff -q` exit 0 for both files).

### Nothing else moved

```
diff manifest-source-rfrdfo-post.txt  manifest-source-rfrdfo-rework.txt    0
diff manifest-source-jrrgw9-post.txt  manifest-source-jrrgw9-rework.txt    0
diff manifest-baseline-twin.txt       manifest-baseline-twin-rework.txt    0
```

The rfrdfo source, the main jrrgw9 candidate and the never-edited twin are all
byte-identical to their recorded state. Nothing was committed, staged or
published.

---

## 3. Gates — real exit codes

Barrier proven empty **before** any Go command
(`gates/gate-rw-preflight.barrier`: `SCAN 1 empty` / `SCAN 2 empty` /
`BARRIER_OK`, exit 0), plus an independent `ps -ax` sweep at 20:19:40 matching
zero `run-gates`, `wait-then-run`, `golangci`, `go test` or `go build`
processes. Every gate below additionally took its own two-scan barrier through
`bin/run-one.sh` — standalone process, no `tee`, no pipe chain, `.exit` written
last.

### Round 1 — after the three named renames, `gates/gate-rw-*`

| gate | exit | wall |
| --- | ---: | ---: |
| `gate-rw-gofmt` | 0 | 0 s |
| `gate-rw-transaction` | 0 | 15 s |
| `gate-rw-race-transaction` | 0 | 18 s |
| `gate-rw-namespace-verbose` | 0 | 1 s |
| `gate-rw-lint-transaction` | **1** | 1 s |

Round 1 is retained deliberately. Its **red** lint gate is the evidence that the
fourth finding existed and was discovered by running the command, not assumed
away. It is reported as failing.

### Round 2 — after the fourth rename, the accepted set, `gates/gate-rw2-*`

| gate | exit | wall | log |
| --- | ---: | ---: | --- |
| `gate-rw2-gofmt` | **0** | 0 s | empty — both changed paths formatted |
| `gate-rw2-transaction` | **0** | 14 s | `ok … 13.662s` |
| `gate-rw2-race-transaction` | **0** | 19 s | `ok … 18.486s` |
| `gate-rw2-namespace-verbose` | **0** | 0 s | `ok … 0.413s`, 25 PASS, **0 SKIP** |
| `gate-rw2-lint-transaction` | **0** | 2 s | `0 issues.` |
| `gate-rw2-lint-full` | **1** | 0 s | 1 inherited `ineffassign` |
| `gate-rw2-equivalence` | **0** | 1 s | `ok … 0.741s`, baseline product code |

`gate-rw2-namespace-verbose` still shows **25 PASS and 0 SKIP** — identical to
the pre-rework count. The five `t.Skipf` hard-link/symlink capability guards did
not fire, so the alias coverage remains real and the rename did not silently
neuter a subtest.

### Why `gate-rw2-lint-full` is red and that is the accepted state

```
internal/godriver/builddriver_positive_conformance_test.go:178:4:
    ineffectual assignment to environment (ineffassign)
1 issues:
* ineffassign: 1
```

One issue. Inherited, proven statically:

```
$ diff -q worktree/internal/godriver/builddriver_positive_conformance_test.go \
          worktree-baseline/internal/godriver/builddriver_positive_conformance_test.go
$ echo $?
0
```

The file is byte-identical to the never-edited twin and does not appear in
`manifest-pre-post.diff`. It is not this prototype's regression and this
prototype does not fix it — the constraint explicitly permits retaining it.

Running full lint was one command beyond the constraint's enumerated list. It
was run because the constraint's acceptance bar is stated *against full lint*,
and because the previous cycle's full-lint report had already been shown to be
truncated by `max-same-issues: 3`. Asserting "only the inherited ineffassign
remains" from that truncated report would have repeated the exact error this
cycle exists to fix. The command is read-only, finished in under a second on a
warm build cache, touched no product code, and ran behind the same barrier.

---

## 4. Why the preserved performance evidence is not staled

The accepted atomicity and benchmark numbers — non-race
`internal/install/atomicity` **66 s**, race **84 / 76 / 75 s**, race
`internal/install` **72 s**, and both benchmark gates — were **not** re-run, per
the constraint. They remain valid, and the argument is mechanical rather than a
judgement call:

1. **No product code changed.** `internal/transaction/namespace.go` hashes
   `bb332038…` before and after this cycle. Every one of those gates measures
   product code paths, and the product tree is bit-for-bit what produced them.
2. **A parameter name is not observable.** Renaming a function parameter to `_`
   changes no type, no signature arity, no call site, no control flow and no
   emitted behavior. The parameter was, by the linter's own finding, unread in
   all four closures.
3. **The atomicity and install gates do not compile the changed file.**
   `namespace_pass_test.go` is in `package transaction`. `go test
   ./internal/install/atomicity` and `./internal/install` build their own
   packages against the non-test `internal/transaction` archive; a `_test.go`
   file in a dependency package is never part of a dependent package's build.
   The changed file is not an input to those gates at all.
4. **The gates that *do* compile it were all re-run** — `gate-rw2-transaction`,
   `gate-rw2-race-transaction`, `gate-rw2-namespace-verbose` — and are green.
5. **`gate-equivalence` was re-run rather than argued about.** An orchestrator
   directive received at a checkpoint required the same renames in the adapted
   copy `equivcheck/internal/transaction/namespace_pass_test.go` (`26dd7405…` →
   `c86e3fbb…`, four closures at lines 70/78/87/95, the identical four cases;
   lines 103/116/128 use `t` and were left alone) followed by exactly one
   barriered re-run. `gate-rw2-equivalence` exits **0** in 1 s
   (`ok … 0.741s`). `equivcheck`'s product `namespace.go` is untouched at the
   baseline `997d53df…`, so the gate still measures prototype behavior against
   baseline product code.

So the decisive AC clause is unchanged: worst-case race atomicity **84 s**
against the **480 s** bar, **396 s** of headroom, from a tree whose product code
is identical to the one now handed off.

---

## 5. What this cycle did not do

- Did not run atomicity, install, benchmark or baseline-race gates.
- Did not start a detached process or any baseline script.
- Did not edit `namespace.go` (either copy), journal schema, timeout values, CI,
  protocol/spec behavior, or the main candidate. `equivcheck`'s **test** file
  was edited only under the explicit orchestrator directive recorded in §3/§4,
  with the identical four renames; its product code stayed at baseline
  `997d53df…`.
- Did not weaken or remove an assertion — the four renames are inside table
  constructors that build fixtures, not inside any assertion.
- Did not commit, stage or publish anything.

Cancelled baseline-race artifacts from earlier runs remain excluded per
`evidence/excluded-artifacts.md`; this cycle produced no new excluded debris.
