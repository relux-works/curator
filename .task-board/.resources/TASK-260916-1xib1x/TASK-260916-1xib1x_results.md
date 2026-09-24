# TASK-260916-1xib1x results — R4 script audit labels

Story STORY-260822-2h0v9j. Worktree HEAD `d275e2d9` (R1–R3 landed) plus the
uncommitted R4 change described here.

## Label rules (manager profile §7, CI pin 87a0d006)

| Command shape | Labels |
|---|---|
| script, no `execution_policy` (schema 7, schema 8 w/o field) | `script-command-declared-only` |
| script, `execution_policy: "script-worker-v1"`, empty `network` | (none) |
| script, `execution_policy: "script-worker-v1"`, non-empty `network` | `script-command-unfiltered-declared-network` |
| build / system | (none) |
| script, unknown policy (refused at admission) | (none) |

Enforced is never declared-only; declared-only with hosts carries only
declared-only (the network label is about enforcement, not declaration).
Both classes are always `warn` in every mode, never errors, never install
refusals; neither is a finding, neither is subject to `fail_on`, neither is
cached with the verdict.

Vector note: `audit_label_cases` names policy + labels but no interpreter
or hosts, so the two enforced rows would be identical manifests without the
case name. Every consumer binds enforced rows to the closed `python3-v1`
interpreter and gives `schema8-enforced-unfiltered-network` one declared
host (`audit-label.example.com`) while its twin declares none. Confirmed
against the vector rows: the name is the distinguishing input.

## Row table — all four `audit_label_cases` at production entries

| Vector case | audit Gate / GateReadOnly | skillcheck Validate | install (messages + gate) | `curator audit` CLI |
|---|---|---|---|---|
| schema7-script → [declared-only] | `TestGateEmitsScriptAuditLabels` | `TestValidateEmitsScriptAuditLabels` | `TestScriptAuditLabelsAtInstallEntry` | `TestCLIAuditEmitsScriptAuditLabels` |
| schema8-declared-only-script → [declared-only] | same | same | same | same |
| schema8-enforced-script → [] | same (silent) | same (silent) | same (silent) | same (`audit clean`) |
| schema8-enforced-unfiltered-network → [unfiltered] | same | same | same | same + `--json` |

Negative rows: enforced ⇒ no declared-only (`TestAuditLabelsEnforced…`,
`TestGateEnforcedIsNeverDeclaredOnly`, `TestValidateEnforced…`, install +
CLI assertions); declared-only with hosts ⇒ no network label
(`TestAuditLabelsDeclaredOnlyWithHosts…`, `TestGateDeclaredOnly…`,
`TestValidateDeclaredOnly…`, `TestDeclaredOnlyWithHostsInstalls…`).
Legacy install preserved: declared-only keeps the ordinary shim with no
sidecar, enforced keeps the native launcher + sidecar; install `Status`
stays `ok` and CLI audit exits 0 with empty stderr.

Production call sites: `audit.Gate`/`GateReadOnly` (`internal/audit`),
`skillcheck.Validate` (`internal/skillcheck`, via install `validateNodes`
and `skill check`), install audit gate (`internal/install`), `auditTarget`
(`cmd/curator`).

## Mutant table (all killed, no survivors)

| Mutant | Attack | Killed by |
|---|---|---|
| M1 label enforced declared-only too | broaden declared-only | unit + vector consumer (both enforced subtests) + gate + validation |
| M2 enforced always unfiltered (drop host check) | broaden network label | unit + vector consumer (`schema8-enforced-script`) + gate + validation |
| M3 omit declared-only | drop label | unit + vector consumer (both declared-only subtests) + all gate paths + validation incl. pre-existing tests |
| M4 validation severity error | warn→error | all validation label tests + install `schema7-script` (install refuses) |
| M5 gate appends labels to errs | warn→error | all gate label tests on the allow path |

## Ratio line

`script-host-execution-policy.json`: 12/12 top-level keys consumed
(`audit_label_cases` reclassified `not implemented` → `consumedAtAuditEntry`;
no key remains unimplemented or unreachable). Behavioral rows: 4/4 audit
cases driven at four production entries each.

## Bounds

- Local spec root used for vector tests (checkout main); `audit_label_cases`
  verified byte-identical to the CI pin 87a0d006 rows.
- Narrow suites rerun in this session (real exit codes, `zsh`): `audit`,
  `skillcheck`, `scriptpolicy` full; `install` audit+script subset
  (`TestAudit|TestScript|TestDeclaredOnly|TestActiveScript|TestActiveEnforced|TestEnforced|TestGlobalEnforced`,
  385s); `cmd/curator` audit subset; `envprofile` full. `go vet`,
  `gofmt -l cmd internal`, and `golangci-lint run` clean on touched
  packages. Full landing suite runs once via the handoff runtime (remote
  CI), not locally.
- No platform-specific code: rows run everywhere via the standard matrix;
  hosted windows/macos/linux evidence comes from the landing CI run.
  Platform-ledger registration is R5 scope (this task adds no ledger rows).

## Files

- `internal/scriptpolicy/labels.go`: `LabelDeclaredOnly`,
  `LabelUnfilteredDeclaredNetwork`, `AuditLabels`, `AuditLabelsForCommands`
- `internal/audit/audit.go`: `Subject.Commands`, `scriptAuditWarnings`,
  gate appends on every decision incl. allow/block/require-pin/revoked
- `internal/skillcheck/skillcheck.go`: `scriptAuditLabelIssues` (warning,
  code = label text)
- `cmd/curator/main.go`, `internal/install/install.go|global.go`: pass
  `node.Spec.Commands` into audit subjects
- Tests: `scriptpolicy/labels_test.go` + `conformance_test.go`
  (`TestScriptAuditLabelCases`), `audit/auditlabels_test.go`,
  `skillcheck/auditlabels_test.go`, `install/auditlabels_test.go`,
  `cmd/curator/auditlabels_test.go`; two pre-existing `skillcheck` tests
  updated for the new warning (now first of 3 / 1 of 1, still no errors)
- Docs: `docs/troubleshooting.md` (§Script audit labels),
  `docs/cli.md` (audit section), `CHANGELOG.md` (### Added)

## Revision 2 (rework 1: per-command record + global coverage)

Verdict rev1 `CHANGES_REQUESTED` fixed on the revision-1 tree; no
checkout/clean/stash. Base still `d275e2d9`, work uncommitted.

### 1. Per-command audit identity (manager profile §7)

`scriptpolicy.AuditPoliciesForCommands` reports every script command's
entry — command name → declared policy verbatim (`script-worker-v1` for
enforced, empty for absence, verbatim for unknown/refused policies) — in
lexical order, independent of warning eligibility; build/system commands
omitted. `audit.Report` carries `ScriptPolicies` (fresh on every path
incl. cache hits), the stored verdict persists `script_policies`,
`curator audit` prints `audit info: <skill>: command '<name>'
execution_policy=<script-worker-v1|(none)>` lines (informational, never
warnings) and carries `script_policies` under `audit --json`. The
enforced no-warning CLI row now asserts its exact record line plus
`audit clean`, exit 0, empty stderr — the bare-clean assertion is gone.

Record row table:

| Case | Unit | audit gate/verdict | CLI text | CLI JSON |
|---|---|---|---|---|
| mixed (declared-only + enforced no-host) | `TestAuditPoliciesForCommandsCoversEveryScriptCommand` | `TestScriptPolicyRecordsCoverEveryCommand`, `TestReportCarriesScriptPoliciesOnEveryPath`, `TestStoredVerdictPersistsScriptPolicies` | `TestCLIAuditMixedSkillRecordsBothPolicies` (exactly 1 warning) | — |
| enforced no-host (no warnings) | same | `TestNoWarningCommandKeepsItsRecordWithoutWarnings` (Gate silent) | `TestCLIAuditEmitsScriptLabels/schema8-enforced-script` (exact output) | `TestCLIAuditJSONCarriesScriptLabels` (identity entry) |
| declared-only absence | same + unknown-policy-verbatim | `TestFormatScriptPolicyRendersIdentityOrExplicitAbsence` | all declared-only rows assert `(none)` | `TestCLIAuditJSONRecordsExplicitAbsence` |
| vector rows | `TestScriptAuditLabelCases` now also pins each case's record entry (identity/absence) | — | — | — |

### 2. Global-install coverage (default production gate)

`TestScriptAuditLabelsAtGlobalInstallEntry` (4 vector rows) +
`TestDeclaredOnlyWithHostsGlobalInstallsWithOnlyDeclaredOnly` run
through `install.Global` with NO `AuditGate` override. Each labelled row
asserts an audit-gate line containing both `audit warning:` and the
exact label on the same message line, a distinct `warning: <label>`
skillcheck line, no label in errors, `Status ok`, and the legacy
launcher in `global/bin` (shim without sidecar for declared-only,
native launcher + sidecar for enforced). The enforced no-host row
asserts total label silence.

### Mutant table, revision 2 (all killed, no survivors)

| Mutant | Attack | Killed by |
|---|---|---|
| R2-M1 drop enforced (no-warning) commands from `AuditPoliciesForCommands` | revive the rev1 record bug | unit + vector consumer (both enforced subtests) + all audit record tests + all CLI record rows — FAIL at all 3 levels |
| R2-M2 record `Policy: ""` always | drop the enforced identity | unit + vector consumer enforced rows + CLI enforced row — FAIL |
| R2-M3 `FormatScriptPolicy` renders absence verbatim (empty) instead of `(none)` | narrow explicit absence | format test + CLI declared-only/mixed rows — FAIL |
| R2-M4 `global.go` `Subject.Commands` → nil (reviewer's rev1 survivor, global-only) | bypass the global audit-label route | both new global tests FAIL with `no global audit-gate warning carries declared-only` while the skillcheck line survives — kill is specific to the global audit wiring |
| R2-M5 gate appends script warnings to errs (rev1 M5 rerun) | warn→error | `TestGateEmitsScriptAuditLabels` — FAIL |

All mutants restored from saved bytes after each run; `diff` confirms
the tree carries only the intended delta.

### Verification (real exit codes, `zsh`, Darwin x86_64, Go 1.26.0)

Spec served from the EXACT CI pin: `git archive
87a0d006.../conformance/v1` extracted to a temp root (4.8M);
`audit_label_cases` additionally verified byte-identical between the pin
and the local spec checkout main (`05053cd7`).

- `CURATOR_CONFORMANCE_ROOT=<pin> go test -count=1
  ./internal/scriptpolicy/ ./internal/audit/ ./internal/skillcheck/
  ./internal/envprofile/`: exit 0 (0.63s, 1.45s, 2.12s, 336.36s).
  Vector consumer `TestScriptAuditLabelCases` + section classification
  run unskipped, 4/4 subtests pass.
- `go test -count=1 ./cmd/curator/ -run 'TestCLIAudit'`: exit 0 (2.16s).
- `go test -count=1 ./internal/install/ -run
  'TestScriptAuditLabelsAtInstallEntry|TestDeclaredOnlyWithHostsInstallsWithOnlyDeclaredOnly'`:
  exit 0, 89.21s (project regression green).
- `go test -count=1 ./internal/install/ -run
  'TestScriptAuditLabelsAtGlobalInstallEntry|TestDeclaredOnlyWithHostsGlobalInstallsWithOnlyDeclaredOnly'`:
  exit 0, 88.85s (final green after mutant restore; 81.00s on first run).
- `gofmt -l cmd internal`: clean (one alignment fix applied).
  `go vet ./...`: exit 0. `golangci-lint run` on the five touched
  packages: 0 issues.
- Full landing suite runs once via the handoff runtime, not locally.

### Ratio / bounds update

`script-host-execution-policy.json`: still 12/12 keys consumed; the
`audit_label_cases` consumer additionally pins the per-command record
entry per case. Global label coverage: 0/4 → 4/4 vector rows + negative
row at the production global entry. No platform-specific code; hosted
evidence comes from the landing CI run. Platform-ledger registration
remains R5 scope.

## Revision 3 (rework 2: backfill old cached verdicts)

Verdict rev2 `CHANGES_REQUESTED` with ONE finding, fixed on the
revision-2 tree; no checkout/clean/stash. Base still `d275e2d9`, work
uncommitted. rev2→rev3 touches only `internal/audit/audit.go` (the
backfill) and `internal/audit/auditlabels_test.go` (the rows).

### The defect

`auditSubject` returned on a findings-cache hit before the only call
that persists `script_policies`, and the cache identity/version was
unchanged — a valid pre-R4 cached verdict never gained the per-command
record, so the stored audit record stayed incomplete indefinitely while
the in-memory report and CLI were correct.

### The fix (backfill, not version bump)

On a WRITABLE cache hit, `backfillScriptPolicies` rewrites the stored
verdict with the cached findings unchanged and the current per-command
entries, so the file ends up as a fresh audit would have written it.
The rewrite happens only when the stored verdict lacks a non-null
`script_policies` member:

- Cached-findings semantics kept: no re-detection, entries preserved
  (the mixed row pins them with `reflect.DeepEqual`).
- No clobber: verdicts that already record policies are left
  byte-identical, so a subject without parsed commands (the source-audit
  path carries none) never erases a stored record. An explicit null
  counts as absent — null is the fresh-path encoding of "no script
  commands", so a later commands-carrying audit still backfills.
- `GateReadOnly` never writes: the backfill sits behind `persist`, and
  read-only still reports policies from the current manifest via the
  in-memory report computed fresh on every path.

A cache version bump was rejected: it would invalidate every verdict
and force re-detection for a record-format gap the backfill closes
with one conditional rewrite.

### Row table, revision 3

| Case | Test | Production entry |
|---|---|---|
| reviewer's row: pre-R4 verdict + enforced no-warning command gains the exact identity | `TestReviewerExistingVerdictGetsPolicyRecord` (committed review flow + exact-entry pin) | `audit.Gate`, twice |
| mixed old verdict: both entries backfilled, cached finding preserved, warning still emitted | `TestExistingVerdictGainsMixedPolicyRecord` | `audit.Gate`, twice |
| read-only hit: file byte-identical, in-memory report carries the exact identity | `TestGateReadOnlyCacheHitReportsPoliciesWithoutWriting` | `GateReadOnly` + `auditSubject(..., false)` |
| commands-less writable hit: stored record byte-identical | `TestCacheHitWithoutCommandsKeepsStoredPolicies` | `auditSubject(..., true)` |

### Mutant table, revision 3 (all killed, no survivors)

| Mutant | Attack | Result |
|---|---|---|
| R3-M1 skip the backfill (`return` first in `backfillScriptPolicies`) | revive the rev2 defect | BOTH backfill rows FAIL with `production Gate left existing verdict without per-command identity`; the two no-write guards pass (expected — they assert write absence, orthogonal to this mutant) |
| R3-M2 unconditional rewrite (drop the `verdictHasScriptPolicies` guard) | clobber stored records | `TestCacheHitWithoutCommandsKeepsStoredPolicies` FAILs (byte-identity broken); both backfill rows still pass — kill is specific to the conditional |

All mutants restored from saved bytes after each run; `cmp` confirms
the file is back, and `git status` shows only the intended delta.

### Verification (real exit codes, `zsh`, Darwin x86_64, Go 1.26.0)

Spec served from the EXACT CI pin: `git archive
87a0d006.../conformance/v1` extracted to a temp root.

- New rows: `go test -count=1 ./internal/audit/ -run
  'TestReviewerExistingVerdictGetsPolicyRecord|TestExistingVerdictGainsMixedPolicyRecord|TestGateReadOnlyCacheHitReportsPoliciesWithoutWriting|TestCacheHitWithoutCommandsKeepsStoredPolicies'`:
  exit 0, 0.571s, 4/4 pass.
- `CURATOR_CONFORMANCE_ROOT=<pin> go test -count=1
  ./internal/scriptpolicy/ ./internal/audit/ ./internal/skillcheck/`:
  exit 0 (0.687s, 1.726s, 2.186s). Vector consumer unskipped, 4/4 pass.
- `go test -count=1 ./cmd/curator/ -run 'TestCLIAudit'`: exit 0, 1.435s.
- `go test -count=1 ./internal/install/ -run
  'TestScriptAuditLabelsAtInstallEntry|TestDeclaredOnlyWithHostsInstallsWithOnlyDeclaredOnly|TestScriptAuditLabelsAtGlobalInstallEntry|TestDeclaredOnlyWithHostsGlobalInstallsWithOnlyDeclaredOnly'`:
  exit 0, 161.238s (Gate-behavior regression: project + global label rows).
- `go vet ./...`: exit 0. `gofmt -l cmd internal`: clean.
  `golangci-lint run ./internal/audit/...`: 0 issues.
- Full landing suite runs once via the handoff runtime, not locally.

### Ratio / bounds update

Unchanged from revision 2: 12/12 keys consumed, 4/4 audit cases at four
production entries each, global 4/4 + negative row. New coverage: the
cache-upgrade path (old-format verdict → backfilled record) at the
production gate. No platform-specific code; hosted evidence comes from
the landing CI run. Platform-ledger registration remains R5 scope.

revision 4 = revision 3 unchanged; gate rerun after a Windows managerlock timing flake

revision 5 = revision 4 unchanged; gate rerun after an ubuntu internal/install package timeout on a 2x slow runner
