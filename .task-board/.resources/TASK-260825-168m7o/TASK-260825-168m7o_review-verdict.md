# TASK-260825-168m7o review verdict: ACCEPTED

Reviewer run: RUN-260824-effc22 (reviewer archetype, read-only).

## What was reviewed

- internal/config/buildhttps.go (new): build_https parsing, validation, matching, serialization.
- internal/config/buildssh.go (modified): longest segment-aware prefix match extracted into generic longestScope[T] shared by MatchBuildSSH and MatchBuildHTTPS; diff inspected, behavior-preserving.
- internal/config/config.go (modified): Config.BuildHTTPS field, parseBuildHTTPS wired into Parse, build_https added to managerKeys only.
- internal/config/buildhttps_test.go (new): full AC coverage.

## AC verification

- Parse/serialize roundtrip: TestBuildHTTPSSerializationRoundtrip writes BuildHTTPSObject output back through Load and requires DeepEqual. PASS.
- Literal-secret rejection: parse-level (token=ghp_abc123secret) and validator-level cases both assert the error mentions secrets never live in the config. PASS.
- Exactly-one-source rejection: XOR check in ValidateBuildHTTPS; tests cover none, empty entry, and both-set. PASS.
- Grammar rejections: uppercase host, empty/dot/dotdot segments, trailing slash, scheme, port, empty scope, unknown fields, wrong types -- all fail closed with labeled verr errors. Scope validity is ValidBuildSSHScope directly, and TestBuildHTTPSReusesBuildSSHScopeGrammar pins both surfaces to the same verdicts. PASS.
- Longest-prefix match incl. segment boundary: portals-evil and portals.git-mirror fall back to the org scope, relux-works-evil to the host scope, git.example.community matches nothing; host-only identity correctly selects nothing (not a canonical identity). Confirmed against identity.MatchesPrefix/ValidCanonical semantics. PASS.
- Manager key set extended; TestBuildHTTPSIsAManagerKey exercises the system-config default path. build_https absent from LockableKeys (config.go lines 46-51), pinned by TestBuildHTTPSIsNotLockable, consistent with build_ssh and the ratified rule. PASS.

## Gates rerun independently by the reviewer

- go build ./... -- PASS
- go test ./internal/config/... -count=1 -- ok, 0.527s
- go vet ./internal/config/... -- PASS
- golangci-lint run ./internal/config/... -- 0 issues

## Architecture fit

Mirrors buildssh.go in structure, error labeling (verr), sorted-scope deterministic fault reporting, and object-serialization shape. Reuse is real, not copy-paste: one grammar function, one generic matcher. Token source enum (git-credentials, keyring) matches the two-source language of the sibling tasks that will consume it. No write helpers added -- correctly deferred to TASK-260825-2gyhq8 and documented.

## Non-blocking observations

- In parseBuildHTTPS the per-entry field loop iterates a Go map, so when two fields of one entry are simultaneously malformed the reported field can vary between runs. Scope-level ordering is deterministic (sorted), matching build_ssh, and parseBuildSSH has the same per-entry property -- consistent with the existing surface, not worth a rework cycle.

## Verdict

accepted -> done. Reviewer supplies no commit_ack; the commit-owning mover commits the scope (internal/config/buildhttps.go, buildhttps_test.go, buildssh.go, config.go) and makes any enforced parent done transition.