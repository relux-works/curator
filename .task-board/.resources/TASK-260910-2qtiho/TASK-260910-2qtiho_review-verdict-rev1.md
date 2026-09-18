# TASK-260910-2qtiho — review verdict, revision 1

Verdict: **changes_requested**, route **to-dev**. Reviewed curator-spec HEAD e8b53a003256433761cebce6080d6a955d777f25 plus its uncommitted candidate. No candidate edits, index edits, commits, or LOGBOOK edits performed. Run goal query reports no active goal.

## Findings requiring correction

### F1 — Scenario pinning admits removal of required branches (high)

`tools/validate.py:7599` pins the MCP refusal only to `profile-install`, declarations present, and absent environments; it does not pin the effective hardened posture. In memory, I changed `refusal-mcp-allowlist-empty-with-declarations` to explicit `security_posture: permissive`, recomputed expected values/provenances/diagnostics/rows, and set outcome `proceeds`. `validate_security_posture_vectors` accepted it. The case now warns instead of refusing, but retains its negative-case name. This is precisely the campaign rule-7 forbidden replacement.

An additional exhaustive whole-case substitution probe tried each of the other 16 published cases under each of the 17 names: 269/272 substitutions rejected, 3/272 accepted. Two accepted replacements erase distinct required branches:

- `revision-A-default-permissive-status` replaced with `locked-value-beats-explicit`: ceases to test advisory profile defaults.
- `revision-B-default-hardened-flip-install` replaced with `schema1-machine-is-permissive`: ceases to test the hardened default flip.
- `revision-A-permissive-warning-once-install` replaced with `unreachable-registry-permissive-warns`: accepted, but still tests the named warning; not independently a defect.

Correction: pin every required scenario to schema version, effective posture, precedence conditions and the inputs distinguishing its branch, including absence of overriding locks where necessary. Add same-name, internally consistent replacement negatives for the demonstrated holes. Re-run all gates. The measured substitution ratio is a bounded corpus test, not exhaustive proof over arbitrary inputs. The attached probe reproduces the MCP hole without editing candidate bytes.

### F2 — Both status commands do not expose the brief's full gate inventory (medium)

The settled brief requires `curator status` and `env status` to carry the posture and per-gate effective values/provenance for the listed gates. `profiles/manager.md:1397` closes curator status to four rows; line 1407 says **“No other `curator status` posture row exists.”** `protocol/environments.md:2832` prescribes eight environments rows. Thus curator status excludes eight of the twelve requested gates even when environments capability exists, and the env-status contract does not clearly require the four manager rows. The vectors encode that split.

Correction: require the complete applicable twelve-gate inventory on both commands when environments capability exists, using a shared closed vocabulary, and pin both command outputs in vectors. Keep the schema-1/no-environments case explicitly bounded. The `shipped` provenance for revision/always-on gates is a reasonable explained extension and is not itself a finding.

## Per-deliverable review

| Requirement | Evidence and assessment |
|---|---|
| Exact patch/worktree | `git diff HEAD | git patch-id --stable` and attached spec patch both yield `425adf2a41b41dc8011c16bac2125cb8164bf730`. Base is e8b53a0, not moved origin/main. PASS. |
| Manager §1 knob and lock | `profiles/manager.md:46`: “one closed top-level knob”; line 80: “only in the direction of `hardened`”. Schema v2 enum and lock set agree. PASS. |
| Manager §7 effective policy | `profiles/manager.md:1095` table: audit/registry strict, non-empty allowlists, explicit-null refusal, transitive error, signers true, unreachable registry error. “Precedence is lock, then explicit machine value, then the profile default” with three exceptions. PASS. |
| Two labelled revisions | `profiles/manager.md:1122`: “Revision A (this release)” and “Revision B (a later release)”; permissive warning exactly once with named migration hint. PASS. |
| Registry §4 residual and notice | `protocol/registry.md:135`: “revocation is network-dependent”; prominent install/update notice names registry/artifacts and warning/error posture. PASS against settled brief. |
| Registry §8 cross-reference | `protocol/registry.md:262`: “The grace widens the advisory-policy residual”. Existing grace retained. PASS. |
| SECURITY text | `SECURITY.md:40` recommends hardened defaults; line 519 names network-dependent advisory residual and availability tradeoff. PASS. |
| Environments interaction | `protocol/environments.md:523` escalates empty MCP list with declarations; line 2611 refuses explicit null; diagnostic tables updated; §12.1 paragraph at 2952 references profile without changing S4 default. PASS. No new environments lock key is introduced, so §12.2 needs no added key. |
| Posture report | Manager §10 and environments §12 use closed row/provenance names and non-current checks for mandatory refusals. F2: full inventory not required on both commands. |
| Schemas | `schemas/v1/manager-config-v2.schema.json:46` enum permissive/hardened, revision-A default permissive; `system-config-v2.schema.json:18,43` lock member and hardened-only enum. Frozen v1 schemas unchanged. PASS. |
| Generator/schema cases | Four manager and three system schema cases; positive values and invalid values/types/direction; generator property/lock tests updated. PASS. |
| Vector family/manifest | 17 cases in `conformance/v1/vectors/security-posture.json`, manifest registration at 4392. Cases read individually: both defaults, explicit/lock precedence, three refusals, permissive warning, registry warning/error, schema1, status contradiction, shipped revisions. F1: required branches removable despite green validation. |
| CHANGELOG | `CHANGELOG.md:10`: “S1/S3”, both revisions, residual, exact diagnostics. PASS. |
| Scope | 24 paths: prose, schemas, conformance/manifest/index/release digest and spec generators/validator/tests. No runtime implementation, proposals, LOGBOOK, or pre-existing vector edits. `git diff HEAD --check` passes. |
| Existing vectors | `git diff HEAD --name-only -- conformance/v1/vectors` returns only new `security-posture.json`; pre-existing vectors unchanged. PASS. |
| Empty curator CR | Exact curator base-to-candidate diff is empty, as declared. Appropriate repository boundary: this leaf delivers a curator-spec candidate and task-scoped artifacts, not curator implementation. Empty curator delta is not a rejection reason and does not establish spec integration. |

## Independent validation

No producer test result is substituted for reviewer execution. Candidate mutation probes were in memory. Regeneration runs in a disposable tracked-file copy with its candidate staged as the comparison index, preserving the read-only review worktree. Command: `make -C /tmp/TASK-260910-2qtiho-regen regenerate-check`, exit **0**:

```
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
```

MCP replacement probe, venv Python, exit **0** (unexpected acceptance):

```
SURVIVED: same-name MCP refusal replaced with permissive warning/proceeds; validator accepted
```

One exploratory substitution command initially used system Python and failed to import jsonschema; rerun with the repository venv completed the 272 substitutions reported above. This environment failure is not counted as a candidate gate failure.


Validation orchestration: `/bin/zsh`, `set -o pipefail`, from the candidate:

```sh
PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" make validate
```

The first leg completed: `validated 62 schemas and 1103 vector files`. I terminated the monolithic Python leg deliberately before the command time bound (make exit 2, `Terminated: 15`); this invocation is NOT claimed as a passing make invocation. Replayed the entire Python discovery suite as bounded shards in disposable candidate copies, with the 11-scenario repeated-main test partitioned 4/4/3 without changing validator logic. The attached shard harness records the exact selection. No producer-only validation is relied on. Go leg independently run from the candidate:

```
go test ./tools/...
ok  github.com/relux-works/curator-spec/tools/generate-vectors  1.218s
```

Go exit 0. Final candidate `git diff HEAD` SHA-256 and attached patch both equal `723cfaa5fb6edb767cba2982494627db80a86f47bfc6b1c082dcc8d69176f3e6`; no review edits persisted.

### Completed Python replay transcript

Command for each shard: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin/python -B /tmp/TASK-260910-2qtiho-shard.py <shard>` from its disposable candidate copy. Ordinary shards partition all 421 tests other than the one repeated-main test; heavy shards partition that test’s 11 scenarios. Coverage: 422/422 discovered Python tests, 11/11 repeated-main scenarios. Every shard exited 0.

Shard 0:
```text
Tests 141
.............................................................................................................................................
----------------------------------------------------------------------
Ran 141 tests in 239.307s

OK
```

Shard 1:
```text
Tests 140
............................................................................................................................................
----------------------------------------------------------------------
Ran 140 tests in 276.090s

OK
```

Shard 2:
```text
Tests 140
............................................................................................................................................
----------------------------------------------------------------------
Ran 140 tests in 251.058s

OK
```

Shard heavy0:
```text
Heavy scenario shard 0 count 4 of 11
Tests 1
.
----------------------------------------------------------------------
Ran 1 test in 77.647s

OK
```

Shard heavy1:
```text
Heavy scenario shard 1 count 4 of 11
Tests 1
.
----------------------------------------------------------------------
Ran 1 test in 77.667s

OK
```

Shard heavy2:
```text
Heavy scenario shard 2 count 3 of 11
Tests 1
.
----------------------------------------------------------------------
Ran 1 test in 62.735s

OK
```

All three make-validation legs passed independently in the bounded replay; regeneration proof exited 0. These green results do not override F1/F2. Findings are recorded in task notes and this outcome; campaign prohibition on LOGBOOK.md edits respected. Exactly one verdict: changes_requested / to-dev.
