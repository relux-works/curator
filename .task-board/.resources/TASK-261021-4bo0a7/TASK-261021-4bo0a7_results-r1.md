# TASK-261021-4bo0a7 results, revision 1

Branch: `research/project-source-exclusion-muse`.
Head: `9b657867a0c65f6e86caf85f6ff6fb7fa4235c0a`.
Base: `cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7` (`origin/main` at task start).
Commit signature: `G` from `git log -1 --format='%G?'`.
Design read: curator-spec `e41f96387287f5e76604627f9e63662f7d6d86d4`, CIP-0002.

## Result

Research harness, fixtures, sanitized native observations and P1-P8 qualification
record are committed. Product code, adapter registration and board status are
unchanged. Complete strict project exclusion is **not qualified**.

A fresh untrusted headless home excludes the measured startup project rules,
skills, UserPromptSubmit hooks and `.mcp.json` while selected home copies load.
Empty managed-home MCP alone is not exclusive in a trusted checkout. Resume
retains the root-rule snapshot, but granting the resume trust prompt revives
project skills, hooks and MCP; the project marker hook executes. Lazy tools,
native child invocation and permission enforcement are explicitly unmeasured.
No inference, real credentials or other worker agents were used.

The binary was copied read-only from the authorized pin into ignored scratch and
both original and copy verified at SHA-256
`6c04ce1719844685153ef6f04b98113cb61b942d97f0a8b0bfb3932c343984d9`.
All native sessions used echo, env -i, disposable HOME/XDG, auto-update disabled
and an outer network/native-store deny sandbox. Synthetic checkouts use external
Git metadata and .git files. Private PTYs and loopback control servers terminate
inside their probes.

## Files changed

- `.research/project-exclusion-muse/README.md`
- `.research/project-exclusion-muse/evidence/control-digest.log`
- `.research/project-exclusion-muse/evidence/control-keychains.log`
- `.research/project-exclusion-muse/evidence/control-network.log`
- `.research/project-exclusion-muse/evidence/control-run-root.log`
- `.research/project-exclusion-muse/evidence/control-trust.log`
- `.research/project-exclusion-muse/evidence/controls.log`
- `.research/project-exclusion-muse/evidence/gated.log`
- `.research/project-exclusion-muse/evidence/observations.txt`
- `.research/project-exclusion-muse/evidence/ungated.log`
- `.research/project-exclusion-muse/evidence/validation.md`
- `.research/project-exclusion-muse/mutants.sh`
- `.research/project-exclusion-muse/replay_test.go`
- `.research/project-exclusion-muse/testdata/AGENTS.md`
- `.research/project-exclusion-muse/testdata/SKILL.md`
- `.research/project-exclusion-muse/testdata/hook.sh`
- `.research/project-exclusion-muse/tui_test.go`
- `docs/qualification/project-source-exclusion-muse.md`

`replay_test.go` covers startup discovery and sandbox controls. `tui_test.go`
covers actual resume and late-file listings. `mutants.sh` checks intended red
verdicts and setup refusals. Fixtures contain only synthetic markers. Evidence
logs contain sanitized native counts, terminal observations and actual test
results. The qualification record maps each observation to its limit.

## Tests and claims

| Test | Claim established by executable assertions |
| --- | --- |
| TestChecker | Clean zero and positive one records pass; missing, failed, malformed, duplicated and contradictory event/count records refuse. |
| TestP1RootAndAncestors | Four instruction filenames load individually at checkout root or an inside-checkout ancestor when trusted, and are absent from untrusted startup while the selected home root loads. |
| TestP1OutsideGitBoundary | The planted source above the Git root is not discovered in either trust mode; the neighboring in-checkout positive loads. |
| TestP2NestedStartup | A nested fresh launch sees two trusted project sources or zero untrusted, alongside the home root; this is not a lazy-read proof. |
| TestP3Skills | Each of three project skill roots is cataloged when trusted and excluded when untrusted, retaining a distinct home skill. |
| TestP4Hooks | Trusted project and selected user hooks execute; under untrusted exec only the user marker is written and the project handler is inactive. |
| TestP4Plugins | A raw repository bundle is inert; an explicit untrusted project install is rejected, while trusted project and selected home installs admit the same command capability. |
| TestP5MCP | Empty/non-empty project MCP config participates only when trusted, home config remains admitted, and empty home MCP does not exclude a trusted non-empty project source. Configuration enumeration only, no successful handshake claim. |
| TestP6SettingsCandidates | Three planted candidate settings files do not have the user parser's poison-byte startup effect; valid home settings supply the model source, while identical poison user bytes refuse. Tool permission decisions remain unmeasured. |
| TestP7ResumeTrustGate | A real resume offers trust or quit; granting it preserves the saved root-rule listing but revives project skill, hook and MCP catalogs, with project-hook execution. Home sources remain present. |
| TestP7ChildAvailability | Untrusted exec reports delegation unavailable and excludes the project root; real child execution/inheritance remains unmeasured. |
| TestP8LateRoot | Two distinct native rule listings omit a root created between turns, retain the home root, and a fresh trusted process loads the late file. No lazy-tool or late-other-surface claim. |
| TestOfflineBoundary | An ordinary fixture file and unsandboxed loopback connection succeed; the sandbox rejects that connection without another hit and denies an unchanged synthetic Keychains file. |

The full gated matrix has 54 passing top-level/subtest records. Final checker
and distinct-late-listing assertions also pass in focused reruns. All 13 research
tests skip explicitly without the binary gate. A skip is not qualification.

## Commands and exits

All Go test commands use `env -u GIT_DIR -u GIT_WORK_TREE go test`. Go 1.25.5
and workspace-local caches are used; no external tool installation was needed.
Native run roots are fresh task-local `.temp/` paths and are omitted from public
argv to avoid publishing personal locators.

| Command/run | Exit | Evidence |
| --- | ---: | --- |
| a2a inbox | 0 | Newest authorized request in the assigned room/topic selected. Other recipients and senders ignored. |
| task-board q get(TASK-261021-4bo0a7) | 0 | Assignee and development status confirmed; only the listed brief is a precondition resource. |
| task-board resource get TASK-261021-4bo0a7 W-ADP-muse-brief.md | 0 | Brief materialized and read before work. |
| a2a post ack TASK-261021-4bo0a7 | 0 | Ack in the request topic to the orchestrator. |
| git fetch origin; branch creation from origin/main | 0 each | Base above; no stale feature branch used. |
| go test ./.research/project-exclusion-muse -count=1 -v, ungated | 0 | Explicit 13-test skip transcript in committed evidence. |
| go test ./.research/project-exclusion-muse -count=1 -v -timeout 8m, gated-04 | 0 | Full matrix, committed gated.log. |
| go test ./.research/project-exclusion-muse -count=1 -v -run '^TestChecker$', final gate | 0 | Contradictory-event hardening validated. |
| go test ./.research/project-exclusion-muse -count=1 -v -run '^TestP8LateRoot$', final gate | 0 | Requires two distinct native before/after listings. |
| sh .research/project-exclusion-muse/mutants.sh | 0 | All five controls detected intended failures; each inner test command exits 1 with its required invariant. |
| sh -n .research/project-exclusion-muse/mutants.sh | 0 | Syntax check. |
| go vet ./.research/project-exclusion-muse ./... | 0 | Default and research packages. |
| go vet ./.research/project-exclusion-muse, final source | 0 | Final research source after assertion hardening. |
| go build ./... | 0 | Product module builds. |
| public artifact/English/local-link checks | 0 | Bounded scan over all 18 delivered files; no personal locators. |
| git diff --cached --check | 0 | Reviewed task-only scope. |
| git commit -S | 0 | Signed head above. |
| git log -1 --format='%G?' | 0 | G. |
| disposable normal Git clone and exact-head checkout | 0 each | Validation source equals the handover SHA. |

Exploratory gated-01/02/03 and focus-01 exited 1 for harness assumptions, not
product regressions. The committed validation record preserves each cause:
wrong pre-catalog extraction boundary, a source above the Git boundary mistaken
for an in-checkout ancestor, retained resume rule snapshot, and curl progress
text in a body-only assertion. Corrected full and focused runs pass. Initial
parser probes also establish that `context.project_rules=false` is refused as an
unknown setting, exit 2; no fictional native project-off switch is claimed.

## Default suite at the signed head

The default inventory contains 85 packages at the signed head. All 80 other packages completed their full package runs (including native platform/tool skips); five large packages were partitioned. Every listed test in those packages completed with pass or an explicit skip. The detailed command argv, inventories, case lists and exit codes are in companion outcome `TASK-261021-4bo0a7_validation-r1.json`.

| Package | Completed inventory |
| --- | ---: |
| `cmd/curator` | 434/434 |
| `internal/envprofile` | 393/393 |
| `internal/install` | 371/371 |
| `internal/artifactpolicy` | 60/60 |
| `internal/crossconformance` | 54/54 |

The activation matrix is additionally checked against all 36 expected leaf cases, and the network-contact matrix against all 40 expected vector parents. Filtered reruns do not establish completion merely by emitting a passing parent test. All expected cases are present in the persisted pass/skip set.

All default tests used a normal disposable Git clone detached at the exact handover SHA, a short TMPDIR and a task-local Go build cache. Corrected invocations also set GOTMPDIR to that short root and isolated PATH. Initial environment exceptions are preserved in the companion; successful package results from those attempts remain part of the inventory. No research binary opt-in was set. A separate exact-head ungated research command exited 0 with all 13 tests skipped. These local results do not substitute for hosted CI, conformance-root admission, another OS, or native qualification of a skipped case.

| Recorded command group | Exit | Claim or interpretation |
| --- | ---: | --- |
| `default-artifactpolicy-flood` | 1 | Unsuccessful environment or time-limited attempt; resolved by recorded reruns and inventory coverage above. |
| `default-artifactpolicy-flood-long` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-artifactpolicy-rest` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-crossconformance-s0` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-crossconformance-s1` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-crossconformance-s2` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-curator-0` | 1 | Unsuccessful environment or time-limited attempt; resolved by recorded reruns and inventory coverage above. |
| `default-curator-last` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-curator-matrix-tail` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-curator-remaining-1` | 1 | Unsuccessful environment or time-limited attempt; resolved by recorded reruns and inventory coverage above. |
| `default-curator-remaining-2` | 1 | Unsuccessful environment or time-limited attempt; resolved by recorded reruns and inventory coverage above. |
| `default-curator-s0` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-curator-s1` | 1 | Unsuccessful environment or time-limited attempt; resolved by recorded reruns and inventory coverage above. |
| `default-curator-s1-rest` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-curator-s2` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-curator-tail` | 1 | Unsuccessful environment or time-limited attempt; resolved by recorded reruns and inventory coverage above. |
| `default-environment-witness` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-envprofile-0` | 1 | Unsuccessful environment or time-limited attempt; resolved by recorded reruns and inventory coverage above. |
| `default-envprofile-final` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-envprofile-network-0` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-envprofile-network-1` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-envprofile-s0` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-envprofile-s1` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-envprofile-s2` | 1 | Unsuccessful environment or time-limited attempt; resolved by recorded reruns and inventory coverage above. |
| `default-envprofile-s2-rest` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-envprofile-s3` | 1 | Unsuccessful environment or time-limited attempt; resolved by recorded reruns and inventory coverage above. |
| `default-envprofile-s3-rest` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-envprofile-s4` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-envprofile-s5` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-envprofile-s6` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-envprofile-s7` | 1 | Unsuccessful environment or time-limited attempt; resolved by recorded reruns and inventory coverage above. |
| `default-install-s0` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-install-s1` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-install-s2` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-install-s3` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-install-s4` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-install-s5` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-install-s6` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-install-s7` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-list` | 0 | Exact-head inventory discovery. |
| `default-rest-0` | 1 | Unsuccessful environment or time-limited attempt; resolved by recorded reruns and inventory coverage above. |
| `default-rest-1` | 1 | Unsuccessful environment or time-limited attempt; resolved by recorded reruns and inventory coverage above. |
| `default-rest-2` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-rest-3` | 1 | Unsuccessful environment or time-limited attempt; resolved by recorded reruns and inventory coverage above. |
| `default-rest-3-retry` | 0 | Selected default contracts pass, with explicit skips retained. |
| `default-rest-4` | 0 | Selected default contracts pass, with explicit skips retained. |
| `exact-head-ungated` | 0 | Research gate skips all 13 tests before fixture/native startup. |
| `flood-cleanup` | 0 | Two abandoned, uniquely identified synthetic fixture roots removed. |
| `list-artifactpolicy` | 0 | Exact-head inventory discovery. |
| `list-crossconformance` | 0 | Exact-head inventory discovery. |
| `list-curator` | 0 | Exact-head inventory discovery. |
| `list-envprofile` | 0 | Exact-head inventory discovery. |
| `list-install` | 0 | Exact-head inventory discovery. |

Initial trials were not counted as passing commands. Long GOTMPDIR caused a legacy golden path mismatch; ambient provider executables in PATH caused two CLI status failures. The corrected short-root/status witness passes. Restricting PATH initially hid rustc from two Rust tests; direct links to the installed rustc/cargo restored the toolchain and the whole package group passes. Several large groups exceeded their 8-minute test alarm, and the single 100,001-file directory test exceeded 9 minutes during fixture deletion. Remaining tests/cases were partitioned explicitly. The same unmodified directory test passes with a 25-minute alarm (about 18 minutes including cleanup). A monitored process allowed short shell interactions and finished before handover.

The first cleanup guard refused the macOS /tmp symlink before deletion; canonical-parent validation corrected that guard. Start-time matching and ended-process receipts identified only this task's two abandoned empty-file fixtures. Their cleanup completed successfully. The validation clone was verified unchanged at the signed head and deleted before attachment. No task process or private PTY is left running.

## Open decisions and limits

- Keep strict Muse project capability unavailable from this evidence. A fresh
  untrusted exec recipe is startup-scoped and relies on an empty managed trust
  store; it does not override a reused trusted store.
- Native untrusted interactive/resume entry needs a separately qualified
  contract. Trusting the checkout on resume defeats live catalog exclusion.
- Add actual lazy read/entry and native child tools before claiming their
  inheritance. Echo cannot supply those paths, so no model assertion is used
  as a substitute.
- Permission-policy decisions, other hook events/plugin capability families,
  successful MCP handshakes and late non-rule sources remain unmeasured.
- Selected synthetic home copies demonstrate native delivery, not production
  manager approval or protection against an unrestricted same-UID attacker.
- No hosted CI, PR, merge, registry bump or other platform qualification was
  attempted. The owning orchestrator retains board status and integration.

## Publication

`git push -u origin research/project-source-exclusion-muse` exited 0. Remote head equals the signed SHA above, signature verification is G, and the source checkout is clean. `git diff --quiet <base> HEAD -- cmd internal tools go.mod go.sum` exits 0: default product/test sources are unchanged. Results and the command/inventory companion are attached as outcome resources; the orchestrator owns integration and board status.
