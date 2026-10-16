# Review verdict — TASK-261016-i5gwaf: M2 Muse 1.4.8 exclusion replay

**ACCEPT revision 3 as a qualification record, not successful interactive exclusion.**

CR-TASK-261016-i5gwaf-3. Base `dccac3199aa5c543c589c129f5ef2bc437523a2a`; candidate tree `187619f43617eeaec4d7cac2df091199581c9c1b`. Candidate delta is present and limited to the evidence/probe module and justified qualification, README, logbook and unresolved-question documentation. Converged README retains the independent login reference and tools entries without conflict.

## Prior findings and surface sweep

Revision 1's `contradictory-loader-attestation` and `snapshot-link-attestation` no longer reproduce in the named regression tests. Duplicate count/outcome/event fields refuse; actual frozen source and real reference CLI reject listed and dangling links beside valid regular controls. Five targeted narrowing mutants and three retained mutants fail at the intended assertions. No repeat finding remains.

| Supplied surface | Result | Attacks |
|---|---|---|
| Vendor-build evidence authorizing form-specific exclusion | held | Contradictory fields; missing/failed/truncated observations; actual-source and real-CLI listed/dangling symlinks; altered/missing/extra root bytes; eight narrowing mutants; exec omission on both roots; offline boundary controls |

The sweep was written before free hunt. Additional reference-CLI attacks rejected directory, metadata and bundle-root links, changed/missing roots and extra files, with an accepted regular control and unchanged source bytes. Free hunt found no blocking mechanism. Interactive omission remains a measured survivor because the vendor boolean does not exclude; it is not credited as protection.

## AC coverage and fresh execution

**6 of 6 in-scope AC rows driven through named production entrypoints.**

| Row | Test / production entrypoint | Result |
|---|---|---|
| Parser | TestMuse148ParserRefusals and LoaderMatrix / real Muse exec and TUI parsers | Two intended exit-2 refusals; valid forms accepted |
| Root/fallback/skills | TestMuse148LoaderMatrix / real context and skill loaders | Both forms, three root states, separate sources and exclusions |
| Required exclusion | LoaderMatrix and InteractiveExclusionRegression / loader and startup banner | Exec removal; interactive retained-skill FAILURE on both roots |
| Boolean without root | LoaderMatrix/interactive/absent / real TUI loader | Rules fallback retained |
| Link and frozen snapshot | FrozenSnapshot, FrozenSnapshotSourceInputs, ReferenceSnapshotPhysicalClosure, matrix / Curator install+resolve, reference CLI, Muse | Real link and independently checked regular closure |
| Workspace residual | TestMuse148WorkspaceResidual / loader and TUI /rules | Recorded counts/listing reproduced; no model precedence claim |

Fresh commands (vendor/reference locators privately supplied as documented in the candidate README):

- `go build -o .temp/review-muse-r3/curator ./cmd/curator`: exit 0.
- `go test ./.research/muse148 -run '^(TestObservationChecker|TestReviewContradictoryObservation|TestFrozenSnapshot|TestFrozenSnapshotSourceInputs|TestReferenceSnapshotPhysicalClosure)$' -count=1 -v`: exit 0, no reference skip.
- `go test ./.research/muse148 -count=1 -v -timeout=5m`: exit 0, all 44 vendor cases and offline controls, no skips.
- `go vet ./.research/muse148`: exit 0; `golangci-lint run ./.research/muse148/...`: exit 0, zero issues; `git diff --check`: exit 0.
- Eight candidate narrowing patches applied separately in disposable copies. Named masks from the candidate mutant README each exit 1 at their intended behavioral assertion; no setup failure counted.
- `MUSE148_OMIT_EXCLUSION=exec go test ./.research/muse148 -run '^TestMuse148LoaderMatrix$/exec/(linked|snapshot)/skills' -count=1 -v`: exit 1, required-exclusion cases fail on user_skills=1; both controls pass.
- Corresponding interactive omission command: exit 0, explicit survivor, no protection claim.
- `python3 .temp/review-muse-r3/free-hunt.py`: exit 0 for the attack driver; regular reference CLI exit 0, six invalid variants exit 1 at physical-entry-type, entry-set or file-identity.

Muse identity is the candidate's exact `1.4.8-R6256.1` and SHA-256 `5721763b4723f961709cd28f69aa049763965b46c4d4d70f883a0a90e5ddf39c`, enforced by the harness. Real disposable HOME/XDG, echo provider, auto-update off and sandbox network/Keychain denial remained active. No native login or credential access. macOS 27 arm64, Go 1.25.5. Normative main `15d970111277b1f68b3f7a4ddd63bc339b2c0a3f` reread: environments 7.1, 7.9, 10.2; CIP-0017 leaf 2/minimal probes; Muse v4 conformance. No conflict with the leaf's explicitly separate model/credential boundary.

## Full gate identity and limitation

[Producer full gate 38869688515](https://github.com/relux-works/curator/actions/runs/38869688515), attempt 2, independently read as completed/success: 20 successful jobs, 2 configured optional skips. Signed head `a88ce7997152b97e99389b6bbf9303e5e210e75f` has exactly candidate tree `187619f43617eeaec4d7cac2df091199581c9c1b`; no source/test/configuration delta exists. Hosted job/toolchain configuration is the same. Its recorded initial Windows timing-test failure is retained; subsequent hosted retry is legitimate success, not a local override. Hidden vendor probe is separately rerun above.

A redundant reviewer `sh scripts/remote-gate.sh` run was launched and then cancellation requested once exact-tree prior success was established: [38890939343](https://github.com/relux-works/curator/actions/runs/38890939343), signed snapshot `943be40892ff2253aa4ca0c71a6e34aa47287894`, same tree. Waiter exit 137 is intentional termination; this duplicate is **unverified**, never reported green. The owned temporary branch was deleted. Final hosted cancellation readback is in the attached evidence.

## Required verdict schema

```json
{
  "verdict": "accepted",
  "candidate_revision": 3,
  "candidate_tree": "187619f43617eeaec4d7cac2df091199581c9c1b",
  "findings": [],
  "notes": [
    "6 of 6 in-scope AC observation rows have named driving tests. Full 44-case real vendor replay passed as an observation witness. Exec exclusion succeeds; interactive exclusion fails on both root representations, as required by Rework 1. The section 7.9 interactive release gate remains pending.",
    "No catalog-keyed surface table or wall-clock budgets were supplied. The one declared surface was swept and persisted before free hunt. Held describes executed attacks only, not proof of absence.",
    "No runtime qualification constants or credential policy changed. Model-visible precedence/removal, refresh, actual materialize integration and credential backend qualification remain unmeasured/out of scope.",
    "Full gate reused from the exact candidate tree, not a claim of a fresh reviewer full-gate pass. Producer run 38869688515 attempt 2 is completed/success on signed snapshot a88ce7997152b97e99389b6bbf9303e5e210e75f, whose tree equals this candidate. Its attempt-1 Windows timing failure remains documented and was not reclassified as external failure.",
    "Reviewer started a duplicate full gate on signed snapshot 943be40892ff2253aa4ca0c71a6e34aa47287894 (also the exact candidate tree), then requested cancellation after independently confirming exact-tree prior success. Owned waiter was stopped (exit 137), owned branch deleted. This duplicate is unverified, not green; no local mirror or synthesized status used.",
    "Fresh upstream main 58b2aac337f73fb313893611d9de28e9a4fcff0a differs from base by non-overlapping snapshot-verifier/board paths. This review accepts only the supplied candidate. Landing must apply normal freshness and combined-tree validation requirements; no landing is claimed.",
    "Initial unquoted normative-fetch URL failed shell glob expansion, and an initial board query used unknown task operation; corrected reads succeeded. Neither is validation evidence.",
    "All 69 candidate paths matched frozen blobs before and after review. Reviewer writes are confined to ignored worktree scratch and task-scoped board mutations. No code or candidate documentation edited."
  ]
}
```
