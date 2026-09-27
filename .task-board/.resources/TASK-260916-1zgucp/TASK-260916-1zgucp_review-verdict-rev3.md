# TASK-260916-1zgucp review verdict — rev3: ACCEPTED

Reviewer: claude-opus-5-5 (low). Candidate tree 78bcb664 == worktree (temp-index write-tree); base bd3c0f43; 31 paths, only product/test/.github/ci (no CHANGELOG/LOGBOOK, no stray files; nothing from the earlier control-root write is missing — every item on the changed-path list is present in the tree).

## Rules checked through production entry
- Signer allowlist: contextresolve.verifyCandidate (unsigned → context_source_unsigned, wrong/unknown → context_source_signer_rejected, require without allowlist → context_source_signers_missing; empty allowlist admits none). Driven through `curator profile install` using real git+ssh signatures (cmd/curator/profile_signers_test.go:116,194), env status posture asserted (JSON + human).
- System delta: envprofile update/reinstall refuse with profile_update_confirmation_required unless --confirm-system-delta (envprofile.go:949,1212; flags profile.go:111,278); goldens profile-update-{system,mcp}-delta.golden.
- Gatefix items 1–3: TestManagerOwnedAbsenceReadsAreGuarded, TestUpdatePathWithoutStatePinIsSourceInvalid, and the signer-posture tests are green locally.

## Independent runs (zsh, set -o pipefail)
- go test ./internal/contextresolve ./internal/contextlock ./internal/config → ok rc=0
- go test ./cmd/curator -run 'Signer|DeltaConfirmation|Surfacing' → ok rc=0
- go test ./internal/envprofile -run 'Signer|Delta|ManagerOwnedAbsence|UpdatePathWithoutStatePin|Surfacing|Status' → ok (169s) rc=0
- With CURATOR_CONFORMANCE_ROOT=curator-spec@23435129 conformance/v1: TestSourceSignerVectorsAtProfileInstall (4 vectors), TestRevisionDoesNotBorrowTagSignature, TestSourceSignerVectorsAtResolve, TestSourceSignerMergeVectorsAtLoad → PASS rc=0

## Mutants (disposable git-archive copies)
- M1 verifyCandidate → `return nil`: contextresolve FAIL rc=1; CLI unsigned/wrong-signer/empty-allowlist vectors FAIL rc=1 (accepted vector passes, as expected). KILLED.
- M2 both `len(triggered)>0 && !confirm` gates → false: cmd/curator -run DeltaConfirmation FAIL rc=1. KILLED.

## Gap ledger
ioemse-owned rows: 39 → 0. Total rows 73 → 73: the 39 rows were re-attributed to STORY-260922-1cenbr, with the reason stated as still blocked by the separately owned environments.permissions field. That is consistent with the ratchet, which rejects any row that passes.

## Bounds / notes
- Without CURATOR_CONFORMANCE_ROOT, the CLI vector subtests SKIP locally, so the evidence depends on the hosted gate exporting it (ci.yml:207,328 does).
- Windows/Linux signature verification was not rerun here; the hosted gate is the arbiter.
