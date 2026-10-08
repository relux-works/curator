# TASK-260927-1e5qqm — flip-codex-seed-to-revision-b

Verdict: accepted, CR-TASK-260927-1e5qqm-3 revision 3. No blocking findings. LANDING HELD until the operator schedules the B release. Acceptance routes to integrating; this reviewer performs no code edits, commits, checkpoints, integration, or done transition.

Reviewed exact base 35cac65962fda160cbb0dd52d42277a70ea23381 to candidate tree 41bf2cb318b2aa9337e92c29c9def7e958cba388, all six changed paths. Every working file matches its captured candidate blob.

| Swept surface | Result and evidence |
| --- | --- |
| Warning release | Held. Independently queried GitHub release v0.15.0-rc.3: published 2026-10-04T03:03:40Z, draft=false; tagged registry selects A. https://github.com/relux-works/curator/releases/tag/v0.15.0-rc.3 |
| Registry and architecture | Held. Only production code change is CodexSeedRevision A to B. Registry adapter consumes it at internal/envregistry/envregistry.go:279; Resolve dispatches through gatherSeeds at internal/envprofile/managed.go:2619, which strips before publication. CLI production entries call Resolve and StatusOf at cmd/curator/env.go:157 and :194. No new production seam. SecurityPostureRevision is B on the new base and unchanged by this CR; the older review note's A expectation is superseded by the explicit post-posture-B reapplication brief. |
| Semantic reapplication | Held. envstatus_test.go, codex_seed_test.go and envregistry.go are byte-identical to accepted revision 1 and revision 2. The sole marker-file difference versus revision 2 exactly matches upstream posture-B's explicit permissive fixture. All other marker changes versus revision 1 are upstream v1/v2 changes plus the same seed-B intent. |
| Gap ledger | Held. Independently filtering only this task's owner rows from base bytes produces candidate bytes exactly. Same 36 rows as revision 1, nine per each of four manifest identities; zero additions, foreign removals or remaining owned rows. Headers/comments/separators and all other bytes retained. |
| Fresh-main marker assertions | Held. Candidate equals base with exactly two intended replacements: explicit empty historical-A seed record after hash_version deletion, and pre-rule guidance A to B. All schema-3/hash-v2, v1 writer pins, genuine schema-1 downgrade, empty-A/nil-record and byte-preservation assertions are intact. |
| Documentation | Held. Exact base-byte substitution replaces the obsolete nine-seed-gap sentence with seed-B release and count text. All posture-B text is retained verbatim. Other pre-existing documentation text is unchanged. |
| Production vectors and negatives | Held. Current hosted Linux evidence measures 7/7 provisioning driven (B production 4/4, historical A 3/3), 8/8 posture driven (B production 5/5, historical A 3/3); zero gaps/bounds/skips. B uses empty override through shipped Resolve/StatusOf, A retains historical seam. All 15 child cases passed, including empty-table/no-warning, subtable stripping, A-home-unstripped-under-B and pre-rule-home guidance. Invalid TOML refuses before home publication; table/inline CLI stripping, preserved non-MCP members, B marker names, immutable repair snapshot and both legacy CLI repair byte-preservation tests pass. |
| Hosted gate identity | Held. Independently queried run 37718550711: completed/success, head 770bfe546f583ae6d244ae0f352a3895caa702bc. Commit tree exactly equals 41bf2cb318b2aa9337e92c29c9def7e958cba388. All 20 executed jobs succeeded: configured Linux/macOS/Windows tests, Linux/macOS race, lint, interop, naming, gate self-tests and Go-driver jobs. Optional rose-air and candidate-suite jobs skipped. https://github.com/relux-works/curator/actions/runs/37718550711 |
| Protected files and freshness | Held. LOGBOOK.md, CHANGELOG.md, scripts/remote-gate.sh and SecurityPostureRevision source are byte-identical to base; no dependencies or stray paths changed. Reviewer diff-check and gofmt checks pass. Fresh main advertisement equals exact-ref FETCH_HEAD 4b0ba5a8bc01c86c898a13849cf58e361e19b34a; its entire delta since base consists of board paths, none overlapping the six reviewed paths. Future landing must recheck freshness and honor the hold. |

Evidence inspected independently: TASK-260927-1e5qqm_change-request_rev3-validation.log records remote-gate exit 0, required=1 green=1 failed=0 missing=0 exact-command shards; aggregate test-case coverage remains unknown. Downloaded current run's Linux artifact 11525795341, test-evidence-ubuntu-latest, SHA-256 cd7f1a0fe81105185989b2af572a1403b83cd1c214de6349c61fbadc9fb90848. Parsed test/go-test.json: final pass for all 15 seed-vector children and these nine checks: TestEnvironmentsCodexSeedVectors, TestCodexSeedRevisionAWholeCopyWarningAndPostureRegression, TestCodexSeedStripsInlineMCPTable, TestCodexSeedRejectsInvalidTOMLBeforePublishingHome, TestCodexSeedSnapshotAndBytesDoNotRefreshOnRepair, TestEnvStatusReportsNotInheritedNativeCodexServers, TestEnvResolveStripsAndReportsInlineNativeCodexMCPTable, TestEnvResolveKeepsSchema1BytesForMetadataOnly, TestEnvResolvePreservesPreRuleCodexSeedAndReportsUnstrippedHome.

Exact observed count output:
- environments-codex-seed/provisioning-cases: 7 driven, 0 known-gap, 0 bound, 0 skipped, 7 total.
- environments-codex-seed/posture-cases: 8 driven, 0 known-gap, 0 bound, 0 skipped, 8 total.

Counts agree with the independent manifest-specific pins. Current corpus execution is for SPEC_PIN 43bf0a2506d5c354a73bbc3ea4623d4653db10c7, rc.14 manifest 6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5, not four separate corpus runs. Other platforms are verified at job level; Linux alone was inspected at case level. Skips are not passes.

Reviewer ran no local Go build/test/vet/run and made no adversarial source mutations. R194 excludes local cmd/curator tests; current exact-tree hosted evidence is the arbiter. Producer reports revision-3 locked targeted tests exit 0, envprofile 46.096s and stable syspolicyd count 40; these are accepted producer observations, not reviewer reruns. Local vectors skip without CURATOR_CONFORMANCE_ROOT. Prior switch-A and header-only narrowing mutants remain historical design evidence from revision 1; they are not claimed as revision-3 executions.

A preliminary byte-identity comparison against revision 2 flagged the marker test's upstream explicit permissive fixture. Inspection of the upstream diff and a corrected exact replacement comparison confirmed this is solely the expected rebase, not seed rework. No source changed.

Run goal queried immediately before verdict: no goal binding. Live merged checklist already complete. Findings are recorded in this task-scoped board outcome; LOGBOOK edits are prohibited by the brief and host rules. No prior rejection findings need resolution: revision 1 was accepted, revision-2 review was cancelled during convergence. This verdict supports accept_cr only, without commit_ack.

```verdict-findings
{
  "findings": [],
  "notes": [],
  "surface_results": [
    {"row":"Warning release","result":"held"},
    {"row":"Registry and architecture","result":"held"},
    {"row":"Semantic reapplication","result":"held"},
    {"row":"Gap ledger","result":"held"},
    {"row":"Fresh-main marker assertions","result":"held"},
    {"row":"Documentation","result":"held"},
    {"row":"Production vectors and negatives","result":"held"},
    {"row":"Hosted gate identity","result":"held"},
    {"row":"Protected files and freshness","result":"held"}
  ],
  "free_hunt": []
}
```
