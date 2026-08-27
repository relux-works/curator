## 11. Cycle-7 rebuild — what the `references/info.md` trim actually removes, and what it measured

Cycle 6 answered "can §4.3 supply the missing margin?" with **No**, on arithmetic the cycle-6 verdict
correctly rejected as invalid. This section does two things:

1. Replaces that arithmetic with the model the source actually implements (§11.1–§11.2).
2. Reports the **measurement**, which already exists. A sibling task, **`TASK-260729-2afulh`**
   (`prototype-atomicity-fixture-trim`, parent `STORY-260720-3plyvy`, status **`done`**), applied
   exactly this trim, instrumented the staging inventory before and after, and ran the focused race
   gate. Its result and its own accepted review verdict are in §11.4.

The short version: **cycle 6's conclusion — the trim does not reach the 480s bar — survives
measurement, but its arithmetic was wrong in the direction that understated the trim.** The measured
staging-work reduction is **14.5–17.1 %**, not 7.1–12.5 %, and the measured race result is
**492.231s** — 12.231s above the 480s bar, and **107.8s below the 600s alarm** where Patch B alone
had 8.72s.

### 11.1 The save-count model, read off the source

Three call sites in `stageTarget` (`internal/transaction/staging.go:16-61`) fire **once per staging
entry**, not once per target:

```go
for target.StagingIndex < len(target.StagingEntries) {   // staging.go:18-19
        …
        engine.saveJournal(journal)                      // staging.go:26   before create
        engine.createStagingEntry(journal, targetIndex, entry)
        engine.saveJournal(journal)                      // staging.go:33   after create
        if entry.Kind == "file" { engine.copyStagingFile(journal, targetIndex, entry) }
        …
        engine.saveJournal(journal)                      // staging.go:56   after advance
}
```

`copyStagingFile` (`staging.go:118-176`) adds **two more per 32 KiB chunk actually read** — `:141`
write-ahead, `:161` chunk-commit — both inside `if count > 0`. A zero-byte file therefore adds
**nothing**: its first `Read` returns `0, io.EOF` and the branch never runs.

One staged target `t` therefore costs

> **`3 · E_t + 2 · C_t`** journal saves,

where `E_t = len(target.StagingEntries)` and `C_t = Σ ⌈size / 32 KiB⌉` over the *file* entries of `t`.

`StagingEntries` is produced by `captureRemovalEntries(target.Kind, source, desiredDigest)`
(`engine.go:201`). For a directory target that is a `filepath.Walk` of the staged source
(`journal.go:507-533`) and it records **the walk root itself** (its relative path is normalised to
`""` at `journal.go:518-522`), **every subdirectory**, and **every regular file**. For a `KindEntry`
symlink it is `captureLinkRemovalEntries` (`journal.go:553-…`) — exactly one `link` entry, three
saves.

Every other `saveJournal` call site is per-phase or per-target, never per-file:

| Site | Frequency |
| --- | --- |
| `engine.go:64`, `:105` | once each per `Prepare` |
| `engine.go:325`, `:339` | once each per `commit` |
| `engine.go:359`, `:443` | ≤ 2 per target that reaches backup / install |
| `engine.go:332`, `:401`, `:413`, `:538`, `:547`, `:588`, `:658` | failure, rollback and cleanup phases |
| `engine.go:772`, `:823` | 2 per removed sidecar — **independent of `len(RemovalEntries)`** (`removeRecordedSidecar` / `finishRecordedRemoval`) |
| `staging.go:218`, `:244` | preparation discard |

So the total per installation is

> **`saves(install) = O(phases) + O(targets) + Σ_t (3 · E_t + 2 · C_t)`**

and the fixture trim can only move the last term. It does **not** move `N`, the target count, so it
does not touch the `O((7N)²)` pairwise work *inside* each surviving save (§2.2). That part of cycle
6's reasoning is correct and survives.

**Cycle 6's `3 × targets + 5 × files` formula is withdrawn.** It charged three saves per *target*
where the source charges three per *entry*; it treated files as the only entries, where the source
also walks directories and the root; and it therefore had no term at all for the directory entry the
trim removes. `TASK-260729-2afulh`'s reviewer independently confirmed the corrected formula:
*"source inspection confirms `stageTarget` calls `saveJournal` three times per staging entry and
`copyStagingFile` calls it twice per copied chunk, so `3*entries + 2*chunks` reproduces every row."*

### 11.2 What the trim removes, exactly — derived, then measured

**Derived from source.** `stageNode` (`internal/install/targets.go:39-90`) builds a context target
with `whitelist.CopyContext` (`targets.go:66`) and then `marker.Write` (`targets.go:85`):

- `whitelist.IncludeRoots` (`whitelist.go:20-23`) is
  `SKILL.md, agents, references, .skill_triggers, assets, templates, examples, data`. The fixture
  skill (`atomicity/fixture_test.go:75-100`) supplies exactly two of them: **`SKILL.md`** and
  **`references/`**.
- **`csk-skill.json` never enters context** — it is not in `IncludeRoots`. Cycle 6's §11.1 said a
  context target carries all four fixture files. It does not.
- **`scripts/` never enters context** either: `Commands` is non-empty (`fixture_test.go:84-92`), so
  `includeScripts` is false (`targets.go:59-62`), and `runtime_roots: ["scripts"]` also places it in
  `excludeRoots` (`targets.go:65`, `whitelist.go:106-120`).
- `marker.Write` adds exactly one file, `.csk-install.json` (`marker.go:27`).
- `locale.Render` (`targets.go:70`) adds **no** file for this fixture: the snapshot has no `locales/`
  tree, so `analysis.LocaleToRender == ""` and it returns before writing anything
  (`locale.go:114`, `:123-125`).

| Context target | Staging entries `E` | Non-empty file chunks `C` | Staging saves `3E + 2C` |
| --- | ---: | ---: | ---: |
| **Now** — `""`, `SKILL.md`, `references`, `references/info.md`, `.csk-install.json` | 5 | 3 | **21** |
| **After the trim** — `""`, `SKILL.md`, `.csk-install.json` | 3 | 2 | **13** |
| **Delta** | **−2** | **−1** | **−8** |

That is the verdict's `3 × 2 entries + 2 × 1 file chunk = 8`, confirmed. The `references/` directory
disappears together with its only child because the fixture creates it only implicitly, through
`e.write`'s `os.MkdirAll` of the parent (`fixture_test.go:63-73`), and `CopyContext` skips an include
root whose `os.Stat` fails (`whitelist.go:57-59`). **−8 saves is 38.1 % of that target's own staging
saves.**

**Measured, independently.** `TASK-260729-2afulh` instrumented both arms with a measurement-only test
file (`zz_measure_test.go`, byte-identical in both arms, absent from its deliverable) and dumped the
staging manifest of a context target. From `evidence/measurement-raw.txt`, verbatim:

```
before: MANIFEST scenario=project entries=directory:,directory:references,file:.csk-install.json,file:SKILL.md,file:references/info.md
after:  MANIFEST scenario=project entries=directory:,file:.csk-install.json,file:SKILL.md
```

**That is the derived 5-entry and 3-entry trees exactly**, including the absence of `csk-skill.json`
and `scripts/` and the presence of the walk root as its own `directory:` entry. The static derivation
above is confirmed element by element, not merely in total.

**How many context targets are affected — derived, then measured.** Cycle 6 counted three, once. The
trim applies to *every* context target of *every* installation. Deriving from the scenarios
(`commit_atomicity_test.go:41-93`, `:131-215`) and then checking against the measured per-phase save
deltas:

| Scenario / phase | Context targets (derived) | Measured entries | Measured chunks | Measured staging saves | Δ saves | = 8 × contexts? |
| --- | ---: | ---: | ---: | ---: | ---: | :---: |
| project, baseline (`skill-a`, `skill-h`) | 2 | 24 → 20 | 14 → 12 | 100 → 84 | **−16** | ✔ 8×2 |
| project, upgrade (`skill-a`, `skill-b`, `skill-h2`) | 3 | 34 → 28 | 19 → 16 | 140 → 116 | **−24** | ✔ 8×3 |
| global, baseline (`skill-a`, `skill-drop`) | 2 | 26 → 22 | 16 → 14 | 110 → 94 | **−16** | ✔ 8×2 |
| global, upgrade (`skill-a`, `skill-b`) | 2 | 25 → 21 | 16 → 14 | 107 → 91 | **−16** | ✔ 8×2 |

Every row agrees with the derivation. The **measured per-installation reduction in staging-path
journal saves is 14.55 %–17.14 %**; `TASK-260729-2afulh`'s reviewer quotes the same span as
"12.5–17.6 %" over a slightly different set of shapes.

Aggregating over one patched chain (`baseline → injection → final`, where the injection and final
installations both have the upgrade shape — this step is *derived*, not separately measured):

| Chain | Staging saves before | after | Δ | Reduction |
| --- | ---: | ---: | ---: | ---: |
| project-hybrid | 100 + 140 + 140 = **380** | 84 + 116 + 116 = **316** | −64 | **16.84 %** |
| global | 110 + 107 + 107 = **324** | 94 + 91 + 91 = **276** | −48 | **14.81 %** |

Removal-class targets are excluded throughout: their `StagedSource` is empty (`engine.go:182`), so
`StagedPath` is empty, so `Prepare` skips `stageTarget` for them (`engine.go:75-77`), and
`removeRecordedSidecar` saves twice regardless of how many entries the removal walk found.

### 11.3 How wrong cycle 6 was, and in which direction

| Quantity | Cycle 6 (withdrawn) | Cycle 7 derived | `TASK-260729-2afulh` measured | Direction of the cycle-6 error |
| --- | ---: | ---: | ---: | --- |
| saves removed, project upgrade installation | 15 | 24 | **24** | understated ×1.6 |
| staging-save reduction per installation | 7.1–12.5 % | — | **14.55–17.14 %** | understated |
| files in a fixture context target | 4 | 3 | **3** (`SKILL.md`, `references/info.md`, `.csk-install.json`) | overstated |
| entries removed per context target | 3 files | 1 file + 1 directory | **1 file + 1 directory** | wrong kind |

Cycle 6's ceiling was below the requirement it stated (14.4–18.8 %); the measured reduction is
**inside** that band. Its conclusion happened to hold, but not for the reason it gave: the trim fails
on **wall clock by 12.2s**, not because its work reduction is too small.

### 11.4 The measured result — `TASK-260729-2afulh`

Scope: the accepted 13-file rfrdfo state plus exactly one 14th path,
`internal/install/atomicity/fixture_test.go`. Integrity checker over freshly regenerated manifests:
`modified=14 added=0 deleted=0 unexpected=0 forbidden_touched=0`, `INTEGRITY_OK`, independently
re-run by its reviewer. Assertion, test, subtest, `t.Parallel`, `t.TempDir`, `t.Setenv`, skip and
timeout-token counts are all **unchanged**; the only mechanical change is `e.write(` 8 → 7.

| Gate | Real exit | Wall | Against |
| --- | ---: | ---: | --- |
| `gate-gofmt` | 0 | 0s | |
| `gate-build` | 0 | 1s | |
| `gate-vet` | 0 | 0s | |
| `gate-atomicity-nonrace` (`-count=1`) | 0 | 273s (`ok … 272.580s`) | control 285.434s ⇒ **−4.50 %** |
| `gate-race-atomicity-1` (`-count=1 -race`) | **0** | 493s (`ok … 492.231s`) | controls 560.828 / 564.022 / 591.280s ⇒ **−12.23 % / −12.73 % / −16.75 %** |
| `gate-race-atomicity-2` | **no exit file** | — | driver received SIGTERM mid-run; partial log correctly excluded from evidence |
| `gate-race-atomicity-3` | **not run** | — | stopped by orchestrator directive once the strict predicate was already false |

No `DATA RACE` marker in either race log. Every executed gate carries its own two-scan `BARRIER_OK`.
`DRIVER-DONE` was written manually at a cooperative stop and says so.

**Reading the number honestly.**

- **It misses §7's 480s bar by 12.231s** (2.5 %). That is the disqualifying fact and it is why
  `TASK-260729-2afulh`'s reviewer rejected the fixture-only patch as the timeout solution.
- **It clears the 600s alarm by 107.769s**, where Patch B alone cleared it by **8.72s**. That is a
  materially different risk posture, and §7's bar exists precisely to absorb the focused-versus-`./...`
  gap that 8.72s cannot survive and ~108s might.
- **It is one repetition.** §10.3 measured a 30.45s (5.4 %) run-to-run spread on this package, so a
  single 492.231s observation sits in a band of roughly **465–520s**. It neither demonstrates the bar
  is cleared nor demonstrates it cannot be. Repetitions 2 and 3 do not exist.
- The saving is **larger under `-race` (12.2–16.8 %) than without it (4.5 %)**, which is what §2.2
  predicts: the removed work is in-process Go that the race detector instruments, not syscall time.

### 11.5 Disposition of §4.3

**Measured, and not promoted.** The reasons are now evidential rather than arithmetical:

1. Its one valid focused race gate is **492.231s**, above §7's 480s pass condition.
2. It costs a **14th file**, widening the §4.0 allowlist and touching
   `atomicity/fixture_test.go`, which §4.0 explicitly excludes.
3. `TASK-260729-2afulh` is `done` with an **accepted** review verdict that reproduces and adjudicates
   exactly this, and routes onward to the production fix.

**What is withdrawn from cycle 6, and stays withdrawn:** the `3 × targets + 5 × files` model, the
15-save numerator, the 120–210 denominator, the 7.1–12.5 % ceiling, the 490.7–517.4 s projection
table, and the claim that the trim is *bounded below* the required saving. Measurement puts the
reduction at 14.55–17.14 %, inside the band cycle 6 said was required.

**What should not be claimed either way:** that one repetition at 492.231s proves the trim cannot
reach 480s. If `TASK-260729-365r5r` is rejected or lands short, the cheapest remaining evidence is
**repetitions 2 and 3 of the gate `TASK-260729-2afulh` already built**, on the tree it already
produced — not a new prototype:

```sh
CURATOR_CONFORMANCE_ROOT=<immutable rc.5 root> GOTMPDIR=<task-owned> \
  go test -count=1 -race ./internal/install/atomicity/
```

run twice more in `.temp/TASK-260729-2afulh/worktree`, each as a standalone process behind the §7
two-scan barrier, each exit code recorded, no `tee`, no `-timeout`, no `./...`. Promote the trim only
if **all three** repetitions land at or below 480s. Combining it with the production fix is the
better bet regardless: the two levers are independent — the trim reduces the *number* of saves, the
production fix reduces the *cost* of each one.

### 11.6 One further candidate, examined and rejected on protocol grounds

Per-save cost is dominated by **path canonicalisation**, and that cost is linear in path **depth**:
`canonicalNamespaceTargetPath` (`namespace.go:121`) calls `filepath.EvalSymlinks`, which walks one
`lstat` per component, and `namespaceCaseInsensitive` / `namespaceNormalizationInsensitive`
(`namespace_case_darwin.go:13`, `:25`) each walk `existingNamespaceAncestor` (`namespace.go:245`)
over the same path again. Every fixture path is rooted at `t.TempDir()`, which is rooted at
`GOTMPDIR`; the alarm frames quoted in §1.3 show live paths of **198–216 bytes**. Shortening that
root would reduce every save's cost without touching a single assertion.

**Rejected, and it should stay rejected.** `GOTMPDIR` is set by the *gate command*, not by the tests,
and the verifier protocol pins it to a task-owned directory whose post-run size and removal are
themselves integrity gates (verifier-3 ledger rows "Task-owned GOTMPDIR size" and "Exact GOTMPDIR
removal", both exit 0). A test that reached outside it for a shorter root would break that gate and
give up `testing.T`'s automatic cleanup and per-test isolation. It is recorded here as
examined-and-refused, not as available headroom.
