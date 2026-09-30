# TASK-260917-8vfgxf review verdict — rev2 (tree 6dd496cf) — CHANGES REQUESTED

The worktree matches the candidate tree: a temp-index write-tree gives 6dd496cf.

## What is correct
- envprofile lane: contextaudit.Detect (envprofile.go:1534 auditMember, :1573 auditNewUpdateMember, :2001 migrateGlobalSkills) walks every regular file with opaquescan.NULPaths through stateread and does not follow symlinks. The finding is blocking and cannot be waived, and the error names the class and the file. The colliding trees are built from v1 framing with an equal-hash assertion, and both are refused through Install. The NUL-free tree installs, and deep assets/ and docs/ files are refused (nul_opaque_test.go).
- The stateread guard: the allowlist entry moved from detect to detectWithOpaquePaths with the same reason and nothing added. No Windows-reserved names. No CHANGELOG or LOGBOOK changes.

## F1 (blocking): the legacy skill admission lane skips the rule under default config
The rule sits inside audit.auditSubject (audit.go:247). But audit.gate returns nil,nil when `!cfg.Audit.Enabled` (audit.go:164), and Enabled defaults to false (config_test.go:113). So a skill snapshot with a NUL file is admitted by:
- internal/install/install.go:580-582 (curator install, skills)
- internal/install/global.go:198-200
- cmd/curator/main.go:1896-1898 and :2198
- envprofile.go:1712 (strictAuditMember goes through audit.Gate too)

Probe: I added a temporary test in internal/audit and removed it afterwards. The snapshot had docs/deep/x.bin = "a\x00b". Gate with Audit.Enabled=false returned `warnings=[] errs=[]`: admitted. With Enabled=true it returned `audit blocked: s: critical audit.opaque.nul-byte ...`. The brief says the rule applies "regardless", at the production entry. It must not depend on the opt-in audit toggle.
Fix: run opaquescan.NULPaths before the `!cfg.Audit.Enabled` early return, or at each install/resolve admission site. Add a row that drives install.Install (and the global lane) with the default config (audit disabled) on both colliding trees plus a deep docs/ or assets/ file.

## F2: the audit.Gate block message does not name the file
audit.go:197 formats `severity ID - Evidence`. finding.File is dropped, and the Evidence text is generic. The rule requires the finding to name the file. Include finding.File in the error, and assert it at the install-lane row.

## Mutants
I did not re-run them on this revision, because F1 already fails item 1 of the review note. Once F1 is fixed, the mutant rows must include a default-config install-lane row. Without it, the "demoted to a warning" mutant cannot be killed on that lane.

## rc.13
Not re-verified; carry forward the producer's statement.
