# TASK-260905-30zs8t — review findings, stage (a) core, cycle 4

Subject: curator worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-a-core`, branch
`feat/agent-environments-stage-a`, head **`dea3f5ac`** — one signed commit ("Stage (a) rework 3:
migration policy from CLI (F10), narrowing gate tests (F11), file operand rejection (F12)") on the
cycle-3 head `314ae748`; 10 files, +752/−147. Ten commits on curator main `a2406dfe`, all `G`
`Ivan Oparin <oparin@me.com>`. Authority: curator-spec main `f39f4a9`.

**Note on the head under review.** The spawn prompt for this run carries the cycle-3 brief
(`314ae748`) and `CR-TASK-260905-30zs8t-3`; the producer has since delivered rework 3 and published
revision 4 (`…_change-request_rev4.patch`, 06:33). This verdict judges `dea3f5ac`, the current branch
head and the tree revision 4 describes.

**Change Request note.** `repository_delta=empty` against the curator-spec story workspace is correct
by construction: every brief for this leaf states the story workspace carries an empty delta by design
and forbids writing into the control root. The reviewable work is the curator branch, and that is what
this verdict judges.

Verdict: **CHANGES REQUESTED** — F10, F11 and F12 are genuinely fixed and hold under attack. Two new
**blocking** findings on surfaces the brief names but no test drives, plus one major evidence gap that
repeats the F11 class on the F10 fix.

Reviewer scratch: the worktree's `.temp/review-4/` (attack scripts `a1.sh`–`a9.sh`, mutant tree,
produced locks/markers, range differential). Read-only on tracked files; the mutant tree is an `rsync`
copy under `.temp/`, restored byte-identical afterwards. `git status --short` in the worktree is clean.

---

## Rework verification — F10, F11, F12, reproduced against this head

Reproduced through the **built CLI binary** (`.temp/review-4/curator`, built from `dea3f5ac`) with real
machine configuration files, not through the library or the producer's tests.

### F10 — FIXED (verified, six scenarios)

`loadPolicyForHome` is gone; `PolicyFromConfig` + `loadMachinePolicy` (`envprofile.go:349`, `:1011`)
route every policy through `config.Load`, and `cmd/curator/profile.go` threads it into install, list,
use, update and sync. The cycle-3 reproduction now refuses on **every** command that reaches
`migrateGlobalSkills` (`.temp/review-4/a1.sh`, system config locking `allowed_sources` + `audit`):

```
profile install https://example.com/skills/hello  rc=1  outside the machine's allowed sources
profile list                                      rc=1  (same)
profile sync                                      rc=1  (same)
profile use default                               rc=1  (same)
default lock: absent    profile-repos/: empty (the refusal precedes the clone)
```

Matrix (`a2.sh`), each on its own isolated home:

| # | Policy source | Result |
|---|---|---|
| s1 | system locks `audit` with a revocation of the skill identity | refused: `migrated global skill "hello" audit blocked: source … is revoked` |
| s2 | same, with **`audit.enabled: false`** | refused identically — §9.1's "an advisory profile install does not exist" holds on the migration path |
| s3 | permissive system (positive control) | rc=0; lock members `[(context,default,""), (skill,hello,"example.com/skills/hello")]` — canonical identity, confirming the F8 migration boundary |
| s4 | no system config; gates in `CURATOR_CONFIG=<home>/curator.json` | refused — the path override applies |
| s5 | user config tries to override a locked `allowed_sources` | warning emitted, locked value wins, refused |
| s6 | malformed user config | rc=1 config error — a failed read is not an empty policy |

`Remove` no longer reaches `ensureDefault`, and `grep` finds no production caller of the bare
`List`/`Use`/`Sync`/`Update`/`EnsureDefault` entry points — the CLI uses the `WithPolicy` variants
throughout. `run2/s3_permissive/profiles/default/lock.json` is **VALID** under `context-lock-v1`
(ajv 2020 against `schemas/v1` at `f39f4a9`).

*Non-blocking note.* `loadMachinePolicy` returns an empty policy when `config.UserPath()` does not
exist. That branch is not CLI-reachable — `fileConfigSource.Load` → `config.Load` → `readObject` fails
the whole command with "global config not found" first — so no production path can reach a system
overlay through it. It is a latent shape for a future library caller only; the doc states the bound.

### F11 — FIXED (verified by re-running cycle 3's three surviving mutants)

Each mutant applied to an `rsync` copy at `.temp/review-4/mutant/`, `go test -count=1 ./internal/envprofile/`:

| Cycle-3 mutant | Cycle 3 | This head |
|---|---|---|
| M-A `gateSource` matches host only, ignoring the path | SURVIVED | **KILLED** — `TestSourceAllowlistRejectsAdjacentPathSegment` FAILS (`allowlist err = <nil>`) |
| M-B revocation applied only when `Kind == KindContext` (the single `RevocationFor` line) | SURVIVED | still survives — **equivalent mutant**, see below |
| M-C canary guard removed from `strictAuditMember` | SURVIVED | **KILLED** — `TestStrictAuditCanaryFailureBlocksInstall` FAILS (`canary err = <nil>`) |

M-B survives because revocation is enforced twice in `strictAuditMember` — the explicit
`audit.RevocationFor` at `envprofile.go:754` and `audit.Gate` at `:773`, whose `cfg.Audit.Revocations`
is populated from the same policy with `Enabled: true` hardcoded. Narrowing **both** layers by kind
(M-B2, `.temp/review-4/mutant`) kills both named tests:

```
--- FAIL: TestRevokedSkillMemberIsRefused   revocation err = <nil>, want profile_source_invalid
--- FAIL: TestRevokedMCPMemberIsRefused     revocation err = <nil>, want profile_source_invalid
```

So the class "revocation applies to every member kind" **is** covered; the producer's table row is
imprecise about which line it mutated, not wrong about the property. No finding.

The `canaryPasses` package-level seam (`envprofile.go:735`) is documented at its declaration and no
production path reassigns it. Accepted.

### F12 — FIXED (verified, operand and requirement, every spelling)

`canonicalGit` rejects the operand and `gateSource` backstops it at the clone boundary
(`.temp/review-4/a3.sh`, `a4.sh`):

```
file://<repo>   file://<repo>/   FILE://<repo>   File://<repo>   file:<repo>
  -> rc=1  profile_source_invalid: file:// git sources carry no network identity and are not accepted
transitive requires.contexts.dep.git = file://<dep>
  -> rc=1  same diagnostic; the dependency is never cloned (profile-repos holds only the root)
control: transitive https requirement -> rc=0, lock members carry canonical identities and commits
```

The stage-(a) `path`-kind lock (`run3/h/profiles/t1/lock.json`) is **VALID** under `context-lock-v1`.
The withdrawn "test-only shim" bound is genuinely withdrawn.

---

## New findings

## F14 — BLOCKING — the `path`-vs-`git` operand distinction probes the filesystem, so a directory in the operator's working directory shadows a git identity and installs local bytes with the machine source allowlist bypassed

`repeat-of: TASK-260905-30zs8t_review-findings-stage-a-3.md#F10` (same class: a §9.1 gate that is
present and correct but is not in force on a production path)

**Where.** `internal/envprofile/envprofile.go:812`

```go
func isPathOperand(operand string) bool {
	info, err := os.Stat(operand)
	return err == nil && info.IsDir()
}
```

**What is wrong.** environments §9.1 line 1512:

> `<source>` is either a git URL or a `path` operand. The distinction is
> **syntactic, never probed from the filesystem**: an operand beginning with `/`,
> `./`, or `../` (or a platform absolute-path spelling) is a `path`
> declaration; every other operand resolves as `git` under section 1.

The implementation does exactly what that sentence forbids. The consequence is not cosmetic: §9.1 also
says a `path` package "has no network identity: its identity for local revocation is its state hash,
**the core §6.1 network allowlist does not apply (local sources bypass it)**". So misclassifying a git
operand as `path` silently removes the allowlist gate F9 was written to install.

**Evidence — one operand, one locked machine allowlist, two working directories
(`.temp/review-4/a9.sh`):**

```
CURATOR_SYSTEM_CONFIG = {"schema_version":1,"locked":["allowed_sources"],
                         "allowed_sources":["github.com/relux-works"]}

control (clean cwd):
  curator profile install github.com/evil-org/pkg --as c1
  -> rc=1  profile_source_invalid: source github.com/evil-org/pkg is outside the machine's allowed sources

attack  (cwd contains a planted directory "github.com/evil-org/pkg"):
  curator profile install github.com/evil-org/pkg --as a1
  -> rc=0  installed and activated profile a1 (root shadow 9.9.9, lock sha256:1a7bb9b0a288…)
  -> source.json: {"kind":"path","path":"github.com/evil-org/pkg","requirement":{}}
  -> curator profile use a1 ; grep PLANTED ~/agents/claude/CLAUDE.md
     PLANTED LOCAL CONTENT — never fetched from the network
```

The operator typed a network identity, the machine forbids that identity, and curator installed
attacker-controlled local bytes into every agent home instead — because a directory of that name
existed relative to the process working directory. A checked-out repository, an unpacked archive, or a
`GOPATH`-shaped tree is enough to plant it; no privilege is needed beyond writing a directory into a
path the operator will `cd` into.

**Second consequence — the wrong diagnostic for a syntactic path operand that does not exist.** §9.1:
"a requirement flag or `--directory` with a `path` operand is `profile_install_ref_conflict`". Because
a non-existent absolute path fails the `os.Stat`, it is routed to `git` and handed to `git clone`:

```
curator profile install /tmp/definitely-not-here --range '^1.0.0'
  -> profile_source_invalid: clone /tmp/definitely-not-here: … fatal: repository … does not exist
curator profile install ./missing-relative --tag v1.0.0
  -> profile_source_invalid: clone ./missing-relative: … fatal: repository … does not exist
```

Both must be `profile_install_ref_conflict`, and neither operand should ever reach a clone.

**Why no test caught it.** `TestProfileInstallRefConflictIsAFailure`
(`cmd/curator/profile_test.go:123`) passes a `t.TempDir()` — an operand that **exists** — so it holds
under both the stat rule and the syntactic rule and cannot distinguish them. It is a positive-path
test for a refusal.

**Fix.** Classify syntactically: `/`, `./`, `../`, and the platform absolute spellings (a Windows
drive-letter or UNC prefix) are `path`; everything else is `git`. Never stat the operand to decide.

**Regression cover required.** Driving `Install`/the CLI: (1) an operand naming an existing directory
whose spelling is a git identity resolves as `git` and is refused by a locked allowlist — the a9.sh
shape; (2) `/tmp/<absent>` and `./<absent>` with `--range` are `profile_install_ref_conflict`, with no
clone attempted; (3) a NARROWING mutant that keeps the classifier but re-admits the stat fallback for
operands that are not syntactic paths — a named test must fail.

---

## F13 — BLOCKING — `profile install --use` (and first-install activation) records the machine current profile without materializing anything, reports "activated", and leaves every agent home on the previous profile with no warning

**Where.** `internal/envprofile/envprofile.go:493-499` (`installLocked`), `cmd/curator/profile.go:75`

```go
	activated := false
	if machine == "" || options.Use {
		if err := op.publish(map[string][]byte{CurrentFile(home): []byte(name + "\n")}); err != nil {
			return Info{}, false, false, err
		}
		activated = true
	}
```

Activation writes the `current` pointer and nothing else. No adapter entry is attempted, no marker is
written, no surface is materialized, and no warning is printed.

**What is wrong.** §9.2: "The new current profile is recorded only when the whole scope materialized."
`cli/curator.md:30`: "`--use` takes no name and **activates the installed root**." Here the pointer
moves while every in-place surface keeps the previous profile's bytes and the previous profile's
marker — the recorded current and every `agent-environment-marker-v1` disagree, and nothing surfaces
the divergence.

**Evidence (`.temp/review-4/a6.sh`), four adapters with real homes:**

```
1. install alpha --use ; profile use alpha        -> CLAUDE.md = alpha, marker.profile.name = alpha
2. curator profile install https://example.com/org/beta --use
   rc=0  "installed and activated profile beta (root beta 1.0.0, lock sha256:8a97f128…)"
   <home>/profiles/current            = beta
   curator profile list               = beta … current
   ~/agents/claude/CLAUDE.md          = "alpha module body"      <-- unchanged
   marker profile.name                = alpha                    <-- unchanged
   stderr                             = (empty)                  <-- no warning
3. curator profile sync                            -> CLAUDE.md = beta, marker = beta
```

The operator is told beta is active, `profile list` says beta is current, and every agent launched
from those homes reads **alpha**. `profile sync` repairs it — but only for an operator who already
knows something is wrong.

The first-install branch has the same shape on a fresh machine, where §9.1 promises "first install,
and the activation is reported, never silent":

```
curator profile install https://example.com/org/alpha       (fresh home, no current)
  rc=0  "installed and activated profile alpha"
  <home>/profiles/current      = alpha
  ~/agents2/claude/.agent-environment.json : No such file or directory
  ~/agents2/claude/CLAUDE.md               : No such file or directory
```

The activation is reported; it did not happen.

**Why no test caught it.** `--use` has **zero** coverage anywhere in the repository: no
`InstallOptions{… Use: true}` in any `internal/envprofile` test, and no `"--use"` argument in any
`cmd/curator` test. `TestProfileInstallListUse` (`cmd/curator/profile_test.go:56`) asserts the string
"installed and activated profile acme" and then runs a **separate** `profile use acme` before
asserting the marker — so it passes whether or not activation materializes anything.

**Fix (author's call, both are defensible; the current shape is neither).** Either make install
activation perform the §9.2 switch — attempt every entry, per-adapter results, record the current only
when the whole scope materialized, and report `profile_use_partial` when it did not — or stop writing
the current pointer at install and print the same "activate with 'curator profile use <name>'"
guidance the non-activating branch already prints. A silent pointer move plus an "activated" claim is
the one option the spec excludes.

**Regression cover required.** Through the CLI: `profile install <src> --use` on a machine already
current on another profile must leave the marker and the materialized bytes in agreement with the
recorded current (whichever of the two fixes is chosen), and first install on a fresh machine likewise.
Plus a NARROWING mutant: keep the activation but drop the materialization for one adapter — a named
test must fail.

---

## F15 — MAJOR — the CLI policy threading that the F10 fix depends on is pinned for `profile list` only; dropping it from `profile use` and `profile sync` leaves the whole `cmd/curator` suite green

`repeat-of: TASK-260905-30zs8t_review-findings-stage-a-3.md#F11` (same class: the gate is correct and
the production call site that carries it is unpinned)

**Where.** `cmd/curator/profile.go:146` (`UseWithPolicy(… PolicyFromConfig(cfg))`), `:241`
(`SyncWithPolicy(… PolicyFromConfig(cfg))`); rework report 3 names only
`TestProfileListMigrationHonoursSystemPolicy` for the CLI.

**Evidence — my mutant M-D on the `rsync` copy:** replace both call sites with `envprofile.Policy{}`,
leaving install/list/update untouched.

```
go test -count=1 ./cmd/curator/     ->  ok  github.com/relux-works/curator/cmd/curator  264.538s
```

**SURVIVES.** That mutant reinstates exactly the F10 bypass on `curator profile use` and
`curator profile sync`: the migration would run under an empty policy, so a system-locked
`allowed_sources` and system-locked `audit.revocations` would not reach `migrateGlobalSkills` on those
two commands. I verified behaviourally (`a1.sh`) that the shipped code refuses on all four commands —
this is an evidence gap, not a behavioural defect, and it is the same shape cycle 3 raised as F11 about
the F9 gates.

**Fix.** Extend `TestProfileListMigrationHonoursSystemPolicy` into a table over `list`, `use`, `sync`
and `update` (each on a fresh isolated home with a declared global skill and a system config that
forbids it), then re-run M-D and put the killed test in the table.

---

## Stated bounds — judged

| Bound (rework report 3) | Judgement |
|---|---|
| MCP package allowlist has no schema-1 machine-config surface; CLI passes empty (permits all) | **Defensible**, unchanged from cycle 3: `mcp_package_allowlist` is a manager-config **v2** key, out of scope for stage (a), and §2's default for an empty allowlist is "permits every network identity". Enforcement path exists and fires. |
| `fetchRaw` first-spelling-wins for transport selection | **Defensible**; the reviewer's SSH-only-operator note is now in the `gitManager` doc as asked. |
| `file://` test-only shim | **Withdrawn and replaced by rejection** — verified above. Correct disposition. |
| F7 scoped waivers, F2 migration warning | Unchanged, correctly carried to the stage that lands the operator surfaces. |
| `loadMachinePolicy` bare entry points require `home == cfg.Home()` | **Defensible** — no production caller uses the bare entry points; the bound is stated on each. |

---

## Verified correct — attacked, held

| Surface | Attack | Result |
|---|---|---|
| F10 migration policy | 4 commands × system-locked allowlist; 6-scenario policy matrix incl. `audit.enabled=false`, `CURATOR_CONFIG` override, locked-key override, malformed config | all refuse; positive control migrates with the canonical identity |
| F11 gate narrowing | cycle 3's M-A / M-B / M-C re-run on a copy; M-B2 added | M-A and M-C now KILLED; M-B shown equivalent, class covered by M-B2 |
| F12 `file://` | 5 operand spellings + transitive requirement + control | all refused at the boundary, dependency never cloned |
| F1 detector scope (re-run) | secret in a `--directory sub` module; transitive `requires.contexts{directory}` member; control with the secret outside the addressed root | both refused `member … carries a blocking context-secret-material finding`; control installs |
| F2 fresh home (re-run) | isolated home: `profile list`, `use default`, `sync`, `use --clear --env claude_code` | all rc=0 |
| F4 skip honesty (re-run) | `grep CURATOR_STAGE_B` over the tree | 0 hits in the repository (only the cycle-1 findings resource on the board); `stage-deferred` registered at `skip-classes.tsv:106` and `platform-cases.tsv:187` |
| F5/F6 ranges (re-run, **my own 30 fresh cases** × 24 versions vs `semver@7.7.4`) | caret on `0.x`/`0.0.x`/prerelease, tilde with fewer components, x-ranges, `^*`, `<=x`, `>2`, `-0` upper bounds, `\|\|`, `latest`, spaced operator, `v`-in-range, hyphen, build metadata | **24 of 30 exact agreement**; all 6 deviations are §1.4 restrictions (`latest`=`*`; `^v1.2.3`, `~ 1.2.3`, `1.2.3 - 1.9.9`, `1.2.3+meta`, `^1.2.3+meta` rejected). 0 unexplained |
| Schema validity | ajv 2020 against `schemas/v1` at `f39f4a9` | git-root default lock, path-kind lock both VALID under `context-lock-v1` |
| Materialization / lock / marker / versions | `git diff 314ae748..dea3f5ac` is **empty** for `internal/{pkgversion,contextmaterialize,contextlock,envmarker,contextpkg,contextaudit,contextstore,interop}` | cycle 2 and cycle 3's byte-for-byte header, weight-ordering and marker verdicts carry over unchanged |
| Marker shape | `profile use` on 4 adapters, marker read back | `profile.name`, `mode: linked`, per-adapter homes written |
| CLI rows | 19-row refusal/acceptance sweep through the built binary | `profile_install_ref_conflict` ×2, `--use` takes no name (rc=2), `profile_not_found` ×3, `profile_in_use`, `environment_unknown`, `list takes no arguments`, re-install-as-update, `remove --purge`, compose refusal — all as `cli/curator.md` names them, **except** the two F14 rows |
| Scope / signatures | `git log --format='%G? %GS' a2406dfe..HEAD` | 10 commits, all `G`, `Ivan Oparin <oparin@me.com>`; no unrelated behaviour change |

## Gates rerun by the reviewer at `dea3f5ac`

| Gate | Result |
|---|---|
| `go build ./...` | exit 0 |
| `go vet ./...` | exit 0 |
| `gofmt -l cmd internal` | clean |
| `golangci-lint run` (envprofile, contextresolve, cmd/curator) | **0 issues** |
| `go test -count=1 -race` on 11 new/touched packages | all `ok` (envprofile 25.2s) |
| interop vectors, `CURATOR_CONFORMANCE_ROOT=<spec>/conformance/v1` | 5 top-level PASS (detectors, header, monolithic, resolution, versions), 0 FAIL |
| conformance sub-skips | exactly **7**: `referenced-{claude-code-composed,opencode,opencode-zero-modules}` + `mcp-{claude-code,codex-cli,opencode,pi-none}` — all genuinely stage (b) |
| `bash .github/ci/gate-selftest.sh` | 81 passed, 0 failed (exit 0) |
| `bash .github/ci/no-broad-suppression.sh` | ok (exit 0) |
| `bash .github/ci/ledger-consistency.sh .temp/rework3-evidence/ledger` | ok, 103 rows across linux/darwin/windows |
| `go test ./cmd/curator/` | **run by me on the mutant copy** (M-D): `ok … 264.538s`. Not rerun on the unmutated tree; the producer's first-hand run is cited. |

## AC coverage as measured

Rework-3 author decisions: **3 of 3** findings (F10, F11, F12) delivered and independently verified.

Producer-brief items against `f39f4a9`: items 1, 2, 3, 4, 5, 6, 8 delivered and verified. **Item 7 is
not**: the `linked` switching shape holds for `profile use`/`profile sync`, but the `profile install
… [--use]` row of the same item records the current outside the transactional shape (F13) and its
`<git-url|path>` operand rule is implemented against the filesystem rather than the spec's syntactic
rule (F14). **7 of 8 brief items fully delivered.**

CLI rows driven by a named committed test: `--use` is **0 of 1** — the flag has no test anywhere. The
19-row CLI sweep above is my evidence, not the suite's.

## Reviewer housekeeping

Every probe ran with `HOME`, `CURATOR_CONFIG`, `CURATOR_SYSTEM_CONFIG`, `CLAUDE_CONFIG_DIR`,
`CODEX_HOME`, `XDG_CONFIG_HOME`, `GIT_CONFIG_GLOBAL`, `GIT_CONFIG_NOSYSTEM` and `PI_CODING_AGENT_DIR`
pointed at directories under `.temp/review-4/`. `~/.curator` was never touched, and no probe reached
the network (`insteadOf` rewrites onto local repositories throughout). The mutant tree is a copy under
`.temp/review-4/mutant/`; every mutation was reverted from a saved original and the copy now differs
from the worktree in no tracked file. `git status --short` in the worktree is clean and `git diff`
against `dea3f5ac` is empty. Nothing was written into the control root.
