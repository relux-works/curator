# TASK-261018-70enqt review, revision 3

Candidate tree: `b895a985af0758f3166367c7c1357fdfe57df6e1`.
Base: `59746a4545db06607446dcfe5fc35f28118a9125`.
Fresh upstream: `5ecb68da8f8ceab4dda118f2263abfb3da63ab8c`.
All 117 changed working files match the exact candidate blobs. Fresh upstream and candidate changed-path intersection is empty. README and CHANGELOG are unchanged from the base. The latest rc.7 brief supersedes the initial CHANGELOG request.

## Coverage and sweep

Producer coverage is **11 of 12 AC rows driven**, with row 12 explicitly a preservation/review bound. This is research qualification, not a new production gate. The brief supplies neither a catalog surface table nor review/free-hunt budgets. This review uses the supplied 12-row AC matrix, with bounded adversarial checks; it does not claim catalog-wide security coverage. No additional production coverage is demanded of the research leaf.

All eight vendor probes were independently rerun against the exact candidate, with synthetic credentials, temporary HOME/CODEX_HOME and a loopback endpoint. Actual binary identity is recorded in the review evidence. All passed. Default invocation independently produced eight skips before vendor fixture creation. Registry tests passed.

| Row | Result | Executed attack / production entry point and bound |
| --- | --- | --- |
| 1 CODEX_HOME | held | Changed only selected home; root and skill markers disappear in codex debug prompt-input; TestCodex162HomeRootSkills. |
| 2 root/advisory/replacement | held | Oversized root tail retained; small/large production Install -> Resolve warning controls; replacement captured in real exec request. TestCodex162RootSize, ManagedAdvisory, ManagedChannels. |
| 3 skills | held | Real and linked skill discovery versus empty-home control through debug prompt-input; TestCodex162HomeRootSkills. |
| 4 config seed | held | Real exec selects configured loopback model; production-managed seed loads; TestCodex162ExecSeedProfile and ManagedAdvisory. |
| 5 auth | held | Other home and missing link refuse login status; login writes through link without changing inode. TestCodex162LinkedAuth; synthetic file backend only. |
| 6 outside Git | held | Confirmed non-Git cwd; unflagged exec refuses and flagged exec succeeds. TestCodex162ExecSeedProfile. |
| 7 profile/MCP | held | Missing profile falls back, repeated flag refuses, absent MCP layer hides server; real exec and mcp list, TestCodex162ExecSeedProfile and ManagedChannels. |
| 8 hook trust | held | First and second project prompt independently; same-project restart skips after consent; no pre-consent marker. Real TUI, TestCodex162ProjectHookTrust. Hook execution itself unmeasured. |
| 9 sandbox | held | Root/additional-root .git write attempts refused with unchanged bytes; ordinary and nested .git controls writable; real separate-dir git config succeeds. TestCodex162GitSandbox, macOS only. |
| 10 release pin | held | Production envregistry.ByID pin test passes. Neighbor-release narrowing control recorded below; production credential policy reads this field. |
| 11 opt-in | held | Unsetting CODEX162_BINARY yields eight skips, while explicit opt-in executes eight tests; vendor entry points are behind the early fixture gate. |
| 12 preservation bound | held | Exact candidate comparison leaves root README/CHANGELOG untouched; explicit subprocess environment excludes operator auth/config, synthetic linked auth resides inside fixture. No whole-HOME filesystem attestation is claimed; Go caches excluded as documented. |

## Additional hunt and evidence identity

No additional blocking finding reproduced. Reviewed wrapper scope, loopback request assertions, missing-auth negative paths, default early skip, subprocess environment allowlist and raw-log publication redaction. The bounded sensitive-locator scan found no matches. No new source-text checker is shipped; token-preserving source-checker mutation is inapplicable.

The candidate source-hash manifest matches replay, mutants, registry, dependent tests and lint config, but its go.mod/go.sum hashes are historical and no longer match after converge. Those two historical identities are not accepted as current evidence. This review reran the entire opt-in replay and default/registry suite with current dependencies. Full hosted validation is independently bound to commit `58b43638b8a14a1b2e3a4a79176573c7cebc6941`, whose tree is exactly this CR tree and signature reports G. The attached revision-3 validation log reports run 39178320359 success and exit 0, including Complete runner; it is reused for full-suite/build/lint checks, not represented as a reviewer rerun. The reviewer did not launch a second hosted gate.

Research text honestly records the descendant .git counterexample and states that it does not establish the pilot phrase's intended meaning. This is a measurement correction, not a silent product adaptation or a changed registry field. Synthetic API-key file auth does not qualify OAuth refresh or native stores. No whole-host HOME mutation audit is claimed. All these bounds remain visible to downstream consumers.

The five extra credential-test expectation updates follow the registry release field; no credential production logic or assertions were removed. Changelog text is retained in the gaps resource under the latest wave instruction. Historical checkpoint prose is historical, not the current identity supplied above.

## Narrowing controls and commands

Reviewer reran `sh .research/codex162/mutants.sh` with the same qualified binary and fresh temporary output. Runner exit 0: all seven individual tests exit 1 with their intended named assertion (skip-git, repeated-profile, auth-other-home, root-git, trust-second, root-git-extra-root, release). These are narrowing sensitivity controls, not proprietary vendor source mutants. No mutant survivor is counted as a killed control.

Reviewer commands: `env -u CODEX162_BINARY go test ./.research/codex162 ./internal/envregistry -count=1 -v` (exit 0); `CODEX162_BINARY=<qualified-binary> CODEX162_EVIDENCE=<scratch>/replay TMPDIR=<scratch> go test ./.research/codex162 -count=1 -v` (exit 0, eight probes pass); `CODEX162_BINARY=<qualified-binary> CODEX162_MUTANTS_ROOT=<scratch>/mutants TMPDIR=<scratch> sh .research/codex162/mutants.sh` (exit 0, seven controls detected). Binary SHA-256 matches producer evidence: `74c6a6d263f40d65f99c27a815275f5b4ed0bcc2baecc396b4cf3255498bdfc0`, Codex 0.162.1, darwin/arm64, Go 1.25.5.

Raw rerun logs, identities, and fetched hosted evidence are attached as `TASK-261018-70enqt_review-evidence-rev3.json`. Local scratch is `.temp/review-codex162/`. No candidate source was edited, committed or landed by this reviewer.

## Verdict

```json
{
  "verdict": "accepted",
  "revision": 3,
  "findings": [],
  "notes": [
    "Historical source-sha256 manifest has stale go.mod/go.sum hashes; current replay and exact-tree hosted validation replace that historical reuse claim.",
    "No catalog surface table or wall-clock budgets were supplied; the complete expanded AC matrix was swept.",
    "AC coverage is 11/12 executable rows plus the explicitly stated preservation bound, not a whole-HOME filesystem audit.",
    "Nested .git is writable in the measured vendor sandbox; research records this counterexample rather than claiming recursive protection."
  ],
  "free_hunt_findings": []
}
```

Task-scoped review logbook: 2026-10-19 — exact candidate independently replayed; all eight probes and seven narrowing controls behave as recorded, full hosted validation matches candidate tree, no blocking finding reproduced. Accepted for downstream landing, not landed.
