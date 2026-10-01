# BUG-261001-2n70px: install-skips-non-git-project-root

## Description
curator install at a non-git project root (product folder of repos) skips project skills with exit 0 and installs nothing: the gitignore hygiene gate runs git check-ignore, gets 128, and skips. Second-operator (op2-product) blocker G2.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
Non-git project root installs declared project skills via curator install (bytes present), git-repo hygiene unchanged, non-128 git errors fail closed; entry-point rows + mutant; docs line.
