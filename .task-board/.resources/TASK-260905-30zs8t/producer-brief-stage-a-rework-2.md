# Producer brief: stage (a) core — rework 2 (F8, F9)

Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-a-core`, branch
`feat/agent-environments-stage-a`, head `ac9d0037` (already rebased onto curator main `a2406dfe`).
Findings: `TASK-260905-30zs8t_review-findings-stage-a-2.md` — F1–F6 verified fixed and held; two new
**blocking** findings. Read both in full; each carries the reviewer's reproduction. New signed commits
on top of `ac9d0037`, no rewrite.

## Author decisions
- **F8 (blocking)** — canonicalize the git operand through `internal/identity.Parse` at every
  boundary: `Install`, the requirement reader, and `migrateGlobalSkills`. `canonicalGit` as it stands
  (trim space, trim one trailing slash) is not the core §6.1 identity, so the lock member `source` and
  the marker `profile.source` fail the published schemas and two spellings of the same repository
  resolve to two identities. Keep the raw URL only where the spec asks for it (the `path` kind's
  `source_path` is a different field). A malformed network source is rejected, not passed through.
  Compare canonical identities in the one-name-one-commit agreement check. Regression cover: install
  one repository under `https`, `scp`-style and `ssh://` spellings (git `insteadOf` keeps it offline,
  as the reviewer did) and assert one identical canonical identity in the lock and the marker, and
  that both validate against `context-lock-v1` and `agent-environment-marker-v1`.
- **F9 (blocking)** — run the gates §9.1 says every closure member passes, not only the two classes
  the context detector adds: read the machine configuration in the profile path; gate every `git`
  member's canonical identity against the core §6.1 source allowlist before the clone (reuse the shape
  of `closure.gateSource`); populate the MCP package allowlist from the same config
  (`mcp_package_not_allowed`); and run the manager §7 source audit in strict mode over every member new
  to the lock — raw-tree hashing, the static canary whose failure always blocks, the deterministic
  detectors and **revocation**. §9.1's "an advisory profile install does not exist" is stronger than
  `audit.Gate`, which no-ops when `cfg.Audit.Enabled` is false: the profile path applies revocation and
  the canary regardless of that flag. Tests: a member whose identity is outside the allowlist is
  refused; a revoked source is refused; the canary blocks; an `mcp` member outside the package
  allowlist is refused with its diagnostic. If any single part genuinely cannot land in stage (a), it
  is a **stated** bound in the package doc *and* the report — not silence.
- The reviewer's non-blocking note on F2 (an operator should be warned when a global skill did not
  migrate) is for the stage that wires the operator surface; record it in the report, do not implement.

## Gates and delivery
`go build ./...`, `go vet ./...`, `gofmt -l`, `golangci-lint run ./...` if installed,
`go test -count=1 -race` on the touched packages, the vector families with
`CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator-spec/conformance/v1`,
`bash .github/ci/gate-selftest.sh`, the platform-case gate for the three GOOS values, and
`go test -count=1 -timeout 30m ./cmd/curator` once at the end. For each fix, a mutant that restores the
old behaviour must fail a named test — put the table in the report. Signed commits; do not push, tag,
or open a PR. Attach `TASK-260905-30zs8t_rework-report-2.md`;
`task-board handoff TASK-260905-30zs8t --role developer`. Never write LOGBOOK.md or anything into the
control root.
