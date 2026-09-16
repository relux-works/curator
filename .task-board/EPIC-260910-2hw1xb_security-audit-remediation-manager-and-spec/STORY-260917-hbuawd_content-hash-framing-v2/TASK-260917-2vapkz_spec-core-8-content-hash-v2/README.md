# TASK-260917-2vapkz: spec-core-8-content-hash-v2

## Description
curator-spec: revise core.md section 8 to curator-content-v2 framing (domain prefix curator-content-v2 plus 0x00, then per file F || uint64be(len(path)) || path || uint64be(len(bytes)) || bytes), version the identity (prefix or hash_version member wherever the hash is carried), add the colliding-tree and empty-tree vectors, amend registry.md so a record matches only with an equal framing version, and record the interim NUL-opaque rule for v1 readers.

## Scope
protocol/core.md section 8, protocol/registry.md, schemas, conformance vectors

## Acceptance Criteria
Revision merged with vectors; issue #59 closed by the landing commit
