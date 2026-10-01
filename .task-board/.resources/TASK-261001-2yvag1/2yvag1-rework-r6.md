# THE ONLY CURRENT INSTRUCTION — TASK-261001-2yvag1 rework after rev6 review (binding)

Astra's review verdict for rev6 is CHANGES REQUESTED (`TASK-261001-2yvag1_review-verdict-rev6.md`). Read it fully. Fix exactly these three findings and nothing else.

## R1 (P1, security): symlinked auth parent accepted
`internal/envprofile/muse.go:58`, via `checkPassthrough`; `managed.go:2514` returns current before repair preflight.

Scenario: `<managed-home>/config` replaced by a symlink to an outside directory that holds a valid-looking `muse/auth.json` link. Today both `Resolve` (Repair=false) and `--repair` succeed and emit v3.

Fix:
- Establish the managed parent boundary (nofollow, the existing boundary abstraction) BEFORE auth liveness is accepted.
- Keep the intentional final credential file-link.
- On refusal: preserve outside and native bytes, emit no fragment.

Add production-entry rows for bare resolve and for `--repair` with such an outside parent. Each must refuse with a non-zero exit.

## R2 (P2): fragment corpus not bound to production emission
`cmd/curator/muse_test.go:267/275/293`. Today 0/36 rows invoke Resolve, and a mutant that drops `permissions` from `Fragment.Object` survives.

Fix:
- Validate the ACTUAL emitted document from production `Resolve` (and the CLI) against the v3 schema, as permanent assertions.
- Explicitly classify reader-only or static rows as such.

Show that the missing-permissions mutant now fails, with a real exit code.

## R3 (P2): remove unrelated broker EPIPE changes
Remove all changes to `internal/testcli/cli.go` and `internal/testcli/cli_test.go`; they must be byte-identical to base. That flake belongs to BUG-261001-2772iz. If a test then fails ONLY because of that flake, say so; do not fix it here.

## Verify
Run with real exit codes:
- `go test ./internal/envprofile ./cmd/curator -run 'TestMuse|TestEnvStatus|TestManagerOwnedAbsenceReadsAreGuarded|TestEmptyAllowlistWarning' -count=1`
- the new R1 rows
- the R2 mutant (fails)
- the R1 mutant: drop the boundary check, the rows fail

Add a "Revision 7" section to the results. Then run `task-board handoff TASK-261001-2yvag1 --role developer` and END YOUR TURN.

Never spell any employer name. No LOGBOOK; CHANGELOG at most one line.
