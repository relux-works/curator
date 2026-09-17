# TASK-260910-14dnb7 — revision 2 reviewer verdict

Verdict: **accepted**. Both revision-1 corrections are closed. Route with `accept_cr(... revision=2 ...)` to integrating; acceptance is not landing or done.

## Candidate and review inputs

- Base b13ad660c8aa7fb0a484315fa9bed3d2dd9c9810; candidate tree 70e1b62c79d273f7fb7db4f0a6e188ca746164e1.
- Compared all 27 tracked candidate files with the managed Story worktree: 27/27 identical, zero mismatches. Seven changed paths, exactly the published list. Candidate was not edited; no review caches or files were created there.
- Read campaign rules, producer brief, results (including superseding Revision 2 section), published revision-2 patch and validation log, audit R1, normative protocol §§5/9/9.3 and profile §§2/5/11, pagination vectors and v2 schemas. Specification checkout HEAD is dced9b8317e0e8af79edf2d0539b32bd22b6c85b.
- Downloaded patch SHA-256 cbe263ad4de98fdaff02b73153768e88796f80e36ab57033ac588e6a102fbbd0 matches the assignment.
- `task-board spawn goal "$TASK_BOARD_RUN_ID"`: Active Goal: none (run is not goal-bound). No directives recorded.

## Per-item review

| Requirement | Evidence and assessment |
| --- | --- |
| Full signed boundary on every records/log success page | app.py:237,246,268 and :286,295,319 capture/build or recover the boundary and return the closed three-member envelope. snapshot.py:9 builds all seven snapshot fields using the immutable selected body. No separate success-return path omits it. |
| Page evaluated at emitted boundary | app.py:254,301 pass the carried boundary log_size to Store; store.py:576,626 bound records/log selection by that prefix. Existing boundary_available at store.py:708 compares the complete immutable body. |
| Chain byte equality, appends, timestamps | app.py:261,312 propagate the same snapshot into next cursors; :426 carries the complete signed object; :468-490 validates and returns it unchanged. tests/test_registry.py:886,926 check both complete chains after append, full canonical-document bytes, signature, snapshot equality, and immutable fields. No per-request timestamp refresh. |
| Rotation stability and retirement | tests/test_registry.py:739 drives genkey, prepare/activate rotation and retire-key, both HTTP endpoints, all overlap pages, original-key signatures, new active snapshot signer, old-envelope and new-envelope/old-boundary continuations refused after retirement. app.py:469 verifies the carried object against accepted keys. Independent replay passed. |
| Real schema validation | tests/test_protocol_conformance.py:179 uses Draft202012Validator with all local schema resources registered by $id; :206 enumerates registered cases for both v2 schemas; :250 validates actual endpoint responses. :222 rejects a served envelope modified to entry_hash="invalid". Independent replay and production-output mutation confirm enforcement. |
| Shared vectors | tests/test_protocol_conformance.py:250 drives both pagination flags through records/log endpoints; records append-after-first-page semantics and expected IDs are asserted, along with signed boundaries and snapshot equality. Coverage: 2/2 requested flags, 2/2 endpoint envelopes. Cursor disagreement vector is explicitly the sibling's scope. |
| CI pin and existing tests | .github/workflows/ci.yml:30 is exact dced9b8 full OID; dev installation already installs .[dev], including pyproject.toml:26 jsonschema>=4.23. Independent full suite passes at that root. Existing conformance tests remain unchanged; new vector test signs fixture bodies because real nested log schema requires signed audit records. |
| Release/operator docs | CHANGELOG.md:36 names R1, boundary, both v2 schemas, dced9b8, rotation/retirement behavior. README.md:26,60 describes the boundary. |
| Error paths/scope/architecture | Existing error paths are unchanged except required carried-snapshot malformed/untrusted refusal via existing 404 invalid_cursor mapping (app.py:491). No new disagreement-refusal algorithm, store edits, Docker/deploy edits, client/spec edits or unrelated changes. Existing staged-rotation test strengthened; cursor representation change is documented for sibling. |
| Static/build evidence | Strict mypy and diff whitespace check pass independently. No separate linter configured. Hosted gate evidence reports all six OS/Python matrix cells, strict mypy, distribution and Docker build green; accepted as attached evidence, not rerun. |

## Independent validation transcripts

Shell: zsh, `set -o pipefail`. Interpreter `/tmp/csk-venv/bin/python --version` → Python 3.14.6. Existing external venv reused; pytest's configured src path resolves the disposable candidate. Candidate extracted with `git archive 70e1b62c79d273f7fb7db4f0a6e188ca746164e1` into `/tmp/csk-review-r2`; no candidate-worktree execution artifacts.

```
cd /tmp/csk-review-r2
set -o pipefail
CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 /tmp/csk-venv/bin/python -m pytest -q 2>&1 | tee /tmp/csk-review-r2-pytest.log
126 passed, 2 warnings in 108.33s (0:01:48)
PYTEST_EXIT=0
/tmp/csk-venv/bin/python -m mypy 2>&1 | tee /tmp/csk-review-r2-mypy.log
Success: no issues found in 13 source files
MYPY_EXIT=0
```

Warnings are FastAPI/Starlette httpx and anyio BlockingPortal deprecations. `git diff --check` in the managed worktree: exit 0.

Hosted evidence used: TASK-260910-14dnb7_change-request_rev2-validation.log, remote run https://github.com/relux-works/curator-skill-registry/actions/runs/35231278201, exit 0, six matrix tests plus mypy/build/Docker success. I did not rerun the hosted gate, Python 3.11 or other OSes.

## Adversarial checks

Separate candidate extraction `/tmp/csk-review-r2-mutants`. Attached executable attack script and output log reproduce the following. Each mutant was applied individually to original app.py bytes and restored; baseline validation was never run against a moving source tree.

| Narrowing mutant | Committed detector | Result |
| --- | --- | --- |
| Re-sign boundary only on records cursor pages, preserving first pages and log behavior | test_rotation_overlap_keeps_chain_boundary_on_both_endpoints | exit 1, 1 failed (2.10s), chain equality |
| Re-sign boundary only on log cursor pages, preserving records behavior | same rotation test | exit 1, 1 failed (2.63s), chain equality |
| Verify carried snapshot signature only for records, retaining outer cursor signatures on both endpoints | same rotation test | exit 1, 1 failed (2.50s), post-retirement log continuation incorrectly returns 200 |
| Emit invalid entry_hash only in served log items, preserving all boundary fields | test_shared_service_page_boundary_vectors | exit 1, 1 failed (1.81s), real nested JSON Schema rejects invalid hash |

4/4 targeted mutants killed, 0 survivors. This is bounded evidence for these four shapes, not exhaustive mutation coverage of every cursor/schema clause.

After restoration the script independently invokes the committed rotation scenario with a wrapper measuring actual issued cursors: records [781,781,781], log [781,781]. Both chains retain original signed bytes throughout overlap; post-retirement refusal passes for old and overlap-issued cursors. Cursor size 781/4096 characters, 3315 headroom in this scenario (not a proof of every possible numeric-size boundary). The malformed served-entry regression independently passes. Attack script exit 0.

## Findings and limitations

No blocking findings or new spec gaps. Producer results retain historical revision-1 statements about re-signing and a purported rotation exception; these are superseded by their explicit Revision 2 correction and must not be read as the current contract. Profile §5 permits snapshot signature replacement, while §2 and protocol §9.3 require the complete chain boundary to stay identical; revision 2 correctly satisfies both.

Unsigned-snapshot cursors from the pre-change representation now fail validation; this deliberate compatibility limit is documented in Revision 2 results. No broader migration requirement was added by this review. Sibling cursor/boundary disagreement work, R2 performance, manager-client behavior and unrelated protocol conformance remain out of scope.
