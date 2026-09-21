# Script-worker-v1 delivery reconciliation

Task: TASK-260916-3gcc00. Parent: STORY-260822-2h0v9j. Research date: 2026-09-16.

## Recommendation

**Keep the Story open.** Schema-8 parsing, fail-closed unsupported-policy admission, and conformance classification are delivered. The requested script worker, capability containment, portable controls, native probing/evidence, control-unavailable preflight, and legacy audit warning are missing. Refusing an unsupported policy is correct behavior for this manager, but does not deliver a manager that implements that policy.

Decision: whether to close this Story or create residual implementation tasks. Scope is the board Story description and its AC (script-worker-v1 conformance green on Ubuntu/macOS/Windows with gates/evidence like go-v1). The Story scope field is still a placeholder and adds no requirements. Budget: one research document, no archive attachment, no further research prerequisite; exit at an evidenced recommendation. The consuming production slice is R1 below. No code changes or mutants were made.

## Evidence identities

- Curator HEAD and local main: `559447efe4a9d6f0c5c0f2a9e254cd3cedea883d`; clean worktree before research. Fresh `git ls-remote origin refs/heads/main` returned the same OID. This is a pinned read-only reconciliation; no branch refresh/reset was needed.
- CI spec pin: `87a0d0060bad64ab883d007dcdf35df7485368bf`. Its conformance manifest SHA-256 was independently calculated as `0e195ecd26af2fcb5e5afb5c3fef0ebe60981888b057e8efc60c0b48cf584529`. The workflow identifies this as the spec release tag v1.0.0-rc.11 publishing **protocol 1.0.0-rc.9**; these version numbers must not be conflated.
- All vector data used here came from `git show`/`git archive` of that exact pin, not the spec checkout's moving main. A temporary suite was extracted inside the assigned worktree for narrow tests.
- Code links below resolve against the exact target [Curator tree](https://github.com/relux-works/curator/tree/559447efe4a9d6f0c5c0f2a9e254cd3cedea883d). Normative reference: [manager profile at the spec pin](https://github.com/relux-works/curator-spec/blob/87a0d0060bad64ab883d007dcdf35df7485368bf/profiles/manager.md), script execution sections 3.1–3.6 and audit labels.

## Reconciliation table

Every clause in the Story description is represented; the last row covers its separate platform AC. “Missing” means the requested script behavior is absent, even where reusable build-worker machinery exists.

| Story clause / prerequisite | Assessment | Landed code and test evidence | rc.9 vector mapping and exact gap |
|---|---|---|---|
| Schema-8 script execution surface (prerequisite) | Delivered | `internal/skillspec/parse.go:857` `parseScriptExecution`; `types.go` command fields; `TestSchema8ScriptExecutionPolicyIsParsed`, `TestSchema8ScriptExecutionRejections`, `TestScriptExecutionSurfaceRequiresSchema8`, `TestScriptExecutionSurfaceRejectedOnOtherCommandTypes`; PR #33 | Both `agent-skill-v8` and `csk-skill-v8` contain 66 schema files each, consumed by `TestReleasedSchemaCases`. Examples: `valid-script-worker-enforced`, `valid-script-worker-node-interpreter`, `valid-script-worker-mixed-enforcement`, `invalid-script-worker-null-policy`, `invalid-script-worker-missing-interpreter`, `invalid-script-worker-on-build-command`. Syntax acceptance is not execution support. |
| Manager-owned worker re-execution for script commands | Missing; admission prerequisite delivered | `internal/scriptpolicy/scriptpolicy.go:83` `Admit` unconditionally returns `script_execution_policy_unsupported` for any enforced command. `cmd/curator/main.go:93` dispatches the existing **go-v1 build** worker only. `TestEnforcedScriptCommandIsRefusedAtInstall` drives installation and asserts refusal/no shim; PR #34 | `opt_in_cases` parses accepted shapes but enforced scripts never launch. `TestARefusalPrecedesEveryWorkerSurface` calls `skillspec.Load` and `Admit` directly for node-v1/python3-v1; it is helper-level refusal evidence, not a production worker invocation. No script dispatch, authenticated launch, interpreter identity or script worker lifecycle. |
| Containment derived from declared capabilities, deny by default | Missing | `skillspec` parses declarations; `scriptpolicy.Admit` blocks the script before derivation. Existing build behavior does not establish script behavior. | `capability_derivation_cases`: 0/4 driven as worker behavior. Cases: `all-fields-absent-deny-by-default`, `declared-network-hosts-are-reporting-only`, `declared-exec-is-manager-resolved`, `declared-secrets-remain-identifiers`. Need private roots/working directory, environment and PATH, manager-resolved executable grants, identifiers rather than secret injection, and correctly bounded network reporting. |
| Mandatory portable controls | Missing | No script implementation behind admission; `internal/scriptpolicy/conformance_test.go:50` labels `mandatory_controls` unreachable. | 0/11 script controls applied: fixed process graph, worker identity, interpreter identity, manager environment, manager PATH, offline configuration, private runtime area, explicit streams, inventory application, closed evidence record, worker-domain teardown. Build-worker tests cannot discharge these script obligations. |
| Native-control inventory probing | Missing | Same explicit unreachable classification for `native_control_inventory`. No per-script invocation probe path. | 0/8 inventory controls exercised by script worker across the three specified platforms. Need per-invocation pre-worker-launch probes; honor available, unavailable and Linux host-conditional mechanisms. The eight controls are descendant termination, process count, aggregate memory, file size, inherited handles, descendant exec, filesystem writes, network isolation. |
| Capability-evidence records | Missing | `capability_evidence_record` and `capability_evidence_cases` explicitly classified unreachable. | 0/14 evidence cases driven. Need exactly one `script-capability-evidence-v1` result-only record per invocation, exactly one entry per inventory control, pre-worker-launch timing, truthful probe/status consistency; reject missing/duplicate/extra entries, cached probes, foreign versions/policies and forbidden hardened claims. Keep evidence out of stdout/stderr, markers, receipts, cache keys and conformance claims. |
| `script_execution_control_unavailable` preflight | Missing | The actual error is `script_execution_policy_unsupported`; this intentionally precedes all control checks. `internal/skillcheck/skillcheck.go:61` maps that refusal to an error issue; install calls `skillcheck.Validate` at `internal/install/install.go:918`, with a second writer guard at `internal/install/targets.go:278`. | 0/5 `preflight_cases` driven. Mandatory-control loss at install/invoke must refuse without starting worker; Linux conditional probe available/unavailable and fixed unavailable controls must produce the specified success/evidence behavior. Unsupported-policy refusal is not a substitute for these distinct outcomes. |
| Audit warning class for declared-only legacy script commands | Missing | `internal/scriptpolicy/conformance_test.go:47` explicitly says audit warning classes are not implemented and owned by this Story. `TestDeclaredOnlySchema8ScriptCommandInstallsUnchanged` only proves install compatibility. CLI audit routes via `auditTarget`/`audit.Gate` (`cmd/curator/main.go:1755`); no script label implementation found in production Go code. | 0/4 `audit_label_cases` driven: schema7 and schema8 declared-only need `script-command-declared-only`; enforced needs no such label; declared-network enforced case needs `script-command-unfiltered-declared-network`. No warning-producing entry-point test. |
| Ubuntu/macOS/Windows conformance and gates/evidence like go-v1 | Partially delivered | `.github/ci/platform-cases.tsv:313–316` requires four scriptpolicy test functions on linux/darwin/windows. `.github/workflows/ci.yml:140` runs the platform-case gate. PR #37 added the vector consumption and classification. Hosted checks are green (below). | Gate currently proves schema/identity/classification/refusal, not worker execution. All 12 vector top-level keys classified: 5 consumed (4 metadata plus opt-in), 6 unreachable, 1 not implemented. Only 6/33 named behavioral cases (6 opt-in + 4 derivation + 14 evidence + 5 preflight + 4 audit) are consumed for their specified behavior. Add real entry-point coverage for the remaining 27 cases and controls; replace unreachable classifications. |

Coverage ratios are an inspection of committed consumers, not mutation scores. The schema tests drive `skillspec.Load`; policy vector tests drive helpers; the narrow install tests drive the actual install entry point. No new CLI-level execution coverage is claimed.

## Landing history and later changes

GitHub PR metadata was read directly with `gh`; each merge commit was separately verified as an ancestor of 559447ef (each ancestry command exited 0).

| PR | Merge commit | Delivered contribution |
|---|---|---|
| [#33](https://github.com/relux-works/curator/pull/33) Admit schema-8 module roots and the script execution surface | `62d578c5fceb442df97c1b5bae53542a725754ca` | Schema-8 parser/model support (implementation commit a27f78f); not a script runtime. |
| [#34](https://github.com/relux-works/curator/pull/34) Refuse an enforced script command instead of installing it uncontained | `77aafa09557c27567f0765b3571e6beb597fd933` | Unsupported-policy admission, validation/writer guards and refusal/control tests (a74f1f5). |
| [#37](https://github.com/relux-works/curator/pull/37) Consume the schema-8 families instead of merely publishing them | `a3abcf3468b4854904313295672eef6f7d8826fd` | Schema/vector consumers and explicit script worker/audit gaps (b902023). |

`git log -- internal/scriptpolicy` at the target contains only a74f1f5 and b902023. Inspection of the target code, later parser/skillcheck/ledger history and CLI dispatch found no later implementation replacing the refusal. Later CI commits `4d240ba` and `559447e` add and conditionally gate rose-air. The title “goal's worker policy” in 4d240ba concerns board worker configuration, not delivery of the script execution policy.

## Verification and platform limits

Locally rerun on Darwin x86_64, Go 1.26.0, zsh. Commands ran as standalone processes, without pipes or tee. Both returned exit **0**:

```sh
env CURATOR_CONFORMANCE_ROOT="$PWD/.temp/reconciliation-spec/conformance/v1" go test -count=1 ./internal/scriptpolicy ./internal/skillspec ./internal/skillcheck
go test -count=1 ./internal/install -run 'Test(EnforcedScriptCommandIsRefusedAtInstall|DeclaredOnlySchema8ScriptCommandInstallsUnchanged|ActiveScriptCommandsRefusesEnforcedCommand)$'
```

First command: scriptpolicy 0.799s, skillspec 2.848s, skillcheck 2.085s. Second: install 4.262s. The pin was served explicitly, so the script family was available. These are narrow checks, not a rerun of the landing suite. No expected-red gate or mutant was run; no mutant resistance is claimed.

Accepted external evidence: [CI run 35012468253](https://github.com/relux-works/curator/actions/runs/35012468253), exact target SHA, reported success for Test ubuntu/macos/windows, Gate self-test all three, Race ubuntu/macos, interop, lint and naming. This is GitHub job-status evidence; raw artifact contents and shell exit codes were not independently downloaded/replayed. Rose-air and candidate-conformance were **skipped**. A second run 35016800812 was queued when observed and provides no passing evidence. Thus hosted platform refusal/schema lanes are evidenced; real script execution on every platform remains unverified because absent, and ARM64 adds no result here.

## Exact residual implementation tasks to create

1. **R1 — Implement the script manager/worker invocation path.** Consume the existing schema-8 fields, resolve and bind node-v1/python3-v1 identities, create the fixed authenticated manager re-execution path, route installed enforced shims through it, bind streams/private runtime paths and tear down descendants. Keep unsupported/refusal behavior until enforcement is available. Test through actual install and invoked shim/CLI process boundaries, including forged worker identity and bypass attempts.
2. **R2 — Derive capabilities and enforce mandatory portable controls.** Connect declaration-derived environment, PATH, working directory/private roots, exec grants, offline configuration and secret identifier handling to R1. Drive all four derivation cases and all eleven mandatory controls. Do not promise total network denial or host filtering where the protocol only specifies reporting/configuration.
3. **R3 — Add script native probes, evidence and preflight.** Implement the eight-control script inventory for Linux/macOS/Windows, per-invocation probing, mandatory-control unavailable install/invoke refusals and the closed result-only evidence record. Drive all fourteen evidence and five preflight cases through production invocation, including forged/missing evidence and probe contradictions. Reuse build primitives only with script-specific wiring and proof.
4. **R4 — Emit script audit labels.** Add declared-only warning for schema7 and schema8 and the unfiltered-declared-network warning; avoid declaring enforced scripts declared-only. Drive all four audit vectors through production audit/validation output and preserve existing legacy install behavior.
5. **R5 — Qualify runtime conformance across platforms.** Replace six unreachable and one unimplemented classifications with real consumers; preserve the six opt-in cases. Register real launch/control/evidence/audit tests in the platform ledger, use narrowing negative tests for refusal/attestation gates, capture hosted Ubuntu/macOS/Windows evidence, and explicitly report any unavailable ARM64 lane. This is implementation validation, not another research prerequisite.

R1 is the first vertical production slice; R2/R3 extend it, R4 can proceed independently, R5 validates the resulting integrated behavior. These are recommendations for the orchestrator to create, not newly created board tasks.

## Logbook / anomaly record

2026-09-16 — This Story is not delivered at 559447ef despite green schema-8 CI. The checked-in conformance classification explicitly preserves this distinction and names this Story as audit owner. Do not close on a proxy signal or remove the fail-closed admission without enforcement. This task-scoped record and board note preserve the finding. No `logbook` executable is available on PATH; campaign rules explicitly prohibit LOGBOOK.md edits, so that file was not changed.

## Primary references

- [Admission implementation](https://github.com/relux-works/curator/blob/559447efe4a9d6f0c5c0f2a9e254cd3cedea883d/internal/scriptpolicy/scriptpolicy.go)
- [Conformance consumers and explicit gaps](https://github.com/relux-works/curator/blob/559447efe4a9d6f0c5c0f2a9e254cd3cedea883d/internal/scriptpolicy/conformance_test.go)
- [Production install negative/control tests](https://github.com/relux-works/curator/blob/559447efe4a9d6f0c5c0f2a9e254cd3cedea883d/internal/install/scriptpolicy_test.go)
- [Authoritative rc.9 script vector](https://github.com/relux-works/curator-spec/blob/87a0d0060bad64ab883d007dcdf35df7485368bf/conformance/v1/vectors/script-host-execution-policy.json)
- [Platform ledger](https://github.com/relux-works/curator/blob/559447efe4a9d6f0c5c0f2a9e254cd3cedea883d/.github/ci/platform-cases.tsv) and [workflow](https://github.com/relux-works/curator/blob/559447efe4a9d6f0c5c0f2a9e254cd3cedea883d/.github/workflows/ci.yml)
