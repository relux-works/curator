# TASK-260918-ryh3kw — review verdict rev6 (delta review): ACCEPTED

Candidate: base 6bd98d49, tree 3229dac8 (worktree `git write-tree` == 3229dac8), 17 paths.
Method: `git merge-tree --write-tree --merge-base eca2bf27 6bd98d49 refs/campaign/1ll22r-rev5-20260927` (= 1b526693, conflicted) diffed to 3229dac8 on the two conflict files under review.

1. internal/envprofile/managed.go gatherSeeds (≈l.794-822): resolution keeps trunk E3 body verbatim (switch on Kind: absent→continue, present→proceed, other→environment_seed_unreadable; stripCodexSeedMCPServers + not-inherited names; bundle.files[seed]=payload) and only replaces `stateread.ReadFile(seedPath)` with rev5's injectable `req.readRegularFile` (default stateread.ReadRegularFile) and renames state→file. Rev5's simplified side (which would have dropped the E3 strip) was correctly discarded. Behaviour unchanged except rev5's regular-file strictness. No other lines in neither side.
2. internal/envregistry/envregistry.go const block: trunk block + rev5's two constants (DiagPassthroughUnreadable, DiagStoreUntrusted) at gofmt alignment of the wider trunk block; DiagSeedShadowed moved below the MCP diagnostics (pure reordering of a constant, same value). No semantic change.

Runs (zsh, pipefail, real exit codes):
- gofmt -l internal/ → empty; go vet envprofile/envregistry/stateread → 0
- go test ./internal/envprofile -run 'Seed|Codex|Mcp|MCP|ReadFailure|Guarded' -count=1 → ok, exit 0
- go test ./internal/stateread ./internal/envregistry -count=1 → ok, exit 0
- Mutant: gatherSeeds seed read reverted to os.ReadFile + os.IsNotExist → TestManagerOwnedAbsenceReadsAreGuarded FAIL ("managed.go:*ResolveRequest.gatherSeeds tests not-exist after os.ReadFile without a seam route"); file restored byte-identical (git diff --quiet 3229dac8 → 0).
Rev5 acceptance (review 2) stands for everything else; hosted gate green per orchestrator.
