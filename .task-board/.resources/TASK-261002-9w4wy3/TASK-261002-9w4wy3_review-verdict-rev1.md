# TASK-261002-9w4wy3 — windows-exec-hardlink-origin-checks

Verdict: accepted, revision 1. No revision requests.

Candidate tree 91fbebfbc4f368b876cafe200844e8edc5d94a23; base 2cb29dac8a4c82c5a07d7ca2d107aa6e6e2c93e7. All 13 changed paths independently compared byte-for-byte to candidate. Fresh origin main advertisement equals base. Full hosted gate commit 1a3c01ca777eade070347acd70cd17dd3a545ffe has no tree delta from candidate. Native qualification snapshot 494088911f5d76fc1084f530ab6afda10dc08078 differs only by its temporary workflow.

## Swept surfaces

| Surface | Result |
| --- | --- |
| Production resolution and VerifyExec | Both new refusals execute through readExecIdentityAt before/after hashing; launch revalidation repeats the proof. Component-store positive and uncaptured-SystemRoot negative preserved. |
| Native observation | Real GetFileInformationByHandleEx FileIdInfo, GetSecurityInfo owner SID, FindFirstFileNameW/FindNextFileNameW enumeration. Read/API errors, incomplete counts, changed IDs and unknown ownership fail closed. Every alias must match the open executable ID; complete count and physical component-store boundaries enforced. |
| Non-Windows seam | Fixture validates expected open target; positive case proves reachability. Independent mutants demonstrate neither refusal passes for an unrelated reason. |
| Published family | 8/8 production entries driven, zero gaps/bounds/skips. Interpreter cases use ResolveInterpreter; declared executable cases use deriveProfileForPlatform. |
| Ledgers/docs | Exactly 2 gap rows removed, 0 added. Platform ledger adds two explicit coverage rows and no skips. CI documentation matches selected suite coverage. |
| Hygiene/architecture | One CHANGELOG entry, no LOGBOOK change. Platform observations isolated behind existing resolver boundary; no metadata-controlled authorization seam. |

## Independently executed checks

Targeted go test ./internal/scriptworker -run 'Test(ExecutableIdentityCasesAtProductionEntry|UncapturedSystemRootHardlinkRegression|WindowsExecHardlinkOriginEvidence|DeriveProfileUsesDefaultExecSearchDirsAndBuildsDeclaredExecFarm)$' -count=1 -v: exit 0. Family reports 8 driven, 0 known-gap, 0 bound, 0 skipped, 8 total.

GOOS=windows go vet ./... on candidate: exit 0, no diagnostics. Same command on an isolated git archive of base: exit 0, no diagnostics. git diff --check: exit 0.

Mutants used Go overlays in a temporary directory; repository code never modified. Each ran go test with -count=1 -timeout 2m:
- Drop owner check: exit 1; windows-exec-unowned-file-hardlinks incorrectly accepted.
- Drop store boundary check: exit 1; windows-exec-noncomponent-store-hardlinks incorrectly accepted.
- Narrow owner check to empty SID only: exit 1; unowned-file case incorrectly accepted.
- Narrow traversal to first two links: exit 1; third-link-outside-store incorrectly accepted.

Exact source overlays and failure logs are in TASK-261002-9w4wy3_reviewer-mutants.zip.

## Reused hosted evidence, inspected independently

Native run https://github.com/relux-works/curator/actions/runs/36956099103: Windows 2022 and latest both succeeded, including vet/build/package tests. Logs explicitly show real owner SIDs, actual aliases, all five native subtests passing and 8/8 family counts. SHA256 windows-2022 tests.log: 9114a8ad3c6bd8b20ba9e75478cfb201ed0d0d771c092bdfa817a660a263f9f7; windows-latest: 34532f3bca97ca493ba27352565ee250e94251dbc8cb328f13796cbe47406ce2. Both match producer evidence. No new Windows run launched by reviewer.

Exact-candidate full gate https://github.com/relux-works/curator/actions/runs/36958450989: independently queried success; attached validation log exits 0 and records Windows/Linux/macOS tests, race lanes, lint, naming and conformance successes. Existing optional/candidate lanes skipped as recorded. Accepted existing gate evidence rather than replaying unchanged full suite.

## Host incident and lifecycle

This final verdict supersedes the interim stop-the-line resource. Host execution stalled, then recovered with syspolicyd running and successive crashes increasing from 352 to 353. The interim blocked status was returned to reviewing, and all remaining checks completed. No external blocker remains. spawn goal reports run not goal-bound (exit 0). No commits, integration, commit_ack or source edits performed by reviewer. Acceptance routes to integrating for producer-owned integration.
