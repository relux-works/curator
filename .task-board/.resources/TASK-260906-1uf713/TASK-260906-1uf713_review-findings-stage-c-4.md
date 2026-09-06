# TASK-260906-1uf713 — review findings, stage (c) cycle 4 (rework 3)

Verdict: **CHANGES REQUESTED**. Two blocking, one major, two minor.

`repeat-of:` **cycle-3 C3-B2** — finding C4-B1 below is the same environments §1
`path`-source read discipline, in the same `switch` arm, reached through a different
addressing mode, and it is a **regression introduced by the C3-B2 fix itself**
(bisected to `7fcf9c1b`). It is the fifth consecutive instance of the class this chain
keeps naming — a defect living in an addressing mode nobody tested, under a doc comment
and a ledger row that both claim the coverage (stage (a) F1, cycle-1 B3, cycle-2 C2-B1,
cycle-3 C3-B2, now C4-B1).

`repeat-of:` **the orchestrator's Windows finding on `833918d2`** — finding C4-B2 is a
Windows-only lane failure that no Unix runner and no local run can see, the fourth
consecutive stage to ship one. It has been red since rework 2 (`221e0082`) and no review
cycle has read a settled Windows lane since `0bcea201`.

- Subject: `~/Developer/ReluxWorks/.worktrees/curator-stage-c`, branch
  `feat/agent-environments-stage-c` at `4a8a1d42db8ed4941dbca67bacc2efa16a9444c4`,
  25 commits past `origin/main`. `gh pr view 61 --json headRefOid` =
  `4a8a1d42db8ed4941dbca67bacc2efa16a9444c4`, so the hosted lanes run this exact tree.
  The producer worktree was read-only: `git status --short` empty and HEAD unchanged
  before and after.
- Authority: curator-spec `550579d` (this story worktree) — `protocol/environments.md`
  §1, §6, §8.4, §9.1, §9.2, §9.5, §9.6, §12.1, §12.2; `profiles/manager.md` §1;
  `cli/curator.md`.
- Method: a `curator` binary built from `4a8a1d42` (`go build ./cmd/curator`, exit 0)
  invoked as a standalone process against scratch homes with `CURATOR_CONFIG`,
  `CURATOR_SYSTEM_CONFIG`, `CLAUDE_CONFIG_DIR`, `CODEX_HOME`, `XDG_CONFIG_HOME`,
  `PI_CODING_AGENT_DIR`, `GIT_CONFIG_GLOBAL` and `HOME`/`USERPROFILE` pinned to the
  scratch tree. Comparison binaries were built from `4df4d507` and `7fcf9c1b` the same
  way. Every mutant was applied to a throwaway `git archive` copy (`/tmp/rev4/mut`),
  never to the worktree. Reproducer:
  `TASK-260906-1uf713_review-probes-stage-c-4.tgz`.

---

## The Change Request's empty repository delta

`CR-TASK-260906-1uf713-4` revision 4 reports `repository_delta: empty`: candidate tree
`4829dd34` is identical to base `550579d1` and the patch is zero bytes. **That is
structurally correct for this leaf and is not what the verdict turns on** — it was equally
true of revisions 1, 2 and 3, all of which were reviewed on their merits.

The story workspace is a *curator-spec* worktree and `550579d1` is the curator-spec
authority commit. This leaf's declared scope is the *curator* repository, branch
`feat/agent-environments-stage-c`, delivered on PR #61, and every producer brief in this
chain forbids writing into the control root. A non-empty delta here would itself have been
a scope violation. The reviewable work is `git diff origin/main..HEAD` in the curator
worktree (the stage) and `git diff 4df4d507..HEAD` (rework 3: 11 files, +579/−44), which
is what I reviewed.

So the emptiness is right. I am rejecting the revision for C4-B1, C4-B2 and C4-M1, which
live in the curator tree.

---

## Hosted lanes on the reviewed head

`gh pr checks 61` on `4a8a1d42`, run 34045070090, read after every lane settled:

| Lane | Result |
|---|---|
| Lint | pass 1m15s |
| Naming gate | pass 10s |
| Interop conformance gate | pass 21s |
| Gate self-test (ubuntu / macos / windows) | pass 6s / 9s / 27s |
| Test (ubuntu-latest) | pass 3m11s |
| Test (macos-latest) | pass 9m30s |
| Race (ubuntu-latest) | pass 7m28s |
| Race (macos-latest) | pass 19m2s |
| **Test (windows-latest)** | **FAIL 28m43s** → C4-B2 |
| Candidate suite | skipping (gated matrix — it does **not** run on this PR) |

**Not every hosted lane is green on the head I reviewed.** `Test (windows-latest)` is red,
and per this cycle's brief a red lane on this head is blocking. Details in C4-B2, extracted
from the run's `test-evidence-windows-latest` artifact rather than from `--log-failed`.

C4-B1 and C4-M1 are invisible to every lane, green or red, which is the standard this chain
has been held to since cycle 2.

---

## BLOCKING

### C4-B1 — `profile install <same path>` no longer re-reads the source, so the §1 reinstall is dead and reports success anyway

`repeat-of: cycle-3 C3-B2` (regression introduced by its fix, commit `7fcf9c1b`)

§1 is one sentence with two halves:

> Installation copies the directory's tree into the profile store as an immutable
> snapshot and never reads the source directory again: later edits to the source
> directory change nothing **until the operator reinstalls**, and nothing about a `path`
> source is ever fetched from a network.

Rework 3 fixed the first half and broke the second. `installLocked`
(`internal/envprofile/envprofile.go:633-645`) routes a same-name, same-source install
straight into `updateLocked`:

```go
if prior, err := readSource(home, name); err == nil {
    // Re-installing an installed source with the same requirement
    // re-resolves exactly as profile update does and is reported as an
    // update; a different source under the same name is taken.
    if prior == source {
        info, moved, err := updateLocked(op, home, name, options.Policy)
```

For a `git` source that delegation is right — the requirement is a range and a reinstall
*is* an update. For a `path` source it is now exactly wrong, because `updateLocked`'s
`KindPath` arm was changed by C3-B2 to resolve from the store snapshot and never from
`source.Path`. The `stateForPath` call eleven lines above has already read and stored the
edited tree; the delegation throws that snapshot away and re-pins the old one.

Driven through the production CLI:

```
$ curator profile install ./src --as pk            # pin ff6c35c7…, body "original"
$ printf 'EDITED AFTER INSTALL\n' > ./src/context/a.md
$ curator profile install ./src --as pk            -> exit 0
    updated profile pk (lock sha256:e67d89a7…)     # the SAME lock hash as before
$ # pin ff6c35c7… -> ff6c35c7…                     NO RE-READ
$ grep -c EDITED <home>/native/claude/CLAUDE.md    -> 0
$ curator profile install ./src --as pk2           # control, a fresh name
$ # pin 43d9f269…                                  # so the edit is real
```

Bisected to the commit:

| binary | reinstall of the same name over an edited source |
|---|---|
| `4df4d507` (pre-rework-3) | pin `ff6c35c7…` → `921fe8db…` — **re-read** |
| `7fcf9c1b` (the C3-B2 fix) | pin `ff6c35c7…` → `ff6c35c7…` — **no re-read** |
| `4a8a1d42` (this head) | pin `ff6c35c7…` → `ff6c35c7…` — **no re-read** |

Cycle 3 explicitly recorded the pre-fix behaviour as correct ("`profile install <same
path>` correctly does — that is the reinstall §1 sanctions"), and the cycle-4 brief asked
in terms for it to be confirmed. It was not.

**Consequences.** `cli/curator.md` at `550579d` defines exactly one row that reads a
`path` source — `profile install <git-url|path> …` — and §1 names the reinstall as the
only refresh. With `update`, `sync` and `use` all correctly frozen, there is now **no
operator path that refreshes a `path` profile in place**. The only way to pick up an edit
is to install under a different name, which changes the profile identity, or to
`profile remove` first, which requires switching the machine away from it. And the command
reports `updated profile pk` at exit 0: a false attestation of an update that did not
happen.

**The attestation half.** Two artifacts assert exactly the coverage this defect needs:

- `TestPathSnapshotImmutableAcrossUpdateSyncUse`'s doc comment says the test "extends it to
  sync, use, **reinstall**, and the imported-profile update-all shape". It does not: the
  "not vacuous" step is `Install(home, InstallOptions{Operand: source, As: "pk2"})` — a
  *fresh* install under a different name, which takes the branch above the
  `prior == source` delegation. The one addressing mode the fix could break is the one the
  test avoids.
- `.github/ci/platform-cases.tsv:329` states "edits to a path source after install move
  nothing across update, sync and use **until reinstall**". The clause after "until" is
  neither asserted by the test nor true of the code.

That is the C2-m3 "ledger row states what the test does not" shape and the cycle-1 M1
"comment claims a path it does not have" shape, together, directly over the hole.

**Fix direction and the test it needs.** Decide from §1 which operation is the reinstall
and stop routing a same-source `KindPath` install through the update arm — the freshly
computed `stateForPath` result is already in hand at that point. Then extend
`TestPathSnapshotImmutableAcrossUpdateSyncUse` (or add a sibling) to drive `Install` **under
the same name** over an edited source and assert the pin *moves*, alongside the existing
update/sync/use assertions that it does not. Correct the doc comment and ledger row 329 so
neither claims reinstall coverage the suite does not have. Narrowing mutant to apply and
quote: make the `prior == source` delegation exempt exactly `KindPath` and show the new
named test failing.

### C4-B2 — `Test (windows-latest)` is red on this head, and has been since rework 2

The platform-case gate fails on Windows with one case:

```
FAIL  ledger case skipped for the wrong reason on windows:
      internal/envprofile :: TestImportUnreadableMarkerIsLoss
platform-case gate: FAILED
```

From `test-evidence-windows-latest`, `test/skips-observed.tsv:33`:

```
internal/envprofile  TestImportUnreadableMarkerIsLoss  UNCLASSIFIED  FATAL-wrong-class
    this environment can read a mode-000 file; unreadability is untestable here
```

The ledger row is fine — `platform-cases.tsv:355` already carries
`linux,darwin | windows | host-capability`. What is missing is the **reason regex**.
`.github/ci/skip-classes.tsv` registers

```
host-capability   this environment can (inspect|read) a mode-000 directory   allow
```

and the skip introduced by rework 2 (`internal/envprofile/import_test.go:785`) says
**file**, not directory — correctly, because the artifact is `.agent-environment.json`, a
file. The three sibling skips (`pathkind_test.go:187`, `:328`, `import_test.go:250`) all
say "directory" and classify. One word, and the whole Windows lane is red.

Runs on this branch:

| run | head | Test (windows-latest) |
|---|---|---|
| 34037510726 | `0bcea201` | success (cycle 2's green head) |
| 34041073632 | `4df4d507` | **failure** — same case, same UNCLASSIFIED reason (I pulled its artifact and diffed) |
| 34045070090 | `4a8a1d42` | **failure** — same case |

So it has been red since rework 2 (`221e0082`, which added the test) and nothing caught it:
cycle 3 recorded while the lane was still running and honestly claimed nothing about it;
the rework-3 report honestly says CI was not consulted because the head was not pushed. No
local lane can see it — on a non-root Unix runner `chmod 0o000` works, the test does not
skip, and the classifier never sees the message. This is the mechanism the orchestrator's
`833918d2` finding described, one layer up: **a skip whose class is registered but whose
reason text is not.**

Fix: widen the registry pattern to cover the file case (or split it into its own row) and
re-push so the lane is read green. The gate is doing its job — it is refusing to let a new
skip reason through unclassified — so do not silence the case and do not change the skip to
say "directory" when the artifact is a file.

---

## MAJOR

### C4-M1 — the three refusals C3-B2 introduced are undriven, and the §8.4 one collapses into a silent no-op with a fully green suite

The new `KindPath` arm adds three refusals and no test drives any of them:

```go
return … "profile %q lock carries no state pin for its path root"        // (1)
return … "profile %q path snapshot cannot be read: %v"                   // (2)
return … "profile %q snapshot names %q, want lock root %q"               // (3)
```

`grep` finds no test asserting any of the three messages, and my mutants confirm it. Each
keeps the surrounding code and weakens exactly one member; each was run against the **full**
`./internal/envprofile/` and `./cmd/curator/` suites, not a `-run` subset:

| # | Mutant | Result |
|---|---|---|
| P2 | drop `\|\| rootMember.StateHash == ""` from (1) | **SURVIVES** |
| P3 | delete the snapshot-name check (3) | **SURVIVES** |
| P4 | on (2)'s read failure return the old lock as `unchanged`, `nil` error | **SURVIVES** |

P4 is the one that matters: it is the §8.4 shape verbatim — *a failed read of manager-owned
state turned into "nothing to do"* — the same class as `readSkillsLedger` (fixed in
`ee6743a2`) and `readRootSurface` (cycle-2 C2-B1). Driven end to end with the mutant built,
against a store entry removed under the pin:

```
stock      curator profile update pk -> exit 1
           pk: profile_source_invalid: profile "pk" path snapshot cannot be read:
               context_manifest_invalid: …
P4 mutant  curator profile update pk -> exit 0
           pk: unchanged
```

A corrupted or half-written store entry would make `profile update` report `unchanged`
forever while the profile is unrepairable, and the suite would not notice. Refusal (2) is
reachable without any code change — that is how I produced the stock line above.

The stage's own definition of done says "every new refusal is proven by a narrowing mutant
that kills a named test". Three of three are unproven. The class C3-B2 was *about* — falling
back to `source.Path` — **is** covered: my mutant P1 (below) dies on two named tests. This
finding is the rest of the arm.

---

## MINOR

- **C4-m1 — the §12.2 gate can be narrowed to admit exactly `default`, green.** The
  production behaviour is correct: I drove `curator profile use default` under a
  `require_current_profile: acme` lock and got exit 1 naming the knob. But
  `Policy.CheckMachineUse` narrowed with `|| name == DefaultProfile` **survives** the full
  `./internal/envprofile/` and `./cmd/curator/` suites. Every committed refusal row uses an
  ad-hoc name (`other`, `bogus`, `p5`, `other2`, `other3`); none uses `default`, which is
  the builtin, always-present profile and the first thing an operator under a lock would
  try. One more driven row closes it. (Coverage point, not a defect — recorded because this
  is the fourth cycle in which this gate's coverage has been the question.)
- **C4-m2 — `profile update` on a machine that already violates the lock leaves the lock
  moved and the surfaces stale, at exit 1.** `updateLocked` publishes the new lock and
  *then* calls `resyncCurrentScopes`, whose `useLocked` hits the §12.2 gate. Driven with a
  git root under a lock to `acme` while the machine current is `groot`:

  ```
  $ curator profile update groot   -> exit 1
      groot: environments.require_current_profile: profile use of "groot" is refused: …
  $ # lock commit c684b700… -> fcd5fc41…   (MOVED)   surfaces not re-materialized
  ```

  This is the generic resync-failure shape rather than anything new in rework 3 — the gate
  has been on this path since rework 1 and two cycles passed it — and the state is
  recoverable by switching to the required profile. Recorded with evidence rather than as a
  new defect; the producer may reasonably answer that a machine in a configuration-error
  state is expected to fail loudly. If so, say so in the report rather than leaving it
  undocumented.

---

## What I verified and found correct

**C3-B2, the immutable snapshot (the half the fix got right).** Install, edit the source,
then `update`, `sync` and `use`: the pin stays `ff6c35c7…` through all three and the
materialized `CLAUDE.md` still carries `original`, not the edit. With the source directory
deleted entirely, `update`, `sync` and `update --all` all exit 0. `profile update --all`
with a §9.6 imported profile present exits 0 — cycle 3's bricked command is fixed.
**Overlays still re-resolve under a `path` root**, which the rework-3 brief warned "return
the old lock" would break: with a `git` overlay declared through
`profile compose add … --range` and the overlay repository moved to a new tag, the lock
records `ovl 1.0.0 → 1.1.0` while the root stays pinned at `ff6c35c7…`, and the emitted
bytes carry `original` plus `overlay v2`. `curator gc` does not prune the context store, so
the snapshot the arm now depends on survives maintenance (driven: `gc`, then `update` with
the source deleted, exit 0). My narrowing mutant P1 — the store gate kept but weakened to
admit exactly a source directory that is still readable — is **killed** by
`TestPathSnapshotImmutableAcrossUpdateSyncUse` and `TestUpdatePathLeavesLockInPlace`.

**C3-B1, the structural seam.** `CurrentFile` has exactly one production writer,
`switch.go:248`, inside `useLocked` below the gate; `SetCurrent` (`envprofile.go:295`) is
exported but has no production caller (tests only), and `purgeHomes` does not touch the
pointer. `useLocked`'s only callers are `UseWithPolicy`, `installLocked` (which `Import`
also routes through) and `resyncCurrentScopes`. Every one driven through `run()` under a
lock to `acme`:

| caller | observed |
|---|---|
| `profile use other` | exit 1, names the knob, current unchanged |
| `profile use other --clear` | exit 2 usage, nothing written |
| `profile use other --clear --env claude_code` | exit 2 usage |
| `profile use --clear --env claude_code` (scoped clear) | exit 0, machine current unchanged |
| `profile use other --env claude_code` (scoped switch) | machine scope unaffected |
| `profile install ./other --as other2 --use` | exit 1, current unchanged |
| first-install auto-activation on a fresh machine | exit 1, no current written |
| `profile import --as other2` (auto-activation) | exit 1, no current written |
| `profile import --as other3 --use` | exit 1, no current written |
| `profile use default` | exit 1, names the knob |
| `profile use acme` (the requirement) | exit 0 |
| `profile update acme` (resync of the requirement) | exit 0 |
| `profile sync` (no switch) | exit 0 |

The producer's own mutant (`if scope == "" && !clearScope`) is **killed** by
`TestMachineClearReachesSeamGate`. I found no path that reaches the seam in machine scope
without passing the gate. The enumeration comment is now true.

**C3-M1, the transcribed pin.** `TestSystemV2LockableSubsetIsClassWide` transcribes §12.2's
six keys as a literal, asserts set equality with `LockableEnvKeys` in both directions, and
decides each knob's branch from the transcription rather than from the map. The
transcription matches §12.2 verbatim (`overlays_allowed`, `precedence`,
`mcp_package_allowlist`, `passable_env_names`, `require_current_profile`, `isolation`). Six
mutants, all killed hermetically **and** with `CURATOR_CONFORMANCE_ROOT`:

| # | Mutant | Result |
|---|---|---|
| W1 | `LockableEnvKeys += secret_material_waivers` (cycle 3's one-liner) | killed — `carries 7 keys, §12.2 transcribes 6` |
| W2 | `LockableEnvKeys += backup_retention` (mine) | killed — same |
| N1 | `LockableEnvKeys -= precedence` (mine) | killed — `carries 5 keys, §12.2 transcribes 6` |
| S1 | **count-preserving swap** `precedence → forms` (mine) | killed — `§12.2 key "environments.precedence" missing from LockableEnvKeys` |
| C1 | consumer `parseSystemEnvironments` admits exactly `secret_material_waivers` | killed — `non-lockable/secret_material_waivers admitted` |
| C2 | consumer admits exactly `targets` (mine) | killed — `non-lockable/targets admitted` |

S1 is the fifth mutant the brief asked for and the sharpest: it preserves the map's
cardinality, so only a real transcription catches it. The pin does not read
`LockableEnvKeys` for its expectations, directly or indirectly. **The gate is derived — this
class is closed.**

**C3-m1 and C3-m2.** `profile compose add` now warns `overlays_allowed is false: the added
declaration is inert; resolution joins the root alone` — driven under a *locked*
`overlays_allowed: false` and under an unlocked machine one, both exit 0 with the warning,
and `compose list` keeps its own. `env status` still reports
`require_current_profile: acme (locked)` in text and both JSON fields.
`TestWeightRulesAllFourDisagree` is committed and asserts weight 40, `overlay: true`,
`required_by: [mid1 mid2]` — cycle 3's drive exactly — over a `Policy` shape an operator can
actually produce (I built the same closure through `profile compose add … --range`). Note
for the record that it pins rule 4 over rules 1–3; 3-over-2 and 2-over-1 remain covered only
pairwise by `TestWeightRulesApplyInOrder`.

**Ledger rows 303 and 304** still describe what their tests assert, naming the manufactured
`Policy` and the TASK-260906-3x0w4y bound. Row 329 is the exception — see C4-B1.

**Windows sweep.** I re-swept the stage delta for the class the orchestrator flagged. The
two `filepath.Join(...) + "/"` constructions (`switch.go:491`, `managed.go:1665`) are
display-only notice strings, not path comparisons — harmless on Windows. The one new
`filepath.IsAbs` (`import.go:409`) takes `os.Readlink` output, a real OS path, not a POSIX
fixture literal. No new `$HOME`-relative lookup and no new well-known-location list. The
producer's "found nothing new" holds for the *code*; what it missed is the skip-reason
registration in C4-B2, which is not a code path and which only the hosted lane can show. The
standing `dotfileStateDirs` bound (TASK-260906-vlrjo1) is still stated in code and in the
report and its case still runs unskipped on all three runners.

---

## Observations (not findings)

- **A second, unpinned copy of the §12.2 set.** `config.LockableKeys`
  (`internal/config/config.go:54`) lists the same six `environments.*` keys and gates the
  system file's `locked` array at `config.go:313`; the transcription pin does not cover it.
  I checked exploitability rather than inferring it: with
  `LockableKeys += environments.secret_material_waivers` **built into a binary**, both
  reachable shapes are still refused — carrying the knob hits `parseSystemEnvironments`
  (`is not lockable and must not be carried by the system file`) and locking without
  carrying hits `checkLockedEnvSet` (`locks … but does not set it`). So the widening
  direction is **inert**; the mutant survives the suite but reaches nothing. Narrowing it is
  a loud refusal rather than a bypass, and two of the six narrowings
  (`mcp_package_allowlist`, `passable_env_names`) survive hermetically and die only with
  `CURATOR_CONFORMANCE_ROOT` — which matters because the Candidate suite is `skipping` on
  this PR. One line asserting `LockableKeys` against the same transcription would close it.
  Not a blocker.
- **`envprofile.SetCurrent` is an exported second writer of `CurrentFile` with no production
  caller.** Inert today (tests only), but it is a writer *below* the structural seam the
  C3-B1 fix rests on, and the single-writer claim in `switch.go:215` is true only because
  nobody calls it. A doc line or an unexported rename would keep the claim honest.
- **`CURATOR_SYSTEM_CONFIG` overrides the system-file path** (`config.go:227`), so any
  process can point every §12.2 lock at a file it controls. Pre-existing — `git log -S` puts
  it at `5e183d58` and it is not in the stage-c delta — so out of scope for this verdict,
  but stage (c) is what turns it into a fleet-policy bypass and it deserves a decision
  somewhere.

---

## My mutants

Applied to a throwaway `git archive` copy (`/tmp/rev4/mut`), never to the worktree. `config`
mutants were judged on `./internal/config/` with and without `CURATOR_CONFORMANCE_ROOT`;
`envprofile` mutants on the full `./internal/envprofile/` and `./cmd/curator/` suites.

| # | Target | Mutant | Result |
|---|---|---|---|
| W1 | `LockableEnvKeys` | += `secret_material_waivers` | killed |
| W2 | `LockableEnvKeys` | += `backup_retention` (mine) | killed |
| N1 | `LockableEnvKeys` | −= `precedence` (mine) | killed |
| S1 | `LockableEnvKeys` | `precedence` → `forms`, count preserved (mine) | killed |
| C1 | `parseSystemEnvironments` | admits exactly `secret_material_waivers` | killed |
| C2 | `parseSystemEnvironments` | admits exactly `targets` (mine) | killed |
| LK-W | `config.LockableKeys` | += `environments.secret_material_waivers` (mine) | SURVIVES — proven inert end to end, see Observations |
| LK-N | `config.LockableKeys` | −= `environments.precedence` (mine) | killed |
| LK-N2/3 | `config.LockableKeys` | −= `mcp_package_allowlist` / `passable_env_names` (mine) | survive hermetic, killed with the root |
| P1 | `updateLocked` `KindPath` | store gate admits exactly a still-readable source (mine) | killed — `TestPathSnapshotImmutableAcrossUpdateSyncUse`, `TestUpdatePathLeavesLockInPlace` |
| P2 | `updateLocked` `KindPath` | drop the empty-state-pin refusal (mine) | **SURVIVES** → C4-M1 |
| P3 | `updateLocked` `KindPath` | drop the snapshot-name refusal (mine) | **SURVIVES** → C4-M1 |
| P4 | `updateLocked` `KindPath` | failed snapshot read → silent `unchanged` (mine) | **SURVIVES**, driven end to end → C4-M1 |
| S1s | `useLocked` | `scope == "" && !clearScope` (the producer's) | killed — `TestMachineClearReachesSeamGate` |
| S3 | `Policy.CheckMachineUse` | admits exactly `DefaultProfile` (mine) | **SURVIVES** → C4-m1 |

C4-B1 and C4-B2 are not mutants: C4-B1 is a defect driven through the production CLI on the
unmutated tree and bisected to `7fcf9c1b`; C4-B2 is a red hosted lane read from its own
evidence artifact.

---

## Gate table (what I ran, as standalone processes, on `4a8a1d42`)

| Gate | Exit | Output |
|---|---|---|
| `go build ./...` | 0 | empty |
| `go vet ./...` | 0 | empty |
| `gofmt -l cmd internal` | 0 | empty |
| `golangci-lint run ./...` | 0 | `0 issues.` |
| `bash .github/ci/gate-selftest.sh` | 0 | `gate-selftest: 94 passed, 0 failed` |
| `bash .github/ci/no-broad-suppression.sh` | 0 | `no-broad-suppression: ok` |
| `go test -count=1 ./internal/config/` (throwaway copy, unmutated baseline) | 0 | `ok 0.498s` |
| `go test -count=1 ./internal/envprofile/ ./cmd/curator/` (throwaway copy, unmutated baseline) | 0 | the kill/survive oracle for P1–P4, S1s, S3 |
| `gh pr checks 61` | — | 11 of 12 pass, `Test (windows-latest)` **fail** — see C4-B2 |
| `gh api …/artifacts/9993282621/zip` (test-evidence-windows-latest, head `4a8a1d42`) | 0 | the failing case above |
| `gh api …/artifacts/9992198260/zip` (test-evidence-windows-latest, head `4df4d507`) | 0 | the identical failing case |

**Not run by me, stated plainly:** the two `test-gate.sh` lanes (materialized `SPEC_PIN`
root, and curator-spec main with `CI_REQUIRE_FULL_ROOT=1`),
`bash .github/ci/ledger-consistency.sh`, and `go test -timeout 30m ./cmd/curator` as an
unmutated full run in the worktree. Reasons: they cost roughly an hour of serialised wall
clock, must not run concurrently or beside a `-race` suite, the hosted lanes execute the
same suites on this exact head — and did, which is how C4-B2 surfaced — and the verdict is
CHANGES REQUESTED on defects none of them can see. The rework-3 report's two-root table for
`4a8a1d42` stands unchallenged and I found nothing that contradicts it. The ledger row I
dispute (329) is a wording defect, not a consistency one, so `ledger-consistency.sh` would
pass either way. **No `-race` suite and no `test-gate.sh` lane ran at any point during this
review**, so nothing here is contention noise.

---

## Judging the rework report against the attestation standard

The rework-3 report is accurate on everything I could check except two rows, and both are
findings above.

Correct: the finding→resolution table covers all six; the C3-B2 before/after and the
`update --all` drive reproduce exactly; the C3-B1 caller enumeration is true and every row
is driven through `run()`; the C3-M1 mutant table's kills are real and I added four more,
including a count-preserving swap it did not try; the carried bounds list is unchanged
rather than trimmed; the stale-oracle red (`TestUpdatePathMovesLock`) is reported rather
than buried, with the lesson drawn; the CI row honestly says the head was not pushed when
the report was written; and the "no `git add -A`" discipline held (`git status --short`
empty, no stray artefacts).

The two exceptions:

1. **The AC ratio.** The report claims **15 of 15** and says the path row's immutability gap
   "is now driven". Half of that row is driven. The row is "The `path` source kind
   **snapshots immutably**, pins by state hash, and rejects every `profile_source_invalid`
   condition", and §1's own sentence has a second clause — *until the operator reinstalls* —
   that the named test does not drive, that its doc comment claims it does drive, that
   ledger row 329 asserts, and that the code gets wrong. That is **14 of 15** with one row
   unchecked, and the unchecked half is where the regression is. Cycle-1 M1's class again: a
   row reported as driven that the production surface does not establish.
2. **The three new refusals.** They are not in the report's mutant table at all. The stage's
   definition of done requires a narrowing mutant per refusal; three shipped without one,
   and one of the three (P4) is the §8.4 class this stage has already paid for twice.

---

## Is stage (c) safe to land?

**No, not at `4a8a1d42`.**

C4-B1 removes the only operator path §1 provides for refreshing a `path` profile and reports
success while doing nothing. Every machine onboarded through §9.6 has a `path` profile, and
this is the command the operator reaches for when the source changes. It is a regression
this rework introduced, in the arm this rework rewrote, under a comment and a ledger row
that both say it is covered.

C4-B2 is a red required lane on the exact head, red since rework 2, and it is one word in
`skip-classes.tsv`.

Everything else I attacked held, and several things are now genuinely closed: the seam
enumeration is complete and every caller is driven, the transcribed §12.2 pin survives six
mutants including a count-preserving swap and no longer derives from the thing it protects,
the immutability half of C3-B2 is right and overlays still re-resolve beneath a `path` root,
the compose warning and the four-rule weight test are in, and the code-level Windows sweep
found nothing new. Both blockers are small: one branch plus one test, and one regex.

## Rework guidance

C4-B2 first — it is one line, and until it is green nobody can read the lane that would show
the next regression. C4-B1 next: decide from §1 which operation is the reinstall, stop
routing a same-source `KindPath` install through `updateLocked`, drive `Install` under the
**same** name over an edited source, and correct the doc comment and ledger row 329. C4-M1 is
three named tests and one mutant each — write P4's shape first, because it is the §8.4 class
this stage has now paid for twice. C4-m1 is one more row in an existing test; C4-m2 is a
paragraph in the report unless the producer decides it is a defect.
