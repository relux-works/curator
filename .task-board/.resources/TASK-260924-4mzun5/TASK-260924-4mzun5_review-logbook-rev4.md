TASK-260924-4mzun5 — record-dependency-directory-in-legacy-lane: review logbook, revision 4

The status declaration-binding repair passed independent project/global CLI probes, including read-only marker preservation and reinstall recovery. The moved-tag repair uses the presence of any live tag at the recorded commit as evidence that the installed tag is unchanged. Hosted probes reproduced silent strict-tag bypass with a lightweight or annotated archive tag on the prior commit (4/4), while no-alias controls refused (2/2). An explicit change to unmoved v2 after deleting old v1 was also falsely refused (2/2; P2 BUG note).

This is a source-policy defect, not an external blocker or unresolved human decision. Rework stays in this task and must preserve exact previous declaration evidence instead of inferring it from unrelated tags. All targeted tests and mutations ran on hosted runners; local checks were compile-only. The full CR gate is green but does not cover these missing tag-state combinations.

Evidence: TASK-260924-4mzun5_review-evidence-rev4.log and review-verdict-rev4.md; hosted run https://github.com/relux-works/curator/actions/runs/37888437685. No repository LOGBOOK or CHANGELOG edits were made by the reviewer.
