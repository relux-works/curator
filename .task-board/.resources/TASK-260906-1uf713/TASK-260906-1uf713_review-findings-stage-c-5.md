# TASK-260906-1uf713 — review findings, stage (c) cycle 5 (rework 4)

Verdict: **CHANGES REQUESTED**. One major, four minor, three observations.

`repeat-of:` **cycle-1 M1 / cycle-2 C2-m2** — finding C5-M1 below is a doc comment on new
stage-(c) code that asserts a production path the code does not have, sitting directly over
a behavioural hole. The chain has named this class three times and the rework-2 brief said
in terms that "a comment that lies about reachability is how cycle-1 M1 happened."

`repeat-of:` **cycle-4 C4-B1** — C5-M1 is also the same attestation shape: a command that
reports success at exit 0 for work it did not do.

`repeat-of:` **cycle-4 C4-B2** — finding C5-m2 is two more stage-(c) skip reasons that the
registry does not classify, two source lines from the two that carry the classifying prefix.

- Subject: `~/Developer/ReluxWorks/.worktrees/curator-stage-c`, branch
  `feat/agent-environments-stage-c` at `71e6baecbb0bcd9e671b2fa5351b0e5ffdadda49`, 29 commits
  past `origin/main`. `gh pr view 61 --json headRefOid` = the same OID, so the hosted lanes
  run this exact tree. The producer worktree was read-only throughout:
  `git status --short` empty and HEAD unchanged before and after.
- Authority: curator-spec `550579d` (the task authority) and `87a0d006` (curator-spec main
  today) — `protocol/environments.md` §1, §6, §8.4, §9.1, §9.2, §9.5, §9.6, §12.1, §12.2;
  `profiles/manager.md` §1; `cli/curator.md`.
- Method: a `curator` binary built from `71e6baec` via `git archive` into `/tmp/rev5`
  (`go build ./cmd/curator`, exit 0), invoked as a standalone process against scratch homes
  with `CURATOR_CONFIG`, `CURATOR_SYSTEM_CONFIG`, `CLAUDE_CONFIG_DIR`, `CODEX_HOME`,
  `XDG_CONFIG_HOME`, `PI_CODING_AGENT_DIR`, `GIT_CONFIG_GLOBAL` and `HOME`/`USERPROFILE`
  pinned to the scratch tree. Every mutant was applied to a second throwaway copy
  (`/tmp/mut5`) and reverted after each run. Reproducer:
  `TASK-260906-1uf713_review-probes-stage-c-5.tgz`.
- **No `-race` suite and no `test-gate.sh` lane ran at any point during this review**, so
  nothing here is Go-test-lock contention noise.

---

## Hosted lanes on the reviewed head

`gh pr checks 61` on `71e6baec`, run 34050303111, read after every lane settled:

| Lane | Result |
|---|---|
| Lint | pass 33s |
| Naming gate | pass 6s |
| Interop conformance gate | pass 21s |
| Gate self-test (ubuntu / macos / windows) | pass 8s / 11s / 25s |
| Test (ubuntu-latest) | pass 3m21s |
| Test (macos-latest) | pass 7m53s |
| **Test (windows-latest)** | **pass 41m38s** |
| Race (ubuntu-latest) | pass 9m21s |
| Race (macos-latest) | pass 17m6s |
| Candidate suite | `skipping` (gated matrix; it does not run on a PR event) |

**All eleven hosted lanes are green on the exact head I reviewed, Windows included.** C4-B2
is fixed and the lane confirms it. Every finding below is invisible to all eleven, which is
the standard this chain has been held to since cycle 2.

## The candidate dispatch

Run 34052590291, `workflow_dispatch` on head `71e6baec`, candidate revision
`87a0d0060bad64ab883d007dcdf35df7485368bf` (curator-spec main), `CI_REQUIRE_FULL_ROOT=1`:

| Job | Result |
|---|---|
| Candidate suite (ubuntu-latest) | **failure** |
| Candidate suite (macos-latest) | **failure** |
| Candidate suite (windows-latest) | **failure** |

Extracted from the three `candidate-evidence-<os>` artefacts (9995047563, 9995114725,
9995506885), not from `--log-failed`. All three runners fail identically and **only** in
`internal/config`, on `TestManagerConfigV2SchemaCases` and `TestManagerConfigV2Vectors`. On
ubuntu that is exactly 30 subcases — every one of them a case
`bd39adb`/`2f2dfa4` added to curator-spec after the task authority `550579d`
(`valid-overlay-path-*` now rejected because this tree requires a form on every overlay;
`invalid-overlay-*-with-form` now accepted for the same reason; plus `valid.json` and
`schema2-every-knob`). Every other package serves and passes. The producer's rework-4
characterisation of this lane is accurate in every particular, including the failing-case
list, and it reported the lane red rather than omitting it.

I reproduced the same failure locally and bounded it:

| Root | `go test ./internal/config/` | Note |
|---|---|---|
| curator-spec `550579d` (**task authority**) | **exit 0**, `ok 0.571s` | `suite-plan: served=72 deferred=0` |
| curator-spec `87a0d00` (main today) | **exit 1**, 30 failing subcases | the `bd39adb` cases only |
| `SPEC_PIN` `0ed5c691` | deferred, never run | `suite-plan: served=69 deferred=3` — `internal/config`, `internal/envfragment`, `internal/envmarker`, exactly the three rows `root-artifacts.tsv:39-41` register |

So, answering the brief's question directly: on the candidate lane the environments
families **ran on all three runners and passed on none of them** — every other package
served and passed, and the only failures are the schema-2 manager-config families, which are
exactly the artefacts curator-spec changed after this task's authority. Against the task
authority `550579d` the same families run and pass, on this host and in the producer's
authority lane, and the root-artifact deferral does what it was registered to do on the
`SPEC_PIN` root. See C5-m3.

---

## MAJOR

### C5-M1 — the documented retry after the §9.5 stop is a silent no-op that reports success

`repeat-of: cycle-1 M1 / cycle-2 C2-m2` (a comment claiming a production path it does not
have) and `cycle-4 C4-B1` (a command reporting success for work it did not do).

Stage (c) adds the `--takeover` flag to `profile install` (commit `671f9925`,
"Stage (c): onboarding takeover across the mutating operations"; `origin/main` has zero
occurrences of `takeover` in `cmd/curator/profile.go`), and adds the §9.5 stop that makes it
necessary. `cli/curator.md` at `550579d`:

> `curator profile install <git-url|path> … [--use] [--takeover]` — … `--use` takes no name
> and **activates the installed root**; … `[--takeover]` takes over the unmanaged files the
> install would write (section 9.5 notice, section 8.3 backup); without it the write fails
> with `environment_surface_unmanaged_conflict`

Driven on `71e6baec`, the ordinary §9.5 onboarding case — a hand-written `CLAUDE.md` in the
claude_code home:

```
$ curator profile install ./src --as tk --use
    [EXIT 1]
    claude_code: environment_surface_unmanaged_conflict: CLAUDE.md exists and no marker records it
    codex_cli: switched   opencode: switched   pi: switched
    installed profile tk (lock sha256:d667296b…)
    curator: profile_use_partial: the scope is partially switched; the recorded current is unchanged
    STATE current=<none>  CLAUDE.md=OPERATOR HAND-WRITTEN CONTEXT  backups=[]

# the notice tells the operator to take over; they add the flag and retry the same row
$ curator profile install ./src --as tk --use --takeover
    [EXIT 0]
    updated profile tk (lock sha256:d667296b…)
    STATE current=<none>  CLAUDE.md=OPERATOR HAND-WRITTEN CONTEXT  backups=[]
```

**Nothing happened.** No takeover, no backup, no activation, machine still has no current —
and the command exited 0 saying `updated profile tk`. Identical with a foreign-manager
symlink: after `environment_foreign_manager_detected: … abort, or take over with backup`,
the retry leaves the symlink in place, writes no backup, and reports success.

Root cause: the failed first attempt **does** publish `source` and `lock` (the switch fails
after the install), so the retry takes `installLocked`'s `prior == source` branch
(`internal/envprofile/envprofile.go:644`), which for a `path` source now calls
`reinstallPathLocked`. That function never reads `options.Use` or `policy.Takeover`, and
because the lock hash is unchanged it returns at `envprofile.go:768` **before**
`resyncCurrentScopes` — so no code below the branch can honour either flag.

Only one of the four `--takeover`-bearing rows recovers from this state:

| retry after the stop | exit | result |
|---|---|---|
| `profile install ./src --as tk --use --takeover` | 0 | **no-op**, `updated profile tk` |
| `profile update tk --takeover` | 0 | no-op, `tk: unchanged` |
| `profile update --all --takeover` | 0 | no-op, `tk: unchanged` |
| `profile sync --takeover` | 0 | no-op (current is empty, so it syncs `default`) |
| `profile use tk --takeover` | 0 | **correct** — switches, notice, backup generation 1 |

**The attestation half.** `reinstallPathLocked`'s own doc comment
(`internal/envprofile/envprofile.go:715-726`) states:

> Like the git reinstall delegation it replaces, it reports updated and never activates: a
> reinstall refreshes the lock and re-materializes scopes already on the profile, and a
> `--use` switch of a non-current profile **takes the fresh-install path, not this one**.

The last clause is false. A `--use` of a non-current profile whose source is already
recorded takes *this* path, and the switch never happens. That is the cycle-1 M1 shape the
rework-2 brief named explicitly, in code written after that brief.

**Scoping, stated honestly.** The `--use` drop on a same-source reinstall is *structurally*
pre-existing: `origin/main`'s `installLocked` has the byte-identical `prior == source`
delegation returning before `options.Use` is consulted (`envprofile.go:507-518` on main). I
could **not** drive that on main — an absolute local path is classified as a `path` operand
and a `file://` operand is refused (`file:// git sources carry no network identity`), so I
have no local git source and I report the main behaviour as a code reading, not as a
measurement. What is unambiguously stage (c)'s is: the `--takeover` flag itself, the §9.5
stop that makes this the documented retry, the new function that re-implements the drop, and
the comment asserting it cannot happen.

**Fix direction.** Decide from `cli/curator.md` what `--use` and `--takeover` mean on a
reinstall — either honour them in the reinstall arm (activate when `--use` is given and the
profile is not current; thread `policy.Takeover` through the resync), or refuse the
combination loudly. Do not leave a third outcome where both flags are accepted and dropped.
Then make the comment true. Evidence: one test that drives the whole §9.5 sequence through
`run()` — stop, retry with `--takeover`, assert the machine current moved and a backup
generation exists — plus a narrowing mutant that exempts exactly the already-installed case.

---

## MINOR

### C5-m1 — the C4-B1 regression test pins only the `--as` addressing mode

`TestPathSnapshotImmutableAcrossUpdateSyncUse` drives the reinstall as
`Install(home, InstallOptions{Operand: source, As: "pk"})`. Two narrowing mutants on the
arm it protects survive the full `./internal/envprofile/` suite:

| # | mutant | result |
|---|---|---|
| my1 | `if isPath` → `if isPath && options.As != ""` | **SURVIVES** |
| my4 | `if isPath` → `if machine, _ := Current(home); isPath && machine == name` | **SURVIVES** |

`my1` is not academic. Built into a binary and driven, it restores C4-B1 exactly, through
the CLI row's **default** form (`--as` is optional in `cli/curator.md`):

```
stock    curator profile install ./src   (no --as, edited source)
         pin ff6c35c7… -> 921fe8db…   RE-READ    CLAUDE.md EDITED lines: 1
my1      curator profile install ./src   (no --as, edited source)
         pin ff6c35c7… -> ff6c35c7…   NO RE-READ CLAUDE.md EDITED lines: 0
         "updated profile pk" at exit 0
```

Production is correct on both forms — I drove the no-`--as` reinstall and the non-current
reinstall on the stock binary and both behave right. This is a coverage point, recorded
because the blocking finding of the previous cycle lived in precisely the addressing mode
its test did not take. One extra assertion (or a table over `As: ""` / `As: "pk"` × current
/ not-current) closes both.

### C5-m2 — two stage-(c) skip reasons are still unregistered

C4-B2 is fixed and the fix is the right one: `skip-classes.tsv` was *widened* to
`this environment can (inspect\|read) a mode-000 (directory\|file)`, no skip text was changed
to lie about the artefact, and no case was silenced. Reproduced locally against
`platform-case-gate.sh`'s own `classify()`:

```
reason: this environment can read a mode-000 file; unreadability is untestable here
  4a8a1d42 registry -> UNCLASSIFIED     (reproduces the red Windows lane)
  71e6baec registry -> host-capability/allow
sibling "…mode-000 directory…" reasons -> host-capability/allow on both
```

Sweeping every `t.Skip` reason in the twelve test files this stage touches through the same
classifier, **two remain UNCLASSIFIED**, both introduced by stage (c):

```
internal/envprofile/import_test.go:781    t.Skipf("chmod refused: %v", err)
internal/envprofile/pathkind_test.go:358  t.Skipf("chmod refused: %v", err)
```

Their two siblings (`import_test.go:246`, `pathkind_test.go:185`) spell the same condition
as `"this environment can read a mode-000 directory: chmod refused: %v"` and classify — and
carry a comment explaining that they do it precisely so the classifier sees them. The other
unclassified reason, `no git on PATH`, is pre-existing on `origin/main` and out of scope.

This cannot redden a lane today: `os.Chmod` does not fail on any of the three runners, so
the branch is dead there. It is latent, it is the C4-B2 class one sibling over, and it is a
one-line fix either way (register the bare reason, or spell it like its siblings).

### C5-m3 — the AC's "against curator-spec main" clause is unmet, and needs an orchestrator decision

The task AC says the conformance subset "passes through `CURATOR_CONFORMANCE_ROOT` against
curator-spec main with `CI_REQUIRE_FULL_ROOT=1`". At today's main (`87a0d00`) it does not:
two candidate-suite jobs are red and I reproduced the 30 failing subcases locally. At the
task authority (`550579d`, named in the task description, in every brief, and as the CR base
OID) it does, and I verified that too.

The gap is entirely the `path`-overlay reconciliation the orchestrator itself landed on
curator-spec after this task's authority was frozen (TASK-260906-3x0w4y / curator-spec
PR #47, `bd39adb`), plus the scheme-discriminator follow-ups (`2f2dfa4`, `87a0d00`).
Consuming it means porting the spelling discriminator into the Go reader and reaching the
feature from all three surfaces — real work the cycle-4 brief explicitly said not to require
in-cycle.

**On the M1 bound, which the brief asked me to judge:** it is now *half* stale and the
producer says so honestly in rework 4. The **spec contradiction** it was raised for is gone —
§6/§12.1/`manager-config-v2` now agree that a form is git-only. What survives is an
**implementation gap**: this tree still enforces the `550579d` form-on-every-overlay rule, so
no operator can declare a `path` overlay from any of the three surfaces. Ledger rows 303–304
still describe what their tests assert and name the bound, so nothing untruthful lands. The
bound must be **retired and reissued** as an implementation-consumption task rather than
carried as a spec bound. That is an orchestrator call, not a producer one: either stage (c)
lands with this gap and a named follow-up, or the consumption work joins this leaf. I am not
treating it as a reason to reject the code.

### C5-m4 — `reinstallPathLocked`'s copied blocking-audit gate has no test; exploitability unknown

The new function copies `updateLocked`'s new-member gate:

```go
if report.Blocking() { … DiagUpdateBlocked … "the old lock stands" }
```

Mutant `my2` narrows it to `report.Blocking() && resolved.Kind != contextlock.KindContext`
— admitting exactly a blocking context member — and **survives** the full
`./internal/envprofile/` suite. So does the `updateLocked` twin it was copied from: `grep`
finds only one test naming `DiagUpdateBlocked`, and it is the "builtin local profile does not
move" case. A second, independent gate (`strictAuditMember`, strict/`fail_on: high`) remains
in the same loop under the mutant, so I could not establish that a blocking-but-not-strict
finding exists. **I report exploitability as unknown rather than inferring it** — the finding
is that a refusal introduced by this delta has no narrowing mutant, which the stage's own
definition of done requires.

---

## Observations (not findings)

- **`--directory` on a path operand is unpinned, and the mutant admits it end to end.**
  `installLocked`'s `isPath && (Range \|\| Tag \|\| Revision \|\| Directory)` refusal is
  **stage (a)** code (`99cc576c`, on `origin/main`), and its only test
  (`TestAbsentPathOperandWithRequirementIsRefConflict`) covers `Range` and `Tag` only.
  Dropping `\|\| options.Directory != ""` survives the suite, and driven:
  `curator profile install ./src --directory sub --as pd` → **exit 0, installed**, where
  stock refuses `profile_install_ref_conflict`. §1 makes `directory` on a `path` declaration
  `profile_source_invalid`. Out of the stage-(c) delta, so not a finding here, but it is one
  `--revision`/`--directory` row from being closed.
- **A fourth §8.4 site already exists, on `origin/main`.** The producer's answer to cycle 4's
  structural question ("nothing prevents a fourth site; the class is closed by vigilance plus
  per-site mutants") is honest, and within the stage-(c) delta I checked every read of
  manager-owned state and each of the three now has a pin. But `purgeHomes`
  (`internal/envprofile/switch.go:706`, unchanged from `origin/main`) is the same shape
  verbatim — `marker, err := envmarker.Read(native); if err != nil || marker == nil … {
  continue }` — so an unreadable marker makes a purge silently skip that home. Pre-existing,
  out of the delta, and worth a task rather than a stage-(c) rework.
- **Carried from cycle 4, unchanged and still true:** `envprofile.SetCurrent` is an exported
  second writer of `CurrentFile` with test-only callers, sitting below the C3-B1 seam (I
  re-enumerated: `switch.go:248` is the only production write); and `CURATOR_SYSTEM_CONFIG`
  overrides the system-file path, which I re-confirmed is pre-existing (`5e183d58`, on
  `origin/main`) and therefore out of scope for this verdict, though stage (c) is what turns
  it into a fleet-policy bypass.
- **The Change Request's repository delta is another element's work.** `CR-…-5` revision 5
  has `repository_delta: present` and its candidate tree `2e6ca472` is
  `550579d..87a0d00` — the five *curator-spec* commits of the `path`-overlay reconciliation
  (`bd39adb`, `2f2dfa4`, `18dca85`, `407424e`, `87a0d00`), i.e. TASK-260906-3x0w4y's
  deliverable, which happens to sit on the shared story branch. Revisions 1–4 were all
  `empty`, which cycle 4 correctly explained as structurally right for a leaf whose scope is
  the *curator* repository and whose briefs forbid writing into the control root. Revision 5
  is not empty and not this leaf's work. Since my verdict is CHANGES REQUESTED I do not call
  `accept_cr`, so nothing is stamped — but the next reviewer should not accept a revision
  whose delta belongs to another task.

---

## What I verified and found correct

**C4-B1 — both halves of §1, driven through the production entry point.** Install, edit the
source, then:

| command | exit | pin | verdict |
|---|---|---|---|
| `profile update pk` | 0 | `ff6c35c7…` | immutable |
| `profile sync` | 0 | `ff6c35c7…` | immutable |
| `profile use pk` | 0 | `ff6c35c7…` | immutable |
| `profile install ./src --as pk` | 0 | `ff6c35c7…` → `43d9f269…` | **re-read** |

and the re-materialized `CLAUDE.md` carries `EDITED AFTER INSTALL`. The reinstall pin equals
the pin a *fresh* install of the same edited tree produces (`43d9f269…`), so the two paths
agree on content. The seams the brief asked me to attack:

| seam | observed |
|---|---|
| same path, different name (`--as pkalt`) | exit 0, fresh install, identical pin |
| different path, same name | exit 1, `profile_name_taken: profile "pk" is already installed` |
| source deleted after install → `update` / `sync` / `use` / `update --all` | exit 0 each, pin unmoved |
| source deleted after install → reinstall | exit 1, `profile_source_path_missing` |
| reinstall with no `--as` (name from the manifest) | exit 0, re-read, surface refreshed |
| reinstall of a non-current profile | exit 0, pin moves, current profile's surface untouched |
| `install --use` on an already-installed path root | **see C5-M1** |

The producer's own mutant (`if false && isPath`) is **killed** by
`TestPathSnapshotImmutableAcrossUpdateSyncUse` with the message
`reinstall of the edited tree pinned the same hash: the §1 reinstall is dead`.

**C4-M1 — all three refusals driven, all three mutants dead.** Each refusal is reachable from
the CLI, not just from `UpdateWithPolicy`:

```
lock's state pin stripped   -> exit 1  profile_source_invalid: profile "pk" lock carries no state pin for its path root
store entry removed         -> exit 1  profile_source_invalid: profile "pk" path snapshot cannot be read: context_manifest_invalid: …
snapshot manifest renamed   -> exit 1  profile_source_invalid: profile "pk" snapshot names "other", want lock root "pk"
```

| # | mutant | result |
|---|---|---|
| P2 | drop `\|\| rootMember.StateHash == ""` | **KILLED** — `TestUpdatePathWithoutStatePinIsSourceInvalid` |
| P3 | delete the snapshot-name check | **KILLED** — `TestUpdatePathSnapshotNameMismatchIsSourceInvalid` |
| P4 | failed snapshot read → `unchanged`, nil error (§8.4) | **KILLED** — `TestUpdatePathMissingSnapshotIsSourceInvalid` |

P4 was cycle 4's worst survivor and it is genuinely dead. A reinstall over a broken lock is
also the repair — `profile install ./src --as pk` with the state pin stripped exits 0 and
republishes a valid lock — which is the right asymmetry: `update` refuses what it cannot
prove, `install` re-establishes it.

**C4-B2 — the registry, and the sweep.** Fixed as described in C5-m2; the classifier
reproduction is above and in the archive.

**C4-m1 — the builtin `default` under a lock.** Driven: `curator profile use default` with
`require_current_profile: acme` locked → exit 1,
`environments.require_current_profile: profile use of "default" is refused: the system
configuration requires current profile "acme"`, current unchanged; `profile use acme` → exit
0. Cycle 4's surviving mutant (`|| name == DefaultProfile`) is now **KILLED** by
`TestPolicyFromConfigCarriesEnvGates` (`locked require must refuse the builtin default
profile`).

**C4-m2 — documented, and the documentation is a defensible answer.** The producer's position
— a machine already violating the lock fails loudly at exit 1 with the lock moved and the
surfaces stale, recoverable by switching to the required profile, silent success being the
worse outcome — is stated in the report rather than left implicit, which is what cycle 4
asked for.

**Things nobody had driven yet, all correct.** `env config set overlays_allowed true` and
`env config unset overlays_allowed` under a lock → exit 1 naming the knob and the system file,
while unlocked knobs (`overlay_default_weight`, `precedence.winner`) write normally.
`isolation` locked toward `isolated` → exit 1
(`a system file locks isolation only toward shared`); locked toward `shared` → exit 0.
Locking a non-lockable key → exit 1 (`cannot lock "environments.secret_material_waivers"`);
*carrying* it unlocked → exit 1 (`is not lockable and must not be carried by the system
file`). `schema_version: 3` → exit 1 (`unsupported config schema_version 3; this config
requires a newer tool`); `schema_version: 1` → valid. `compose add` and `compose list` under
a locked `overlays_allowed: false` both print the manager §1 inert warning and exit 0, and
`profile update` then reports `unchanged`. `--takeover` on a *clean* machine takes over the
foreign symlink correctly: the surface stops being a symlink, the foreign file's bytes are
unchanged, backup generation 1 is written, and the notice is printed. `profile import`
against an **absent** marker imports; against a **corrupt** marker →
`environment_import_lossy` naming `environment_marker_invalid: malformed protocol JSON`;
against an **unreadable** marker → `environment_import_lossy` naming `permission denied` —
three distinct outcomes with the §8.4 distinction intact.

**The §8.4 sweep inside the delta.** Every read of manager-owned state in the stage-(c)
production files carries the distinction explicitly in code and has a pin: `readSkillsLedger`
(absence → empty set, any other failure → error), `readRootSurface` (no marker → detect;
unreadable marker → loss), `readSkillsSurface` (absent dir → nothing; unreadable → loss;
unreadable ledger → the ledger is the loss), and `updateLocked`'s snapshot read (P4). The
fourth site I found is outside the delta — see Observations.

---

## Judging the rework-4 report against the attestation standard

Accurate on everything I could check. The finding→resolution table covers all five; the
C4-B1 paired evidence reproduces exactly, including the reinstall-pin-equals-fresh-pin
assertion; the three refusal mutants are real kills and I re-ran each; the local classifier
reproduction is correct and I reproduced it independently; the C4-m1 mutant dies at both
levels; the C4-m2 answer is stated rather than omitted; the deferral counts it quotes
(`served=69 deferred=3` on the pin, `served=72 deferred=0` on the authority) match what I
measured; the `git archive` / `export-subst` lesson is right and the authority lane was
correctly re-run from a real checkout; the candidate lane is reported **red** with its exact
failing families rather than buried; the CI row honestly says the head was not pushed when
the report was written (the orchestrator pushed afterwards, and all eleven lanes are green);
the bounds list is updated rather than trimmed, and the M1 bound now names the landed spec
commit; and the "no `git add -A`" discipline held — `git status --short` empty, no stray
artefacts in the tree.

Two rows I would not sign as written:

1. **"AC ratio: 15 of 15 … the path row now drives both halves."** Both halves are driven,
   but only in one addressing mode; the CLI row's default form is unpinned (C5-m1), and the
   `--use`/`--takeover` behaviour of the same row is wrong (C5-M1). I read the path row as
   driven and the onboarding/takeover row as **partially** driven — 14½ of 15, with the
   half being `profile install --takeover` after the §9.5 stop.
2. **"the class is closed by vigilance plus the per-site narrowing mutants now committed."**
   True inside the delta. `purgeHomes` shows the class is not closed in the file the delta
   edits, which is worth saying out loud rather than leaving to the next stage to rediscover.

---

## Gate table (standalone processes, on `71e6baec` in a `git archive` copy at `/tmp/rev5`)

| Gate | Root | Exit | Output |
|---|---|---|---|
| `go build ./...` | — | 0 | empty |
| `go vet ./...` | — | 0 | empty |
| `gofmt -l cmd internal` | — | 0 | empty |
| `golangci-lint run ./...` | — | 0 | `0 issues.` |
| `bash .github/ci/gate-selftest.sh` | — | 0 | `gate-selftest: 94 passed, 0 failed` |
| `bash .github/ci/no-broad-suppression.sh` | — | 0 | `no-broad-suppression: ok` |
| `bash .github/ci/ledger-consistency.sh` | — | 0 | `206 rows checked across linux darwin windows`, `ok` |
| `bash .github/ci/suite-plan.sh` | `SPEC_PIN` `0ed5c691` | 0 | `served=69 deferred=3 excluded=0` (config, envfragment, envmarker) |
| `bash .github/ci/suite-plan.sh` | curator-spec `550579d` | 0 | `served=72 deferred=0 excluded=0` |
| `go test -count=1 ./internal/config/` | curator-spec `550579d` | 0 | `ok 0.571s` |
| `go test -count=1 ./internal/config/` | curator-spec `87a0d00` | **1** | 30 failing subcases, all `bd39adb` — C5-m3 |
| `go test -count=1 ./internal/envfragment/ ./internal/envmarker/ ./internal/contextstore/` | curator-spec `550579d` | 0 | all `ok` |
| `go test -count=1 ./internal/envprofile/` | — | 0 | `ok 67.259s` (the kill/survive oracle) |
| `gh pr checks 61` | — | — | 11 of 11 **pass** on `71e6baec` |
| `gh api …/artifacts/9995047563/zip` (candidate-evidence-ubuntu-latest) | — | 0 | the 30 failing subcases above |

**Not run by me, stated plainly:** the two full `test-gate.sh` lanes and
`go test -timeout 30m ./cmd/curator`. The hosted `Test (ubuntu/macos/windows)` lanes execute
the default-root lane and the full `cmd/curator` suite on this exact head and are green on
all three platforms, and the candidate dispatch executes the `CI_REQUIRE_FULL_ROOT=1` lane
against curator-spec main on all three — strictly better evidence than a fourth local rerun
on one host, and both are read above from their own artefacts. The lane with **no** hosted
coverage is `CI_REQUIRE_FULL_ROOT=1` against the task authority `550579d`; I covered its
decisive part directly by running every root-reading package against that root (rows above),
and the producer's full run of it stands unchallenged. A gate I did not run is listed as not
run.

---

## Is stage (c) safe to land?

**Not at `71e6baec`, and one thing must change.**

Everything the previous four cycles found is fixed and stays fixed: the §1 snapshot is
immutable across `update`, `sync` and `use` and refreshes on reinstall; the §9.2 seam is
structural and single; the §12.2 pin transcribes the spec independently; the three new
refusals are driven and their mutants die; the Windows lane is green for the first time in
this stage; and every lane on the PR passes. Composition, the precedence primitives and the
import were attacked by cycle 3 and re-confirmed by cycle 4, and I found nothing that
contradicts either.

The one thing is C5-M1: `profile install … --use --takeover`, the exact command the new §9.5
stop invites an operator to run, accepts both flags, does nothing, and exits 0 saying
`updated profile <name>` — under a doc comment that says this case takes another path. Fix
that row (honour the flags or refuse the combination), make the comment true, and drive the
whole stop-and-retry sequence through `run()`. C5-m1, C5-m2 and C5-m4 are one assertion, one
registry line and one mutant respectively and should ride along. C5-m3 is not the producer's
to decide.

With C5-M1 closed, I would sign this stage without hedging.
