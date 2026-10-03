# STORY-261004-3oognx: inline-audit-2026-10-remediation

## Description
Remediate the four defects confirmed by the 2026-10-03 inline code-level audit of curator (N1–N4). Report: docs/security-audit-2026-10-inline.md (EN) and docs/security-audit-2026-10-inline.ru.md (RU). Audited source: main aeab5334 (unchanged on 68210ecc for every cited file). The probe tests named in the report (TestInlineAudit*) were not preserved; each leaf recreates its own red regression first, through the production entry point.

## Scope
internal/scopes (GC), internal/buildrepo (admission), internal/envprofile (unmanage/restore), cmd/curator CLI regressions

## Acceptance Criteria
(define acceptance criteria)
