# N8 revision 3 (republish on trunk 3cb461b3)

No code changes. Workspace converged onto trunk 3cb461b3 (N9 landed). Ran git diff HEAD on CHANGELOG.md and internal/audit/audit.go: no conflict markers. CHANGELOG.md keeps the N7 and N8 entries. audit.go adds the N8 hunks (ParseDigest before filesystem access in PinAtVersion, pinDir containment) on top of N9 created_at.

Compile-only checks (R223, no local go test): go vet ./... exit 0; go build ./... exit 0; gofmt -l cmd internal printed nothing.

Hosted green: the merged candidate (commit 2a6fbbbc on base 3cb461b3) was pushed to gate/TASK-261008-1j34ro-rev3-green. https://github.com/relux-works/curator/actions/runs/37904101780 concluded success (gh run watch --exit-status exit 0). All jobs passed: Test ubuntu/macos/windows, Race, Lint, drivers, gate self-tests. The candidate-suite and rose-air jobs were skipped by the workflow. The scratch branch has been deleted.

The production fix and the tests are byte-identical to revision 2, so the red and ordering-mutant evidence from revision 2 still applies.