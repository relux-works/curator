# TASK-261021-i2f2uq results, revision 1

Outcome: proposal delivered for CLI grammar and UX across Curator, curator-run, Werkstatt, and the task-board registration seam. Board assignment was verified and the request acknowledged before review. Board status was not changed.

## Deliverables

- `TASK-261021-i2f2uq_cli-review.md`: exhaustive 72-topic Curator inventory with parser corrections, launcher/station/registry flags and exit families, registration seam and its exit-evidence limit, 18 source-backed findings, now-to-proposed grammar grouped by the seven requested nouns, compatibility windows, hook-free registration and conditional trusted lazy registration, five prioritized cross-repository slices, and open decisions.
- `TASK-261021-i2f2uq_help-evidence.md`: 92 help/version capture sections and seven bounded probes with asserted results. Absolute fixture locations are replaced by neutral placeholders.
- This results file.

Repository files changed: none. Both inspected worktrees were clean before and after review. All reports are outside the repositories; task-local capture scripts, fixture policy, build output and private logs are temporary artifacts. No PR, branch, commit, or push was created because the task brief explicitly makes both repositories read-only and requests only a proposal. Signing and publication of a repository change are therefore inapplicable.

## Reviewed heads

| Repository | Read-only head |
| --- | --- |
| curator | `cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7` |
| curator-agent-launcher | `dd70beb42239fc80f81bcfd512358f8d3844b615` |

Curator HEAD equaled fetched origin/main at review. The launcher was inspected and built at its recorded local HEAD; no claim is made that it equals remote main. Curator executable: v0.15.0-rc.7. Launcher review build: 0.3.0, reported specification 0.5.1-draft. Go toolchain: 1.25.5, darwin/arm64. Python: 3.14.7.

## Commands and verification

Commands are given relative to the repository or task-output directory. Workspace-specific absolute paths in the original invocation are expressed through `$PWD` below; no private paths or identities are published. `$CURATOR_GO` was the preconfigured toolchain. Initial readiness used command lookup and help/version, with output retained under `.temp/bootstrap/` and `.temp/TASK-261021-i2f2uq/`.

| Command / check | Exit | Claim and limits |
| --- | ---: | --- |
| `a2a inbox` | 0 | Received the addressed request in the authorized room from the authorized orchestrator; no other request was acted on. |
| `task-board q 'get(TASK-261021-i2f2uq) { status assignee }'` | 0 | Task existed and matched the assigned developer identity before work. |
| `task-board resource get TASK-261021-i2f2uq W-CLI-brief.md` | 0 | Materialized and read the only precondition resource; later narrow task projection confirmed the complete precondition list. |
| `task-board q 'get(TASK-261021-i2f2uq) { description scope ac preconditionResources outcomeResources }'` | 0 | Confirmed proposal-only scope and all five acceptance requirements. |
| `a2a post ... --kind ack ... 'ack TASK-261021-i2f2uq'` | 0 | Acknowledgment published to the request topic. |
| `git fetch origin` with the required agent Git environment loaded | 0 | Refreshed remote references; subsequent HEAD/origin-main comparison matched. |
| From launcher: `env -u GIT_DIR -u GIT_WORK_TREE GOCACHE="$PWD/../.tmp/go-cache" GOPATH="$PWD/../.gopath" GOTMPDIR="$PWD/../.tmp" "$CURATOR_GO" build -o "$PWD/../.tmp/curator-run-review" ./cmd/curator-run` | 0 | Built the actual launcher entry point at the recorded source head. This is build/help evidence, not hosted runtime qualification. Dependencies downloaded into workspace-local caches; no additional CLI tool was installed. |
| `python3 .temp/TASK-261021-i2f2uq/capture.py` | 0 | Captured all 72 registered Curator topic helps, each exit 0, plus root/version/launcher/task-board help. Administrative launcher help returned expected observed refusal exits 2, rather than being counted as successful help. |
| `python3 .temp/TASK-261021-i2f2uq/probes.py` | 0 | Seven exact exit/text assertions passed; claims appear below and beside every probe in the evidence. |
| `python3 .temp/TASK-261021-i2f2uq/validate.py`, first pass | 1 | The Markdown table-width check caught an unescaped pipe in a proposed synopsis. This was an artifact-format failure, corrected before attachment. |
| Same validation, final pass | 0 | All 72 inventory topics present; source-reference targets exist; five requested sections present; table widths consistent; English/public-path checks passed. A removed-topic control is rejected and a planted private-path pattern is detected. This is a bounded document validator, not a general secret detector. |
| `git status --short` for each explicitly selected repository | 0 | No tracked or untracked repository changes were reported; task scratch is ignored. |

An exploratory query using an unsupported `resources` projection was rejected by the query parser; `schema()` identified `preconditionResources`, and the corrected query above succeeded. Exploratory missing-path searches returned no relevant local skills. No workflow depended on those failed lookups.

### Bounded behavior probes

1. `curator run session-manager register --help`, exit 0: proves the umbrella returns the generic run topic, losing nested provider help.
2. `curator-run session-manager register --help`, exit 2: proves direct nested registry help is rejected before registration.
3. `curator-run werkstatt status example --help`, exit 2: proves adding the complete station operand does not make nested station help work.
4. `curator-run --help` with absent fixture user policy, exit 0: positive control for top-level informational output.
5. `curator-run codex --native --hosted`, exit 2: proves conflicting host flags are refused by syntax admission; no agent launch is attempted.
6. `curator-run werkstatt down example --json`, exit 2: proves a status/list-only flag is rejected on down; no station mutation occurs.
7. `curator-run --help` after malformed fixture `launcher.json`, exit 1 with `defaults_config_invalid`: proves policy parsing can block top-level help. Only the task-local fixture was changed.

No product behavior was changed and no new product gate was introduced. The full Go suites, race tests, mutation suites, real session-manager registration, and hosted/device tests were not run: they do not validate the delta of a read-only proposal and would add unrelated effects. Required future positive/negative controls and narrowing mutants for the proposed gates are specified in proposal section 4.3. No .git directory or disposable validation clone was created.

## Key conclusions

- Prioritize consistent static nested help and truthful documentation before broad command renames.
- Explicit `task-board session-manager register --for curator-run --select` is the supported grammar for hook-free setup; choosing no-select preserves the existing selection.
- Lazy registration is conditional on trusted launcher-owned discovery. A PATH search, repeated re-registration, or silently selecting a peer would violate existing trust, selection or station-owner contracts.
- Move explicit station metadata edits and owner recovery out of advertised status grammar, retaining compatible aliases and current unknown/stale/generation safeguards.

## Open decisions and evidence limits

- Approve or decline the trusted catalog and lazy first-hosted-use design; decide separately whether auto-empty routing should change.
- Approve any optional grammar aliases and the proposed two-minor-release/90-day minimum compatibility window.
- Publish or source-verify task-board's registration-wrapper failure exit mapping; help only establishes the surface and help exit 0.
- Reconcile draft implementation-status banners and confirm upstream specification acceptance before implementation.
- Curator-spec and skill-project-management were named as affected owners through the inspected launcher references; their current sources were not independently reviewed.

These decisions are documented for subsequent tasks and do not leave this requested proposal unfinished. Branch: not applicable, read-only proposal. Handover head: curator `cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7`; launcher `dd70beb42239fc80f81bcfd512358f8d3844b615`.

## Publication evidence

Both proposal and help-evidence resource-add commands exited 0. Separate resource-get commands exited 0 and byte-for-byte comparisons matched the local deliverables. The results resource is published last.
