# TASK-261021-jr7tu7 review r2

Requested head: `71b24e02d2f9fbeebf70a6c881adc1ba1a1648ff`.
Branch: `repository-transport-lanes`, curator-spec.
Original base: `e41f96387287f5e76604627f9e63662f7d6d86d4`.
Fetched main: `3a3e041b3fbf76292e69b6abb4a0eca855524007`.

The authenticated orchestrator request, named reviewer resource, all five preconditions, producer results r3 and attached frozen inventory were read. The r3 brief explicitly withdraws the central transport edit and limits this rework to the lane-side clarification. Both introduced commits have signature `G`. Both the detached review worktree and the normal-.git suite clone are at the exact requested head. Main has one additional unrelated local-hosted-launch amendment commit; this review does not claim to test that combined tree.

## Surface sweep

| Surface | Held or broken | Attack and evidence |
| --- | --- | --- |
| Prior P1 R1: frozen corpus | Held | Recomputed the production gate's closed source inventory directly from its roots, independently read both manifests, hashed every file and compared raw Git blobs against fetched main. All 1432 entries match the producer inventory: 138 accepted-source entries and 1294 core entries. Every non-exempt file is byte-identical to main and matches its digest. The sole permitted live exception remains protocol/skillfile-sources.md. Accepted/core manifests, candidate pin, rc.14 metadata and repository-transport.md are unchanged. |
| Production release gate | Held | Called unchanged release_gate.validate_skillfile_sources_manifest on an isolated exact-head copy. Raw live draft refuses only the expected Skillfile-prose digest. Restoring that one document from rc.14 makes it pass. A newline planted in transport fails with the exact transport-digest diagnostic; updating that digest without the release pin fails with the exact metadata-pin diagnostic. Removing both plants restores acceptance. This closes the r1 witness without weakening a gate. |
| Seven lane-side statements | Held | Read each changed sentence against transport sections 1–3 and its surrounding owner rules. Independently searched protocol, profiles and CLI for Git sources, repositories, clones and fetches. The statements cover skill dependencies, legacy/configured sources and substitutions, external build repositories, profile roots/overlays/transitive closure, context packages, Git MCP packages and new network sources. They describe network acquisition and do not grant it to local/path/committed-HEAD sources or store-only operations. No canonical-HTTPS contradiction was found. |
| No widening / frozen lane owners | Held | The complete branch delta is exactly four prose files and the draft manifest. Core, manager and environments are outside the accepted inventories; Skillfile prose uses the existing explicit draft-owned exception. Schemas and tools have zero diff. Existing legacy URL schemes, stricter external-build endpoint grammar, port/alias restrictions, logical-source scope, exact refs, bounded fallback and package-execution boundaries remain unchanged. |
| Independence gate | Held | Production checker passes: 89/89 definitions match rc.10, 34/34 cited clauses present, the same single conditional exception. No checker, test or exception change exists in the branch. Its positive/negative controls are included in the full Python run below. |
| Links and prose | Held | Independently resolved all seven added links including heading anchors. Exercised production validate_local_links with an existing target and a planted missing target; the latter fails with its intended broken-local-link diagnostic. Production validator passes 74 schemas and 1294 vector files. |
| Reproducibility / free hunt | Held | Ran the real generator into an isolated non-repository copy of exact-head bytes; every tracked file remains byte-identical. No source/release-history/schema/expectation change is concealed by digest regeneration. |
| Public hygiene | Held | Inspected the full added prose and digest diff for personal paths, host/user identities, client/employer names and secrets. None found. git diff --check passes. |

## Independent evidence

Evidence is retained in the review workspace under `.temp/TASK-261021-jr7tu7/r2/`; the independently recomputed inventory is `independent-inventory.json`. No repository source was edited. Control plants ran only in an isolated scratch copy. The first inventory helper used git archive and encountered the intentional export-subst fixture transformation; it was corrected to read raw blobs with git cat-file --batch. All final inventory assertions pass. That helper correction is not a product failure.

Local platform/toolchain: macOS arm64, Go 1.25.5, Python 3.14.7 with the existing task-local jsonschema 4.25.1 venv. CI specifies Python 3.12; exact hosted-toolchain equivalence is not claimed. Every test invocation unsets GIT_DIR/GIT_WORK_TREE, uses the workspace-local short `.t` TMPDIR, and suppresses Python bytecode output. The full Python suite and full Go suite run in a normal-.git exact-head clone; validator, independent probes and prose checker use the separate exact-head worktree.

| Check | Exit | Evidence |
| --- | ---: | --- |
| Production validator | 0 | validator.log: 74 schemas, 1294 vector files |
| Production independence gate | 0 | independence.log: 89/89 definitions, 34/34 clauses, existing exception |
| Full Go suite, normal-.git clone: go test ./... | 0 | go-normal-suite.log |
| Independent frozen inventory, production release/link controls and seven anchors | 0 | probes.log and independent-inventory.json |
| Production generator and all-tracked-file byte comparison | 0 | generator.log; zero changed file bytes |
| Full normal-.git Python suite: python -B -m unittest discover -s tools -p 'test_*.py' | 0 | python-suite.log: 879 tests in 841.155s, OK; includes all 40 release-gate tests and eight independence-gate controls |

All make validate recipes were independently executed; the Python recipe ran as one complete discovery invocation. Test-generated EXPECTED RED diagnostics are passing negative-control output, not suite failures. These are specification/checker validations, not a claim of manager runtime qualification. The source and normal-.git clone remain clean at the exact requested head after validation.

R1 is resolved. No new findings. Verdict: **accepted** for the narrowed r3 lane-side scope at this exact head. A central transport prose revision remains outside this task's accepted scope.

```json
{
  "verdict": "accepted",
  "head": "71b24e02d2f9fbeebf70a6c881adc1ba1a1648ff",
  "findings": []
}
```
