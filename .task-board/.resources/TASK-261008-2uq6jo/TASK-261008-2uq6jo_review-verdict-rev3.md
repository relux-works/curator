# TASK-261008-2uq6jo — N6 credential answer bound — revision 3 review

Verdict: accepted. No blocking findings. Accept CR-TASK-261008-2uq6jo-3 revision 3 and route to integrating; this is review acceptance, not landing.

## Identity and evidence

Reviewed the complete four-path delta from base 3b6481c15d61b08d3f0fae7c3269329555336709 to candidate tree 794492d4b7ac5cff6acdfb5304cb3f13808fe26f. Fresh origin HEAD advertisement and exact refs/heads/main fetch both equal the base; upstream has not advanced. The candidate hosted commit 422b5e187aca27b6a3ea8d0fa03b7d8f22098a32 and the producer's disposable-clone unmutated snapshot c1d56bc33a4626bed141f69223dad22bc6a2c491 both have exactly the candidate tree (independently verified via GitHub Git commit objects).

Candidate gate: https://github.com/relux-works/curator/actions/runs/37854690852 — success. Independently inspected the three test-evidence-{ubuntu,macos,windows}-latest artifacts, including terminal events in test/go-test.json. All six size subtests pass on each platform: 18/18 overall; no gitcred skips. All four report controls pass on each platform: 12/12 overall. Lint, race on Linux/macOS, and all other enabled jobs succeed. The optional self-hosted and candidate-conformance jobs are skipped; no claim is made about those lanes.

## Previous verdict dispositions

The legacy revision-2 verdict named two mechanisms, without findings IDs. Both are resolved:

- Windows coverage: revision 2 unconditionally skipped both regressions and reworded their reasons. Revision 3 uses the real Git built-in store helper and the existing cross-platform test-binary fake-Git transport. Windows now executes and passes 6/6 size cases; no new skip class or reason laundering.
- Missing hosted mutation proof: revision 3 attaches three diagnostic mutants from the same exact candidate snapshot. Independently checked their single-file, single-hunk GitHub diffs and downloaded all nine platform evidence artifacts. Results below verify production wiring and both sides of the bound.

There are no current findings, so repeat-of is not applicable. No prior finding is being relabeled as a new mechanism.

## Swept surfaces

| Surface | Disposition and evidence |
| --- | --- |
| Production entry and bypass paths | Held. ReadHost reaches Access.call before any answer parsing. ReadScoped, ReadProvider, approve and reject also share Access.call. Overflow refuses the whole answer after cmd.Run; no alternate parsing path bypasses it. |
| Bounded writer and framing | Held. Strict > preserves exact-cap acceptance. Overflow is sticky across writes, including a nonempty write after capacity reaches zero. Remaining capacity cannot become negative; retained data stays bounded. Returning the full payload length continues draining stdout. M1/M2/M3 below exercise the two changed clause sites and exact boundary. |
| Real Git regression | Held. TestCredentialAnswerBoundRefusesOversized drives ReadHost with synthetic store-helper credentials: secrets 32 and 65280 survive unchanged; 65664 is refused. 3/3 cases on each of three platforms. |
| Exact frame regression | Held. TestCredentialExactFrameBound drives ReadHost with byte-exact 65535, 65536 and 65537 frames via fakeAccess/TestMain/fakeGitMain. At/below cap accepted; cap+1 refused. 3/3 cases on each platform. |
| Adjacent report controls | Held. TestEveryCallDisablesInteractivePrompting, TestReadHostRefusesAManagerNamespacedAnswer, TestReadProviderRejectsNearMissAnswers and TestStoreRejectsAHelperThatPersistsNothing pass, 4/4 per platform, in the candidate and all three mutants. |
| Windows fixture and test harness | Held. Store helper uses an isolated synthetic home; controlled Git reuses the existing executable re-entry fixture, with SYSTEMROOT supplied on Windows. No shell dependency, platform skip, or fixture route that avoids production stdout collection. |
| CHANGELOG and architecture | Held. Unreleased Fixed entry accurately explains oversized-answer refusal and absent credential. Four scoped paths, small change at the owning boundary, existing fixture extended without a separate harness. No unrelated code changes. |
| R223, outcomes and checklist | Held. Full live checklist is checked. Producer results and candidate validation are attached. Reviewer ran compile-only checks and inspected hosted evidence; no local go test, code changes, commits, branch changes or commit_ack. No LOGBOOK edit per explicit task brief. |

Free hunt: empty after sweeping the full delta and adjacent callers. Bounds of the review: six size cases across three platforms establish this framing invariant; no claim is made about every malformed answer grammar, duplicate key, platform keychain, or unlimited-output CPU/lifetime case.

## Hosted mutation proof independently verified

The producer ran these in a disposable clone; the reviewer reused and independently inspected the hosted artifacts and exact mutant diffs, without rerunning suites or mutants. Workflow Test lanes use -count=1.

| Mutant | Hosted run | Verified failures on Linux, macOS and Windows |
| --- | --- | --- |
| M1: remove only Access.call overflow refusal; 144d761d475d087eed2f3121c3d412ccd35c3f00 | https://github.com/relux-works/curator/actions/runs/37847866844 | TestCredentialAnswerBoundRefusesOversized/65664 and TestCredentialExactFrameBound/1. Truncated credential accepted and cap+1 admitted. |
| M2: narrow detection to len(accepted) > b.remaining+1; a16f0cb6e0238fde72cb4f472491390f24516482 | https://github.com/relux-works/curator/actions/runs/37847865978 | Only TestCredentialExactFrameBound/1 fails; exact-cap and real-Git oversized refusal still pass. |
| M3: change strict > to >=; 1fe63ffca2b1633979e45c0102c5b3d8444dcebc | https://github.com/relux-works/curator/actions/runs/37847865854 | Only TestCredentialExactFrameBound/0 fails; exact-cap acceptance is protected. |

All nine mutant platform lanes execute 6/6 size subtests with zero gitcred skips and no unrelated gitcred failures. The gate/** throwaway prefix is justified by ci.yml push triggers (scratch/** does not trigger CI); fresh ls-remote confirms all three mutant branch refs were deleted. The candidate tree contains none of the diagnostic mutations.

Reviewer local verification: go vet ./..., go build ./..., gofmt -l on the three changed Go files, and git diff --check; combined invocation exit 0 with no error diagnostics or formatting output. Producer compile-only tail also records individual exit 0 results. Test results above are exclusively hosted. No additional full-suite replay was needed for this identical frozen tree. spawn goal reports no active goal for this run.

Acceptance evidence satisfies the live checklist and the N6 invariant. Integration remains the authorized producer's next step.
