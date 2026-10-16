# SNAP-ID revision 2 deciding review

Task: TASK-261016-1v3gbc — implement-rc7-snapshot-identity-verifier.
CR: CR-TASK-261016-1v3gbc-2, revision 2.
Base: 0d00a109ad1da77a5506747c89f24af9442eb341.
Candidate tree: 357a1081c0bc514e4543085cab5ae0264cf6ccb5.
Reviewer signed CI snapshot: d313a63f9d905ebe0a496e6de74f72d95649404e; exact candidate tree.
Repository delta: present, 116 files under the two declared leaf directories.

**Verdict: ACCEPT — no P1/P2/P3 blocking findings.**

```json
{
  "verdict": "accepted",
  "findings": [],
  "notes": [
    "Prior unicode-casefold-renormalization finding is fixed; four case-sensitive public-entry regressions pass and the M12 narrowing control is killed.",
    "4 of 4 AC rows and 4 of 4 mandatory paired rows are driven at the public library boundary; no materialize CLI or hosted-launch qualification is claimed.",
    "Unix actual-tree backend only: Windows fails closed and remains a backend gap for later cross-platform materialize delivery.",
    "The brief supplied no catalog-keyed surface table or wall-clock/free-hunt budgets; the four mandatory paired rows were the sweep floor.",
    "No findings in the additional Unicode free hunt; held means only that the named attacks did not reproduce."
  ]
}
```

## Revision response and scope

Reviewed the four-file revision-1-to-2 delta first: metadata collision key, retained reviewer probes, M12, and documentation. Canonical caseless matching now normalizes before AND after full case folding. The previous finding `unicode-casefold-renormalization` is resolved; no repeat finding is recorded. TestReviewerGreekCollision keeps both singleton positive controls and all four iota/upsilon file/directory cases, demands the exact path-collision refusal and verifies unchanged manifest bytes. All prior reviewer test function names remain present. M12 restores the former algorithm without deleting collision checks and fails the named behavioral regression with child exit 1.

Current curator-spec main independently read at 15d970111277b1f68b3f7a4ddd63bc339b2c0a3f: managed-home-snapshots §§2, 4, 6, 7 and CIP-0014 implementation/test plan. Protocol matches the producer pin. Fresh origin/main changes have no overlap with either leaf directory. All 116 live files match their immutable candidate blobs. The separate corpus, embedded schemas, released fragments and all code outside the leaf are unchanged by rework.

## AC coverage

**4 of 4 AC rows driven; 4 of 4 mandatory paired-test rows driven.** The exported library API is the declared production seam for this helper leaf. Caller search finds no existing CLI integration; capture, publication and hosted consumers are separate leaves. Tests are captured in the immutable CR tree and remain uncommitted in the managed Story checkout as required.

| AC | Production seam | Named driving tests |
|---|---|---|
| Closed codecs and hash domains | EncodeManifest, DecodeManifest, EncodeDescriptor, DecodeDescriptor, FragmentDigest, Digest, Verify | TestPinnedSchemaCases, TestMetadataFramingAndHashDomains, TestLaunchHashIncludesLF, TestTypedEncodingRejectsInvalidUTF8, TestPinnedTrees |
| Inventory, bytes, modes, collisions, references, bindings | Verify → validateMetadata, verifyEntries, verifyFile | TestActualTreeRefusals, TestMetadataNeighbors, TestReviewerGreekCollision, TestMuseDerivedDirectoriesAndRootFact, TestNoFollowSnapshotRoot, TestNoFollowBeforeParentSegment |
| Exact and portable identities, original baseline | Verify, PortableEqual, Verified.ManifestSHA256 | TestPortableIdentity, TestBaselineNeverRecaptured, TestReviewerPortableLeftDrift, TestReviewerRecapture |
| Fixed separate corpus, no manifest-only tree attestation | public codecs, Verify, PortableEqual | TestCorpusPins, TestPinnedSchemaCases, TestPinnedTrees, TestReviewerManifestOnly, TestPinnedSchemaCheckoutBytes |

## Swept surface table

Each result was appended to local sweep.txt before advancing to the next row. Logs are bundled in TASK-261016-1v3gbc_review-local-rev2.log.

| Row | Result | Executed attacks |
|---|---|---|
| Closed metadata and semantic bindings | held | row1.log: pinned closed-schema and semantic cases, framing/hash domains, metadata neighbors, invalid UTF-8, null channel, checkout byte conversion, Greek collision regression; public codecs and Verify |
| Complete actual inventory and independent files | held | row2.log: missing/extra entries, bytes, modes, unreadable content, links/aliases/special files, linked root/ancestor including before parent segment, Muse derived directories, manifest-only directory, metadata hardlink/sticky bits; Verify and PortableEqual |
| Exact versus portable identity | held | row3.log: differing source provenance allowed portably; producer/profile/descriptor/file changes remain binding; left-side drift with unchanged metadata refuses; Verify and PortableEqual |
| Original baseline retention | held | row4.log: changed actual tree and attempted recapture remain refused with original baseline; repeated Verify calls do not rewrite metadata |

Free hunt beyond the table: five additional full-fold classes (sharp-s, sigma, ligature, long-s and Kelvin sign), each as file and directory component. Ten public Verify cases with singleton positive controls and unchanged-manifest assertions pass. No blocking findings. Native filesystem cannot establish coexistence for every folded physical spelling; these attacks establish precise metadata refusal before tree enumeration. The prior Greek bypass was separately tested on a case-sensitive filesystem.

Replay: download TASK-261016-1v3gbc_review-extra-rev2_test.go and TASK-261016-1v3gbc_review-replay-rev2.py into one directory, then run `python3 <artifact-directory>/TASK-261016-1v3gbc_review-replay-rev2.py` from the candidate root. The script copies exact candidate package bytes into scratch and changes only the test corpus lookup before adding the probe. No production source mutation.

## Independent validation

Reviewer executed, all top-level exits 0:

- Row 1: `go test -count=1 -run '^(TestPinnedSchemaCases|TestMetadataFramingAndHashDomains|TestPinnedMetadataSemantics|TestPinnedTrees|TestMetadataNeighbors|TestLaunchHashIncludesLF|TestTypedEncodingRejectsInvalidUTF8|TestPinnedSchemaCheckoutBytes|TestReviewerGreekCollision|TestReviewerNullChannel)$' -v ./internal/envsnapshot`
- Row 2: `go test -count=1 -run '^(TestActualTreeRefusals|TestNoFollowSnapshotRoot|TestNoFollowBeforeParentSegment|TestMuseDerivedDirectoriesAndRootFact|TestReviewerManifestOnly|TestReviewerFreeHunt)$' -v ./internal/envsnapshot`
- Row 3: `go test -count=1 -run '^(TestPortableIdentity|TestReviewerPortableLeftDrift)$' -v ./internal/envsnapshot`
- Row 4: `go test -count=1 -run '^(TestBaselineNeverRecaptured|TestReviewerRecapture)$' -v ./internal/envsnapshot`
- `TMPDIR="$CASE_TMP" go test -count=1 -run '^TestReviewerGreekCollision$' -v ./internal/envsnapshot`
- `TMPDIR="$CASE_TMP" go test -tags snapshotcontrols -count=1 -run '^TestNarrowingControls$' -v ./internal/envsnapshot`
- `go test -count=1 -race ./internal/envsnapshot`
- `python3 .temp/review-snapid-r2/TASK-261016-1v3gbc_review-replay-rev2.py`

CASE_TMP was a private case-sensitive HFS+ image, detached after the targeted regression and controls. Native environment: Go 1.25.5, darwin/arm64. 13 of 13 narrowing controls killed, zero survivors. Child failures are expected exit 1, not compilation/setup failures. Includes M5 mode/extra, M6 saved-manifest trust, M12 post-fold normalization, and token-preserving M11 checkout behavior. The general race suite used the native filesystem; no broad HFS+ suite success is claimed.

Reviewer independently executed `sh scripts/remote-gate.sh`: exit 0. Workflow https://github.com/relux-works/curator/actions/runs/38869438224 completed success on signed snapshot d313a63f9d905ebe0a496e6de74f72d95649404e, whose tree equals the CR candidate. Twenty jobs passed; two conditional jobs were skipped, not counted as executed. This includes Test on Linux/macOS/Windows, Race on Linux/macOS, Lint, interop, naming, gate self-tests and toolchain drivers. Final provider evidence and sanitized gate log: TASK-261016-1v3gbc_review-hosted-rev2.json and TASK-261016-1v3gbc_review-gate-rev2.log. Final byte comparison again matched all 116 candidate files. No previous producer green is substituted for this execution.

## Bounds and task-scoped logbook

This accepts the reusable metadata/actual-tree helper at its declared platform boundary, not full materialize or runtime delivery. Windows actual-tree support remains unimplemented and refuses. No Windows ACL, reparse, successful tree verification or runtime claim is made. Ownership/ACL admission, capture consistency, authorization, immutable external baseline/overlay retention, and ongoing protection against hostile concurrent mutation belong to consuming leaves. Self-minted digests are integrity metadata, not plan authorization. No baseline recapture/write API is exposed.

No product code, Story branch, checkpoint or index was modified. The gate's task-local Git adapter preserved configured human author identity, signed and verified a detached CI snapshot, and left the repository gate and CI source unchanged. No unsigned snapshot was published. All review evidence was written inside the worktree and attached using board CRUD. Early invalid board projections were corrected, never treated as absence. No subagents were used. The producer's reported cross-chain board attribution warning was read; it is not independently attributed to producer writes and does not change candidate source evidence.
