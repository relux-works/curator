# BUG-261017-47xdyt — install-deletes-unowned-store-entries: review revision 1

Verdict: changes_requested
Reviewer-Run: RUN-261018-80ae8a
Change Request: CR-BUG-261017-47xdyt-1, revision 1
Candidate tree: 54bffdc4a8460c329fd5d68cd4dd412458720c4d
Base: 1724b4a52b4c2d163c0a3e6ab7c2ad2bbe54ac0c

The ownership implementation passes this review's bounded attacks. Acceptance is withheld solely because the candidate crossed the mandatory pre-review freshness boundary. This is ordinary producer rework, not a human-only blocker. No code was modified, no branch moved, and no commit created by this reviewer.

## Required fix

1. Refresh/converge the candidate against current protected main through the authorized producer workflow, preserve both README additions and all upstream code, validate the resulting candidate, and publish a fresh handoff. The producer need not change the ownership implementation for any reproduced behavioral defect in this round. Do not call acceptance against this stale revision.

Source of the gate: project-management SKILL.md critical invariant 16 and references/tracked-background-spawn.md, "Pre-review freshness": "A non-empty set is a stale-review boundary even when the hunks look textually separate." The current upstream README addition documents command-help tests and help invocation. Exact-ref fetch and ls-remote agree on 5d8e6b07f7f485557d9bbb065c9692348ca065c2; that commit is timestamped before this run started. Overlap is exactly README.md. The bound is path-level; no textual merge conflict is claimed.

## Findings and notes

```json
{
  "verdict": "changes_requested",
  "findings": [
    {
      "id": "stale-candidate-path-overlap",
      "row": "delivery-freshness",
      "invariant": "Before review, the recorded CR-base/candidate path set must not overlap changes in freshly resolved protected authority; overlap requires converge/producer refresh, validation and a new handoff.",
      "mechanism": "CR base 1724b4a predates protected main 5d8e6b07; both candidate and upstream change README.md. Upstream was committed before this reviewer run started. The workflow explicitly treats path overlap as stale even when hunks are separate. This is a delivery eligibility finding, not a reproduced ownership implementation defect.",
      "reproductions": [
        {
          "test_file": "BUG-261017-47xdyt_freshness-probe.py",
          "command": "python3 BUG-261017-47xdyt_freshness-probe.py",
          "expected_failure": "AssertionError: pre_review_path_overlap: README.md",
          "actual_exit": 1,
          "log": "BUG-261017-47xdyt_review-evidence-rev1.txt: freshness-probe.log"
        }
      ],
      "severity": "regression",
      "repeat-of": "none"
    }
  ],
  "notes": [
    "No product ownership bypass, regression or unstructured failure reproduced.",
    "No catalog surface table was supplied: six explicit AC/review rows form the bounded sweep; no catalog-complete claim.",
    "The producer explicitly bounds v6-unaware substitution, marker.Write-only declaration controls, Windows FIFO/permission substitutions, and hostile concurrent filesystem races. Those are not silently counted as end-to-end coverage.",
    "Full local producer-suite evidence was not replayed: this review ran the named targeted suites and reused exact-candidate hosted success. Two conditional hosted jobs remain skipped, not green execution.",
    "Initial board reads used unsupported element/resources queries; these errors were not treated as absence. Supported get overview/checklist and resource get reads succeeded."
  ]
}
```

## Sweep and measured coverage

# BUG-261017-47xdyt review sweep

Candidate: 54bffdc4a8460c329fd5d68cd4dd412458720c4d.
No surface table or wall-clock budget was supplied. The six supplied review/AC rows are the sweep denominator; no catalog-complete claim. Local attack budget capped at 25 minutes, including a five-minute free hunt. All tests use candidate blobs; no product code changes.

1. AC1 stale preservation: held. TestInstallPreservesUnownedStaleStoreEntries and TestInstallRefusesUnprovenStaleOwnership passed via Project. Three unmanaged shapes named together, exact snapshots and foreign-link resolution; mixed owned state preserved during refusal.
2. AC2 selected destination: held. TestInstallRefusesUnownedSelectedCanonicalDestination and TestInstallOwnedCanonicalPositiveControls passed via Project. Missing, invalid, wrong-name, linked and non-directory evidence refuses; absent install and owned repair succeed.
3. AC3 ownership evidence: held. TestInstallDraftOwnershipCases passed all 21 mapped rows through Project, including permission failure, FIFO, canonical-parent links and changed ownership before publication. Unknown-reader substitution remains an explicit producer bound; no v6-unaware implementation claim. Corpus bytes match spec 3ca4b84f.
4. AC4 positive control: held. Mixed fixture removes owned stale scan after explicit conflict removal; TestRemovedSkillCleanedUp also passed.
5. AC5 callers: held. Global, hybrid, native adapter refusal and retained external-evidence repair tests passed through their production entries. Supporting suites still running when this row was recorded; their results will be appended separately.

6. AC6 gate/mutants: held. All seven narrowing overlays returned 1 at the intended behavioral assertions. Hosted run 39078199935 independently re-read as success on signed 1c85e638, whose tree equals the CR. Twenty jobs succeeded, two conditional jobs skipped. No hosted run was launched by this reviewer.

Supporting marker/staging/transaction/scopes suites exited 0. CLI A/C2 probe exited 0 (individual refusal invocations exited 1), preserving bytes and link resolution. Free hunt: duplicate name, string schema and trailing JSON document all refused through CLI without publication. No product defect reproduced.


Measured: 6 of 6 AC rows driven and held within the stated bounds; 7 of 7 narrowing mutants killed for the intended reason. Additional delivery-freshness row: broken, finding stale-candidate-path-overlap. The local build and diff whitespace check passed. Both vendored corpora match spec 3ca4b84f byte-for-byte. All 20 changed candidate paths match the reviewed tree blobs. The signed hosted snapshot reports signature G and the same tree.

## Direct commands and evidence

- `go test ./internal/install -run '^TestInstall(PreservesUnownedStaleStoreEntries|RefusesUnownedSelectedCanonicalDestination|RefusesUnprovenStaleOwnership|OwnedCanonicalPositiveControls|DraftOwnershipCases|DraftMarkerWriterCases)$' -count=1 -v`: exit 0, preservation-01.log.
- `go test ./internal/install -run '^Test(GlobalCanonicalOwnershipPreservation|HybridCanonicalOwnershipPreservation|DraftUnmanagedAdapterDestinationRefuses|RemovedSkillCleanedUp|DraftExternalExecutionPolicyInvalidMarkerReDerives)$' -count=1 -v`: exit 0, callers-01.log.
- `go test ./internal/marker ./internal/staging ./internal/transaction ./internal/scopes -count=1`: exit 0, support-01.log.
- `go build -o .temp/BUG-261017-47xdyt-review/curator ./cmd/curator`: exit 0.
- `python3 .temp/BUG-261017-47xdyt-review/cli_probe.py`: exit 0; standalone subprocess commands and individual exits in cli-probe.log. Private HOME and local-only Git fixtures; no OS-level network isolation claim.
- Seven exact overlay commands, PIDs and exits are in mutant-results.json; each full log includes its intended named assertion failure. Overlay deltas are included; no production file was mutated.
- Freshness probe: exit 1 at the named README overlap, not a setup failure.

Hosted evidence reused: https://github.com/relux-works/curator/actions/runs/39078199935 . Fresh GitHub API read confirms completed/success on 1c85e638c92edfe1bd87a593b40bb524766c3824. No full hosted replay or local mirror is claimed.

## Task-scoped logbook

Review independently reproduced original A/C2 preservation plus owned-update/stale-removal controls, malformed ownership grammar, cross-scope contracts and seven narrowed gates. The acceptance boundary is procedural freshness: upstream changed a candidate path before reviewer launch. Route once to producer rework; this is not grounds for blocked, and no user decision is requested. All evidence attached before lifecycle transition. The next reviewer should review the refreshed combined tree and any fix-induced delta; this review is not acceptance of that future tree.
