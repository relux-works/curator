# TASK-260906-1uf713 — review findings, stage (c) cycle 3 (rework 2)

Verdict: **CHANGES REQUESTED**. Two blocking, one major, three minor.

`repeat-of:` **cycle-1 B1** (finding C3-B1 below — a second machine-scope switch that
reaches the §9.2 activation seam without the §12.2 gate, which the rework-1 brief said
in terms would be a repeat) and **cycle-1 m1 / cycle-2 C2-M1** (finding C3-M1 — the
§12.2 lockable subset, still pinned by one hand-written knob name, now one level up in
the derivation source). That makes `parseSystemEnvironments`/`LockableEnvKeys` a
**third consecutive cycle** on the same class: the next step there is a pin that does
not derive from the map it is meant to protect.

- Subject: `~/Developer/ReluxWorks/.worktrees/curator-stage-c`, branch
  `feat/agent-environments-stage-c` at `4df4d5079141e1a034c457e10ef13ffbc7ad2a42`,
  21 commits past `origin/main`. PR head confirmed with `gh pr view 61 --json
  headRefOid` = `4df4d507…`, so the hosted lanes run this exact tree. The producer
  worktree was read-only: `git status --short` empty and HEAD unchanged before and
  after.
- Authority: curator-spec `550579d` (this story worktree) — `protocol/environments.md`
  §1, §6, §8.4, §9.1, §9.5, §9.6, §12.1, §12.2; `profiles/manager.md` §1; `cli/curator.md`.
- Method: a `curator` binary built from `4df4d507` (`go build ./cmd/curator`, exit 0)
  and invoked as a standalone process against scratch homes, with `CURATOR_CONFIG`,
  `CURATOR_SYSTEM_CONFIG`, `CLAUDE_CONFIG_DIR`, `CODEX_HOME`, `XDG_CONFIG_HOME`,
  `PI_CODING_AGENT_DIR`, `GIT_CONFIG_GLOBAL` and `HOME`/`USERPROFILE` pinned to the
  scratch tree. My narrowing mutants were applied to a throwaway copy
  (`/tmp/review3/mut`), never to the worktree. Reproducer:
  `TASK-260906-1uf713_review-probes-stage-c-3.tgz`.

---

## The Change Request's empty repository delta

`CR-TASK-260906-1uf713-3` revision 3 reports `repository_delta: empty`: candidate tree
`4829dd34` is identical to base `550579d1` and the patch is zero bytes
(sha256 `e3b0c442…` is the empty string). **That is structurally correct for this leaf
and is not what the verdict turns on** — the same was true of revisions 1 and 2, both
of which were reviewed on their merits.

The story workspace is a *curator-spec* worktree and base `550579d1` is the
curator-spec authority commit. This leaf's declared scope is "curator repository,
branch `feat/agent-environments-stage-c` in worktree
`/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-c`" — a different repository,
delivered on PR #61 — and every producer brief in this chain forbids writing into the
control root. A non-empty delta here would itself have been a scope violation. The
reviewable work is `git diff origin/main..HEAD` in the curator worktree (the stage) and
`git diff 0bcea201..HEAD` (rework 2: 6 files, +255/−80), which is what I reviewed.

So the emptiness is right. I am rejecting the revision for C3-B1, C3-B2 and C3-M1,
which live in the curator tree.

---

## Hosted lanes on the reviewed head

`gh pr checks 61` on `4df4d507`, run 34041073632, read at the moment this verdict was
recorded:

| Lane | Result |
|---|---|
| Lint | pass 50s |
| Naming gate | pass 7s |
| Interop conformance gate | pass 20s |
| Gate self-test (ubuntu / macos / windows) | pass 7s / 10s / 26s |
| Test (ubuntu-latest) | pass 3m16s |
| Test (macos-latest) | pass 9m16s |
| Race (ubuntu-latest) | pass 9m31s |
| Race (macos-latest) | pass 19m25s |
| Test (windows-latest) | **still running** |
| Candidate suite | skipping (expected: gated matrix) |

**I cannot state that every hosted lane was green on the head I reviewed**:
`Test (windows-latest)` was still running when I recorded. Every other lane is green.
I make no claim about the Windows lane in either direction, and the next cycle should
read it before accepting. Nothing in this verdict rests on it — all three substantive
findings below are invisible to every lane, which is the standard the cycle-3 brief set.

---

## BLOCKING

### C3-B1 — `profile use <name> --clear` reaches the §9.2 activation seam and skips the locked `require_current_profile`

`repeat-of: cycle-1 B1`

The B1 fix moved the gate to the activation seam and its comment claims the
enumeration is complete:

```go
// internal/envprofile/switch.go:200-208
// Every machine-scope switch — profile use, install --use,
// first-install auto-activation, import --use, and resync —
// funnels through this seam, not through any one CLI row. A
// scoped switch records a scoped current and is unaffected.
if scope == "" {
    if err := policy.CheckMachineUse(name); err != nil {
```

That gate is inside the `else` of `if clearScope`. `CurrentFile(home)` has exactly one
production writer — `switch.go:235`, which the cycle-2 review correctly identified —
but that writer fires for `scope == ""` in **both** branches. The `clearScope` branch
reaches it ungated.

Cycle 2 checked this and concluded "`--clear` requires `--env` or `--target` at the
CLI, so the ungated `clearScope` branch is not reachable with an empty scope". That is
true only for the zero-positional form. `cmd/curator/profile.go:203-206` reads

```go
if len(positional) == 1 {
    name = positional[0]
} else if len(positional) > 1 || !*clearScope || (*env == "" && *target == "") {
```

so `profile use <name> --clear` — one positional, no `--env`, no `--target` — parses,
and `useLocked` then takes the `clearScope` branch with `scope == ""`.

Driven through the production CLI on a machine with the key locked to `acme`:

```
$ curator profile use other                 -> exit 1
    environments.require_current_profile: profile use of "other" is refused: …
$ curator profile use bogus --clear         -> exit 0
    claude_code: switched   codex_cli: switched   opencode: switched   pi: switched
$ cat <home>/profiles/current
default
```

The machine scope is materialised for `default` and the machine current pointer is
written to `default`, while §12.2 requires `acme`, at exit 0 with no warning. The same
end state through the documented row is refused, which is the cleanest statement of the
bypass:

```
$ curator profile use default        -> exit 1
    environments.require_current_profile: profile use of "default" is refused: …
$ curator profile use whatever --clear -> exit 0 ;  cat profiles/current  ->  default
```
 §12.2 is
explicit — "A system file that locks `require_current_profile` to a profile name makes
`profile use` of any other profile in the machine scope a configuration error" — and
`default` is another profile. `overlays_allowed` and `require_current_profile` are the
entirety of revision 1's fleet-policy surface, so this is the second bypass the
rework-1 brief named in advance as a repeat finding.

Two further facts about the same line:

- the operand is ignored entirely — `bogus` is not installed, and the `readSource`
  installed check is in the same `else` the gate is in, so an unknown name switches the
  machine scope at exit 0;
- `cli/curator.md` at `550579d` defines exactly two `profile use` forms, and
  `profile use <name> --clear` with no scope is neither. Row 34 is
  `curator profile use --clear --env <env-id>|--target <target-id> [--takeover]`. The
  CLI admits a form the row does not define, and that undefined form is the one that
  bypasses the gate.

Fix at the seam, not at the CLI row — the seam is what the B1 fix chose and the choice
was right. Either the gate covers the `clearScope` branch when `scope == ""`, or that
combination never reaches the seam at all. Then re-do the enumeration honestly: the
comment above the gate is currently false, which is the C2-m2 "comment that lies about
reachability" shape sitting directly over the hole. No committed test drives
`profile use <name> --clear`; `cmd/curator/profile_test.go:318` and
`internal/envprofile/envprofile_test.go:663` both pass `--env`.

Narrowing mutant to apply and quote once fixed: make the new gate admit exactly the
`clearScope && scope == ""` path, and show a `run()`-level test failing.

### C3-B2 — `profile update` re-reads a `path` source, so the immutable snapshot is not immutable, and `profile update --all` is broken for every imported profile

§1 is unambiguous: "Installation copies the directory's tree into the profile store as
an immutable snapshot **and never reads the source directory again**: later edits to
the source directory change nothing **until the operator reinstalls**, and nothing
about a `path` source is ever fetched from a network."

`updateLocked` (`internal/envprofile/envprofile.go:766-776`) does exactly what that
sentence forbids:

```go
case KindPath:
    manifest, err := contextpkg.LoadManifest(source.Path)
    …
    state, err := stateForPath(home, manifest.Name, source.Path)
```

Compare the `KindGit` arm eleven lines above, which short-circuits an exactly-pinned
source and returns the old lock unchanged, and the `default` arm, which refuses the
`local` root with `DiagUpdateBlocked` because it "does not move". A `path` root is
exactly pinned by construction — §1 makes `range`, `tag`, `branch`, `revision` and
`directory` on a `path` declaration `profile_source_invalid` — so it belongs with those
two, not with a re-resolvable git range.

Driven through the production CLI:

```
$ curator profile install ./pathsrc --as pk        # state sha256:ff6c35c7…
$ printf 'EDITED AFTER INSTALL\n' > ./pathsrc/context/a.md
$ curator profile update pk                        -> exit 0
$ # lock member pk state_sha256                     sha256:43d9f269…
```

The pin moved on an edit to the source with no reinstall. `profile sync` correctly does
*not* re-read (I re-drove cycle 1's check and it holds), and `profile install <same
path>` correctly does (`updated profile pk2 …` — that is the reinstall §1 sanctions).
`profile update` is the one command that has it wrong, and it is the one command §1's
sentence excludes.

The second consequence is operational and immediate. §9.6 reassembles into a staging
directory inside the machine home and `importLocked` deletes it (`defer os.RemoveAll(stage)`),
so every imported profile's recorded `source.Path` points at a directory that no longer
exists. Because `update` re-reads it:

```
$ curator profile import --as imported            # profile installed, marker imported_from_native
$ curator profile update --all                    -> exit 1
    imported: profile_source_path_missing: path "…/.import-2105254010" names no existing entry
```

One onboarding import bricks `profile update --all` for the whole machine — the routine
maintenance command — and it does so for the profile §9.6 exists to produce. This is
not a corner: `--all` is the documented fleet form of the row.

Decide the fix from §1: the `KindPath` arm must resolve the root from the snapshot the
store already holds under the old lock's `state_sha256`, not from `source.Path`. Note
that overlays must still re-resolve after that — `resolveOverlays` runs below the
switch and a `path` root with git overlays is a legitimate `update` — so "return the
old lock" is not the whole answer.

The evidence gap under this is the AC row itself. "The path source kind snapshots
immutably" is a definition-of-done row, and **no committed test drives it**:
`pathkind_test.go` has nine tests, all refusals and the `.git` exclusion, and none edits
a source after install. Cycle 1 verified immutability by hand against `profile sync`
only; the untested addressing mode is where the defect lives, which is the same shape as
stage (a) F1, cycle-1 B3 and cycle-2 C2-B1. The replacement test must edit the source
and assert the pin across `update`, not only `sync`.

---

## MAJOR

### C3-M1 — the §12.2 lockable **set** is pinned by one hand-written knob, and widening it by one line reaches §9.1 secret material with a fully green suite

`repeat-of: cycle-1 m1, cycle-2 C2-M1`

The consumer gate is genuinely derived now, and I confirmed it properly. I applied
eight narrowing mutants to `parseSystemEnvironments` — the four the brief names plus
four of my own choosing (`system_prompt_files`, `shadow_acknowledged`, `scoped_current`,
`backup_retention`) — and **all eight die**. That half of C2-M1 is fixed, and the
payload-count guard against `EnvKnobNames` is a good pin.

But the derivation moved the unpinned list rather than removing it.
`TestSystemV2LockableSubsetIsClassWide` decides which branch a knob takes with
`if LockableEnvKeys["environments."+knob]`, so it derives its expectations from the very
map that encodes §12.2. A change to that map is invisible to it by construction. The
only other pin in the repository is `environments_test.go:300`,
`TestSystemV2Refusals/unlockable_knob_carried`, and its knob is `current_profile` —
one hand-written name, which is precisely what cycles 1 and 2 rejected one level down.

Widening mutant, one line, gate present and admitting exactly one member:

```go
// internal/config/environments.go — LockableEnvKeys
    "environments.isolation":               true,
+   "environments.secret_material_waivers": true,
```

| mutant | `go test ./internal/config/` | with `CURATOR_CONFORMANCE_ROOT` |
|---|---|---|
| `LockableEnvKeys += secret_material_waivers` | **ok (SURVIVES)** | **ok (SURVIVES)** |
| `LockableEnvKeys += current_profile` | killed (`TestSystemV2Refusals/unlockable_knob_carried`) | killed |

The conformance root does not help: the published `system-config-v2` family carries
`invalid-unlockable-environments-knob.json` and
`invalid-locked-unlockable-environments-key.json`, and both use `current_profile`. So
the entire §12.2 closed set is anchored, in curator and in the root, by exactly one
counter-example.

Driven end to end with the mutant built, on the same shape my predecessor used:

```
# stock 4df4d507
$ curator env config show
curator: environments.secret_material_waivers: is not lockable and must not be
         carried by the system file (system config …/system.json)          # correct

# mutant
$ curator env config show | grep -A3 secret
  "secret_material_waivers": [
    { "file": "context/a.md", "pin": "0123…", "reason": "org policy", …    # injected
```

`mergeSystemEnvironments` rule 3 then carries it into the effective machine
configuration. Manager §1 states the rule this map is the sole implementation of — "no
key that selects or constrains credential material is lockable" — and
`secret_material_waivers` and `xdg_seed_allowlist` are its most load-bearing members.

There is an attestation half. The test's own doc comment claims the opposite of what it
does:

```
// … so a knob added to the reader without a payload — or to LockableEnvKeys
// without §12.2 standing — fails here instead of widening the system file's
// reach silently.
```

The second clause is false, and I proved it false in one line. That is the same
comment-lies-about-reachability shape C2-m2 was raised for, now in the test that is
supposed to be the gate.

The fix is a pin that does not derive from `LockableEnvKeys`: transcribe §12.2's six
keys as a literal in the test and assert set equality with the map (the literal is the
spec sentence, the map is the implementation, and either drifting fails), or drive the
membership from the published `system-config-v2` cases. Prove it by re-applying both
widening mutants above and both directions — a knob added and a knob removed — and
showing each dies.

---

## MINOR

- **C3-m1 — `profile compose add` is silent under a locked `overlays_allowed: false`.**
  `compose list` prints the manager §1 inert warning (`compose.go:164`) and resolution
  correctly empties the list — I drove it: with the lock in place `profile update acme`
  drops the overlay and `leaf` reverts from the overlay weight 40 to the rule-3 weight
  30 with `overlay=false`. But `compose add` under the same lock exits 0 with
  `added overlay … the lock moves on profile update` and no indication the declaration
  will be inert. C2-m4 asked for exactly this kind of asymmetry to be closed or
  explained; the same reasoning that justified the `list` warning applies to `add`.
  Judgement call: warn, or state why the write row is right to stay silent.
- **C3-m2 — no committed test drives all four §6 weight rules disagreeing at once.**
  `TestWeightRulesApplyInOrder` covers rules 1→2 and 1→3 pairwise, and the overlay tests
  cover rule 4 against the manifest weight and the machine default. Nothing covers
  2-versus-3, or 4 overriding 3. I built the case and drove it through the CLI
  (`profile install` → `profile compose add` → `profile update`) with a leaf whose
  manifest weight is 5, two agreeing non-root requirers at 20, the root's `weights` map
  at 30 and a machine overlay at 40: the lock records **weight 40, `overlay: true`,
  `required_by: [mid1, mid2]`**, which is correct. The implementation is right; the
  suite would not notice it stopping being right.
- **C3-m3 — `TestTakeoverWarnsDotfileHeuristic` drives `UseWithPolicy`, not `run()`.**
  That is a production package entry point the CLI calls directly with the same
  arguments, so it is not the uncalled-gate shape and I am not treating it as a hole —
  noting it because the §9.5 heuristic is the one surface in this stage whose only
  driver is below the CLI.

---

## What I verified and found correct

**C2-B1, the corrupt marker, as a class.** Three distinct outcomes, each driven through
the production `profile import` against a machine curator manages (4 markers written):

```
absent marker      surface detected as native; installed as a path profile at 1.0.0
                   with a state-hash pin (exit 1 is profile_use_partial from the
                   activation over the unmanaged native file, not a refusal)
intact marker      exit 0; managed root imported as native: no
corrupt marker     exit 1  environment_import_lossy: 1 detected surfaces cannot be
                   carried over: claude_code …/.agent-environment.json: cannot be read:
                   environment_marker_invalid: malformed protocol JSON …
                   managed root imported as native: no
unreadable (000)   exit 1  same, "permission denied"
--allow-lossy      exit 0, the loss re-reported as a warning, and **no claude_code
                   module reassembled** — I checked the store: the profile has no
                   modules at all
```

The §8.4 distinction is intact in all four. The producer's own mutant
(`err != nil && envID != "claude_code"`) kills `TestImportCorruptMarkerIsLoss` and
`TestImportUnreadableMarkerIsLoss` with the exact bypass bytes printed — a full §5.1
generation header re-detected as native — and I applied a second of my own (collapse
the failed read into `return nil, nil`, i.e. silently drop rather than lose) which kills
both as well. The sweep table in the rework report is accurate as far as I could check
it: `readSkillsLedger` returns the error, `readSkillsSurface` records the ledger file as
the loss, `isStoreEntry` and `recoverSkill` fail toward recovery-then-loss, and
`switch.go:419`, `verifyHome` and `purgeHomes` are fail-closed or conservative. I found
no third site of the class inside this stage's delta. (`listLocked`'s
`machine, _ := Current(home)` swallows a failed read of the current pointer for the
informative `Current:` column — stage (a) code, `git log -S` confirms `99cc576c`, and the
same disposition as `status.go:432/521`: not this stage's finding.)

**C2-M1's consumer half.** Eight narrowing mutants on `parseSystemEnvironments`, all
killed — see the table below. The gate is genuinely class-wide at that level. The
finding above is one level up.

**C2-m2.** `Config.CheckMachineUse` is gone from code, tests and docs; the only
remaining `CheckMachineUse` is `Policy.CheckMachineUse` with one production caller at
`switch.go:207`. Nothing lost coverage: `profile use other` under the lock is refused
through `run()` (exit 1, `environments.require_current_profile`). No other function in
the stage delta has zero production callers — I enumerated every func added in
`overlays.go`, `import.go`, `environments.go`, `compose.go` and `envconfig.go` and every
one has a caller. The one comment in the stage that still claims a production path it
does not have is the seam enumeration in C3-B1.

**C2-m1, the stated bound.** The bound is on `dotfileStateDirs` in `managed.go`, names
TASK-260906-vlrjo1, says plainly that no Windows location was invented and why, and is
repeated in the rework report. It is truthful: the mechanism resolves the operator home
through `os.UserHomeDir` and joins portable relative paths, and the list is what is
POSIX-only. The case is not skipped into silence — ledger row 331 keeps
`linux,darwin,windows` with no tolerance and the test asserts the warning names
`chezmoi`.

**C2-m3 / C2-m4.** Row 304 now describes what `TestPathOverlayJoinsClosure` actually
asserts, naming the manufactured `Policy` and the TASK-260906-3x0w4y bound; row 303 is
the model it follows. `applyPlan`'s seed and marker writes remove first, and every
target is `filepath.Join(plan.homeDir, …)` where `homeDir` is
`ManagedHomeDir(req.Home, req.Profile, adapter.ID)` — the manager's own tree, so the
change is inert with respect to anything outside it, exactly as cycle 2 said. The
`claudeSeed` exclusion is correctly reasoned in the comment.

**Composition and the precedence primitives.** Beyond C3-m2's four-rule drive: the two
primitives are independent and drive the materialised bytes, checked against the
emitted `CLAUDE.md` rather than an ordering function, with a closure of one weight-40
member and three weight-0 members:

| `winner` | `placement` | emitted chapter order |
|---|---|---|
| higher-weight | winner-last | mid1, mid2, acme, **leaf** |
| higher-weight | winner-first | **leaf**, mid1, mid2, acme |
| lower-weight | winner-last | **leaf**, mid1, mid2, acme |
| lower-weight | winner-first | mid1, mid2, acme, **leaf** |

Ties keep topological order in all four, and the §5.1 generation header records
`precedence: winner=… placement=…` in each. `environment_composition_invalid` fires for
two overlays sharing a name and for an overlay repeating the root's name; an overlay
naming a package the root *requires* correctly does not (rule 4's own case), because
overlays are seeded before the root's requirements expand.

**The schema-2 surface.** A schema-1 machine file stays valid (exit 0); an unknown
`schema_version` is rejected explicitly (`schema_version: unsupported config
schema_version 7; this config requires a newer tool`); `isolation` locked toward
`isolated` is refused with the system-file path and toward `shared` is accepted;
`env status` reports `require_current_profile: acme (locked)` in text and
`require_current_profile` / `require_current_profile_locked` in `--json`. Manager §1
rule 2 is right in both directions: a machine file that sets `overlays_allowed: true`
under a system lock gets `warning: config key "environments.overlays_allowed" is locked
by …/system.json; the user override is ignored`, and a machine file that leaves the knob
absent gets no warning.

**Import reassembly.** Normalization is exactly §9.6 — `line1\r\nline2\rline3\n\n\n`
reassembles to `line1\nline2\nline3\n`, one trailing LF — and applies only at
reassembly. The manifest is `schema_version` 1, `version` `1.0.0`, `weight` 0, no
`weights`, with one `class: root` module per adapter carrying
`environments: ["<env-id>"]` in ascending environment-identifier order.
`profile_import_name_taken` fires before any write. The marker records
`"imported_from_native": true`. `profile use imported --takeover` emits the §9.5 notice
and lands backups under `.agent-environment-backup/1`.

---

## My mutants

Applied to a throwaway copy (`/tmp/review3/mut`), never to the worktree. Each keeps the
gate present and weakens it to admit exactly one member.

| # | Mutant | Admits | Result |
|---|---|---|---|
| S1 | `readRootSurface`: `if err != nil && envID != "claude_code"` (the producer's) | one adapter's failed marker read | **killed** — `TestImportCorruptMarkerIsLoss`, `TestImportUnreadableMarkerIsLoss` |
| S2 | `readRootSurface`: failed read `return nil, nil` (silent drop, not a loss) | every failed marker read | **killed** — same two tests |
| S3–S6 | `parseSystemEnvironments`: `&& key != "forms" / "secret_material_waivers" / "xdg_seed_allowlist" / "in_place_mode"` | one non-lockable knob | **killed** ×4 — `TestSystemV2LockableSubsetIsClassWide/non-lockable/<knob>` |
| S7–S10 | same, my own knobs: `system_prompt_files`, `shadow_acknowledged`, `scoped_current`, `backup_retention` | one non-lockable knob | **killed** ×4 — same test |
| S11 | `LockableEnvKeys += "environments.secret_material_waivers"` | §9.1 secret-material waivers from a system file | **SURVIVES**, with and without the root → C3-M1 |
| S12 | `LockableEnvKeys += "environments.current_profile"` | the one knob the hand-written case names | killed — `TestSystemV2Refusals/unlockable_knob_carried` |

C3-B1 and C3-B2 are not mutants: both are defects driven through the production CLI on
the unmutated tree.

---

## Gate table (what I ran, as standalone processes, on `4df4d507`)

| Gate | Exit | Output |
|---|---|---|
| `go build ./...` | 0 | empty |
| `go vet ./...` | 0 | empty |
| `gofmt -l cmd internal` | 0 | empty |
| `golangci-lint run ./...` | 0 | `0 issues.` |
| `bash .github/ci/gate-selftest.sh` | 0 | `94 passed, 0 failed` |
| `bash .github/ci/no-broad-suppression.sh` | 0 | `no-broad-suppression: ok` |
| `bash .github/ci/ledger-consistency.sh /tmp/review3/ledger-ev2` | 0 | `194 rows checked across linux darwin windows`, `ok` |
| `go test -count=1 ./internal/config/ ./internal/envprofile/ ./internal/contextresolve/` (unmutated baseline) | 0 | `ok 0.5s / 54.8s / 1.1s` |
| `go test -count=1 ./internal/config/` with `CURATOR_CONFORMANCE_ROOT=<story worktree>/conformance/v1` | 0 | `ok` |

**Not run by me, stated plainly:** the two `test-gate.sh` lanes (materialized `SPEC_PIN`
root, and curator-spec main with `CI_REQUIRE_FULL_ROOT=1`) and
`go test -timeout 30m ./cmd/curator` in full. Reasons: they cost roughly an hour of
serialised wall clock, must not run concurrently or beside a `-race` suite, and the
hosted lanes execute the same suites on this exact head. The rework report's two-root
table for `4df4d507` stands unchallenged and I found nothing that contradicts it.
`Test (windows-latest)` had not finished when I recorded, and I make no claim about it.

**On the rework report's CI row.** It says "Hosted CI was not consulted: nothing pushed,
PR #61 untouched per the brief, so no `gh pr checks` output exists for `4df4d507`". That
was honest when written; the head has since been pushed by the orchestrator and the
lanes above are the current state.

---

## Judging the rework report against the attestation standard

The standard holds for what it covers, with one exception. The finding→resolution table
covers all six; the three marker outcomes are driven through production `Import` and I
reproduced them; the derived-gate mutant table's four kills are real and I added four
more; the C2-m1 bound names TASK-260906-vlrjo1 in code and in prose; the carried bounds
list is unchanged rather than quietly trimmed; the ratio is stated as a ratio (14 of 15)
with the one bound named. The gate table is honest, including the CI row.

The exception is the AC ratio itself. The definition-of-done row is "The path source
kind **snapshots immutably**, pins by state hash, and rejects every
`profile_source_invalid` condition", and increment 2's AC table — carried forward
unchanged into both rework reports — lists the whole row as driven with these entries
under "Path kind (§1)":

```
- root `.git` excluded, state-hash pin: TestRootGitExcluded
```

`TestRootGitExcluded` asserts that a root-level `.git` is excluded from the snapshot.
**No named test drives the immutability half of that row** — that installation "never
reads the source directory again" and that "later edits change nothing until the
operator reinstalls" — in any command, and the behaviour is wrong under `profile update`
(C3-B2). An AC row whose named driver establishes a different clause of the row is an
unchecked AC row. It is the same class as cycle-1 M1 — a row reported as driven that the
production surface does not establish — reached by a different route: not a manufactured
`Policy`, but a named test that covers the neighbouring clause.

---

## Is stage (c) safe to land?

**No, not at `4df4d507`.** Two of the three findings are behavioural defects on the
production CLI:

- C3-B1 defeats one of the two fleet-policy knobs the revision ships. A machine under a
  §12.2 lock can be switched to another profile from a documented command with an
  undocumented flag, at exit 0.
- C3-B2 breaks the immutability guarantee that is the entire point of the `path` source
  kind, and makes `profile update --all` fail for any machine that has onboarded through
  §9.6 — which is the machine every operator of this feature will have.

Neither is a corner case and neither is expensive to fix; both are one arm of one switch
statement plus the test that was missing. C3-M1 is the third cycle on one class and
wants a pin that does not derive from the thing it protects. Everything else in the
stage that I attacked held, including all six cycle-2 findings, the composition weight
rules, both precedence primitives against materialised bytes, and the import's loss list
and reassembly.

## Rework guidance

C3-B2 first: it is one `case KindPath:` arm and the test that should have existed —
edit the source after install and assert the pin across `update`, `sync` and `use`, and
assert `profile update --all` succeeds with an imported profile present. C3-B1 next,
fixed at the seam with the enumeration comment made true. C3-M1 wants a pin that
transcribes §12.2 rather than reading `LockableEnvKeys`, proved by widening *and*
narrowing that map. C3-m1 and C3-m2 are a warning and a test.
