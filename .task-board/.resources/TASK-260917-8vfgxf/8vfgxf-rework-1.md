# TASK-260917-8vfgxf — rework 1 (THE ONLY CURRENT INSTRUCTION, with 8vfgxf-brief.md)

Review of rev2 (`TASK-260917-8vfgxf_review-verdict-rev2.md`) = CHANGES REQUESTED. The envprofile lane is correct. Fix these:
F1 (blocking). The rule must not depend on the opt-in audit toggle. audit.gate returns early when `!cfg.Audit.Enabled` (audit.go:164),
and Enabled defaults to false. So under the default config a snapshot with a NUL file is admitted by:
- internal/install/install.go:580-582
- global.go:198-200
- cmd/curator/main.go:1896-1898 and :2198
- envprofile.go:1712 (strictAuditMember)

Run opaquescan.NULPaths unconditionally, before the early return, so every admission site enforces it. Leave the rest of the audit
opt-in as it is.

Rows, with the DEFAULT config (audit disabled): drive install.Install, the global install lane and the context/env lane on both
colliding trees and on a deep docs/ or assets/ NUL file. Each must be refused.

F2. audit.go:197 drops finding.File. Include the file path in the blocking error, and assert it in the install-lane row.

Mutants (real exit codes, killed on the install lane with the default config):
- the rule gated back behind Audit.Enabled;
- the rule limited to one directory;
- the finding demoted to a warning.

Run go test for ./internal/audit, ./internal/install -run 'NUL|Opaque', ./internal/envprofile -run 'NUL|Opaque' and
./cmd/curator -run 'NUL|Opaque'. Set status development, update the results, run `task-board handoff TASK-260917-8vfgxf --role
developer`, then END YOUR TURN. No CHANGELOG/LOGBOOK edit. No Windows-reserved file names. Never spell any employer name.
