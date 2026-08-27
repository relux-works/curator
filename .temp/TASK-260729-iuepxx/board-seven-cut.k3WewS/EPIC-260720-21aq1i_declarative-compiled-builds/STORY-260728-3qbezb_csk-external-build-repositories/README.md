# STORY-260728-3qbezb: csk-external-build-repositories

## Description
Independently implement the accepted schema-7 external build repository contract in Python cocoaskills/csk after the protocol and Curator reference implementation are accepted. Match wire bytes and security behavior without copying Curator internals, while retaining csk-specific manager-home and activation mechanics.

## Scope
Python csk schema models, clean Git and raw-object admission, independent external audit subject, protected snapshot/artifact caches, mixed receipts/markers, project/global transactions, status/repair/GC, docs, and macOS/Windows native validation. No generic build-command escape hatch.

## Acceptance Criteria
csk independently matches schema-7, receipt-v2, marker-v3, claim-v3, source identity, exact-tag, substitution, audit ordering, cache, rollback, and shim/PATH semantics; inaccessible source and offline protected-snapshot behavior match the spec; schemas 1-6 and existing Python-script/go-v1 flows do not regress; shared and manager-local tests pass on macOS and Windows.
