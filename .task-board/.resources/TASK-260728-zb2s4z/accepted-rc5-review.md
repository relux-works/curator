# TASK-260728-3b8qym independent review cycle 2

## Verdict

ACCEPTED. No remaining findings.

## Rework closure

1. CCJ-1 oracles: independent jq canonicalization recomputed SHA-256(CCJ-1(receipt.input)) as sha256:012564909df8f333004eb5aec867210d8973c74d9b71948e1f4fdb0d00c76559 and SHA-256(CCJ-1(full receipt)) as sha256:d86ca882ff336480a277010e886bea46767bf9a23d928fe3b43ee1009c3bd0ed. Receipt, mixed marker, mixed plan, schema examples, Python validation, and Go validation agree. Targeted false-cache-key and false-receipt-hash mutation tests passed in both harnesses.
2. Whole-snapshot ordering: executable external-source-dry-run and external-audit-only vectors are non-mutating, claim source and audit coverage, and enforce exact acquisition -> graph/object proof -> LFS scan -> immutable snapshot -> whole-snapshot validation -> source digest -> descriptor validation -> independent audit before cache or compiler. The same checker covers cache hit, cache miss, repair, dry-run, and audit-only; syntax-only claims none of source, audit, cache, or mutation.
3. Pack negatives: the checksum case materializes base plus final-index-byte XOR and differs at exactly byte offset 1071; full PACK and index-v2 header, fanout, trailers, filenames, embedded pack checksum, and index checksum parsing proves the intended sole fault. The family case preserves exact valid SHA-1 bytes under a SHA-256 declaration and fails SHA-256 validation with the stable build_repository_local_object_format_unsupported error.

## Release qualification

- Assigned HEAD is exactly 57c1f56846d221ecc55786bd3c2467ec32f11730; index is unstaged and there are zero commits after the pin. Content matches the disposable clean probe byte-for-byte. Reviewer regeneration changed generated-file mtimes only, not bytes.
- Two consecutive clean make regenerate-check runs passed. Clean make release-check VERSION=1.0.0-rc.5 passed at disposable probe commit baed6d17303344c8c48dfdbb9cc6f6681aab1e1d.
- Full validation passed: 42 schemas, 411 manifest files, 18 Python tests, all Go tool tests, go vet, gofmt, compileall, go build, git diff --check, and tracked probe cleanliness.
- Independent manifest audit found 411 unique, complete, non-self-referential entries with exact file hashes. Manifest pin is sha256:33fd7aed900ec4f9f9c72be6298823a16674116b18e3b8efdd2d147574dba2b8. Release metadata SHA-256 is 1f4f6fcc7c57f26f050d730f7b84091cf3f58970871bd7f511bea03f4b2b8f31.
- Exact file SHA-256 values: receipt 28d3340295731b4271ceb002ef3fc063bbf187237f4012e2269b0c702bf78622; mixed marker 0088cb9536eaacea2c087efeb70a5bc453c923020c5c01d3a87146f6e85773e3; pack/index 09768bac46eb9966e76baf7fbe613ed87a88dd665962f4d3c18f20dc6a98791c; lifecycle 175d709f183775b22ed3db27bc923ca78d6394d93a13096b5d6890c509aab072. Largest fixture is 18377 bytes, below the 65536-byte ceiling.
- Exact downstream protocol pin is the manifest pin above through CURATOR_CONFORMANCE_ROOT; committed_release_pin_advanced remains false and the obsolete sha256:30a64ed0... pin is absent.
- Candidate claim-v3 claims are empty; macOS and Windows remain pending downstream native evidence; Linux is explicitly excluded until TASK-260728-1skseh. No manager implementation, generic driver, real remote contact, fabricated platform evidence, or schema rollback entered scope.
- 196 protected schemas 1-6, go-v1, receipt-v1, marker-v1/v2, claim-v1/v2, and generated legacy cases match the accepted predecessor byte-for-byte.

The rc.5 shared conformance corpus and candidate release layer satisfy the task acceptance criteria and fit the specification architecture.