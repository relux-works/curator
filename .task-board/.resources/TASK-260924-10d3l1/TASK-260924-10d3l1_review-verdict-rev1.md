# TASK-260924-10d3l1 review verdict — CR rev1: ACCEPTED

Candidate tree 9b607c44 verified (worktree matches; disposable git-archive copy used). Delta: CHANGELOG.md + internal/install/draftevidence_test.go only; no production change (ResolveExact untouched).

## Driven rows (production entry install.Project, strict policy)
- exact-admits (unchanged), wrong-name, wrong-context, NEW wrong-repository-only (source_identity only), NEW wrong-commit-only (commit only).
- Refusal rows now also seed a prior ok install and assert byte-identical lock, draft bindings, marker, SKILL.md, references; zero Attestations; no stub URL / pinned key / ed25519: / 127.0.0.1 in Errors+Messages.

## Independent reruns (zsh, pipefail)
- `go test ./internal/install -run TestDraftEvidenceExactMatch -count=1 -v` → all 5 subtests PASS (30s). `go vet ./internal/install` ok.
- Mutant M1 MatchesExact `record.SourceIdentity == sourceIdentity` → `record.SourceIdentity != ""`: KILLED (wrong-repository-only-refuses fails, only it).
- Mutant M2 `record.Commit == commit` → `record.Commit != ""`: KILLED (wrong-commit-only-refuses fails, only it).
- Validation log: exit 0, exact_command_shard green=1.

## Residual (non-blocking)
- Leak list replaced generic "http://"/"https://" with the concrete stub URL; loopback 127.0.0.1 still covered and Messages newly covered, so net stronger for this stub, but a leak of an unrelated non-loopback URL would no longer be caught.
- Logbook DoD satisfied by this resource (no LOGBOOK.md edits).
