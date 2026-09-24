# TASK-260916-1xib1x revision 1 — independent review

Verdict: changes_requested; route to `to-dev`.

Reviewed candidate tree `eb0d7e104d32db6651c5ec6fba4a122355eee37a`, base `d275e2d9d23afdde75abecdc51971b663047ffd5`. Tested an isolated archive of that exact tree under the assigned Story worktree; product working files were not changed. Mutations were confined to the disposable archive and restored from original bytes.

## Required rework

1. **Per-command audit identity is missing** (`internal/scriptpolicy/labels.go:78`, `internal/audit/audit.go:168`, Report at line 75, persisted verdict at line 329). The pinned manager profile section 7 explicitly requires the audit record to carry the effective execution-policy identity or absence for EVERY script command. The binding revision-1 review note repeats this requirement. AuditLabelsForCommands drops commands with no warning; Report and the stored verdict contain no per-command policies. Consequently an enforced no-host skill produces only `app: audit clean`, with no command or policy identity. The existing CLI row explicitly asserts that output (`cmd/curator/auditlabels_test.go:121`). A mixed skill likewise loses its enforced no-host commands from the record. Add a per-command record independent of warning eligibility, including exact `script-worker-v1` or explicit absence, expose it through the production audit record, and test mixed and no-warning commands without adding erroneous warnings. Reproduction: run TestCLIAuditEmitsScriptLabels/schema8-enforced-script; inspect its exact-output assertion and Report/storeCachedFindings. The separate draft-source ScriptPolicy list is a manager-wide unsupported-policy string, not a per-command record, and does not satisfy this requirement.

2. **Global-install label coverage is absent** (`internal/install/global.go:189`, `internal/install/auditlabels_test.go:89`). All four new install rows call project installation only. The binding review note requires global scope because global.go is changed. Existing TestGlobalInstallUsesManifestLocaleAndAuditGate injects a gate and uses a non-script fixture; it does not check these labels. Add the four vector rows and declared-only-with-hosts negative row through actual Global with its default production gate. Assert each audit warning on the same message line as the exact label, distinct from the skillcheck warning, plus absence of errors and unchanged legacy shims. See mutation evidence below.

## Independent verification

Host Darwin x86_64, Go 1.26.0; commands launched by zsh, no output pipelines for Go checks. Spec extracted from exact pin `87a0d0060bad64ab883d007dcdf35df7485368bf`.

- `CURATOR_CONFORMANCE_ROOT=<pinned archive>/conformance/v1 go test -count=1 ./internal/scriptpolicy ./internal/audit ./internal/skillcheck`: exit 0 (1.227s, 2.550s, 1.165s).
- `go test -count=1 ./cmd/curator ./internal/install -run 'Test(CLIAudit.*ScriptLabels|ScriptAuditLabelsAtInstallEntry|DeclaredOnlyWithHostsInstallsWithOnlyDeclaredOnly|DeclaredOnlySchema8ScriptCommandInstallsUnchanged)$'`: exit 0 (CLI 1.977s, install 110.199s).
- `go build -o ../TASK-260916-1xib1x-curator ./cmd/curator`: exit 0.
- Docs explain both labels accurately; CHANGELOG Added present. Production project/global changes pass Commands only; legacy materialization code unchanged. Warning labels do not appear in scriptworker applied-control code on inspection.
- Both warning classes bypass fail_on and remain warning outputs on all audit decisions; strict, pinned, revoked and blocked tests passed in the audit package.

## Independent mutation attacks

Each run used `-count=1`; mutated source restored after every run. CLI mask: TestCLIAuditEmitsScriptLabels$. These are reviewer reruns, not reliance on the producer's mutant table.

| Mutation | Outcome |
|---|---|
| Also label enforced commands declared-only | Killed, exit 1; both enforced CLI rows fail |
| Narrow network warning to more than one host | Killed, exit 1; single-host network CLI vector fails |
| Narrow declared-only warning to commands lacking win_path | Killed, exit 1; both declared-only CLI rows fail |
| Append script warnings to errors | Killed, exit 1; all three labelled CLI rows exit 1 |

The third mutant's log filename says schema7-only-declared; its actual mutation is the win_path predicate stated above, not schema filtering.

Global wiring mutant: replace only `global.go` Subject.Commands with nil (project route remains intact). `go test -count=1 ./internal/install -run 'Test(ScriptAuditLabelsAtInstallEntry|DeclaredOnlyWithHostsInstallsWithOnlyDeclaredOnly|GlobalInstallUsesManifestLocaleAndAuditGate)$' survived, exit 0, 93.247s. This is a measured bypass of the global audit-label route, not a product failure in the unmutated wiring. Attack ratio 4/5 killed, 1/5 survived. The survivor establishes the coverage gap in finding 2; the full install suite was not run against this mutant.

## Reused landing evidence and bounds

Independently queried CI run https://github.com/relux-works/curator/actions/runs/35698876152: success. headSha `b88dcea45064b85eca037283e4eceaa24dbb3ac5` resolves locally to the exact candidate tree. Job-status evidence: Ubuntu/macOS/Windows tests, gate self-tests, Ubuntu/macOS race, lint, interop and naming succeeded. Rose-air and candidate suite skipped. Raw hosted artifacts were not replayed/downloaded. The full landing suite was not rerun manually; its runtime-owned result was reused. No ARM64/rose-air passing claim.

Coverage: 4/4 named vector shapes driven in CLI, project install, validation and audit gate tests. Global install: 0/4 label-specific committed rows. Vector helper reads the exact 4-row set and classification now says consumedAtAuditEntry; this does not by itself prove global coverage or complete per-command records. Runtime refusal/attestation implementation is outside this audit-label delta; no new claim of full runtime mutation coverage.

## Logbook note

2026-09-22: green exact-tree CI does not discharge the missing per-command record or global label-path tests. No logbook executable is available; campaign rules prohibit LOGBOOK.md edits. This task-scoped artifact plus the board note persist the findings.
