# TASK-260929-jup8re: naming-gate-report-without-content-and-echo-allowlist

## Description
naming-gate.sh prints path:line only (never line content); existing immutable machine-recorded echoes are allowed by an explicit ledger .github/ci/naming-gate-allow.tsv keyed by path + sha256 of the exact line; everything else still fails.

## Scope
(define task scope)

## Acceptance Criteria
1. Gate output never contains line content (path:line only). 2. Historical CI-echo lines with the step-name+TAB+timestamp+./ signature are exempt; all other lines scanned. 3. Selftest rows f-i. 4. Mutants content-printed, signature-relaxed, exemption-removed each killed. 5. Gate exits 0 on an origin/main archive.
