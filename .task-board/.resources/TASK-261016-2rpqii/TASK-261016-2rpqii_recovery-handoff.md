# TASK-261016-2rpqii: agent login reference, recovery evidence

Ready for review. Candidate retained uncommitted in the managed Story worktree.
No branch switch, rebase, merge, commit, or repository push was performed by this successor.

## Candidate and contract

The candidate contains `docs/agent-logins.md`, README visibility and tools commands,
the environment configuration link and session-visibility correction, and four
focused PlantUML sources with embedded SVG outputs. Every capability is marked
works today, rc.7, or planned. Tool observations remain bound to Claude Code
2.1.299, Codex CLI 0.153.2, and Muse 1.4.8. No real login was rerun or read.

All eleven files are byte-identical to the predecessor candidate at hosted
snapshot `cd4487462837154c94ae958aca4bea557836d0e1`. Product source, tests,
configuration, dependencies, and workflow inputs are unchanged. The historical
snapshot is evidence only, not a signed delivery commit created by this run.

Fetched `origin/main` is `0d00a109ad1da77a5506747c89f24af9442eb341`.
The worktree checkpoint remains `205272e75ea7d40fd10e073c20ce49bbc0787cd2`,
two board-only commits behind local and remote main. Direct comparison confirms
that code, docs, scripts, workflows, and module inputs on main equal this
checkpoint. The document's availability baseline therefore still matches main.
Current normative login-readiness and CIP-0017 content was fetched through the
GitHub API and matches the predecessor's saved source exactly.

## Checks executed by this successor

Tools: Go 1.25.5 on macOS arm64, Lychee 0.24.2,
PlantUML 1.2026.6 with Graphviz 14.0.4.
Commands ran directly, with output redirection, without a pipeline masking exits.
Raw logs remain under `.temp/TASK-261016-2rpqii/` and are excluded from public
attachments because raw diagnostics can contain machine paths.

| Command | Actual exit | Evidence and bound |
|---|---:|---|
| `lychee --offline --include-fragments --no-progress README.md docs/environment-config.md docs/agent-logins.md` | 0 | `links-local-04.log`: 31 successful unique local targets, 0 errors, 17 external exclusions. |
| `lychee --no-progress --timeout 30 docs/agent-logins.md` | 0 | `links-external-04.log`: 12/12 unique targets, 0 errors, one redirect. |
| `plantuml -tsvg -nometadata diagrams/plantuml/agent-login-*.puml` | 0 | `plantuml-render-04.log`: 4/4 sources rendered; all four SVG XML documents separately parsed successfully. Output bytes match the original candidate. |
| `plantuml -tpng -o ../../.temp/TASK-261016-2rpqii/preview-02 diagrams/plantuml/agent-login-*.puml` | 0 | `plantuml-preview-02.log`; four previews rendered and visually inspected for clipping and readability. |
| `git diff --check` | 0 | `diff-check-04.log` and pre-handoff `diff-check-05.log`; tracked delta only. No untracked source was changed in this recovery. |
| `bash .github/ci/naming-gate.sh .temp/TASK-261016-2rpqii/public-scope-02` | 0 | `naming-gate-04.log`: byte-verified current copies of all 11 candidate files. This gate checks its declared naming patterns; added prose and SVGs were also inspected for identifying details. |
| `lychee --offline --no-progress .temp/TASK-261016-2rpqii/link-control.md` (valid fixture) | 0 | `links-control-positive-02.log`: 4/4 diagram links resolve. |
| `lychee --offline --no-progress .temp/TASK-261016-2rpqii/link-control.md` (one retargeted diagram) | 2 | Expected red; `links-control-negative-02.log`: one missing target with three valid controls. |
| `go test -count=1 -timeout=3m ./cmd/curator -run '^TestReviewProbeLegacyLaneIdentityInvalid$'` | 0 | `tests-hosted-regression-01.log`: the reviewer regression that failed in the original hosted attempt passes locally. This is not Linux qualification. |
| `gh run rerun 38823186484 --failed` | 0 | `hosted-rerun-01.log`: GitHub accepted the request to rerun only the failed job on the existing exact snapshot; request success alone is not test success. |

## Existing evidence reused

No product code, test, configuration, or environment identity changed.
The predecessor's exit-0 focused suites and build are accepted without replay:

- `go build -o .temp/TASK-261016-2rpqii/curator ./cmd/curator` (`build-01.log`).
- `go test -count=1 -timeout=3m ./internal/envprofile -run '^Test(CodexIsolatedAdmission|CodexUnknownStoreSharedRefuses|CodexCredentialStoreTOMLSpellings|CodexMalformedStoreSharedRefuses|MuseDirectoriesAndIsolation|Reviewer.*)$'` (`tests-envprofile-01.log`).
- `go test -count=1 -timeout=3m ./cmd/curator -run '^TestEnvConfig(ShowDefaults|SetUnsetRoundTrip|Refusals)$'` (`tests-cli-01.log`).
- Built CLI env/profile help and source dispatch/config inspection, confirming
  env login, profile prepare, and credential_store remain unimplemented.

All committed reviewer tests retain their original names. This run adds no
product tests: the delta changes operator documentation and rendered assets,
and documentation link/render checks exercise those artifacts directly.

## Narrowing control

| Mutant | What it narrows the gate to | Named failing test/check | Surviving bound |
|---|---|---|---|
| `LinkCheckMissingDiagram` | Retarget only stores.svg to a missing SVG, retaining three accessible neighboring diagrams. | `lychee --offline --no-progress .temp/TASK-261016-2rpqii/link-control.md`, exit 2, one missing target. | Killed; this establishes missing-target detection only. Diagram meaning, external page correctness, and arbitrary Markdown syntax are not established by this fixture. |

No product mutation claim is made.

## Handoff bounds

The brief supplies no attack-surface table; this is a declared brief gap.
No catalog coverage ratio or product security coverage is inferred from its
absence. No brief row is treated as out of contract. Muse Linux measurement and
refresh behavior remain explicitly unqualified, as required by the supplied
facts. The candidate stays within the requested documentation and diagram
module, with no implementation or workflow edits.

## Task-scoped logbook

- Retained the predecessor's candidate and evidence rather than rewriting it.
- The original configured gate exited 1: only the Linux test job failed.
  Uploaded test events identify `TestReviewProbeLegacyLaneIdentityInvalid`:
  the test expected the frozen source-unavailable diagnostic, but the Go probe
  hit its process timeout. This is a recorded failure, not passing evidence.
  The local diagnostic rerun passed; it does not establish why Linux timed out.
- The original naming scratch scope contained stale versions of two diagram
  labels. A byte-identity check detected the mismatch. Recreated all 11 copies
  and reran the naming gate green on the exact candidate. Earlier scratch-scope
  evidence is not used as exact-candidate hygiene evidence.
- The rc.7 managed Muse native-to-effective-file exception remains grounded in
  CIP-0017. Linux measurements and token-refresh safety remain unclaimed.

## Candidate SHA-256

| File | SHA-256 |
|---|---|
| `README.md` | `523974dc6ac771ed3afd1af4e7868a1c004ff470ed3a2431f65cd134fe181214` |
| `docs/environment-config.md` | `dc992319aaf518935f848b55d04e5e9a5d8a015f125ada0d752dc94753e6f975` |
| `docs/agent-logins.md` | `715d6721cc92b6adc6c8d3a5f798d4275cd28d6fcf98bf171777439718ac6358` |
| `diagrams/plantuml/agent-login-homes.puml` | `88f09cc1e94f78da6234085a48f7249b37193da932a8bb7ecd18b7d18e3ecd28` |
| `diagrams/plantuml/agent-login-homes.svg` | `d454bfdb44b37acd15d64e3c843675eb5411ec9d3b1f3aae6deaf3098ad03af7` |
| `diagrams/plantuml/agent-login-prepare.puml` | `fd08b7dcb6c19e1205537bcce98af44bd0c7d59b83839c8b8ea681b182fe8b7a` |
| `diagrams/plantuml/agent-login-prepare.svg` | `8f9bb1d2d455e5302c4fd1665838e696991504e3308ad79ba5f23581367b42d7` |
| `diagrams/plantuml/agent-login-resolution.puml` | `abd572aee619227d558d99c7a67f610bc3ce7e62f494eaa7692ca4ec3a09a934` |
| `diagrams/plantuml/agent-login-resolution.svg` | `a0f04f91ea2878544a2a7a21f18087b430d5ef80b6d4463cbf8dee6e0e8b67fe` |
| `diagrams/plantuml/agent-login-stores.puml` | `897d2c69738bd6a445de3ea087d92db7d7c27de5a58269432507f518cfc38244` |
| `diagrams/plantuml/agent-login-stores.svg` | `2b95b8958ee09d6b119bffb4247ce3b4cc6a8609538d70222f1499dc51727dde` |

## Hosted validation verdict

[CI run 38823186484, attempt 2](https://github.com/relux-works/curator/actions/runs/38823186484)
is **success** on `cd4487462837154c94ae958aca4bea557836d0e1`.
The fresh API read reports 20 successful jobs and two conditional skipped jobs.
The optional candidate-suite and optional self-hosted test job were skipped;
no skipped job is represented as passing evidence.

Only the previously failed Linux test job was rerun by this successor.
It now reports success, including the uploaded test stream and platform gate.
The other successful jobs are reused on the unchanged exact snapshot, workflow,
source, tests, configuration and dependency inputs. No full-suite replay is
claimed. Required test, race, Go driver, lint, self-test and interoperability
results are green. All original reviewer tests are unchanged.

| Command | Actual exit | Evidence |
|---|---:|---|
| `gh run watch 38823186484 --compact --exit-status --interval 30` | 0 | `hosted-watch-01.log`; completed hosted workflow is successful. |
| `gh run view 38823186484 --json headSha,status,conclusion,attempt,jobs` | 0 | `hosted-status-02.json`; verified success, attempt 2, exact expected SHA, 20 successful jobs and 2 conditional skips. |
| `gh run download 38823186484 -n test-evidence-ubuntu-latest -D .temp/TASK-261016-2rpqii/hosted-linux-evidence-02` | 0 | `hosted-download-02.log`; fresh Linux evidence downloaded and inspected. |

The fresh Linux `go-test.json` has no failed test or package records.
`TestReviewProbeLegacyLaneIdentityInvalid` has a passing event from this rerun.
The platform-case gate reports ok. This establishes that the sole failing
reviewer regression passes in the actual Linux suite without changing code,
expectations, timeouts, or configuration. The original timeout's cause remains
unknown; no external-infrastructure cause or defect fix is inferred.

The original `sh scripts/remote-gate.sh` invocation remains recorded as exit 1.
This successor did not rerun that snapshot-creation command or create a new
commit. The green evidence above is the actual hosted workflow rerun against
the existing byte-identical snapshot. Normal Change Request publication and
its configured validation remain owned by the board lifecycle.

Task-scoped logbook update: recovered the original failed hosted attempt through
an unchanged-candidate failed-job rerun. All required hosted jobs now report
success. Attached evidence distinguishes rerun commands, unchanged green
results reused, historical failures, explicit platform limits, and unknown
failure cause. All eleven candidate digests were rechecked immediately before
handoff. No product implementation change was necessary.
