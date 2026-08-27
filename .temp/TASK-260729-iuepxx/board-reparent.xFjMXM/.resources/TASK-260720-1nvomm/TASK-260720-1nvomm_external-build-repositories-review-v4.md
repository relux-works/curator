# External build repositories v4 — independent review verdict

Verdict: **changes requested**. Route `TASK-260720-1nvomm` to `analysis` for bounded architecture rework and a fresh reviewer cycle.

Reviewed artifact: `TASK-260720-1nvomm_external-build-repositories-architecture-v4.md`, SHA-256 `edd2e705685f588d90355e9588a7f8b02605f2bef9cad3308f1958ac62714e64`. The active run `RUN-260727-d2dd94` is not goal-bound.

## Accepted portions

- The previously accepted schema-7/versioning, immutable Git lock, logical-target/output boundary, declared/effective identity, audit-before-cache/compiler, receipt-v2/marker-v3, status/repair/GC, signing, future-driver, and board-impact decisions remain intact.
- V4 resolves the prior raw-object child-graph contradiction. One exact direct `git cat-file --batch` session now has fixed argv, clean environment, full-OID-only requests, byte-framed raw output, no transformation flags, bounded resources, no child process, no network, manager object-ID recomputation, and a common network/local-store path.
- V4 resolves the prior implementation-selected local pack subset. Pack versions 2/3 and index version 2 are common for SHA-1 and SHA-256, with hash-width trailers, CRC32, offsets, private inert copying, trusted reading, and stable rejection of other formats.
- V4 defines a bounded all-reachable-blob LFS scan and a stable pre-audit error. The current extension grammar is not complete enough to close the invariant, as finding 1 explains.
- Exact init/common-fetch flags and the full-OID destination worked independently on Git 2.50.1; the only destination was `refs/curator/locked` and `FETCH_HEAD` remained absent.

## Required changes

### 1. The LFS detector misses pointers accepted as canonical by the cited official implementation

Section 6.6.3 accepts extension names only as `[a-z0-9][a-z0-9.-]*`. The cited current Git LFS implementation instead recognizes extension keys with `extRE = \Aext-\d{1}-\w+`, does not validate the remainder of the extension name, and its canonical encoder reproduces the parsed name. On the installed official Git LFS 3.7.1:

```text
ordinary current pointer                    strict exit 0
ext-1-a_b                                  strict exit 0
ext-1-A_B                                  strict exit 0
ext-1-a!                                   strict exit 0
```

V4 treats the last three as ordinary blob content because underscore, uppercase, and `!` do not match its canonical or legacy/noncanonical extension-name rule. They are nevertheless accepted as canonical by `git lfs pointer --check --stdin --strict`, so a repository can carry bytes that Git LFS will hydrate while the manager lets the pointer text reach `go:embed`. That contradicts the claimed rejection of recognized Git LFS pointers.

The packet also calls a non-empty pointer with `size 0` canonical, while the cited encoder emits an empty blob for size zero and Git LFS strict checking returns exit 2 for the non-empty spelling. Rejection may remain conservative, but the canonical/noncanonical classification must be accurate.

Required correction:

1. State whether the invariant covers the published pointer specification only or every pointer accepted by the pinned/tested Git LFS family. The current text claims both.
2. To retain the no-LFS-pointer invariant, define a bounded conservative detector that includes every key/value form accepted by the supported current parser family, including its actual extension-name behavior, while still keeping an explicit false-positive boundary.
3. Classify non-empty `size 0` as deliberately noncanonical if it remains rejected.
4. Add cross-language vectors for lowercase/dot/dash extensions, underscore, uppercase, punctuation accepted by the tested parser, duplicate priorities, non-empty size zero, and near misses.

Primary evidence: [Git LFS pointer specification](https://github.com/git-lfs/git-lfs/blob/main/docs/spec.md) and [current pointer parser/encoder](https://github.com/git-lfs/git-lfs/blob/main/lfs/pointer.go).

### 2. Local raw-object proof lacks an exact commit/tag semantic grammar

Section 6.6.2 precisely defines tree entries but only says that manager code parses the commit's exact `tree <id>` header and validates tag `object`/`type` headers. It does not define:

- whether a commit must contain exactly one tree header;
- which position that header occupies;
- whether duplicate/misordered tree headers, malformed continuation records, NULs, or a missing header/message separator reject;
- whether tag `object` and `type` occur exactly once and in the required order;
- whether the declared tag type must match the actual recomputed target object type.

This is observable on the local-substitution path because `cat-file` reports the stored type and raw bytes but does not fsck commit semantics. Independent Git 2.50.1 smoke created this object with `hash-object --literally`:

```text
tree 4b825dc642cb6eb9a060e54bf8d69288fbee4904
tree 2561a62d4223eb7660d3b6b02b707048382f4019

message
```

`git cat-file -t` reports `commit`, and the exact batch reader returns it as a valid typed object. Git fsck rejects the format, but v4 invokes fsck only indirectly through network fetch policy; local stores are copied and sent directly to `cat-file`. Two conforming managers can therefore choose different trees or fail under different errors for the same effective commit.

Required correction:

1. Define one byte-level manager grammar for the selected commit and annotated-tag chain, including required header order, uniqueness, object-ID width/case, continuation handling, separator handling, and bounded ignored metadata/message bytes.
2. Require exactly one root tree and require each tag `type` header to match the actual recomputed target type.
3. Give malformed commit/tag inputs one stable pre-audit error and add network/local parity vectors for duplicate tree/object/type, wrong target type, malformed separators, and valid signed/extra-header objects.
4. Alternatively, add a separately fixed trusted fsck child for both network and local private stores, but that would expand the exact child graph and requires its own reviewed argv/output contract. The manager-parser option is narrower and recommended.

Primary evidence: [git-cat-file batch behavior](https://git-scm.com/docs/git-cat-file), [Git commit-tree model](https://git-scm.com/docs/git-commit-tree), and [Git user-manual commit object](https://git-scm.com/docs/user-manual).

## Independent validation evidence

- All ten fenced JSON blocks parsed.
- `validation-venv/bin/python tools/validate.py` passed: `validated 30 schemas and 93 vector files`.
- Full pinned-venv `make validate` passed: 30 schemas, 93 vectors, 8 Python tests, and `go test ./tools/...`.
- Host `make validate` failed only because host Python lacks `jsonschema`; this matches the producer anomaly and does not invalidate the pinned gate.
- `git diff --check` passed.
- Exact `cat-file --batch` framing passed for SHA-1 and SHA-256 object IDs on Git 2.50.1.
- Exact common fetch options passed with a full-object refspec; the selected ref matched, only the manager destination existed, and `FETCH_HEAD` was absent.
- Git LFS 3.7.1 strict checks produced the extension and size-zero results recorded above.
- No product, specification, schema, code, test, release, producer resource, or prior verdict resource was modified by this review.

These are bounded architecture defects, not a stop-the-line decision.
