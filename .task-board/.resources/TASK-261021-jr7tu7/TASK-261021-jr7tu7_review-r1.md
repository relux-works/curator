# TASK-261021-jr7tu7 review r1

Exact head: `d63a1ddf375ed5e40e7dde641a9beea83d67bf05`.
Base: `e41f96387287f5e76604627f9e63662f7d6d86d4`.
Repository: curator-spec. One introduced commit, signature `G`.

The task, reviewer assignment, all three preconditions, both producer results and the orchestrator's reference-direction redirect were read. This review covers all eight changed files. The detached review worktree and normal-.git validation clone both use the exact requested head. No repository source was modified.

## Surface sweep

| Surface | Held or broken | Attack and evidence |
| --- | --- | --- |
| Completeness | Held | Independently searched protocol, manager profile and CLI for git/repository/clone/fetch/acquisition, then read each owning source rule. Skillfile selections, configured/legacy project/global/hybrid sources, skill dependencies and substitutions, profile roots/overlays/direct/transitive closure requirements, context packages, MCP declaration packages and external build repositories are all named. Operation scope covers update/refresh/upgrade/repair/reacquisition only when already authorized to perform Git I/O. |
| Exclusions | Held | Compared local/path/file and committed-HEAD rules in core sections 5–6 and environments section 1; MCP server execution prohibition in environments 2.2; registry transport and remote-worker R-TRANSFER/R-NET-TOOL. Local sources and non-Git registry/server connections are explicitly excluded. Draft remote-workplace archive/import supplies no additional worker Git acquisition lane. |
| No grammar widening | Held | Compared all changed prose against unchanged core 6.1/6.3, Skillfile-source section 1, profile source kinds and manager 11.2/11.3. Legacy http/git support is not replaced by the narrower new-source grammar. Profile branch rejection, logical repository exclusivity, strict external-build no-port/no-alias rules, bounded policy fallback and read/store-only operation boundaries remain. Schemas and all conformance verdicts are byte-unchanged. |
| Reference direction | Held | Seven backlinks use the same identity/connection-hint reading of sections 1–3. Covered lanes uses plain names without new manager/environments citations. The revision-2 pointer resolves to Covered lanes and retains endpoint restrictions. |
| Contradiction hunt | Held | The old environments phrase canonical git source refers to canonical identity, not an HTTPS fetch mandate. No inspected lane requires canonical-HTTPS fetching. The explicit no-policy SSH example restates section 2's declared-endpoint rule. No contradictory lane rule was silently removed or edited. |
| Independence gate | Held | Diff against base for tools and schemas is empty. Ran the production gate: 89/89 baseline definitions identical, 34/34 clauses present, one existing exact exception. All eight gate controls passed, including post-baseline citation, changed-definition, wrong-file and unconditioned-exception refusals with valid controls. |
| Validator and links | Held | Production validate.py passes 74 schemas and 1294 vector files. Independently resolved all eight added file/anchor links. Production validate_local_links accepted an existing-file fixture and rejected a planted missing target with the intended broken-local-link diagnostic. |
| Digest reproducibility | Held locally, insufficient for release compatibility | Ran the real generator into an isolated non-repository copy and compared every tracked file to the exact head: zero changed file bytes. Digest regeneration is internally consistent; it does not preserve the frozen release pin (R1). |
| Accepted-release integrity / free hunt | **Broken, P1 R1** | Ran unchanged production release_gate.validate_skillfile_sources_manifest against an isolated exact-head copy after restoring only the already documented draft-owned Skillfile source document. It rejects the changed accepted-manifest hash. Restoring repository-transport.md and the accepted manifest from base in the same fixture makes it pass. Existing LiveCandidateAcceptedCorpusTests also detect the extra live deviation. |
| Public hygiene | Held | Inspected every added prose line and metadata diff. No personal paths, host/user identities, client/employer names, credentials or private material added. Only the expected eight files changed; diff --check passes. |

## R1 — P1: the editorial update changes the frozen accepted source corpus

Location: `conformance/skillfile-sources-v1/manifest.json:509`, together with `protocol/repository-transport.md:34`.

The source suite is accepted and pinned by `release/1.0.0-rc.14.json`. Its live-tree exemption covers only the draft-owned `protocol/skillfile-sources.md`, as enforced by `tools/test_release_gate.py`'s LiveCandidateAcceptedCorpusTests. Updating the transport document's digest changes the accepted manifest itself. Updating candidate.json makes the generic validator happy, but does not change the frozen rc.14 release pin.

Independent production reproduction:

1. Copy the exact-head files to a scratch directory, excluding Git metadata.
2. Restore `protocol/skillfile-sources.md` from `v1.0.0-rc.14`, exactly as the existing live-candidate positive control does.
3. Point `release_gate.ROOT` to that scratch directory and call `validate_skillfile_sources_manifest()` with its unchanged rc.14 release metadata.
4. Observe `ReleaseFailure: 1.0.0-rc.14 metadata does not pin the accepted source suite separately`.
5. In that same fixture, restore only `protocol/repository-transport.md` and `conformance/skillfile-sources-v1/manifest.json` from the review base. The same production call passes.

The second existing regression witness is `test_live_document_is_the_only_manifest_deviation`: transport prose now differs from the published tag in addition to the sole permitted draft-owned document. The companion positive release-gate test also fails.

Required fix: choose a supported publication path for this editorial clarification that preserves the frozen accepted artifact set, or obtain explicit authorization for a separately versioned/draft-owned contract treatment with matching release-gate controls. Do not rewrite historical rc.14 metadata or weaken the independence/release gates merely to accept this head. Rerun the two live-candidate tests and the complete validation suite at the replacement head.

## Evidence and execution

Evidence is retained privately under `.temp/TASK-261021-jr7tu7/` in the review workspace. Source-only comparison and gate failure reproduction are independently derived; the producer's passing generic-validator result is confirmed but is not release-gate acceptance.

The stock Python lacked jsonschema. A task-local venv was created with `python3 -m venv` and the exact repository `requirements-dev.txt` was installed with its Python (`jsonschema==4.25.1`); setup exited 0. This is an isolated test dependency installation, not a repository change. Local Python is 3.14.7; CI config names Python 3.12, so no exact hosted-environment equivalence is claimed. The R1 failure is an explicit digest-contract rejection reproduced with an adjacent base control, not a Python-version inference.

Every test command unset GIT_DIR and GIT_WORK_TREE and used the short workspace-local `.t` TMPDIR. Python used -B/PYTHONDONTWRITEBYTECODE to preserve the checkout. Logs retain complete results; public report text uses repository-relative paths.

| Check | Exit | Evidence |
| --- | ---: | --- |
| Production `python -B tools/validate.py` | 0 | validate-01.log: 74 schemas, 1294 vectors |
| Production independence gate | 0 | prose-01.log: 89/89 definitions; 34/34 clauses |
| Eight independence-gate unit controls | 0 | gate-controls-01.log: eight tests passed |
| `go test ./...` in normal-.git exact-head clone | 0 | go-suite-01.log: generator suite passed |
| New file/anchor links and real link-checker positive/negative controls | 0 | review-checks-01.log: eight links; intended missing-link rejection |
| Real generator into scratch copy, followed by all-tracked-file comparison | 0 | regenerate-01.log and regenerate-compare-01.log: zero changed bytes |
| Production release-gate head/base attack | 0 for the assertion script | release-probe-01.log: head rejected for its pin, base accepted; this is evidence of R1, not passing head validation |
| Full `python -B -m unittest discover -s tools -p 'test_*.py'` in normal-.git clone | 1 | python-suite-01.log: 879 tests in 1199.941s; failures=3, errors=1 |

The complete Python run reports only R1-related failures:

- `LiveCandidateAcceptedCorpusTests.test_live_gate_admits_draft_with_tagged_document`: release pin rejection (error).
- `LiveCandidateAcceptedCorpusTests.test_live_document_is_the_only_manifest_deviation`: transport is an additional forbidden deviation (failure).
- `AcceptedSourceCorpusIsolationTests.test_accepted_corpus_matches_rc14_tag_byte_for_byte`: accepted manifest differs (subtest failure) and enclosing candidate pin differs (failure).

The schema-drift negative control passes; the two adjacent corpus failure markers represent the subtest and enclosing test, not two separate test methods. No additional independent defect was found. Final checkout status is clean and HEAD remains the requested SHA.

Verdict: **changes_requested**. One blocking P1; no P2 findings.

```json
{
  "verdict": "changes_requested",
  "head": "d63a1ddf375ed5e40e7dde641a9beea83d67bf05",
  "findings": [
    {
      "id": "R1",
      "severity": "P1",
      "file_line": "conformance/skillfile-sources-v1/manifest.json:509",
      "mechanism": "Regenerating the transport-document digest changes the frozen rc.14 accepted manifest and enclosing candidate pin. Only protocol/skillfile-sources.md is draft-owned; the unchanged release gate rejects this additional accepted-artifact drift.",
      "reproduction": "At the exact head, run the full Python suite: 879 tests, exit 1, three failures and one error in the accepted-corpus controls. Independently restore only the draft-owned Skillfile document from v1.0.0-rc.14 in an isolated copy and call production validate_skillfile_sources_manifest with rc.14 metadata: it rejects the accepted-suite pin. Restoring repository-transport.md and the accepted manifest from base makes that same call pass.",
      "required_fix": "Publish the clarification through a supported path that preserves frozen accepted artifacts, or obtain explicit authorization for separately versioned/draft-owned contract treatment with matching release controls. Do not rewrite historical rc.14 metadata or weaken existing gates merely to accept this head. Rerun affected controls and the full suite at the replacement head."
    }
  ]
}
```
