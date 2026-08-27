# TASK-260720-1u7hes implementation evidence

## Provenance

Created the isolated curator-spec worktree at /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1u7hes/worktree from origin/main 57c1f56846d221ecc55786bd3c2467ec32f11730. Imported the accepted TASK-260720-cw39jh product tree byte-for-byte from /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-cw39jh/worktree, excluding ignored caches, binaries, alternate indexes, virtual environments, board config, and .temp content. No files were staged or committed.

The task-only delta against the accepted predecessor is limited to tools/validate.py, tools/release_gate.py, tools/test_release_gate.py, tools/test_validate.py, and tools/generate-vectors/main_test.go.

## Behavior

The validator now pins the exact 35-schema rc.4 inventory, the ordered 129-case positive and negative index, the complete 189-file suite inventory, canonical manifest entry shapes and hashes, all build-driver case identities, all compiled manager-lifecycle coverage groups, and byte-identical shared lifecycle/build-driver identity. Missing schema references fail as validation errors.

The release gate now requires decision 0004, both manifest v6 schemas, receipt v1, marker v2 with schema-6 support, claim v2 at rc.4, claim v1 byte-frozen at rc.3, complete schema cases and manifest inventory, valid current suite hashes, and the build-driver plus lifecycle vectors. Optional --claim evidence is schema-validated and must name the SHA-256 of the current manifest bytes. The current rc.4 suite identity is sha256:70cd274150d629f16ca04f2073a8eae5a0f3fff4584f905d807775040af21eae; the real rc.3 identity sha256:7951cda1711d34d2a9dd9a873cf9d537c41ca4e9527e94f138f38743610a379e is rejected as rc.4 evidence.

Negative tests remove every rc.4 schema and every generated rc.4 schema case individually, and cover removed or renamed required artifacts, missing manifest entries, stale hashes, unresolved references, renamed lifecycle coverage, claim v1 and v2 version mismatch, duplicate claim keys, and rc.3 suite evidence.

## Verification

- python3 -B -m unittest discover -s tools -p test_*.py: 27 passed
- go test ./tools/...: passed
- go vet ./tools/...: passed
- gofmt -d tools/generate-vectors/main_test.go: clean
- make validate: passed; 35 schemas and 189 vector files validated
- make regenerate-check: passed with a disposable alternate Git index seeded from the accepted uncommitted conformance baseline; the real index was untouched
- git diff --check: passed
- post-regeneration manifest SHA-256 remained 70cd274150d629f16ca04f2073a8eae5a0f3fff4584f905d807775040af21eae

System Python lacks jsonschema, so the exact Python and make commands ran with PATH prefixed by the predecessor task-local environment installed from pinned requirements-dev.txt.