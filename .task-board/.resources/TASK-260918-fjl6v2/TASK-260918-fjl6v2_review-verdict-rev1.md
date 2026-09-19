# TASK-260918-fjl6v2 — review verdict, revision 1 (story-final Change Request `CR-TASK-260918-fjl6v2-1`)

Verdict: **accepted**, revision 1 — routed with `accept_cr` to `integrating`;
acceptance is not landing. Reviewer run `RUN-260918-a93e60` (claude-opus-5),
not goal-bound, no directives. No code edits, no commits, nothing written
into the Story worktree; every build, test and mutant ran in disposable
clones under `/tmp/fjl6v2-review/` with
`CURATOR_CONFORMANCE_ROOT=/tmp/spec-rc12-review/conformance/v1`
(`git worktree add` of curator-spec at `dced9b8317e0e8af79edf2d0539b32bd22b6c85b`
= tag `v1.0.0-rc.12`; `manifest.json` sha256
`ea9dd5a0030b889079cf5655517056e66a0f1d61812f8c2920b4beddc1fd24ed`, the digest
the candidate's `ci.yml` comment names).

Scope of this review (per the review brief): prove the candidate is exactly
"trunk + the accepted rc.12 union + the stated combination work" and that the
combination is correct. The union's design was accepted at
`TASK-260917-16l2md` rev 5 and is not re-reviewed here.

## 1. What was reviewed

| Fact | Value |
|---|---|
| Story worktree HEAD | `1c464c56d5ba05b759b8dd879a7a5b816bfb8551` (the CR base OID; current trunk) |
| Candidate tree | `b7dfd84c9a4095b110a1b3ac15465e49de0bf567` (CR record) |
| Reconstruction | `git clone --no-checkout` of the control root → `git checkout --detach 1c464c5` → `git apply --index TASK-260918-fjl6v2_change-request_rev1.patch` (sha256 `e03a0085…25db7b2`, matches the CR record) → `git write-tree` = `b7dfd84c…` |
| Clone vs worktree | `diff -rq --exclude=.git <worktree> <clone>` → exit 0 (byte-identical, no extra files either side) |
| `git status --short --untracked-files=all` in the worktree | exactly 42 lines, all `M`/`A`; `--ignored` adds nothing; 14 new files all mode `100644`; no mode changes, no executables |
| Union input | `rc12-union-73fc8a4-vs-3c45d4b.patch` sha256 `18a5cd68…3c489d`, byte-identical to `TASK-260917-16l2md_change-request_rev5.patch` **and** to `git diff 3c45d4b 73fc8a4` recomputed from the object database; whole-patch `patch-id --stable` `ccc9574a4bb4d6191c8f3b1ef829882092489946`; `73fc8a4^{tree}` = `559e88e4…` = the tree accepted at rev 5 |
| Prepared inputs | the four `TASK-260918-11f9l1_*` resources attached to this task are byte-identical to the originals on `TASK-260918-11f9l1`; sha256 of the three resolutions = the §6 table of `TASK-260918-11f9l1_results.md` |
| Hosted gate (rev 1) | run `35368185420`, commit `787e1e1c…` whose tree is `b7dfd84c…` (verified with `git cat-file -p`): conclusion `success`; read independently through `gh run view` — see §6 |

## 2. Identity, file by file

Per-file `git patch-id --stable` of `git diff 3c45d4b 73fc8a4 -- <file>` (union)
vs `git diff HEAD -- <file>` (candidate), all 40 union files — my table
reproduces the producer's exactly:

| File | Union patch-id | Candidate patch-id | Verdict |
|---|---|---|---|
| .github/workflows/ci.yml | d51ea717e5ef | d51ea717e5ef | IDENTICAL |
| CHANGELOG.md | dde4c9eb24e0 | 24eab169007f | DIFFER (resolved, §3) |
| cmd/curator/env.go | 34809ba5f12f | 34809ba5f12f | IDENTICAL |
| cmd/curator/env_test.go | c42d7602ff15 | c42d7602ff15 | IDENTICAL |
| cmd/curator/envconfig_test.go | e12547829c5e | e12547829c5e | IDENTICAL |
| cmd/curator/envstatus.go | de5f4651d440 | e6cc447f9f96 | DIFFER (resolved, §3) |
| cmd/curator/main.go | faa95dd84ae2 | faa95dd84ae2 | IDENTICAL (auto-merge, §3) |
| cmd/curator/profile.go | 9fe7facee243 | 9fe7facee243 | IDENTICAL |
| cmd/curator/profile_surfacing_test.go | 27e9d42d3ef0 | 27e9d42d3ef0 | IDENTICAL |
| cmd/curator/umbrella.go | 5fb66c06f33b | 5fb66c06f33b | IDENTICAL |
| cmd/curator/umbrella_conformance_test.go | 2b260a886e11 | 2b260a886e11 | IDENTICAL |
| cmd/curator/umbrella_test.go | 4a3cac8fcd90 | 4a3cac8fcd90 | IDENTICAL |
| cmd/curator/umbrella_trustroot_windows_test.go | 246efc59f7f8 | 246efc59f7f8 | IDENTICAL |
| docs/environment-config.md | f9a85f4eff96 | f9a85f4eff96 | IDENTICAL |
| internal/config/config.go | e3e5d86dd93a | e3e5d86dd93a | IDENTICAL |
| internal/config/environments.go | 1adcd62a9cc3 | 1adcd62a9cc3 | IDENTICAL |
| internal/config/environments_conformance_test.go | de10c4db0cfa | de10c4db0cfa | IDENTICAL |
| internal/config/environments_test.go | d91bd230bfe6 | d91bd230bfe6 | IDENTICAL |
| internal/config/system_module_schema_test.go | cdd40b98a1ff | cdd40b98a1ff | IDENTICAL |
| internal/contextmaterialize/admission.go | 44d2ff25995d | 44d2ff25995d | IDENTICAL |
| internal/contextmaterialize/admission_test.go | b981f03ed95e | b981f03ed95e | IDENTICAL |
| internal/contextmaterialize/contextmaterialize.go | 601cefe5e2f7 | 601cefe5e2f7 | IDENTICAL |
| internal/contextmaterialize/mcp.go | 5b33e2790e5e | 5b33e2790e5e | IDENTICAL |
| internal/contextmaterialize/mcp_test.go | f43b69a5a408 | f43b69a5a408 | IDENTICAL |
| internal/contextmaterialize/system_module_admission_test.go | 4480c2e7323d | 4480c2e7323d | IDENTICAL |
| internal/envfragment/envfragment.go | a8fb5197ba28 | a8fb5197ba28 | IDENTICAL |
| internal/envfragment/envfragment_test.go | 754e4962711b | 754e4962711b | IDENTICAL |
| internal/envprofile/admission_test.go | 046d167a59dc | 046d167a59dc | IDENTICAL |
| internal/envprofile/envpassthrough_conformance_test.go | 0b3177167a02 | 0b3177167a02 | IDENTICAL |
| internal/envprofile/envprofile.go | df507404a050 | df507404a050 | IDENTICAL |
| internal/envprofile/import.go | dd8ec5b0eb42 | dd8ec5b0eb42 | IDENTICAL |
| internal/envprofile/managed.go | 30c2edb49488 | 30c2edb49488 | IDENTICAL |
| internal/envprofile/status.go | 5d923f007212 | 40e59b1b3f6a | DIFFER (resolved, §3) |
| internal/envprofile/surfacing.go | fcd6df0f1790 | fcd6df0f1790 | IDENTICAL |
| internal/envprofile/surfacing_order_test.go | 915a1e39f1f6 | 915a1e39f1f6 | IDENTICAL |
| internal/envprofile/surfacing_test.go | 36113721090b | 36113721090b | IDENTICAL |
| internal/envregistry/envregistry.go | 8e15d91303aa | 8e15d91303aa | IDENTICAL |
| internal/globalbins/globalbins.go | dac49f9b94ed | dac49f9b94ed | IDENTICAL |
| internal/globalbins/globalbins_test.go | 13e9075fdeef | 13e9075fdeef | IDENTICAL |
| internal/interop/environments/context_materialization_test.go | 8ea1153316d8 | 8ea1153316d8 | IDENTICAL |

IDENTICAL = 37, DIFFER = 3 (exactly the three resolved files). Stronger than
patch-id: the trunk touched only four of the union's 40 files since `3c45d4b`
(`git diff --name-only 3c45d4b 1c464c5` ∩ union = `CHANGELOG.md`,
`cmd/curator/envstatus.go`, `cmd/curator/main.go`,
`internal/envprofile/status.go`), and `git diff --cached --name-only 73fc8a4`
over the 40 union paths + the 2 fix-up paths lists exactly those four plus the
two fix-up files — so the other **36 union files are byte-identical to the
accepted checkpoint `73fc8a4`**.

Changed-set identity: `git status` paths = union 40 ∪ {`cmd/curator/hook_posture_test.go`,
`cmd/curator/hook_test.go`}; `comm` in both directions shows nothing else.

Fix-up: `git diff HEAD -- cmd/curator/hook_posture_test.go cmd/curator/hook_test.go`
is **byte-identical** to `TASK-260918-11f9l1_fixup.patch` (`diff` empty;
patch-id `63ddb9ed0bb49aef6278ed0a684dc495f0a0e204` both), `+17/−0` and
`+16/−0`, test files only. Whole candidate: 42 files, +6812/−219.

## 3. Three-way union evidence for the four merge-surface files

For each file: `base` = `3c45d4b:<file>`, `trunk` = `1c464c5:<file>` (verified
`== 64cacfc:<file>`, i.e. the S6 squash is the only trunk change to each),
`union` = `73fc8a4:<file>`, `resolved` = candidate. The ordered sequence of
`+`/`−` lines (`git diff --no-index -U0`, headers stripped) was compared in
both directions; `patch-id` differs here only because context lines differ,
which is expected for an interleaved merge.

| File | diff(base→trunk) vs diff(union→resolved) | diff(base→union) vs diff(trunk→resolved) | Candidate == prepared resolution |
|---|---|---|---|
| CHANGELOG.md | 27 lines, identical sequence | 67 lines, identical sequence | byte-identical (sha256 `38b42f0d…5b1ca9`) |
| cmd/curator/envstatus.go | 6 lines, identical sequence | 91 lines, identical sequence | byte-identical (`ceeccf45…c304cf`) |
| internal/envprofile/status.go | 75 lines, identical sequence | 222 lines, identical sequence | byte-identical (`d696023d…a62090`) |
| cmd/curator/main.go (auto-merge) | 40 lines, identical sequence | 6 lines, identical sequence | n/a — git auto-merge; `trunk→resolved` is exactly the union's `cmdUmbrella(cfg, home, args)` dispatch hunk |

So every S6 line and every union line is present verbatim and in its own
side's order, and nothing else was added: "union of both sides, nothing
dropped, reordered or reworded" holds for all four. Placement (read from the
`union→resolved` diffs): in `envstatus.go` the six S6 printer lines sit
directly after the `require_current_profile` block and before the union's
`// §12 posture` comment; in `status.go` the two S6 imports are alphabetical,
the two S6 `Status` fields follow `NonCurrent` and precede the union's
`S4Profile`, the S6 assessment block sits between the union's
orphans/`NonCurrent` line and the homes sort, and the two S6 helpers follow
`StatusOf`; in `CHANGELOG.md` the S6 entry keeps its trunk position (after
the schema-8 module-roots entry) and the union's E2/S4/pin entries follow it,
E4 Added/Fixed auto-merged. gofmt-clean (`gofmt -l cmd internal` empty).

## 4. Pin, ledger, hygiene

- `.github/workflows/ci.yml:53` `SPEC_PIN: dced9b8317e0e8af79edf2d0539b32bd22b6c85b`,
  referenced by `${{ env.SPEC_PIN }}` in all five suite checkouts (`test`,
  `test-self-hosted`, `race`, `interop`, `candidate-conformance`; lines
  113/255/377/510/604). The `ci.yml` delta is the union's pin hunk only
  (patch-id identical).
- `.github/workflows/release.yml`: unchanged vs trunk and unchanged on trunk
  since `3c45d4b`.
- `.github/ci/platform-cases.tsv`: not in the candidate (unchanged vs trunk)
  and unchanged by the union (as accepted at rev 5). The trunk's own +9 rows
  since `3c45d4b` are the S6 `internal/shell TestShellHookTrustVectors`
  rows landed by `64cacfc`, not this candidate's doing. No new skip rows.
- Rule 8: no executables, `.review/`, `LOGBOOK.md`, coverage or evidence
  files in the candidate or the worktree (see §1); `go.mod`/`go.sum`
  untouched; nothing left in the worktree by this review (all work under
  `/tmp/fjl6v2-review`, spec worktree `/tmp/spec-rc12-review`).

## 5. Combination correctness

**Closed output order.** Read from the candidate's `printEnvStatus`
(`cmd/curator/envstatus.go`) and confirmed against the call sites: after the
optional `require_current_profile` line, `env status` prints (1) the S6
`shell-hook-trust: warning:` rows, (2) the S6 `shell-hook-trust: <path> …`
rows, (3) the union `s4_profile: …, passable_env_names: …` line, (4) the union
`warning: …` machine rows, (5) the union `scope <s> profile <p>
mcp-declarations:` blocks, then the unchanged homes/scopes/adapters rows, the
union `provider <name>: …` rows, then targets, profiles (with the union
`transitive_system_modules` and dropped-module rows), unregistered, orphans,
notes — exactly the order the producer's results §3 state. `curator status`
(`main.go:700` `cmdStatus`) never calls the env printer; its only posture rows
are the S6 hook rows printed after the per-target rows, and the union's
`main.go` hunk touches only the umbrella dispatch — so `curator status` is
the S6 trunk behaviour, unchanged by the combination. No committed test pins
the relative order of the S6 block against the union block (both suites use
`strings.Contains`), so the order is a documented, code-defined choice;
rc.12 `profiles/manager.md` §8.6 only requires the posture be reported and
§10 orders nothing, and the chosen order is consistent with the later
spec-main §10 closed inventory (`hook-trust` first, then `env-passthrough`,
`transitive-system-modules`, `provider-trust-roots`, `mcp-package-allowlist`).

**Fix-up rationale, reproduced.** In the disposable clone with the fix-up
reverted (`git apply -R TASK-260918-11f9l1_fixup.patch`; both test files then
equal trunk), `go test ./cmd/curator/ -count=1 -run '^(TestEnvStatusReportsShellHookTrustPosture|TestEnvStatusMissingAndUnreadableKeepRecord|TestEnvStatusUnreadableApprovalStateSurfaced|TestEnvStatusMatrix)$' -v`
→ exit 1: **exactly the three S6 tests fail**, each at "env status --check
without approvals = 1", and their captured stdout shows why:
`provider session: missing (subcommand_provider_missing, non-current)` (this
host has `/usr/local/bin/curator-run` but no `curator-session`; a hosted
runner has neither). The union control `TestEnvStatusMatrix` passes. With the
fix-up restored (`diff -rq` clone == candidate again) the same four tests pass
(exit 0). The stub block is verbatim the union's own `TestEnvStatusMatrix`
setup; no production code is touched.

**Combination side-effect worth recording.** rc.12 publishes
`vectors/shell-hook-trust.json`, so the trunk's S6 driver
`internal/shell TestShellHookTrustVectors` (ledger class `root-content`)
stops skipping once this candidate's pin lands: locally it ran at the rc.12
root and passed (9.99 s; 40 subtests pass, 6 `pwsh` host-capability skips),
and the hosted lanes ran it too (all green). The rev-5 verdict had noted the
S6 family as outside the union; the landing makes it live.

## 6. Gates

Disposable clone, `set -o pipefail`, zsh, Go `go1.26.0 darwin/amd64`, root
exported as above:

| Gate | Exit |
|---|---|
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `gofmt -l cmd internal` (and `gofmt -l .` minus the tracked `.task-board` resource copies) | 0, empty |
| `GO_TEST_TIMEOUT=90m EVIDENCE=/tmp/fjl6v2-review/evidence make ci-test` (= `bash .github/ci/test-gate.sh`, `go test -json -count=1 -timeout 90m` over the planned 75 served packages, then `platform-case-gate.sh`) | **2** — see below |

The full local gate ran end to end (21:28–22:59 local, `suite-plan:
served=75 deferred=0 excluded=0`): **66 packages pass, 3 no-test-file skips,
6 packages fail**. Every failure has a process-launch stall signature and none
of the failing tests is in a file the candidate touches: `internal/artifactpolicy`
("wait for selected Go executable: signal: killed"), `internal/buildrepo`
("git [--exec-path]: signal: killed" after the 20 s probe deadline, every
`TestResolvedTransportClosedGrammarTable` group), `internal/godriver` ("Go probe exceeded its deadline"),
`internal/npmsource` ("portable process exceeded its deadline"),
`internal/install` (`TestDraftTransportAuth` "trusted Git version probe
failed", `TestAuthoritativeDryRunCasesMutateNothingPersistent` 1200 s), and
`cmd/curator` ("panic: test timed out after 1h30m0s" — `sample` of the hung
binary showed its stub `/bin/sh …/git --exec-path` child, spawned by
`TestProductionExternalDepsFalseDrivesMirrorFetchToSink`, sitting in
`_dyld_start` with 0 CPU for >70 min). The same host pathology stalled this
review's own shell for ~20 min at 21:31–21:51 (an `uptime` needed >300 s to
exec; other sessions' `go test`/mutation binaries were running; load 10–23 on
6 cores). It is the exec-stall the producer's results §8 describe, not a
candidate defect. The ledger then also reported the 4 pre-existing
host-capability skips (`internal/pnpmsource` ×2, `internal/rustsource` ×2,
untouched packages, reproduced on bare trunk by the producer) plus the
cmd/curator cases that never ran because of the timeout.

Re-runs of the failed packages, one at a time, same flags
(`go test -json -count=1 -timeout 90m ./<pkg>/`), in a quieter window:

| Package | Exit | Tests |
|---|---|---|
| internal/artifactpolicy | 0 | 59 pass, 0 skip (188 s) |
| internal/buildrepo | 0 | 105 pass, 2 skip (131 s) |
| internal/godriver | 0 | 125 pass, 5 skip (165 s) |
| internal/npmsource | 0 | 23 pass, 0 skip (53 s) |
| internal/install | 0 | 221 pass, 2 skip (519 s) |
| cmd/curator (whole package, `-timeout 150m`) | 1 | 240 pass, **3 fail** — `TestCompiledProjectRepairsCorruptCompiledState`, `TestCompiledProjectRestoresCacheWhenCommitFails`, `TestCompiledProjectStatusAndUntrustedRecovery`: `no space left on device` on the boot volume (214 MiB free at 23:39 local; another session freed it minutes later) and `go-v1 go_build_failed` of the fixture build during that window; compiled-build tests in untouched `status_test.go`/`global_status_test.go` |
| cmd/curator, those 3 tests (`-run`) | 0 (third attempt) | 3 pass (1627 s; attempt 1 died on a transient `package math/big is not in std`, attempt 2 on a concurrent `go clean -cache` by another session that emptied `GOCACHE` mid-build — both host events, both outside the candidate) |

The union's vector drivers executed (not skipped) at the rc.12 root in the
local stream: `internal/config TestManagerConfigV2Vectors`,
`internal/contextmaterialize TestSystemModuleAdmissionVectors`,
`internal/envprofile TestEnvironmentsEnvPassthroughVectors` — all pass;
`cmd/curator TestUmbrellaProviderResolutionVectors` passed in the whole-package
cmd/curator re-run, as did all 52 union/S6-relevant cmd/curator tests there
(`TestUmbrella*`, `TestEnvStatus*`, `TestEnvConfig*`, `TestHook*`,
`TestStatus*ShellHookTrust*`, `TestProfile*SurfacesBeforePublication`,
`TestProviderPostureRows`, `TestActiveRevisionIsWarningRelease`: 52 pass, 0 fail).

**Hosted gate, read independently** (`gh run view 35368185420`): status
`completed`, conclusion `success`, head `787e1e1c…` (tree `b7dfd84c…` = this
candidate). Job logs: Test (macos-latest) `suite-plan: served=75`, `go test
overall exit=0`, `platform-case gate: 22 skips recorded … ok`, `go test
exit=0, platform-case gate exit=0` (whole suite in 8 min on a clean runner);
Test (ubuntu-latest) `served=74 excluded=1`, both exits 0, 52 skips; Test
(windows-latest) `served=75`, both exits 0, 175 skips; Race (ubuntu/macos)
both exits 0; Lint `golangci-lint found no issues`; Interop conformance gate,
Gate self-tests ×3, Naming gate all `success`; Candidate suite and rose-air
lanes `skipped` (optional). The hosted gate is the arbiter per the campaign
rules and it is green on exactly this tree.

## 7. Narrowing mutants (disposable clone, restored from byte copies, `diff -rq` clone == candidate afterwards)

| Mutant (file, side) | Killed by |
|---|---|
| M1 `envstatus.go` (S6): printer hides non-current trust rows (`if row.NonCurrent() { continue }`) | `TestEnvStatusMissingAndUnreadableKeepRecord` ("env status lacks 'shell-hook-trust: …'"), `TestEnvStatusReportsShellHookTrustPosture` ("lacks the changed row") — exit 1 |
| M2 `envstatus.go` (union): printer hides `outside_trust_roots` provider rows | `TestEnvStatusMatrix` ("status rows miss 'provider run:'") — exit 1 |
| M3 `status.go` (S6): only a changed file makes the matrix non-current (`row.State == PostureChanged` instead of `row.NonCurrent()`) | `TestEnvStatusMissingAndUnreadableKeepRecord` ("env status --check with file missing = 0, want 1") — exit 1 |
| M4 `status.go` (union): a scope with a single declaration row is dropped (`len(rows) < 2`) | `internal/envprofile TestStatusS4Posture` ("declarations = [], want one scope") — exit 1 |
| M5 `status.go` (union): `PassableEnvNames` always nil | `TestStatusS4Posture` ("passable_env_names = []") — exit 1 |
| M2b `envstatus.go` (union): `attachProviderPosture` folds only **refused** providers into `NonCurrent` (missing/unreadable no longer non-current) | **survived** `TestEnvStatusMatrix`, `TestUmbrellaMissingProvider`, `TestEnvStatusReportsShellHookTrustPosture` (exit 0) |

5/6 killed. M2b is a bound, not a defect of this task: the accepted union
pins the missing/unreadable verdicts at the `providerPostureFor` seam
(`umbrella_test.go:427/440`) but has no CLI-level negative test that
`env status --check` fails on a missing provider; the fix-up-revert probe in
§5 shows the production folding is live (the S6 tests failed on exactly it),
and it existed at rev 5 unchanged. Recommended follow-up (outside this
landing): one committed `env status --check` negative for a missing
`curator-session`.

## 8. Findings

None blocking. Observations for the record:

1. M2b coverage bound above (accepted-union scope, follow-up test suggested).
2. The producer's results §5 "full suite green locally" was obtained by
   partitioning `cmd/curator`/`envprofile`/`install` across `-run` masks; on
   this host a single-process `make ci-test` cannot be relied on to finish
   without exec stalls (this review's run needed per-package re-runs). The
   hosted gate on the exact tree is the evidence that carries AC 3.
3. `cmd/curator` and `internal/install` test binaries serialize on the
   host-wide `$TMPDIR/curator-host-goroot-test-lock-v1`; a hung cmd/curator
   binary from any session blocks every other cmd/curator run on the machine
   for its whole timeout ("acquire package host GOROOT test lock: context
   deadline exceeded" after 5 min) — worth knowing for future local gates.

## 9. AC mapping

1. Candidate = trunk `1c464c5` + 40-file union delta (37 patch-id identical,
   36 byte-identical to `73fc8a4`; the 3 resolved files byte-identical to the
   11f9l1 resolutions and proven to be the union of both sides; `main.go`
   auto-merge = checkpoint + S6 hunks = trunk + union hunk) + the fix-up
   (byte-identical, +33/−0, tests only) — §2, §3. ✓
2. `SPEC_PIN dced9b8…` unchanged, no other change, rule 8 — §4. ✓
3. build/vet/gofmt clean; full suite: hosted gate green on all lanes for
   this tree; locally every served package has passed at the rc.12 root in
   this review — 66 in the single gate run, 5 on per-package re-run,
   cmd/curator as 240/243 in a whole-package re-run plus the 3 disk-full
   casualties in a targeted re-run (§6). The single-process `make ci-test`
   itself exited 2 on this host for the environmental reasons stated. ✓
4. Identity proof, closed output order, gate transcripts, hygiene — §2–§6. ✓

Evidence directory (not attached; regenerable): `/tmp/fjl6v2-review/`
(`patchid-table.md`, `threeway.log`, `ci-test.log`, `evidence/test/*.json`,
`rerun/*.json`, `mutants/*.log`, `hosted-*.log`).
