# TASK-260917-3j8sjw — registry-service-hash-version

Ready for review; changes are uncommitted in the assigned Story worktree.

Implemented framing declarations on schema-2 records; frozen schema-1 records retain implicit framing 1. Countersigning preserves schema/version. Store publication validates declarations, and content query filtering plus latest-record partitioning carry the version from authoritative JSON without rewriting history. Mixed exports use bundle schema 2; imports validate both envelope and record versions before existing signature/chain/Merkle/P4 checks. HTTP metadata advertises both record schemas. CI now pins the normative spec revision. Deployment note: docs/content-hash-versions.md; Unreleased changelog updated.

Normative interpretation: curator-spec b1a2efb6fa28d014968a2a8fd7641823b5f3cf28 freezes registry-snapshot-v1 (no hash_version field). Snapshot head/Merkle commits versioned records; adding a top-level version would violate the specified schemas. Both framings use identical sha256 digest syntax: actual preimage framing cannot be established without the tree. Publication checks schema/framing agreement with stable HTTP 400 invalid_record and hash_version_mismatch diagnostic. Source-only queries are discovery and return both versions per registry HTTP contract; content queries (including conjunctive identity filters) enforce version equality. Clients own artifact matching after source-only discovery.

## Commands executed directly and observed exit codes

Interpreter P=/tmp/registry-3j8sjw-venv/bin/python (Python 3.14, macOS).
Conformance root C=/tmp/curator-spec-3j8sjw/conformance/v1, disposable git clone checked out to full revision above.

- P -m pytest -q tests/test_hash_version.py: exit 0, initial 14 tests passed.
- CURATOR_CONFORMANCE_ROOT=C P -m pytest -q: exit 0, initial 239 passed; rerun after adding cursor binding test: exit 0, 240 passed in 33.56s, no skips. One upstream Starlette httpx deprecation warning.
- P -m mypy: exit 0, no issues in 15 source files. No source changes after this run.
- P -m build: exit 0 (twice; rebuilt after the additional packaged test), sdist and wheel produced.
- P -m twine check dist/*: exit 0 on initial build; repeated on rebuilt artifacts before handoff.
- git diff --check: exit 0.

No separate lint target is configured; strict mypy and whitespace validation are the repository's applicable static checks. Docker and cross-platform CI were not run locally; only the local macOS/Python 3.14 suite is asserted here. No prior attached test evidence was substituted for execution.

## Negative evidence

Production route POST /v1/records -> validate_record / _countersign -> Store._append_locked; GET /v1/records -> Store.records_page; bundle.import_bundle -> validate_record -> transactional upstream import.

HTTP tests exercise v1 and v2 positive matches and both mismatch directions, content-only and content-plus-source queries, implicit v1 compatibility, projection coexistence, pagination boundaries and cross-version cursor refusal. Publication has 7/7 invalid declaration cases, exercised via HTTP and direct store append. Mixed bundles test valid import, no-op reimport, durable reopen, and 3/3 rejection cases (frozen envelope with v2, wrong record framing, missing v2 public key); rejected imports leave no records.

Pinned schema cases: 12/12 exercised, including 7/7 invalid cases across audit-record v1/v2, log entry v2, bundle v2 and log-response v3. Actual HTTP log/snapshot and mixed export/import are schema-validated. Existing registry-service query vectors: 6/6 exercised, along with the full P4/R regression suite. Registry client resolution vectors are client policy and not claimed as service matching coverage.

Disposable source copy /tmp/registry-3j8sjw-mutant; repository source was never mutated:
1. Delete version SQL predicate and its parameter in Store.records_page. Run P -m pytest -q -o pythonpath=/tmp/registry-3j8sjw-mutant tests/test_hash_version.py -k http_matching_is_version_scoped. Real exit 1, 2 failed / 12 deselected; expected failures because mismatched records were returned.
2. Restore copy from production, then narrow the predicate to only `if not source_identity` (identity queries bypass comparison). Same command: real exit 1, 2 failed / 12 deselected; expected failures in the identity-filter cases. This proves the gate's scope beyond delete-only mutation.

Mutant result: 2/2 mutants killed. The later 240-test run used unmodified production source.
