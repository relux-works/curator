# TASK-260910-1tvf2t — review verdict, revision 2

Verdict: **accepted**. Record with `accept_cr(..., revision=2, ...)`; route to integrating, not done. No commit acknowledgement.

## Candidate and scope

Reviewed curator-spec Story worktree `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-6bo7ej/worktree`, HEAD `4a2fa3ec428b5021e9e8a80f29fa7fc5a88e4990`. Read campaign rules, producer/rework/reviewer briefs, both patches, producer evidence, round-1 verdict, S2 audit finding, project-management skill and negative-evidence reference.

Attached revision-2 patch and `git diff HEAD` have identical stable patch-id `a1f4a5486c5e7516a6ff87152ed669e0cabb73b8`. Comparison uses HEAD, not moving origin/main. The curator CR delta from `b9a15458c2e0f815ed3c0b197e350e9812a19546` to `4e65e3a02041516b4e81f1961dc188adc31a1c74` is independently confirmed empty. This is the correct repository outcome: the requested normative deliverable is in the separate curator-spec worktree and matching attached spec patch; no curator implementation change belongs to this leaf.

23 candidate paths cover spec prose, schema, vectors/manifest/release metadata, and spec conformance generator/validator/tests. No product implementation, LOGBOOK, proposal 0014–0018, commit, push or branch changes. The campaign explicitly prohibits LOGBOOK edits; this task-scoped verdict records review observations instead.

## Rework findings resolved

| Finding | Evidence and result |
|---|---|
| F1 present/typed inputs | `tools/validate.py:2791` requires a boolean via `bootstrap_bool`; `:2781` rejects boolean/noninteger/negative versions; `:2897` scenario pins require explicit false branches. `tools/test_validate.py:4835` onward adds missing/null/mistyped negatives, `:4865` drives the real main entry, and `:4891` adds internally consistent substitutions. Independent clean-copy main replay refuses all five original removals; exhaustive phase-relevant input sweep and all pairwise same-name substitutions refuse (details below). Pass. |
| F2 per-row status | `profiles/manager.md:1294`: checkpoint is “a current row (informational, --check passes)”; first-use warning “stays current”; refused checkpoint is “an error row” with “non-current”; divergence stays current under advisory and becomes non-current error under strict. `registry-client.json` has boolean `check_current` on all 15 cases, including false on both regression refusals. Oracle recomputes it; flip negatives at `tools/test_validate.py:4910`. No new diagnostic. Pass. |
| F3 cache consistency | `protocol/registry.md:206`: “First fixation comes from the network only”; previously accepted cached bytes “reflect persisted state rather than a fresh fixation”. `:241` diagnostic condition and §8 `:340` agree; `:347` forbids never-accepted cache state fixation, bypass, and known-loss recovery. `:350` says cached reuse “does not re-report” TOFU. `registry-behavior.json:6` pins `checkpoint-or-network-only`; existing TOFU scenario pins warning posture, and summary negatives at `tools/test_validate.py:4940` protect source/posture. Pass. |
| CHANGELOG | `CHANGELOG.md:343`: “except the new once-per-registry registry_bootstrap_tofu warning on checkpoint-less first use.” Pass. |

Revision-1 versus revision-2 file/hunk comparison: only CHANGELOG, manager §10, registry §5/§8, generator main, validator/tests, the two registry vectors, manifest and rc.9 differ. Each maps to F1–F3 or the rollout sentence; manifest/release differences are generated hashes. All 15 old bootstrap case names and input/output values remain identical after removing the newly required `check_current` field. No widening of the authorized rework.

## Standing brief conformance

| Requirement | Quote / location / result |
|---|---|
| Existing signed object and one form | `protocol/registry.md:180`: “path to a signed registry-snapshot-v1”; `:184`: “there is no inline-object form”; references service §6 operator checkpoint, reused without respecification. Pass. |
| Verification and persistence ordering | registry `:189` signature verification “against the pinned keys”; `:195`: “initial high-water BEFORE any network response is accepted”; `:196` first snapshot/page boundary subject to §5. Below/equal-different snapshot is tampered, page uses existing stale-boundary diagnostic. Pass. |
| Protected file and failure posture | manager `:55` “MUST validate ownership, private mutation permissions, containment, and link safety”; `:58` “MUST NOT follow a symlink”; missing/unreadable/malformed/bad signature fails closed naming the path. No absence fallback. Pass. |
| TOFU | registry `:212` “MUST report that fixation once as posture”, URL and pinning hint, durable status source; `:217` later operations do not re-warn. Pass. |
| Rebootstrap | registry `:222` fresh checkpoint, `:225` never lowers/forks state, `:227` regression diagnostic, unchanged state, equal-same no-op, higher atomic advance. Pass. |
| Residual and optional detection | registry `:257`: “monotonic for that client yet divergent across clients”; `:264` “MAY implement”; `:268` enabled same-group registries at same log_size, compare roots; `:273` advisory warning/strict error; `:275` “never changes resolution, never excludes a registry” and “no quorum”. Pass. |
| Cross references | registry §2.1 `:73` same pinned keys; §8 `:340` retained high-water; §10 `:525` publication never bootstraps. registry-service `:287` residual and optional comparison, no service quorum. Existing service §5/§6 untouched. Pass. |
| Knobs and defaults | manager §1 `:46`: bootstrap_checkpoint path and mirror_group identifier; `:50` OPTIONAL/absent, schema-2 only, path length 1–4096, relative resolution; `:64` existing audit_registries lock, no new lock. No environments §12.1 knob/§12.2 lock is introduced. Pass. |
| Status | manager §10 `:1287` curator status/env status, per-registry high-water, source checkpoint/first-use, comparison agree/diverged/not-compared; F2 explicit outcomes above. Pass. |
| Closed sets | Exactly registry_bootstrap_tofu, registry_checkpoint_regression, registry_view_divergence in registry table `:241`; same spellings in prose, vectors, generator/oracle. Exactly bootstrap_checkpoint and mirror_group extend v2 registry shape. No mirrors_of alias or new diagnostic/lock. Pass. |
| Frozen schema | `schemas/v1/manager-config-v2.schema.json:41` uses local registry definition; `:55` closed definition preserves v1 fields and adds two optional members. manager-config-v1 schema, vectors and schema cases have zero diff. Five v2 schema fixtures and v2 vector cases registered. Pass. |
| Vector preservation and gates | Old registry-client top-level values 8/8 and registry-behavior values 5/5 identical to HEAD; diffs insertion-only, preserving old bytes including rollback_state_cases/snapshot_transitions. Fifteen bootstrap branches, exact-name pins, oracle and main registration `tools/validate.py:8286`. Pass. |
| Manifest and generation | New schema fixtures, registry/v2 vectors registered in conformance manifest; rc.9 regenerated. Independent byte-preserving regeneration proof below. Pass. |
| SECURITY and rollout | SECURITY `:497` checkpoint plus TOFU and loss rules, `:509` residual and optional report-only comparison. CHANGELOG `:308` S2 finding, `:341` direct rollout, new warning acknowledged. Warn-first not applicable to settled direct S2 rollout. Pass. |
| Artifacts and scope | Matching spec-patch_rev2 and updated evidence attached to this task; no implementation or LOGBOOK changes. Pass. |

The prior nonblocking scope note remains: the frozen system-config registry shape cannot itself set the new members; locking audit_registries still replaces the user array wholesale. Revision 2 intentionally preserves that pre-existing reviewed limitation per the bounded rework brief. Acceptance does not claim a new system-config extension or implementation conformance.

## Independent validation and bounds

Shell zsh; repository venv selected with `PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH"`. The Makefile validate recipe was executed as bounded component calls, not one >10-minute command. No single-invocation `make validate` exit is claimed. Every component exit and transcript is recorded below; no producer-only results substitute for independent execution.

Regeneration: `make regenerate-check` ran in a disposable candidate copy with literal fixture bytes preserved. Make exited **2** because its final `git diff --exit-code` compares the deliberately uncommitted candidate with HEAD (underlying diff exit 1). This is not relabelled a green literal gate. The generation command completed; before/after SHA-256 maps over **1234/1234 conformance/release files** were identical, changed list `[]`, hash-proof wrapper exit **0**. This independently establishes regeneration idempotence against the candidate, as in round 1.

Review harness observations: an initial copy was made during the existing nofollow test's temporary on-disk mutation, so its generation result was discarded. An archive copy also expanded the repository's deliberate `$Format` fixture (export-subst); that result was discarded. The final copy restores literal original fixture bytes and retains symlinks before regeneration. No candidate correction was made. Existing tests restore their temporary mutations; final patch identity is verified below. Main-entry probes and the focused S2 suite were repeated in the clean disposable candidate so unrelated transient test mutations cannot supply their refusal evidence.

Adversarial scope: phase-relevant inputs, not irrelevant placeholder fields, are required. The independent sweep removes each active input and substitutes null/array/object/invalid string, booleans versus integer versions, and 0/1 versus boolean fields. Invalid container-valued policy inputs may fail with TypeError rather than ValidationFailure; they are not admitted. This is a fixture-validator rejection bound, not production-client execution or cryptographic verification. All 210 ordered substitutions of one published scenario into another scenario's name are rejected; each replacement includes the other scenario's self-consistent expected outputs. Main replay injects at load_json after integrity checks, proving semantic gate reachability, not bypassing manifest integrity.

### Captured passing transcripts

python3 -B tools/validate.py (final worktree replay), exit 0

```text
validated 62 schemas and 1108 vector files
```

python3 -B -m unittest test_validate.RegistryBootstrapVectorTests (clean copy), exit 0

```text
.......................................
----------------------------------------------------------------------
Ran 39 tests in 49.037s

OK
```

test_validate classes 1–8, exit 0

```text
...................................................................................
----------------------------------------------------------------------
Ran 83 tests in 87.218s

OK
```

test_validate classes 9–16, exit 0

```text
........................................................................................................................................................................
----------------------------------------------------------------------
Ran 168 tests in 227.449s

OK
```

test_validate classes 17–22, exit 0

```text
..............................................................................................................
----------------------------------------------------------------------
Ran 110 tests in 355.176s

OK
```

remaining four test modules, exit 0

```text
..............................................................................
----------------------------------------------------------------------
Ran 78 tests in 115.093s

OK
```

go test ./tools/..., exit 0

```text
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	1.510s
```

independent clean-copy semantic probe, exit 0

```text
main removal checkpoint-signature-invalid-first-use signature_valid 1
main removal rebootstrap-signature-invalid-ignored signature_valid 1
main removal checkpoint-first-network-equal-different-tampered candidate_same_body 1
main removal divergence-different-sizes-skipped same_log_size 1
main removal divergence-detected-advisory roots_equal 1
typed/removal refused 573 / 573
same-name substitutions refused 210 / 210
```

Total Python coverage: **478/478 tests passed** (83 + 168 + 110 + 39 + 78), matching the independently counted test inventory. The S2 clean-copy replay replaces the earlier overlapping focused run as authoritative evidence. All Makefile validate components passed. Some independent slices initially overlapped; the main-entry semantic probe and S2 suite were repeated in isolation from the mutating nofollow tests.

Final worktree `git diff HEAD` SHA-256: `7c7324ea611047dd7f880c71052b699c8d475c5c0cbeee5cccad2e7bc84d965f`, identical to initial capture and attached revision-2 patch. Final status has the same 23 submitted paths; no reviewer files or residual fixture changes. `git diff --check HEAD` exit 0.

Run goal queried before verdict: **none (not goal-bound)**. No directives. Acceptance is bounded to the normative spec and conformance artifacts; implementation is a separate task. No outstanding rework finding.

### Probe source

```python
import sys,copy,contextlib,io,json
sys.path.insert(0,'tools')
import validate as v
p=v.SUITE/'vectors/registry-client.json'; original=v.load_json; data=original(p)
probes=[('checkpoint-signature-invalid-first-use','signature_valid'),('rebootstrap-signature-invalid-ignored','signature_valid'),('checkpoint-first-network-equal-different-tampered','candidate_same_body'),('divergence-different-sizes-skipped','same_log_size'),('divergence-detected-advisory','roots_equal')]
for name,field in probes:
 d=copy.deepcopy(data);del next(c for c in d['bootstrap_cases'] if c['name']==name)[field]
 v.load_json=lambda path: d if path==p else original(path)
 with contextlib.redirect_stdout(io.StringIO()),contextlib.redirect_stderr(io.StringIO()): result=v.main()
 print('main removal',name,field,result);assert result==1
v.load_json=original
count=0
for i,c in enumerate(data['bootstrap_cases']):
 fields=['phase']
 if c['phase']=='compare': fields+=['group_size','policy','same_log_size','roots_equal']
 else:
  fields+=['prior_state','checkpoint_configured']
  if c['checkpoint_configured']:
   fields+=['signature_valid']
   if c['signature_valid']: fields+=['checkpoint_version','candidate_same_body','first_network_version' if c['phase']=='bootstrap' else 'stored_version']
 for f in fields:
  bad=[None,[],{},'invalid']
  if type(c[f])==bool:bad += [0,1]
  elif type(c[f])==int:bad += [True,False,1.5,-1]
  for val in ['REMOVE']+bad:
   d=copy.deepcopy(data)
   if val=='REMOVE':del d['bootstrap_cases'][i][f]
   else:d['bootstrap_cases'][i][f]=val
   try:v.validate_registry_bootstrap_vectors(d)
   except (v.ValidationFailure,TypeError):count+=1
   else:raise AssertionError((c['name'],f,val))
print('typed/removal refused',count,'/',count)
count=0
for i,c in enumerate(data['bootstrap_cases']):
 for other in data['bootstrap_cases']:
  if c['name']==other['name']:continue
  d=copy.deepcopy(data);d['bootstrap_cases'][i]=dict(other,name=c['name'])
  try:v.validate_registry_bootstrap_vectors(d)
  except v.ValidationFailure:count+=1
  else:raise AssertionError((c['name'],other['name']))
print('same-name substitutions refused',count,'/',count)

```

### Exact bounded unittest slice commands

Run from the spec `tools` directory with the venv PATH above (each exit 0):

```sh
python3 -B -m unittest test_validate.WireSemanticValidationTests test_validate.AssuranceRelationalValidationTests test_validate.RepositoryDescriptorIdentityTests test_validate.ManagerLifecycleValidationTests test_validate.BuildDriverGoldenSuiteTests test_validate.SharedFixtureMarkerTests test_validate.WorkflowRegenerationScopeTests test_validate.EnvironmentVectorTests
python3 -B -m unittest test_validate.EnvPassthroughVectorTests test_validate.StoreBoundaryVectorTests test_validate.SourceSignersVectorTests test_validate.CodexSeedVectorTests test_validate.ContextVersionVectorTests test_validate.ContextDetectorVectorTests test_validate.SnapshotAcquisitionVectorTests test_validate.ShellHookTrustVectorTests
python3 -B -m unittest test_validate.WriteNofollowVectorTests test_validate.UmbrellaProviderVectorTests test_validate.ManagerConfigVectorTests test_validate.SystemConfigV2SchemaTests test_validate.RegistryPageBoundaryVectorTests test_validate.RegistryCheckpointVectorTests
python3 -B -m unittest test_implementation_coverage test_release_gate test_verify_release_commit test_verify_release_merge_policy
```
