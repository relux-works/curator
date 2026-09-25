# BUG-260923-11jgkt: windows-snapshot-concurrent-get-sharing-violation

## Description
Hosted windows-latest: internal/snapshot TestConcurrentGetAcceptsOneImmutablePublisher fails intermittently on candidates that do not touch it. Gate run 35855672743 (BUG-260922-306v4m republish): snapshot_test.go:61: worker 11 = "", snapshot destination conflicts with immutable commit: authenticate destination: open C:\...\cache\in... . One concurrent reader opens the destination while another worker is publishing it; on Windows an open during the publisher rename/replace can fail with a sharing violation, which the reader reports as a destination conflict. This may be a REAL product defect (a legitimate concurrent get refused on Windows), not only a test race.

## Scope
internal/snapshot publication and destination authentication on Windows; the concurrency test. No semantic weakening of the immutability check.

## Acceptance Criteria
1) root cause proven: test race vs product defect, with the exact Windows error (sharing violation / access denied) captured; 2) if product: concurrent readers on Windows never report a false destination conflict while a publisher completes (bounded retry on sharing-violation only, or a publication protocol that makes the destination openable), with a Windows row; 3) the immutability refusal for a genuinely different destination still fails closed (row); 4) repeated hosted windows runs green; 5) CHANGELOG
