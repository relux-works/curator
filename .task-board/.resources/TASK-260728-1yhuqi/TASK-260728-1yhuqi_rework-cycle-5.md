# TASK-260728-1yhuqi — rework cycle 5

Closes the single blocking finding of
`TASK-260728-1yhuqi_review-verdict-cycle-5.md` (reviewer run `RUN-260729-36adf9`)
and the two binding-model defects that came with it.

Nothing accepted in cycles 1–4 is reopened: the SwiftPM rejection, the
direct-`swiftc` graph/compile model, the line-1-only manifest classifier, the
byte-exact banner grammar, the toolchain/SDK/native admission rules,
`curator-swift-module-v1`, `curator-swift-relpath-v1`, the closed per-job token
grammar, resolved containment, the byte-total source rule
`curator-swift-source-admission-v1`, the plugin policy, `manager-worker-v2`
cardinality at one graph plus one compile command, and every platform claim are
unchanged. No execution-policy identity, claim schema or capability-evidence
version is minted.

---

## The finding

Reference 4.1.4 and decision section 4 required the manager to re-resolve and
re-identify every binding immediately before `compile_argv`. The probe
implemented that re-check and exercised it in `S20`/`S21`. **The session did
not call it.** It ran `AdmitSources` → `swiftc -###` → `VerifyPlan` →
`swiftc compile_argv`, with nothing between the last two steps.

That is worth stating plainly rather than softening: the contract claimed a
defence, the code had the defence, and the thing that actually runs skipped it.
No check went red, because no check asserted the step. The reviewer found it by
reading the session.

## What closed it

**The permit is now a step in the normative session order**, in both documents
and in the probe:

```text
1. AdmitSources            Stage B, before any process
2. swiftc -###             manager command 1 of 2
3. VerifyPlan              the closed grammar over every token
4. the compile permit      re-bind everything  ← this
5. swiftc compile_argv     manager command 2 of 2
```

`compile_argv` is reachable only through step 4. The session records that the
step ran and how many bindings it re-examined, so "it was called" is measured
rather than inferred. Any finding is
`build_execution_control_unavailable` / `swift_permit_binding_changed`, leaves
**exactly one** manager-started command, and produces no artifact.

### Repair 1 — a binding re-checks the path it actually identified

The reviewer's second point was that the function could not be integrated
unchanged. `checkBucket` resolved and identified an output's **parent** when the
output did not exist yet, then stored the still-absent final path as the thing to
re-resolve. `Reverify` called `EvalSymlinks` on it, which fails while the session
is behaving correctly.

A binding now carries `raw` (the plan token) and `checked` (the path whose
identity was established) as two fields. They differ in exactly one case, and the
distinction is normative, not an implementation note — writing it the other way
rejects every happy path, and the only ways out of that are to weaken the check
or to skip it.

Measured on the default vector: **5** of the 33 plan bindings are absent outputs
re-checked at their parent. Expected-red control `C17` restores the raw-token
model and reports **5** findings on a verified happy path where the live permit
reports **0**.

### Repair 2 — absence means `ENOENT`

The absent-plugin branch treated any `Lstat` error as continued absence. A
permission error, an I/O error or a dangling symlink means the manager cannot
establish the state of a path it is about to hand to a compiler. That is now a
finding. `C17` reports the retired branch raising **0** findings on an
unreadable absent plugin path where the live permit raises **1**.

### Repair 3 — the Stage-B rule leaves a binding

File identity is size, mode and mtime, all of which a writer can restore. Stage B
already reads every byte, so it now records the **digest** of what it scanned,
and the permit re-reads every admitted source and compares. A unit test writes
one byte in place, restores the mtime, asserts the file identity is byte-equal to
what was recorded, and requires the finding anyway — and requires it to name
`swift_source_macro_selector_forbidden`, because the byte it wrote was `@`.

Without this the source rule would be a statement about a file the manager no
longer has.

---

## Integrated cases, not only unit checks

Seven new structural checks run **whole sessions**. Each adversarial one mutates
real state after step 3 has already accepted the entire plan — the only window
the graph phase leaves — and each is required to name its own mechanism, so a
case cannot pass by producing some unrelated finding.

| ID | Mutation between verification and permit | Measured |
|---|---|---|
| `S65` | none | permit ran over **35** bindings (33 plan + 2 sources), 0 findings, 2 commands, artifact produced |
| `S66` | an absent plugin path is created | 2 findings, **1** command, no artifact |
| `S67` | an admitted source gains a `@` | 3 findings, **1** command, no artifact; names `swift_source_macro_selector_forbidden` |
| `S68` | an admitted source is replaced by rename | 3 findings, **1** command, no artifact; names the digest change |
| `S69` | the presented SDK is re-pointed | 4 findings, **1** command, no artifact |
| `S70` | the recorded output parent is replaced | 1 finding, **1** command, no artifact |
| `S71` | a bound executable's bytes change | 1 finding, **1** command, no artifact |
| `S72` | — | live vs retired re-check on the same bindings, at the permit: 0 vs 5 and 1 vs 0 |

Every adversarial case runs its own **control** session first, over the same
copied source set, and requires it to reach 2 commands. A case whose control run
is already rejected proves nothing about the permit.

`S71` is the one case that cannot use the real toolchain: its executables are
read-only host state this task does not mutate. It runs a full session against a
**synthetic** manager-owned root — a `swiftc` that prints a fixed plan under
`-###` and writes its output otherwise, plus the executables that plan names. The
plan is verified by the same closed grammar and the same bucket rules as the real
one; only the executables behind it are writable. It is labelled as synthetic in
the probe, in the reference and here, and it makes no claim about Swift.

Two things the probe learned while building this, both recorded rather than
worked around:

- The synthetic shim must use shell builtins only. The session runs its children
  with an **empty** `PATH` — that is the contract — so a `cat` in the shim does
  not resolve and the plan arrives empty. The first run measured exactly that.
- `S72` and `C17` measure **at** the permit, through a read-only `observe` hook,
  not after the session. Measured: the graph command's own intermediate directory
  does not survive the compile command, so a post-session re-check reports a
  difference that has nothing to do with either function. The first version of
  `S72` did that and was wrong.

---

## Expected-red controls

| ID | Restores | Reports |
|---|---|---|
| `C16` | the retired cycle-4 session with no permit step | the same appearing plugin path then reaches a compile permit and **2** manager commands |
| `C17` | the retired binding model | **5** vs **0** findings on a verified happy path; **0** vs **1** on an absent plugin path that cannot be stat'ed |

Both restore code from the same binary, and both report numbers rather than
restating the review.

---

## Document changes

**Reference** — 3.3 gains the digest binding and why the rule needs one; 4.1.4 is
rewritten as the normative session order, the binding fields, the four permit
conditions, the single failure outcome and the measured table; 4.2.2 states that
"after the permit" is a precondition; section 9 names both Swift details beneath
`build_execution_control_unavailable`; 7.6 adds `C16`/`C17`; 11 adds the residual
below; 13 adds the `CP01`–`CP08` vector group.

**Decision** — section 4's binding bullet is rewritten with the session order,
the `raw`/`checked` distinction, the digest, the `ENOENT` rule and the measured
results; section 8 states the precondition; section 9's matrix gains three rows;
the stable-failure-class list names both Swift details; section 14 adds the
residual; three rejected alternatives are added — exercising the permit only in
isolation, re-checking the raw token, and treating any error as absence.

## Residual added

**The permit narrows the mutation window; it does not remove it.** It runs
immediately before `compile_argv`, so a write landing between the permit and the
compile child's own `open` is outside it. That interval is bounded by the
ownership requirement on the declaration channels, not by a check. Closing it
needs a compiler that accepts content-addressed inputs, which this toolchain does
not offer. Stated in reference 11 and decision 14.

---

## Results

| Gate | Result |
|---|---|
| probe `gofmt -l` / `go vet` / `go test -count=1` / `go build` | exit 0, 0, 0, 0 |
| native run | 23 cases / 23 matched / 0 divergences; 32 closure checks / 0 verdicts; **17 of 17** controls red; **70** structural checks / 0 divergences; executed P2 admission ok; `green: true`; exit 0 |
| degraded run (no resolvable toolchain) | 23 `not_run` with the reason recorded, exit 0, nothing installed |
| each control replayed individually | `C1`–`C17`, every one exit 1 |
| tarball round-trip | 22 non-test / 10 test files, 69 test functions, 9683 lines, `go test` exit 0; sha256 `4465da8689f68031ef7c1908369c6cdf44aa8e18ebdc3f6db5de9dbfef37f5b9` |

Counts against cycle 4: structural **62 → 70**, controls **15 → 17**, non-test Go
files **21 → 22**, test files **9 → 10**, test functions **63 → 69**, lines
**8662 → 9683**. Cases and closure checks are unchanged at 23 and 32.

### Expected-red gates, attributed

- **curator repo `gofmt -l .`** — exit 2, 1141 paths. **0** under
  `.temp/TASK-260728-1yhuqi`, **0** outside `.temp/`, **0** modified tracked
  files. The paths are other tasks' scratch trees; the count grew since cycle 4
  because more of them exist, not because this task added any.
- **spec `validate.py` in the task worktree** — exit 1, failing only its link
  check on `docs/external-build-repositories.md`. The clean baseline at `57c1f56`
  exits 0 (30 schemas, 93 vector files). Scoped over the two documents this task
  authored: `docs/swift-build-drivers.md` 4 links / 0 broken,
  `decisions/0011-swift-driver-pair.md` 2 links / 0 broken.

### Hygiene

Spec worktree: 0 staged, 0 tracked modifications, and both copies are byte-equal
to the board artifacts. Curator repo: 0 tracked modifications. Nothing staged,
committed, pinned, published or installed on any host. No platform claim widened:
macOS arm64 remains the only measured tuple, Windows remains an implementation
contract with no claim, Linux remains deferred.

---

## Open items carried forward

Unchanged, and none affected by this rework:

1. the Windows plan-derived closure member **count** is unmeasured — an
   implementation takes it from the plan it verifies;
2. the beyond-ordinal-0 SDK argument template for a multi-root platform is
   unmeasured — minting it is part of the Windows obligation;
3. the compile phase re-derives its own plan, so "the plan verified is the plan
   executed" is an equality of inputs rather than of processes. Closing it needs
   an execution-policy identity that admits a manager-driven job set, which is a
   decision 0008 change.

New, and stated rather than closed:

4. the permit closes the graph-to-permit interval, not the permit-to-`open`
   interval.
