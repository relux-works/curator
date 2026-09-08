# E6 reconciliation — TASK-260909-1zcwvs

Date: 2026-09-09 (Asia/Tbilisi). Researcher handoff for STORY-260908-2utz8k.

**Finding: E6 already matches the landed specification and fragment implementation. Recommend evidence-only acceptance after independent review; no new implementation is needed.** `repository_delta: empty`. This outcome substitutes for LOGBOOK under the explicit task instruction. No README, SPEC, LOGBOOK, repository source, installation, home, or control-root files were changed. Board resource/lifecycle writes are the authorized deliverable. Acceptance and signed board completion remain the reviewer/coordinator's steps.

## Provenance

- [Fragment PR5](https://github.com/relux-works/curator-agent-launcher/pull/5): hosting API reports MERGED at 2026-09-08T16:28:26Z, head and merge commit `84e659e1bda41c0b70fad72e9e29b3c7ad474a7d`. Its two commits are `8f0c68e84d2de037d24e70116727a34b779d0647` (resolver, tests and E6 SPEC paragraph introduced) and `84e659e1bda41c0b70fad72e9e29b3c7ad474a7d` (Unicode path/corpus follow-up). The E6 delivery belongs to PR5; it was not newly introduced by its last commit.
- Inspected/tested worktree HEAD: `adf627607eb334e9839288cfffce63e1268ae688`. Fresh `git ls-remote --symref origin HEAD` advertised main at that exact OID; exact-ref `git fetch origin refs/heads/main` yielded the same FETCH_HEAD. Local HEAD..main count was 0. PR5 is its ancestor, and both resolver files are byte-identical between PR5 and HEAD. No branch was switched, rebased or committed.
- [Current SPEC §4.1](https://github.com/relux-works/curator-agent-launcher/blob/adf627607eb334e9839288cfffce63e1268ae688/SPEC.md#L240): mapping at lines 245–261, E6 transport paragraph at 263–274. `git blame` attributes the entire E6 paragraph to `8f0c68e84d2de037d24e70116727a34b779d0647`.
- Accepted A0: TASK-260908-qblycn is `done`; resources [findings](board-resource://TASK-260908-qblycn/outcome/TASK-260908-qblycn_a0-verification-findings.md), [accepted review rev1](board-resource://TASK-260908-qblycn/outcome/TASK-260908-qblycn_review-verdict-rev1.md), and [sanitized evidence bundle](board-resource://TASK-260908-qblycn/outcome/TASK-260908-qblycn_a0-evidence.tar.gz) were retrieved through the CLI and inspected. That acceptance applies to A0, not this new candidate.

## Exact behavior

[resolve.go](https://github.com/relux-works/curator-agent-launcher/blob/adf627607eb334e9839288cfffce63e1268ae688/internal/fragment/resolve.go#L144) implements the following:

| Curator result | Launcher result |
|---|---|
| nonzero + `environment_unknown` | `resolve_environment_unknown` |
| nonzero + `profile_unknown` | `resolve_profile_unknown` |
| nonzero + `environment_repair_failed` | `resolve_repair_failed` |
| nonzero + `environment_lock_unavailable` | `resolve_lock_unavailable` |
| nonzero + another recognizable code, including `environment_home_stale` | `resolve_invocation_failed`; CuratorCode retained |
| nonzero without recognizable code | `resolve_invocation_failed`; CuratorCode empty |
| zero + valid matching fragment + warning | successful resolve; warning forwarded |
| zero + invalid/mismatched fragment | `resolve_fragment_invalid` |

The parser (lines 200–226) takes the first syntactically recognizable line with exact `curator: ` prefix, nonempty lowercase-letter/digit/underscore code and a following colon. It trims detail whitespace. Warnings, uppercase codes, embedded prefixes and lines missing the detail colon are skipped. An unknown but syntactically valid first code is not skipped in favor of a later mapped code.

Mapped `Detail` is `curator env resolve exited N: CODE: DETAIL`. An unmapped code produces `curator env resolve exited N with unmapped diagnostic CODE: DETAIL`; no recognized code produces `curator env resolve exited N without a curator diagnostic line`. These strings are source-inspected, not exact-string assertions in the tests below.

`Request.Argv` always supplies `--repair`. `Resolve` executes once, forwards stderr through `io.MultiWriter` while capturing it, checks nonzero status before parsing stdout, and never returns a fragment on failure. `ExecRunner` routes process stderr to that writer. Thus unrecognized failure text such as `panic: boom` is preserved even though it is not copied into the structured Detail. Exit-zero warnings do not become a resolve error; valid stdout is still required. A nil Stderr writer deliberately discards operator output, so operator forwarding depends on the caller providing its writer.

The production caller is [cmd/curator-run/main.go](https://github.com/relux-works/curator-agent-launcher/blob/adf627607eb334e9839288cfffce63e1268ae688/cmd/curator-run/main.go#L75): main constructs `fragment.New()`, run passes its stderr into Request, then appends `curator-run: CODE: DETAIL` and returns 1 on resolve failure. Successful resolution currently continues to mapping and the explicit `not_implemented` refusal. E6 is delivered; full agent execution is not established by this finding.

## What A0 proves, and its bounds

A0 findings §2/§6 E6 and bundle `03-resolve-readonly.log` record `profile_unknown` and `environment_home_stale` diagnostic lines, each with exit **1**: expected failing resolve probes, not passing commands. The first-pass candidate was `v0.14.1-0.20260906225215-a66eec88fa4d+dirty` (source `a66eec88fa4dfb8f3ca098f594f6e48af57a6835`, documented dirty board activity). Do not relabel those initial failures as fresh observations on the later installed release.

A0 §7 and `11-resolve-installed.log` explicitly name installed `v0.14.1-0.20260907213730-04550e282705`, source `04550e282705e8dc361ca02233ad1557589a4b18`. With current homes, default-profile read-only and `--repair` resolves exited **0** for claude_code, codex_cli and pi. Claude/codex stderr contained `warning: environment_tool_version_unverified`; pi stderr was empty. Review rev1 independently reproduced successful read-only resolves and confirmed E6 against the refusal transcript.

This is historical transport evidence for those builds and inputs. It does not prove all four mapped failures occurred on a real installed binary, a stale-home failure under `--repair`, today's installed-tool behavior, or current model/home/ax behavior. No installed Curator or native agent was invoked in this reconciliation. The source enforces the broader mapping, covered with deterministic fixtures below. A0's unrelated combined suite timeout (exit 124) is not treated as a passing test or rerun here.

## Narrow verification performed here

Platform inspected: Go CLI, Go 1.25.5, darwin/arm64; no iOS target. Tests ran at the HEAD above, with `-count=1`, no tee/pipeline, no new tests, and shell fakes instead of installed Curator.

1. `go test ./internal/fragment -run '^(TestResolveSuccessForwardsWarnings|TestNonZeroExitMapping|TestNonZeroExitNeverYieldsFragmentEvenWithValidStdout|TestExecRunnerEndToEnd|TestExecRunnerNonZeroAndInvalidOutput)$' -count=1 -v` — **exit 0**, 5/5 selected tests passed. `TestNonZeroExitMapping` contains 13/13 table rows, covering 4/4 mapped codes plus warning-before-code, unmapped/stale, malformed and absent diagnostics; each checks verbatim stderr, error code, CuratorCode and exit. The nonzero/valid-stdout negative test refuses fallback. ExecRunner tests drive `New().Resolve` through a real subprocess with a shell fake.
2. `go test ./cmd/curator-run -run '^(TestRunParsedLaunchResolvesThenRefuses|TestRunResolveFailuresExit1|TestRunProductionResolverAgainstFakeCurator)$' -count=1 -v` — **exit 0**, 3/3 selected tests passed. Failure table has 9/9 rows including all 4/4 mapped codes and unrecognized `boom` stderr; operator stderr prefix, exit 1, one resolve and empty stdout are asserted. The success-path tests assert warning-before-later-refusal ordering; the production-resolver test uses the real ExecRunner with a shell fake.
3. Standalone `git merge-base --is-ancestor 84e659e1bda41c0b70fad72e9e29b3c7ad474a7d adf627607eb334e9839288cfffce63e1268ae688` — **exit 0**. Standalone `git diff --exit-code 84e659e1bda41c0b70fad72e9e29b3c7ad474a7d HEAD -- internal/fragment/resolve.go internal/fragment/resolve_test.go` — **exit 0**, empty.
4. Standalone `git diff --exit-code`, `git diff --cached --exit-code`, and `git status --porcelain=v1` — each **exit 0**, empty. `gh pr view 5 --repo relux-works/curator-agent-launcher --json number,url,state,headRefOid,mergeCommit,mergedAt,commits` — **exit 0**; JSON attached.

Bounds: the listed tests do not assert every Detail byte or the choice among multiple recognizable diagnostic lines; those properties were verified by source inspection. No mutation campaign, broad suite, CI or live-tool validation was run or claimed. Table ratios describe these existing cases, not universal input-space coverage. Initial project-local skill discovery returned exit 2 because the three searched skill directories were absent; this was discovery, not a green gate. Installed project-management skill was used. No implementation gap was found within E6; the remaining bounds do not justify manufacturing code for this erratum.

Evidence logs are attached as TASK-260909-1zcwvs_fragment-tests-01.log, TASK-260909-1zcwvs_cli-tests-01.log and TASK-260909-1zcwvs_pr5-01.json. The board outcome records the finding in place of the expressly forbidden LOGBOOK edit. Ready for independent review.
