# TASK-261001-2kqi6r — spec-muse-environment-adapter: review revision 1

Verdict: ACCEPTED, subject to the recorded acceptance transaction. No revision-blocking findings.

Review identity: CR-TASK-261001-2kqi6r-1 revision 1; base add50233fe64d29d1ba64e71b0538788efc79657; candidate tree f72405d2d23c3c27a206262c604999c9e26dc21a. Reviewed all 57 changed paths (+2908/-13). A Git-blob comparison of all 1612 candidate files found zero working-tree mismatches. Fresh origin main advertised the base OID during review. No source code, worktree index, commits, or branches were changed by this reviewer.

## Swept surfaces

| Surface | Finding and evidence |
| --- | --- |
| Adapter registry and layout | Accepted. Decision 0010 §3 and environments §7.1 declare all four XDG parents with home/config, home/data, home/state, home/cache and MUST NOT replace HOME. Opencode remains a config-parent adapter; Muse explicitly retargets data/state/cache as well. Root context is unverified and not admitted; the skills location is measured but discovery remains unverified. Only managed-home is admitted. |
| Credentials and seeds | Accepted. Decision 0010 §7, Decision 0017, and environments §7.4 specify config/muse/auth.json as a file-link to the native effective XDG config root captured before managed overrides. Optional settings.json and trust.json seeds come from the root profile snapshot, never overlays or credentials. Reads distinguish failure from absence. Every resolve checks the link and target. Missing links with healthy native targets can be repaired; forked files, wrong targets, and dangling targets refuse repair without overwriting credential bytes. |
| Evidence and isolation | Accepted. Independently read issue 117 body and all comments via gh JSON; comments=[] and no refresh result. A surviving link on four runs is not called a refresh probe. The spec handles write-through and temp+rename, records per-tree lock/concurrent-refresh uncertainty, and names the foreign-personal-context isolation gap. No live Muse, login, refresh, skills-discovery, or root-context probe is claimed by this review. |
| Permission ownership | Accepted. Decision 0018 and environments §10.1 record exec --yolo and serve --disable-sandbox/--trust-workspace plus separate session/start approvalMode. Measured provider facts remain distinct from normative policy and agents-management mapping ownership. Native silence, fleet locks, and unsupported-mapping refusal are retained. |
| Schemas and release metadata | Accepted. v3 is needed because released v1/v2 close the environment enum and allow only one home variable. The new v3 admits Muse/four parents, forbids HOME and unverified channels, and retains v2 permission rules. Normalized structural comparison confirmed all old adapter constraints are retained. Released-schema immutability gate passes. No existing schema changes in the CR. rc.13 metadata differs only at candidate_protocol_pin.manifest_sha256 and downstream_consumption.required_manifest_sha256. |
| Generated fixtures and validator | Accepted. environments_muse.go is called by the real generator CLI; preservation tests delete copied Muse artifacts before driving that CLI. 36 v3 schema cases (8 valid/28 invalid), one marker case, and layout/fragment/marker fixtures are generated. Manifest adds 41 files and changes only the existing schema-case index digest. Validator production dispatch is tested; no runtime implementation is implied. |
| Scope and hygiene | Accepted. CHANGELOG entry is under Unreleased, no LOGBOOK.md or unrelated changed paths, whitespace and Go formatting checks pass. Priority brief explicitly forbids LOGBOOK.md, so review findings and limitations are persisted here and in board notes instead. |

## Link-state review

Coverage: 16/16 declared combinations, eight states times bare resolve and repair. Live and write-through states emit fragments. Missing-link bare resolve is stale; repair relinks only with a readable regular native target. Temp-rename fork, retargeted link, and dangling target emit no fragment (bare resolve: stale; repair: credential conflict). Metadata-unreadable and target-unreadable emit no fragment and never repair. All cases preserve credential bytes. The status member records the observed liveness classification, including detached before a successful missing-link repair.

Bound: this is specification consistency coverage, not an exhaustive filesystem state space or manager execution. Empty/nonempty directories and concurrent in-session refresh races are not among these 16 cases; generic §7.4 directory rules remain applicable, and race freedom is explicitly unknown.

Independent adversarial checks ran in memory without editing sources: narrowing the validator to temp-rename-fork only produced 14 failing subtests out of the 16 emission mutations; removing the production Muse dispatch caused the production-entry test to fail (1/1). Both deliberately weakened gates were detected. The in-memory modules were restored and discarded.

## Reviewer-run validation

Validation used an isolated shared clone with the candidate tree loaded into its index and files (no commit), preserving release tags. Thus regenerate-check compared regenerated output with the exact candidate, not with the producer's uncommitted base. Python dependencies were installed from requirements-dev.txt in /tmp/muse-review-python-261001.

| Command/check | Actual exit/result |
| --- | --- |
| Ambient python3 -B tools/validate.py | 1: jsonschema missing; this attempt did not validate anything |
| Temporary Python dependency installation | 0 |
| Temporary Python -B tools/validate.py | 0: validated 73 schemas and 1294 vector files |
| make regenerate-check | 0: no generated diff from the candidate index |
| go test -count=1 ./tools/... | 0: generator suite passed, 5.748s |
| go vet ./tools/... and gofmt -l on changed Go sources | 0; no diagnostics or formatting paths |
| git diff --check add50233 f72405d2 | 0 |
| Narrowed-gate / missing-dispatch adversarial harness | 0: both expected failure counts observed, no harness errors |
| Full Python discovery | 130: deliberately interrupted after about nine minutes, before the headless command bound, during test_release_gate.py:281 fixture copying. Incomplete; not a full-suite pass. No failure marker was observed before interruption. |
| Temporary Python -B -m unittest test_validate.ReleasedSchemaImmutabilityTests test_validate.EnvironmentVectorTests test_validate.WireSemanticValidationTests test_muse (cwd tools) | 0: 58/58 relevant tests passed in 29.386s |
| Final temporary candidate git diff --exit-code and untracked inventory | 0: no tracked changes or untracked files |

All passing checks above were run personally; producer logs were inspected for context, not substituted for reviewer execution. Broader discovery is explicitly not accepted as passing evidence; the required validator/regeneration gates and the bounded relevant tests are green. The conditional nonacceptance-routing checklist item is not applicable because this verdict accepts revision 1. Reviewer goal query returned: Active Goal none (run is not goal-bound).

## Routing

Accept revision 1 with task-board accept_cr and route to integrating. This is acceptance of the spec candidate, not a claim that Muse runtime integration has landed. No commit_ack, done transition, source commit, or integration is performed by the reviewer.

Routing transport note: the first accept_cr attempt exited 1 before acceptance because the canonical ssh://git@github.com/relux-works/curator-spec authority read failed with publickey denial. A fresh HTTPS query to the same GitHub repository returned refs/heads/main at add50233fe64d29d1ba64e71b0538788efc79657. A process-scoped Git URL transport rewrite from ssh://git@github.com/ to https://github.com/ also resolved the canonical authority URL to that same remote HEAD. Acceptance is retried with this transport mapping only; no repository/global configuration or authorized repository identity is changed, and no cached/local authority is substituted.
