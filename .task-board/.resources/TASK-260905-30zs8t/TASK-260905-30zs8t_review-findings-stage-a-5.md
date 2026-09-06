# TASK-260905-30zs8t — review findings, stage (a) core, cycle 5

Subject: curator worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-a-core`, branch
`feat/agent-environments-stage-a`, head **`b6f00e1a`** — one signed commit ("Stage (a) rework 4:
syntactic operand classification (F14), install --use performs 9.2 switch (F13), policy table over CLI
paths (F15)") on the cycle-4 head `dea3f5ac`; 5 files, +428/−50. Eleven commits on curator main
`a2406dfe`, all `G` `Ivan Oparin <oparin@me.com>`. Authority: curator-spec main `f39f4a9`.

**Change Request note.** `repository_delta=empty` against the curator-spec story workspace is correct
by construction: every brief for this leaf states the story workspace carries an empty delta by design
and forbids writing into the control root. The reviewable work is the curator branch, and that is what
this verdict judges.

Verdict: **CHANGES REQUESTED** — F13, F14 and F15 are genuinely fixed and hold under attack, including
under the three mutants of the producer's table plus one narrowing mutant the producer declined to run.
One new **major** finding: a machine-scope switch (both `profile use` and, newly, `profile install
--use`) overwrites a scoped adapter's surfaces while the scope record and `profile list` keep claiming
the scoped profile. Plus one minor recorded as a follow-up.

Reviewer scratch: the worktree's `.temp/review-5/` (attack scripts `b1.sh`–`b7.sh`, `b_a7/b_a8/b_f9/
b_f10/b_f12a/b_f12b.sh`, mutant tree, range differential, schema validation, platform-case evidence).
Read-only on tracked files; the mutant tree is an `rsync` copy under `.temp/`, restored byte-identical
afterwards. `git status --short` in the worktree is clean.

---

## Rework verification — F13, F14, F15, reproduced against this head

Reproduced through the **built CLI binary** (`.temp/review-5/curator`, built from `b6f00e1a`) with real
machine configuration and four real adapter homes, not through the library or the producer's tests.

### F14 — FIXED (verified, eight shapes, `b1.sh`)

`isPathOperand` (`internal/envprofile/envprofile.go:833`) no longer stats anything. `grep` over
`internal/envprofile` finds exactly four `os.Stat`/`os.Lstat` sites (`envprofile.go:220` profile-record
probe, `:1060` `loadMachinePolicy`, `gitsource.go:88` repo-dir cache, `switch.go:388/398/496/509`
surface bookkeeping) — none on the classification path.

```
1. cwd contains a planted package at ./github.com/evil-org/pkg, allowlist locked to github.com/relux-works
   curator profile install github.com/evil-org/pkg --as a1
   -> rc=1  profile_source_invalid: source github.com/evil-org/pkg is outside the machine's allowed sources
   -> no source record, profile-repos/ absent (the refusal precedes the clone)   [the cycle-4 a9.sh shape]
2. /tmp/definitely-not-here-5 | ./missing-relative-5 | ../missing-parent-5  with --range '^1.0.0'
   -> rc=1  profile_install_ref_conflict: a path operand takes no requirement flag or --directory
6. operand "barepkg" with ./barepkg an existing package directory
   -> classified git; clone attempted and fails. Syntactic, not probed.
7. C:/pkg, C:\pkg, D:, \\host\share\pkg  -> classified path on darwin
8. github.com/relux-works/x (inside the allowlist) -> reaches the git path
```

**Narrowing mutant (producer table row 1), re-run by me on the `rsync` copy:** keep the syntactic rule,
re-admit the `os.Stat` directory fallback for operands the syntactic rule does not claim.

```
--- FAIL: TestOperandShadowedByDirectoryResolvesAsGit
    envprofile_f13f14_test.go:138: shadowed operand err = <nil>, want profile_source_invalid
```

KILLED.

### F13 — FIXED (verified, six shapes, `b2.sh` + `b6.sh`)

`installLocked` (`envprofile.go:503-518`) routes activation through `useLocked` under the same held
operation. Driven through the CLI with four real adapter homes:

| # | Scenario | Result |
|---|---|---|
| A | fresh machine, first install (no `--use`) | rc=0; all four adapters `switched`; `current=alpha`; markers and `CLAUDE.md` carry alpha. §9.1's "the activation is reported, never silent" now holds in substance |
| B | `install beta --use` on a machine current on alpha | rc=0; four `switched` lines; `current=beta`; every marker and every home's bytes are beta; `profile list` agrees |
| C | claude home replaced with a regular file, `install gamma --use` | rc=1; `claude_code: mkdir …: not a directory`, three `switched`; `installed profile gamma (lock …)`; `profile_use_partial`; **`current` stays beta**; gamma's lock written and listed non-current — exactly §9.2's "the successfully switched entries are reported as `profile_use_partial` (non-current…)" |
| D | recovery: `profile use beta` after C | rc=0; scope converges, markers and bytes all beta |
| E | non-activating install while current | unchanged: "installed profile delta …; activate with 'curator profile use delta'"; current and markers untouched |
| F | `profile install … --use extra` | rc=2, usage line — `--use` takes no name |
| G (`b6.sh`) | fresh machine, first install with a broken adapter home | rc=1; `current` stays **empty**; lock written; partial reported; `profile use` after repair converges |

Every lock and marker the new activation path produced is **VALID** under `context-lock-v1` and
`agent-environment-marker-v1` (ajv 2020 against `schemas/v1` at `f39f4a9`; 5 locks, 4 markers).

**Mutants.** Producer row 2 (keep the `activated` claim, drop the switch — the pre-fix pointer-only
shape), re-run by me:

```
--- FAIL: TestInstallUseSwitchesAndAgrees      activation results = 0, want 4
--- FAIL: TestInstallUsePartialLeavesCurrent   install beta --use err = <nil>, want profile_use_partial
```

The producer declined to run the cycle-4 brief's *single-adapter* narrowing mutant, calling it
"subsumed". I ran it (`materializeScope` reports `pi` as `OK` without materializing anything, every
other entry untouched):

```
--- FAIL: TestUseMaterializesLinkedHomes            envprofile_test.go:183: pi marker <nil>
--- FAIL: TestDefaultProfileMaterializesOnFreshHome     envprofile_test.go:661: pi marker <nil>
```

KILLED — by two pre-existing tests rather than the F13 tests. The class "every adapter actually
materializes" is covered; the producer's table row is imprecise about which test kills it, not wrong
about the property. No finding (same disposition as cycle 4's M-B).

### F15 — FIXED (verified by re-running M-D myself)

M-D on the `rsync` copy: both CLI call sites pass `envprofile.Policy{}` instead of
`PolicyFromConfig(cfg)`, `list`/`update` untouched.

```
--- FAIL: TestProfileListMigrationHonoursSystemPolicy/use-default
--- FAIL: TestProfileListMigrationHonoursSystemPolicy/sync
      (migration admitted; the allowlist reason is gone and a clone is attempted)
    /list and /update-default still PASS
```

KILLED, exactly where the finding predicted.

---

## New finding

## F16 — MAJOR — a machine-scope switch overwrites a scoped adapter's surfaces while the scope record and `profile list` keep claiming the scoped profile

`repeat-of: TASK-260905-30zs8t_review-findings-stage-a-4.md#F13` (same class: the manager's recorded
state and the materialized bytes disagree, and nothing surfaces the divergence)

**Where.** `internal/envprofile/switch.go:185` — `useLocked` with `environment == ""` calls
`materializeScope(home, effective, "")`, which iterates the whole `Adapters` registry
(`switch.go:298-318`) with no regard for `ScopedCurrents(home)`. `SyncWithPolicy` (`switch.go:245-270`)
does the machine pass *and then* a scoped pass; `useLocked` has no scoped pass.

**What is wrong.** environments §9.3:

> A scoped switch records a per-scope current profile. `env status` and `profile list` MUST surface
> every scope whose current profile differs from the machine default: a split-brain configuration is
> always visible, never implicit. … A scoped current is cleared in either of two ways …: a scoped
> `profile use --clear`, or a scoped `profile use` naming the profile that is the machine default.

Neither clearing form is a machine-scope switch, so the scope record survives one — and it must mean
something. It does not: the machine switch materializes the machine profile into the scoped adapter's
home, writes a marker naming the machine profile, and leaves the record and the `profile list` column
asserting the scoped profile.

**Evidence (`.temp/review-5/b7.sh`, four real adapter homes, hermetic `insteadOf` fixtures):**

```
1) after 'profile use beta --env codex_cli'   (machine current = alpha)
   scope record env:codex_cli = beta
   profile list row           = env:codex_cli=beta
   codex AGENTS.md bytes      = beta body
   codex marker profile.name  = beta
2) after machine-scope 'profile use alpha'
   scope record env:codex_cli = beta
   profile list row           = env:codex_cli=beta
   codex AGENTS.md bytes      = alpha body     <-- overwritten
   codex marker profile.name  = alpha          <-- overwritten
3) after 'profile install gamma --use'
   scope record env:codex_cli = beta
   profile list row           = env:codex_cli=beta
   codex AGENTS.md bytes      = gamma body     <-- overwritten again
   codex marker profile.name  = gamma
```

The operator scoped `codex_cli` to beta; `profile list` says codex is on beta; the agent launched from
that home reads gamma. Nothing warns. `profile sync` repairs it (`b5.sh` step 5 — and note it prints
`codex_cli: synced` **twice**, once for the machine pass and once for the scope pass, so every sync
burns two backup generations on a scoped home: after `b5.sh` codex holds generations `1 2 3 4 5` where
claude holds `1 2 3`) — but only for an operator who already knows something is wrong.

**Why this is a finding for this cycle.** The defect is on the machine-scope path, which predates this
rework; **this commit newly extends it to `profile install --use`**, which the cycle-5 brief asked me
to drive. It is the same shape cycle 4 ruled blocking as F13 — a recorded state that contradicts the
bytes on disk, silently — moved to the scoped-current axis. It ships a wrong artifact: a marker and a
`profile list` row that contradict each other about which profile a registered environment is on.

**Why no test caught it.** No test anywhere combines a scope record with a machine-scope switch.
`TestScopedUseAndClear` (`envprofile_test.go:487`) records a scope, asserts the machine current did not
move, then clears it. `TestProfileScopedUseAndClear` (`cmd/curator/profile_test.go:220`) is the CLI
mirror. Neither runs an unnarrowed `profile use` (or an `install --use`) while a scope record exists.

**Fix — author's call; the current shape is neither.** §9.2 step 1 read literally ("re-materializes
every in-place surface of every registered adapter") admits the machine pass touching scoped adapters,
and §9.3 requires the record to keep meaning something. So either:
(a) the machine pass skips adapters that carry a scope record — `materializeScope` consults
`ScopedCurrents(home)` for the unnarrowed case, as `SyncWithPolicy` already effectively does in two
passes; or
(b) the machine switch re-materializes them and then re-materializes each scope from its record in the
same operation (one write per home, not two), so the recorded state and the bytes agree when the
command returns.
Whichever is chosen, the manager must not report `env:<id>=<profile>` for a home it has just
overwritten with a different profile.

**Regression cover required.** Driving the CLI and `Use`/`Install`: with `env:codex_cli` scoped to a
non-machine profile, (1) a machine-scope `profile use` and (2) a `profile install --use` must leave the
codex home's bytes and marker in agreement with whatever `profile list` reports for that scope; plus a
NARROWING mutant that keeps the scope-aware machine pass but re-admits the unconditional overwrite for
exactly one adapter — a named test must fail. A third case worth pinning: `profile sync` writes a
scoped home once, not twice (the backup-generation churn above is the observable).

---

## Follow-up, not a cycle (recorded, not blocking)

**FU-1 — §1.1's `profile_source_path_missing` and `profile_source_path_unreadable` do not exist.**
`grep -rn 'path_missing\|path_unreadable' internal/ cmd/` → 0 hits. §1.1 names both and adds the rule
this protocol repeats everywhere: "A missing `path` operand and an unreadable one are different facts …
`profile_source_path_missing` never fires on a failed read, and `profile_source_path_unreadable` never
fires on absence." Observed (`b1.sh` §3–§5):

```
install /tmp/definitely-not-here-5   -> profile_source_invalid: context_manifest_invalid: agent-context.json is absent at …
install ./unreadable  (chmod 000)    -> profile_source_invalid: read agent-context.json: … permission denied
install ./afile       (regular file) -> profile_source_invalid: … not a directory          [correct per §1.1]
```

Both refuse, both fail closed, and the substance of the refusal is right — so this is a diagnostic-name
gap, not a wrong outcome, and it does not meet this cycle's bar. It is newly *reachable in this shape*
because F14 correctly routes an absent path operand to the `path` branch instead of to a clone, so it
belongs in the same package's next touch.

**FU-2 — a partial install whose activation fails before `materializeScope` prints no "installed
profile" line.** `cmd/curator/profile.go:78` gates that line on `len(info.Activation) > 0`; if
`useLocked` returns early (`readSource`/`loadMaterial` error) `results` is nil, so the operator sees
only the error while the lock and source record are on disk. Hard to reach and always recoverable by
`profile list`; noted, not blocking.

---

## Stated bounds — judged

| Bound | Judgement |
|---|---|
| `loadMachinePolicy` returns an empty policy when `config.UserPath()` is absent (rework report 4) | **Defensible.** Absence is legitimate, any other read/parse failure fails the caller (`envprofile.go:1058-1071`), and the branch is not CLI-reachable — `fileConfigSource.Load` fails the command first. The bound is stated on the function doc. Matches the absence-vs-unreadable rule |
| skills tree, MCP files, managed homes, seeds, passthrough, secondary targets are stage (b) (`switch.go` package doc) | **Defensible**, and it covers `cli/curator.md:32`'s "re-point the command shims on a machine-scope switch" (§9.4): shim re-pointing follows the skills tree, which this stage's package doc explicitly defers. Stated, not silent |
| the per-entry agent-home payloads are not transaction targets; entries are direct writes under the held lock, recovery is re-running the switch (`switch.go` package doc) | **Defensible** and demonstrated: 8 SIGKILLs at random points in a switch left `current` unchanged, no broken marker, one journal, and the next command converged (`b3.sh`) |
| MCP package allowlist has no schema-1 machine-config surface | **Defensible**, unchanged from cycles 3–4 |
| `fetchRaw` first-spelling-wins for transport selection | **Defensible**, unchanged. Observed again in `b4.sh`: `http://example.com/org/pkg` installs from the cached repo dir of the earlier `https` spelling — the same canonical identity under core §6.1, which admits `http` |
| F7 scoped waivers have no production path; F2 migration warning | Unchanged; correctly carried to the stage that lands the operator surfaces |

---

## Verified correct — attacked, held

| Surface | Attack | Result |
|---|---|---|
| F14 syntactic classification | 8 shapes incl. the planted-directory shadow, absent absolute/relative/parent operands with `--range`, a bare operand naming an existing package dir, four Windows spellings; `grep` for every `Stat`/`Lstat` on the path | all as §9.1 requires; classifier never stats |
| F13 install activation | 7 shapes through the CLI incl. first install, `--use` over a current profile, forced adapter failure, fresh-machine partial, recovery, non-activating branch, `--use <name>` | §9.2 shape holds in every one |
| F15 policy threading | M-D re-run | `use-default` and `sync` subtests die; `list`/`update` survive |
| F1 detector scope (re-run, `b_a7.sh`) | secret in a `--directory sub` module; transitive `requires.contexts{directory}` member; control with the secret outside the addressed root | both refused `member … carries a blocking context-secret-material finding`; control installs |
| F2 fresh home (re-run, `b_a7.sh`) | isolated home: `profile list`, `use default`, `sync`, `use --clear --env claude_code` | all rc=0 |
| F3 interrupted switch (`b3.sh`) | 8 SIGKILLs at 0.1–30 ms into `profile use`, 3 into `install --use`; journal and backup inspection | `current` never half-moved, no broken marker, journal replays, next command converges; backups pruned to 5 generations |
| F4 skip honesty (re-run) | every `CURATOR_*` name a skip mentions, grepped repo-wide | only `CURATOR_CONFORMANCE_ROOT`, read in 30+ places; `CURATOR_STAGE_B` gone; `stage-deferred` registered at `skip-classes.tsv:106` and `platform-cases.tsv:187` |
| F5/F6 ranges (re-run, **my own 33 fresh cases** × 29 versions vs `semver@7.7.4`) | caret on `0.x`/`0.0.x`, tilde with fewer components, x-ranges, `<1.0.0-0`, `\|\|`, `latest`, `~>`, empty range, `>*`/`<*`/`>x`/`<X`, `^v`, hyphen, build metadata | **23 of 33 exact agreement**; all 10 deviations are §1.4 restrictions (`latest`=`*`; `^v2.0.0`, `1.0.0 - 2.0.0`, `2.0.0+build.7`, `~>1.2.3`, `""`, `>*`, `<*`, `>x`, `<X` rejected — the grammar is closed and names none of them). 0 unexplained |
| F8 canonical identity (re-run, `b4.sh`) | one repository under `https`, scp-style and `ssh://`, three isolated homes | one identical canonical identity `example.com/org/pkg` in all three locks **and** all three markers; all six artifacts schema-VALID; explicit port, whitespace and query rejected at the boundary |
| F9 gates (re-run, `b_f9.sh`) | transitive member outside `allowed_sources`; malformed transitive source; revoked transitive member **with `audit.enabled: false`** | refused, refused, refused — §9.1's "an advisory profile install does not exist" holds |
| F10 migration policy (re-run, `b_f10.sh`) | 4 commands × system-locked `allowed_sources` | all four refuse before any clone; no default lock written |
| F12 `file://` (re-run, `b_f12a/b.sh`) | 5 operand spellings + transitive requirement + control | all refused at the boundary; the dependency is never cloned; the network control installs |
| Lock/marker schema validity | ajv 2020 against `schemas/v1` at `f39f4a9` | 9 locks + 7 markers from every path exercised: all VALID |
| Materialization / lock / marker / versions | `git diff dea3f5ac..b6f00e1a` is **empty** for `internal/{pkgversion,contextmaterialize,contextlock,envmarker,contextpkg,contextaudit,contextstore,interop,contextresolve,identity,audit}` | cycles 2–4's byte-for-byte header, weight-ordering, resolution-replay and marker verdicts carry over unchanged; the vector families re-run green here |
| CLI rows (`b_a8.sh`) | 19-row refusal/acceptance sweep through the built binary | all as `cli/curator.md` names them, **including** the two rows F14 fixed |
| Scope / signatures | `git log --format='%G? %GS' a2406dfe..HEAD` | 11 commits, all `G`, `Ivan Oparin <oparin@me.com>`; no unrelated behaviour change; README profile rows present |

## Gates rerun by the reviewer at `b6f00e1a`

| Gate | Result |
|---|---|
| `go build ./...` | exit 0 |
| `go vet ./...` | exit 0 |
| `gofmt -l cmd internal` | clean |
| `golangci-lint run ./internal/envprofile/... ./cmd/curator/...` | **0 issues** |
| `go test -count=1 -race` on 11 new/touched packages | all `ok` (envprofile 27.9s) |
| interop vectors, `CURATOR_CONFORMANCE_ROOT=<spec>/conformance/v1` | 6 top-level PASS (versions, resolution, header, monolithic, detectors, snapshot-acquisition), 0 FAIL |
| conformance sub-skips | exactly **7**: `referenced-{claude-code-composed,opencode,opencode-zero-modules}` + `mcp-{claude-code,codex-cli,opencode,pi-none}`, every one `stage-deferred` / `tolerated-by-ledger` |
| `bash .github/ci/gate-selftest.sh` | 81 passed, 0 failed (exit 0) |
| `bash .github/ci/no-broad-suppression.sh` | ok (exit 0) |
| `bash .github/ci/ledger-consistency.sh` | ok, 103 rows across linux/darwin/windows |
| platform-case gate, `CI_GATE_GOOS=linux\|darwin\|windows` over a real `go test -json` interop stream with the ledger scoped to `internal/interop` | **ok** on all three; 7 skips recorded and classified |
| `go test -count=1 -timeout 30m ./cmd/curator/` | **run by me**: `ok … 284.994s` |

Reran myself: everything above. Accepted from prior evidence: nothing.

## AC coverage as measured

Rework-4 author decisions: **3 of 3** findings (F13, F14, F15) delivered and independently verified,
each killed by a named test under a narrowing mutant I applied myself.

Producer-brief items against `f39f4a9`: items 1–6 and 8 delivered and verified across cycles 1–5.
**Item 7 (`linked` switching, §8.1/§9.2) is now correct for `profile install [--use]`, `profile use`
(machine and scoped), `profile update`, `profile remove`, `profile sync` and the marker — but the
interaction between a machine-scope switch and an existing scope record leaves the recorded state and
the materialized bytes in disagreement (F16). 7 of 8 brief items fully delivered.**

CLI rows driven by a named committed test: `--use` moves from **0 of 1** to **1 of 1**
(`TestProfileInstallUseActivatesThroughSwitch`, `TestProfileInstallUsePartialLeavesCurrent`, plus the
library pair). The machine-switch-over-a-scope row is **0 of 1** — that hole is F16.

## Reviewer housekeeping

Every probe ran with `HOME`, `CURATOR_CONFIG`, `CURATOR_SYSTEM_CONFIG`, `CLAUDE_CONFIG_DIR`,
`CODEX_HOME`, `XDG_CONFIG_HOME`, `GIT_CONFIG_GLOBAL`, `GIT_CONFIG_NOSYSTEM` and `PI_CODING_AGENT_DIR`
pointed at directories under `.temp/review-5/`. `~/.curator` was never touched and no probe reached the
network (`insteadOf` rewrites onto local repositories throughout; the one outbound attempt was inside
the M-D *mutant*, which is the point of the mutant). The mutant tree is an `rsync` copy at
`.temp/review-5/mutant/`; every mutation was reverted from a saved original and the three mutated files
are byte-identical to the worktree again. `git status --short` in the worktree is clean and `git diff`
against `b6f00e1a` is empty. Nothing was written into the control root.
