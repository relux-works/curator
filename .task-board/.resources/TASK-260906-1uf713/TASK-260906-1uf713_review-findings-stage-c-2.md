# TASK-260906-1uf713 — review findings, stage (c) cycle 2 (rework 1)

Verdict: **CHANGES REQUESTED**. One blocking, one major, four minor.

`repeat-of:` **cycle-1 B3** (finding C2-B1 below — the fix's gate is bypassed by a
different addressing mode) and **cycle-1 m1** (finding C2-M1 — the same
`parseSystemEnvironments` gate, still not class-wide). Two consecutive same-class
findings on `parseSystemEnvironments`: the next step there is a gate that derives the
test from the knob list, not a third hand-written list.

- Subject: `~/Developer/ReluxWorks/.worktrees/curator-stage-c`, branch
  `feat/agent-environments-stage-c` at `0bcea201984be360bec42ba49296b9d82ac862ce`,
  16 commits past `origin/main`. PR head confirmed with `gh pr view 61 --json
  headRefOid` = `0bcea201…`. The producer worktree was read-only: `git status
  --short` empty and HEAD unchanged before and after.
- Authority: curator-spec `550579d` (this story worktree).
- Method: a `curator` binary built from `0bcea201` (`go build ./cmd/curator`, exit 0)
  and invoked as a standalone process against scratch homes, with `CURATOR_CONFIG`,
  `CURATOR_SYSTEM_CONFIG`, `CLAUDE_CONFIG_DIR`, `CODEX_HOME`, `XDG_CONFIG_HOME`,
  `PI_CODING_AGENT_DIR` and `HOME`/`USERPROFILE` pinned to the scratch tree. My own
  narrowing mutants were applied to a throwaway copy (`.temp/review2/mut`), never to
  the worktree. Reproducer: `TASK-260906-1uf713_review-probes-stage-c-2.tgz`.

---

## The Change Request's empty repository delta

`CR-TASK-260906-1uf713-2` revision 2 reports `repository_delta: empty`: candidate tree
`4829dd34` is identical to base `550579d1`, and the patch has zero paths. That is
**structurally correct for this leaf and is not what the verdict turns on.**

The story workspace is a *curator-spec* worktree; base OID `550579d1` is the
curator-spec authority commit. This leaf's declared scope is "curator repository,
branch `feat/agent-environments-stage-c` in worktree
`/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-c`" — a different repository,
delivered on PR #61. No change to curator-spec was ever in scope; a non-empty delta
here would itself have been a scope violation, and the producer brief forbids writing
into the control root. The reviewable work is `git diff origin/main..HEAD` in the
curator worktree (27 files, +5398/-151 for the stage; `git diff 833918d2..HEAD`, 16
files, +628/-37, for rework 1), and that is what I reviewed at `0bcea201`, the exact
PR #61 head.

So the emptiness is correct and I would not have rejected the revision for it. I am
rejecting it for C2-B1 and C2-M1, which live in the curator tree.

---

## Hosted lanes on the reviewed head

`gh pr checks 61` on `0bcea201`:

| Lane | Result |
|---|---|
| Lint | pass 42s |
| Naming gate | pass 6s |
| Interop conformance gate | pass 26s |
| Gate self-test (ubuntu / macos / windows) | pass 7s / 11s / 18s |
| Test (ubuntu-latest) | pass 3m43s |
| Test (macos-latest) | pass 7m22s |
| Race (ubuntu-latest) | pass 10m1s |
| Race (macos-latest) | pass 18m34s |
| Test (windows-latest) | pass 33m57s |
| Candidate suite | skipping (expected: gated matrix) |

**Every hosted lane was green on `0bcea201`, the exact head reviewed**, including
`Test (windows-latest)`, which was still pending for the first ~50 minutes of this
review and which I waited out rather than record a verdict around. Cycle 1 could not
make that statement; this one can. The orchestrator's `pinOperatorHome` fix
(`0bcea201`) is confirmed by the runner that produced the original failure —
`TestTakeoverWarnsDotfileHeuristic` no longer fails there.

Green lanes are not why this review does not accept. Both findings below are invisible
to every lane: C2-B1 is an untested addressing mode, C2-M1 is a surviving mutant.

---

## BLOCKING

### C2-B1 — the B3 managed-home gate fails OPEN on a failed marker read, and re-imports curator's own generated root context

`repeat-of: cycle-1 B3`

The B3 fix is:

```go
// internal/envprofile/import.go — readRootSurface
if marker, err := envmarker.Read(native); err == nil && marker != nil {
    return nil, nil            // managed home: skip
}
```

`envmarker.Read` returns `(nil, nil)` for **absence** and `(nil, err)` for an
unreadable or undecodable marker — its own package doc says an invalid marker "fails
closed and is reported as `environment_marker_invalid`". Here `err != nil` falls
through the skip, so a home whose marker cannot be read is treated as an unmanaged
native home. Driven through the CLI on a machine curator manages:

```
$ curator profile install ./root --use          # 4 markers written
$ curator profile import --as clean             # exit 0
  managed root imported as native: no           # gate holds

$ printf 'not json at all' > <native>/claude/.agent-environment.json   # a FAILED READ
$ curator profile import --as corrupt           # exit 0, no warning, no loss
  managed root imported as native: YES
$ head -4 <store>/context/corrupt/<hash>/context/claude_code.md
<!--
curator-root-context-v2
root: acme 1.0.0 state sha256:38074ce9…
member: acme 1.0.0 state sha256:38074ce9… weight 0
```

This is cycle-1 B3 verbatim — the reassembled module carries a complete §5.1
generation header, so activating the imported profile nests one generation header
inside another inside a chapter, a shape the closed §5.1 grammar does not admit and
which drift and marker verification then read. It is reached by an ordinary
addressing mode (a truncated or hand-edited marker), not a race. Stage (a)'s F1 and
the cycle-1 review brief both name this shape: a gate bypassed by one addressing mode
nobody tested.

It is also the **§8.4 absence-versus-failed-read class the same commit fixed
correctly one function away**: `ee6743a2` changed `readSkillsLedger` to return an
error so a failed ledger read becomes a loss (cycle-1 m2), and left the marker read
collapsing failure into absence. The rule was known and applied inconsistently.

§9.6 decides the outcome without ambiguity — "an absence and a failed read stay
different facts (§8.4): an absent surface never appears in the loss list, and a
failed read is always a loss, never treated as absence". A marker that cannot be
read is a failed read of the evidence that classifies the surface.

`TestImportSkipsManagedRootContext` is otherwise good — my narrowing mutant
(`&& envID != "codex_cli"`, gate present, one member admitted) kills it. It
constructs only intact markers, so the failed-read mode has no test at all.

---

## MAJOR

### C2-M1 — `parseSystemEnvironments` is still not class-wide: three surviving narrowing mutants, one of them on the §9.1 secret-material waivers

`repeat-of: cycle-1 m1`

Cycle 1 found `&& key != "forms"` surviving and asked for a class-wide test "so no
single knob can be admitted". `TestSystemV2LockableSubsetIsClassWide` covers all six
§12.2 lockable keys and **five** of the twelve non-lockable §12.1 knobs
(`forms`, `current_profile`, `overlays`, `overlay_default_weight`,
`backup_retention`). The predecessor's mutant now dies:

```
--- FAIL: TestSystemV2LockableSubsetIsClassWide/non-lockable/forms
    environments_test.go:404: non-lockable knob "forms" admitted: err = <nil>
```

The other seven are unattacked. Three narrowing mutants I applied to
`internal/config/environments.go`, each keeping the gate and admitting exactly one
member, **all survive** — with and without `CURATOR_CONFORMANCE_ROOT` pointed at
curator-spec main:

| Mutant (`parseSystemEnvironments`) | `go test ./internal/config/` | with root |
|---|---|---|
| `&& key != "secret_material_waivers"` | **ok (SURVIVES)** | **ok (SURVIVES)** |
| `&& key != "xdg_seed_allowlist"` | **ok (SURVIVES)** | **ok (SURVIVES)** |
| `&& key != "in_place_mode"` | **ok (SURVIVES)** | **ok (SURVIVES)** |

The first one matters beyond coverage arithmetic. `mergeSystemEnvironments` applies an
**unlocked** system `environments` knob as a default the machine file inherits when it
leaves the knob absent (manager §1 rule 3), so `parseSystemEnvironments` is the only
thing standing between a system file and the effective machine configuration. Driven
end to end with the mutant built:

```
# stock 0bcea201
$ curator env config show
curator: environments.secret_material_waivers: is not lockable and must not be
         carried by the system file (system config .../system.json)      # correct

# mutant admitting exactly secret_material_waivers
$ curator env config show | grep secret
  "secret_material_waivers": [                                           # injected
```

A system file would supply the §9.1 secret-material audit waiver list. The manager §1
credential rule — "no key that selects or constrains credential material is lockable"
— is exactly what this gate enforces, and `secret_material_waivers` and
`xdg_seed_allowlist` (which selects the entries seeded from XDG config, `ssh` and
`git` among them) are its most load-bearing members. The production gate is correct
today; what is missing is any test that would notice it stopping being correct.

The fix is a gate, not a sixth hand-written case: derive the refused set from the
§12.1 knob list minus `LockableEnvKeys`, so a knob added to §12.1 without a matching
case cannot silently widen the system file's reach.

---

## MINOR

### C2-m1 — the §9.5 dotfile heuristic list is POSIX-only, and the orchestrator finding's first requirement is unanswered

`dotfileStateDirs` (`internal/envprofile/managed.go:727`) is
`{".local/share/chezmoi", "chezmoi"}` and `{".config/home-manager", "home-manager"}`,
joined under `os.UserHomeDir()`. The *mechanism* is portable and `0bcea201`'s
`pinOperatorHome` correctly makes the fixture assert real behaviour on all three
runners — that commit is right and I have no objection to it. But the *list* is not
platform-aware, so on Windows a real chezmoi install (`%LOCALAPPDATA%\chezmoi`) is
never detected and the heuristic can only ever fire on a path chezmoi does not use
there.

`orchestrator-finding-stage-c-windows-heuristic.md` required exactly this to be
decided: "Decide from §9.5 whether the closed list is meant to be platform-specific …
Say which, with the sentence that decides it. If §9.5 does not decide it, that is a
spec gap: state it as one rather than picking whichever makes the test pass." Neither
`TASK-260906-1uf713_rework-report-1.md` (whose Windows sweep covers fixture shapes
only) nor `0bcea201`'s message answers it, and no bound carries it. §9.5 gives
`~/.local/share/chezmoi` and "and the like" without saying whether the list is
per-platform, so on my reading this is a spec gap and should be declared as one.

Non-blocking by construction (§9.5: "The heuristic never blocks"), which is why this
is minor and not major — but the requirement was explicit and a bound once demanded is
dropped only by fixing it.

### C2-m2 — the B1 fix left `config.CheckMachineUse` with no production caller, and a test comment that says otherwise

`cmdProfileUse` no longer calls `cfg.CheckMachineUse`; the gate is
`Policy.CheckMachineUse` at the seam. `Config.CheckMachineUse`
(`internal/config/environments.go:768`) now has zero production callers — only
`TestRequireCurrentProfileGate` and `TestRequireCurrentProfileUnlockedCarriesNoRefusal`
reach it — and that test's doc comment still reads "CheckMachineUse, **the production
path `profile use` takes**". That sentence is false at this head. This creates no
hole (the real gate is covered — see the mutant table), but it is the uncalled-gate
shape the definition of done names, sitting under a comment asserting the opposite.
Delete it or repoint the tests at `Policy.CheckMachineUse`.

### C2-m3 — a ledger row still asserts the behaviour the M1 bound says is unreachable

Row 303 was correctly fixed. Row 304 still reads:

```
internal/envprofile  TestPathOverlayJoinsClosure  …  a machine-declared path overlay
joins the closure flagged overlay at the machine default weight
```

No machine can declare a path overlay at this head — that is the M1 bound, and the
test builds a `Policy` by hand. Same class of untruthful row cycle 1 flagged in 303.

### C2-m4 — the remove-first hardening is asymmetric inside `applyPlan`

`applyPlan` gained `_ = os.Remove(...)` before the copied-surface writes
(`managed.go:852`) but not before the provisioning seed write (`:876`) or the marker
write (`:889`), while `switch.go` added it before its marker write (`:571`). These
targets are under `EnvRoot(home)` — the manager's own tree — so this is **not** the
B2 class and I found no write in the stage that can land outside it (see the sweep
below); it is a consistency gap, not a data-loss path. Judgement call: close it or
say why the manager's own tree needs no remove-first.

---

## What I verified and found correct

Every cycle-1 finding reproduces as fixed. Driven with the cycle-1 probe script
(`probes.sh` from `TASK-260906-1uf713_review-probes-stage-c-1.tgz`) against
`0bcea201`:

```
B1  profile use other  -> exit 1 (refused);  install --use -> exit 1;  current stays acme
B2  use --takeover -> exit 0;  surface is still a symlink: no;  foreign file overwritten: no
B3  import -> exit 0;  imported module carries a generation header: no
B4  import -> exit 1 (environment_import_lossy)
```

**B1, the seam and its enumeration.** The enumeration in the comment is true and
I checked it structurally rather than taking it: `CurrentFile(home)` has exactly one
production writer, `switch.go:235`, inside `useLocked`; `SetCurrent` is exported but
has no production caller; `UseWithPolicy` has one production caller; `--clear`
requires `--env` or `--target` at the CLI, so the ungated `clearScope` branch is not
reachable with an empty scope. Driven individually: `import --as imp --use` under the
lock is refused at the seam (exit 1, `environments.require_current_profile`); a
scoped switch (`use other --env codex_cli`) is unaffected, exit 0; `profile use ghost`
reports the configuration error, not `profile_not_found`, so the ordering claim holds;
`profile sync` moves no pointer and is correctly ungated. `Update` reaches the seam
through `resyncCurrentScopes` only when the lock moved, and refuses fail-closed there;
the operator is not stranded, since `profile use <required>` always converges.

**B2 and the write sweep.** Mutant killed (below). The strengthened
`TestForeignSymlinkStopsSwitch` asserts the post-state (no longer a symlink), the
foreign bytes (`"foreign\n"` intact) and that a backup generation exists — not
`result.OK`. I swept every `os.WriteFile`/`os.Create`/`os.OpenFile` in
`internal/envprofile`: the backup writer writes only into a freshly created numbered
generation guarded by an `Lstat` existence check; `writeStoreDocument` and the
profile-record writers stay under the manager home; `copyLinkFallback`'s target is
already removed by `replaceLink`; `purgeHomes` only removes. **No write in the stage
can land on a path the manager does not own.** The three `applyPlan` asymmetries are
C2-m4.

**B4's decision.** The chosen outcome — one `requires.skills` entry, each divergent
drop recorded as a loss, the import stopping under the consent gate — is the safe one
and nothing is silently lost. `deduplicateSkills` is deterministic (`detectNative`
sorts by env-id then name before it runs) and an identical repeat collapses with no
loss, which is right. I do not agree that the reassembly sentence *decides* it:
§9.6's loss list is a closed enumeration of four conditions and a divergent
cross-adapter skill is in none of them, while the classification rule ("a skills entry
maps when the manager can recover a complete exact declaration") is satisfied by both
copies. The honest description is a **spec gap** — §9.6 does not contemplate one skill
name in two adapter homes — resolved fail-closed. Not a finding, because the behaviour
is right and the alternative was the cycle-1 blocking defect; noted so the next report
does not present a contested derivation as settled.

**M2 / m3, the path diagnostics.** All four §1.1 conditions driven at the CLI, each
with the specific diagnostic leading and no doubled prefix:

```
unreadable root (chmod 000)   profile_source_path_unreadable: path "…" cannot be read: …
unreadable dir below root     profile_source_path_unreadable: cannot read …/context/sub: …
unreadable file below root    profile_source_path_unreadable: cannot read context/b.md: …
missing operand               profile_source_path_missing: path "…" names no existing entry
symlink / fifo / nested .git  profile_source_invalid: … (single prefix)
root-level .git               excluded from the snapshot, install exit 0, no leak in the store
```

**The declared survivor, re-verified rather than accepted.** I applied increment 2's
M3 mutant myself (collapse absence into unreadable in `contextstore.EnsureState`) and
ran `./internal/contextstore/ ./internal/envprofile/` — **it survives**, so the
rework's re-run is truthfully reported. I also checked the rationale for an addressing
mode: both production callers of `EnsureState` (`stateForPath` from `Install`, and
`resolveOverlay`) call `pathManifestDiag`/`LoadManifest` first, and `ensureDefault`
passes a staging directory it just created, so the store's *missing* branch is indeed
TOCTOU-only. The bound is honest. (`contextstore` has no test of its own for either
branch of its documented §1.1 contract; cheap to add, not a finding.)

**M3.** `env status` reports `require_current_profile: acme (locked)` in text and
`"require_current_profile": "acme"` / `"require_current_profile_locked": true` in
`--json`, and `(unlocked)` for a machine-file value. Driven through `run()`.

**m4.** The inert-declaration warning fires on stderr only, so machine consumers of
`compose list` stdout are unchanged. Reasonable.

**CLI rows against `cli/curator.md` at `550579d`.** `--takeover` is accepted on
exactly the five §9.5 operations and refused (exit 2) on `profile import`,
`env resolve` without `--repair`, `env status`, `profile list`, `profile remove`,
`profile compose` and `env config`. `profile import [--as] [--allow-lossy] [--use]`,
`env config show|set|unset` and `profile compose add|remove|list` all match their rows.

**The consent gate.** `AllowLossy` is assigned only from the CLI flag
(`cmd/curator/profile.go:63`) and `Policy.Takeover` only from the five flag sites;
`PolicyFromConfig` sets neither. Machine configuration cannot pre-record consent.

**Windows/POSIX sweep of my own.** Beyond C2-m1 I found no POSIX literal fed to a
`filepath` operation, no hardcoded `/` in a path comparison, and no new
`$HOME`-relative lookup in the rework delta; the new fixtures build over `t.TempDir()`
with `filepath.Join`. One thing I could **not** establish on macOS and therefore report
as unknown rather than infer: on a platform that takes no symlink, managed skills are
materialised by `copyLinkFallback` as copies rather than store symlinks, and
`readSkillsSurface`'s managed-entry guard is `isStoreEntry` (a symlink test). Whether
such copies are then detected as native skills — most likely as losses, since a copy
carries no recoverable declaration — has no probe here.

**Root artifacts and the ledger.** No rework test reads a conformance artifact
unguarded (`git diff 833918d2..HEAD --name-only | grep _test.go | xargs grep -l
CONFORMANCE_ROOT` is empty); `internal/config` remains registered in
`root-artifacts.tsv` with all three artifacts; the six new ledger rows are present and
`ledger-consistency.sh` accepts them.

---

## My mutants

Applied to a throwaway copy, never to the worktree. Each keeps the gate present and
weakens it to admit exactly one member.

| # | Mutant | Admits | Result |
|---|---|---|---|
| R1 | `useLocked`: gate only when `Current(home) != ""` | first-install auto-activation | **killed** — `TestProfileInstallFirstActivationLockedRequireRefuses`, `TestProfileInstallUseLockedRequireRefuses`, `TestProfileUseLockedRequireRefuses` (all via `run()`) |
| R2 | `materializeOne`: skip `os.Remove` when the target is a symlink | the foreign-manager symlink | **killed** — `TestForeignSymlinkStopsSwitch`: "takeover left … a symlink" |
| R3 | `readRootSurface`: managed skip `&& envID != "codex_cli"` | one adapter | **killed** — `TestImportSkipsManagedRootContext` |
| R4 | `deduplicateSkills`: no loss when `winner.git == skill.git` | same source, divergent revision | **killed** — `TestImportDivergentSkillsAreLoss` |
| R5 | `parseSystemEnvironments`: `&& key != "forms"` (cycle-1's survivor) | `forms` | **killed** — `TestSystemV2LockableSubsetIsClassWide/non-lockable/forms` |
| R6 | `parseSystemEnvironments`: `&& key != "secret_material_waivers"` | the audit waiver list | **SURVIVES** → C2-M1 |
| R7 | `parseSystemEnvironments`: `&& key != "xdg_seed_allowlist"` | the XDG seed allowlist | **SURVIVES** → C2-M1 |
| R8 | `parseSystemEnvironments`: `&& key != "in_place_mode"` | the in-place mode map | **SURVIVES** → C2-M1 |
| R9 | `EnsureState`: collapse absence into unreadable (inc-2 M3 re-run) | absence | **SURVIVES** — reasoning verified, TOCTOU-only, honestly declared |

C2-B1 is not a mutant: it is a defect driven through the production CLI on the
unmutated tree.

---

## Judging the rework report against the attestation standard

The standard cycle 1 set — every "driven" row driven from the production entry point,
every declared bound either fixed or still declared, the gate table honest — is met by
`TASK-260906-1uf713_rework-report-1.md`, with the two exceptions above:

- The `path`-overlay row is correctly demoted from driven to a stated bound naming
  TASK-260906-3x0w4y; the ratio is stated as a ratio (14 of 15) rather than prose; the
  ledger row that stated the opposite of its test is fixed. The bound C2-m3 leaves is
  a wording gap in row 304, not a re-inflated claim.
- Increment 1's dropped `env status` bound is implemented rather than re-omitted.
- The inc-2 M3 survivor re-run is reported truthfully — I re-ran the same mutant and
  got the same result, and the "TOCTOU-only" rationale holds under an addressing-mode
  check.
- The gate table's CI row ("Hosted CI was not consulted: nothing pushed") was honest
  for `7997d97b`. It is stale for `0bcea201`, which is not the producer's fault: the
  orchestrator pushed and added the last commit. For the record, that head is now
  fully green — see the lane table above.
- The Windows sweep is the weak part. It sweeps fixture shapes competently but never
  reaches the question the orchestrator finding actually asked (C2-m1).

---

## Gate table (what I ran, as standalone processes, on `0bcea201`)

| Gate | Exit | Output |
|---|---|---|
| `go build ./...` | 0 | empty |
| `go vet ./...` | 0 | empty |
| `gofmt -l cmd internal` | 0 | empty |
| `golangci-lint run ./...` | 0 | `0 issues.` |
| `bash .github/ci/gate-selftest.sh` | 0 | `94 passed, 0 failed` |
| `bash .github/ci/no-broad-suppression.sh` | 0 | `no-broad-suppression: ok` |
| `bash .github/ci/ledger-consistency.sh /tmp/rev2-evidence` | 0 | `191 rows checked across linux darwin windows`, `ok` |
| `go test -count=1 ./internal/envprofile/` (unmutated baseline) | 0 | `ok … 52s` |
| `go test -count=1 ./internal/config/`, with and without `CURATOR_CONFORMANCE_ROOT` | 0 | `ok` |
| `go test -count=1 ./internal/contextstore/` | 0 | `ok` |

**Not run by me, stated plainly:** the two `test-gate.sh` lanes (materialized
`SPEC_PIN` root, and curator-spec main with `CI_REQUIRE_FULL_ROOT=1`) and
`go test -timeout 30m ./cmd/curator` in full. Reasons: they cost roughly an hour of
serialised wall clock (they must not run concurrently or beside a `-race` suite), the
hosted lanes execute the same suites on this exact head, and neither finding above
turns on them. The producer's two-root table for `7997d97b` stands unchallenged; the
delta from `7997d97b` to `0bcea201` is two test files setting `USERPROFILE`. `go test
./cmd/curator` was exercised through targeted `-run` masks for every mutant that
touches a CLI row.

## Rework guidance

C2-B1 first: it is a one-line fail-closed change plus the test that was missing — drive
`profile import` with an unreadable marker and assert the surface is neither detected
nor silently dropped. C2-M1 wants a gate, not a case: two cycles have now found a
knob the hand-written list forgot. C2-m1 wants a sentence, not code: decide from §9.5,
or declare the spec gap. C2-m2 and C2-m3 are deletions and wording.
