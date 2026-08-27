## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(3))

## Blocked By
- (none)

## Blocks
- TASK-260822-2505vo
- TASK-260822-3pkc80

## Checklist
- [x] build_ssh parsed into config with fail-closed grammar (host lowercase, segment charset, no empty segments)
- [x] Longest segment-prefix match helper; boundary test portals vs portals-evil
- [x] Serialization roundtrip + managerKeys extended; unknown entry fields rejected
- [x] go test ./internal/config green; no other implementation named anywhere
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260822-13cbfb, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260822-13cbfb)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260822-fad725, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260822-fad725)
build_ssh landed in internal/config/buildssh.go + buildssh_test.go, wired through Parse/managerKeys in config.go.

Grammar (fail closed): scope = lowercase DNS host (<=253) + optional /-separated segments [A-Za-z0-9._-]+, path case preserved per spec 6.3. Rejected: empty scope/segment, leading or trailing slash, . and .. segments, uppercase host, git..example.com, trailing dot, leading -, explicit port, scheme, scp form, >4096 runes. Entry = {agent?, identity?, known_hosts?} only; unknown keys reported as unsupported field(s). Each value must be absolute or ~/ (Windows C:\, C:/, \\host\share accepted platform-independently), no control chars; a present-but-empty string is a fault, not an absent field. At least one of agent/identity required, so known_hosts-only is rejected. Scope faults reported in lexicographic order so a multi-fault config always fails the same way.

Match: BuildSSHFor reuses identity.MatchesPrefix (same matcher as the source allowlist) behind an identity.ValidCanonical guard, longest scope wins. portals vs portals-evil boundary covered, plus portals.evil, relux-works-evil, git.example.community, local/empty identity, bare host, uppercase host, ssh:// URL, .. path - all select nothing.

API for downstream tasks: BuildSSHFor, BuildSSHScopes (sorted), BuildSSHObject (serialization), ValidBuildSSHScope, ValidateBuildSSH, Expanded (~ expansion for 2505vo), Empty. build_ssh added to managerKeys, deliberately NOT to LockableKeys - locking operator credentials from an org config is a policy call nobody asked for.

Verification, real exit codes: go build ./... 0; go test ./internal/config -count=1 0 (13 build_ssh tests PASS); same with -race 0; go test ./... 0 (31 packages ok, 0 FAIL); go vet ./... 0; golangci-lint run 0 issues; gofmt -l cmd internal empty. 100% statement coverage on every buildssh.go function.

ANOMALY: two implementer runs (RUN-260822-13cbfb and RUN-260822-fad725) were dispatched for this one task and both wrote buildssh.go, leaving config.go with a duplicated parseBuildSSH block that did not compile. Reconciled; logged in LOGBOOK.md 2026-08-22 2024. Nothing is committed - working tree only, per the no-auto-commit rule.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-13cbfb, pid=13673, exit=0)
RECONCILIATION (run RUN-260822-fad725, the second of the two duplicate runs). The peer run RUN-260822-13cbfb and this run both completed. The working tree now holds the merged superset, re-verified end to end after landing. One grammar detail in the earlier note is superseded: agent now accepts true as well as a socket path. Rationale: the sibling CLI task specifies --agent [SOCKET] with an optional socket, and on macOS an agent socket path is per-login-session (/private/tmp/com.apple.launchd.*/Listeners), so persisting one into a machine-global config goes stale after a relogin. agent: true records the operator choice of an agent and lets resolution read the live socket and pass it explicitly as IdentityAgent=<socket>; agent: "<socket>" still parses, so the shape is a strict superset. agent: false, null and other types stay rejected, keeping exactly one spelling for identity-only. Struct is BuildSSHCredential{Scope, Agent bool, AgentSocket, Identity, KnownHosts}; the three authentication shapes are identity-only, agent alone, agent+identity (pinned). Folded in from the peer suite: identity null and known_hosts empty rejection cases, and TestValidateBuildSSHAcceptsEveryParsedCredential. Both peer files kept verbatim at .temp/TASK-260822-96m5pj/buildssh.peer-run-13cbfb.go.txt and buildssh_test.peer-run-13cbfb.go.txt. Post-landing verification, real exit codes: gofmt -l internal/config clean; go build ./... 0; go vet ./... 0; golangci-lint run 0 issues; go test -count=1 ./internal/config 0; go test -race -count=1 ./internal/config 0; go test ./... 0 (31 packages ok). 100% statement coverage on every buildssh.go function. Nothing committed - working tree only.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-fad725, pid=14345, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [reviewer] reviewer (claude) (run=RUN-260822-b910b6, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260822-b910b6)
REVIEW VERDICT (RUN-260822-b910b6): ACCEPTED. Every AC verified against the tree, not the notes. Roundtrip, empty-segment, uppercase-host, portals vs portals-evil boundary, longest-prefix, unknown-field rejection all covered by named tests that pass.

Independent verification, real exit codes from this reviewer: gofmt -l internal cmd clean; go build ./... 0; go vet ./... 0; golangci-lint run 0 (0 issues); go test -count=1 ./internal/config 0 (13 build_ssh tests PASS); go test -race -count=1 ./internal/config 0; go test ./... -count=1 0 (31 packages ok, 0 FAIL); go tool cover -func: 100.0% on all 9 buildssh.go functions. The attached patch artifact was re-diffed against a regenerated git diff internal/config and is identical modulo blob hashes, so the evidence is faithful.

Architecture fit: matching reuses identity.MatchesPrefix (internal/identity/identity.go:125), the same segment-aware matcher as the 6.1 source allowlist, behind an identity.ValidCanonical guard - no second prefix rule. write.go preserves the raw object so build_ssh survives AddProject/Bootstrap. applySystem treats build_ssh as a whole-key manager default. CLI/resolution/precheck/docs are separate siblings, so their absence here is correct scope, not a gap. No other manager implementation named in either new file.

Three non-blocking findings, none worth a rework cycle: (1) LOW - buildssh.go:251 ranges a map literal for identity/known_hosts, so when both paths in one entry are invalid the reported field flips run to run (measured 172/28 over 200 loads in a scratch copy); contradicts the determinism comment at buildssh.go:208, though the scope-level ordering that test pins does hold, and the config is rejected either way. Fix is a two-element ordered slice - fold into 3pkc80 or 4p3dcq. (2) INFO - BuildSSHObject normalizes an invalid struct (Agent false + AgentSocket set) into agent: <socket>; TASK-260822-3pkc80 must call ValidateBuildSSH before BuildSSHObject. (3) INFO - build_ssh deliberately not in LockableKeys; a defensible policy call worth confirming with the spec owner before the story closes.

Nothing committed - working tree only. internal/config/{buildssh.go,buildssh_test.go,config.go} are still uncommitted and belong to the commit-owning mover. Full report: TASK-260822-96m5pj_review.md.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-b910b6, pid=74874, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260822-96m5pj_spawn-log_-implementer--developer--claude-_RUN-260822-13cbfb.log](file://TASK-260822-96m5pj/TASK-260822-96m5pj_spawn-log_-implementer--developer--claude-_RUN-260822-13cbfb.log) — System spawn log captured by task-board
- [TASK-260822-96m5pj_spawn-log_-implementer--developer--claude-_RUN-260822-fad725.log](file://TASK-260822-96m5pj/TASK-260822-96m5pj_spawn-log_-implementer--developer--claude-_RUN-260822-fad725.log) — System spawn log captured by task-board
- [TASK-260822-96m5pj_results.md](file://TASK-260822-96m5pj/TASK-260822-96m5pj_results.md) — build_ssh config field: grammar, match helper, serialization, verification evidence
- [TASK-260822-96m5pj_full-test-01.log](file://TASK-260822-96m5pj/TASK-260822-96m5pj_full-test-01.log) — go test ./... full suite log, exit 0, 31 packages ok
- [TASK-260822-96m5pj_implementation-notes.md](file://TASK-260822-96m5pj/TASK-260822-96m5pj_implementation-notes.md) — build_ssh config field: surface, grammar, design decisions, concurrency incident, verification table
- [TASK-260822-96m5pj_build-ssh-config.patch](file://TASK-260822-96m5pj/TASK-260822-96m5pj_build-ssh-config.patch) — Full diff of internal/config (buildssh.go, buildssh_test.go, config.go) as landed and verified
- [TASK-260822-96m5pj_full-test-02.log](file://TASK-260822-96m5pj/TASK-260822-96m5pj_full-test-02.log) — go test ./... on the reconciled tree after landing the merged superset, exit 0, 31 packages ok
- [TASK-260822-96m5pj_spawn-log_-reviewer--reviewer--claude-_RUN-260822-b910b6.log](file://TASK-260822-96m5pj/TASK-260822-96m5pj_spawn-log_-reviewer--reviewer--claude-_RUN-260822-b910b6.log) — System spawn log captured by task-board
- [TASK-260822-96m5pj_review.md](file://TASK-260822-96m5pj/TASK-260822-96m5pj_review.md) — Reviewer verdict: accepted; AC/DoD evidence, independent verification table, 3 non-blocking findings
- [TASK-260822-96m5pj_reviewer-probe_test.go.txt](file://TASK-260822-96m5pj/TASK-260822-96m5pj_reviewer-probe_test.go.txt) — Reviewer probe tests run against a throwaway source copy; source of the determinism finding
- [TASK-260822-96m5pj_review-config-tests.log](file://TASK-260822-96m5pj/TASK-260822-96m5pj_review-config-tests.log) — Reviewer run of go test -v ./internal/config -run BuildSSH, 13 tests PASS, exit 0

## Created
2026-08-22T16:12:05Z

## Last Update
2026-08-22T16:38:02Z

## Assigned To
[reviewer] reviewer (claude)
