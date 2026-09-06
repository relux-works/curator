# BUG-260906-1bdotx: registry-snapshot-future-timestamp-flake-on-windows

## Description


## Scope
internal/install/TestRegistryAttestationLandsInMarker failed on Test (windows-latest) in curator run 34022458940 (a workflow_dispatch of the candidate lane on feat/agent-environments-stage-b at 1a936e77) with: registry test-reg snapshot timestamp is too far in the future, then every trusted audit registry served a tampered snapshot. The same test on the same head passed on the pull_request run 34021551479 for PR #60, and every other job in the dispatch run was green including all three Candidate suite jobs. Stage (b) does not touch internal/install: git diff origin/main..HEAD -- internal/install is empty. The fixture writes created_at with time.Now().UTC().Format(time.RFC3339) at internal/install/registry_e2e_test.go:44, and the gate at internal/registry/snapshot.go:158 rejects a snapshot when parsed.CreatedAt.After(now.Add(clockSkew)). RFC3339 formatting truncates fractional seconds downward, so a fixture written before the check should never be ahead of it; the failure therefore points at the now the checker is given, at runner clock adjustment during the 39m44s Windows job, or at a clockSkew tolerance that is zero or too tight for a contended runner.

## Acceptance Criteria
The mechanism is identified from evidence rather than guessed, and either the fixture pins a deterministic clock the checker also sees, or the tolerance is stated and justified. Proven by a narrowing mutant: shrinking the tolerance or reintroducing the skew must fail a named test. The test is not weakened into ignoring a genuinely future timestamp, which is a real tampering signal the gate exists to catch.
