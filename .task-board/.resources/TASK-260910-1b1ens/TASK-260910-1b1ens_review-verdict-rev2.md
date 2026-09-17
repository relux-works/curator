# TASK-260910-1b1ens — review verdict, revision 2

Verdict: **accepted**. Revision-1 finding F1 is resolved by the settled revision-2 verification order. No blocking findings.

Reviewed candidate: curator-spec Story worktree `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-25yc0h/worktree`, against `origin/main` (`23dafa7`). The attached `TASK-260910-1b1ens_spec-patch_rev2.patch` equals the live diff byte-for-byte, before and after regeneration. SHA-256: `04bb2902ac9934bbafab087c9ef43e693598701a38fb3fc37819093fc7f09629`; both stable patch IDs: `b99f12815b48919d29e004b9bcfbd92750433497`. New files were already tracked/intent-to-add, so no reviewer index change was needed.

The runtime CR's empty **curator** repository delta is correct: this leaf delivers a **curator-spec** contract, whose 18-file delta is captured by the attached spec patch and reviewed worktree. It must not modify the curator implementation. Acceptance does not claim that the protocol has landed: `accept_cr(... revision=2 ...)` routes to integrating, and the producer integration transaction remains responsible for landing. No commit acknowledgement or done transition is supplied by this reviewer.

Read the campaign rules, producer brief, producer evidence and both patches, and audit R1/P1 (`docs/security-audit-2026-09.md:88`). The audit describes record pages hiding revocations by using an older boundary than the separately fetched snapshot. Architecture-diagrams skill is not applicable: no diagram change or architecture redesign is requested. The project-management skill was used for reviewer lifecycle and evidence.

## Per-item review

All paths below are relative to the curator-spec Story worktree.

| Requirement | Evidence, exact excerpt and location | Result |
|---|---|---|
| Full signed boundary on every records/log page | `protocol/registry.md:351`: “Every successful `/v1/records` GET and `/v1/log` response carries a REQUIRED `boundary`”; line 353: “all fields including `sig`”; line 354: “All pages of one cursor chain carry a byte-identical” boundary. | Pass |
| Versioned closed envelope schemas | Both `schemas/v1/{records,log}-response-v2.schema.json:6` include `"boundary"` in `required`; line 10 has `"$ref": "registry-snapshot-v1.schema.json"`; line 12 is `"additionalProperties": false`. Referenced snapshot schema line 6 requires version, size, head, root, timestamp, schema version and sig. | Pass |
| Frozen v1 and registration | Both v1 envelope files are byte-identical to `origin/main`. `schemas/v1/README.md:20`: “The R1/P1 page-boundary revision carries `records-response-v2` and `log-response-v2`”; line 23: “The v1 envelope schemas are byte-frozen and stay valid”. | Pass |
| HTTP endpoint table and legacy parsing | `protocol/registry.md:259` names `records-response-v2.schema.json`; line 261 names `log-response-v2.schema.json`; line 252: “a client validating the frozen v1 records or log envelope treats the v2 `boundary` member as ignorable”. | Pass |
| Signature first, then chain comparison | `protocol/registry.md:364`: “presence and section 2 signature verification”; line 367: “for every page after the first: the boundary MUST be byte-identical to the chain boundary”; mismatch has “no state change”. | Pass |
| First-page rollback/persistence | `protocol/registry.md:372`: “for the first page: the section 5 rollback rules”; line 377: “a higher version is persisted atomically as the new high-water ... BEFORE the page contributes anything”; line 381: “Only a first page that passes steps 1 and 3 changes rollback state”; line 382: “rejected at any step leaves it untouched”. Section 5 line 166 applies the rules to “the chain boundary — the first page's boundary”. | Pass |
| Settled diagnostic precedence | `protocol/registry.md:383`: missing over mismatch; line 385: stale “reported only for a first page”. `tools/validate.py:2409` docstring matches this order; executable oracle checks missing/signature, mismatch, then version/body. The published case shape uses false chain equality for a later mismatching page, documented at `tools/generate-vectors/main.go:667`; no first_page field was required by the settled alternative. | Pass |
| Closed diagnostics/exclusion/direct rollout | `protocol/registry.md:395`–397 table contains exactly `registry_page_boundary_stale`, `registry_page_boundary_mismatch`, `registry_page_boundary_missing`; line 387: “the registry is excluded from the resolution for that operation”; line 399: “There is no legacy-accept mode and no configuration knob”. Spellings agree in prose, table, vectors, generator and gate. | Pass |
| Knob/lock/status tables | No knob or lock key is introduced: §12.1/§12.2 changes and warn-first profiles are inapplicable to the explicit direct R1 rollout. Manager/environments text has no registry client diagnostic list to extend. `protocol/registry.md:409`: “read-only status reports, per trusted registry URL, the persisted high-water (`version`, `log_size`) and whether the last page boundary ... was verified.” | Pass |
| P1 cursor binding/refusal | `profiles/registry-service.md:41`: “Every page of either endpoint carries that boundary on the wire”; line 64: cursor bound to first-page boundary; line 66: “`404 invalid_cursor` (cursor-boundary disagreement), and the service MUST NOT re-evaluate a cursor at a newer boundary”. Line 128 emits immutable captured snapshot; line 239 includes “page-boundary emission and cursor-boundary refusal” in conformance. | Pass |
| Inclusion wording and optional replay purpose | `protocol/registry.md:403`: “The `boundary` is the stated inclusion evidence”; line 405: replay “stays optional” to independently re-derive head, size and root. `profiles/registry-service.md:227`–232 agrees. Compact inclusion proofs remain explicitly out of scope. | Pass |
| Schema cases/manifest | Both valid v2 fixtures contain full signed boundary objects; invalid fixtures omit boundary and equal the existing v1-valid shape. Four entries registered at `conformance/v1/schema-cases/index.json:3803` and `:4248`, and manifest `:3496` and `:3852`; independently validated by the schema gate. | Pass |
| Original client branches | `conformance/v1/vectors/registry-client.json:3` contains 9/9 required cases: fresh advance, equal-same accept, equal-different reject, stale reject, mismatch reject, missing reject, bad-signature reject, and the two revision-2 overlaps. Rejected cases all pin exclusion and no advance. First seven case-body lines are identical to rev1 (91/91 checked). | Pass |
| Revision-2 overlap vectors/gate | `registry-client.json:95`: boundary 7/stored 8 and unequal chain → mismatch/no advance/excluded; `:108`: boundary 9/stored 7 and unequal chain → mismatch/no advance/excluded. `tools/validate.py:2403`–2404 requires both names; `:2465` onward recomputes outcomes through the oracle. Gate is called from main at `:5781`. | Pass |
| Service vectors | `registry-service.json:100` pins `boundary_emitted_on_every_page: true`; `:102` pins byte-identical chain; `:103` contains cursor-boundary-disagreement, 404 invalid_cursor and no newer-boundary reevaluation. | Pass |
| CHANGELOG | `CHANGELOG.md:91`: “R1/P1: records and log pages carry their committed snapshot boundary”; names v2 envelopes, client rejection, missing reporting, direct rollout at line 110 and both Story IDs at line 123. Unchanged as required by the revision-2 brief. | Pass |
| Scope and byte stability | Exactly six patch sections differ from rev1: protocol, oracle, generator, client vector, manifest and rc.9 digest; other 12/12 are hunk-identical. Both registry vector diffs against base have zero removed lines. No implementation, proposals, tags, unrelated files, branch operations or commits. Spec generator/validator changes are the explicitly required conformance tooling. | Pass |

## Independent negative evidence and bounds

Through the real conformance gate `validate_registry_page_boundary_vectors` (wired into `tools/validate.py::main`), independently mutated each of the two new cases three ways: missing case, admitted rejected page, high-water advanced on rejection. **6/6 mutants rejected**. Independently narrowed mismatch handling to cover only non-stale mismatches: **1/1 narrowed oracle mutant rejected**, specifically by `stale-and-mismatch-reports-mismatch`. No candidate source was changed; these were in-memory inputs/oracle replacement in a separate Python process.

These are contract/oracle tests, not service/client integration tests. Cryptographic validity and body identity are represented by fixture predicates in these cases; this review does not claim to have executed the downstream registry implementation, tested filesystem crash atomicity, or established compact cryptographic inclusion. Those are outside this spec leaf. The first-page/later-mismatch distinction uses the documented compact case convention accepted by the rework brief.

## Validation

Commands run independently from the Story worktree using `/bin/zsh`, with `set -o pipefail` and repository venv first on PATH:

```sh
PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" make validate
PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" make regenerate-check
```

Actual `make validate` output (exit 0):

```text
python3 tools/validate.py
validated 62 schemas and 1071 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
.............................................................................................................................................................................................................................................................................................................
----------------------------------------------------------------------
Ran 301 tests in 272.054s

OK
go test ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	0.858s
EXIT=0
```

Actual `make regenerate-check` output (exit 0):

```text
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
EXIT=0
```

Both literal Makefile targets were rerun independently; no producer gate result was substituted. `git diff --check` exited 0. Regeneration left candidate bytes unchanged, rechecked against the attached patch. No source edits, index changes, commits, pushes or LOGBOOK.md edits were made. The review results are persisted in this task-scoped verdict, consistent with the campaign prohibition on LOGBOOK.md edits.

The run goal was queried before recording the verdict: `Active Goal: none (run is not goal-bound)`. No directives were recorded.
