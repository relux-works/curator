# TASK-260910-1tvf2t — review verdict, revision 1

Verdict: **changes_requested**, route to `to-dev`. Candidate code was not modified.

## Candidate and scope

Reviewed curator-spec Story worktree `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-6bo7ej/worktree`, against HEAD `4a2fa3e`, together with the attached producer evidence and spec patch. Read campaign rules, producer and reviewer briefs, S2 audit finding, project-management skill and negative-evidence reference.

Attached patch and `git diff HEAD` have identical stable patch-id `5b60fabaa8e7ad40df2d570f3d19c47cf987e17c`. Worktree diff SHA-256: `cbcaa195db32df638e2b0a46f8b2e934ca96b2d049f219188a484b4dc393ccd5`.

The curator CR delta is empty (independently checked against the supplied base and candidate tree). That is appropriate for this leaf: its deliverable lives in the separate curator-spec worktree and attached spec patch, not in curator implementation files. Empty curator delta is not a finding and does not justify acceptance by itself.

23 changed paths are spec prose, schema, conformance fixtures/manifest, release metadata, and spec generator/validator/tests. No product implementation, LOGBOOK.md, proposal 0014–0018, branch, commit or push changes were made. The explicit campaign prohibition on LOGBOOK edits governs; findings are recorded here and on the board instead.

## Findings requiring rework

### F1 — scenario pins treat missing evidence as a negative input

`tools/validate.py:2917`, `:2948`, `:2973`, `:3040`, `:3056` use `is True` / `is not True` predicates without requiring a present boolean. Five independent probes removed the discriminating input from a required scenario, leaving its name and expected result unchanged. All five malformed scenarios were accepted:

| Required scenario | Removed input | Gate result |
|---|---|---|
| checkpoint-signature-invalid-first-use | signature_valid | accepted |
| rebootstrap-signature-invalid-ignored | signature_valid | accepted |
| checkpoint-first-network-equal-different-tampered | candidate_same_body | accepted |
| divergence-different-sizes-skipped | same_log_size | accepted |
| divergence-detected-advisory | roots_equal | accepted |

Measured refusal coverage for these probes: **0/5**. With all five removals injected through `load_json` at the actual `tools/validate.py main()` entry, the validator printed `validated 62 schemas and 1108 vector files` and returned **0**. This is an in-memory probe; no candidate file was edited. The check is called from main, but absent evidence is interpreted as false. A scenario without signature-validity evidence no longer pins the invalid-signature branch; a scenario with no root-equality evidence cannot establish divergence.

Correction: require each phase's discriminating inputs to exist with exact types, and require explicit `False` where false is the scenario discriminator. Add missing/null/wrong-type negatives and self-consistent scenario replacements, including a main-entry negative, so these five probes refuse. Preserve the existing valid vectors and 15 scenario names. The measured 0/5 is the bound of this probe set, not a claim about all validator coverage.

### F2 — status downgrades a checkpoint refusal to a warning

`protocol/registry.md:237` declares `registry_checkpoint_regression` severity **error**. `profiles/manager.md:1293-1296` immediately includes a refused checkpoint, then says `--check` treats strict divergence as non-current and **“every other bootstrap row as a warning row.”** This includes the checkpoint-regression error and also successful checkpoint rows. The two normative surfaces disagree.

Correction: state status severity/outcome for successful checkpoint bootstrap, TOFU warning, checkpoint-regression error, and advisory/strict divergence explicitly; do not blanket-downgrade the refusal. Add a conformance case/pin for the refusal's status mapping. No new diagnostic is needed.

### F3 — offline first-use wording contradicts itself

`protocol/registry.md:335` says **“Caching and offline grace never bootstrap ... rollback state”**, but lines 338–340 require TOFU posture when the **“fixing view came from ... the cache.”** Section 5 at lines 205–206 and the diagnostic table at line 236 describe first fixation from the network only. A client cannot tell whether an otherwise valid cached snapshot may establish first-use high-water.

Correction: make §5, §8 and the diagnostic condition agree about cache-backed first fixation. Preserve the settled rule that caching cannot bypass checkpoint validation, lower high-water, or silently recover known-lost state; specify the permitted first-use path consistently and pin its posture in a vector.

## Brief conformance matrix

| Requirement | Evidence / result |
|---|---|
| Signed existing object, one form | `protocol/registry.md:179-185`: “path to a signed registry-snapshot-v1”; “there is no inline-object form”. Pass. |
| Pinned verification, persist before network | `protocol/registry.md:189-202`: signature against pinned keys; “initial high-water BEFORE any network response is accepted”; first snapshot/page subject to §5. Pass prose. |
| Protected file, failure distinct from absence | `profiles/manager.md:55-61`: MUST validate ownership/permissions/containment/link safety; missing/unreadable/malformed/signature failure is configuration error. Pass prose; F1 affects vector input handling. |
| TOFU once plus status | `protocol/registry.md:205-212`: MUST report once, URL and pinning hint; later operations do not re-warn. Status source closed to checkpoint/first-use at manager:1289-1290. Pass except F3. |
| Rebootstrap refuses regression | `protocol/registry.md:217-227`: fresh checkpoint; lower/equal-different refused, unchanged state, equal-same no-op, higher atomic advance. Pass. |
| Equivocation residual | `protocol/registry.md:252-257`: monotonic per-client yet divergent views; protection proves nothing about another client. Pass. |
| Optional comparison / closed group / policies | `protocol/registry.md:259-276`: MAY implement, shared mirror_group, same log_size, MUST compare roots, warning/advisory and error/strict, never changes resolution/excludes, no quorum. Pass. |
| Service and verification cross-references | registry-service §10 line 287 points to §5.1; registry §2.1 line 73 and §10 line 514 point to checkpoint/high-water rules. Existing service §5/§6 object reused. Pass. |
| Offline cross-reference | registry §8 lines 335–340 added; F3 requires correction. |
| Config members/defaults/locks | manager §1 lines 46–66: bootstrap_checkpoint path 1–4096, mirror_group identifier, optional/absent, schema-2 only, audit_registries lock. No new environments §12.1 knob or §12.2 lock key. See scope note below. |
| Status | manager §10 lines 1287–1296 lists high-water/source/comparison, but F2 requires correction. |
| Schema versioning | manager-config-v2 `$defs/registry` retains old fields and adds exactly bootstrap_checkpoint/mirror_group, additionalProperties false. manager-config-v1 has zero diff. Pass. |
| Vectors / old cases | 15 bootstrap cases added; old top-level values of registry-client and registry-behavior compare equal to HEAD, including rollback_state_cases and snapshot_transitions; diffs are insertion-only. Five new schema fixtures plus manager-config-v2 cases registered. Pass content; F1 prevents gate acceptance. |
| Manifest / rc.9 / generation | generator updates manifest and rc.9. Independent regeneration produced no changed bytes across 1114 conformance/release files. Pass reproducibility. |
| Security text | SECURITY.md:495–517 covers checkpoint, once-only TOFU, surviving state, residual, policy severity and report-only/no quorum. Pass. |
| Changelog / rollout | CHANGELOG.md:308–355 names S2 and direct rollout. No warn-first requirement for this settled direct rollout. Minor wording correction advisable: “no behavior change without a checkpoint or a mirror group” overlooks the new TOFU warning. |
| Scope / evidence | producer evidence and matching spec patch attached; no forbidden implementation/LOGBOOK edits. Pass. |

Scope note for the producer: manager §1 says the existing system lock covers the new members, but system-config-v2.schema.json:44 still references the frozen v1 registry shape, so a system file cannot set either new member. Locking audit_registries replaces the user array wholesale (manager §1 lines 91–100). Clarify the claim/limitation within the authorized scope; do not silently widen frozen system schemas. This is not an external blocker or a request to reopen the settled bootstrap design.

## Independent validation

Shell: zsh, `set -o pipefail`; Python selected with `PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH"`. Commands run from the spec worktree unless noted. The Makefile's validate recipe is executed in bounded component calls rather than one long shell invocation.

- `python3 tools/validate.py`: exit 0; `validated 62 schemas and 1108 vector files`.
- From tools: `python3 -B -m unittest test_validate.RegistryBootstrapVectorTests`: exit 0; `Ran 22 tests in 26.941s`, `OK`. Includes the producer's scenario-replacement/main-entry negative.
- `go test ./tools/...` from spec root: exit 0; `ok github.com/relux-works/curator-spec/tools/generate-vectors 5.299s`. An earlier invocation mistakenly from tools failed because tools/tools does not exist; corrected at the root above.
- `make regenerate-check` in a disposable candidate copy (original remains read-only): make exit 2, underlying `git diff --exit-code` exit 1, showing the expected uncommitted candidate delta against HEAD. This is NOT reported as a green literal gate. Hashes before/after regeneration match for **1114/1114** conformance/release files; changed list empty. The Makefile check compares with the base, not the uncommitted candidate, so idempotence supplies the substantive generation proof.
- Adversarial main-entry probe: unexpectedly exit 0 with five discriminating inputs removed (F1); review failure even though normal validation is green.

Additional component results:

- `python3 -B -m unittest test_implementation_coverage test_release_gate test_verify_release_commit test_verify_release_merge_policy` (tools directory): exit 0; `Ran 78 tests in 303.466s`, `OK`.
- First test_validate slice (classes 1–8 in source order): exit 0; `Ran 83 tests in 104.211s`, `OK`.

- Second test_validate slice (classes 9–16 in source order): exit 0; `Ran 168 tests in 268.757s`, `OK`.
- Third test_validate slice (classes 17–22 in source order): exit 0; `Ran 110 tests in 366.338s`, `OK`.

Total independently rerun Python tests: **461/461 passed** (83 + 168 + 110 + 22 + 78). All three Makefile validate components passed; no literal single-invocation `make validate` exit is claimed. No producer-only test evidence substitutes for this independent run. The known failing adversarial semantic probe remains F1.

Final candidate diff hash rechecked unchanged: `cbcaa195db32df638e2b0a46f8b2e934ca96b2d049f219188a484b4dc393ccd5`.

The regeneration hash-comparison wrapper exited 0. The underlying Makefile comparison did not: its expected-red result is retained above rather than relabelled as a passing `make regenerate-check`.

### Reproduction of F1 (in memory; original files unchanged)

Run with the repository venv Python from the spec worktree:

```python
import copy, sys
sys.path.insert(0, "tools")
import validate
path = validate.SUITE / "vectors" / "registry-client.json"
mutant = copy.deepcopy(validate.load_json(path))
for name, field in [
    ("checkpoint-signature-invalid-first-use", "signature_valid"),
    ("rebootstrap-signature-invalid-ignored", "signature_valid"),
    ("checkpoint-first-network-equal-different-tampered", "candidate_same_body"),
    ("divergence-different-sizes-skipped", "same_log_size"),
    ("divergence-detected-advisory", "roots_equal"),
]:
    del next(c for c in mutant["bootstrap_cases"] if c["name"] == name)[field]
original = validate.load_json
validate.load_json = lambda p: mutant if p == path else original(p)
print(validate.main())  # observed 0; must refuse this malformed corpus
```

This probe tests semantic validation after the data read. It does not claim a bypass of manifest byte-integrity validation; the published vector bytes are unchanged. The defect is the scenario gate accepting a corpus which no longer carries the required discriminating evidence.

### Exact unittest slice commands (tools directory)

```sh
python3 -B -m unittest test_validate.WireSemanticValidationTests test_validate.AssuranceRelationalValidationTests test_validate.RepositoryDescriptorIdentityTests test_validate.ManagerLifecycleValidationTests test_validate.BuildDriverGoldenSuiteTests test_validate.SharedFixtureMarkerTests test_validate.WorkflowRegenerationScopeTests test_validate.EnvironmentVectorTests
python3 -B -m unittest test_validate.EnvPassthroughVectorTests test_validate.StoreBoundaryVectorTests test_validate.SourceSignersVectorTests test_validate.CodexSeedVectorTests test_validate.ContextVersionVectorTests test_validate.ContextDetectorVectorTests test_validate.SnapshotAcquisitionVectorTests test_validate.ShellHookTrustVectorTests
python3 -B -m unittest test_validate.WriteNofollowVectorTests test_validate.UmbrellaProviderVectorTests test_validate.ManagerConfigVectorTests test_validate.SystemConfigV2SchemaTests test_validate.RegistryPageBoundaryVectorTests test_validate.RegistryCheckpointVectorTests
```

Each slice used a bounded subprocess (550-second timeout). The reviewer retained the live session until all commands finished and attached evidence before the verdict transition.

### Captured slice transcripts

```text
...................................................................................
----------------------------------------------------------------------
Ran 83 tests in 104.211s

OK

EXIT:0
```

```text
........................................................................................................................................................................
----------------------------------------------------------------------
Ran 168 tests in 268.757s

OK

EXIT:0
```

```text
..............................................................................................................
----------------------------------------------------------------------
Ran 110 tests in 366.338s

OK

EXIT:0
```

Run goal queried before recording verdict: no active goal (run is not goal-bound). Route: changes_requested → to-dev. No acceptance, done transition, or commit acknowledgement was supplied.
