# Review brief: stage (a) core — cycle 3 (rework of F8, F9)

Subject: curator worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-a-core`, branch
`feat/agent-environments-stage-a`, head `314ae748` (one signed commit on the cycle-2 head `ac9d0037`,
8 files, +647/−39). Cycle-2 findings: `TASK-260905-30zs8t_review-findings-stage-a-2.md`; author
decisions: `producer-brief-stage-a-rework-2.md`; rework report: `TASK-260905-30zs8t_rework-report-2.md`
(it carries a mutant table and a "stated bounds" section — judge both).

Verify by reproducing, not by reading:
1. **F8** — install one repository under `https`, `scp`-style and `ssh://` spellings (git `insteadOf`
   keeps it offline) and confirm one identical canonical identity in the lock member `source` and the
   marker `profile.source`, and that both instances validate against `context-lock-v1` and
   `agent-environment-marker-v1` with a real schema validator; confirm a malformed network source is
   rejected at the boundary rather than passed to the clone; confirm the one-name-one-commit agreement
   check compares canonical identities; grep for any remaining raw-operand path (`canonicalGit`, the
   requirement reader, `migrateGlobalSkills`) that could reintroduce the divergence.
2. **F9** — a `git` member whose canonical identity is outside the machine allowlist is refused
   **before** the clone; a revoked source is refused; the static canary blocks; an `mcp` member outside
   the package allowlist is refused with `mcp_package_not_allowed`. Critically, check the strength the
   brief demanded: §9.1's "an advisory profile install does not exist" — with `cfg.Audit.Enabled` false
   the profile path must still apply revocation and the canary. Set that flag false in a scratch config
   and prove it.
3. **Stated bounds** — the report lists bounds for stage (a). For each, decide whether the spec permits
   deferring it to a later stage or whether it is a gate §9.1 requires now; a bound that hides an
   unimplemented mandatory gate is a finding, not a bound.
4. **Regression** — re-run your cycle-1 and cycle-2 "verified, held" set: the directory-addressed
   detector cases, the fresh-home default profile (list/use/sync/use --clear), the mutation lock and
   journal with an interrupted switch, the range differential against node-semver, the resolution
   replay, the header bytes, the marker shape, the CLI rows. This rework touched the install path, so
   re-do the F1/F2 reproductions rather than trusting them.
5. **Gates** — `go build ./...`, `go vet ./...`, `gofmt -l`, `golangci-lint run ./...` if installed,
   `go test -count=1 -race` on the touched packages, the vector families through
   `CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator-spec/conformance/v1`,
   `bash .github/ci/gate-selftest.sh`, the platform-case gate for the three GOOS values; cite the
   producer's `./cmd/curator` run. Confirm the producer's mutant table by running two of its mutants
   yourself.
6. Commits signed by the repository's human identity; nine commits on curator main `a2406dfe`.

Read-only (scratch under the worktree's `.temp/`). Never write into the control root. Findings resource
`TASK-260905-30zs8t_review-findings-stage-a-3.md`. Blocking/major → `development`; else explicit ACCEPT
at `to-review` with `accept_cr` on the recorded revision. Do not mark done.
`task-board handoff TASK-260905-30zs8t --role reviewer`.
