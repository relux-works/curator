# SNAP-ID revision 1 deciding review

Task: TASK-261016-1v3gbc — implement-rc7-snapshot-identity-verifier.
Candidate tree: e1515c85c5ed4e3dbf932fb4549bf7eaea66e621.
Base: 0d00a109ad1da77a5506747c89f24af9442eb341.
Repository delta: present; 115 additions confined to the two declared leaf directories.
Reviewer CI snapshot: 655965c022e6387281140c161f9da34f6777028d; tree equals candidate, signature G.

**Verdict: REJECT / changes_requested. P1 bypass: Unicode casefold renormalization.**

```json
{
  "verdict": "changes_requested",
  "findings": [
    {
      "id": "unicode-casefold-renormalization",
      "row": "Closed metadata and semantic bindings",
      "invariant": "Snapshot paths must reject case/Unicode-normalization collisions, including directory components (managed-home-snapshots section 4).",
      "mechanism": "internal/envsnapshot/metadata.go:14 computes Fold(NFC(path)) without normalizing the fold result. Greek iota/upsilon plus diaeresis and acute produce distinct, canonically equivalent keys after case folding; validateMetadata lines 44-48 therefore misses both file and directory collisions.",
      "reproductions": [
        {
          "test_file": "TASK-261016-1v3gbc_reviewer_test.go",
          "test": "TestReviewerGreekCollision/{iota,upsilon}/{directory,file}",
          "command": "TMPDIR=\"$PWD/.temp/review-snapid/case-mount\" go test -count=1 -run '^TestReviewerGreekCollision$' -v ./.temp/review-snapid/probe",
          "expected_failure": "All four subtests fail with refusal = <nil>, want environment_materialize_snapshot_mismatch/path-collision. Each singleton positive control verifies; combined manifest remains unchanged.",
          "log": "TASK-261016-1v3gbc_review-greek-collision-rev1.log",
          "candidate_tree": "e1515c85c5ed4e3dbf932fb4549bf7eaea66e621"
        }
      ],
      "severity": "bypass",
      "priority": "P1",
      "repeat-of": "none"
    }
  ],
  "notes": [
    "4 of 4 mandatory paired rows have driving tests, but the Unicode collision class is incomplete.",
    "Windows/non-Unix actual-tree backend remains explicitly unsupported; no Windows success qualification is claimed.",
    "CLI/capture/publication/hosted integration are separate consumers and were not exercised.",
    "Reviewer hosted rerun was cancelled after the blocking bypass reproduced; it is not passing evidence. Earlier producer CI is historical evidence on the same candidate tree.",
    "No catalog-keyed surface table or review/free-hunt budgets were supplied; the four mandatory paired rows were swept."
  ]
}
```

## Scope and coverage

4 of 4 mandatory paired-test rows driven through the production library entry points. The brief supplied no catalog-keyed table or time budgets; its four paired rows are the sweep floor. CLI materialize/capture/publication and hosted launch are explicitly separate consumers, so no production CLI coverage is claimed. Repository caller search confirms no current integration; this review concerns the reusable library seam only.

Current curator-spec main was independently read via the hosting API at 15d970111277b1f68b3f7a4ddd63bc339b2c0a3f. Its managed-home-snapshots protocol is byte-identical to the producer pin. Sections 2, 4, 6 and 7 and CIP-0014 were read. Current upstream repository changes have no overlap with either candidate directory.

| Row | Result | Named attacks and production entry |
|---|---|---|
| Closed metadata and semantic bindings | broken | TestPinnedSchemaCases, TestMetadataFramingAndHashDomains, TestPinnedTrees, TestMetadataNeighbors, TestLaunchHashIncludesLF TestReviewerNullChannel and the failing TestReviewerGreekCollision; public codecs and Verify |
| Complete actual inventory and independent files | held | TestActualTreeRefusals, TestNoFollowSnapshotRoot, TestNoFollowBeforeParentSegment, TestMuseDerivedDirectoriesAndRootFact, TestReviewerManifestOnly; Verify and PortableEqual |
| Exact versus portable identity | held | TestPortableIdentity and TestReviewerPortableLeftDrift; Verify, PortableEqual and Verified.ManifestSHA256 |
| Original baseline retention | held | TestBaselineNeverRecaptured and TestReviewerRecapture; Verify with original digest |

Rows were appended to the local sweep artifact before proceeding to the next row. Held means the named attacks did not reproduce, not proof of absence.

Free hunt: metadata hardlinks, sticky content modes, missing metadata and valid Unicode singleton paths passed. Additional Greek normalization-after-casefold attacks reproduced one bypass mechanism in four variants. All other table rows remain held for their named attacks; no rows were omitted.

## Validation

Reviewer-executed commands, each exit 0:

- go test -count=1 ./internal/envsnapshot
- go test -count=1 -race ./internal/envsnapshot
- go test -tags snapshotcontrols -count=1 -run '^TestNarrowingControls$' -v ./internal/envsnapshot
- Initial TestReviewer* row probes and initial replay against isolated copies of the exact candidate package. The later Greek regression and final replay intentionally exit 1, as recorded below.

All 12 narrowing controls killed the named behavioral test; expected child exits were 1, not build/setup failures. This includes execute mode, extras, saved-manifest trust, source-token-preserving checkout conversion, portable producer version and hardlink count. Original source blobs remained unchanged.

Reviewer probes and replay script are task-scoped outcome resources. Download both into one directory, then run python3 TASK-261016-1v3gbc_review-reproduce.py from the exact candidate checkout. The script copies the package to scratch and changes only the test corpus lookup location; production source bytes are unchanged.

## Notes and bounds

- Windows/non-Unix actual-tree verification remains explicitly unsupported and fails closed. This review does not qualify Windows ACLs, reparse protection or successful tree verification. Unix verification and cross-platform codecs are the implemented helper slice; the platform backend remains outstanding for cross-platform materialize delivery.
- Ownership/ACL admission, capture consistency, hostile concurrent-writer protection, original-plan authorization and overlay identity remain consumer responsibilities. A self-minted digest is not authorization.
- No current CLI consumer calls this package. Wiring and production CLI qualification remain the consuming leaves' work.
- The repository gate's unsigned synthetic snapshot mechanism was not used. The reviewed existing task-local Git adapter signs and verifies its detached snapshot with configured identity; CI source/configuration and the Story branch/index are unchanged.
- No source-text-only gate is offered as behavioral evidence. The checkout control preserves the searched token and runs real Git checkout assertions.

## Task-scoped logbook

No product changes, branch switches or commits on the Story branch were made. Review artifacts were written only within the worktree and attached through board mutations. Fresh candidate byte comparison matched all 115 files. Early board query syntax failures and a Git read from an extracted spec directory were corrected; neither was treated as absence or normative evidence. The authoritative spec read above used the hosting API. No child agents were used.

## Blocking reproduction and required rework

On a private case-sensitive HFS+ image, each of these components verifies independently:
`\u03b9\u0308\u0301` versus `\u0399\u0308\u0301` (iota), and
`\u03c5\u0308\u0301` versus `\u03a5\u0308\u0301` (upsilon).
Combining either pair in one snapshot is incorrectly accepted, for both leaf files and directory prefixes with different children. All content bytes, modes, metadata hashes and inventory entries are otherwise valid. This is an actual successful public Verify call, not merely a diagnostic mismatch. The earlier case-insensitive-filesystem probe refused later with entry-set and is not the bypass evidence.

Required rework: compute collision keys with canonical equivalence preserved after case folding. Add the named production-entry regression with singleton positive controls and both file/component variants, plus a narrowing mutant that retains collision checks but admits this normalization-after-fold class. Do not create a separate harness or research leaf. Preserve baseline bytes on refusal. Re-run focused controls and the full gate, then return for another reviewer cycle.

Case-sensitive fixture commands used (scratch only):

```sh
hdiutil create -size 64m -type SPARSE -fs 'Case-sensitive HFS+' -volname SNAPIDReview .temp/review-snapid/case-sensitive.sparseimage
mkdir -p .temp/review-snapid/case-mount
hdiutil attach .temp/review-snapid/case-sensitive.sparseimage -nobrowse -noautoopen -mountpoint "$PWD/.temp/review-snapid/case-mount"
TMPDIR="$PWD/.temp/review-snapid/case-mount" python3 TASK-261016-1v3gbc_review-reproduce.py '^TestReviewerGreekCollision$'
hdiutil detach "$PWD/.temp/review-snapid/case-mount"
```

Download the replay script and reviewer test resource into the same directory before invoking it from the candidate root. On an already case-sensitive filesystem the mount is unnecessary; choose a case-sensitive TMPDIR. The image was detached after the observed exit-1 reproduction. An initial APFSX image-creation attempt failed with Invalid argument and was not counted as test evidence.

## Hosted rerun disposition

The reviewer launched `sh scripts/remote-gate.sh` on signed snapshot 655965c022e6387281140c161f9da34f6777028d, exact candidate tree. Run: https://github.com/relux-works/curator/actions/runs/38850704016 . After the P1 bypass reproduced, the reviewer requested cancellation and stopped only the recorded own gate/sleep PIDs; the local gate returned exit 143 and cleaned its temporary remote branch. No hosted green is claimed for this rerun. The producer's prior run 38841469667 is not relabelled as reviewer execution. A green suite cannot override the concrete four-variant public-entry bypass.

## Final task-scoped logbook

The free hunt changed the tentative no-findings assessment: normalization after casefold is a real gap. Private case-sensitive filesystem isolation was necessary to distinguish an actual bypass from the case-insensitive host's later entry-set refusal. No production code was changed. Record this revision as changes_requested and route once to to-dev.

Provider readback confirms reviewer run 38850704016 completed/cancelled. The temporary gate branch is absent after cleanup (successful ls-remote read). Final attached-script Greek replay also exits 1 with all four intended nil-refusal failures. Local environment: Go 1.25.5, darwin/arm64; isolated case-sensitive HFS+ fixture.
