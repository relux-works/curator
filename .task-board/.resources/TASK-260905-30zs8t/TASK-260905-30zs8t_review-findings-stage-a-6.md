# TASK-260905-30zs8t — review findings, stage (a) core, cycle 6 (acceptance)

Subject: curator worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-a-core`, branch
`feat/agent-environments-stage-a`, head **`834b40f6`** ("Stage (a) rework 5: machine pass skips adapters
with a scope record (F16)"), one signed commit on the cycle-5 head `b6f00e1a`; 3 files, +377/−10.
Twelve commits on curator main `a2406dfe`, every one `G` and authored **and** committed by
`Ivan Oparin <oparin@me.com>`. Authority: curator-spec main `f39f4a9`.

Verdict: **ACCEPT**. F16 is fixed and holds under my own attack, including a narrowing mutant and the
`profile update` / `SyncWithPolicy` shapes the brief asked me to check separately. The cumulative
"verified, held" set from cycles 1–5 was re-run against this head and every item held. Every gate is
green, including the full suite (70 packages, 0 FAIL). Three follow-ups are recorded below for
TASK-260906-1f2ng0; none of them ships a wrong artifact, bypasses a gate, or leaves recorded state
contradicting the bytes, which is the bar this cycle's brief set.

**Change Request note.** `repository_delta=empty` against the curator-spec story workspace is correct by
construction and is not a failure: every brief for this leaf places the reviewable work in the curator
repository and forbids writing into the control root ("the story workspace carries an empty delta by
design"). I verified it rather than accepting the claim: `git diff f39f4a9 c3e47989` in the story
workspace is zero paths, `HEAD^{tree}` equals the candidate tree OID `c3e47989`, and the rev6 patch is
0 bytes with sha256 `e3b0c442…` (the empty-input digest). The delivered artifact is the twelve-commit
curator branch plus the outcome resources; this verdict judges that.

Reviewer scratch: the worktree's `.temp/review-6/` (`c1.sh`–`c4.sh` written by me; `b1/b2/b3/b4/b5/
b_a7/b_a8/b_f9/b_f10/b_f12a/b_f12b.sh` re-pointed from cycle 5 at a binary built from **this** head;
`mutant/` rsync tree; `ci/` gate evidence; `interop.txt`). Read-only on tracked files. `git status
--short` in the worktree is clean and `git diff` against `834b40f6` is empty.

---

## 1. F16 — FIXED, verified by reproduction, not by report

`materializeScope` (`internal/envprofile/switch.go:304-330`) now joins `ScopedCurrents(home)` in the
unnarrowed branch and skips every adapter carrying an `env:<id>` record; the narrowed branch is
untouched. The join fails closed — a `ScopedCurrents` error aborts the switch rather than defaulting to
the full registry.

Driven through a CLI binary built from `834b40f6`, four real adapter homes, hermetic `insteadOf`
fixtures, isolated `HOME`/`CURATOR_CONFIG`/`CLAUDE_CONFIG_DIR`/`CODEX_HOME`/`XDG_CONFIG_HOME`/
`PI_CODING_AGENT_DIR`/`GIT_CONFIG_GLOBAL`/`GIT_CONFIG_NOSYSTEM` (`c1.sh`, `c2.sh`):

| # | Step (machine current alpha, `env:codex_cli`=beta) | Observed |
|---|---|---|
| 1 | `profile use beta --env codex_cli` | one result (codex); record=beta; codex bytes+marker beta; claude untouched |
| 2 | machine `profile use alpha` | **codex absent from the results**; codex bytes+marker stay beta; three other homes switch; `profile list` = `alpha[current] beta[env:codex_cli=beta]` — record, list and bytes all agree |
| 3 | `profile install gamma --use` | same: codex skipped and stays beta, current moves to gamma, list agrees |
| 4 | `profile sync` | each adapter appears **exactly once**; codex gains exactly one backup generation (1→2), claude one (2→3). The cycle-5 double-write churn is gone |
| 5 | scoped `profile use delta --env codex_cli` | one result, machine current unmoved, other homes untouched |
| 6 | `profile use --clear --env codex_cli` | record removed from disk, codex re-materialized from the machine default, list agrees |

The brief's extra targets, checked as its own shapes (`c2.sh`):

- **`profile update` on a scoped-only profile** (machine=alpha, codex scoped to beta, `update beta`):
  codex moves to beta v2, claude untouched. `resyncCurrentScopes` (`envprofile.go:1075-1099`) takes the
  scoped branch only.
- **`profile update` on the machine current while an adapter is scoped** (`update alpha`): claude moves
  to alpha v2, codex stays on beta v2. The machine branch routes through `useLocked`, so it inherits the
  skip — one code path, no second join to drift.
- **`profile remove beta` while scoped**: `profile_in_use: profile "beta" is current in a scope` (§9.2).
- **Every adapter scoped, machine `profile use gamma`**: all four skipped, no home overwritten, list
  reports `gamma[current]` plus four scope rows. (Silent output — recorded as FU-5.)
- **Crash window on the new path** (`c3.sh`): six `SIGKILL`s at 3–25 ms into a machine `profile use`
  while `env:codex_cli`=beta. In every one the codex home's bytes and marker stayed beta, `current`
  never half-moved, no journal was left behind, and `profile sync` converged. The scoped home is
  untouched by an interrupted machine switch, which is the point of the fix.

**Mutants — I ran both myself on an `rsync` copy under `.temp/review-6/mutant/`.**

| Mutant | Shape | Result |
|---|---|---|
| M-D (the producer's table row): unnarrowed branch restored to `adapters = Adapters` | **deletion** | KILLED — `TestMachineUseSkipsScopedAdapter`, `TestInstallUseSkipsScopedAdapter`, `TestSyncWritesScopedHomeOnce` (library) and `TestProfileMachineUseSkipsScopedAdapter` (CLI) all FAIL with the messages the report predicts |
| M-N (mine): gate present, `skippedOne` bool so the pass skips **at most one** scoped adapter | **narrowing** | **SURVIVED** the entire committed suite — see FU-4 |

Schema validity of everything these paths produced: **17 locks and 16 markers**, ajv 2020 against
`schemas/v1` at `f39f4a9`, all **VALID** — including a marker written on a machine switch that skipped a
home, and the stale marker of the skipped home itself. The claude marker carries the copied-surface
reason (`copies: [{path: CLAUDE.md, reason: claude-code-root-context}]`), `members`, `precedence`
(`higher-weight`/`winner-last`) and `mode: linked` as `agent-environment-marker-v1` requires.

**Verdict on the author's option (a).** Defensible. §9.2 step 1 read literally admits the machine pass
touching every registered adapter, but §9.3 requires a scope record to keep meaning something, and only
option (a) satisfies both without writing a home twice per operation. The implementation states the
choice in four doc comments (package, `Use`, `Sync`/`SyncWithPolicy`, `materializeScope`), so it is a
stated design rule rather than an accident.

---

## 2. Regression — the cumulative cycle-1…5 set, re-run at this head

Every row below was re-executed by me against a binary built from `834b40f6`; nothing is carried over on
the producer's word.

| Cycle | Surface | Attack | Result |
|---|---|---|---|
| F1 | detector scope | secret under a `--directory sub` module; transitive `requires.contexts{directory}` member with a secret; control with the secret outside the addressed root | both refused `member … carries a blocking context-secret-material finding`; control installs |
| F2 | fresh isolated home | `profile list`, `use default`, `sync`, `use --clear --env claude_code` | all rc=0 |
| F3 | interrupted switch | 8 SIGKILLs into `profile use` + 3 into `install --use`; backups and journals inspected | `current` never half-moved, one journal replayed, next command converged, backups pruned to 5 generations (14–18) |
| F8 | canonical identity | one repository under `https`, `scp`-style and `ssh://`, three isolated homes | one identical canonical identity `example.com/org/pkg` in all three locks **and** all three markers; 6/6 schema-VALID; explicit port, whitespace and query rejected at the boundary |
| F9 | §9.1 gates | transitive member outside `allowed_sources`; malformed transitive source; revoked transitive member **with `audit.enabled: false`** | refused, refused, refused — "an advisory profile install does not exist" holds |
| F10 | migration policy | `install`, `list`, `sync`, `use default` × a system config locking `allowed_sources` | all four refuse before any clone; no default lock written |
| F12 | `file://` | operand spellings + transitive `file://` requirement + network control | `file:// git sources carry no network identity and are not accepted`; the dep is never cloned; the network control installs |
| F13 | install activation | fresh install, `--use` over a current profile, forced adapter failure (`profile_use_partial`, previous current survives), recovery, non-activating branch, `--use <name>` refusal | §9.2 shape holds in all seven |
| F14 | syntactic classification | planted-directory shadow under a locked allowlist, absent absolute/relative/parent operands with `--range`, bare operand naming an existing package dir, four Windows spellings, allowed control | all as §9.1 requires; the classifier never stats |
| F4 | skip honesty | every `CURATOR_*` name a skip mentions, grepped repo-wide | only `CURATOR_CONFORMANCE_ROOT`; the stage-(b) skip reason is emitted at `context_materialization_test.go:145` and classified `stage-deferred` at `skip-classes.tsv:106` / `platform-cases.tsv:187` |
| — | CLI rows | 19-row refusal/acceptance sweep through the built binary | all as `cli/curator.md` at `f39f4a9` names them |
| F5/F6, materialization, lock, marker, resolution, versions | — | `git diff b6f00e1a..834b40f6` over `internal/{pkgversion,contextmaterialize,contextlock,envmarker,contextpkg,contextaudit,contextstore,interop,contextresolve,identity,audit,closure}` and `cmd/curator/profile.go` is **empty**; this commit touches exactly `switch.go`, `envprofile_f16_test.go`, `cmd/curator/profile_test.go` | cycles 2–5's byte-for-byte header, weight-ordering, range-differential and resolution-replay verdicts carry over unchanged, and the vector families re-run green here |

---

## 3. Gates rerun by the reviewer at `834b40f6`

| Gate | Result |
|---|---|
| `go build ./...` | exit 0 |
| `go vet ./...` | exit 0 |
| `gofmt -l cmd internal` | clean |
| `golangci-lint run ./internal/envprofile/... ./cmd/curator/...` | **0 issues** |
| `go test -count=1 -race` on 15 stage-A/dependency packages | all `ok` (envprofile 39.6s, transaction 31.6s) |
| interop vectors, `CURATOR_CONFORMANCE_ROOT=<spec>/conformance/v1` | **95 PASS, 0 FAIL**, exactly **7 sub-skips**: `referenced-{claude-code-composed,opencode,opencode-zero-modules}`, `mcp-{claude-code,codex-cli,opencode,pi-none}` — all `stage-deferred`, all in the ledger |
| `bash .github/ci/gate-selftest.sh` | 81 passed, 0 failed, exit 0 |
| `bash .github/ci/no-broad-suppression.sh` | ok, exit 0 |
| `bash .github/ci/ledger-consistency.sh <evidence>` (as `ci.yml:112`) | 103 rows across linux/darwin/windows, ok, exit 0 |
| platform-case gate over a real `go test -json ./internal/interop/` stream, `CI_GATE_GOOS=linux\|darwin\|windows` | **ok** on all three, 7 skips recorded and classified each |
| `go test -count=1 -timeout 30m ./cmd/curator/` | `ok 261.222s`, exit 0 |
| the other 69 packages | 66 `ok`, 3 no-test-files, **0 FAIL** |

Reran myself: everything above. Accepted from prior evidence: nothing.

---

## 4. Follow-ups for TASK-260906-1f2ng0 (recorded, not blocking)

### FU-3 — §9.3's "a scope record equal to the machine default is never kept" is not enforced on a machine-scope switch, and the F16 fix makes the stale record load-bearing

`repeat-of: none`

**Reproduction (`.temp/review-6/c4.sh`, CLI, four real homes):**

```
1) machine=alpha, then scoped 'use beta --env codex_cli'
     current=alpha  record=beta   codex=beta   claude=alpha
2) machine-scope 'profile use beta'
     current=beta   record=beta   codex=beta   claude=beta
     profile list: beta[current,env:codex_cli=beta]      <-- record equal to the machine default, kept
3) machine-scope 'profile use gamma'
     current=gamma  record=beta   codex=beta   claude=gamma   <-- codex now pinned to beta forever
CONTROL (the implemented clearing form): scoped 'use alpha --env codex_cli' with machine=alpha
     record removed, codex re-materialized from the machine default
```

`useLocked`'s machine branch (`switch.go:206-214`) publishes `current` and touches no scope record.
§9.3 says both implemented clearing forms "remove the scope record … ; a scope record equal to the
machine default is never kept". Step 2 keeps one. Before this commit the consequence was cosmetic —
the machine pass overwrote the home anyway. After it, the retained record silently pins that adapter:
the operator scoped codex to beta, then made beta the machine default (which §9.3 says clears the
scope), and codex now stops following the machine default with no command that would naturally reveal
it as unintended.

**Why this is a follow-up and not a cycle.** The state stays self-consistent: `profile list` reports
`env:codex_cli=beta`, the bytes *are* beta, and every artifact is schema-valid — there is no
recorded-state-versus-bytes contradiction, which is this cycle's bar. It is one command to recover
(`profile use --clear --env codex_cli`). And the spec text is genuinely ambiguous: the invariant
sentence sits inside the paragraph enumerating the two clearing *actions*, so it can be read as the
rationale for the second form rather than as a rule binding a third path, and §9.2's `profile use`
steps never mention scope records. That reading is the author's to settle. Fix if it is settled as an
invariant: in the machine branch, fold `SetScoped(home, scope, "", true)` for every scope record equal
to `effective` into the same journaled publish, and cover it with a test at the shape above plus a
narrowing mutant that clears only the first such record.

### FU-4 — the F16 gate ships a delete-only mutant; a narrowing mutant survives the whole suite

`repeat-of: none`

The rework report's mutant table has a single entry, and it deletes the gate (`adapters = Adapters`).
The DoD asks for a mutant that keeps the gate and weakens it to admit exactly one member of the class.
I wrote one — the pass skips **at most one** scoped adapter:

```go
skippedOne := false
for _, adapter := range Adapters {
    if _, ok := scoped["env:"+adapter.ID]; ok && !skippedOne { skippedOne = true; continue }
    adapters = append(adapters, adapter)
}
```

`go test ./internal/envprofile/` → `ok 27.181s`; `go test -run TestProfile ./cmd/curator/` → `ok 8.492s`.
**Survived**, because all five F16 fixtures (`envprofile_f16_test.go:36,…`, `profile_test.go:292`) scope
exactly one adapter — `codex_cli` — so "skips the scoped adapter" and "skips at most one scoped adapter"
are indistinguishable to the suite. The mutant is live, not equivalent: I built it and ran my `c2.sh`
case D against it, and it reproduces the F16 defect exactly — `profile list` reports
`env:codex_cli=beta` while codex's bytes and marker read gamma.

**The shipped code is correct** — I drove the N=4 case through the CLI at this head (`c2.sh` step D:
all four scoped adapters skipped, no home overwritten). So this is a test-coverage hole, not a
behaviour defect, which is why it is a follow-up. Closing it is one line: scope a second adapter in the
`TestMachineUseSkipsScopedAdapter` fixture and assert both are absent from the results.

### FU-5 — a machine-scope switch that materializes nothing prints nothing and exits 0

When every registered adapter carries a scope record, `materializeScope` returns an empty result set;
`profile use <name>` then moves the recorded current with **no stdout at all** and rc=0 (`c2.sh` step
D). The state is correct and `profile list` describes it accurately, but §9.2's "reports per-adapter and
per-target results" degenerates to silence in the one case where the operator most needs to be told
that the switch touched no home. One line: say so when the result set is empty.

### Carried forward unchanged from cycle 5

- **FU-1** — `profile_source_path_missing` / `profile_source_path_unreadable` (§1.1) do not exist; both
  shapes still refuse fail-closed as `profile_source_invalid` with substantively correct text
  (re-observed in `b1.sh` §3–§5 at this head).
- **FU-2** — a partial install whose activation fails before `materializeScope` prints no "installed
  profile" line (`cmd/curator/profile.go:78`). Note the success path is unaffected: `activated` drives
  its own line regardless of `len(info.Activation)`, so the all-scoped case of FU-5 still reports.

## Stated bounds — re-judged at this head

All six bounds carried into this cycle (`loadMachinePolicy` absence-vs-unreadable rule, stage-(b)
surfaces incl. shim re-pointing, the non-journaled per-entry payloads with re-run recovery, the MCP
package allowlist awaiting manager-config v2, `fetchRaw` first-spelling-wins, F7 waivers and the F2
migration warning) are untouched by this commit and remain defensible for the reasons cycles 3–5
recorded. The per-entry-payload bound was re-demonstrated here: `c3.sh`'s six SIGKILLs and `b3.sh`'s
eleven left `current` and the scope record intact and converged on the next command.

## AC coverage as measured

- **Rework-5 author decision: 1 of 1** (F16, option (a)) delivered and independently verified, killed by
  four named tests under the deletion mutant; one narrowing mutant survives (FU-4), and I closed that
  evidence gap myself by driving the N=4 case through the CLI.
- **Producer-brief items against `f39f4a9`: 8 of 8.** Cycle 5 held item 7 (`linked` switching, §8.1/§9.2)
  open at 7 of 8 solely on F16; the machine-switch-over-a-scope interaction is now correct for
  `profile use`, `profile install --use`, `profile update` and `profile sync`, so item 7 closes.
- **CLI rows driven by a named committed test:** the machine-switch-over-a-scope row moves from **0 of 1
  to 1 of 1** — `TestMachineUseSkipsScopedAdapter`, `TestInstallUseSkipsScopedAdapter`,
  `TestSyncWritesScopedHomeOnce`, `TestScopedUseStillSwitchesOnlyThatHome` (library, through `Use`,
  `Install` and `Sync`) and `TestProfileMachineUseSkipsScopedAdapter` (CLI, through `run()`).
- **Blind spot, named:** `env status` is stage (b), so §9.3's split-brain visibility requirement is
  discharged in this stage by `profile list` alone. Unknown at this stage whether `env status` will
  surface the same rows; not inferred from `profile list`.

## Reviewer housekeeping

Every probe ran with `HOME`, `CURATOR_CONFIG`, `CURATOR_SYSTEM_CONFIG`, `CLAUDE_CONFIG_DIR`,
`CODEX_HOME`, `XDG_CONFIG_HOME`, `GIT_CONFIG_GLOBAL`, `GIT_CONFIG_NOSYSTEM` and `PI_CODING_AGENT_DIR`
pointed under `.temp/review-6/`. `~/.curator` was never touched; `insteadOf` rewrites kept every fixture
offline (the one outbound attempt is `b1.sh` case 8's deliberate allowed-identity control, which fails
closed). The mutant tree is an `rsync` copy at `.temp/review-6/mutant/`; its `switch.go` was restored
from a saved original and is `cmp`-identical to the worktree's. `git status --short` in the curator
worktree is clean; nothing was written into the control root, and no LOGBOOK.md was touched.
