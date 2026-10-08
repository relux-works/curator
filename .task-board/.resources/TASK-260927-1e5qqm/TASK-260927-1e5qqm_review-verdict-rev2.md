# TASK-260927-1e5qqm — flip-codex-seed-to-revision-b

Verdict: accepted, CR-TASK-260927-1e5qqm-2 revision 2. LANDING HELD until the operator schedules the B release. This reviewer performed no source edits, commits, checkpoints, integration, or done transition.

Reviewed base `75ab9a71a9b9049ec4242b4acdb60c3fead6aff1` to candidate tree `357b286a881a72071664fe531f028a2c17360b7e`, all six changed paths. The six working-tree files equal their captured candidate blobs. No actionable findings.

The previous verdict resource is an acceptance, not a rejection: revision 1 was accepted on its exact candidate. The generic round brief's reference to a rejection does not match that artifact. Its legacy prose has no numbered findings; there are no previous rejection mechanisms to answer or repeat. This round verifies the reapplication and fresh-main combination.

| Swept surface | Result and evidence |
| --- | --- |
| Warning release prerequisite | Held. GitHub release v0.15.0-rc.3 remains published, draft=false, published_at 2026-10-04T03:03:40Z; tagged registry selects A. https://github.com/relux-works/curator/releases/tag/v0.15.0-rc.3 |
| Registry and collateral defaults | Held. CodexSeedRevision selects B. SecurityPostureRevision remains A in unchanged internal/config/security_posture.go. No alternate production path or new seam is introduced. |
| Semantic reapplication | Held. Compared revision-1 tree 40da016696c6a0673cc4b7f260ba8a923559e358 against revision 2: envstatus_test.go, docs/ci-gates.md, codex_seed_test.go and envregistry.go are byte-identical. The only other changed reviewed paths contain upstream ledger cleanup and upstream marker-test changes plus the same seed-B intent. |
| Gap ledger | Held. Independently removed the owner-matching lines from the fresh-base bytes and compared with candidate bytes: exact equality. Exactly the same 36 rows as revision 1 are removed, nine per manifest identity, zero additions, zero foreign-owned removals, zero remaining rows for this leaf. No header, comment, separator or other byte changed. |
| Fresh-main marker assertions | Held. Candidate env_credential_marker_test.go equals the fresh-base file with only the historical empty A seed-record fixture added after hash_version deletion and the pre-rule guidance changed from A to B. All v1 writer pins, v3/hash-v2 credential assertions, genuine schema-1 downgrade checks, nil-record/empty-A-record assertions and byte-preservation checks remain intact. Hosted CLI tests for both legacy repairs passed. |
| Production vectors and exact counts | Held. B vectors have empty override and exercise Resolve and StatusOf through the shipped registry; A vectors retain the explicit historical seam. Current hosted Linux evidence measures provisioning 7/7 driven (B production 4/4, A historical 3/3) and posture 8/8 driven (B production 5/5, A historical 3/3), zero gaps/bounds/skips. Count pins and the independently pinned rc.14 corpus agree. CLI production call sites: cmd/curator/env.go:157 and :194; Resolve consumes the registry adapter and calls gatherSeeds at internal/envprofile/managed.go:2619. |
| Stripping, failures and preservation | Held. Current hosted evidence passes table and inline CLI stripping/status, invalid-TOML refusal before home publication, immutable B seed/snapshot repair, historical A-home preservation under B and pre-rule marker/config preservation with B guidance. All 15 seed-vector child cases pass, including empty-table/no-server no-warning cases and A-home-unstripped-under-B. CLI checks assert stripped TOML, preserved non-MCP members, B snapshot names and not-inherited warnings without the old migration hint. |
| Hosted gate and candidate identity | Held. Independently queried GitHub run 37710932288: completed/success, head 813ae4fe29fab2b76c56c80a1c7cc143d52d86d2. GitHub commit tree is exactly 357b286a881a72071664fe531f028a2c17360b7e. Configured Linux/macOS/Windows tests, Linux/macOS race, lint, interop, naming, gate self-tests and Go-driver jobs succeeded. Optional rose-air and candidate-suite lanes were skipped. https://github.com/relux-works/curator/actions/runs/37710932288 |
| Protected files and freshness | Held. No LOGBOOK.md, CHANGELOG.md, scripts/remote-gate.sh, dependency or other stray delta. Reviewer diff checks and gofmt checks pass. Fresh main advertisement and exact-ref FETCH_HEAD both equal base 75ab9a71a9b9049ec4242b4acdb60c3fead6aff1; no current upstream combination remains unreviewed. Future landing must recheck freshness and honor the release hold. |

Current hosted proof: board artifact TASK-260927-1e5qqm_change-request_rev2-validation.log records remote-gate exit 0, required=1 green=1 failed=0 missing=0 exact-command shards. It states aggregate test-case coverage unknown. Reviewer additionally downloaded GitHub artifact 11522471890 (`test-evidence-ubuntu-latest`) from that same run and parsed its merged go-test.json; SHA-256 `9f8e77a568988cc63dfa0a17793ae503ab3c2f94c40a7591f4b45124ab838677`. Observed final pass events for all 15 vector children and these nine parent checks: TestEnvironmentsCodexSeedVectors, TestCodexSeedRevisionAWholeCopyWarningAndPostureRegression, TestCodexSeedStripsInlineMCPTable, TestCodexSeedRejectsInvalidTOMLBeforePublishingHome, TestCodexSeedSnapshotAndBytesDoNotRefreshOnRepair, TestEnvStatusReportsNotInheritedNativeCodexServers, TestEnvResolveStripsAndReportsInlineNativeCodexMCPTable, TestEnvResolveKeepsSchema1BytesForMetadataOnly, TestEnvResolvePreservesPreRuleCodexSeedAndReportsUnstrippedHome. The precise published-family output is:

- environments-codex-seed/provisioning-cases: 7 driven, 0 known-gap, 0 bound, 0 skipped, 7 total.
- environments-codex-seed/posture-cases: 8 driven, 0 known-gap, 0 bound, 0 skipped, 8 total.

Corpus identity: SPEC_PIN 43bf0a2506d5c354a73bbc3ea4623d4653db10c7, rc.14 manifest SHA-256 6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5. All four historical manifest ledgers lose the same owned nine gaps; actual case execution/counts in this round are proven for the pinned rc.14 corpus, not four separate corpus runs. Other platform success is proven at job level; case-level evidence was independently inspected on Linux only. Optional skipped lanes are not claimed as passes.

Reviewer ran no Go build/test/vet/run locally and performed no mutations for adversarial testing. R194 excludes local cmd/curator execution. Accepted current exact-tree hosted execution as the arbiter. Producer results report the revision-2 locked targeted command exit 0, vector execution skipped locally because CURATOR_CONFORMANCE_ROOT was unset, vet exit 0 and stable syspolicyd count 40; these are producer observations, not reviewer reruns or local vector proof. Revision-1 narrowing/switch-A mutant evidence remains historical design evidence, not validation of the changed revision-2 tree. Current behavior and negative refusal are established by current hosted tests.

Run goal queried before verdict: no goal binding. All live merged checklist items were already checked. There is no task surface-table precondition; the full task-specific sweep above is recorded with structured results below. Findings are recorded on the board and in this verdict rather than LOGBOOK.md, as explicitly required by the seedB brief and host rules. This artifact supports accept_cr only, with no commit_ack and no done transition.

```verdict-findings
{
  "findings": [],
  "notes": [],
  "surface_results": [
    {"row": "Warning release prerequisite", "result": "held"},
    {"row": "Registry and collateral defaults", "result": "held"},
    {"row": "Semantic reapplication", "result": "held"},
    {"row": "Gap ledger", "result": "held"},
    {"row": "Fresh-main marker assertions", "result": "held"},
    {"row": "Production vectors and exact counts", "result": "held"},
    {"row": "Stripping, failures and preservation", "result": "held"},
    {"row": "Hosted gate and candidate identity", "result": "held"},
    {"row": "Protected files and freshness", "result": "held"}
  ],
  "free_hunt": []
}
```
