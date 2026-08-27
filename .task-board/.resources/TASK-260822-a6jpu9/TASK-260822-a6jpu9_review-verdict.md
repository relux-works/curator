# TASK-260822-a6jpu9 — reviewer verdict: ACCEPTED

**Run:** RUN-260822-2559c8 (reviewer, claude), read-only. No file in the worktree modified.
**Repo:** `/Users/iv/Developer/ReluxWorks/curator` @ `handoff/cocoaskills-parity-20260731`
**Spec:** `/Users/iv/Developer/ReluxWorks/curator-spec` @ `a2d44eb` (verified by `git rev-parse`)

## Verdict

**ACCEPTED** on the gap branch of the AC. Every load-bearing claim in the
implementer's report was re-derived here by command, not read from the report.

## AC mapping

| AC clause | State | Evidence re-derived by the reviewer |
| --- | --- | --- |
| Reproducing fixture with a vendored third-party Makefile | met | `internal/audit/vendor_test.go::vendorFixture` (0644 Makefile with `curl … \| sh`, 0755 vendor script, first-party `scripts/setup.sh`); `internal/install/install_test.go::TestAuditAdmitsVendoredThirdPartyText` (real git skill repo with `go.mod` + `vendor/modules.txt`) |
| install admits the vendored Makefile | met | both tests pass uncached under `mode=strict, fail_on=low`; install-level result `ok`, no warning naming `vendor-inert.example.com` |
| first-party text still blocks | met | negative control in the install test moves the identical curl-pipe line into `scripts/tool`, retags, and the install fails with `network-undeclared` — this also proves the gate was live in the positive case |
| an executable vendor script still blocks | **not met — recorded gap** | correctly deferred; see below |
| — or an evidence-backed note that the audit surface does not exist yet | met | see verification 1 |
| go test green | met | `go test -count=1 ./...` exit **0**, 31 `ok`, 0 `FAIL` |

## Verification performed (independent, not read from the report)

1. **The external-repository audit subject genuinely does not exist here.**
   `internal/skillspec/types.go:9` — `SupportedSchemaVersions = {1,2,3,4,5}`.
   `grep -rn -E "go-repository-v1|build_repository|skill-build|BuildRoot" --include='*.go' internal cmd`
   → exit 1, no matches. `grep -rn -E '"go"|go build|go list' --include='*.go' internal cmd`
   → no output. There is no build driver and no `go list`/`go build` child,
   so the "fixed build session" the question is framed around is unimplemented.
   The conditional in the task description ("*if* the audit blocks … *if no such
   audit exists*, record the gap") therefore fires on its second clause. Writing
   a demotion would have been a demotion of nothing.

2. **Spec citations are exact, not approximate.** Checked verbatim at `a2d44eb`:
   - `protocol/core.md:531` — "The whole external snapshot is validated, hashed, and audited."
   - `protocol/core.md:844` — "The consuming skill and each effective external repository are independent audit subjects."
   - `profiles/manager.md:616` — "Detectors and analysis backends may differ; decisions do not."
   - `profiles/manager.md:626-632` — the `allow`/`warn`/`block`/`require_pin` decision list, unchanged by this task.
   The "spec-legal" argument for leaving the scope alone holds.

3. **The detector scope claim is the code, verified.** `internal/audit/audit.go:307`:
   `if !strings.HasPrefix(posix, "scripts/") && posix != "csk-skill.json" { return nil }`.
   Literal prefix, first-party only.

4. **The finding is meaningful, not an artifact of vendor/ being absent.**
   `internal/install/install.go:261-277` builds each `audit.Subject` from
   `node.Snapshot` — the full repository snapshot, `vendor/` included. So the
   vendored files are physically present at audit time and are skipped by
   *scope*, not by absence. This is the check that makes the whole task's answer
   load-bearing, and it holds.

5. **The non-exploitability argument holds structurally.**
   `internal/whitelist/whitelist.go:19-22` — `IncludeRoots` is an allowlist that
   does not contain `vendor`, so no vendored byte reaches installed context.
   `internal/skillspec/parse.go:546` — a command path outside `runtime_roots` is
   rejected at parse. Both as claimed.

6. **Gates, each re-run standalone with a direct exit code** (not `${PIPESTATUS}`,
   which is empty under zsh):

   | Command | Exit | Result |
   | --- | ---: | --- |
   | `go build ./...` | 0 | builds |
   | `go vet ./...` | 0 | clean |
   | `gofmt -l internal cmd` | 0 | empty output |
   | `go test -count=1 ./...` | 0 | 31 `ok`, 0 `FAIL` |
   | `go test -count=1 -run Vendor ./internal/audit/... ./internal/install/... -v` | 0 | 5/5 PASS |
   | `golangci-lint run` | 0 | `0 issues.` |

## Why the missed AC half is not rework

The AC's first branch wants an executable vendor script to block. It does not.
Making it block means widening the detector scope beyond the literal `scripts/`
prefix, which changes install outcomes for **every existing skill** at schemas
1-5 — a policy decision, and precisely the "follow-up decision" the task
description routes the gap to. Forcing it inside this task would be the forced
fit the role contract tells the implementer to refuse. Recording it, with the
tightening condition written into the test comment, is the right call.

## Findings (non-blocking, for the follow-up owner)

- **F1 — the adjacent gap deserves a board item, and does not have one.**
  The detector scope is the literal `scripts/` prefix, not the manifest's
  declared `runtime_roots`. `runtime_roots` is arbitrary
  (`internal/skillspec/parse.go:482`, copied to the runtime store by
  `internal/runtimestore.InstallRuntimeRoots`), so a skill declaring
  `runtime_roots: ["bin"]` ships with **zero** audit coverage of its real
  executable surface, today, at schemas 1-5. Confirmed by reading
  `audit.go:298-310`: a path `bin/tool` matches neither branch and is never
  read. This is not vendor-specific and does not wait on schema 6/7. It is
  captured in `LOGBOOK.md` 2026-08-23 0001 and in the research note, but
  `STORY-260822-2jdj9o` has no child task for it (children:
  `27bvo4`, `5wfdfx`, `a6jpu9`). Recommend the orchestrator open one — a
  reviewer run should not widen the board unilaterally.
- **F2 — `TestVendorTextIsOutsideTheDetectorScope` is a deliberate
  change-detector.** Its `len(findings) != 1` assertion will fail the moment
  anyone widens the scope. That is by design and the test comment says so
  ("when the external repository audit subject of Spec §6.5 lands, the vendor
  executable expectation below must be tightened to a blocking finding"), but
  the follow-up owner should expect to edit this test rather than treat the
  failure as a regression.
- **F3 — cosmetic.** `TestGateAdmitsVendorTextAndBlocksFirstPartyText` joins its
  two positive assertions with `&&`, so it passes if *either* the file name or
  the host appears. Both do today. Harmless, worth tightening to `||` if the
  file is touched again.

## Fit with the project

Tests only, no product code — correct for a question whose answer is "there is
nothing here to change". `vendor_test.go` reuses the package's existing
`newCfg` helper and `Subject`/`Gate`/`detect` seams rather than inventing a
parallel harness; the install-level test uses the established `newEnv`/`e.git`/
`e.write`/`e.declare` fixture shape. Evidence is persisted in four places
(board resource, research note under `.research/`, `LOGBOOK.md`, task notes),
so nothing depends on conversation context surviving.

## Scope note

Read-only reviewer run. No `commit_ack` supplied. The working tree carries
uncommitted work from sibling tasks (`internal/config/buildssh*.go`,
`internal/closure`, `internal/skillspec/parse.go`, `cmd/curator/main.go`);
this task's delta is exactly `internal/audit/vendor_test.go` (new), the
`TestAuditAdmitsVendoredThirdPartyText` block appended to
`internal/install/install_test.go`, `.research/260823_vendor-inert-text-audit-policy.md`,
and the `LOGBOOK.md` 2026-08-23 0001 entry. The commit-owning mover lands them.
