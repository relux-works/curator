# STORY-260910-1py4f3: boundary-verification-performance

## Description
Finding R2 (Medium): every records/log/snapshot request, every cursor validation, and every /health call recomputes the full boundary (O(n) scan + Merkle rebuild), and all reads serialize on one lock; large logs degrade into a self-DoS.

## Scope
curator-skill-registry store.py/app.py

## Acceptance Criteria
Boundary per log_size is memoized and recomputed only when the head changes; /health serves a cached verdict refreshed by a background verifier
