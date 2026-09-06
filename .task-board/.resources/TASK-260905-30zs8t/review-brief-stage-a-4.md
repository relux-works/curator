# Review brief: stage (a) core — cycle 4 (rework of F10, F11, F12)

Subject: curator worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-a-core`, branch
`feat/agent-environments-stage-a`, head `dea3f5ac` (one signed commit on the cycle-3 head `314ae748`).
Cycle-3 findings: `TASK-260905-30zs8t_review-findings-stage-a-3.md`; author decisions:
`producer-brief-stage-a-rework-3.md`; rework report: `TASK-260905-30zs8t_rework-report-3.md`.

Verify by reproducing:
1. **F10** — on an isolated home with `CURATOR_SYSTEM_CONFIG` locking `allowed_sources` and
   `audit.revocations` so a declared global skill is excluded or revoked, the default-profile migration
   is refused; the mirror case succeeds when the system config permits it; a config read or parse
   failure fails the migration rather than emptying the policy (drive the unreadable case). Grep for any
   remaining path in `envprofile` that loads configuration on its own.
2. **F11** — the three narrowing tests exist and bite: allowlist `example.com/org` vs operand
   `example.com/org-evil/pkg`; a revoked `skill` member and a revoked `mcp` member; the canary forced to
   fail on the profile path. Run each as a *weakening* mutant (not a deletion) yourself and confirm a
   named test fails. Confirm `TestMigratedSkillSourceIsCanonical` now asserts the canonical identity.
3. **F12** — a `file://` operand and a `file://` requirement source are refused with
   `profile_source_invalid` at the same boundary as a malformed network source; no operator-reachable
   path writes a schema-invalid lock or marker (drive the CLI shapes the cycle-3 reviewer used); the
   converted tests use `insteadOf` fixtures.
4. **Regression** — re-run your cycle-1/2/3 "verified, held" set, especially the F1 directory-addressed
   detector cases, the F2 fresh-home default profile, the F3 interrupted switch, F8's three spellings,
   and F9's gates: this rework touched the migration and install boundaries.
5. **Gates** — the full set as in the previous briefs, plus `bash .github/ci/gate-selftest.sh` and the
   platform-case gate for the three GOOS values; confirm the producer's mutant table by running two
   rows yourself.
6. Commits signed by the repository's human identity; ten commits on curator main `a2406dfe`.

If this cycle is clean, this is the ACCEPT that lets stage (a) land. Read-only (scratch under the
worktree's `.temp/`). Never write into the control root. Findings resource
`TASK-260905-30zs8t_review-findings-stage-a-4.md`. Blocking/major → `development`; else explicit ACCEPT
at `to-review` with `accept_cr` on the recorded revision. Do not mark done.
`task-board handoff TASK-260905-30zs8t --role reviewer`.
