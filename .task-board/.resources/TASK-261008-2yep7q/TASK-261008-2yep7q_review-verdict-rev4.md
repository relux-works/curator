# TASK-261008-2yep7q — Re-apply the accepted manifest dependency-directory amendment: review revision 4

Verdict: **accepted**. Accept CR-TASK-261008-2yep7q-4 revision 4 and route to **integrating**. This is reviewer acceptance; it does not attest landing on main. No blocking findings remain.

Reviewed base `210d3619051248cf9b87a7bfe02e5a9142a6b7cf` and exact candidate tree `7bb3b5d0e2819e73ff1e8aa18b7aa2e2fc759ca2`. Independent execution used a disposable shared clone materialized from that exact tree. Its HEAD tree remained identical and its working tree was clean after tests, regeneration and restored data mutations. The assigned Story worktree still matches the candidate on 47/47 changed paths; the reviewer did not change or commit source code. Fresh remote HEAD discovery advertised refs/heads/main at the base; an exact-ref fetch in the disposable clone confirmed that OID, and final remote discovery still matched. The run is not goal-bound.

## Previous finding and required rework

`accepted-source-corpus-drift` from revision 2 is **fixed**, rather than waived. Decision 0022 explicitly excludes manifest v9 from accepted sources revision 1 and sends it to the next draft namespace (lines 25–26, 65–66). The revision-4 layout follows that repository decision.

1. The accepted corpus is restored: all 126 conformance files and 10 schema-directory files match the rc.14 tag byte-for-byte, with identical inventories and no extra files. The accepted manifest retains 138 entries and `sha256:061ec05ddb1746b72168d157f81d2930047cd62371ffa46cf585844da61ee6be`. Candidate metadata matches tagged bytes too.
2. The amendment lives in `conformance/draft-sources-v2/` and `schemas/draft-sources-v2/`. Unique schema IDs and references resolve through the combined core/accepted/draft registry. Canonical and legacy schema-9 manifests agree beyond identity metadata, reuse the accepted directory grammar, and retain earlier-version rejection. Source marker v5 is frozen at maximum manifest version 8. Draft marker v6 is justified by recording schema 9; its shape equals source marker v5 after the schema/manifest version changes and the relative package reference are normalized. The already-released lock-v1 member carries a directory with exactly the shared grammar; no lock widening or new lock revision is needed.
3. Core §4.4, closure §7, source and manager cross-references preserve optional selection, omitted/explicit `.` equivalence, containment/metadata checks, full package identity, same-name conflicts, independent per-package validation/audit and lock/marker binding. Only the amended source-contract document is read from tagged bytes during accepted-manifest verification; the draft manifest pins its live bytes. Other accepted input files stay checked live. No accepted manifest rewrite was used to admit the document change.
4. Live-candidate negative regressions accompany historical fixtures. Accepted case/schema drift is rejected, non-owned document drift still fails, missing-tag evidence fails loudly, candidate status/pins/absence fail through validate.main, and draft additions pass. The frozen corpus identity test derives its entire expected inventory from the tag, including README/index/manifest bytes. Historical lifecycle removal and stale compiled-fixture mutations still fail even with updated manifest hashes. Revision 4 also aligns both workflows and the test inventory with the Makefile's five generated paths, including the draft manifest.

## Swept surfaces

| Surface | Result and evidence |
|---|---|
| Core normative directory selection and cross-references | Held: reviewed against the accepted patch and current core/source/manager wording; dependency identity includes directory throughout the specified closure and audit contract. |
| Wire schema revisions and released-schema freeze | Held: v9 manifests and v6 marker isolated in draft; accepted schemas unchanged; existing lock directory grammar equivalent; carrier shape and version-band assertions pass. |
| Schema cases, index and vector recipe | Held: 25/25 cases indexed and present, 16/16 negatives; three draft schemas have positive and negative cases. Nine grammar and eight resolution cases remain. 23/23 JSON files originating in the accepted patch match, allowing only the intended marker 5→6 adjustment. validate.main invokes both draft-schema/manifest checks and the vector recipe. |
| Accepted source corpus and draft isolation | Held: 136/136 accepted files plus 1/1 candidate metadata match the tag, with exact inventories; original 138-entry manifest and release pin retained. |
| Release-gate tests and evidence bounds | Held: historical tagged fixtures preserve refusal tests; live-candidate cases run the unchanged strict source-suite gate on the live copy with only the owned document restored. Missing tag does not take a fallback. Production CLI attacks below fail as required. |
| Changelog and repository scope | Held: CHANGELOG describes final draft namespaces, frozen corpus and generator ownership. No LOGBOOK changes. No schema/claim release or downstream implementation is represented as completed. |
| Workflow and Makefile regeneration inventory | Held: five-path inventories agree and regenerate-check is clean against the tracked exact candidate snapshot. |

## Independent execution

Python 3.14.6. Initial validate.py used the repository development venv. Because that venv has no pytest, the reviewer created an external disposable venv with jsonschema 4.25.1 (the repository pin) and pytest 9.1.1. No repository dependency files were changed.

- `python -B tools/validate.py`: exit 0; 73 core schemas and 1294 core vector files reported. Draft cases are separately exercised by the registered checks; the printed core counts do not include the 25 draft cases.
- Bounded pytest selection across test_validate.py and test_release_gate.py: exit 0, **50 passed**, 564 deselected, **200 subtests passed**. Selected ManifestDependencyDirectoryTests, AcceptedSourceCorpusIsolationTests, SkillfileSourcesSuiteManifestTests, WorkflowRegenerationScopeTests, ProtocolRC14ReleaseGateTests and LiveCandidateAcceptedCorpusTests. The release subset includes removal of a lifecycle case and stale compiled-fixture rejection against the tagged tree.
- CandidateMetadataTests separately: exit 0, **2 passed**, 572 deselected, **3 subtests passed**, exercising wrong status, both wrong pins, and absent metadata through validate.main.
- `go test -count=1 ./tools/...`: exit 0; includes the frozen-tag digest check and both new manifest-writer tests.
- `make regenerate-check`, disposable venv first on PATH: exit 0. Published release-record guard, generator and all five generated-path comparisons passed.
- `git diff --check BASE CANDIDATE`: exit 0. `gofmt -l` on both changed Go files produced no paths.

Accepted as producer evidence, not independently replayed: the exact revision-4 attached `TASK-261008-2yep7q_change-request_rev4-validation.log` reports **689 tests / OK** in 1098.582 seconds, Go tests / OK and gate exit 0. Recovery details are in `TASK-261008-2yep7q_gate-fix-rev3.md`. An initial independent full pytest attempt was interrupted with exit 2 after 16 tests and 17 subtests (86.32 seconds) once the published full-suite duration was read; it is not claimed as a full pass. It was replaced by the bounded selections above. The review does not claim a new independent full 689-test replay or cross-platform CI execution.

## Negative evidence and stated bounds

Four selected drift classes were attacked through the real `tools/validate.py` CLI, with exact bytes restored after each mutation: accepted case bytes, added accepted inventory member, candidate source pin, and live draft-contract bytes. **4/4 refused** with named errors. This is measured coverage of those four classes, not a claim of exhaustive mutation coverage of every validator clause.

A narrowly widened shared directory schema that additionally permits only `../developer` exposes both indexed canonical/legacy escape cases: **2/2** become wrongly valid. This proves that refusal class at the JSON-schema validation entry, not manager filesystem behavior. The historical rc.14 fixture setup was also driven against a real Git repository lacking the tag: **1/1** refused loudly at git archive. Only **1/3** accepted document inputs has a tagged-document exemption; the other **2/3** remain live-checked and have negative regressions.

The eight resolution cases are specification requirements checked for internal consistency, not downstream resolver/install execution. Manager filesystem containment, install publication, audit caching and runtime behavior remain unverified by this specification-only review. Neither the draft README nor this verdict claims implementation conformance.

## Output tails

inventory.log

```text
conformance/skillfile-sources-v1: 126/126 tag files identical; no extra files
schemas/skillfile-sources-v1: 10/10 tag files identical; no extra files
accepted manifest: 138 entries; sha256:061ec05ddb1746b72168d157f81d2930047cd62371ffa46cf585844da61ee6be
candidate.json: tag bytes identical
draft indexed cases: 25/25 present; 16 negatives
draft schema inventory: 25/25 indexed, no unindexed JSON cases
```

validation.log

```text
validated 73 schemas and 1294 vector files
```

targeted-tests.log

```text
........ [ 16%]
..........................................                                       [100%]
50 passed, 564 deselected, 200 subtests passed in 182.27s (0:03:02)
```

candidate-pin-tests.log

```text
..                                                                    [100%]
2 passed, 572 deselected, 3 subtests passed in 50.94s
```

go-tests.log

```text
WORK=/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/go-build3641685354
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	11.402s
```

regenerate-check.log

```text
python3 tools/validate.py --release-history-only
validated published release record bytes
go run ./tools/generate-vectors -root .
WORK=/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/go-build1022490313
git diff --exit-code -- conformance/v1 conformance/candidate.json conformance/skillfile-sources-v1/manifest.json conformance/draft-sources-v2/manifest.json release/1.0.0-rc.14.json
```

adversarial.log

```text
accepted case drift: refused by tools/validate.py CLI: validation failed: skillfile-sources-v1 manifest digest mismatch: conformance/skillfile-sources-v1/schema-cases/build-receipt-v3/valid.json
accepted added file: refused by tools/validate.py CLI: validation failed: skillfile-sources-v1 manifest inventory is incomplete or unsorted
candidate source pin drift: refused by tools/validate.py CLI: validation failed: candidate skillfile_sources_v1 pin does not match the suite manifest
live draft document drift: refused by tools/validate.py CLI: validation failed: draft-sources-v2 manifest digest mismatch: protocol/skillfile-sources.md
Production CLI negative classes: 4/4 refused
Shared grammar widening limited to ../developer: 2/2 canonical/legacy negative cases expose mutant
Historical fixture absent tag: real git archive refusal surfaced loudly
Tagged document exemption: exactly 1/3 accepted document inputs; other 2/3 remain live-checked
```

source-identity.log

```text
Assigned source worktree matches reviewed candidate bytes on 47/47 changed paths. No source changes or commits made by reviewer.
```

```verdict-findings
{
  "findings": [],
  "notes": [],
  "surface_results": [
    {
      "row": "core normative directory selection and cross-references",
      "result": "held"
    },
    {
      "row": "wire schema revisions and released-schema freeze",
      "result": "held"
    },
    {
      "row": "schema cases, index and vector recipe",
      "result": "held"
    },
    {
      "row": "accepted source corpus and draft isolation",
      "result": "held"
    },
    {
      "row": "release-gate tests and evidence bounds",
      "result": "held"
    },
    {
      "row": "changelog and repository scope",
      "result": "held"
    },
    {
      "row": "workflow and Makefile regeneration inventory",
      "result": "held"
    }
  ],
  "free_hunt": []
}
```
