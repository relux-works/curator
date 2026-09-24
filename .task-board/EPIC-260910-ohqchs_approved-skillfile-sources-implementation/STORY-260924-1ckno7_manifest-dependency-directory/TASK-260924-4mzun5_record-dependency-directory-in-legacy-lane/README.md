# TASK-260924-4mzun5: record-dependency-directory-in-legacy-lane

## Description
Review residual R1 of TASK-260924-1kpw4w: in the schema-1 (legacy) Skillfile lane a schema-9 skill with a subfolder manifest dependency resolves through closure.Build, but the legacy marker (no directory outside v5) and audit.Subject do not record the directory. Decide per spec whether the legacy lane must record it (then record it) or must refuse a schema-9 directory dependency (then refuse with the spec diagnostic); plus spec follow-up vectors for ? and [ globs on the dependency path (R2).

## Scope
(define task scope)

## Acceptance Criteria
legacy lane either records the directory in marker and audit or refuses per spec, with production-entry rows; glob class pinned on the dependency path
