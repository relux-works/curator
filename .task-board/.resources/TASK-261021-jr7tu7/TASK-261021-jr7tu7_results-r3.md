# TASK-261021-jr7tu7 results r3

## Delivery

- Repository: curator-spec; branch: `repository-transport-lanes`.
- New signed head: `71b24e02d2f9fbeebf70a6c881adc1ba1a1648ff`, signature `G`.
- Appended to reviewed head `d63a1ddf375ed5e40e7dde641a9beea83d67bf05`; ordinary push exited 0. Remote tracking head readback equals the new head. No force push, PR, main push or board-status mutation.
- This is rework on the existing branch, as requested. The branch retains its original `e41f96387287f5e76604627f9e63662f7d6d86d4` base. Fetched main is `3a3e041b3fbf76292e69b6abb4a0eca855524007`, one later commit not incorporated here. All evidence is for the exact delivered head.

## R1 fix and final scope

The new commit restores three files byte for byte to fetched `origin/main`:

- `protocol/repository-transport.md`: remove the Covered lanes list, SSH example and revision-2 pointer from the frozen document.
- `conformance/skillfile-sources-v1/manifest.json`: restore the accepted source-suite manifest, including its original transport digest.
- `conformance/candidate.json`: restore the accepted source-suite pin.

The final branch diff against its merge base contains only the seven lane-side backlinks in `protocol/core.md` (dependencies, legacy/configured sources and network substitutions), `profiles/manager.md` (external builds and the whole profile closure), `protocol/environments.md` (context and Git MCP packages), and `protocol/skillfile-sources.md` (network sources), plus that draft-owned prose digest in `conformance/draft-sources-v2/manifest.json`. No grammar, transport/port/alias rule, checker or checker exception changes remain. The acquisition inventory and no-contradiction findings in results r2 remain applicable; the proposed central list is withdrawn.

## Frozen files found

The complete 1432-entry path/digest inventory is attached separately as `TASK-261021-jr7tu7_frozen-inventory-r3.json`, verified in `outcomeResources`. Its local copy is `~/Werkstatt/curator-dev4/.out/TASK-261021-jr7tu7-frozen-inventory-r3.json`. It includes both manifest inventories, actual/expected digests, the sole draft-owned exception, and release/candidate pin metadata.

| Pinned surface | Files and disposition |
| --- | --- |
| Accepted source prose | `protocol/repository-transport.md` and `docs/skillfile-sources.md` are frozen. Transport is restored; docs were never changed. |
| Accepted source manifest's prose exception | `protocol/skillfile-sources.md` is pinned as rc.14 bytes for accepted validation, but its live bytes are explicitly draft-owned by `tools/validate.py` and pinned by the draft manifest. Only that permitted live prose differs. No new exception was added. |
| Accepted source schemas | All ten files under `schemas/skillfile-sources-v1/`: `README.md`, `build-receipt-v3.schema.json`, `install-marker-v5.schema.json`, `local-snapshot-v1.schema.json`, `skillfile-lock-v1.schema.json`, `skillfile-v2.schema.json`, `source-audit-v1.schema.json`, `source-policy-v1.schema.json`, `source-policy-v2.schema.json`, `source-types-v1.schema.json`. Unchanged. |
| Accepted source conformance data | All 125 files enumerated below `conformance/skillfile-sources-v1/` by its manifest. Unchanged; the manifest itself is restored and its hash matches rc.14 release metadata and the candidate pin. |
| Core conformance data | All 1294 `conformance/v1/` entries in its manifest, pinned by rc.14 release metadata and the candidate. Unchanged. |
| Pin/control documents | Accepted and core manifests, `conformance/candidate.json`, and `release/1.0.0-rc.14.json` preserve their main bytes. Historical metadata checks and release-gate implementation are unchanged. |
| Lane owners outside those frozen inventories | `protocol/core.md`, `profiles/manager.md`, and `protocol/environments.md` are not digest-frozen by these manifests. Their backlinks remain. The Skillfile backlink uses the existing draft-owned exception above. |

All non-exempt inventory bytes match their recorded hashes. A central transport edit would need an explicitly authorized new transport revision and corresponding release disposition; it is not part of this lane-side rework.

## Tests, claims, commands and exit codes

Evidence: `~/Werkstatt/curator-dev4/curator-spec/.temp/TASK-261021-jr7tu7/r3/` (E below). PY is the existing local validation venv's Python 3.14.7/jsonschema 4.25.1. Go is configured CURATOR_GO, version 1.25.5. No installation was needed.

| Command/check | Exit | Claim proved and evidence |
| --- | ---:| --- |
| `a2a inbox`; task/assignee/precondition query; all four resource gets; ack | 0 each | Latest addressed orchestrator request and board assignment confirmed; all briefs read before work. |
| `git fetch origin` | 0 | Named remote refs refreshed before existing-branch rework. |
| Restoration/inventory script | 0 | Three restoration paths equal fetched main; 138 accepted-source and 1294 core inventory entries checked, with only the pre-existing draft-owned prose exception. `frozen-inventory.json`. |
| `PY -B E/check-editorial.py` | 0 | Final scope is five files against the merge base; seven added backlink anchors resolve; frozen paths and gate/checker files equal main; added prose hygiene passes. Real link checker accepts an existing target and rejects a planted missing target with its intended diagnostic. `editorial.log`. |
| `PY -B E/production-gate.py` | 0 | Unchanged production `release_gate.validate_skillfile_sources_manifest` accepts the accepted corpus when only the draft-owned document is restored from rc.14. A newline plant in transport refuses with the exact transport-digest mismatch; updating its manifest digest without changing the release pin refuses with the exact metadata-pin mismatch. Restoring both bytes returns to acceptance. `production-gate.log`. |
| Same production-gate check on raw live draft | expected refusal, wrapper 0 | The sole draft-owned live Skillfile document produces the exact Skillfile-prose mismatch, as required by the existing live-corpus tests. This is explicitly not a claim that the unchanged strict gate accepts raw draft prose. |
| `PY -B tools/verify_skillfile_sources_independence.py` | 0 | 89/89 schema references byte-identical to rc.10, 34/34 cited clauses present; existing single allowlisted exception unchanged. `prose.log`. |
| `env -u GIT_DIR -u GIT_WORK_TREE PY -B tools/validate.py` | 0 | 74 schemas and 1294 vector files validate at delivered head. `validator.log`, `validation-exits.json`. |
| `env -u GIT_DIR -u GIT_WORK_TREE PY -B E/suite-groups.py N`, N=0..124 | 0 each | All 879 uniquely discovered tests run exactly once in 125 groups. All 40 `test_release_gate` tests pass, including all five live-corpus controls that expose the prior P1. Each group records its selected test IDs, command, count, time, exit and matching unittest OK summary. `suite-groups.json`, `python-001..125.{json,log}`, `validation-summary.json`. |
| `env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test ./tools/...` | 0 | Full Go tools package test suite passes at delivered head. `go.log`, `validation-exits.json`. |
| `git diff --check HEAD^ HEAD`; signature/ancestry/status checks | 0 | Whitespace clean; signature G; old remote task head remains ancestor; source and both exact-head clones clean. `validation-summary.json`. |
| `git push -u origin repository-transport-lanes`; tracking-head readback | 0 each | Fast-forward publication succeeded and readback equals tested signed head. |

The complete three recipes shown by `make -n validate` were run in disposable normal-.git clones at the exact delivered head. Python discovery was split to honor the standing per-call limit; this is full recipe-equivalent validation, not a claim of a monolithic `make validate` invocation. Every release-gate test ran separately; other groups contained at most ten tests. The slowest group took 194.917 seconds, below ten minutes. `TMPDIR=/tmp/cw-dev4`; inherited GIT_DIR/GIT_WORK_TREE were removed for every suite/test invocation. Go caches stayed under this Werkstatt. Both clones were deleted after clean HEAD/status verification; logs are retained in E.

One task-local editorial helper initially named a nonexistent checker-test filename; it was corrected to `tools/test_skillfile_sources_independence.py` and rerun successfully. No product check failure was hidden, skipped or waived. This editorial restoration changes no behavior, so no new product tests/gates were added; retained executable release regressions and bounded real-checker controls cover the affected integrity claim.

## Open decisions

None blocks this rework. The lane-side clarification is complete. Publishing new prose inside the frozen transport document would be a separate owner-authorized revision. No native-runtime qualification is claimed by these spec tests.
