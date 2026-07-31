# TASK-260728-3u1nho: run-linux-kotlin-driver-pair-qualification

## Description
Qualify the selected local and external Kotlin driver pair on the dedicated Linux host after shared suites are accepted.

## Scope
Pinned compiler/runtime, clean bootstrap, local and external fixtures, full corpus, rollback/concurrency/PATH gates and immutable evidence.

## Acceptance Criteria
Both source modes pass without skips or fabricated claims; failures block only affected Linux tuples and retain exact host/toolchain evidence.
