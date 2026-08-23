# External build repositories v5 — independent review verdict

Verdict: **changes requested**. Route `TASK-260720-1nvomm` to `analysis` for one bounded architecture correction and a fresh reviewer cycle.

Reviewed artifact: `TASK-260720-1nvomm_external-build-repositories-architecture-v5.md`, SHA-256 `1ac3715cbc13ea85c71bac1b4ae427b6647dab8956ce5c80839744ffd9b716b4`. The active run `RUN-260727-12bce4` is not goal-bound.

## Accepted portions

- V5 closes both v4 findings. The pinned `git-lfs-pointer-parser-v3.7.1` algorithm matches the official 3.7.1 decoder/encoder behavior for aliases, the less-than-1024 cutoff, Unicode trimming and ScanLines behavior, one-digit prefix-only extension matching, arbitrary accepted extension suffix bytes, exact-key overwrite, distinct-priority rejection, signed-int64 sizes, and non-empty size-zero classification.
- Independent Git LFS 3.7.1 smokes returned strict exit 0 for the current canonical pointer and the lowercase/dot/dash, underscore, uppercase, and punctuation-suffix extension family; exit 2 for exact duplicate-key and non-empty size-zero noncanonical forms; and exit 1 for distinct duplicate priority, punctuation-start, and uppercase-OID near misses. The packet detector classifies each consistently.
- The common commit/tag parser now fixes root-tree and structural-header order/uniqueness, continuation and separator behavior, object-ID widths, target-type checks, tag cycles/depth, and one stable semantic error across network/private and local/inert stores. Its signed and extra-header fixtures are compatible with the cited Git object/signature model.
- Previously accepted schema-7 separation, logical-target/output boundary, declared/effective identity, raw-object and pack/index controls, audit-before-cache/compiler, receipt-v2/marker-v3, status/repair/GC, signing, future-driver, and board-impact decisions remain intact except for the tag-verification gap below.

## Required change

### A declared optional tag is never verified when the locked-OID fetch succeeds

Section 3.3 normatively says that when `tag` is present its exact ref MUST peel to the locked commit and a mismatch is `build_repository_ref_moved`. Section 6.3 instead makes the exact-tag fetch only a fallback after the full locked-OID fetch fails. On the normal successful locked-OID path the fixed fetch uses `--no-tags` and writes only `refs/curator/locked`; the process graph contains no second ref query or tag fetch. Manager code therefore has no tag ref or tag object from which to prove the assertion.

Independent Git 2.50.1 smoke created an annotated `v1.4.0`, then used the packet full-OID refspec with `--no-tags`, empty refmap, and no FETCH_HEAD. Fetch succeeded, `show-ref` contained only the locked commit at `refs/curator/locked`, and `refs/tags/v1.4.0` was absent. The same successful path would also be unable to distinguish a moved or missing tag while receipt v2 records the unverified declaration.

This contradicts the stated tag assertion, moved-ref failure, audit subject, and cache identity. It is not repaired by the immutable commit lock: the lock protects compiler bytes, while the optional tag is separately promised as a verified human-facing assertion.

Required correction:

1. Define one exact verification flow for every declaration containing `tag`, regardless of whether the server accepts a full-OID fetch. The narrowest option is to fetch the exact tag ref into a fresh private repository, parse/recompute its tag chain, require the terminal commit to equal the lock, and use that exact fetch as acquisition when `tag` is present. An alternative second fresh exact-tag verification after full-OID acquisition is acceptable if its argv, state, and failure ordering are equally fixed.
2. Reserve the full-OID-only flow for declarations without `tag`, or explicitly specify and validate both required attempts when tag and OID acquisition are separated.
3. Add conformance vectors for matching, moved, and missing tags on servers that do and do not permit direct full-OID fetches. All mismatched/missing assertions must fail before audit success, artifact-cache lookup, or compiler execution; no branch/all-tags fallback is allowed.
4. Keep tag verification syntax-only offline behavior consistent with section 7: a check that cannot obtain the exact source may warn without a claim, while install, repair, and coverage-claiming audit fail before mutation.

This is ordinary bounded architecture rework, not a human-only decision or external blocker.

## Independent validation evidence

- All 12 fenced JSON blocks parsed.
- `/tmp/curator-review-venv.AztOCZ/bin/python tools/validate.py` passed: `validated 30 schemas and 93 vector files`.
- Full pinned-venv `make validate` passed: 30 schemas, 93 vectors, 8 Python tests, and `go test ./tools/...`.
- `git diff --check` passed.
- Local tools were Git 2.50.1, Git LFS 3.7.1, and Go 1.25.5.
- Primary sources checked: pinned Git LFS 3.7.1 `lfs/pointer.go` and `docs/spec.md`, plus Git `gitdatamodel`, `git-commit-tree`, `signature-format`, `git-cat-file`, fetch, and object-model source.
- No product, specification, schema, code, test, release, producer resource, or prior verdict resource was modified by this review.