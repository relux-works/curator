# TASK-260905-30zs8t — review verdict, CR revision 4 (stage (a) core)

Reviewer run `RUN-260906-fc66d4`. Subject: curator worktree
`/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-a-core`, branch `feat/agent-environments-stage-a`,
head **`dea3f5ac`** — unchanged since the cycle-4 review. Ten commits on curator main `a2406dfe`, all
`%G? = G`, `Ivan Oparin <oparin@me.com>`. Authority: curator-spec main `f39f4a9`.

## Verdict: CHANGES REQUESTED → `to-dev`

Two blocking findings recorded at this head are **live and independently reproduced by this run** with
its own build of `dea3f5ac`. Findings body: `TASK-260905-30zs8t_review-findings-stage-a-4.md` (F13, F14
blocking; F15 major). This document is the verdict record plus my independent confirmation; it does not
restate the findings.

## Why this run exists, and a process finding for the orchestrator

`TASK-260905-30zs8t_review-findings-stage-a-4.md` was written by `RUN-260906-7e9471`, which was
**cancelled** (exit 143) seconds after attaching it. The three reviewer runs before it were cancelled
with the same runner diagnostic:

> `reviewer run RUN-260906-0d4f9b remains unsatisfied: reviewer run has no verdict branch while
> TASK-260905-30zs8t is development`

**Root cause.** Every review brief in this leaf instructs "Blocking/major → `development`". `development`
is **not** a reviewer verdict branch. The reviewer archetype recognises exactly `done` / `to-dev` /
`analysis` / `blocked` (plus `accept_cr`). Routing to `development` therefore leaves the run with no
verdict, the runner cancels it and respawns — four times so far, each losing a ~20-minute review. This
run routes to **`to-dev`**, which is the changes-requested branch, and the element lands in the same
producer queue.

**Ask:** correct the phrase in the next review brief (`Blocking/major → to-dev`), otherwise the loop
repeats on cycle 5.

## Change Request `repository_delta=empty` — judged

Correct by construction, and it is **not** the reason for this verdict. Every brief for this leaf states
that the curator-spec story workspace carries an empty delta by design and forbids writing into the
control root; the reviewable work is ten signed commits in the curator repository. An empty
curator-spec delta is the expected shape for this leaf and would not have blocked acceptance on its own.

## Independent reproduction at `dea3f5ac` (this run, own build)

Binary built by me from the worktree head into `.temp/review-5/curator`. Scripts under
`.temp/review-5/` (`f13.sh`, `f14.sh`), hermetic: `HOME`, `CURATOR_CONFIG`, `CURATOR_SYSTEM_CONFIG`,
`CLAUDE_CONFIG_DIR`, `CODEX_HOME`, `XDG_CONFIG_HOME`, `PI_CODING_AGENT_DIR`, `GIT_CONFIG_GLOBAL` all
under `.temp/`; `GIT_CONFIG_NOSYSTEM=1`; no network. `~/.curator` untouched.

### F14 — confirmed BLOCKING (machine source allowlist bypassed by a planted directory)

`internal/envprofile/envprofile.go:812`

```go
// isPathOperand tells a path operand from a git URL syntactically, never by
// probing the network: an existing local directory is a path.
func isPathOperand(operand string) bool {
	info, err := os.Stat(operand)
	return err == nil && info.IsDir()
}
```

environments §9.1 line 1513, verbatim at `f39f4a9`:

> `<source>` is either a git URL or a `path` operand. The distinction is
> syntactic, **never probed from the filesystem**: an operand beginning with `/`,
> `./`, or `../` (or a platform absolute-path spelling) is a `path` declaration;
> every other operand resolves as `git` under section 1.

The doc comment states the spec rule ("syntactically") and the body does the opposite. My run, one
operand, one machine-locked `allowed_sources: ["github.com/relux-works"]`, two working directories:

```
control (clean cwd):
  profile install github.com/evil-org/pkg --as c1
  -> profile_source_invalid: source github.com/evil-org/pkg is outside the machine's allowed sources

attack (cwd contains a planted directory "github.com/evil-org/pkg"):
  profile install github.com/evil-org/pkg --as a1
  -> rc=0  installed and activated profile a1 (root shadow 9.9.9, lock sha256:1a7bb9b0a288…)
  -> source.json  {"kind":"path","path":"github.com/evil-org/pkg","requirement":{}}
  -> profile use a1 ; CLAUDE.md contains "PLANTED LOCAL CONTENT — never fetched from the network"
```

Identical lock hash `1a7bb9b0a288…` to the cycle-4 reviewer's run — the same defect, reproduced
independently. §9.1 gives a `path` package no network identity and states the core §6.1 allowlist does
not apply to it, so the misclassification silently removes exactly the gate F9 installed. Second half
also reproduced: `/tmp/definitely-not-here --range '^1.0.0'` and `./missing-relative --tag v1.0.0` both
reach `git clone` and report `profile_source_invalid`, where §9.1 requires `profile_install_ref_conflict`
and no clone.

### F13 — confirmed BLOCKING (`--use` and first install report an activation that did not happen)

`internal/envprofile/envprofile.go:493-499` publishes the `current` pointer and nothing else — no
adapter entry attempted, no marker, no surface, no warning. My run, four adapters with real homes:

```
install alpha --use ; profile use alpha      -> CLAUDE.md = alpha, marker.profile.name = alpha
profile install https://example.com/org/beta --use
  rc=0  "installed and activated profile beta (root beta 1.0.0, lock sha256:2c915479…)"
  profiles/current    = beta
  profile list        = beta … current
  CLAUDE.md           = "alpha module body"     <- unchanged
  marker profile.name = alpha                   <- unchanged
  stderr              = (empty)                 <- no warning
profile sync                                    -> CLAUDE.md = beta, marker = beta

fresh machine, first install (no --use):
  rc=0  "installed and activated profile alpha"; profiles/current = alpha
  .agent-environment.json : No such file or directory
  CLAUDE.md               : No such file or directory
```

§9.1: activation on first install "is reported, never silent". §9.2: "The new current profile is
recorded only when the whole scope materialized." Here the report is emitted, the pointer moves, and
every agent launched from those homes reads the previous profile. This is a capability claim that does
not reproduce, in the standard negative-shape sense.

I confirm the cycle-4 coverage claim by my own grep at this head: **zero** `--use` coverage in the
repository. `grep -rn '\-\-use' --include='*_test.go' .` returns four hits, all unrelated
(`--use-yarnrc`, `--username`, `--userconfig`); no `InstallOptions{… Use: true}` exists in any test.
`TestProfileInstallListUse` (`cmd/curator/profile_test.go:56`) asserts the "installed and activated"
string and then runs a separate `profile use` before asserting the marker, so it passes either way.

### F15 — confirmed MAJOR as an evidence gap (not re-mutated by this run)

I confirmed the gap statically at this head: `grep` finds exactly one CLI-level policy-threading test,
`TestProfileListMigrationHonoursSystemPolicy` (`cmd/curator/profile_test.go:309`). `profile use`
(`cmd/curator/profile.go:146`) and `profile sync` (`:241`) carry `PolicyFromConfig(cfg)` with nothing
pinning it. I did **not** re-run the cycle-4 M-D mutant (`./cmd/curator` is a ~265 s run and this run is
time-bounded); I cite `RUN-260906-7e9471`'s first-hand result that M-D survives the whole `cmd/curator`
suite. Reported as cited evidence, not as my own measurement.

## Gates rerun by this run at `dea3f5ac`

| Gate | Result |
|---|---|
| `go build ./...` | exit 0 |
| `go vet ./internal/envprofile/... ./cmd/curator/...` | exit 0 |
| `gofmt -l internal cmd` | clean |
| `go test -count=1 ./internal/envprofile/` | `ok … 19.388s` |
| `git log --format='%G? %GS' a2406dfe..HEAD` | 10 commits, all `G`, `Ivan Oparin <oparin@me.com>` |

Not rerun by this run, cited from `RUN-260906-7e9471` at the same head: `golangci-lint`, `-race` on the
eleven touched packages, the `CURATOR_CONFORMANCE_ROOT` vector families (5 PASS / 0 FAIL, 7 stage-(b)
sub-skips), `gate-selftest.sh` (81 passed), `ledger-consistency.sh`, `./cmd/curator`.

**The gates are green while both blocking defects are live.** That is the point, not a mitigation: a
green suite around an unexercised production surface is exactly the shape this leaf's evidence rules
reject.

## AC coverage as measured

Producer-brief items against `f39f4a9`: **7 of 8** fully delivered. Item 7 (`linked` switching and the
profile CLI) is not: the `profile install … [--use]` row records the machine current outside the §9.2
transactional shape (F13), and its `<git-url|path>` operand rule is implemented against the filesystem
rather than §9.1's syntactic rule (F14). Rework-3 decisions F10/F11/F12 are **3 of 3** delivered — the
cycle-4 reviewer verified each by reproduction and I did not find them regressed.

CLI rows driven by a named committed test: `--use` is **0 of 1**.

## Required for the next cycle

1. **F14** — classify the operand syntactically (`/`, `./`, `../`, platform absolute spellings are
   `path`; everything else is `git`); never `os.Stat` to decide. Tests through `Install`/the CLI:
   (a) an operand naming an existing directory whose spelling is a git identity resolves as `git` and is
   refused by a locked allowlist; (b) `/tmp/<absent>` and `./<absent>` with `--range` are
   `profile_install_ref_conflict` with no clone attempted; (c) a **narrowing** mutant that keeps the
   classifier but re-admits the stat fallback for non-syntactic-path operands must fail a named test.
2. **F13** — either make install activation perform the §9.2 switch (attempt every entry, per-adapter
   results, record the current only when the whole scope materialized, `profile_use_partial` otherwise),
   or stop moving the current pointer at install and print the existing "activate with
   `curator profile use <name>`" guidance. Cover `--use` on an already-current machine and first install
   on a fresh machine through the CLI, plus a **narrowing** mutant that keeps activation but drops
   materialization for one adapter.
3. **F15** — table `TestProfileListMigrationHonoursSystemPolicy` over `list`, `use`, `sync`, `update`,
   then re-run M-D and record the killed test.

## Reviewer housekeeping

Read-only on tracked files. `git status --short` in the curator worktree is clean and `git diff` against
`dea3f5ac` is empty; all scratch is under the worktree's `.temp/review-5/`. Nothing was written into the
control root, and no probe reached the network.
