# TASK-260728-26e3n2: run-linux-rust-driver-qualification

## Description
Qualify rust-v1 and rust-repository-v1 natively on the dedicated Linux host after shared cross-manager suites are accepted.

## Scope
Pinned Linux toolchain and guidance, clean bootstrap, local vendored and external locked fixtures, full positive/negative corpus, rollback/concurrency/PATH gates, evidence and claim update only after review.

## Acceptance Criteria
Both Rust source modes reproduce the pinned suite without skips or fabricated claims; failures block only the affected Linux tuple; immutable host/toolchain evidence is attached.
