# TASK-260916-55g9dg review verdict — revision 2

Verdict: **changes_requested**. Route to **to-dev**; do not accept CR revision 2.

Reviewed candidate tree `76c9b710a2be05c0ef01d22806b230782fae07c9` against base `1de6f8e12f33212843820be756923b1179063d2c`. All 14 changed files match candidate blobs byte-for-byte, including untracked additions. Downloaded patch SHA256 matches `850852be480cff6ed453e0fd92ae9eb563d5d60c896845b49f32f5bcbb61db63`. Candidate code was not modified. Probes/mutants used Go overlays in ignored `.temp/review-55g9dg`.

## Required corrections

1. **High — remove the forbidden vector-comparison workaround.** `internal/config/environments_conformance_test.go:33-66,224` introduces `postPinKnobs` / `prunePostRevisionKnobs`, deleting the two new keys from actual normalized output whenever the expected vector lacks them. This is precisely the key-subset/root-version workaround forbidden by attached `spec-pin-lag-hold.md`. Runtime rendering does include the knobs; the test hides the difference at rc.11. Restore exact comparison, do not change SPEC_PIN, and report the resulting pin incompatibility to the orchestrator/release owner. The green hosted gate does not supersede this instruction. The new root also fails independently: 41 config subtests fail because E4 provider_directories is unsupported/missing. In particular valid-system-module-waiver fails before E2 validation. Coordinate the sibling integration/root qualification rather than adding a no-op E4 parser or weakening tests. Do not claim the config conformance subset is green.

2. **Medium — reject explicit null waiver lists.** `internal/config/environments.go:281` treats a present null as absence and supplies the empty default. The landed `schemas/v1/manager-config-v2.schema.json:501` requires an array; §12.1 says a list with empty default. Reviewer production-entry probe `Load` with `{"schema_version":2,"skills_root":"/tmp/skills","projects":{},"environments":{"system_module_waivers":null}}` succeeds when it must reject. Remove the nil bypass for this knob and add a committed regression through Load; keep omitted and empty-array inputs accepted. Exact probe and failure transcript are in the evidence bundle.

3. **Medium — expose the admission subset's root-content coverage.** `internal/interop/environments/context_materialization_test.go:170-175` iterates whatever cases the root contains; `.github/ci/platform-cases.tsv:229` remains the general family row with no root-content classification. The five new cases run at the supplied root (5/5 verified), but their absence at rc.11 silently yields no admission cases, not the task-required explicit root-content skip. Add a dedicated root-derived admission subset driver/coverage check, explicit absence skip and corresponding root-content ledger row. Do likewise for new E2 schema-case coverage as required by the task; retain exact published case bytes and avoid false green invalid cases rejected solely by an unrelated E4 field. Do not vendor vectors or suppress unrelated validation failures.

These are actionable implementation/test corrections, so this is to-dev, not blocked. Pin promotion remains owned by the release tasks; this reviewer requests no human approval or scope expansion.

## Per-item review

| Requirement | Assessment and evidence |
| --- | --- |
| Knob defaults, enum, waiver grammar, rendering | Defaults drop/empty and enum/object validation implemented at config/environments.go:138,271,769; rendering at :979,1016; existing generic writer consumes it. Null exception is defect 2. |
| Lock direction / waiver not lockable | config/config.go:65 and config/environments.go:880,1037; TestSystemTransitiveDirection covers locked error overriding machine drop, drop refusal and waiver exclusion. |
| Direct/overlay/dependency/waiver admission | contextmaterialize/admission.go:72,135; required_by edges implement root/overlay directness without new I/O; byte-exact SystemPrompt at contextmaterialize.go:255. 5/5 published admission cases pass. |
| Error refusal and unchanged lock | envprofile/envprofile.go:718,816,1031 calls pre-publish :1209 for install/reinstall/update; manifest read failures return error. Production Install/UpdateWithPolicy tests pass; no-lock and unchanged-lock assertions retained. Reinstall path inspected, not independently adversarially exercised. |
| Always-warn provenance | Existing audit surfacing remains; TestDropRepairWarnsAndFragmentFollowsAdmitted checks provenance warning under drop. |
| Fragment admitted-set semantics | managed.go:1787 consumes admitted output and warnings; marker-driven envfragment system_prompt presence follows it. Both some-admitted and all-dropped production Resolve/fragment tests pass. Launcher extension itself is out of scope. |
| Status posture | status.go:379,404 and cmd/curator/envstatus.go:72 report effective policy and package/path list; warnings remain separate from currency findings. Status tests pass. Manifest-read failures in the new profile summary are skipped there and delegated to home verifier findings; no independent unreadable-store fault injection performed, so that edge is not proven by this review. |
| Linux fixture repair | admission_test.go:150,164,277 uses one NativeHomeOf closure for repair and status. Current assertion and diagnostic assertion remain intact, no blanket filtering. Local macOS pass; Linux pass accepted from exact rev2 hosted log. |
| Rollout / CHANGELOG | CHANGELOG.md:28 describes E2 default drop, opt-in error and knobs/diagnostics. E2 explicitly has no A/B warn-first split; generic brief wording does not impose one here. |
| Conformance and root-content | Materialization 5/5 pass, production SystemPrompt called at interop test :237. Config suite red, test comparison narrowed at old root, absence accounting missing: defects 1/3. Vector warning comparison constructs diagnostic rows from returned dropped modules, not actual manager-emitted warning strings; production envprofile tests separately verify those warnings. |
| Scope / architecture | 14-file delta matches scope; no SPEC_PIN change, no vendored spec, ax or proposal edits. Shared admission helpers reused at assembly and pre-publish; contextresolve need not read module bytes itself. |

## Independent validation

Shell bash, `set -o pipefail`; exported `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1` for conformance runs. Checkout actually reports `23dafa798fa80fc2591ddb287c1c6345e2715b3b`, a later landed revision than the brief's 0da4020; report this provenance accurately. Read environments §3/5.5/5.7/12 and manager §1. No spec files changed.

| Command | Exit / result |
| --- | --- |
| go build ./... | 0 |
| go vet ./... | 0 |
| gofmt -l . | 0; lists five pre-existing .task-board/.resources Go probe files, no candidate file |
| go test ./internal/config/... ./internal/contextresolve/... ./internal/contextmaterialize/... ./internal/contextaudit/... ./internal/envfragment/... ./internal/interop/environments/... -count=1 | 1; config has 41 failed subtests, six other packages pass |
| go test ./internal/envprofile/... -run 'Test(InstallError|UpdateError|DropRepair|DropAllDropped|StatusReportsPolicy|StatusErrorReports|PolicyFromConfigCarries)' -count=1 -v | 0; all eight admission tests plus neighboring policy test pass |
| go test ./cmd/curator/... -run 'TestEnv(Config|Status)' -count=1 | 0 |
| go test ./internal/interop/environments -run 'TestConformanceEnvironmentsMonolithic/system-module' -v -count=1 | 0; all five specified cases execute/pass |
| golangci-lint run ./internal/config/... ./internal/contextresolve/... ./internal/contextmaterialize/... ./internal/contextaudit/... ./internal/envfragment/... ./internal/envprofile/... ./cmd/curator/... ./internal/interop/environments/... | 0; 0 issues |

Full envprofile/CLI suites and Linux/Windows/race execution were not rerun locally. Accepted only as existing evidence from attached exact-revision `TASK-260916-55g9dg_change-request_rev2-validation.log`: run 35181211381 succeeds on hosted lanes, exit 0. This is not evidence that the newer-root config suite passes, nor permission for the comparison workaround. No full remote gate rerun.

## Independent adversarial evidence

Two narrowing mutants attempted, **2/2 killed**, no survivors within this bound. Each uses `go test -overlay <json> ./internal/contextmaterialize ./internal/interop/environments -count=1`, without changing candidate code.

- Narrow error refusal from `len(dropped) > 0` to `> 1`: exit 1; TestSystemPromptErrorRefusesFirst and system-module-transitive-error both fail because one transitive module is admitted.
- Narrow direct dependency classification to root requirers only, removing overlay requirers: exit 1; TestDirectSet and system-module-overlay-direct fail because sysleaf is wrongly dropped.

This is not exhaustive mutation coverage. The additional null-input probe is a failing reviewer-added test against unmutated production Load, not a committed test or a killed mutant.

Producer gaps F1/F2/F4 are reasonable interpretations: applicability to at least one registered adapter for pre-publish checking, fail-closed stale-lock materialization, and policy enforcement before identical-lock update fast path. No spec amendment made. F3 is a real cross-task conformance dependency and prevents claiming the requested schema cases independently green.

## Logbook / lifecycle

Important findings are recorded here and in the board notes. Campaign rules prohibit LOGBOOK.md edits; this artifact is the persistent review record. `task-board spawn goal "$TASK_BOARD_RUN_ID"` reports no goal binding; directives checked, none pending. Attach verdict and transcript bundle before setting to-dev. No accept_cr, commit_ack, code edit, commit, branch change or push performed.
