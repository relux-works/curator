# BUG-261017-47xdyt — install-deletes-unowned-store-entries: review revision 2

Verdict: accepted
Reviewer run: RUN-261018-15b289
Change Request: CR-BUG-261017-47xdyt-2, revision 2
Base: 402fdbc9b77803cb2d6b11f4f1cfd83961e6897b
Candidate tree: 191172290510ad94239590f3d4a695c80bfd738b
Signed hosted candidate: db030e1faca9dfc0506876e14e724414ac563030, signature G

```json
{
  "verdict": "accepted",
  "findings": [],
  "notes": [
    "Previous stale-candidate-path-overlap finding resolved: README composition preserves trunk exactly plus one ownership paragraph. No ownership source/test changes since revision 1.",
    "No catalog surface table or review budget was supplied. The six supplied AC/review rows define this bounded sweep; no catalog-complete claim.",
    "Full local and hosted evidence reused only from this exact converged candidate. Reviewer directly reran targeted production tests, supporting packages, three mutants and two private-HOME CLI probes; no full-suite replay claimed.",
    "The v6-unaware case substitutes an unknown schema for this v6-aware reader; four declaration writer cases exercise marker.Write, not package admission. Windows special-file/permission substitutions remain explicit bounds. Concurrent hostile filesystem races beyond the serialized recheck are not claimed.",
    "Ownership markers are not cryptographic authorization. Correct minimal source identity remains repairable as required by spec; invalid build evidence is not confused with unproven ownership.",
    "An unsupported initial board query failed visibly and was replaced by supported projections/resource reads; it was not interpreted as absence. A process-presence probe returned 1 after the named processes had ended; completed command records, not that probe, establish outcomes.",
    "Accepted means ready for manual PR landing; no landing, done transition, source edits, commits or branch movement performed by this reviewer."
  ]
}
```

## Candidate and combination

Checked the refresh delta first. Nineteen of twenty changed candidate paths are byte-identical to revision 1; only README differs. Removing its single ownership paragraph reproduces current base README exactly, preserving CLIHELP and SNAPPUB additions without duplication. All twenty working paths equal the CR tree before and after testing. Story HEAD/index remain the checkpoint.

Fresh authorized remote HEAD advertisement and exact-ref fetch agree on `2e5ab6dab61fc78a525cbcb0de6c5fcc9f5ae882`. Its changes since the CR base are only CHANGELOG.md and RELEASE_SIGNERS: overlap with candidate is empty. This resolves the previous review's sole finding; it is not acceptance of a future rewritten landing head.

Normative ownership subset read from curator-spec 3ca4b84f, skillfile-sources §4.1, with core §10–§11 and manager §2.5. Both vendored ownership/writer corpora equal that spec's Git blobs exactly. AC coverage checked before reading implementation: **6 of 6 AC rows have named driving tests**, with explicit bounds above.

## Swept surfaces

# Revision 2 review sweep

Candidate: 191172290510ad94239590f3d4a695c80bfd738b. No catalog table or budget supplied; six AC rows are the bounded surface table. Self-imposed 25-minute sweep and five-minute free hunt. Coverage checked before code: 6/6 rows have named production driving tests; writer-only/unknown-reader bounds retained.

Refresh delta inspected first: ownership source/test blobs unchanged from rev1; README adds one ownership paragraph over base, preserving CLIHELP and SNAPPUB text. Fresh upstream changes only CHANGELOG.md and RELEASE_SIGNERS, disjoint from candidate paths.

1. AC1: held. Reviewer ac1 suite passed: unmanaged skill/notes/file conflicts named together; snapshots unchanged; foreign link resolves; malformed/wrong-name/link/read-failing stale variants refuse. CLI probe independently confirms item A.

2. AC2: held. ac2 suite exits 0 through Project: nine selected collision variants refuse, absent/owned repair/reinstall controls succeed. CLI missing-marker and malformed-carrier attacks preserve published state.

3. AC3: held. Six targeted published cases pass via Project: unreadable marker, special marker, linked parent, ownership removed before publication, owned v5/v6 invalid execution policy rederived. Full 21-case producer run reused from exact candidate; corpora independently matched normative Git blobs. Unknown-reader and Windows substitutions remain stated bounds.

4. AC4: held. ac1 mixed fixture and independent CLI probe refuse while conflict exists; once fixture explicitly removes foreign entries, actual owned stale node is removed. This is not a reject-everything implementation.

5. AC5: held. Global/hybrid canonical refusal, owned repair, unrelated hybrid sibling preservation, native adapter refusal and stale cleanup tests exit 0 via public production entries. Exact-candidate full transaction/GC/repair evidence reused; supporting package replay recorded separately.

6. AC6: held. Fresh GitHub API confirms full hosted success on signed db030e1f / exact CR tree (20 success, two conditional skips). All seven producer narrowing overlays inspected against current source and complete named-failure logs; three rerun independently (stale loose file, parseable wrong name, manifestless runtime) return 1 at intended assertions. Local evidence replay confirms all 98 command records end correctly; 455/455 archive payload digests verified. No source-text-only mutant claim.

Supporting marker/staging/transaction/scopes suites exit 0.

Free hunt: held, no findings. Built CLI rejects duplicate/escaped-duplicate keys, string schema, trailing document, null/array marker and nested-only identity. Every refusal preserves full published snapshots; owned repair and reinstall controls remain green. Both CLI probes exit 0 with individual refused installs exit 1.


Measured: **6 of 6 AC rows held; findings = 0**. “Held” means the named attacks did not reproduce, not proof of absence.

## Production calls and evidence

- AC1/AC4: `Project -> stageProjectTargets -> checkCanonicalStore / stageStaleSkillRemovals`; `TestInstallPreservesUnownedStaleStoreEntries`, `TestInstallRefusesUnprovenStaleOwnership`, `TestRemovedSkillCleanedUp`, and independent CLI item A.
- AC2: `Project -> stageProjectTargets -> stageNode -> marker.CheckOwnership`; `TestInstallRefusesUnownedSelectedCanonicalDestination`, `TestInstallOwnedCanonicalPositiveControls`, and CLI C2. No partial publication; named paths and foreign-link resolution checked.
- AC3: `TestInstallDraftOwnershipCases` through Project and serialized stagers; directly repeated six cases for permission failure, special marker, linked parent, ownership change before publication, and v5/v6 invalid-evidence repair. All 21 ownership and 10 writer corpus rows have exact-candidate full-suite evidence.
- AC5: `Global -> stageGlobalTargets` and project hybrid branch through the same ownership checks; `TestGlobalCanonicalOwnershipPreservation`, `TestHybridCanonicalOwnershipPreservation`, native refusal and owned cleanup all pass. Hybrid does not prune unrelated siblings. Supporting marker/staging/transaction/scopes rerun exits 0; full GC/repair suites reused from exact-candidate evidence.
- AC6: seven narrowing mutant sources compared to exact candidate, all complete logs show intended assertion failure. Three independently repeated: stale loose file loses named conflict; parseable wrong-name marker admits forbidden replacement; manifestless runtime guard admits forbidden declaration. Each exits 1 for that specific reason. Other four exact-candidate mutant results reused. No delete-only or static-only mutant evidence.

All direct command arrays, actual exits, full targeted logs, CLI drivers, overlay diffs and evidence digests are attached as `BUG-261017-47xdyt_review-evidence-rev2.txt`. Direct Go commands use `-count=1`; no cache-only result. Build and diff whitespace check exit 0. Candidate hygiene scan and manual scope review found no public operator material.

Full local evidence rechecked: 98/98 command records, 85/85 packages, 3504/3504 top-level tests driven, including all 425 CLI tests. Declared platform/helper skips are retained as skips. Failed/superseded batch is excluded and replaced by explicit green masks. Archive digest matches the attached locator; all 455 payload lengths/hashes verified. Existing full-suite evidence was inspected, not silently counted as reviewer execution.

[Hosted gate](https://github.com/relux-works/curator/actions/runs/39090264021) independently read from GitHub: completed/success on exact signed db030e1f; 20 executed jobs successful, two conditional jobs skipped. No synthetic status or local mirror. Producer lint/build evidence is bound to that candidate.

## Task-scoped logbook

Revision 2 resolves the sole revision-1 freshness finding with the correct combined README and fresh exact-tree validation. Independent refusal, preservation, repair, cross-scope and narrowed-gate attacks found no blocker. Record acceptance once using accept_cr revision 2; manual PR landing remains separate. Run goal read reports not goal-bound. All evidence is attached before the lifecycle mutation.
