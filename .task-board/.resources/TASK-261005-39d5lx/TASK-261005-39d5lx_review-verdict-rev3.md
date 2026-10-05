# Review verdict — Design CIP-0007: manager-provisioned CLI tools

Task: TASK-261005-39d5lx — Design CIP-0007: manager-provisioned CLI tools.
Change request: CR-TASK-261005-39d5lx-3, revision 3.
Verdict: **accepted**. Required route: **integrating**, via accept_cr.
Review date: 2026-10-05. Reviewer run: RUN-261005-2cbde4.

Base: `6ec4158d96200fca47f7d6b999f2fa3da0053274`.
Candidate tree: `d8e4a4ce4030dcb4f0903f30e6177dd1396abf50`.
Board patch SHA-256: `1439a67d60b2c62bc8f398f0680d77aa9b0be92c90a8d1e5033c790fa477854b` (independently verified).
All three saved worktree documents match the candidate byte-for-byte.
Fresh `git ls-remote origin refs/heads/main` returned the base OID.
The base-to-candidate delta contains only the three authorized documentation
paths (796 added lines). Revision 2 to 3 changes only the evidence note
(50 insertions, 16 deletions). No repository file was modified by this reviewer;
no commit or commit_ack was supplied. The reviewer goal was queried: this run
is not goal-bound. Acceptance approves this research delivery, not adoption
of the Draft or a provisioning implementation.

## Previous findings verified

- **rev2/F1 (repeat-of rev1/F1): fixed.** The public research note no longer
  contains the prohibited vendor tokens. Its Fact-checks (lines 102–116) and
  Validation (lines 174–193) accurately distinguish the previous 2 occurrences
  from the final 0. Independent scanning of revision 2's saved note reproduced
  2 occurrences; scanning revision 3's saved CIP and note returned 0.
  The named `final-public-doc-name-scan` and note-only negative control are
  recorded in the candidate. Both were independently rerun by this reviewer.
- **rev1/F2–F6: remain fixed.** CIP bytes are unchanged from revision 2.
  Independently cross-checked the cited current specification and the relevant
  counterexamples, rather than accepting validator success as design evidence:
  federation-wide revocation before selection; upstream publisher/subject
  authorization; snapshot rollback versus semantic downgrade and current replay
  authorization; declaring-Skillfile hash versus member snapshot identity;
  versioned closed-schema extensions; and B's minimal binding preserving
  enforced execution controls. Locations and grounds are in the sweep below.

## Complete review surface sweep

Coverage: **7/7 review areas assessed; 7/7 held; 0/7 broken**.
No new findings. Static design review answers the listed counterexamples;
it does not establish that proposed runtime controls have shipped or passed.

| Area | Result and evidence |
| --- | --- |
| Template/process | Held: 10/10 required top-level sections match TEMPLATE order, 5/5 metadata fields, 2/2 required Design subsections, Draft metadata and exactly one linked Draft index row. CIP-0001 evidence limits are stated at CIP lines 53–55. Relative document links resolve. |
| Options/recommendation | Held: A/B/C compare mechanism, pros, cons and security (CIP lines 72–195). B is justified by smaller policy surface and explicit curation/availability costs. C is gated on operator decisions. Goals and non-goals are stated. |
| Trust/threats | Held: author URL refusal; later-layer revocation and unavailable layer; allow override versus revocation; valid unexpected signer and wrong subject; removed key; revoked store hit; expired air-gap evidence; and lower tool release in a newer snapshot are explicitly addressed (CIP lines 113–145, 208–225, 241–362). Cross-checked registry §§2/2.1/4/5/7/8. New tool semantics are proposals, distinct from cited current rules. |
| Schema/lock/transaction/conformance/audit | Held: core §4 and §4.1.1; Skillfile sources §§1/3/4 (especially lines 161–187); closed root/member lock-v1 schema; manager §§2.1/2.5 (lines 186–189, 519–578); CIP-0005 Design §3 (lines 96–114). CIP Compatibility lines 377–405 correctly separates collection/member identity and proposes schema 9, lock-v2 and a new marker version. Transaction sketch lines 453–458 preserves protected-boundary/preimage revalidation and immutable retention versus mutable rollback. Conformance and production-entry negatives/mutants are proposed at lines 469–482 and 507–540, with no claim they ran. |
| Operator decisions | Held: ten numbered, answerable questions with recommended answers (CIP lines 544–594), including registry owner/service, signing pins, default policy, resolver location, manifest/identity/distribution, probes, upstream identity policy and downgrade/air-gap reauthorization. These are adoption decisions, not blockers to a Draft research delivery. |
| Scope/names/privacy/evidence | Held: exact delta touches only CIP, research note and one README index row. No normative, CHANGELOG or LOGBOOK edits. Known vendor-name count 0 across 2/2 final documents; issue-organization token count 0 and personal-path/private-host count 0 across 3/3 files. Manual reading found no prohibited attribution. Existing unrelated README context is outside the added-text name scope. Final count claims and pre-correction counts reproduce. |
| Repository validator | Held: independent reviewer rerun exited 0, validating 73 schemas and 1,294 vector files. Validator does not inspect CIP prose; this is repository validation, not a runtime provisioning assurance. |

## Independent verification and bounds

Read both prior verdicts in full, the exact candidate CIP and note, the process
and template, related Draft structure, and the cited current specification
sections. Retrieved issue #108's complete body and comments through gh;
comments count was 0. Bare non-JSON issue output was empty, so explicit JSON
was used rather than treating that empty read as absence.

Reviewer rerun: `/tmp/cip0007-venv/bin/python -B tools/validate.py` — **exit 0**,
`validated 73 schemas and 1294 vector files`. The pre-existing venv was used;
this is a fresh execution, not reuse of the producer's reported success.
`git diff --check BASE CANDIDATE` — **exit 0**.

Named check: **final-public-doc-name-scan — passed**.
Patterns were derived from the issue body, written only into ephemeral 0600
files outside the repository, and supplied to rg via `-f`; prohibited token
values were never placed on argv or in a repository file by this reviewer.
The independent case-insensitive occurrence scans returned:

| Check | Observed/expected |
| --- | --- |
| Final vendor-name scan: CIP and evidence note | 0/0 occurrences; coverage 2/2 documents |
| Previous revision evidence note | 2/2 occurrences |
| Negative control: token only in temporary evidence-note copy, complete scan | 1/1 detected occurrence |
| Narrowed negative control: CIP only | 0/0 detected occurrences; deliberately misses the note violation |
| Issue-organization token: both documents and README | 0/0 occurrences; coverage 3/3 files |
| Personal-path and current private-host tokens: same files | 0/0 occurrences; coverage 3/3 files |

The note-only control demonstrates the actual scan entry point detects a
violation outside the CIP and that narrowing the scope hides it. Patterns and
control copies were removed after execution. These scans cover known names
and path/host tokens, complemented by manual reading; they are not universal
named-entity detection. No provisioning runtime, platform execution, signature
exploit or runtime mutant was run or claimed. The producer's runtime Test plan
remains a future implementation proposal.

Candidate document SHA-256 identities:

- CIP: `b0a4f032a5f776e808a88d44efb1cc64a21bca686e137b999bd83756fb9945a7`.
- Evidence note: `dc134e2c81173457aaaae44e7f777a9a0db2b679dd2f32b0a7e94daa6346fe8a`.
- README: `36739bc0d663dee0a787ac85c86ccfdbd09a581501cf7c5849b2ceb48ef36b00`.

The repeated evidence anomaly is now resolved and recorded here on the board.
LOGBOOK remains unchanged under the explicit task prohibition. No external
blocker or human-only decision is needed to accept this Draft delivery.

```verdict-findings
{
  "findings": [],
  "notes": [],
  "surface_results": [
    {"row": "Template/process", "result": "held"},
    {"row": "Options/recommendation", "result": "held"},
    {"row": "Trust/threats", "result": "held"},
    {"row": "Schema/lock/transaction/conformance/audit", "result": "held"},
    {"row": "Operator decisions", "result": "held"},
    {"row": "Scope/names/privacy/evidence", "result": "held"},
    {"row": "Repository validator", "result": "held"}
  ],
  "free_hunt": []
}
```
