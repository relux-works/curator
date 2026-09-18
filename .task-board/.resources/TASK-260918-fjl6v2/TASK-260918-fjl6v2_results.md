# TASK-260918-fjl6v2 results — land the accepted rc.12 union on the current trunk (story-final)

Story `STORY-260918-2yvd86` (conformance-pin-rc12-landing), `EPIC-260910-2hw1xb`.
Base commit (fresh Story worktree, recorded per brief): `1c464c5`
`1c464c56d5ba05b759b8dd879a7a5b816bfb8551`
"Merge PR #78: BUG-260918-18ny6p find a Homebrew rustup in the rose-air lane"
— one of the two trunks `TASK-260918-11f9l1` validated (`6401d3c`, `1c464c5`).
Worktree was clean at start (`git status --short` empty).

## 1. Construction transcript

All commands in `<control-root>/.temp/STORY-260918-2yvd86/worktree`,
branch `task-board/story/STORY-260918-2yvd86`. No commit on the Story branch.

1. `git apply --3way --index rc12-union-73fc8a4-vs-3c45d4b.patch` — exit 1
   (expected: applied with conflicts). Conflicts in EXACTLY the three
   expected files, no new conflict (trunk did not move past `1c464c5`):
   `CHANGELOG.md`, `cmd/curator/envstatus.go`, `internal/envprofile/status.go`.
   `cmd/curator/main.go` auto-merged cleanly. All other 36 union files clean.
2. Replaced the three conflicted files with the prepared resolutions
   byte-for-byte; SHA-256 verified against `TASK-260918-11f9l1_results.md` §6:
   - `CHANGELOG.md`: `38b42f0d6434b1b3d01d7611ee7217a6ca3beff1e00e125678078ef0e85b1ca9` ✓
   - `cmd/curator/envstatus.go`: `ceeccf459dece03b4be4799e0b453c6df290b65bbc1d19fe09174193afc304cf` ✓
   - `internal/envprofile/status.go`: `d696023d369e3afe37f02203678f2fe782bce5438258bff84265d859a8a62090` ✓
   then `git add` on the three.
3. `git apply --index TASK-260918-11f9l1_fixup.patch` — exit 0.
4. `git reset -q`, then `git add -N` on the 14 union-new files, so the delta
   is an uncommitted working-tree change with new files intent-to-added.
5. `git status --short --untracked-files=all` lists exactly 42 paths =
   the union's 40 + the 2 fix-up test files (§6); rule 8 hygiene holds.

Union patch itself: 336619 bytes,
sha256 `18a5cd686f64c8e4f17ab74240b703db86a4a24e4de910ec110bf375b3cc489d`,
whole-patch `git patch-id --stable` = `ccc9574a4bb4d6191c8f3b1ef829882092489946`
(matches the brief).

## 2. Identity proof

### 2a. Per-file `git patch-id --stable`: union chunk vs `git diff HEAD -- <file>`

37 of 40 IDENTICAL (first 12 hex shown); the 3 DIFFERs are exactly the
resolved files. `cmd/curator/main.go` is patch-id IDENTICAL: the auto-merge
reproduced the union's change-set at a shifted offset (patch-id ignores line
numbers), which is stronger than AC 1 requires (it allows main.go to differ
as an auto-merge; §2c proves the merge both directions anyway).

| File | Union pid | Tree pid | Verdict |
|---|---|---|---|
| .github/workflows/ci.yml | d51ea717e5ef | d51ea717e5ef | IDENTICAL |
| CHANGELOG.md | dde4c9eb24e0 | 24eab169007f | DIFFER (resolved §1) |
| cmd/curator/env.go | 34809ba5f12f | 34809ba5f12f | IDENTICAL |
| cmd/curator/env_test.go | c42d7602ff15 | c42d7602ff15 | IDENTICAL |
| cmd/curator/envconfig_test.go | e12547829c5e | e12547829c5e | IDENTICAL |
| cmd/curator/envstatus.go | de5f4651d440 | e6cc447f9f96 | DIFFER (resolved §1) |
| cmd/curator/main.go | faa95dd84ae2 | faa95dd84ae2 | IDENTICAL (auto-merge, §2c) |
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
| internal/envprofile/status.go | 5d923f007212 | 40e59b1b3f6a | DIFFER (resolved §1) |
| internal/envprofile/surfacing.go | fcd6df0f1790 | fcd6df0f1790 | IDENTICAL |
| internal/envprofile/surfacing_order_test.go | 915a1e39f1f6 | 915a1e39f1f6 | IDENTICAL |
| internal/envprofile/surfacing_test.go | 36113721090b | 36113721090b | IDENTICAL |
| internal/envregistry/envregistry.go | 8e15d91303aa | 8e15d91303aa | IDENTICAL |
| internal/globalbins/globalbins.go | dac49f9b94ed | dac49f9b94ed | IDENTICAL |
| internal/globalbins/globalbins_test.go | 13e9075fdeef | 13e9075fdeef | IDENTICAL |
| internal/interop/environments/context_materialization_test.go | 8ea1153316d8 | 8ea1153316d8 | IDENTICAL |

MATCH=37 DIFFER=3.

### 2b. Byte identity vs checkpoint `73fc8a4`

36 of 40 union files byte-identical (`cmp`) to `73fc8a4:<path>`; the 4
diffs are exactly the three resolved files + `cmd/curator/main.go` (S6
hunks, §2c). Changed set vs trunk == union 40-file set + the 2 fix-up
files, verified by sorted-list `diff` (`CHANGED_SET_EXACT`).

### 2c. `cmd/curator/main.go` auto-merge, both directions

- Direction 1 (tree == checkpoint + S6): `diff 73fc8a4:main.go tree/main.go`
  content-equals `diff 3c45d4b:main.go HEAD:main.go` (47 lines each, equal
  modulo diff line-number headers): the S6 `hook` subcommand, hook-trust
  posture assessment, and `shell_hook_trust[_warnings]` JSON keys only.
- Direction 2 (tree == trunk + union hunk): with matched a/b paths,
  `git diff --no-index trunk tree | git patch-id` ==
  `git diff --no-index 3c45d4b 73fc8a4 | git patch-id` ==
  `f5d9227ae7b880fb02c7ff30a61239540ff7dc34`: the union umbrella hunk
  (`cmdUmbrella(home, args)` → `cmdUmbrella(cfg, home, args)`) only.
- The union's main.go hunk touches only the umbrella dispatch path
  (verified by reading the union chunk); `cmdStatus` (main.go:700) never
  calls the env printer — only `cmdEnvStatus` (env.go:128) does — so
  `curator status` carries the S6 hook rows only (§3).

### 2d. Fix-up identity

`git diff HEAD -- cmd/curator/hook_posture_test.go cmd/curator/hook_test.go
| git patch-id --stable` = `63ddb9ed0bb49aef6278ed0a684dc495f0a0e204`,
identical to the fix-up patch's own patch-id. `+33 lines, 0 deletions`,
test files only (hook_posture_test.go +17, hook_test.go +16); neither file
is in the union's 40. Cause (per 11f9l1 §5): the union attaches §12
provider rows on every `env status` and a missing provider is non-current,
so the S6 `--check`-expects-OK assertions needed stub providers.

## 3. Closed output order of `curator status` / `env status`

Verified against the worktree files (envstatus.go print order, main.go:700
vs env.go:128 call sites), identical to `TASK-260918-11f9l1_results.md` §3:

`curator env status` prints, after the optional `require_current_profile` line:

1. `shell-hook-trust: warning: ...` rows (S6; envstatus.go:26)
2. shell-hook trust rows `shell-hook-trust: <path> ...` (S6; envstatus.go:29)
3. `s4_profile: ..., passable_env_names: ...` (union S4; envstatus.go:34)
4. `warning: ...` machine warnings (union; envstatus.go:36)
5. `scope <s> profile <p> mcp-declarations:` + rows (union §2.3; envstatus.go:39-41)
6. homes + surfaces/findings/warnings/passthrough/seeds/backups (unchanged)
7. scopes, `tool ...` adapters (unchanged)
8. `provider <name>: ...` (union E4; envstatus.go:91)
9. targets, profiles (+ `transitive_system_modules`, dropped modules), unregistered, orphans, notes (unchanged positions)

Order authority: manager profile §10 closed posture inventory (hook-trust is
gate 1 of 13, ahead of env-passthrough / provider-trust-roots / mcp rows).
`curator status` carries the S6 hook rows only — the union added no posture
rows there (§2c).

## 4. SPEC_PIN

`dced9b8` unchanged, no other change: `.github/workflows/ci.yml:53`
`SPEC_PIN: dced9b8317e0e8af79edf2d0539b32bd22b6c85b`, referenced by the four
spec-checkout steps (lines 113/255/377/510). Local conformance root
`/tmp/spec-rc12` verified at exactly
`dced9b8317e0e8af79edf2d0539b32bd22b6c85b` (pre-existing worktree, reused).

## 5. Gates at the rc.12 root (all with `CURATOR_CONFORMANCE_ROOT=/tmp/spec-rc12/conformance/v1`)

Builds and evidence all outside the worktree (rule 8): evidence in
`/tmp/ci-evidence-fjl6v2`, no `go build -o` into the tree.

| Gate | Exit |
|---|---|
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `gofmt -l .` (only pre-existing `.task-board/` resource files listed; all 40 candidate `.go` files clean) | 0 |
| `golangci-lint run` on all 8 touched trees (`cmd/curator`, `internal/{envprofile,config,contextmaterialize,envfragment,envregistry,globalbins,interop/environments}`) — 0 issues | 0 |
| `suite-plan.sh` at rc.12 root: served=75 deferred=0 excluded=0 | 0 |

Full suite = the documented `make ci-test` equivalent, split into bounded
chunks (one shell call is time-bounded; a single package alone exceeds it,
and macOS has no `timeout(1)`/`setsid(1)` for guards). Every chunk ran the
gate's exact flags (`-json -count=1 -timeout 30m`, root exported); only
package/test selection was split. Partitioned packages were split by
`-run ^(A|B|...)$` over the enumerated top-level tests, each test in exactly
one partition (names verified regex-plain); merged stream fed to the
unmodified `platform-case-gate.sh` exactly as `test-gate.sh` invokes it.

| Chunk | Content | Exit |
|---|---|---|
| F1 | 21 pkgs (curator-spec-pin … crossconformance): 19 pass, 2 no-test-file skips, 0 fail | 0 |
| F2 | 21 pkgs (devsub … marker, incl. godriver): 21 pass, 0 fail | 0 |
| F3 | 15 pkgs (mcp … skillcheck, incl. shell/rustsource/pnpmsource): 15 pass, 0 fail | 0 |
| F4 | 15 pkgs (skillspec … yarnmodernsource): 14 pass, 1 no-test-file skip (testtoolchain), 0 fail | 0 |
| E1–E4 | internal/envprofile 178/178 tests (45+45+45+43), 0 fail | 0 ×4 |
| I1–I3 | internal/install 223/223 tests (75+75+73), 0 fail | 0 ×3 |
| C1–C8 | cmd/curator 243/243 tests (7×31+26), 0 fail | 0 ×8 |
| ledger | `platform-case-gate.sh` over merged 7.1MB stream | **1 — pre-existing environmental, see below** |

`go test` overall: every served package passes, zero failures, zero
unexpected passes — the full suite is green locally. Partition coverage was
verified by distinct-test counts per package family (178/178, 223/223,
243/243) plus per-chunk exit 0.

Ledger exit 1 is NOT a regression: 4 cases the ledger requires to RUN on
darwin skip for host-capability absence on this machine —
`internal/pnpmsource` ×2 ("pinned pnpm executable unavailable") and
`internal/rustsource` ×2 ("no operator-approved Cargo descriptor for native
target x86_64-apple-darwin"). Neither package is touched by the candidate
(§2). Reproduced on the bare trunk: a scratch worktree at `1c464c5` (removed
afterwards) skips the same 4 tests with `go test` exit 0. The remaining 59
skips are ledger-tolerated (22 pwsh-absent, 15 cargo-descriptor, yarn/go
integration opt-ins, platform probes, etc.). The hosted gate at handoff,
with full toolchains, is the arbiter for these 4.

## 6. Hygiene (rule 8)

- `git status --short --untracked-files=all`: exactly 42 lines, all
  `M`/`A`, equal (sorted-diff) to the union 40 + the 2 fix-up files.
- No commit on the Story branch (HEAD still `1c464c5`); `git diff HEAD`:
  42 files, +6812/−219.
- No executables, `.review/`, coverage, or evidence files in the tree
  (evidence in `/tmp/ci-evidence-fjl6v2`; results file written to `/tmp`).

## 7. AC mapping

1. Candidate = trunk `1c464c5` + 40-file union (per-file patch-id identical
   except the three byte-verified resolutions; main.go auto-merge proven both
   directions, §2c — and patch-id identical too) + fix-up (+33/−0, patch-id
   identical, §2d). ✓
2. SPEC_PIN `dced9b8` unchanged; changed set == union 40 + fix-up 2, nothing
   else. ✓
3. `go build`/`go vet`/`gofmt`/`golangci-lint` clean; full 75-package suite
   green locally (all chunk exits 0, transcripts above); hosted gate runs at
   handoff (runtime-owned). Ledger caveat documented in §5 (pre-existing,
   trunk-reproduced, untouched packages). ✓
4. Per-file identity proof (§2), closed output order (§3), gate transcripts
   (§5), rule 8 hygiene (§6). ✓

Checklist notes: no NEW tests were authored beyond applying the prepared
fix-up — correct for a landing task (the union carries its own conformance
tests; inventing more around an accepted pin-promotion would be scope
creep); the §5 ledger finding is recorded here and in the task notes rather
than LOGBOOK.md (producer LOGBOOK.md edits are forbidden by the campaign
rules attached to this task).

## 8. Operational note (board CLI)

Mid-run the PATH `task-board` (a shim exec'ing the
`.curator/cache/build/go-v1/e77b15…` binary) stopped starting — new
invocations wedged pre-exec (sampled in `_dyld_start`, 0 CPU) while every
other binary launched instantly, and other agents' invocations stalled too.
All board writes from the attach step on (resource add/update, notes,
checklist, handoff) were executed with the functional identical build
`/Users/administrator/.local/bin/task-board-main-6cb09a23-curatorlike`
(already in use by another runner on this machine), with
`--no-update-check`. Board effects were verified via re-reads.
