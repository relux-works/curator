# Review verdict — TASK-260910-1b1ens, revision 1

Verdict: **changes_requested**; route to `to-dev`. No candidate files edited.

Reviewed the campaign rules, producer brief, attached evidence and patch, and security audit R1/P1 (docs/security-audit-2026-09.md:88). Candidate: curator-spec Story worktree for STORY-260910-25yc0h, base origin/main 23dafa7.

## Finding F1 — normative verification order and conformance oracle disagree

`protocol/registry.md:359` orders signature verification, then rollback comparison and advancement, then chain comparison (lines 360–366). It says a higher version “advances the high-water exactly as an accepted snapshot does”. However `tools/validate.py:2420` returns mismatch before checking rollback state at lines 2422–2438. Its docstring at 2411 explicitly describes the opposite order.

This already affects the published `conformance/v1/vectors/registry-client.json:56` case: signed boundary version 8, stored version 7, chain mismatch, but `high_water_advanced: false` at line 62. Under the specified sequential steps the boundary advances the high-water before the page is excluded for mismatch. The gate instead returns `(False, 'registry_page_boundary_mismatch', False)`.

Independent read-only probe also changed only stored_version to 9 on that input. The oracle still returned `(False, 'registry_page_boundary_mismatch', False)`; the normative rollback-first order requires `registry_page_boundary_stale`. Thus passing tests currently protect different behavior from the written protocol.

Required correction: make the oracle, generated vectors, tests, evidence and normative verification order agree. Preserve the settled signature-first/rollback rules; explicitly specify the persistence effect of later chain rejection. Include a combined stale-plus-mismatch case to pin diagnostic precedence and a higher-version-plus-mismatch case to pin persistence. Regenerate artifacts and rerun both gates before another review. This is ordinary producer rework, not a human-only decision or blocker.

## Per-item review

Paths below are relative to the curator-spec candidate root.

| Requirement | Evidence / exact excerpt | Assessment |
|---|---|---|
| Patch equals candidate | Attached and actual `git diff origin/main` byte-identical (`cmp` exit 0); both stable patch IDs `01b3fd10a1465cc1cbfc89acd9c47082098a893e`; diff SHA256 `209e08928268cea32192bbf7dacb4147c7f64a719362aef81542c1cd1367f567` | Pass. New files already indexed; no index mutation needed. |
| §5 rollback state | protocol/registry.md:164: “Page boundaries advance and check the same rollback state”; :171 “A page MUST NOT contribute records or entries” before acceptance | Present; F1 affects conformance consistency. |
| §9 endpoint schemas and compatibility | protocol/registry.md:249 “treats the v2 `boundary` member as ignorable”; endpoint table :256/:258 names records-response-v2 and log-response-v2 | Pass against settled brief. |
| Full boundary, chain equality | protocol/registry.md:348 “carries a REQUIRED”; :350 “all fields including `sig`”; :351 “byte-identical” | Pass. |
| Client ordering and exclusion | protocol/registry.md:359 “verify the boundary signature”; :360 “apply the section 5 rollback rules”; :368 “registry is excluded” | F1. Exclusion itself present. |
| Closed diagnostics | protocol/registry.md:376–378 names exactly `registry_page_boundary_stale`, `registry_page_boundary_mismatch`, `registry_page_boundary_missing`; :380 “No other page-boundary diagnostic exists” | Spellings consistent in text, generator, vectors and validator. |
| Missing boundary, direct rollout | protocol/registry.md:380 “There is no legacy-accept mode”; :381 “no configuration knob”; CHANGELOG.md:109 “Rollout is direct, not warn-first” | Pass; no §12.1/§12.2 knob or lock rows applicable. |
| Service §2 P1 | profiles/registry-service.md:41 “Every page of either endpoint”; :65 boundary differing from cursor returns :66 “404 invalid_cursor”; :66 “MUST NOT” re-evaluate | Pass. |
| Service §5 and §11 | profiles/registry-service.md:128 “immutable snapshot body”; :239 “page-boundary emission and cursor-boundary refusal” | Pass. |
| Inclusion and optional replay | protocol/registry.md:384 “stated inclusion evidence”; :386 “stays optional”; profiles/registry-service.md:230 “re-derives the head, size, and Merkle root” | Pass against settled scope; does not prove dishonest-registry page contents cryptographically. |
| Status posture | protocol/registry.md:390 “read-only status reports, per trusted registry URL”; :391 “persisted high-water (`version`, `log_size`)” and last boundary verification | Required posture content present. No registry diagnostic enumeration found in manager/environment profiles requiring duplication. |
| V2 schemas, frozen v1 | Both v2 schema files :6 require boundary, :10 reference registry-snapshot-v1.schema.json, :12 additionalProperties false. Snapshot schema :6 requires sig. schemas/v1/README.md:20 registers next versions | Pass. Independently compared both v1 schemas with origin/main: byte-identical. |
| Schema cases and registration | All four new cases inspected; valid cases carry full signed boundary, invalid omit boundary. schema-cases/index.json:3803/:3808/:4248/:4253; manifest.json:3496/:3500/:3852/:3856 | Pass: 4/4 registered. |
| Client vectors | registry-client.json:3 page_boundary_cases has 7/7 named branches: fresh, equal same, equal different, stale, mismatch, missing, bad signature | All branches present; mismatch persistence contradicts text (F1). Cases abstract signature/equality as booleans; these are semantic fixtures, not end-to-end crypto/client execution proof. |
| Service vectors | registry-service.json:100 boundary_emitted_on_every_page, :102 chain_boundary_byte_identical, :103 cursor_boundary_cases with 404 invalid_cursor and no re-evaluation | Present. Emission is a declared assertion; no running service was tested in this spec review. |
| Existing vectors frozen | git diff --numstat: client 93 additions/0 deletions; service 10 additions/0 deletions. Removing only new keys yields objects identical to origin/main | Pass. |
| Changelog and stories | CHANGELOG.md:91 “R1/P1”; :98 v2 envelopes; :109 direct rollout; :123 names STORY-260910-25yc0h and STORY-260910-3rvvxh | Pass. |
| Scope | 18 changed files: normative docs, schemas, vectors/index/manifest, generated rc.9 digest, generator and spec validation tools/tests | No product implementation, proposals, tags, or release creation. Tooling edits support generated conformance artifacts; rc.9 change is a derived digest. Reviewer made no candidate edits. |

## Empty curator repository Change Request

The runtime CR repository delta is empty because this task produces a protocol revision in the separate curator-spec worktree, not curator implementation changes. This is expected and not a missing-delivery finding: the attached nonempty spec patch matches the actual 18-file spec candidate exactly. Rejection is based on F1 in that delivered candidate, not on the empty curator CR. Merge/integration has not occurred and is not claimed.

## Independent validation

Shell: /bin/zsh. Exact primary gate command, from candidate root:

```sh
set -o pipefail
PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" make validate 2>&1 | tee /tmp/TASK-260910-1b1ens_validate.log
```

Regeneration is mutating, so to preserve read-only review I copied the exact candidate into a temporary directory, initialized a temporary Git index with that candidate, and ran `make regenerate-check` there. No commit was created and the original worktree/index was untouched. This checks generator idempotence against the candidate, not against main.

A first ad hoc oracle probe used ambient Python and failed to import jsonschema; rerunning with the required repository venv PATH succeeded. This was a probe environment error, not a gate pass. The independent gate uses the required PATH from the start.

Validation outputs follow below.

```text
python3 tools/validate.py
validated 62 schemas and 1071 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
.............................................................................................................................................................................................................................................................................................................
----------------------------------------------------------------------
Ran 301 tests in 492.656s

OK
go test ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	1.847s

make validate exit=0
```

```text
Read-only candidate copy: /var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TASK-260910-1b1ens-regen-u7qsks2f/candidate
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json

exit=0
```

All three make validate gates were independently rerun and passed: 62 schemas / 1071 vector files, 301 Python tests, Go tools tests. No producer validation result substituted for these runs. Both required gates passed, but F1 remains a semantic inconsistency not caught by those gates. git diff --check also exited 0.
