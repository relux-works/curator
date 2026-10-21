# TASK-261021-boww1n results, revision 1

Branch: `research/project-source-exclusion-codex`

Head: `7f97b0939ddb604ad7bbcccc5871e28cac6fe92f`

Base: `origin/main` at `cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7`.

Commit signature: `git log -1 --format='%G?'` returns `G`;
`git verify-commit HEAD` exits 0. Only the named feature branch is published.
No PR is opened and no board status is changed.

## Result

A complete strict-context adapter is not qualified. Project instruction
startup injection is excluded by `project_doc_max_bytes=0` while the approved
managed-home root remains. Explicit project distrust excludes measured
config-driven hooks, plugin enable overrides, MCP and permission/settings
values. Repository `.agents/skills` still loads. An exact-path skill-disable
list misses a newly created skill, including through explicit reload in a
running app-server. A future adapter must refuse unqualified strict entry.

P1-P8 each have a scoped verdict and an approved-copy control in
`docs/qualification/project-source-exclusion-codex.md`. Lazy entry, full resume
exclusion, child agents, tainted history, execution-policy rules and automatic
in-turn discovery remain unmeasured; the record does not promote them.
The resume source enumeration fails the uncapped project-positive baseline
and cannot prove project exclusion. No other agent or model turn is started.

## Files changed

- `.research/project-exclusion-codex/replay_test.go`: digest-gated offline probes.
- `.research/project-exclusion-codex/testdata/mcp.py`: synthetic stdio MCP fixture.
- `.research/project-exclusion-codex/mutants.sh`: four discriminating expected-red controls.
- `.research/project-exclusion-codex/observations.json`: bounded sanitized observations and source hashes.
- `.research/project-exclusion-codex/README.md`: replay instructions, tools and boundaries.
- `docs/qualification/project-source-exclusion-codex.md`: per-row verdicts and residuals.
- `README.md`: research tool entry point.

No product code changes. All inputs are synthetic. No real credentials,
auth commands, remote inference, Keychain access, native home changes or
network-enabled vendor launch is used.

## Added tests and the claims they prove

| Test | Claim and material limit |
| --- | --- |
| `TestInstructions` | A complete valid prompt dump loads approved home and positive project/nested/override markers; cap zero removes the measured project injection and preserves the home marker. Lazy in-turn entry is outside the claim. |
| `TestSkills` | Distrust plus document cap retains a project `.agents/skills` marker; exact-path disables remove known paths but a new path loads. Approved home skill remains. No unsupported `.codex/skills` discovery claim is made. |
| `TestOfflineBoundary` | An ordinary synthetic file and unsandboxed loopback request succeed; sandboxed Keychains-shaped read and same-endpoint connect are specifically denied, with original synthetic file bytes preserved. No real credential file is opened. |
| `TestConfigLayers` | Trusted project config affects hooks/plugins/MCP/settings; distrust excludes project values while home sources remain. Hook enumeration preserves both trusted sources even when effective config array merging differs. Real thread permission policy follows the selected config. Hook execution is outside the claim. |
| `TestLateDiscovery` | One running app-server sees a late project skill on force reload while keeping the approved skill; distrust excludes a late MCP entry and retains the approved entry. No automatic in-turn reload claim is made. |
| `TestResumeApprovedRoot` | Actual start/resume with synthetic durable history retains the approved managed-home instruction source at both document budgets. The instrument omits project sources in the uncapped baseline too, so no project-exclusion or tainted-history claim follows. |
| `TestMCPRuntime` | Four trust/empty-set combinations initialize and call the approved stdio tool. Only trusted/non-empty starts the project process and permits its tool call; negative cases have no project startup receipt. No network MCP or auth claim is made. |
| `TestIntakeBoundary` | Missing opt-in skips, wrong binary bytes fail specifically at the digest check and invalid run root fails specifically before creating a child fixture. The outer qualified run is the positive control. |
| `TestPluginSkills` | Installed approved local plugin skills load with either trust value; trusted project enablement overrides the home deny for the installed project plugin, while distrust preserves that deny. Other plugin entry points are outside the claim. |

## Validation

The validation checkout is a disposable normal `.git` clone at the handed-over
head. Go children have both `GIT_DIR` and `GIT_WORK_TREE` removed. HOME, explicit XDG roots (or defaults under the disposable HOME),
module and build caches are disposable/task-local; `TMPDIR=/tmp/cw-dev3` and
`GOTMPDIR=/tmp/cw-dev3`. Go is 1.25.5 on darwin/arm64.

The explicit research invocation at the exact committed head passed (exit 0,
`head-gated-research.log`). The default explicit research invocation passed
with all vendor entries skipped before fixture creation (exit 0,
`head-default-research.log`). The source build and vet, and their clean-clone
repeats, passed (exit 0). No model turn was needed.

The final narrowing runner exited 0. Each intended child exited 1 and matched
its specific diagnostic: document cap removal, known-path skill-disable
removal, project trust promotion and loopback admission. All four are KILLED
in `mutants-final-03/results.tsv`; earlier invalid plants are not counted.
The three executable input hashes in the committed observations were rechecked
against the handed-over source and match.

Final exact-head commands and exits:

| Effective command / selection | Exit | Claim |
| --- | ---: | --- |
| `$GO build -p=2 ./...` | 0 | All default packages build. |
| `$GO vet -p=2 ./... ./.research/project-exclusion-codex` | 0 | Default packages and explicit research package pass vet. |
| `$GO test -count=1 -v ./.research/project-exclusion-codex`, opt-in absent | 0 | All vendor entries skip before fixture creation. |
| Same explicit test command with pinned `CODEX_EXCL_BINARY` and fresh run root | 0 | Every gated assertion passes at the handed-over head. |
| `sh .research/project-exclusion-codex/mutants.sh` | 0 | All four intended narrowing failures are detected; each child exits 1. |
| `$GO test -p=2 -count=1 -timeout=8m <packages-03.txt entries>` | 0 | 22 default packages pass. |
| Same package-group command with `packages-04.txt` | 1 | 21 packages pass; install exceeds the budget and is replaced below. |
| Same package-group command with `packages-05.txt` | 1 | 21 packages pass; rustsource lacks PATH rustc and is replaced below. |
| Same package-group command with `packages-06.txt` | 0 | 17 default packages pass. |
| `$GO test -v -p=2 -count=1 -timeout=8m ./internal/rustsource`, matching Rust bin in PATH | 0 | Full rustsource retry passes. |
| CLI regex selections below, 14 separate commands | 0 each | All 434 discovered CLI top-level entries selected exactly once. |
| envprofile regex selections below, 12 separate commands | 0 each | All 393 discovered top-level entries selected exactly once. |
| install regex selections 01–11, 11 separate commands | 0 each | 370 discovered entries selected exactly once. |
| `$GO test -v -p=2 -count=1 -timeout=8m -run '^TestAuthoritativeDryRunCasesMutateNothingPersistent$' ./internal/install` | 0 | The remaining install entry passes with uncontended GOROOT lock. |
| `$GO test -count=1 -run '^TestDraftTransportLegacyGolden$' ./cmd/curator` | 0 | Corrected compiler temp-root normalization matches the existing golden. |
| Selection/package/clean-checkout audit | 0 | Exact manifest union covers all 85 default packages; clone is normal Git, clean and at the handed-over head. |

Each Go child removes both Git worktree variables, equivalently prefixed by
`env -u GIT_DIR -u GIT_WORK_TREE`. Regex selections use
`$GO test [-v] -p=2 -count=1 -timeout=8m -run REGEX PACKAGE`;
`REGEX` is the anchored alternation of names in the matching selection file.
Full literal argv is preserved in each adjacent `head-<mode>.log`, and every
listed command has its own `.exit`. Successful split invocations compose the
complete default coverage; no uninterrupted whole-package pass is claimed
for the packages that exceeded the per-call budget.

- `cmd`: `01`, `02`, `03`, `04a`, `04b`, `05`, `06`, `07`, `08`, `09a`, `09b`, `10`, `11a`, `11b`.
- `env`: `01`, `02`, `03`, `04`, `05`, `06a`, `06b1`, `06b2`, `07`, `08`, `09`, `10`.
- `install`: `01`, `02`, `03`, `04`, `05`, `06`, `07`, `08`, `09`, `10`, `11`.

The discovery manifests contain 434 CLI, 393 envprofile and 371 install
entries; the install single-test selection complements its 370 grouped entries.
The audit checks the full name multisets, not only counts, and uses only
successful per-package results from the original groups plus their replacements.
Ordinary upstream optional/platform skips retain their normal meaning.


Earlier failed attempts are retained rather than reported as passing:

- Checkout-readiness attempts ran before the first clone checkout finished and
  found no `go.mod`. The runner now requires the real index and module first.
- Initial CLI and envprofile whole-package runs reached the 8-minute test
  budget. The install package did too, including a later run without its one
  GOROOT-lock test. Final evidence selects all tests in smaller groups.
- A first CLI shard exited 1 specifically while waiting five minutes for the
  package-wide host-GOROOT lock. A second concurrent shard later hit Go's
  9-minute watchdog after waiting and executing most tests. Subsequent CLI
  launches are serialized; the latter shard is replaced by two disjoint parts.
- CLI selection 09 reached the eight-minute package budget during the
  activation matrix, without an assertion failure. Its final disjoint
  replacements are 09a and 09b.
- The original CLI 11 invocation combined an eight-minute recovery-test
  timeout with an unrelated provider-count assertion: inherited PATH exposed
  installed `curator-engines*` programs. The final 11a/11b selections isolate
  the recovery test and use system tools plus the pinned Go bin directory,
  excluding installed optional providers.
- The envprofile 06 and 06b selections reached eight minutes during the
  network-contact matrix. The final disjoint selections are 06a, 06b1 and 06b2;
  the matrix runs alone in 06b2.
- The first CLI golden mismatch used a compiler temporary directory outside
  TMPDIR, so its executable path was not normalized. The corrected short
  GOTMPDIR made the unchanged golden control pass (exit 0).
- The first envprofile 10 run exited 1 because a globally supplied
  `XDG_DATA_HOME` overrode the synthetic operator HOME in the chezmoi positive
  control. A full unchanged group retry with XDG defaults under disposable
  HOME passed (exit 0), including the adjacent symlink negative control.
- The rustsource group initially exited 1 because `rustc` was absent from PATH.
  The already-installed matching Rust/Cargo 1.91.0 toolchain was checked and
  added for a full package retry, which passed (exit 0). No install was needed.

Raw validation artifacts are under `.temp/TASK-261021-boww1n/`. Each exact-head
log begins with the head and full argv and has an adjacent `.exit`. The package
manifest is `packages.txt`, the complete split selections are the `*-tests-*.txt`
files, and their discovery lists and union are retained in `shard-manifest.json`.
Raw prompts and runtime databases are not attached to the public board.

Final teardown and publication verification exited 0: the disposable normal
Git clone and its HOME are deleted, the owned-fixture process scan is empty,
and the owning checkout is clean. `git verify-commit HEAD` exits 0 with `G`,
and `git ls-remote origin refs/heads/research/project-source-exclusion-codex`
exits 0 and agrees with the handed-over head. `final-audit.json` and
`delivery-verification.json` retain these readbacks.

## Open decisions

No product implementation is proposed. Qualification residuals are explicit
follow-up evidence requirements, not silently accepted behavior. The
orchestrator must not interpret a green research run as strict-context
certification or as qualification of the unmeasured lifecycle paths.
