# BUG-260731-38dz6m: curator-windows-transaction-and-fingerprint-residuals

## Description
Curator Windows Test lane retains six failures after the accepted PR 12 scope and PR 13 owned scope: five internal/transaction cases and internal/godriver TestFingerprintReportsUnreadableDirectoryIdentically. They were previously mapped to BUG-260731-lepevi, but that completed bug owned only Linux lint/inventory behavior and did not fix these native Windows failures. Reproduce from raw windows-latest evidence on the current combined main/PR13 base, fix platform behavior without skips or ledger relaxation, publish a signed Curator PR, obtain independent Opus review, and land autonomously.

## Scope
Only internal/transaction and the single internal/godriver fingerprint case plus focused tests/evidence needed for these six Windows residuals. Do not absorb PR13 GOROOT/buildcache/managerlock/staging/globalbins scope or unrelated Curator CI work. Claude Opus 5 developer/reviewer only; no tags/releases.

## Acceptance Criteria
On a native windows-latest runner the five internal/transaction failures and TestFingerprintReportsUnreadableDirectoryIdentically all execute and pass; no new skip, tolerance, platform exclusion, or broad suppression is introduced; Linux and macOS gates remain green; the signed PR is independently accepted and landed to main.
