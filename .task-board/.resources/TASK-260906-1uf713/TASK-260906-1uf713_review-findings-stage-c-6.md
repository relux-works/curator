# TASK-260906-1uf713 — review findings, stage (c) cycle 6 (rework 5)

Verdict: **ACCEPT** — two minor findings, four observations, no blocking and no major.

`repeat-of:` **cycle-5 C5-M1** — finding C6-m1 is the same silent-no-op-reporting-success shape
that C5-M1 named, in the `git` addressing mode the fix did not take. It is **pre-existing on
`origin/main`** (driven there, not read), so it is not this leaf's defect; what belongs to this
leaf is one false clause in the bound that carries it.

`repeat-of:` **cycle-5 C5-m1** — finding C6-m2 is one more unpinned clause of the same
`installLocked` reinstall arm: production is correct, a narrowing mutant survives.

No other cycle-1..5 class recurs.

- Subject: `~/Developer/ReluxWorks/.worktrees/curator-stage-c`, branch
  `feat/agent-environments-stage-c` at `8b8aa0417f49c11246a0a53d2006ef75708d4ea5`.
  `gh pr view 61 --json headRefOid` = the same OID, so the hosted lanes run this exact tree.
  The producer worktree was read-only throughout: `git status --short` empty and HEAD unchanged
  before and after. All scratch under `/tmp/c6`.
- Rework 5 is `git diff 71e6baec..HEAD` — 5 files, +500/−9. The **true stage delta** is
  `git diff d23dcd16..HEAD` (30 files, +7473/−175), not `origin/main..HEAD`: `origin/main` is
  now `db444157` and carries one commit this branch does not, so the `origin/main` diff pulls in
  `internal/closureexec` and `internal/godriver` work that is **not** stage (c)'s. I reviewed
  against the merge base.
- Authority: curator-spec `550579d` (§1, §6, §8.4, §9.1, §9.2, §9.5, §9.6, §12.1, §12.2;
  `profiles/manager.md` §1; `cli/curator.md`), with `87a0d006` read only for the known
  consumption gap.
- Method: a `curator` binary built from `8b8aa041` via `git archive` into `/tmp/c6/head`
  (`go build ./cmd/curator`, exit 0), invoked as a standalone process against scratch homes with
  `CURATOR_CONFIG`, `CURATOR_SYSTEM_CONFIG`, `CLAUDE_CONFIG_DIR`, `CODEX_HOME`, `XDG_CONFIG_HOME`,
  `PI_CODING_AGENT_DIR`, `GIT_CONFIG_GLOBAL` and `HOME`/`USERPROFILE` pinned per case. A second
  binary from `origin/main` (`db444157`) for the trunk comparison. Every script and mutant is in
  `TASK-260906-1uf713_review-probes-stage-c-6.tgz`.
- **No `test-gate.sh` lane ran at any point during this review, and every Go suite ran strictly
  sequentially** — one process at a time, never a `-race` suite beside a plain one — so nothing here
  is Go-test-lock contention noise. The one `-race` run (four packages, all `ok`) was the last thing
  executed, alone. Mutant oracle: `go test -count=1 ./internal/envprofile/ ./internal/config/` on
  stock → exit 0 (`ok 69.974s`, `ok 0.919s`).

---

## The empty repository delta, addressed explicitly

`CR-TASK-260906-1uf713-6` revision 6 has `repository_delta: empty`. **That is the correct
outcome for this leaf, and I verified it rather than assuming it:**

- The story worktree is a checkout of **curator-spec**
  (`origin = github.com/relux-works/curator-spec.git`), while this leaf's entire scope is the
  **curator** repository, on `feat/agent-environments-stage-c` in a different worktree. Every
  brief from the producer brief onward says "do not push, do not touch PR #61, nothing into the
  control root". The leaf structurally cannot produce a delta here.
- Candidate tree `2e6ca472ae8542a95e446ed4afb11cb75820536d` equals
  `git rev-parse 87a0d0060bad…^{tree}` — byte-identical to the base — and the patch resource's
  sha256 is `e3b0c442…b855`, the sha256 of the empty string. Both checked.
- Cycle 5 warned that **revision 5** was *not* empty and that its delta was another element's
  work (the five curator-spec commits of the `path`-overlay reconciliation, TASK-260906-3x0w4y).
  Revision 6 resolves that correctly: the base advanced to `87a0d00`, so those commits are now in
  the **base**, not in the delta. Nothing foreign is being accepted here.
- The leaf's actual deliverable — 32 signed commits (`git log d23dcd16..HEAD`, every one `G`,
  human identity `Ivan Oparin <oparin@me.com>`) and eighteen outcome resources — is what I
  reviewed, directly in the producer worktree and in a `git archive` copy of it.

---

## Hosted lanes on the reviewed head

`gh pr checks 61` on `8b8aa041`, run 34056559695:

| Lane | Result |
|---|---|
| Lint | pass 38s |
| Naming gate | pass 10s |
| Interop conformance gate | pass 24s |
| Gate self-test (ubuntu / macos / windows) | pass 10s / 11s / 26s |
| Test (ubuntu-latest) | pass 3m15s |
| Test (macos-latest) | pass 9m52s |
| Test (windows-latest) | pass 32m34s |
| Race (ubuntu-latest) | pass 7m15s |
| Race (macos-latest) | pass 16m48s |
| Candidate suite | `skipping` (gated matrix; it does not run on a PR event) |

**All eleven hosted lanes are green on the exact head I accepted, Windows included.**

## The candidate dispatch against the task authority

Run 34058365116, `workflow_dispatch` on head `8b8aa041`, candidate revision
**`550579d12ea6d5daeaeabdae69933de87ba28e4b`** — the authority this stage's AC names.
Read from the `candidate-evidence-<os>` artefacts, not from `--log-failed`:

| Job | Result | `suite-plan` | fail events |
|---|---|---|---|
| Candidate suite (ubuntu-latest) | **success** | `served=71 deferred=0 excluded=1` | 0 |
| Candidate suite (macos-latest) | **success** | `served=72 deferred=0 excluded=0` | 0 |
| Candidate suite (windows-latest) | **success** (on re-run; see below) | `served=72 deferred=0 excluded=0` | 0 |

**All three candidate runners are green against the task authority.**
`evidence_class candidate-only`, `release_claim none`, `conformance_claim none`, and
`CI_REQUIRE_FULL_ROOT=1` — `deferred=0` on every runner, so **no package ran without its
vectors**. The single ubuntu exclusion is `internal/godriver`, excluded on linux by the root's
own `vectors/conformance-claim-v3-qualification.json`, and its refusal is still asserted by
`TestProbeRejectsAnUncoveredPlatformBeforeTheWorker`.

**The Windows candidate job, reported exactly.** Its **first attempt failed on infrastructure, not
on a test**: GitHub's own annotation on job 101554326845 is *"The hosted runner lost communication
with the server"*; the job died 52 minutes into step 10 with step 10 left `in_progress`, step 11
(`Upload candidate evidence`) never reached, and no `candidate-evidence-windows-latest` artefact
produced. That is neither evidence of a Windows pass nor of a Windows failure, so rather than infer
either way I re-ran that single job (`gh run rerun 34058365116 --job 101554326845`) and read its
evidence artefact (9998022110):

```
candidate_revision  550579d12ea6d5daeaeabdae69933de87ba28e4b
runner_goos         windows        evidence_class  candidate-only    release_claim  none
suite-plan: served=72 deferred=0 excluded=0
suite-plan: CI_REQUIRE_FULL_ROOT=1 -- every package must be served by this root
fail events: 0    passing cases: 5465
```

**Success.** And it closes the blind spot exactly: all six of the families the green PR lane defers
ran and passed on Windows against the authority —
`TestManagerConfigV2SchemaCases` 42, `TestManagerConfigV2Vectors` 14, `TestSystemConfigV2SchemaCases`
25, `TestFragmentAuthoritativeSchemaCases` 50, `TestFragmentEmissionMatchesReference` 1,
`TestParseAuthoritativeEnvMarkerSchemaCases` 54.

**Exactly what the green Windows PR lane does and does not cover**, read from
`test-evidence-windows-latest` rather than assumed. On the `SPEC_PIN` root, `suite-plan` reports
`served=69 deferred=3` and the three deferred packages **still run** — "it runs with
`CURATOR_CONFORMANCE_ROOT` unset, taking the path its own tests implement" — so `internal/config`
executes 66 top-level cases on Windows and skips exactly 3. The complete Windows blind spot of the
green lane is therefore **six** `root-unset` cases:

```
internal/config      TestManagerConfigV2SchemaCases / TestManagerConfigV2Vectors / TestSystemConfigV2SchemaCases
internal/envfragment TestFragmentAuthoritativeSchemaCases / TestFragmentEmissionMatchesReference
internal/envmarker   TestParseAuthoritativeEnvMarkerSchemaCases
```

Those six are data-driven vector readers, and rework 5 touched none of the three packages: their
git tree hashes are **byte-identical** between `71e6baec` and `8b8aa041`
(`a53826734672` / `7567c7643a7e` / `495ee29b748b`, unchanged). That blind spot is now closed
directly: the Windows candidate job ran all six against `550579d` and all six passed (above).

The environments families demonstrably **ran and passed**, from the ubuntu evidence:

```
42  internal/config    :: TestManagerConfigV2SchemaCases
14  internal/config    :: TestManagerConfigV2Vectors
25  internal/config    :: TestSystemConfigV2SchemaCases
50  internal/envfragment :: TestFragmentAuthoritativeSchemaCases
54  internal/envmarker :: TestParseAuthoritativeEnvMarkerSchemaCases
```

That is the AC's conformance-subset clause, met against the authority.

**The candidate lane against curator-spec main (`87a0d00`) remains red** on the 30
`path`-overlay-reconciliation subcases. Per the cycle-6 brief that is a decided, filed matter
(TASK-260906-19gjyw), not a finding, and I confirm only what I was asked to: the bound is worded
truthfully and ledger rows 303–305 describe what their tests assert (checked below).

---

## MINOR

### C6-m1 — the §9.5 retry is still a silent no-op for a **git** source, and the bound that carries it says it cannot be driven

`repeat-of: cycle-5 C5-M1`.

C5-M1 was fixed for `path` roots and I re-drove the fix (below). The identical row for a **git**
source still drops both flags:

```
$ curator profile install https://example.com/gk --as gk --tag v1.0.0 --use
    [EXIT 1]
    claude_code: environment_surface_unmanaged_conflict: CLAUDE.md exists and no marker records it
    …  installed profile gk (lock sha256:bd788b9f…)
    curator: profile_use_partial: the scope is partially switched; the recorded current is unchanged
    STATE current=<none>  CLAUDE.md=OPERATOR HAND-WRITTEN CONTEXT  backups=[]

# the notice invites the retry; cli/curator.md line 30 puts [--takeover] on this exact row
$ curator profile install https://example.com/gk --as gk --tag v1.0.0 --use --takeover
    [EXIT 0]
    updated profile gk (lock sha256:bd788b9f…)
    STATE current=<none>  CLAUDE.md=OPERATOR HAND-WRITTEN CONTEXT  backups=[]

$ curator profile use gk --takeover
    [EXIT 0]  switched, notice printed, backup generation 1   ← the only recovery
```

Root cause is one branch away from the one that was fixed:
`installLocked`'s `prior == source` arm (`internal/envprofile/envprofile.go:644-653`) routes a
path reinstall into `reinstallPathLocked`, which now honours the flags, and routes a git reinstall
into `updateLocked`, which never reads `options.Use` or `policy.Takeover` and whose result is
returned as `info, false, true, nil`.

The flag itself is not broken for git — a **first** install proves it:

```
$ curator profile install https://example.com/gk --as gk --tag v1.0.0 --use --takeover
    [EXIT 0]
    claude_code: switched  notice: taking over unmanaged CLAUDE.md …
    installed and activated profile gk
    STATE current=gk  backups=[1]
```

So the defect is exactly one branch: the same-source reinstall.

**Why this is a minor and not a repeat blocker: it is pre-existing on trunk, and I drove that too.**
Same probe against a binary built from `origin/main` (`db444157`, schema-1 config):

| binary | `install --use` (first, with an unmanaged CLAUDE.md) | the retry |
|---|---|---|
| `origin/main` `db444157` | exit 1, `environment_surface_unmanaged_conflict`, current `<none>` | exit 0, `updated profile gk`, nothing done |
| head `8b8aa041` | identical | identical |

So the §9.5 stop on `install --use` and the `--use` drop on a same-source git reinstall are both
**already on trunk**. Stage (c) adds one thing here: the `--takeover` flag itself, which joins the
dropped set. Rejecting stage (c) would not remove the defect from trunk, and the rework-5 brief
scoped the fix to `path` roots on purpose.

**What does belong to this leaf** is the bound in `TASK-260906-1uf713_rework-report-5.md`:

> Git same-source reinstall still delegates to `updateLocked` with no `--use`/`--takeover`
> activation handling — pre-existing `origin/main` shape (cycle 5 reported it as code reading),
> **undrivable hermetically here**, and this rework was scoped to path roots. Left unchanged, stated.

The substance is honest and the bound was neither dropped nor disguised — but "undrivable
hermetically here" is false. It drives in about thirty lines using **this repository's own fixture
pattern**: a local repo served under a fake canonical identity through a `GIT_CONFIG_GLOBAL`
`insteadOf` rewrite, which `internal/envprofile/network_fixture_test.go` and
`cmd/curator/profile_test.go:353` already use throughout the stage's own suite. The producer
inherited the claim from cycle 5's honest statement that *it* could not drive it (cycle 5 tried
`file://`, which `canonicalGit` correctly refuses) rather than inventing it — a lesser fault than
a fabricated bound, but under this chain's attestation standard an unverified inherited claim is
still a claim.

The bound also understates the consequence: it says the flags are "not handled", not that the row
**reports success at exit 0 for work it did not do** — the shape cycle 5 called the one blocking
thing when it was the path arm.

**Ask:** correct that clause (drivable, and here is the fixture), state the consequence in the
same words cycle 5 used, and file the behaviour as a trunk task. Do not re-open stage (c) for it.

### C6-m2 — the reinstall activation's first-install clause is unpinned

`repeat-of: cycle-5 C5-m1` (the untested addressing mode of the same command).

`reinstallActivation` (`internal/envprofile/envprofile.go:827-836`) has two clauses:

```go
if machine == "" {
        return true, nil          // clause 1: a machine with no current activates
}
return use && machine != name, nil // clause 2: --use on a non-current root
```

Clause 2 is pinned four ways. Clause 1 is not: my narrowing mutant `r2`, which replaces
`return true, nil` with `return use, nil` — exempting exactly the first-install case and nothing
else — **survives** `./internal/envprofile/` and `./cmd/curator/` together.

Production is correct and the behaviour is operator-reachable, which I drove rather than inferred.
With a `path` profile installed and the machine current cleared:

```
$ curator profile install /tmp/…/s3 --as tk            # no --use at all
    [EXIT 0]  claude_code/codex_cli/opencode/pi: switched … updated profile tk
    STATE current=tk
```

and under a locked `require_current_profile: acme` the same command is refused by the seam:

```
$ curator profile install /tmp/…/s2 --as tk
    [EXIT 1]  environments.require_current_profile: profile use of "tk" is refused: …requires current profile "acme"
    STATE current=<none>
```

This is exactly the §9.5 stop's *other* retry — the operator who adds `--takeover` but not `--use`
— and it works. It just has no test, and the doc comment claims it explicitly ("a machine with no
current activates the same way a first install does"). All five subtests of
`TestProfileInstallReinstallHonoursUseAndTakeover` set up a machine that already has a current.
One subtest with the current cleared closes it.

---

## Observations (not findings)

- **Two more pre-existing §8.4 collapse sites, beyond the one cycle 5 named.** Cycle 5 found
  `purgeHomes` (`switch.go:709`). Enumerating every `envmarker.Read` call site myself, two more
  have the same shape: `status.go:432` (`if marker, err := envmarker.Read(managed); err == nil &&
  marker != nil` — an unreadable marker silently reports the home unprovisioned) and
  `status.go:521` (`if err != nil || marker == nil { continue }` — an unreadable marker hides an
  orphan). Both are `git blame` → `e43dd2b8` ("Stage (b): env resolve and status rows"), an
  ancestor of the merge base, so **all three are outside the stage-(c) delta** and the producer's
  and cycle 5's sweep of the delta stands. Recorded so the follow-up task has the complete list
  rather than one of three.
- **The `CheckMachineUse` doc enumeration is now incomplete, though still true.** Its comment
  (`envprofile.go:470-474`) lists "profile use (clear or not), profile install --use,
  first-install auto-activation, import --use, and resync". Rework 5 adds a sixth path to the seam
  — the reinstall activation — and does not list it. The universal claim ("funnels through the
  single gate in useLocked") is still true and I verified it by driving the reinstall activation
  under a lock, so this is not the cycle-1 M1 class; it is a stale example list.
- **One stage-introduced skip reason is still unclassified, benignly.**
  `cmd/curator/envconfig_test.go:496`'s `t.Skip("no git on PATH")` is in a file this stage *adds*,
  so it is strictly a skip this stage introduces, and my classifier reports it UNCLASSIFIED. The
  producer disclosed it by file and line and cycle 5 ruled the text out of scope as pre-existing
  elsewhere. It cannot shrink a suite silently — git is on PATH on all three runners so the branch
  is dead, and if it ever fired the gate would fail closed, which is the gate working. Left as an
  observation.
- **`--directory` on a path operand** remains admitted by cycle 5's `my3` mutant. Still stage (a)
  code on trunk, still out of the delta, still one row from being closed.

---

## What I verified and found correct

**C5-M1 — the path retry, re-driven.** The cycle-5 fixture (`probe-recovery.sh`, unmodified),
after `environment_foreign_manager_detected`:

```
$ curator profile install …/src_a --as tk --use --takeover
    claude_code: switched   notice: taking over unmanaged CLAUDE.md (foreign-manager symlink) …
    updated profile tk
    symlink=no   current=tk   foreign bytes=FOREIGN SOURCE OF TRUTH (unchanged)   backups=[1]
```

All four combinations are driven through `run()` by
`TestProfileInstallReinstallHonoursUseAndTakeover` (5 subtests, `runProfile` → `run(args, …)`,
not a helper), plus the current-root case. The doc comment
(`envprofile.go:722-733`) now describes the real path, and I checked its every clause against the
code: `--use` activates a non-current root, no-current activates like a first install, `--takeover`
alone activates nothing, both ride `useLocked`, the reinstall always reports `updated`.

My own narrowing mutants on that fix:

| # | mutant | result |
|---|---|---|
| r1 | `reinstallActivation`: exempt exactly the `--use` clause | **KILLED** (`use without takeover…`, `use with takeover…`) |
| r2 | `reinstallActivation`: exempt exactly the first-install clause | **SURVIVES** → C6-m2 |
| r3 | `activateReinstall`: strip `Takeover` from the policy handed to `useLocked` | **KILLED** (`use with takeover takes over and activates`) |
| r4 | unchanged-lock branch loses the activation — C5-M1 restored exactly | **KILLED** (both `--use` subtests) |

r3 matters: it proves the takeover genuinely reaches the switch rather than the notice being
printed by some other path.

**The seam covers the new caller.** Driven under a locked `require_current_profile: acme`:
a reinstall with `--use` of an already-current root needs no switch and exits 0; a reinstall with
no flags on a machine with no current is **refused at exit 1** naming the knob and the required
profile, with the current unmoved. `profile import --use` under the same lock → exit 1, nothing
written into any native home. `profile use <name> --clear` → exit 2 with the usage line, so the
C3-B1 undocumented form is still refused at the CLI. `CurrentFile` still has exactly one
production writer (`switch.go:248`); the only other writer, exported `SetCurrent`, has four
callers and all four are `_test.go`.

**C5-m1 — both cycle-5 survivors are dead.** Against the same oracle:

| # | mutant | cycle 5 | now |
|---|---|---|---|
| my1 | reinstall arm exempts the no-`--as` form | SURVIVES | **KILLED** — `TestPathReinstallBareAndNonCurrentMovesPin` |
| my4 | reinstall arm exempts a non-current profile | SURVIVES | **KILLED** — same test |

**C5-m4 — the audit gates, and the reachability answer.** Cycle 5's `my2` (narrowing the reinstall
blocking-audit to admit exactly a context member) is now **KILLED** by
`TestReinstallBlocksSecretOverlayMember`, and I built the `updateLocked` twin myself
(`m_ug1.py`, same narrowing on `updateLocked`'s `report.Blocking()`) — **KILLED** by
`TestUpdateBlocksSecretOverlayMember`. Both gates are pinned. The producer's reachability answer —
*yes*, a blocking-but-not-strict finding is reachable, because a revoked-but-clean member passes
the deterministic detector and is caught by the strict member audit — is stated rather than
inferred, and the two revoked tests exercise exactly that. Cycle 5 reported this as *unknown* and
was right to; the answer is now established.

**Cycle-4's three refusal mutants, re-run rather than taken on trust.** The rework-5 report states
plainly that "Rework-4's P2/P3/P4 mutants were not re-run; their `updateLocked` arms are
byte-unchanged by this delta". That is an honest disclosure of an unverified row, so I verified it:
the `4a8a1d42..8b8aa041` diff touches only comment lines in that arm, and I re-applied the mutants
anyway. **P2** (drop `|| rootMember.StateHash == ""`) → KILLED by
`TestUpdatePathWithoutStatePinIsSourceInvalid`; **P4** (a failed snapshot read returned as
`unchanged` with a nil error — cycle 4's worst survivor and §8.4 verbatim) → KILLED by
`TestUpdatePathMissingSnapshotIsSourceInvalid`. The claim was true and the pins are live.

**C3-M1 — the transcribed pin, re-proved both ways.** `TestSystemV2LockableSubsetIsClassWide`
transcribes §12.2's six keys as a literal (`environments_test.go:413-420`) and branches on
`transcribed[…]`, never on `LockableEnvKeys` — I read every reference. I checked the literal
against `git show 550579d:protocol/environments.md` §12.2 ("`overlays_allowed`, `precedence`,
`mcp_package_allowlist`, `passable_env_names`, `require_current_profile`, and `isolation`") — it
matches exactly. Then:

| # | mutant | result |
|---|---|---|
| w1 | **widen** `LockableEnvKeys` by `environments.secret_material_waivers` (the §9.1 path cycle 5 drove end to end) | **KILLED** |
| w2 | **narrow** `LockableEnvKeys` by dropping `environments.mcp_package_allowlist` | **KILLED** |

Fourth cycle on this class, and the pin no longer derives from the thing it protects.

**C3-B1 — the seam, re-attacked.** Mutant `s1` narrows the gate to `scope == "" && !clearScope`,
restoring the C3-B1 hole for exactly the `--clear` form and nothing else → **KILLED** by
`TestMachineClearReachesSeamGate`.

**C5-m2 — the skip enumeration, re-run independently.** I extracted every `t.Skip*` the stage
delta *adds* (`git diff -U0 d23dcd16..8b8aa041 -- '*_test.go' | grep '^\+.*t\.Skip'` → 14 sites,
9 distinct reasons) and pushed each through my own port of `platform-case-gate.sh`'s `classify()`.
Result: every one classifies except `no git on PATH` (see Observations). The two bare
`chmod refused: %v` reasons cycle 5 found are gone — spelled like their siblings, not silenced.
And re-running the same reasons against the pre-fix registry (`git show 4a8a1d42:…/skip-classes.tsv`)
reproduces the red Windows lane exactly: the two `mode-000 **file**` reasons come back
UNCLASSIFIED while the three `directory` siblings still classify — so C4-B2's fix **widened the
registry** rather than bending a skip text to lie about its artefact.

**C5-m3 — the reissued bound.** Ledger rows 303–305 name **TASK-260906-19gjyw** (which exists) and
describe an implementation gap, not a spec contradiction. I read each against its test:
row 303 / `TestPathOverlayDeclarationParses` — the test does load a path overlay carrying a
revision and a weight, exactly as the row says; row 304 / `TestPathOverlayJoinsClosure` — the row
now says "from a manufactured Policy (resolution machinery only; no operator surface can declare
it until the TASK-260906-19gjyw consumption lands)", which is precisely the cycle-1 M1 complaint
answered honestly; row 305 / `TestPathOverlayFormIsSourceInvalid` — matches. `TASK-260906-vlrjo1`
is named in code at `internal/envprofile/managed.go:728` for the POSIX dotfile list, and
`TestTakeoverWarnsDotfileHeuristic` still runs unskipped on all three runners (Windows lane green).
Nothing untruthful lands.

**§9.5's read-only rule, driven.** Against a machine with a foreign-manager symlink in the
claude_code home, an unmanaged `AGENTS.md` in the codex home and a `~/.local/share/chezmoi` present
to trip the dotfile heuristic, I ran seven read-only rows — `env status`, `env status --check`,
`env status --json`, `profile list`, `env resolve claude_code`, `profile compose default list`,
`env config show` — and diffed the whole home before and after:

```
backup dirs anywhere: []          markers in any native home: []
claude CLAUDE.md still the foreign symlink: YES
```

No onboarding, no backup, no prompt, nothing written into a native home. The only curator write is
`ensureDefault` materializing the builtin default profile's store entry under the **manager** home
(§9.4), which is the manager's own tree. (One further directory appeared under `CODEX_HOME` — a
`tmp/arg0/` shim of symlinks to `/Users/iv/.local/bin/codex`. That is the real `codex` binary on my
host writing its own scratch when curator probes the tool version, not a curator write; it cannot
occur on a runner with no codex installed. Recorded so the diff is not mistaken for one.)

**The locked `overlays_allowed: false` gate — a survivor I then proved unreachable.** §12.2 says a
locked `overlays_allowed: false` "empties every overlay list". My first narrowing mutant **OV1**
exempted exactly the builtin `default` profile from `Policy.EffectiveOverlays`'s forbid check and
**survived** the `envprofile` and `config` suites. Rather than report that as a hole, I built it into
a binary and drove it against stock, with the machine declaring an overlay for `default` and a system
file locking `overlays_allowed: false`:

```
              profile use default   →  overlay content in CLAUDE.md: 0  (both binaries)
              profile update default →  exit 1  profile_update_blocked: "default" is the builtin local profile and does not move
              profile update --all   →  exit 0, lock members ['default']  (both binaries)
```

The exemption reaches no resolution path at all, because the builtin default is a local root that
never re-resolves — so OV1 is a defect in my mutant, not in the gate. Replacing it with **OV3**, which
lets exactly a *single-declaration* overlay list through the forbid, **KILLED** the mutant twice over
(`TestPolicyFromConfigCarriesEnvGates`, `TestForbiddenOverlaysResolveAlone`). The gate is pinned, and
`compose list` prints the manager §1 inert warning
(`warning: overlays_allowed is false: the listed declarations are inert; resolution joins the root
alone`) beside the declaration, which is the C3-m1 resolution.

One mutant on this surface does survive with no exploit path: **OV2**, admitting exactly a
`directory` on a path overlay declaration, survives — but path overlays are unreachable from all
three operator surfaces under the declared M1 bound, so it sits inside the dead branch that bound
already names. It rides with TASK-260906-19gjyw rather than standing alone.

**The two precedence primitives, against real materialized bytes.** Cycle 3 drove these; stage (c)
is what wires `Policy.Precedence()` into materialization and `env status`, so I re-drove all four
combinations end to end — a root at weight 20 requiring a git leaf at weight 10, each knob pair set
in `manager-config` schema 2 and read by the production loader, then `grep` over the emitted
`CLAUDE.md`:

| `winner` | `placement` | materialized order |
|---|---|---|
| `higher-weight` | `winner-last` | LEAF(10), **ROOT(20)** |
| `higher-weight` | `winner-first` | **ROOT(20)**, LEAF(10) |
| `lower-weight` | `winner-last` | ROOT(20), **LEAF(10)** |
| `lower-weight` | `winner-first` | **LEAF(10)**, ROOT(20) |

Each primitive flips the order independently of the other, in the bytes an agent actually reads —
not in an internal ordering function.

**B4 — the divergent same-named skills, re-attacked.** `deduplicateSkills`
(`import.go:252-275`) names the deciding sentence in code — "one `requires.skills` entry per mapping
skills entry, which a JSON object cannot satisfy for duplicate keys, so the excess mapping entry is
an unmappable detected surface, hence a loss" — and implements it: identical repeats (same git and
revision) collapse with no loss; divergent repeats make the first by ascending environment
identifier the winner and each dropped declaration a loss naming the diverging adapter and
revision. My narrowing mutant **B4** widens "identical" to same-source-only, so exactly a
same-source/different-revision duplicate collapses silently — **KILLED** by
`TestImportDivergentSkillsAreLoss`. That is the precise member of the class cycle 1 found being
dropped in silence.

**The §8.4 class, attacked with two fresh mutants.** The three marker states still produce three
distinct outcomes through `profile import` — absent → the surface is detected and imported; corrupt →
`environment_import_lossy … environment_marker_invalid: malformed protocol JSON`; unreadable →
`environment_import_lossy … permission denied`. Then I narrowed both reads in `readRootSurface`:

| # | mutant | result |
|---|---|---|
| MK1 | collapse a failed **marker** read into absence for exactly one adapter (`envID != "claude_code"`) — cycle-2 C2-B1 restored for one addressing mode | **KILLED** — `TestImportCorruptMarkerIsLoss`, `TestImportUnreadableMarkerIsLoss` |
| MK2 | collapse exactly a **permission-denied** read of the root-context file into absence beside the correct `os.IsNotExist` absence — the same class one line down, which nobody had tried | **KILLED** — `TestImportUnreadableRootIsLoss` |

MK2 is the one worth noting: it is the shape that produced cycle-2 C2-B1 and cycle-4 C4-M1, placed
where a plausible future edit would put it, and the suite catches it.

**The CLI rows, re-driven against `cli/curator.md` at `550579d`.** `curator profile` offers
`install | import | list | use | update | remove | sync | compose` and `curator env` offers
`resolve | status | config`, matching rows 30–41. Driven: `env config set overlay_default_weight 7`
→ `7`, `env config show overlay_default_weight` → `7`, `env config unset …` → `unset
overlay_default_weight`, `env config show` → the whole §12.1 table with its defaults
(`backup_retention: 5`, `overlay_default_weight: 1000`); an unknown knob → `knob: unknown
environments knob "not_a_knob"`. `profile compose default add … --range '^1' --weight 5` →
`added overlay …; the lock moves on profile update` — the row's "Informative … the lock moves only
on `profile update`" verbatim — `compose list` → `https://example.com/ov	range=^1	-	5`, and
`compose remove` → the symmetric message.

**Root-artifact registration, enumerated rather than read off the report.** Only three of this
stage's packages touch `CURATOR_CONFORMANCE_ROOT` at all — `internal/config`,
`internal/envfragment`, `internal/envmarker`; `internal/envprofile` never does, and
`cmd/curator/lifecycle_conformance_test.go` is pre-existing. Every subpath they join is
`schema-cases/manager-config-v2`, `schema-cases/system-config-v2`, `vectors/manager-config-v2.json`,
`schema-cases/launch-env-fragment-v1`, `schema-cases/agent-environment-marker-v1` and the root's own
`schema-cases/index.json`. All five families are covered by `root-artifacts.tsv` rows 39–41, one of
which this stage adds; `index.json` is the root's own manifest, which `suite-plan.sh` itself reads to
decide deferral. Nothing a new package reads is unregistered, which is what makes the `SPEC_PIN`
lane defer instead of fail — the lesson stage (a) paid for.

**No uncalled gate in the delta.** I extracted every function the stage delta *adds* to production
files (81 of them) and counted non-test call sites for each: **all 81 have at least one production
caller.** The detector is not blind — run against `SetCurrent` it correctly reports `prod=0
test=4`. The class that produced stage (a) F9, stage (b) M1 and cycle-1 M1 does not appear here.

**The Windows class, swept myself.** Three consecutive stages shipped a Windows-only defect no Unix
lane could see, so I did not take the producer's "nothing new" sweep on trust. Cross-compilation
first: `GOOS=windows GOARCH=amd64 go build ./...` → exit 0, `go vet ./...` → exit 0, and
`go test -c` for all seven packages this stage touches (`config`, `envprofile`, `envmarker`,
`envfragment`, `contextstore`, `closureexec`, `cmd/curator`) → exit 0 each, so no Windows-only
compile error hides in a test file. Then the pattern sweep over the stage's own 27 Go files:

| pattern | hits | verdict |
|---|---|---|
| `os.Getenv("HOME"\|"USERPROFILE")` | **0** | the operator home is resolved only through `os.UserHomeDir` — ten sites, which reads `%USERPROFILE%` on Windows. This is the `0bcea201` lesson holding. |
| `filepath.IsAbs` | 3 (`main.go:1543`, `import.go:409`, `managed.go:787`) | all three take a **runtime** path from `os.Readlink`/`filepath.EvalSymlinks`, never a POSIX literal — the stage-(b) defect class does not recur |
| bare-`"/"` string ops | 3 | none is a filesystem path: `config.go:983` is URL-path grammar, `environments.go:740-746` is the core §6.3 **git-ref** grammar (where `/` is the ref separator), `envprofile.go:1292` is `isPathOperand`, which explicitly handles `./`, `..\`, drive letters and UNC |
| POSIX-absolute literal in production | 1 (`config.go:238` `/etc/curator/config.json`) | inside the `else` of an explicit `runtime.GOOS == "windows"` branch whose Windows arm uses `%ProgramData%` with a `C:\ProgramData` fallback through `filepath.Join` |
| POSIX-only well-known list | 1 (the §9.5 dotfile locations) | the declared bound, TASK-260906-vlrjo1 |

`sameStoreTree` (`managed.go:790-791`) compares store containment case-**sensitively**, which on
Windows could misread a differently-cased spelling of the same store root — but it is stage (b) code,
both sides derive from the same `storeRoot` string, and the failure direction is fail-**closed**
(refuse and offer takeover, never admit). Recorded, not a finding.

**The §1 path refusals, driven through the reinstall arm that did not exist when cycle 1 checked
them.** Install a clean `path` profile, then corrupt the source and reinstall:

| source mutation | reinstall | pin |
|---|---|---|
| a symlink added under `context/` | exit 1 `profile_source_invalid: context/link.md is not a directory or a regular file` | unmoved |
| a `.git` **below** the root | exit 1 `profile_source_invalid: context/.git carries a .git entry below its root` | unmoved |
| a **root-level** `.git` | exit 0 — correctly **excluded**, not refused | moves (the edit is real) |
| a FIFO under `context/` | exit 1 `profile_source_invalid: context/pipe is not a directory or a regular file` | unmoved |
| `--range` / `--tag` / `--revision` / `--directory` | exit 1 `profile_install_ref_conflict: a path operand takes no requirement flag or --directory` | — |

Every refusal leaves the pin where it was, so a corrupted source cannot half-publish. Note the
`--directory` row: production **does** refuse it, so cycle 5's surviving `my3` mutant is a coverage
gap and not a live hole — I confirmed that by driving rather than reading.

**The B2 class, re-attacked through rework 5's new write path.** The most serious finding of this
chain was a `--takeover` writing *through* a foreign-manager symlink into a file outside every
managed home. Rework 5 routes a new caller into that write path (`activateReinstall` → `useLocked`
→ `materializeScope`), so I re-drove it there, with foreign symlinks in **two** adapter homes at
once — one copied adapter, one linked — and the machine current cleared so the activation comes
from the first-install clause:

```
before:  claude CLAUDE.md -> /…/foreign/CLAUDE.md   (symlink)
         codex  AGENTS.md -> /…/foreign/AGENTS.md   (symlink)

$ curator profile install ./s --as tk                  [EXIT 1]
    claude_code: environment_foreign_manager_detected: CLAUDE.md is a symlink outside the manager store; abort, or take over with backup
    codex_cli:   environment_foreign_manager_detected: AGENTS.md is a symlink outside the manager store; …
    → both surfaces untouched, both foreign files untouched, current unmoved

$ curator profile install ./s --as tk --takeover       [EXIT 0]
    claude CLAUDE.md  regular file, managed content, NOT a symlink
    codex  AGENTS.md  symlink -> $HOME/profiles/tk/rendered/codex_cli/AGENTS.md   (inside the manager home)
    foreign/CLAUDE.md  FOREIGN CLAUDE SOURCE OF TRUTH   unchanged
    foreign/AGENTS.md  FOREIGN CODEX SOURCE OF TRUTH    unchanged
    backups: claude=[1] codex=[1], generation 1 holding the foreign bytes
```

The surviving `codex_cli` symlink is **curator's own**, not the foreign one: I read its target and
confirmed it resolves inside the manager home — that is the linked-adapter shape `replaceLink`
produces, and the foreign link was removed rather than written through. Both adapters backed up
before the first write. The B2 fix holds in both the copied and the linked shape through a caller
that did not exist when it was made.

**The import, re-attacked as a gate.** The consent gate is not merely "not config-reachable by the
knob names I guessed" — it is unreachable **by construction**: the schema-2 `environments` object is
a closed field set, and every candidate spelling I tried
(`environments.allow_lossy`, `.import_consent`, `.allow_lossy_import`) is refused at load with
`environments: has unsupported field "…"` before the import runs. None of the eighteen §12.1 knob
names is a consent knob. A lossy import stops at exit 1 with `environment_import_lossy` naming the
loss (`claude_code …/skills/foo: names no recoverable exact declaration`), and `--allow-lossy`
re-reports and proceeds. Reassembly, driven with deliberately mixed line endings: a CRLF source
whose tail is `\r\n\r\n\r\n` and a CR-only source both land as `…INE TWO\n` — the CRLF/CR-to-LF
normalization with **exactly one** trailing LF, applied at reassembly. Modules are named and
ordered by ascending environment identifier (`claude_code.md`, `codex_cli.md`). And the import
itself writes nothing into any native home: both runs exit 1 only on `profile_use_partial`, because
§9.1 auto-activation on a machine with no current is what meets the unmanaged native file — the
reassembly and store entry are already complete by then.

**Things nobody had driven yet, all correct.** `profile use --clear --env claude_code` against
unmanaged state → exit 1 with `environment_surface_unmanaged_conflict`; with `--takeover` → exit 0,
notice, backup generation **2** beside the existing generation 1, surface re-materialized.
`env resolve --repair` and `--repair --takeover` both repair the managed home and touch no native
file, which is right — the repair writes nothing outside the manager tree, so there is nothing for
`--takeover` to take over. `env status` reports `require_current_profile: acme (locked)` as §12.2
requires. `TestWeightRulesAllFourDisagree` asserts weight **40**, `Overlay: true` and
`RequiredBy: [mid1, mid2]` — exactly what cycle 3 measured by hand — and drives `Install`, and its
`Policy` is one `PolicyFromConfig` can actually produce (a git overlay with a form), unlike the
path-overlay case the bound covers.

---

## Judging the rework-5 report against the attestation standard

Accurate on everything I could check, which is nearly everything. The five finding→resolution rows
are real; the four `--use`/`--takeover` combinations are driven through `run()` and the quoted doc
comment is the one in the tree; the C5-m1 mutants my1/my4 are genuine kills and I reproduced both;
the four audit-gate mutants are genuine and I reproduced two of the four myself and built the
`updateLocked` twin independently; the skip enumeration's result reproduces under my own
classifier; the reissued bound names a task that exists and the ledger rows say what their tests
assert; the CI row honestly states the head was not pushed when the report was written; and the
"no `git add -A`" discipline held — `git status --short` empty, no stray artefacts, three signed
commits touching only named paths.

Two rows I would not sign as written:

1. **The git-reinstall bound's "undrivable hermetically here."** False, and I drove it. See C6-m1.
   The rest of that bound is honest.
2. **"AC ratio: 15 of 15 stage (c) rows driven through the production entry point."** The path
   install row is driven in every addressing mode that matters *for a path operand*, but the same
   `cli/curator.md` row also spells `<git-url> … [--use] [--takeover]`, and for a git operand the
   retry after the §9.5 stop is not driven and does not work. I read that as **14½ of 15**, the
   half being `profile install <git-url> --use --takeover` after the stop — the same half cycle 5
   docked, moved one addressing mode over.

Neither is a fabricated claim and neither hides a gate. Both are corrections to the record.

---

## Gate table (standalone processes, on `8b8aa041`)

| Gate | Root | Exit | Output |
|---|---|---|---|
| `go build ./...` | — | 0 | empty |
| `go vet ./...` | — | 0 | empty |
| `gofmt -l cmd internal` | — | 0 | empty |
| `golangci-lint run ./...` | — | 0 | `0 issues.` |
| `bash .github/ci/gate-selftest.sh` | — | 0 | `gate-selftest: 94 passed, 0 failed` |
| `bash .github/ci/no-broad-suppression.sh` | — | 0 | `no-broad-suppression: ok` |
| `bash .github/ci/ledger-consistency.sh /tmp/c6/ledger-ev` | — | 0 | `212 rows checked across linux darwin windows`, `ok` |
| `GOOS=windows GOARCH=amd64 go build ./...` | — | 0 | empty |
| `GOOS=windows GOARCH=amd64 go vet ./...` | — | 0 | empty |
| `GOOS=windows go test -c` × 7 stage packages | — | 0 each | no Windows-only test compile error |
| `go build ./cmd/curator` (`/tmp/c6/head`) | — | 0 | binary produced |
| `go build ./cmd/curator` (`/tmp/c6/main`, `origin/main` `db444157`) | — | 0 | comparison binary |
| `go test -count=1 ./internal/envprofile/ ./internal/config/` | unset | 0 | `ok 69.974s`, `ok 0.919s` — the kill/survive oracle |
| `go test -count=1 -timeout 30m ./cmd/curator` (run solo) | unset | 0 | `ok 311.812s` |
| `go test -count=1 -race ./internal/envprofile/ ./internal/config/ ./internal/envmarker/ ./internal/envfragment/` | unset | 0 | `ok 85.002s / 1.747s / 2.845s / 2.376s` |
| 20 narrowing/widening mutants (tables above) | unset | mixed | **17 KILLED**, 3 survive (`r2` → C6-m2; `OV1` proved unreachable by driving; `OV2` inside the declared M1 bound) |
| `gh pr checks 61` | — | — | **11 of 11 pass** on `8b8aa041` |
| `gh api …/artifacts/9996710714/zip` (candidate-evidence-ubuntu) | curator-spec `550579d` | 0 | `served=71 deferred=0`, 0 fail events |
| `gh api …/artifacts/9996828237/zip` (candidate-evidence-macos) | curator-spec `550579d` | 0 | `served=72 deferred=0`, 0 fail events |
| `gh api …/artifacts/9998022110/zip` (candidate-evidence-windows) | curator-spec `550579d` | 0 | `served=72 deferred=0`, 0 fail events, 5465 passing cases |
| cycle-4 `P2` / `P4` mutants, re-run (not taken on trust) | unset | 1 | both KILLED |
| `classify.sh` over the stage's 9 distinct skip reasons | — | 0 | 8 classify, 1 (`no git on PATH`) UNCLASSIFIED |
| `classify.sh` against the `4a8a1d42` registry | — | 0 | reproduces the red Windows lane |

**Not run by me, stated plainly:** the two full `test-gate.sh` lanes as one process. The hosted `Test (ubuntu/macos/windows)`
lanes execute the default-root lane and the full `cmd/curator` suite on this exact head and are
green on all three platforms; the candidate dispatch executes the `CI_REQUIRE_FULL_ROOT=1` lane
against the task authority on all three, and I read two of its three evidence artefacts directly.
That is strictly better than a fourth local rerun on one host. A gate I did not run is listed as
not run.

---

## Is stage (c) safe to land?

**Yes. Land it.**

Six cycles, twenty-two findings. Everything the previous five cycles found is fixed and stays fixed
under **my own** mutants rather than the producer's — twenty applied, seventeen killed, and of the
three survivors one is a coverage minor (C6-m2), one I proved unreachable by building it and driving
it (OV1), and one sits inside a bound already declared and filed (OV2). Concretely:

- the §1 snapshot is immutable across `update`, `sync` and `use` and refreshes on reinstall in both
  the `--as` and bare addressing modes, and every §1 path refusal — symlink, `.git` below the root,
  special file, requirement flags — fires through the *new* reinstall arm with the pin unmoved, while
  a root-level `.git` is correctly excluded rather than refused;
- the §9.2 seam is structural, single, and covers rework 5's new caller: a reinstall that activates
  under a locked `require_current_profile` is refused at exit 1 naming the knob, and `CurrentFile`
  still has exactly one production writer;
- the §12.2 pin transcribes the specification independently — I checked the literal against
  `git show 550579d:protocol/environments.md` — and dies under both a widening that would reach §9.1
  secret material and a narrowing that would drop a fleet-policy knob;
- the §9.5 retry works for a path root with the takeover genuinely reaching the switch (proved by a
  mutant that strips `Takeover` from the policy handed to `useLocked`), and the B2 class holds
  through that new write path in both the copied and the linked adapter shape, with the foreign
  file's bytes intact and a backup written before the first write;
- the §8.4 class dies under two fresh mutants including one nobody had tried, the §9.6 divergent-skill
  loss dies under a third, the `overlays_allowed` forbid dies under a fourth, and both blocking-audit
  gates die under the reinstall and update twins;
- the two precedence primitives flip the emission order independently **in the materialized bytes**;
- not one of the 81 functions this stage adds is unreachable from production, every root artefact the
  new packages read is registered, and every skip the stage introduces classifies but one benign dead
  branch;
- the Windows class does not recur: zero `os.Getenv("HOME")`, every `filepath.IsAbs` fed a runtime
  path, every bare-`"/"` operation on a URL or git-ref grammar rather than a filesystem path, and the
  one POSIX-absolute production literal inside an explicit `runtime.GOOS` branch;
- and locally `go build`, `go vet`, `gofmt`, `golangci-lint`, `gate-selftest.sh` (94/94),
  `no-broad-suppression.sh`, `ledger-consistency.sh` (212 rows), the `-race` suite on all four stage
  packages and the full `cmd/curator` suite all pass, on top of eleven green hosted lanes and a
  candidate lane that serves the environments families against the task authority with `deferred=0`
  and zero failures on **all three** runners, Windows included.

The two findings are minor by construction. **C6-m2** is one missing subtest over correct, driven
behaviour. **C6-m1** is a real defect but **not this stage's** — I drove it on `origin/main` and got
the identical result, the rework-5 brief scoped the fix to path roots deliberately, and rejecting
stage (c) would leave the defect on trunk either way. What stage (c) owes is one corrected sentence
in a bound and a follow-up task, both of which belong in the record rather than in a seventh cycle.

Recommended follow-ups, neither blocking:

1. **A trunk task** for `profile install <git-url> --use [--takeover]` on a same-source reinstall:
   honour the flags in the git arm as `reinstallPathLocked` now does for path, or refuse the
   combination loudly. Never the third outcome. `probe-git.sh` in the archive is the repro, and it
   ships the fixture that makes it drivable — the bound should stop saying it cannot be.
2. **A trunk task** for the three §8.4 collapse sites outside the delta — `switch.go:709`
   (`purgeHomes`), `status.go:432` and `status.go:521` — closed as a class, since this is now three
   sites in one file family and vigilance has been asked to carry it for four cycles.
3. **One subtest** for C6-m2, on the next touch of `reinstallActivation`.
