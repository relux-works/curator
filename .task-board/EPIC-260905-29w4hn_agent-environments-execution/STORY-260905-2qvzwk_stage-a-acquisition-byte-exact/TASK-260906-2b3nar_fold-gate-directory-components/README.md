# TASK-260906-2b3nar: fold-gate-directory-components

## Description
Follow-up N1 from the acquisition cycle-2 review (TASK-260905-3r30t1_review-verdict-rev5.md): internal/gitops planWrites keys the platform-path collision gate on the folded FULL path, so two tree entries whose directory components fold together but whose basenames differ (Dir/x.txt + dir/y.txt) are admitted and land in one physical directory on a case-folding filesystem. Pre-existing (the git archive path had it too), not a regression. Fix: fold per path component, or track folded directory prefixes, and refuse the collision with the existing duplicate-platform-path class; add a test that extracts such a tree into a case-insensitive destination and asserts the refusal. Schedule with implementation stage (a) or (b).

## Scope
(define task scope)

## Acceptance Criteria
Directory-component folds are refused with the duplicate platform path diagnostic; a test proves it on a case-folding destination; go build/vet/test green; PR reviewed and landed by fast-forward.
