# Review verdict — Design CIP-0007: manager-provisioned CLI tools

Task: TASK-261005-39d5lx. Change request: CR-TASK-261005-39d5lx-2, revision 2.
Verdict: **changes_requested**. Final route: **analysis**, because this is research/documentation rework.
Date: 2026-10-05. Reviewer run: RUN-261005-5e340e.

Base: `6ec4158d96200fca47f7d6b999f2fa3da0053274`.
Candidate tree: `662f3316dd92af74ae3e71ac2f7a8931cbf065c8`.
All three changed worktree files match the exact candidate bytes. Fresh
`git ls-remote origin refs/heads/main` returned the base OID. The delta is
762 additions across the three authorized documentation paths. No repository
files were modified by this reviewer; no commit or commit_ack was supplied.
The run goal was queried before recording the verdict: not goal-bound.

## Numbered finding

1. **Medium — The research note reintroduces the forbidden names and makes its own scan result false.**
   At `.research/261005_manager-provisioned-cli-tools.md:155`, the note includes
   both prohibited vendor tokens inside the published scan expression, while
   reporting zero matches over the CIP and the note. Independent case-insensitive
   occurrence scanning returns **2 occurrences on 1 line in 1 file**. The CIP
   itself now has zero occurrences. The same zero-count assertion at lines
   100–102 is therefore also false for the final revision.
   This repeats revision 1 finding 1's mechanism: forbidden names in the public
   deliverable coupled with a false absence claim; the location moved from the
   CIP examples into the evidence note. **repeat-of: rev1/F1** (F1 refers to
   numbered finding 1 in the legacy prose verdict; no structured predecessor
   finding array exists).

   Remove the literal named-token expression from the public note, retain only
   generic scan labels and actual counts, and scan the **final saved bytes** of
   both documents after writing the report. Record pre-correction versus final
   counts accurately. Keep the change within the three authorized docs paths.

   Because the same mechanism recurred, carry a named verification check
   `final-public-doc-name-scan` in the next evidence packet. Its negative
   control must place a prohibited token only in the evidence document and
   detect it; narrowing the scan to the CIP alone must miss that control.
   This can be a bounded ephemeral check recorded in the evidence; it requires
   no normative edits or separate harness task. Rerun validation and hand off
   the new candidate.

## Previous findings swept

| Revision 1 finding | Revision 2 assessment |
| --- | --- |
| 1 — Names and false scan | Still broken by the same mechanism; finding above. |
| 2 — Federation-wide revocation | Addressed: CIP lines 110–126 and Security Registry compromise define a complete revocation scan, exact match/withdrawal keys, later-layer deny precedence, allow-versus-revocation precedence, and refusal for an unreachable trusted layer. Test plan names all three negative cases. |
| 3 — Expected publisher/subject | Addressed: Option C and Security Mirror and resolver compromise bind tool identity to operator-approved publisher/subject and digest, revoke removed identities, and describe reviewed signed resolver output. Signer substitution and wrong-subject negatives are named. |
| 4 — Downgrade and replay | Addressed: tools.refresh and Security separate semantic downgrade from registry §5 high-water, require current authorization on store hits, remove key trust per §2.1, bound offline/air-gap evidence by §§5/8, and state the availability residual. Relevant negatives are named. |
| 5 — Manifest hash and versioning | Addressed: Compatibility correctly hashes the declaring Skillfile, distinguishes locked member identity and source_snapshot_changed, and proposes skill-manifest schema 9, lock-v2 and new markers while retaining historical lanes. |
| 6 — Runtime exposure | Addressed: B has a provisioned-only binding with collision refusal; core §4.1.1 enforced PATH remains limited to interpreter and declared exec names. C alone adds system probes and refined scope composition. Exposure negatives and corrected issue/question references are present. |

## Complete review surface sweep

Coverage: **7/7 requested review areas assessed**, **6/7 held**, **1/7 broken**.
Held design rows mean the named static counterexamples were answered by the
Draft; they do not attest runtime execution or absence of other defects.

| Area | Result and evidence |
| --- | --- |
| Template/process | Held: 10/10 top-level sections match TEMPLATE order, 5/5 metadata fields, both Design subsections, Draft status and exactly one linked Draft index row. CIP-0001 evidence limitations are stated. |
| Options/recommendation | Held: A/B/C have mechanism, pros, cons and security; B's smaller policy surface and curation/availability costs support the recommendation rather than relying only on triage. |
| Trust/threats | Held for this design review: evaluated author URL, later-layer revocation/unreachability, unexpected valid signer, wrong subject, removed key, revoked store hit, expired air-gap evidence and semantic downgrade counterexamples against the revised proposal and registry §§2/2.1/4/5/7/8. Proposed behavior is distinguished from current specification mechanics. |
| Schema/lock/transaction/conformance/audit | Held: cross-checked core §4/§4.1.1, Skillfile sources §§1/3/4, closed lock-v1 schema, manager §§2.1/2.5 and CIP-0005 Design §3 (lines 96–114). Protected-boundary/preimage revalidation and mutable rollback versus immutable retention are named. Conformance cases are proposed, not claimed executed. |
| Operator decisions | Held: ten numbered answerable questions with recommended answers, including registry owner/service, signing pins, default policy, resolver location, manifest/identity/distribution, probes, upstream pins and downgrade/air-gap operation. Adoption questions do not block producing the Draft. |
| Scope/names/privacy/evidence | Broken: finding 1. Exact delta has no protocol, profile, schema, conformance, CHANGELOG or LOGBOOK edits. Issue-organization token count: 0; personal path/private-host token count: 0. Existing README context is excluded from the added-text name scope. |
| Repository validator | Held: independently rerun, exit 0, 73 schemas and 1,294 vector files. The validator does not inspect CIP prose. |

## Verification and bounds

Read the full previous verdict, candidate CIP and evidence, process/template,
related Draft structure and the cited current specification sections. Read
issue #108 in full through gh, including its empty comments list.

Independent reviewer command:
`/tmp/cip0007-venv/bin/python -B tools/validate.py` — **exit 0**,
`validated 73 schemas and 1294 vector files`.
This uses the pre-existing producer virtual environment; it is a reviewer
rerun, not reuse of its attached success claim. `git diff --check BASE CANDIDATE`
also exited 0.

The reproducible final-document scan returned 2 occurrences across both
documents, versus 0 when narrowed to the CIP alone. Issue-token scan returned
0 and personal path/private-host scan returned 0. These are bounded known-token
scans backed by manual reading, not a universal named-entity detector.

No provisioning implementation, OS execution, signature exploit or runtime
mutant was run: this leaf delivers a Draft and research evidence only. Future
implementation test plans remain unexecuted proposals. The complete candidate
and hash identity were verified rather than inferred from validation success.

The important repeated evidence anomaly is persisted in this verdict and board
notes. LOGBOOK remains unchanged under the explicit task prohibition. No
external blocker or human-only decision is needed for this correction.

```verdict-findings
{
  "findings": [
    {
      "id": "F1",
      "row": "Scope/names/privacy/evidence",
      "invariant": "Public deliverables contain no prohibited organization/vendor names and reported final-document scan counts match the saved candidate bytes.",
      "mechanism": ".research/261005_manager-provisioned-cli-tools.md:155 publishes both prohibited vendor tokens in the scan expression while asserting zero matches over that same document.",
      "reproductions": [
        {
          "test_file": ".research/261005_manager-provisioned-cli-tools.md",
          "command": "rg -oi '\\x67itlab|\\x67rafana' cips/CIP-0007-manager-provisioned-cli-tools.md .research/261005_manager-provisioned-cli-tools.md | wc -l",
          "expected_failure": "No-name acceptance requires 0 occurrences; the exact candidate returns 2 (one evidence-note line). Narrowing to the CIP alone returns 0 and hides the violation.",
          "pinned_blobs": [
            "sha256:09ab2d365116d5334cc91b9f1a969afbf8639e97341e8c4b9c075dc29d9bb9c0",
            "sha256:b0a4f032a5f776e808a88d44efb1cc64a21bca686e137b999bd83756fb9945a7"
          ]
        }
      ],
      "severity": "bypass",
      "repeat-of": "rev1/F1"
    }
  ],
  "notes": [],
  "surface_results": [
    {
      "row": "Template/process",
      "result": "held"
    },
    {
      "row": "Options/recommendation",
      "result": "held"
    },
    {
      "row": "Trust/threats",
      "result": "held"
    },
    {
      "row": "Schema/lock/transaction/conformance/audit",
      "result": "held"
    },
    {
      "row": "Operator decisions",
      "result": "held"
    },
    {
      "row": "Scope/names/privacy/evidence",
      "result": "broken"
    },
    {
      "row": "Repository validator",
      "result": "held"
    }
  ],
  "free_hunt": []
}
```
