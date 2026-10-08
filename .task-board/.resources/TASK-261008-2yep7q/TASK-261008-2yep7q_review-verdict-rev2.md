# TASK-261008-2yep7q — Re-apply the accepted manifest dependency-directory amendment: review revision 2

Verdict: **changes_requested**, route **to-dev**. CR-TASK-261008-2yep7q-2 is not accepted.

Reviewed base `210d3619051248cf9b87a7bfe02e5a9142a6b7cf` and exact candidate tree `34a792a29178c68529ea6444ebf717aeb28b106c`. Fresh `git ls-remote --symref origin HEAD` advertised `refs/heads/main` at the same base. The disposable shared clone was materialized with the candidate tree in its index; `git write-tree` still returned the exact candidate after testing and regeneration, with no unstaged changes. The assigned source worktree was not edited or committed. `spawn goal` reported this run is not goal-bound.

## Blocking finding: preserve the accepted corpus and put this amendment in the next draft

The attached successor steer explicitly requires `conformance/skillfile-sources-v1/` to remain byte-identical, including `manifest.json`, and requests `conformance/draft-sources-v2/` plus `schemas/draft-sources-v2/`. Revision 2 instead changes the accepted README, index and manifest; adds 25 cases and a vector there; adds v9 manifests and marker v6 under `schemas/skillfile-sources-v1/`; and advances the accepted namespace's candidate pin. Neither next-draft directory exists in the delta. This is the same release-isolation constraint called out by the steer, rather than a need for human approval.

The frozen rc.14 release pin is `sha256:061ec05ddb1746b72168d157f81d2930047cd62371ffa46cf585844da61ee6be`; the candidate manifest is `sha256:c72f021289be19b3a610974ebab282dcd98eca661351224c2d8fc245a3a4d65f`. The production check `release_gate.validate_protocol_artifacts`, called from `release_gate.main`, rejects the candidate at `validate_skillfile_sources_manifest` line 677. The attached recovery explanation does not remove this mismatch.

`tools/test_release_gate.py:282` now materializes the entire tagged rc.14 tree. This can be appropriate for historical-release tests, but those tests no longer inspect the candidate's source corpus. The comment at lines 290–292 and recovery outcome claim released-record/schema checks provide live-tree protection; those checks cover release records and already-released `*.schema.json` files, not the source corpus README/index/manifest or added draft files. Measured on this candidate: validate.py passes, all 40 targeted tests pass, and the production source-suite pin check fails. Thus their green results do not establish accepted-corpus preservation. The CHANGELOG claim at line 15 that the rc.14 suite remains unchanged is also contradicted by the delta.

Required rework:

1. Restore the accepted source corpus byte-for-byte and its candidate pin. Put the amendment's schemas, cases, index and manifest in the requested next draft namespace, following the old draft layout. Register and validate that root independently, with unique schema IDs and correct shared-definition references.
2. Preserve the core §4.4 intent and cross-references, and keep existing released schema JSON bytes frozen. The new marker revision is justified by v5's maximum manifest version 8; place the new carrier in the draft namespace and update links accordingly.
3. Account explicitly for shared normative documents: the accepted manifest also lists `protocol/skillfile-sources.md`, which this task intentionally changes. Historical accepted-suite verification must distinguish its tagged document inputs from current draft inputs. Do not solve that by rewriting the frozen accepted manifest. Historical tag fixtures may remain where justified, alongside a separate check of the candidate's accepted corpus and new draft suite. No rc.15 release-record edit is required by this review.
4. Add a negative regression that detects accepted-corpus or pin drift on the live candidate while admitting draft additions. Report its measured scope. Update the CHANGELOG and outcomes to describe the final namespaces and evidence honestly; rerun validation, the relevant tests and regenerate-check, then publish a new CR for another review.

## Swept surfaces

| Surface | Evidence and result |
|---|---|
| Core §4.4, closure §7, source and manager cross-references | Reviewed against the attached accepted patch: optional directory, omitted `.` identity, metadata/containment, closure conflicts, lock and per-package audit identity are re-applied. Held within the text reviewed. |
| Wire revisions and released schema bytes | v9 canonical and legacy manifests share directory grammar; source marker v6 differs from v5 only in identity/version band. Existing released schemas remain frozen. A new marker version is justified. Draft placement is covered by the blocking finding. |
| Cases, index and recipe | 25/25 added indexed case files present; each new manifest has 4 positives and 7 negatives; v5 has one version-9 refusal and v6 has positive/negative cases. Total source index 146 cases. Recipe is called by validate.py. Nine grammar and eight resolution vectors are present. |
| Directory refusal adversarial check | In-memory weakening of the shared directory grammar to additionally admit only `../developer` exposed both canonical and legacy indexed negatives: 2/2. This proves that refusal class at the JSON-schema validation entry, not downstream manager filesystem execution. |
| Accepted corpus / next draft | Broken: accepted manifest and corpus changed; requested next draft is absent. |
| Release tests and evidence | 40/40 rerun targeted tests pass on historical release fixtures, but live source-suite pin check fails. Record/schema immutability is not corpus immutability. |
| CHANGELOG / scope | Entry present, no LOGBOOK changes; claim that rc.14 suite is unchanged is false until the corpus is restored. |

This is one confirmed blocking mechanism, with its test-fixture masking and inaccurate claim recorded as consequences. No other blocking finding was confirmed in the swept amendment surfaces. Resolver/filesystem behavior in downstream manager implementations is outside this specification-only review and remains unverified.

## Validation evidence

Development interpreter: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.venv/bin/python` (Python 3.14). All independent checks ran in the disposable clone.

- `python -B tools/validate.py`: exit 0; `validated 73 schemas and 1294 vector files`.
- Requested pytest invocation: exit 1 because this development venv has no pytest. Used the repository's unittest runner for the bounded independent subset.
- `PYTHONPATH=tools python -B -m unittest -q test_validate.ManifestDependencyDirectoryTests test_release_gate`: exit 0; `Ran 40 tests in 71.714s / OK` (5 amendment tests plus 35 release-gate tests).
- `make regenerate-check` with the development venv first on PATH: exit 0; published release records validated, generator ran, and its four-path diff check was clean against the exact candidate index.
- `git diff --check BASE CANDIDATE`: exit 0.
- In-memory schema refusal-class widening check: exit 0; `parent-escape refusal-class widening exposed by indexed negative cases: 2/2`.
- Live production artifact check against this candidate: exit 1 with the source-suite pin mismatch above. This is a diagnostic of the rewritten accepted namespace, not a demand to publish the amendment as rc.14.

Accepted without full replay: the exact revision-2 attached validation log `TASK-261008-2yep7q_change-request_rev2-validation.log`, which reports 677 Python tests / OK in 726.462s and Go generator tests / OK, gate exit 0. I did not rerun the 677-test suite or Go tests. An initial full test_validate/test_release_gate attempt was interrupted (exit 130) once the attached timings were read, and is not claimed as a pass; it was replaced by the bounded subset above. Initial direct unittest module imports required `PYTHONPATH=tools`; those import errors are not counted as executed test cases. None of these reused or passing results supersedes the confirmed corpus-isolation finding.

Production diagnostic tail:

```text
rc.14 release pin: sha256:061ec05ddb1746b72168d157f81d2930047cd62371ffa46cf585844da61ee6be
candidate manifest: sha256:c72f021289be19b3a610974ebab282dcd98eca661351224c2d8fc245a3a4d65f
Traceback (most recent call last):
  File "<stdin>", line 8, in <module>
  File "/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TASK-261008-2yep7q-review-g12fzqws/repo/tools/release_gate.py", line 584, in validate_protocol_artifacts
    validate_skillfile_sources_manifest(release)
    ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~^^^^^^^^^
  File "/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TASK-261008-2yep7q-review-g12fzqws/repo/tools/release_gate.py", line 677, in validate_skillfile_sources_manifest
    raise ReleaseFailure(f"{PROTOCOL_VERSION} metadata does not pin the accepted source suite separately")
release_gate.ReleaseFailure: 1.0.0-rc.14 metadata does not pin the accepted source suite separately
```

Regenerate-check tail:

```text
python3 tools/validate.py --release-history-only
validated published release record bytes
go run ./tools/generate-vectors -root .
WORK=/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/go-build2125800791
git diff --exit-code -- conformance/v1 conformance/candidate.json conformance/skillfile-sources-v1/manifest.json release/1.0.0-rc.14.json
```

```verdict-findings
{
  "findings": [
    {
      "id": "accepted-source-corpus-drift",
      "row": "accepted source corpus and draft isolation",
      "invariant": "The rc.14 accepted conformance/skillfile-sources-v1 corpus, including its manifest, stays byte-identical; the amendment belongs in the next draft namespace.",
      "mechanism": "conformance/skillfile-sources-v1/manifest.json:1 is regenerated for 167 members instead of the published 138; new draft cases and schemas are added to the accepted namespace.",
      "reproductions": [
        {
          "test_file": "tools/release_gate.py",
          "command": "PYTHONPATH=tools python -B -c \"import release_gate; release_gate.validate_protocol_artifacts('1.0.0-rc.14')\"",
          "expected_failure": "ReleaseFailure: 1.0.0-rc.14 metadata does not pin the accepted source suite separately",
          "pinned_blobs": [
            "sha256:c72f021289be19b3a610974ebab282dcd98eca661351224c2d8fc245a3a4d65f",
            "sha256:061ec05ddb1746b72168d157f81d2930047cd62371ffa46cf585844da61ee6be"
          ]
        }
      ],
      "severity": "regression",
      "repeat-of": "none"
    }
  ],
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
      "result": "broken"
    },
    {
      "row": "release-gate tests and evidence bounds",
      "result": "broken"
    },
    {
      "row": "changelog and repository scope",
      "result": "broken"
    }
  ],
  "free_hunt": []
}
```
