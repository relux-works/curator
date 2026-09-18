# TASK-260918-2mglq0 review verdict — revision 1

Verdict: accepted. S1+S3 specification Revision 4; CR-TASK-260918-2mglq0-1.

## Candidate and evidence boundary

Reviewed the exact curator-spec Story worktree at HEAD `1e73c0301f2dd95efbf93802707da9619f3b2b09`. Its `git diff HEAD` is byte-identical to `TASK-260918-2mglq0_spec-patch_rev1-rebased.patch`, SHA256 `2c9f9efc678488e13bb16ec7c26cef511b01441c7ddf8a6834cf6dbd809046d1`, stable patch ID `f1b6213925a8004a322cb7623aa2b1f89748ed78`.

The curator CR diff from `640a9df19ba94295cce059562101599996761f4b` to `57d2578ea9d1644d8a8931fdcdfbcb8ce4e15545` is independently verified empty, and its worktree is clean. This is correct: the authorized deliverable is the separate curator-spec worktree and task-scoped spec patch, not curator implementation. Acceptance does not claim the spec has landed.

Read campaign rules, both briefs, producer evidence/patch, prior accepted revision-3 verdict, prior brief/evidence, and audit findings S1/S3/E3/E6. Prior review supports unchanged carried content; all validation results below are independently rerun against the new candidate.

## Per-item assessment

| Requirement | Result and exact evidence |
|---|---|
| Producer start state and own-base patch | PASS. Reconstructed both patches on `1ca4b3d`: prior stable patch ID `43f07deedbb0fed7270b0c81454b4fb94b4baedc`; producer ID `46aacb6df526ae462e3b430bec5876ff82ba34a8`. Both apply cleanly. Producer evidence records verification before editing; historical ordering is evidence-backed, not independently observable now. |
| Exactly the authorized correction | PASS. Full reconstructed tree comparison finds only CHANGELOG, manager, environments, security-posture vector, validator, validator tests, manifest and rc.9 release changed (8 files). All other bytes identical. Subtracting codex_seed inputs, codex-seed output rows, inventory member and count wording yields an identical JSON vector across all 17 cases. |
| Fixed thirteen-row inventory | PASS. profiles/manager.md:1445: “closed to exactly these thirteen, in this order”; :1459: “the shipped section 7.4 revision, `A` or `B` (environments §7.4)” with `shipped` provenance. Position 10, immediately after update-confirmation. :2898 and CHANGELOG.md:29 also say thirteen. No remaining `twelve` in affected surfaces. |
| Environments inventory exclusion | PASS. protocol/environments.md:2995 says “thirteen posture rows”; :3001–3002: “`codex-seed` (the per-home `codex_seed_record` native-server rows above are not posture rows and stay unchanged)”. :3005 closes the inventory. |
| Outputs and both shipped revisions | PASS. security-posture.json:70 inventory member; 32/32 applicable command outputs carry position-10 codex-seed with shipped provenance and value equal to pinned input. All 17/17 scenarios pin the input, including B in posture-rows-flipped-revisions. Schema-1 intentionally remains header plus four manager rows, no env posture section. |
| Validator and rule 7 | PASS. tools/validate.py:8661 ordered member; :9249 admits only A/B; :9410 derives row; :9461 checks thirteen ordered gates; :9564 registers gate in main. tools/test_validate.py:3930 onward adds omission, misplacement and internally consistent A→B scenario-rewrite negatives. Independent main-entry probes below refuse all 5/5 mutations. |
| E6 reference | PASS. manager.md:1460 retains “`enforced` (environments §4)”; environments.md:748 says “The contract extends to the declared” path directory and specifies boundary checks. Always-on extension requires no additional shipped-revision row. |
| Landing rebase | PASS. 19/24 per-file patch IDs identical; exactly the five permitted exceptions: manager, manager schema, manager generator, manifest, rc.9 release. S2 manager additions (2 blocks) and generator additions (3 blocks) retained verbatim; registry paragraph precedes posture paragraph, five registry cases precede four posture cases. Schema is exactly main plus producer security_posture property, with the reviewed description combining both extensions and their defaults. Generated pins verified by regeneration. schema-cases/index.json patch ID unchanged. |
| Existing closed sets and rollout | PASS. manager.md:68–72 retains closed permissive/hardened knob; :1129 “Precedence is lock, then explicit machine value, then the profile default”; :1132 three refusals; :1144 “Revision A (this release)” warning and :1152 “Revision B (a later release)” flip. Prior accepted diagnostics, §12.1 interactions, §12.2 lock set, schemas and cases retained; no new knob, diagnostic or lock in this correction. |
| S3 and docs consistency | PASS. protocol/registry.md:138 and SECURITY.md:536 retain “revocation is network-dependent”; notice and offline-grace rules preserved. CHANGELOG.md:31 names “the thirteenth row is the E3 codex-seed shipped revision” and its position/exclusion. |
| Scope and byte identity | PASS. 35/35 pre-existing vector files equal current main (including E3, E6 and S2); only security-posture is candidate vector content. Exactly 24 intended patch paths; no untracked files, implementation, proposals, or LOGBOOK changes. `git diff HEAD --check` succeeds. |
| Artifacts and integration | PASS. Producer patch/evidence are task-scoped outcomes. This verdict and supplemental replay evidence are attached before accept_cr. Acceptance routes to integrating; producer doc-writer/implementer owns subsequent integration. |

## Validation method and limits

Shell zsh, `set -o pipefail`; PATH prepends `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin`, PYTHONDONTWRITEBYTECODE=1. Independently executed all three Makefile validate legs: validator, complete unittest discovery in bounded sequential shards, and Go tests. This is not a claim of one uninterrupted `make validate` invocation. No producer green substitutes for any test here.

Validator and Go tests run in the actual Story worktree. Python tests run in an exact temporary candidate reconstructed from HEAD plus frozen patch; tests can temporarily mutate fixtures, so they never run in the review worktree. Raw export-subst fixture bytes restored from Git, avoiding archive expansion. Regeneration runs in another exact temporary copy with its own Git index recording the candidate: `make regenerate` and `make regenerate-check` exit 0 without touching the Story index. An initial invocation from the temporary parent directory returned “No rule to make target regenerate”; corrected to the candidate root before the recorded successful gates. This was a reviewer invocation error, not a candidate failure.

Probes call the real validate.main(), substituting only the loaded security-posture JSON via load_json; every preceding main check remains active. This isolates semantic rejection from manifest digest rejection. Coverage is structural conformance validation and normative consistency; downstream manager runtime behavior is not established by these tests.

No product findings or unresolved deviations. No LOGBOOK edit: campaign rules explicitly prohibit it; review observations and the invocation correction are recorded here instead.

## Independent transcripts

Commands: `python3 tools/validate.py`; `python3 -B /tmp/2mglq0-review/shards.py {0,1,2,3,heavy}` (each separately); `go test ./tools/...`; `make regenerate`; `make regenerate-check`; `python3 -B /tmp/2mglq0-review/probe.py`.

### validate.log

```text
validated 62 schemas and 1117 vector files
```

Command exit 0.

### shard0.log

```text
Discovery 538 Selected 135 shard 0
.......................................................................................................................................
----------------------------------------------------------------------
Ran 135 tests in 83.621s

OK
```

Command exit 0.

### shard1.log

```text
Discovery 538 Selected 134 shard 1
......................................................................................................................................
----------------------------------------------------------------------
Ran 134 tests in 84.167s

OK
```

Command exit 0.

### shard2.log

```text
Discovery 538 Selected 134 shard 2
......................................................................................................................................
----------------------------------------------------------------------
Ran 134 tests in 92.218s

OK
```

Command exit 0.

### shard3.log

```text
Discovery 538 Selected 134 shard 3
......................................................................................................................................
----------------------------------------------------------------------
Ran 134 tests in 68.061s

OK
```

Command exit 0.

### heavy.log

```text
Discovery 538 Selected 1 shard heavy
.
----------------------------------------------------------------------
Ran 1 test in 118.630s

OK
```

Command exit 0.

### go.log

```text
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	0.648s
```

Command exit 0.

### regenerate.log

```text
go run ./tools/generate-vectors -root .
```

Command exit 0.

### regen-check.log

```text
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
```

Command exit 0.

### probe.log

```text
omit exit 1 validation failed: security-posture case revision-A-default-permissive-status curator status rows are not the closed ordered set
misplace exit 1 validation failed: security-posture case revision-A-default-permissive-status curator status rows are not the closed ordered set
nonclosed-output exit 1 validation failed: security-posture case revision-A-default-permissive-status curator status rows do not follow their inputs
nonclosed-input exit 1 validation failed: security-posture case revision-A-default-permissive-status codex-seed revision is not closed
consistent-rewrite exit 1 validation failed: security-posture case revision-A-default-permissive-status must carry shipped_revisions.codex_seed == 'A'
main-entry refusal coverage 5/5; untouched on-disk digests; in-memory vector substitution at load_json only
```

Command exit 0 (probe harness; each intended rejection exits 1).

### identity.log

```text
old 43f07deedbb0fed7270b0c81454b4fb94b4baedc 0000000000000000000000000000000000000000
producer 46aacb6df526ae462e3b430bec5876ff82ba34a8 0000000000000000000000000000000000000000
candidate f1b6213925a8004a322cb7623aa2b1f89748ed78 0000000000000000000000000000000000000000
Producer changed: ['CHANGELOG.md', 'conformance/v1/manifest.json', 'conformance/v1/vectors/security-posture.json', 'profiles/manager.md', 'protocol/environments.md', 'release/1.0.0-rc.9.json', 'tools/test_validate.py', 'tools/validate.py']
Rebase differing patch sections: ['conformance/v1/manifest.json', 'profiles/manager.md', 'release/1.0.0-rc.9.json', 'schemas/v1/manager-config-v2.schema.json', 'tools/generate-vectors/manager_config.go']
candidate sha256 2c9f9efc678488e13bb16ec7c26cef511b01441c7ddf8a6834cf6dbd809046d1
```

### scope.log

```text
All 17 vector cases unchanged after subtracting codex-seed addition
Schema union exact except reviewed combined description and producer security_posture member
Pre-existing vectors byte identical: 35 / 35
Candidate patch byte-identical to attached rebase: True
```

Python discovery coverage: 538/538 tests, partitioned 135 + 134 + 134 + 134 + 1. The heavy method retains all 11 original subcases. Main-entry codex-seed refusal coverage: 5/5 selected negative shapes.

Final checks: candidate patch still byte-identical, no untracked files, whitespace check passes, curator CR delta empty and curator worktree clean. Run goal: `Active Goal: none (run is not goal-bound)`. No directives recorded. Conditional changes-requested checklist item is satisfied as not applicable to the accepted branch.
