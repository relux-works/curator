# Revision 10 verdict: CHANGES_REQUESTED

Candidate e90fe25a8e7166b0c0075aa281792182b795e824; base 0945447816cb4ff105eafd79228be1d7a639d805. Independently compared all 16 changed files byte-for-byte to the candidate (16/16 matched). No product/test code modified.

## P1 — failed remote discovery is silently accepted as no remote

internal/gitops/gitops.go:126-142: HasRemote returns false for any `git remote` error; Fetch interprets false as a successful no-op. Production `project refresh` -> resolveDraftPlan -> ResolveDraft -> ensureRepo -> Fetch therefore succeeds without fetching when remote discovery fails. Absence and failure to read must be distinct. This newly added origin-less optimization weakens the required acquisition-failure semantics.

Independent production CLI reproduction (attached Python fixture and log, built from this candidate): configure a legacy branch root backed by a real local upstream, resolve and real-install, advance upstream with runtime-only bytes, inject a Git wrapper which exits 73 only for remote enumeration and delegates all other commands. `project refresh` exits 0 and retains f3313f5d1a10f0fe01da0f2dfef92ad7e0d6240f instead of upstream 3dfa83dbb1e13fc19bc21077f9229590e7caf20c. Restoring Git and fetching permits refresh to lock the new commit. Script exits 0 with both bad-path and control assertions satisfied. No network, real credential or runtime-home writes.

Required rework: preserve and propagate remote enumeration errors; skip fetch only after successful empty enumeration. Add a production CLI negative test injecting this read failure and asserting nonzero refresh plus unchanged lock, bindings and installed state. Retain the genuine origin-less positive control. A mutant collapsing error into empty must fail this negative test. Preserve the Windows fixture fix and prior accepted fixes.

## Verification and bounds

- `go build -o /tmp/curator ./cmd/curator`: exit 0.
- `python3 /tmp/TASK-260910-19w2aj-rev10-probe.py`: exit 0; proves the defect via CLI (one injected failure shape / one tested), not a passing product assertion.
- `git diff --check`: exit 0.
- Independently queried GitHub run 35228921467: success. Commit 5e9de72d7a28581e7de70e0781b7261db7eb73c6 has exact candidate tree e90fe25a8e7166b0c0075aa281792182b795e824. Windows/Ubuntu/macOS tests, Linux/macOS race, lint and gate checks green. Rose-air and candidate-suite jobs skipped; not claimed passing.
- Focused independent command: `go test -p 1 -count=1 -timeout 90s ./cmd/curator -run '^TestProjectRefresh(FetchFailurePreservesStateThroughCLI|LegacyConfiguredGitBranchThroughCLI|LegacyNetworkGitBranchThroughCLI|TransitiveTagAdvanceThroughCLI)$' -v`. Host stalled without test output, including beyond its internal timeout; terminated own processes at about 110 seconds, exit 143. No local Go-test pass claimed. Hosted results and attached producer evidence are accepted only for their recorded checks, not as proof against the reproduced error branch.
- Revision 10's Windows fixture now uses the portable empty-command skill while retaining lock/bindings/installed-skill rollback assertions; hosted Windows confirms that correction.
- Earlier mutant evidence is producer evidence, not independently rerun here. No exhaustive conformance or unavailable-platform claim.

Run goal queried: not goal-bound. Route to to-dev; no accept_cr or commit acknowledgement.
