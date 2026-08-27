# Flight Logbook

> Institutional memory. Concise, factual, high-signal.
> Newest entries first. One block per insight.

## 2026-08-27

### 0123 — `BUG-260823-1vx45a`: the helper protocol reported "someone holds this lock" whenever it meant "I ran out of time"

- ROOT CAUSE: `TestManagerLockHelper` gave one 200ms context to lock-state directory creation, lock-file opening, the 10ms retry loop of `acquireFileLock`, and — in build-key mode — two sequential acquisitions. `acquireFileLock` returns `ctx.Err()` for a deadline expiry no matter what caused it, and the helper serialized every expiry as `blocked`. An uncontended lock on a slow runner was therefore indistinguishable from a lock another process held.
- THE REPORT WAS WRONG, AND THE WRONG FRAMING COST A FULL CYCLE: the bug was filed as "`t.TempDir` cleanup hits a still-held `.lock` file" and an implementation was built on that premise before anyone checked it. No lock handle survives helper exit — `CombinedOutput` returns only after the child is gone. The real symptom was already recorded in this file at entry [[1908]] on 2026-08-23: `independent build key helper = blocked, want acquired`, on a commit that failed it in `pull_request` run `32641695064` and passed it in `workflow_dispatch` run `32641704975`. The diagnosis was sitting in the repository the whole time.
- HOW IT WAS SETTLED, BY MEASUREMENT: the unmodified base passed both named tests 40 consecutive times on a native Windows host (go1.25.5 windows/amd64) — so there was no handle-lifetime defect to fix. Shrinking that single deadline to 1ms on the same untouched base, with no contention present, reproduced the exact CI string three times out of three.
- FIX: the two expectations have opposite failure modes, so they get opposite deadlines. An expected-`blocked` probe keeps the bounded 200ms window — a slow host makes blocking *more* certain, and a held lock cannot be acquired at any deadline. An expected-`acquired` probe gets 30s, which costs nothing in the passing case because a free lock is taken on the first `tryFileLock`. Parent call sites now state which outcome they expect instead of one constant serving both.
- REGRESSION, MUTATION-CHECKED: `TestSubprocessExpectedAcquiredWithTinyDeadlineReportsBlocked` gives an uncontended expected-acquired shape a 1ns deadline and requires `blocked`, and separately asserts the acquired deadline stays strictly wider than the blocked one. Both halves were proved to have teeth: collapsing the constants to `200ms` fails with `acquired helper deadline 200ms must exceed blocked helper deadline 200ms`; widening the tiny deadline fails with `uncontended helper with tiny deadline = "acquired", want blocked`.
- ALSO FIXED: helper release errors were joined with acquisition errors and then classified with `errors.Is(err, context.DeadlineExceeded)`, so a failing `Close` was published as `blocked` and exited successfully. Release is now checked independently and fails the helper before any classification.
- NOTE, THE SAME FAMILY, LEFT ALONE DELIBERATELY: `TestCancellationReleasesPartialAcquisition`'s 50ms is safe — it holds the second lock and asserts cancellation, so slowness cannot turn that negative into a false acquisition. `TestProjectOrderInversionFailsBeforeWaiting` is *not* safe in the same way: it asserts `elapsed <= 100ms` after a lock-order rejection, which a sufficiently slow host can exceed while the code is correct. It has never been observed failing and is out of scope here, but it is the same conflation pointed the other way and it is written down so the next person does not rediscover it from a red run.
- EVIDENCE: revision 2 ran the full `internal/managerlock` package 10 consecutive times green on the native Windows host (16.6s). Producer records 20 consecutive `TestSubprocess*` runs each on native Linux/arm64 and native Windows/amd64 on byte-identical files. `go vet ./...`, `golangci-lint run`, `go build ./...` all 0. A broad `go test ./...` was terminated at roughly seven minutes with no package output (exit 143) and is recorded as interrupted, not passing; the package and platform gates are the acceptance evidence.

## 2026-08-25

### 0650 — An entire epic delivered with zero commits, spread across three divergent trees

- ANOMALY (`TASK-260825-1d0eo5`, the landing task for `EPIC-260825-2p2kl3`): seven reviewer-accepted tasks produced no commit anywhere. Every acceptance was granted against uncommitted files, and those files were split across three working trees that did not agree with each other. The primary checkout held the 02:2x–02:5x state of the source; the `STORY-260825-32bopo` worktree held a later 03:37–04:23 superset with `internal/install/buildhttpsprompt.go` and the SSH-prompt persistence change that the primary checkout never received; the `STORY-260825-39h6vz` worktree held the newest documentation. Nothing in the board recorded which tree was authoritative — that had to be reconstructed from file mtimes.
- CONSEQUENCE, the near miss: assembling the composite from the primary checkout — the tree three task notes point at — would have shipped a documentation page describing a candidate prompt whose implementation was not in that tree, and would have silently dropped the fix that stops the SSH prompt persisting an answer the operator did not ask to save. The docs task's own logbook entry carries a `CORRECTION TO THE DOCUMENTATION BOUNDARY` written when its author hit exactly this split.
- CONSEQUENCE, the one that reached the branch: the primary checkout also carried unrelated `.github/ci/*.sh` edits dated two days earlier. A whole-tree copy would have landed them under this epic's commits. Only files named by the epic's own task outcomes were carried.
- RULE THIS ARGUES FOR: a review that accepts uncommitted files accepts a state nobody can address later. Acceptance should name a commit, or the accepted files should be committed to the task branch before the verdict is recorded — otherwise the landing task is reconstructing the delivery by timestamp, and the reviewer's "I ran the gates" refers to a tree that no longer exists.
- SECOND ANOMALY, still unresolved: [[0627]]'s split lineage is why this epic's nine logbook entries reached `origin/main` only now, carried by this commit. They were written into the local 3000-line `LOGBOOK.md`, which shares no history with the file on `origin/main`. Entries `0052` and `0057` were also numbered as a sequence rather than as the file's `HHMM` convention; they are preserved verbatim rather than renumbered, so they sort below entries that were written hours before them.
- EVIDENCE: eight signed commits, each independently checked out and proved to `go build ./...` and `go vet ./...` at exit 0, so the branch bisects. Local gate set green before the pull request: gofmt, build, vet, `golangci-lint`, gate self-test (81 passed), no-broad-suppression, ledger consistency (80 rows across linux/darwin/windows), the `ci.yml` naming grep run verbatim, and `test-gate.sh` against `SPEC_PIN` `0ed5c691` planning 44/0/0 served/deferred/excluded.


### 0640 — `BUG-260825-11nmd5`: relaxing one directive turned an early-exit scan into a bypass for another

- CLASS: a byte scanner that stops at the *first* matching needle is only safe while every needle carries the *same* verdict. `scanSourceDirectives` (`internal/godriver/graph.go`) looked for `//go:cgo_import_dynamic` and `//go:generate` and returned on whichever it hit first. That was harmless for as long as both were rejected unconditionally.
- REGRESSION INTRODUCED BY THE FIX ABOVE IT: PR 40 (`c9fe49c`) made `//go:generate` *exempt* inside a materialized vendor tree. The scan's early exit then became a hole — a `//go:generate` in the first 64 KiB window set `matched = 2` and terminated the read, so a `//go:cgo_import_dynamic` in any later window of the same file was never seen, and the carve-out admitted the package. Reproduced end to end through `Build()` on an audited, non-replaced vendored module: diagnostic code `""`, build succeeded.
- THE GO COMPILER IS NOT A BACKSTOP HERE: `cmd/compile/internal/noder` permits `//go:cgo_import_dynamic` for general use (the comment names Solaris code in `golang.org/x/sys/unix`), and `/usr/lib/libSystem.B.dylib` satisfies its argument check. Preflight was the only thing standing there.
- FIX: the scan now resolves by **severity, not by first hit**. Only `//go:cgo_import_dynamic` — which nothing weaker can override — ends the read early; a `//go:generate` hit is recorded and the file is still read to EOF. The three severities are named constants (`directiveNone` / `directiveCgoImportDynamic` / `directiveGenerate`) so the call site reads as a verdict rather than as `matched == 1` / `matched == 2`. Cost: files that carry `//go:generate` are read whole, bounded by the frozen build source.
- FINDING, THE NARROWING MUTANT IS WHAT PROVES THE BOUND: reintroducing `return true` on the generate branch is delete-only and proves nothing about the class. The mutant that matters gates the cgo check on `matched == directiveNone`, i.e. keeps the "keep scanning" behavior but lets a recorded generate suppress a later cgo hit. It is killed by `TestDirectiveScanReportsTheStrongestDirectiveAcrossWindows/generate_before_cgo` and by the end-to-end case. Four mutants applied and reverted, all killed, including one that removes the carve-out entirely and reddens PR 40's own allowed-side test — proof the hardening did not quietly undo the relaxation.
- SPEC BASIS: profiles/manager.md §2.3 is a *containment* predicate — an active non-standard `GoFiles` file is scanned as exact bytes and rejected if it **contains** the directive, wherever in the file it sits. An early-exit scan silently weakens containment to "contains, in the prefix before some other token". Worth reading every other early-exit byte scan in the tree with that sentence in hand.
- SCOPE: `internal/godriver/graph.go`, `internal/godriver/graph_test.go`, `internal/godriver/moduleroots_test.go`. Branch `fix/BUG-260825-11nmd5-directive-scan-shortcircuit` off `680f6a6`.


### 0415 — Marker-schema banding: one predicate, or every reader drifts

- ROOT CAUSE: three readers each carried their own inequality on the install-marker schema instead of asking the marker package. `classifySkillBuilds` (`cmd/curator/builds.go:365`) admitted only the written schema; `markerRefusal` (`builds.go:545`) admitted schemas 1-2; `marks.absorb` (`internal/scopes/gc.go:213`) admitted 2-3. Marker v4 failed all three.
- FINDING: `marker.Current` was correct throughout, because it asked the shared `buildBearingSchema` predicate. The three that drifted were the three that hand-rolled the check. The predicate existed; it was just unexported.
- FINDING: the GC one was the dangerous half and went unreported. A marker v4 contributed no live build reference, so a maintenance pass could delete protected cache entries a live schema-8 installation was still running from. Found only by grepping the whole class, not from the bug report.
- FINDING: a stale banding check produces a *self-contradictory* remedy, which is the tell. "install marker schema 4 cannot describe a compiled command; reinstall to record marker schema 2" — the manager would never write schema 2 for that band.
- FIX: exported `marker.BuildBearingSchema` / `marker.SupportedSchema`, added `marker.NewestSchemaVersion`; all three readers ask them. `internal/marker/marker.go`, `cmd/curator/builds.go`, `internal/scopes/gc.go`.
- STATUS: PR #41, resolved pending merge.

### 0416 — Two tests asserted the defect instead of catching it

- REGRESSION: `TestMarkerRefusalSeparatesUnsupportedFromInvalid` pinned `{"schema_version":3}` as `unsupported-marker` — i.e. it asserted that a schema this manager reads is unreadable. The test passed for the whole life of the bug and would have blocked the fix.
- FINDING: the status matrix case named "marker schema cannot be read by this manager" pinned `marker.SchemaVersion + 1`. That is *written* + 1, not *readable* + 1. The moment a newer schema became readable it stopped testing the unsupported band and started testing a readable one. Pin `NewestSchemaVersion + 1`.
- NOTE: general shape — a negative test anchored to the value a system *writes* silently stops being a negative test when the *read* band widens. Anchor negative schema/version tests to the edge of the accepted band.
- SCOPE: `cmd/curator/builds_test.go`, `cmd/curator/status_test.go`.

### 0417 — Conformance suite cannot catch a one-schema reader

- FINDING: `curator-spec` at pin `0ed5c691` (rc.9) proves the marker write side and the currentness side separately and never crosses them with the schema as the variable.
- FINDING: `expected/install-marker-v4.json` pins what a schema-8 install must *record*, and says nothing about what a reader must then *accept*. `vectors/manager-lifecycle.json` → `status_cases` has two cases: `compiled-installation-current` lists `marker-schema` among its `validated` steps but never parameterises which schema; `compiled-currentness-failure-matrix` enumerates 14 `independent_conditions` and none is a marker-schema condition. The `go-build-skill` fixture carries no `schema_version` field at all. `gc_cases` has no marker-schema-parameterised liveness vector.
- NOTE: a manager can be conformant on every published vector while reporting a successful install as needs-install. Gap written up for the suite owner in the BUG-260825-1l1st9 board artifact.
- STATUS: pending — needs suite-owner action, not a curator change.

### 0316 — TASK-260825-3n4bjj review: ambient GIT_ASKPASS no longer selects the HTTPS askpass source

- FINDING (reviewer): `productionGitTool` (`cmd/curator/main.go:1402`) dropped the `admittedOperatorFile(os.Getenv("GIT_ASKPASS"))` read that dated to the schema-7 integration; `GitTool.AskPass` is now always the manager's own executable. Consequence an operator may trip on: exporting `GIT_ASKPASS` no longer authenticates a private HTTPS build fetch — `config build-https` scopes and `CURATOR_BUILD_HTTPS_TOKEN`/`_HOST` are the only selection surfaces.
- DECISION: deliberate and load-bearing, not drift. `materializeHTTPSCredentialBroker` copies `Tool.AskPass` as the broker source and the basename dispatch answers prompts only from the manager binary; honoring an ambient helper there would hand `CURATOR_BUILD_HTTPS_ASKPASS_SECRET` to a foreign program inside the fetch environment. Also the shape core Spec §12.2 requires: no ambient, identity-unbound credential selection.
- STATUS: review verdict ACCEPTED. Independent gates: full `go test -timeout 30m -count=1 ./...` 42/42 exit 0 (`.temp/TASK-260825-3n4bjj/review-go-test-full-01.log`), `golangci-lint run ./...` 0 issues, `go vet`/`gofmt`/`no-broad-suppression.sh` clean, real-binary broker probes (both prompts answered; foreign host/prompt, absent/symlinked/relative/unknown-field state all silent exit 1; normal-name binary refuses prompts with usage on stderr only).

### 0234 — TASK-260825-2gyhq8: `config build-https` picked up already-landed sibling APIs mid-flight in a shared, uncommitted checkout

- SCOPE: `curator config build-https add|login|list|remove` in `cmd/curator/main.go`, plus `config.SetBuildHTTPS`/`RemoveBuildHTTPS` in `internal/config/write.go` (mirroring the `build_ssh` Set/Remove pair). Consumes `internal/config/buildhttps.go` (`TASK-260825-168m7o`) and `internal/gitcred` (`TASK-260825-1tgpcn`) rather than reimplementing either.
- FINDING: this task started with those two dependencies, plus `internal/install/buildhttps.go` (`TASK-260825-1lausy`), already present but **uncommitted** in the working tree — three sibling board tasks under `EPIC-260825-2p2kl3` land into the same shared checkout in parallel rather than isolated worktrees. `1lausy`'s own entry (0055 below) records hitting a compile-time race against a concurrent CLI worker's half-written function — that worker was this task. No conflict this round, but the pattern (git status before trusting any file's completeness, treat "exists" as "maybe mid-edit" until `go build` confirms it) is worth carrying into any task in this epic that reads a sibling's file rather than the board API.
- DECISION: `add` exposes three mutually exclusive flags (`--git-credentials`, `--keyring`, `--token-env NAME`) and no flag accepts a literal token — the config's `token` field only ever holds an enumerated source name, never a secret, so there is no CLI surface that could carry one even by accident (core Spec §12.2).
- DECISION: `list`'s "present" column is a **live** probe (`ReadScoped`/`ReadHost`/`os.Getenv` per source), not a config-file fact — a keyring entry can be dropped from the credential store, an operator's own git-credential can be rotated away, or a `token_env` can go unset, independently of the recorded scope selection surviving.
- DECISION: `remove` deletes the stored secret only for a `token=keyring` scope. A `git-credentials` scope names the operator's own credential, which the manager did not create and must not delete; pinned by `TestConfigBuildHTTPSRemoveNeverTouchesTheOperatorsOwnGitCredential`.
- GATES: `go build ./...` 0, `go vet ./...` 0, `golangci-lint run ./internal/config/... ./cmd/curator/...` 0 issues, `go test ./internal/config/... ./internal/gitcred/... -v` all green, `go test ./cmd/curator/...` green (exit 0, ~550s, matching this package's known ~10min suite duration). New tests: `internal/config/write_test.go` (Set/Remove roundtrip + rejection), `cmd/curator/main_test.go` (add/list/remove matrix, login round-trip against a real isolated `credential.helper=store`, help-text precedence/disclosure ordering).

### 0223 — Killing git does not bound a credential call: the helper is git's child and outlives the kill holding the pipes

- REGRESSION (`TASK-260825-1tgpcn`, caught in review cycle 1): supersedes the "(15s bound)" claim in entry 0158 below — that bound was not delivered. `Access.call` used `exec.CommandContext` with `cmd.WaitDelay` at zero. The deadline kills **git**; a credential helper is git's *child*. Because `cmd.Stdout`/`cmd.Stderr` are not `*os.File`, os/exec wires OS pipes with copy goroutines, and the helper inherits the write end of git's stderr. A wedged helper — locked keychain, no desktop session, i.e. exactly what the deadline exists for — survives the kill still holding that pipe, and `cmd.Run` returns only when the orphan does. Every entry point sat on that path (`Discover` once per scope), so one wedged helper wedged the manager, undismissably.
- FIX: `internal/gitcred/gitcred.go:62` new `drainDelay = 2 * time.Second`, `gitcred.go:275` `cmd.WaitDelay = drainDelay`. One credential call now costs at most `Timeout + drainDelay`.
- FINDING, the reason it went unnoticed: the existing `TestACallIsBounded` wedges the stand-in git *itself*, and killing that closes its own pipes — the bad case is invisible unless the hanging process is a **grandchild**. Harness extended with `modeHangGrandchild` (stand-in git spawns a helper and blocks on it) and `modeSleeper` (that helper: the test binary re-executed once more, handed the stand-in git's own stderr, sleeping 30s, recording nothing so it is not counted as a credential call).
- STATUS: measured on darwin/arm64, go1.25.5, 200ms deadline against a 30s orphan — **30.008s** without `WaitDelay`, **2.21s** with. `TestACallIsBoundedWhenTheHelperOutlivesGit` verified to fail at 30s with the one line removed, so the fix is pinned against a refactor back to `exec.CommandContext` alone. Package tests, `-race`, `go test -timeout 30m ./...` (42 pkgs), `go vet ./...`, package lint and gofmt all exit 0. Handed back to review.
- NOTE, generalizes past this package: any `exec.CommandContext` in this repo whose child may spawn its own child, wired to non-`*os.File` stdout/stderr, has the same hole. The context bounds the process, `WaitDelay` bounds the *call*. Worth a sweep of the fetch and driver paths.

### 0158 — Git's own `store` helper prepends, so a manager entry shadows the operator's own on a host-level read

- SCOPE (`TASK-260825-1tgpcn`): new `internal/gitcred` — operator HTTPS credential reads/writes through `git credential fill|approve|reject` only. Consumed next by `TASK-260825-3n4bjj` (broker), `1lausy` (per-repository resolution), `2gyhq8` (CLI). Depends on nothing in the repo, so none of them can hit a cycle.
- FINDING: `git-credential-store` writes a new record to the **front** of `~/.git-credentials`. Once a manager entry (`username=curator-build-https:<scope>`) exists for a host, a username-less `git credential fill` for that host answers with the manager's record, not the operator's. Verified against git 2.50.1. Consequence: `ReadHost` refuses any answer whose username carries `NamespacePrefix` — the manager's own material must never be its own evidence that the operator configured a credential. Fails closed (no operator credential) rather than wrong. Deleting the manager entry restores the operator read. Pinned by `TestRealGitKeepsTheManagerEntrySeparate` in `internal/gitcred/gitcred_test.go`.
- FINDING: `GIT_TERMINAL_PROMPT=0` alone does **not** disable prompting. Git falls through terminal → `core.askPass` → `GIT_ASKPASS` → `SSH_ASKPASS`, and it reads an *empty* askpass variable as unset, continuing to the next source. So both askpass variables are removed from the inherited environment (not emptied) and `-c core.askPass=` plus `-c credential.interactive=false` are passed on every call.
- FINDING: with no `credential.helper` configured at all, `git credential approve` exits **0** and persists nothing — reproduced with real git in `TestRealGitWithoutAHelperIsCaughtByTheReadBack`. That is why every write is proved by reading it back before a scope is reported as configured; the platform-store failure the read-back was designed for is not exotic, an unconfigured home does it.
- DECISION: reads never error — absent material, a helper failure, a missing git, a helper that hangs (15s bound) all report "nothing here", so a resolution degrades instead of blocking a run on a dialog nobody is watching. Writes do error, with per-platform store guidance.
- NOTE: test harness worth reusing — the stand-in `git` is this test binary re-executed via `TestMain` + an env marker, serving the credential protocol out of a JSON store with selectable defects (approve-and-persist-nothing, answer-a-username-not-asked-for, accept-a-rejection-and-keep, never-answer). No shell script, no build step, runs identically on Windows.
- STATUS: `go test -timeout 30m ./...` exit 0 (42 packages), `-race` on the new package green, `go vet ./...` and `golangci-lint run` clean. Handed to `to-review`.

### 0130 — build_https config field lands with no prior enum to reuse; picked git-credentials/keyring to match the still-unbuilt resolver tasks

- DECISION (`TASK-260825-168m7o`): `internal/config/buildhttps.go` adds `Config.BuildHTTPS`, a scope→{token?, token_env?, username?} map mirroring `build_ssh`. No token-source enum existed anywhere in this repo yet, so the two enumerated `token` values were picked to match the two-source language ("the operator's own entry for a host" vs. "a manager-namespaced entry") already written into the still-unimplemented sibling tasks `TASK-260825-1lausy`/`1tgpcn`/`2gyhq8`: `TokenSourceGitCredentials = "git-credentials"` and `TokenSourceKeyring = "keyring"`. `token_env` is the one field allowed to hold operator-chosen text (a literal env var name); `token` accepts only those two source names, and a literal secret pasted into `token` is rejected with "...; secrets never live in the config".
- NOTE: the two source names, the exactly-one-of `token`/`token_env` rule and the reuse of the `build_ssh` scope grammar were chosen to match the shape the Curator Protocol core §12.2 already describes rather than invented here. No code from any other implementation was copied into this repository.
- FIX/SCOPE: `internal/config/buildssh.go` — extracted the longest segment-aware prefix match into a generic `longestScope[T any]` helper; `MatchBuildSSH` and the new `MatchBuildHTTPS` both call it, so `build_ssh` and `build_https` share one matching implementation instead of two. `build_https` added to `managerKeys` in `config.go` but deliberately left out of `LockableKeys` — same rule as `build_ssh`: an operator credential selection, not lockable org policy.
- STATUS: `go test ./internal/config/...`, `go vet ./...`, `golangci-lint run ./...` all green. Handed to `to-review`. `TASK-260825-1lausy` (resolution precedence incl. host-bound override) and `TASK-260825-2gyhq8` (CLI add/login/list/remove) are the tasks that will actually exercise these source names end to end — worth checking their review lands consistent with this naming before it's load-bearing.

### 0057 — TASK-260825-2fy132: HTTPS credential operator documentation

- DOCUMENTATION BOUNDARY: HTTPS selection is resolved for the whole planned closure before its first repository fetch, but an uncovered HTTPS repository remains anonymous. The delivered surface has no install-time HTTPS credential discovery or interactive candidate prompt; only explicitly configured sources and the `login` command can select a token. The operator page states this boundary instead of borrowing the SSH prompt behaviour.
- SECURITY: `CURATOR_BUILD_HTTPS_TOKEN` without `CURATOR_BUILD_HTTPS_HOST` is identity-unbound and can reach every HTTPS build-repository host in the run. The same plain-language warning is present in the operator page and the CHANGELOG entry, with `Spec core §12.2` cited.
- EVIDENCE: delivery-checkout `go run ./cmd/curator config build-https --help`, `add`, and replacement transcripts all exited 0; primary-checkout `make lint` exited 0. The isolated docs worktree lacks the untracked `agents/skills/skill-go-testing-tools/tuitestkit` Go replacement, so its `make lint` exited 2 before linting the changed documentation.
- CORRECTION TO THE DOCUMENTATION BOUNDARY: the previous boundary incorrectly
  denied the shipped terminal candidate prompt. Before the first fetch,
  an operator-terminal run offers every uncovered HTTPS repository the
  presence-only detected Git host credential (when available) or entry of a
  token now. The operator explicitly selects a candidate and then either
  persists its scope or uses it for this run only; abort stops the run. Only
  headless, non-terminal, and dry-run runs leave an uncovered repository
  anonymous. The operator page now states the shipped behaviour.

### 0056 — TASK-260825-3n4bjj: HTTPS credentials exist only inside one fetch process tree

- BROKER BOUNDARY: an authenticated HTTPS fetch materializes a private hardlink/copy of the manager binary under the operation root and dispatches it by the fixed `curator-build-https-askpass` basename. Its sibling JSON state contains exactly `host` and `username`; the secret is supplied only through the fetch environment. Init, validation, and object-proof Git children receive neither the state path nor the secret.
- FAIL-CLOSED PROMPTS: the broker emits bytes only for Git's exact `Username for 'https://<host>': ` and `Password for 'https://<username>@<host>': ` prompts. A foreign host or prompt, extra argv, absent secret/state, non-regular/symlinked/malformed/oversized state, or output failure returns non-zero without output.
- TWO ASKPASS SURFACES: authenticated fetch construction replaces `GIT_ASKPASS` and sets `core.askPass` to the same materialized wrapper. This is load-bearing because command-line `core.askPass` wins over the environment. The existing pinned URL, `http.sslVerify=true`, and `http.followRedirects=false` arguments remain unchanged. Anonymous HTTPS takes the pre-existing argv/environment path and receives no broker material.
- END-TO-END EVIDENCE: a real bare Git repository served by a local TLS server requiring Basic Auth was queried by the real Git client through the materialized broker; the locked commit was returned only after authenticated requests. Separate acquisition instrumentation proved state/secret presence only on `fetch` and absence on every other Git child.
- VALIDATION: focused broker/TLS/fetch tests 0; focused install wiring tests 0; full `go test -timeout 30m -count=1 ./...` 0; `golangci-lint run ./...` 0; native `go build` 0; Windows amd64 cross-build 0; scoped `git diff --check` 0. The first broad package checkpoint was manually interrupted at exit 1 after unrelated concurrent suites caused prolonged I/O contention; the later full suite is the authoritative green run. Logs and build artifacts: `.temp/TASK-260825-3n4bjj/`.

### 0055 — TASK-260825-1lausy: a host pin narrows only the override, not resolution

- HTTPS run-wide material now has an optional exact canonical-host pin. The security-sensitive rule is that a repository on another host does not stop resolution and does not receive the override: it proceeds as though the override were absent, so its longest `build_https` scope may still authenticate it and an unmatched public repository remains anonymous.
- Capturing environment-backed credentials means capturing both the run-wide token and every configured `token_env` at process entry. Retaining an environment callback would let later project-controlled work change credential material mid-run. All captured/resolved secret-bearing types implement redacted diagnostic formatting; tests cover `%v`, `%+v`, and `%#v`.
- Selected material is fail-closed in three distinct ways with source-specific remedies: set the named `token_env`, populate the host credential with `git credential approve`, or run `curator config build-https login <scope>` for a manager-namespaced entry. No selection is deliberately not a fourth error because anonymous HTTPS is a valid transport, unlike SSH.
- VALIDATION: focused resolver tests 0; full `go test -timeout 30m -count=1 ./...` 0 on the stable retry; `go build ./...` 0; `go vet ./...` 0; `golangci-lint run ./...` 0. The first full-suite attempt exited 1 while a concurrent CLI worker had written its import/dispatch but not yet its function body; a compile-only rerun and the full stable rerun passed after that shared-checkout write completed. Logs: `.temp/TASK-260825-1lausy/`.

### 0052 — TASK-260825-3kb532: HTTPS install precheck and run-only persistence boundary

- DECISION: unmatched HTTPS repositories are collected and resolved before the first acquisition. A terminal resolver receives presence-only Git credential discovery, then requires an explicit existing-credential or enter-token choice; a nil resolver is the headless path and leaves every unmatched HTTPS repository anonymous.
- PERSISTENCE BOUNDARY: the shared scope question defaults to the repository namespace, `r` returns that scope for this process only, and only an explicit saved scope reaches the persistence callback. This closes the same unconditional-persist defect on the older SSH prompt. Tests compare real config and isolated Git credential-store bytes before/after both HTTPS run-only surfaces and the SSH run-only path.
- WORKSPACE ANOMALY: accepted HTTPS dependency deltas (`build_https` config, `gitcred`, resolution, config CLI, broker/fetch wiring) existed only as uncommitted files in the primary checkout while their board tasks were already reviewer-accepted; both managed Story worktrees were clean at the base commit. Only the accepted source/test files named by those tasks' outcome evidence were mechanically brought into this Story worktree so this dependent task could compile and test. No unrelated primary-checkout changes were copied.
- GATE EVIDENCE: fresh `cmd/curator` and `internal/...` package runs cover exactly the root module package list and exited 0; those two exact bounded commands support the Go-test checklist under the headless split-run contract. Build, vet, lint, gofmt, and diff checks exited 0. The first `internal/...` run exited 1 because the tracked testing-tools submodule was uninitialized in this isolated worktree, then passed after `git submodule update --init --recursive agents/skills/skill-go-testing-tools`. Extra root runs at default parallelism and `-p 2` were both terminated at exit 143 after exceeding the headless shell budget; neither is presented as green and they add no package coverage beyond the two successful disjoint runs.


## 2026-08-23


### 1307 — Ratified: operator credential selections are never lockable

`build_ssh` stays outside LockableKeys by decision of the spec owner (2026-08-23): credential material is operator-owned, and a system configuration must not select or constrain it. The manager-profile system-configuration clause now states this explicitly; the peer implementation records the same rule at its lockable-keys definition.

## 2026-08-25

### 0330 — cmd/curator cannot be parallelized: 98.7 % of its wall-clock is behind two process globals (TASK-260824-1kzj22)

- SCOPE: test code only — `t.Parallel()` added to 36 cases in `builds_test.go`, `main_test.go`, `status_test.go`, `toolchain_host_test.go`, plus a doc comment on `capture` (`cmd/curator/status_test.go`). No production file touched.
- FINDING: of 70 test functions, 36 are genuinely independent and hold **5.18 s**; the other 33 hold **381.03 s** (98.7 %). Every one of the 33 drives the CLI through `run()`, which resolves the manager config from the process-global `CURATOR_CONFIG` (`cmd/curator/main.go:94` → `config.Load("")`) and writes through process-global `os.Stdout`/`os.Stderr` via `capture`. Go panics both ways on `t.Setenv` + `t.Parallel()`, and two concurrent `capture` calls would cross streams. There is no in-process ordering that satisfies both.
- FINDING: the 4-minute target is unreachable independent of machine speed — `TestCompiledProjectStatusRepairRollbackRecovery` alone measures 230–268 s (subtests 80.85 / 72.11 / 64.95 / 20.07 / 17.02 s) and its five subtests corrupt and repair **one shared installed fixture**, so they cannot be split either.
- ROOT CAUSE of the cost: not test overhead, cold Go compilation. `godriver` gives every build session a hermetic `GOPATH`/`GOMODCACHE`/`GOCACHE` (`internal/godriver/session.go:411`) and `internal/godriver/guards_test.go:60` explicitly rejects a shared cache, so each compiled-command fixture pays a full stdlib compile.
- DECISION: stop at the safe parallelization rather than exceed scope. Reaching ≤ 4 min needs one of: injectable config path + output streams on `run()` (production change), a subprocess CLI harness (no production change but collapses in-package coverage attribution unless the suite moves to `GOCOVERDIR`), relaxing the hermetic per-session `GOCACHE` (contradicts the guard that exists to enforce it), or sub-packaging (needs `run()` importable). Recommended: injectable `run()`.
- FINDING: coverage is provably untouched — both trees report 57.1 %, and a block-level profile diff is exact: 778 blocks each side, 695 / 1218 statements covered on both, zero differing blocks.
- ANOMALY: wall-clock measurement on this host was unusable during the window. Other agent sessions ran their own `go test` passes concurrently (up to nine foreign `*.test` processes, 1-min load 11 → 51); the same tree varied 305.6 s … 543.7 s. The per-test ratio above is taken from inside a single run and is therefore contention-independent; the absolute run numbers are not.
- STATUS: ready for review. 3 consecutive `go test -timeout 30m -count=1 ./cmd/curator` runs exit 0 (543.7 / 374.0 / 470.3 s), focused `-race` on the 36 parallelized cases exit 0 with no data race, golangci-lint v2.12.2 `0 issues.`, gofmt/vet/`git diff --check` clean, `task-board validate` clean.


### 0100 — Cross-adapter integration: what one suite can honestly prove, and what it must delegate (TASK-260811-x611eq)

- SCOPE: new `internal/crossconformance` (standard-library-only production files, guarded), `docs/source-closure-adapter-conformance.md`, README section and gates row. No accepted adapter production file touched.
- DECISION: the integration oracle imports **no repository package**. `TestIntegrationProductionSourceImportsNoRepositoryPackage` enforces it, so the independent CCJ-1 scanner/emitter in `ccj.go` cannot end up calling `internal/protocoljson` — the encoder it exists to check. Its 53-record summary now matches the accepted Ruby oracle byte for byte from two independent implementations.
- FINDING: `cgp05.platform.darwin` and `cgp10.platform` are one record — same label, same CCJ bytes, one identity. Treating a repeated identity as a corpus defect (the first parser did) is wrong; only two *different* labels or payloads colliding on one identity breaks domain separation. `internal/crossconformance/corpus.go:131`.
- FINDING: the shared artifact corpus is **not** identical across adapter profiles for positive admission, and must not be. A Go or Python source fixture is legitimately `opaque.unknown` under `rust-source-v1`, because the accepted policy lets an adapter narrow its allowed source grammars. The cross-adapter claim is the accepted C12 one — 70 bare compiled leaves return one identical class/decision/code/leaf digest through all six profiles — plus its complement, that each profile admits exactly its own 8 source vectors. Asserting blanket agreement would have forced either a weakened profile or a false claim.
- DECISION: three published rejection vectors (`network-attempted`, `undeclared-write`, `output-drift`) are **delegated, not driven**. `closureexec.Executor` needs a provider-issued `Audit` and `artifactpolicy.LocalOutputAuthorization` is a sealed interface with an unexported method — an integration package could reach them only by forging evidence. They stay in the published matrix with their owning packages named, and a compile-time reference to each owner's own diagnostic constant (`guard_test.go`) breaks the build if a code is renamed away.
- FINDING (pre-existing, not introduced here): `internal/rustsource` already hard-requires the pinned Cargo `1.91.0` / `ea2d97820c16195b0ca3fadb4319fe512c199a43` descriptor — `NewManager` fails without it and the accepted conformance tests `t.Fatal`, with no matching class in `.github/ci/skip-classes.tsv`. `.github/workflows/ci.yml` installs Go only. So `internal/rustsource` is red on every CI runner today, and the cross-adapter Rust cases inherit that. Recorded rather than papered over: a host-conditional skip is exactly what `skip-classes.tsv` polices, and dropping the manager would leave Rust's causal-evidence obligation with no receipts.
- NOTE: modern Yarn reconciles an embedded `package.json` against the lock **verbatim**, so a fixture package's dependency ranges must carry the `npm:` protocol prefix the lock records. Feeding a normalized cache ZIP (rather than a tgz) also makes `CacheChecksum` computable outside the package, since `normalizeCacheZip` is a no-op for ZIP input.
- STATUS: all gates green locally — repository suite (53 packages) exit 0, `cmd/curator` exit 0 in 316.568s, race on the new package exit 0, lint 0 issues, gofmt/vet/`git diff --check` empty, `task-board validate` clean.

## 2026-08-24

### 2350 — A source-only macro oracle is bypassed by the build system one level down (TASK-260811-tkurtl, round 12 rework)

- ROOT CAUSE: rounds 11's two closing rules for the macro-expanded identifier positions (`atPositionIdentifiers` at a `#define`, `rejectMacroDefinedModuleNames` over the scanned closure) both read an oracle populated **only by source `#define`s**. A SwiftPM `.define` build setting is a macro the compiler binds that no admitted file spells, so both closed positions reopened through it.
- FINDING: verified on Apple clang 21.0.0 — `-Dprotocol=import` + `@ protocol SecretKit;` builds SecretKit and reads its header; `-DNoSuchKitXYZ=SecretKit` + `@import NoSuchKitXYZ;` does the same and clang prints `note: expanded from macro 'NoSuchKitXYZ'` at `<command line>:1`. The second is the exact wrong-module evidence defect N2 exists to close, reached through a different oracle input.
- FINDING: `swiftpmsource.decodeBuildSetting` folds every setting to `Kind:"swiftpm-setting"` and flags only `unsafeFlags`, so the declared kind survives only inside the retained raw JSON. Two consequences: the whole kind axis was invisible, and `cxxInteropSetting` — which read the folded `Kind` — could see an `.interoperabilityMode(.Cxx)` from a fixture record but never one SwiftPM actually emits.
- DECISION: closed the **axis**, not the member. `settingKindDisposition` (`internal/swiftpminterop/buildsettings.go`) admits a kind only when provably macro-inert AND resolution-inert. The axis is small enough to prove: exactly one non-`unsafeFlags` kind reaches the compiler as `-D` (`define` — routed into both oracles) and exactly one as `-I` (`headerSearchPath` — rejected). Link kinds are gated on component declaration; nine kinds are inert; an unnamed or unreadable kind rejects, with `settingReject` as the zero value so an absent table entry cannot read as inert.
- FINDING: the emitted kind axis was enumerated empirically, not from memory — `swift package dump-package` over a manifest declaring every `PackageDescription` setting on SwiftPM 6.3.2 gives 14 distinct kinds and the encoding `{"kind":{"<name>":{"_0":…}},"tool":"…"}`. `swiftLanguageVersion` is a deprecated alias already serialized as `swiftLanguageMode`.
- NOTE: the transferable lesson is that a static-analysis security oracle must state its **provenance**, not just its positions. The round-11 position table was exhaustive and every row's disposition was correct; it was incomplete one level down, at the question of where the answer came from. `scanIncludes`'s step-4 comment now names both oracle inputs.
- SCOPE: `internal/swiftpminterop/buildsettings.go` (new), `interop.go`, `headers.go` (comment), `platform.go`, `buildsettings_test.go` (new, H24, 39 subtests). `IncludeGrammarID` stays v10 — the per-file scanner grammar is unchanged. Focused suite 431 PASS / 0 FAIL.


### 2343 — Fixing the SwiftPM read-set drop and object disambiguation: both defects were one level below their own oracle (TASK-260811-2qfnai rework)

- FIX: `internal/swiftpmbuild/readset.go` — `mapObservedRead` now receives the **full** per-package admitted-root map instead of only the root entry. A read under `.curator/scratch/checkouts/<identity>/…` is rewritten to that dependency's admitted protected root; one with no matching identity, or a bare `checkouts` read, fails closed with `swiftpm_header_input_undeclared`. The `inBuildTree` drop is now reserved for genuinely derived state (module cache, generated module maps).
- FIX: `internal/swiftpmbuild/build.go` — produced objects are disambiguated on the **target-source-root-relative** source path (`targetRelativeSource`), compared for equality rather than by suffix. `swiftpmsource.Target.SourceRoot()` applies SwiftPM's convention default (`Sources/<Name>`) exactly once and `enumerateTargetSources` reuses it, so manifest normalization and build reconciliation cannot disagree.
- DECISION: exact resolution alone would have made undeclared local generation *invisible* rather than merely ambiguous, so `requireNoUndeclaredObject` now exhausts the produced set — any `.o` below a selected target build directory that no declared slot claims is `artifact_local_output_unreceipted`. Fixing the matcher without this would have traded a false rejection for a false acceptance.
- FINDING: the accepted layout claim survives the real toolchain — Apple Swift 6.3.2 with `Sources/CLib/{a,b}/shared.c` in one Clang target emits `a/shared.c.o` and `b/shared.c.o`, and the resolved set exactly exhausts the produced set for both the Clang and the multi-source Swift target.
- FINDING: `binding.go`'s treatment of an edge with no activation record as selected is **required**, not an oversight — `closuregraph/projection.go:139` emits activations for conditional edges only, and `closuregraph/validation.go:395` rejects an activation record on an unconditional edge. Documented rather than tightened.
- NOTE: mutation-checked the new tests before trusting them. With F1's checkouts branch disabled and F2's narrowing reverted to the old suffix form, all four new/extended tests fail. A fail-closed test that has never been observed failing is not evidence.
- SCOPE: `internal/swiftpmbuild/{readset,build,binding,plan,types}.go`, `internal/swiftpminterop/{types,interop}.go`, `internal/swiftpmsource/{types,executor_runtime}.go`, plus `conformance_test.go`, `swiftpmbuild_test.go`, `fixture_test.go`, `swift_integration_test.go`.
- STATUS: handed to review. Gates exit 0: focused suite, real-toolchain vector, race, suite minus `cmd/curator` (52 ok), `golangci-lint` v2.12.2 `0 issues.`, gofmt, vet, canonical verifier, `git diff --check`, `task-board validate`.

### 2324 — Two SwiftPM build-adapter defects: dependency reads dropped, object disambiguation compares mismatched path roots (TASK-260811-2qfnai review)

- FINDING: `internal/swiftpmbuild/readset.go:321` treats any observed read below `<work-copy>/.curator/` as build-tree noise and drops it. Verified on real Apple Swift 6.3.2 that SwiftPM materializes source-control dependencies into `<scratch>/checkouts/`, which under the planned `--scratch-path .curator/scratch` is exactly that prefix — so every transitive dependency package's source and header reads are discarded from the observed read set. A `CDep.build/*.d` for a `file://` dependency lists `.../.curator/scratch/checkouts/origin/Sources/CDep/{dep.c,include/CDep.h}`; both are dropped, leaving only the absolute SDK modulemap.
- ROOT CAUSE: the drop rule's justification ("locally produced build outputs") is true for the module cache and generated module maps but false for dependency source checkouts. `admittedPackageRoots` (readset.go:246) already builds a per-package root map and only `packageRoots[rootIdentity]` is ever used — the per-package entries are computed and thrown away.
- FINDING: `swiftpminterop.verifyReads` (`boundaries.go:194`) asserts only that every observed read resolves to something declared. It asserts no coverage, so a read set emptied of its package still reports `Reads.Mode == "observed"`. Containment-only verification cannot detect a truncated read set — a coverage assertion is a separate obligation.
- FINDING: `internal/swiftpmbuild/build.go:320` disambiguates ambiguous produced objects with `HasSuffix(candidate, slot.Source+".o")`, where `candidate` is relative to `<Target>.build` (`a/x.c.o`) and `slot.Source` is package-relative (`Sources/CLib/a/x.c`). The suffix can never match, so the branch is dead and always empties the match set. Reproduced on the real toolchain: a Clang target with `a/x.c` and `b/x.c` produces `a/x.c.o` and `b/x.c.o`, and both slots fail `artifact_local_output_unreceipted`. A legal `swiftpm-source-v1` package is therefore unbuildable.
- ANOMALY: `TestUndeclaredGeneratedObjectFailsClosed` passes for the wrong reason — its planted `nested/lib.c.o` makes the narrowing eliminate *both* candidates. A fail-closed test that green-lights on zero matches cannot distinguish a rejected ambiguity from a broken matcher.
- FINDING: latent-only today. `ObserveReads` returns early outside `AssuranceVerified`, and `closureexec.NewOSBoundary` fails closed on every platform, so no verified run can execute. The verified observation pass (`ReadSetObserver.observe`, `harvestDependencyFiles`) consequently has zero test coverage — which is how the first defect survived.
- NOTE: verified independently that the accepted parts hold — `--triple arm64-apple-macosx26.0` really does produce a `.build/arm64-apple-macosx/` scratch dir (so `unversionedTriple` is right), and `description.json` really is emitted at `<scratch>/<unversioned-triple>/<configuration>/` (so the observation pass's declared evidence path is correct).
- DECISION: verdict changes-requested -> `to-dev`; all other scope items accepted. Reviewer gates rerun green: suite minus `cmd/curator` (52 ok), focused suite, real-toolchain vector, `golangci-lint` 0 issues, gofmt, vet, canonical verifier, `task-board validate`.
- SCOPE: `internal/swiftpmbuild/readset.go`, `build.go`, `conformance_test.go`. Verdict artifact `TASK-260811-2qfnai_review-verdict_RUN-260824-145133.md`.


### 2300 — Publication requires observable outputs, and SwiftPM emits one object per source, not per target (TASK-260811-2qfnai)

- ROOT CAUSE: the shared publication contract admits no unobservable produced path. `closuregraph.PublicationEvidence.ValidateForPublication` requires `execution.WriteSet` to equal the graph-derived set of paths produced by selected actions **and** requires observations to cover exactly the C5 `DeclaredOutputNodeIDs`, which contain `output_artifact` nodes only; `closureexec.ProtectedStore.Publish` then requires the staged tree and the sorted observation paths to equal that same write set. Net effect: every path a selected action produces must be a declared `output_artifact` with a real observation.
- REGRESSION: the accepted interop closure declared one `generated_artifact` object **per target** (`internal/swiftpminterop/graph.go`), which can never carry an observation. Every SwiftPM closure was therefore unpublishable — invisible until a stage actually tried to publish, because tkurtl never builds.
- FINDING: verified against real Apple Swift 6.3.2 that a per-target single object does not exist. SwiftPM's native build system emits one object per **source file**: a Clang target mirrors the nested source path under `<Target>.build` (`CLib.build/sub/a.c.o`), a Swift target flattens to the source base name (`App.build/extra.swift.o`) — even in release with whole-module optimization.
- DECISION: fixed at source rather than locally. Interop now declares one `output_artifact` per source (`OutputRole: "intermediate"`) at `.curator/objects/<pkg>/<target>/<package-relative-source>.o` with a per-source write slot; `swiftpmbuild` resolves each declared slot to the exact produced file and fails closed with `artifact_local_output_unreceipted` when it is absent or ambiguous. Collapsing N objects into one artifact was rejected: it would have required synthesizing bytes at a slot declaring `native.object`.
- FINDING: an action's declared write slot must be bound exactly once, so production cannot be attached to an accepted action from a later stage. The declaration has to live where the action lives — that is why the fix belongs in interop and not in the build adapter.
- FINDING: SwiftPM's scratch subdirectory is the target triple with the platform version stripped (`arm64-apple-macosx14.0` -> `arm64-apple-macosx`), so a planned output path built from the verbatim triple never resolves.
- FINDING: SwiftPM writes per-target compiler dependency files (`<Target>.build/*.d`) listing the real read set, including SDK `.swiftinterface` modules and cross-package headers. These are the evidence source for the new verified `ReadSetProvider` (`internal/swiftpmbuild/readset.go`), which runs one offline network-denied build and answers per-target read-set requests from them — replacing static reproduction of Clang's search behaviour with what the compiler actually read.
- DECISION: the observer reports `Observed: false` in portable assurance. Portable execution cannot confine reads at the OS boundary, so a compiler-emitted dependency file is corroboration and not proof there; tkurtl's portable reject-by-default verdict is unchanged.
- SCOPE: `internal/swiftpmbuild/` (new package), `internal/swiftpminterop/graph.go`, `types.go`, `internal/swiftpmsource/manager.go`. Full suite + `cmd/curator` + race + `golangci-lint` v2.12.2 all exit 0.

### 2145 — Phase 4 closed by rejection, and a fifth macro shape the review did not name (TASK-260811-tkurtl, round 9 rework)

- DECISION: finding M is closed reject-by-default, not by building a macro expander. `readMacroDefinition` splits `##` on where the fragments come from: a paste joining a **parameter** to another identifier-shaped token rejects (the call site can supply anything), a paste of **fixed** fragments is performed and the joined token stream is then scanned normally. `readAtToken` rejects `@` followed by any identifier outside Objective-C's closed `@`-keyword set, since such an identifier can only be a macro and a macro there expands to `import`.
- FINDING (new, not in the verdict): the `@` itself can come from a macro. `#define AT @` + `AT import SecretKit;` builds the module and reads its header on Apple clang 21.0.0 — the source contains no `@import` and no `@`-before-identifier at all. Closed by the same rule: `@` at end of input, or before a byte that is neither an `@`-keyword nor a literal/collection introducer, rejects.
- NOTE: the GNU `, ## __VA_ARGS__` comma-deletion idiom is deliberately preserved — a punctuator operand contributes no identifier characters, so the paste cannot build a keyword and the argument stays a whole token visible at the call site. That is what keeps the narrowing bounded; `#define JOIN(a, b) a##b` is the part that really does invert, and it was recorded as such.
- DECISION: C++ raw strings are **rejected in this grammar** rather than deferred to `artifactpolicy`. The reviewer was right that the phase-3 parity row was wrong: `R"x(" /* )x"` is load-bearing, not conservative. Whether `R"` opens a raw string is language-mode dependent, which is the trigraph rationale, so the construct rejects rather than being translated under an assumed mode. `H22` asserts the scanner's own verdict through `scanIncludes` so no future change in another component can silently open it.
- NOTE: the terminating argument is that macro expansion cannot *create* an identifier token — replacement lists and arguments are token sequences that already exist in the source, and adjacent tokens never merge. `##` is the single exception. So a channel keyword must either appear literally as a token (caught in content and in `#define` bodies) or be built by `##` (rejected or performed-then-scanned). That is what makes phase 4 closed rather than merely patched.
- SCOPE: `internal/swiftpminterop/headers.go`, `modulemap_test.go`; `IncludeGrammarID` v8 -> v9; new `H21` (14 rejections + 5 positives) and `H22` (6 + 1); focused suite 369 PASS / 0 FAIL.

### 2010 — The layer the pivot forgot: reject-by-default matches source-text tokens, macro expansion rebuilds them (TASK-260811-tkurtl, round 9 review)

- FINDING: the reject-by-default pivot enforces its channel rejections on **literal keyword spellings in the translated text** — `startsAsmStatement`, `startsPragmaOperator`, `startsModuleImport` all prefix-match `asm`/`__asm`/`__asm__`/`_Pragma`/`__pragma`/`import`. Phases 1-3 are now correct, so no *lexical* trick reconstitutes those keywords. **Phase 4 macro expansion still does**, and it runs after the scanner has already admitted the file.
- FINDING: four shapes reach a real file read with **no rejection at all** — not a wrong diagnostic, no error returned. `#define J(a,b) a##b` + `J(a,sm)(".incbin \"payload.bin\"")`; the same with `J(__as,m__)`; `J(_Prag,ma)("clang module import SecretKit")`; and `#define I import` + `@ I SecretKit;`. Verified both ways on Apple clang 21.0.0 (`clang-2100.1.1.101`): the two `.incbin` forms produce an object **byte-size identical to the direct `__asm__` control** with the payload bytes present, and both module forms build `SecretKit` and read its header (`#error SECRET_MODULE_WAS_READ` fires). Curator's `Close()` admits all four.
- NOTE: `scanDirectiveChannels` already catches a macro whose *body literally contains* the keyword (`#define K asm`, `#define STMT(x) __asm__(x)` — both in `H18`, both rejecting). What it cannot catch is a keyword **assembled from fragments that are not the keyword**: for a function-like macro the `##` operands arrive from the **call site**, which is ordinary content, so no body-local analysis is sound. And `startsModuleImport` requires the identifier after `@` to be literally `import`, so `@` falls to `default: s.index++` and is consumed as one content byte while clang expands and imports.
- DECISION (review): the pivot's structural argument is right and worth keeping — the closed set to work against is the **channel keyword set the scanner already owns**, and the rule is that portable mode must not admit a construct that can deliver one of those keywords into a position the scanner cannot see. Cost to state plainly: closing `##` parameter pasting inverts `#define JOIN(a, b) a##b`, currently an admitted positive (`modulemap_test.go:895`), and parameter pasting is common in real C headers. That narrowing is consistent with the accepted contract but must be deliberate and recorded, exactly as `#embed` was this round.
- NOTE (secondary): C++ raw strings are unmodeled and the phase-3 parity row claiming the divergence "can only add a rejection" is wrong. `const char* s = R"x(" /* )x";` hands the scanner an unmatched quote followed by `/*`, and `skipBlockComment` swallows the rest of the file while the compiler sees no comment — verified, payload lands in the object. Not currently an admission hole only because `artifactpolicy`'s source-text lexer rejects every spelling tried with `artifact_opaque_dependency_forbidden` (while ordinary `R"x(hi)x"` admits). The defense is real but lives in another component and is pinned by no vector here.
- NOTE: everything else in the pivot checked out — phases 1-3 including all six finding-L separator variants and the correct pre-splice trigraph ordering, the allowlist enforcement (`readDirective`'s terminal reject, `classifyPragmaBody`'s default reject, `#embed` ahead of the inclusion branch, `.s`/`.S`/`.asm` in `classifyTarget`), the assembler-classifier removal as a removal, the positive path at 341 PASS / 0 FAIL with 53 canonical goldens, and seam/scope hygiene.
- SCOPE: verdict `TASK-260811-tkurtl_review-verdict_RUN-260824-ed3a24.md`; probes attached as `..._review-probe-macro-reconstitution_RUN-260824-ed3a24.log` and `..._review-probe-clang-evidence_RUN-260824-ed3a24.log`.
- STATUS: routed to `to-dev`.


### 1930 — When faithful emulation is the wrong proof: portable mode flips to reject-by-default (TASK-260811-tkurtl, round 9)

- DECISION (user, this session): stop reproducing clang's front end to prove header closure in portable mode. Rounds 2-8 each closed one layer and left the one beneath it — directive spellings, then stages, then the assembler's expansion layer, then the lexical decoding feeding all of them, then (finding L) translation phase 2 dissolving every token-level keyword the previous three rounds had added. The class never shrank, because making the static scanner the *entire* proof requires byte-perfect reproduction of a compiler plus an assembler, which does not terminate.
- FIX: `internal/swiftpminterop/headers.go` — the channel axis is now an **allowlist**. Portable mode positively admits literal `#include`/`#import`/`#include_next`, `@import`, `#pragma clang module import` (plus `_Pragma`/`__pragma` forms), a closed pragma head allowlist, and normal module maps. Everything else rejects: any `asm`/`__asm`/`__asm__` at a token boundary rejects the *target* (`swiftpm_target_platform_unsupported`), `#embed` rejects in every operand form, any unclassifiable `#`-line rejects, any pragma outside the allowlist rejects. `IncludeGrammarID` → `c-family-include-scanner-v8`.
- FIX: same file — `spliceTranslationLines` now removes `\` + any run of horizontal white space + `\n`, closing finding L. Phases 1-3 are the one place where reproducing the compiler is both required and bounded: a reject-by-default keyword match is only closed if a splice, trigraph, or comment cannot reconstitute the token past the scanner.
- DECISION: the assembler classifier is **deleted**, not kept as belt-and-braces. `assemblerChannelDirectives`, `classifyAssembly`, `asmTemplateText`, `decodeStringLiteral`, `decodeDelimitedEscape` and their 62-case suite were reachable only after the keyword match that is now itself the rejection, so retaining them would have added zero assurance while implying a parity claim this round abandons. Deleting `decodeStringLiteral` retires the whole round-5/round-8 lexical-parity surface: no admitted decision now depends on reproducing a C escape sequence.
- NOTE: the closure argument is structural — "portable mode admits allowlist X and rejects all else, therefore an unknown channel cannot be admitted." A channel found after this is written fails closed with **no new emulation**. That is the property seven rounds of per-spelling fixes could not buy.
- NOTE: acceptance is now narrower than the language. Ordinary `__asm__("nop")`, an `__asm__` symbol label, `#embed`, and OpenMP/vendor pragmas all reject with a clear diagnostic before any process starts. Deliberate, recorded in the task scope, and reversed only by the observed-read provider in `TASK-260811-2qfnai` where the compiler itself reports what it opened.
- NOTE: assertion count went 362 → 341. The drop is the deleted classifier suite, which pinned code that no longer exists; 64 new cases replace it. Counting a test of deleted code as coverage would be the dishonest option.
- SCOPE: `internal/swiftpminterop` only; `closuregraph`, `swiftpmsource`, and the canonical goldens untouched. Outcome `TASK-260811-tkurtl_rework-outcome_RUN-260824-f1733b.md`.


### 0645 — The layer beneath every closed layer: escape decoding is where the scanner claims parity with the compiler (TASK-260811-tkurtl, round 7 review)

- FINDING: `decodeStringLiteral` (`internal/swiftpminterop/headers.go`) models the C11/C++11 escape set and falls through to "an undefined escape yields its own character". That is right for `\q`. It is wrong for Clang's **delimited numeric escape sequences** `\x{..}` and `\o{..}`, which Apple clang 21.0.0 accepts in every language mode — verified `gnu17`, `c17`, `gnu++17`, `c++20` — with **zero warnings** even under `-Wall -Wextra`. `\x{2e}` is a `.` to the compiler and the three bytes `x{2e}` to the scanner.
- FINDING: this bypasses both round-7 moves at once. The reconstructed directive name carries no `.` for `classifyAssembly`'s name scan and no `\` for its residual-backslash rejection. Executed: `\x{2e}incbin "/etc/passwd"` exits 0, object 9872 B with 3 hits of `root:`; `\o{56}incbin`, `\x{2e}\x{69}ncbin` (name split across two escapes), `\x{2e}include` (`nm` shows the included symbol), and `\x{2e}linker_option "-lSecretProbeLib"` (exactly 1 `LC_LINKER_OPTION`) all read or emit. Curator accepted all six.
- FINDING: the sharpest form is composite — `\x{2e}macro D a` / `\x{2e}\x{5c}a "payload.bin"` / `\x{2e}endm` / `D incbin` re-enters the *exact* macro layer round 7 closed, with no literal `.macro` and no literal `\` anywhere in the source. Payload in the object, exit 0.
- FINDING: the UCN forms are **not** a mechanism. `\u{2e}` and `\N{FULL STOP}` are both `error: character '.' cannot be specified by a universal character name` — the basic-character restriction blocks every character that could build a directive. Undelimited `\x2e69` is `error: hex escape sequence out of range`. So `\x{}` and `\o{}` are the whole of it.
- NOTE: `\\` is safe in the other direction — it decodes to one `\` and is correctly rejected. The hole is the opposite: an escape the *compiler* decodes into the marker byte that the *scanner* does not decode at all.
- DECISION (review): rounds 2–7 each closed a layer and left the one beneath. Round 5 closed directive spellings and missed a stage; round 6 closed stages and missed the assembler's expansion layer; round 7 closed that layer and missed the lexical decoding feeding all of them. `classifyAssembly` and `assemblerChannelDirectives` are correct — they were handed the wrong bytes. The next closure argument belongs on the **lexical** axis: enumerate every escape form the pinned compiler accepts in a string literal and say, per form, whether the decoder decodes it, rejects it, or is provably identical in passing it through.
- SCOPE: verdict `TASK-260811-tkurtl_review-verdict_RUN-260824-3d58b7.md`; finding **J** accepted as closed; finding **K** is the sole blocker. Probes under `.temp/TASK-260811-tkurtl/probe-r7rev/`.
- STATUS: routed to `to-dev`.


### 0625 — Closing an expansion layer by rejection, not evaluation: the assembler macro channel (TASK-260811-tkurtl, round 7)

- FIX: `internal/swiftpminterop/headers.go` — `classifyAssembly` now rejects any decoded asm template containing a residual `\` before it reads a directive name. C escape decoding has already run at that point, so a surviving backslash *is* the assembler's substitution marker and nothing else. Ordinary inline assembly decodes to no backslash at all (`"nop\n\t"` → newline + tab; extended-asm constraints and `__asm__("_symbol")` labels carry none), so the rule costs nothing on admitted shapes.
- FIX: same file — `macro`, `irp`, `irpc`, `rept`, `altmacro`, `purgem`, `macros_on` added to `assemblerChannelDirectives`, each `swiftpm_header_input_undeclared`. Each opens an expansion layer this stage does not evaluate. `IncludeGrammarID` → `c-family-include-scanner-v6`.
- DECISION: two independent moves rather than one. Either alone closes all five known spellings; both are present so the closure does not rest on a single rule. Pinned by 10 `TestAssemblyTemplateGrammarIsClosed` cases — three bodies with a substitution marker and no expansion directive, seven expansion-directive bodies with no backslash.
- FINDING: re-proved on Apple clang 21.0.0 (`clang-2100.1.1.101`) with a unique marker payload instead of `/etc/passwd`, so the read proof is exact rather than size-inferred. Baseline object 512 B / 0 marker hits; the direct `.incbin` control and all four substitution spellings 536 B / **1** marker hit each; the macro-built `.include` puts `_probe_included_symbol_r7` in `nm`; a missing operand errors `Could not find incbin file` at `<instantiation>:1:9` with `note: while in macro instantiation`.
- FINDING: all seven expansion directives are accepted by the shipped assembler (exit 0, no diagnostic); the control `.zzz_not_a_directive_r7` gives `error: unknown directive`. A clean exit is therefore a positive recognition result, not an absence of evidence.
- FINDING: `.altmacro` `&`-concatenation is **not** a channel here — `.inc&a&bin` under `.altmacro` is `error: unknown directive` at `<instantiation>:1:1`, 0 marker hits. Not modeled as a mechanism; only the `.altmacro` directive itself is rejected, because it selects a macro dialect this grammar does not model.
- NOTE: over-rejected residue is an assembler string literal containing an escaped backslash (`.ascii "a\\b"`), which no admitted SwiftPM shape uses. Conservative direction.
- NOTE: the round-6 stage-axis enumeration was necessary but not sufficient. For each stage you also have to name the substitution/expansion layers that run *before* that stage's lookup and prove each is either evaluated or rejected. The C preprocessor's layer was closed in rounds 2–5 (macro-hidden `_Pragma`, classification at the `#define`, keyword aliases); the assembler's was the last open member. It is now closed by rejection — the same answer the grammar already gives every other layer it declines to model.
- SCOPE: `internal/swiftpminterop/headers.go`, `modulemap_test.go` (13 new `H18` vectors), `parser_test.go` (10 new grammar cases). Package 325 PASS / 0 FAIL, was 302/0.
- STATUS: handed to review round 7.

### 0611 — The integrated assembler substitutes macro parameters *before* directive lookup, so a rejected directive name can be built from inert tokens (TASK-260811-tkurtl, review round 6)

- FINDING: `__asm__(".macro D a\n.\\a \"/etc/passwd\"\n.endm\nD incbin\n");` in a plain `.c` compiles clean on Apple clang 21.0.0 and produces an object **byte-identical** to the direct `.incbin` form — `/etc/passwd` in the object, 9856 B against a 544 B baseline. A missing operand errors with `Could not find incbin file` at `<instantiation>:1:9`, `note: while in macro instantiation`, so the read is proved in both directions.
- FINDING: the same construction works through `.irp x,incbin` + `.\x`, through `.inc\a\()bin` with an empty argument, through `.\a\b` with two arguments, and for `.include` as well as `.incbin`. Five spellings, all accepted by Curator's round-6 assembler grammar with `/etc/passwd` in no include set.
- ROOT CAUSE: `classifyAssembly` scans the decoded template for `.` followed by an identifier. In every one of those spellings the byte after the `.` is `\`, so `splitLeadingIdentifier` returns empty and the template reads as content. Scanning position-agnostically covers a `.macro` body that *contains* the literal spelling; it does not cover one that *constructs* it.
- FINDING (negative, recorded so it is not re-litigated): `.altmacro` exists in this compiler but `&`-concatenation of a directive name is `error: unknown directive`, so `\`-substitution is the only construction mechanism to close. Independent enumeration of the shipped clang binary's assembler strings yields exactly two `Could not find … file` channels — `.incbin` and `.include` — plus `.linker_option` and `.secure_log_unique`/`.secure_log_reset`, corroborating the round-6 directive-name table as complete. `.secure_log_reset` names no file.
- LESSON: the stage axis was the right correction to the directive axis, and it is still one level short. For each stage, ask not only which token spellings name a file but **which substitution or expansion layers run before that stage's lookup**. The C preprocessor's layer was closed across rounds 2–5 (macro-hidden `_Pragma`, keyword aliases classified at the `#define`); the assembler's own layer was not, and it is a different engine with a different substitution syntax.
- SCOPE (open): `internal/swiftpminterop/headers.go` — `classifyAssembly`, `asmTemplateText`, `assemblerChannelDirectives`; `H18` vectors.

### 0550 — `clang -c` runs two file-reading stages; five rounds of directive enumeration covered only one (TASK-260811-tkurtl, rework round 6)
- FINDING: `__asm__(".incbin \"/etc/passwd\"");` at file scope in a plain `.c` exits 0 and puts that file's bytes in the object at offset `0x188`. `.include` does the same for assembler text. Both are **integrated-assembler** directives — a second file-reading stage inside the same `-cc1` — and they share no token with the preprocessor, so the entire grammar closed across rounds 2–5 was bypassed by construction.
- FINDING: `clang -fsyntax-only -H` reports **nothing** for that source. Observed-read verification is not a backstop for this channel either, so in portable mode nothing at all saw it.
- FINDING: `.linker_option "-lNAME"` emits exactly **1** `LC_LINKER_OPTION` load command, making the linker load an undeclared library — where `#pragma comment(lib|linker, …)`, the round-5 candidate, emits **0**. `.secure_log_unique` appends to the path in `AS_SECURE_LOG_FILE`. Same stage, non-read channels.
- FINDING: a relative `.incbin` operand resolves against the **assembler process working directory**, not the including file's directory (verified both ways). There is no declared closure root to confine it against, which is why the operands reject rather than resolve.
- DECISION: `.s`/`.S` sources make a C-family target `swiftpm_target_platform_unsupported`. The old behaviour was unsound in both directions — `.incbin` in a `.S` was invisible, while a lowercase `.s` is not preprocessed at all so an ordinary `# comment` line in it would have false-rejected. Admission still hashes the bytes; the verdict is target-level.
- LESSON: enumerate compiler channels on the **stage** axis, not the directive axis. Rounds 2–5 each added members to one stage's list and each felt like closure; the axis itself was the gap. The closure argument now names every stage `clang -c` runs — driver (inert, re-verified), preprocessor, parse/sema/codegen, integrated assembler, linker (not run under `-c`) — and states each one's file-reading channels.
- LESSON (method, fourth round running): after the first fix, a self-check against the real compiler found four *more* live channels — `#define K asm` + `K(".incbin …")`, the same via `__asm__`, `#define STMT(x) __asm__(x)`, and `__asm__(`⏎`#include "tpl.inc"`⏎`);` — all of which read. Bare `asm` without a template is now rejected under the same mode-dependence rule the trigraph decision uses: it is a keyword in the GNU modes SwiftPM selects and an identifier in strict ISO C, and this stage cannot bind the mode per file.
- SCOPE: `internal/swiftpminterop/headers.go`, `internal/swiftpminterop/language.go`, `internal/swiftpminterop/modulemap_test.go` (`H18`), `internal/swiftpminterop/parser_test.go`.

### 0550 — The clang driver is case-sensitive on exactly `.C` and `.M`, and lowercasing them bypassed both C++ interop gates (TASK-260811-tkurtl, rework round 6)
- FINDING: `clang -### -c up.C` selects `-x c++` and `clang -### -c up.M` selects `-x objective-c++`, while `.c`/`.m` select C and Objective-C. Both `swiftpminterop.sourceLanguage` and `swiftpmsource.targetLanguages` lowercased the extension first.
- ROOT CAUSE: a provider implemented as `impl.C` reported `[c]`, so `implementationCxx` stayed false, the restricted C++ standard/toolchain profile gate never ran, the direct-C++ boundary was never bound, and the recorded `languages` capture evidence was wrong. No case-sensitive filesystem needed — a target simply containing `impl.C` was enough.
- FIX: case-sensitive mapping where the compiler is case-sensitive; every other extension stays case-insensitive, which is what the driver does with them. Admission is deliberately unchanged — the driver compiles both, so the bytes are target source either way.
- LESSON: normalizing an identifier before comparing it is a silent behaviour change whenever the thing being modelled does not normalize. The compiler was the spec here, and `-###` answers it in one command.
- SCOPE: `internal/swiftpminterop/language.go`, `internal/swiftpmsource/graph.go`, `internal/swiftpminterop/language_test.go` (`S10`), `internal/swiftpmsource/swiftpmsource_test.go`.

### 0930 — Closing the channel axis surfaced four more compiler-verified reads beyond the three the reviewer named (TASK-260811-tkurtl, rework round 5)
- FINDING (`_Pragma` hidden in a macro): `#define IMP _Pragma("clang module import SecretKit")` followed by a bare `IMP` imports and reads the module, and so does `#define DO(x) _Pragma(#x)` invoked as `DO(clang module import SecretKit)`. The expansion site is ordinary content no grammar short of a preprocessor can recognize, so the **definition** is the only place the channel can be classified. `scanDirectiveChannels` (`internal/swiftpminterop/headers.go`) now re-scans every classified directive body for the `@import`/`_Pragma`/`__pragma` tokens.
- FINDING (encoding-prefixed literal): `_Pragma(u8"clang module import SecretKit")` imports. The grammar rejects prefixed and raw literals rather than reading them — the probe proves that refusal is necessary, not cosmetic.
- FINDING (`__pragma`): the Microsoft spelling `__pragma(clang module import SecretKit)` is a syntax error in default mode but imports under `-fms-extensions`. Its operand is raw tokens, not a string, so it needs balanced-paren reading rather than the `_Pragma` literal path.
- FINDING (`#pragma include_alias`): silently accepted and **unwarned** in default mode, and under `-fms-extensions` it really substitutes the aliased file — proved by aliasing a missing header onto an `#error` marker. A directive that redirects `#include` resolution is a channel even when the current flags make it inert.
- FINDING (negative, recorded so it is not re-litigated): `#pragma comment(lib, "SecretLib")` is provably inert on this Darwin clang — compiling it emits **0** `LC_LINKER_OPTION` load commands and the name appears nowhere in the object. `#pragma clang include_instead("…")` reads nothing (`-H` lists only the including header). Both stay benign with the evidence attached.
- DECISION: `#pragma GCC dependency` is rejected despite this clang not implementing it. The spelling names a file a conforming implementation opens; no admitted SwiftPM shape uses it, so the rejection costs nothing and the grammar stays closed against a toolchain that does implement it.
- LESSON: The reviewer's axis ("which channels open files") was right and still under-counted by four. Enumerating an axis is not the same as walking it — each member has to be taken to the compiler individually, including the ones that look like obvious no-ops.
- SCOPE: `internal/swiftpminterop/headers.go`, `internal/swiftpminterop/modulemap_test.go` (`H17`).

### 0630 — The read-set class is defined by "does the directive open a file", not by directive familiarity (TASK-260811-tkurtl, review round 4)
- ROOT CAUSE: `classifiableDirectives` (`internal/swiftpminterop/headers.go:33`) is the closed set of non-inclusion directives the scanner drops as benign. It was derived as "directives that are not `#include`" — and two of its nineteen members open files. `embed` is the C23 resource-inclusion directive; `pragma` covers Clang's `module import` and `module build`.
- FINDING (`#embed`): `#embed </etc/passwd>` and `#embed "../../../../etc/passwd"` read the real file in the **default** GNU C mode the pinned Apple Clang 21.0.0 selects when a SwiftPM target declares no `cLanguageStandard` — only a `-Wc23-extensions` warning, exit 0. Proved by content, not by absence of error: `_Static_assert(sizeof(d) == 1, ...)` fails against the real 9344-byte file, and `-H` lists the path. This needs no `-fmodules` and no ISO `-std`, so it is strictly more reachable than the trigraph hole of round 3. Errors only under `-std=c++17`/Objective-C++, which narrows nothing — C and Objective-C are what a SwiftPM Clang target compiles.
- FINDING (`#pragma clang module import`): the `#pragma` and `_Pragma("clang module import …")` spellings both import and read the module's headers. In a plain `.c` file `@import Secret;` — the spelling the scanner *does* cover — does **not** import (`error: expected identifier or '('`), so the pragma is the only module-import channel in C, and it was the invisible one.
- FINDING (`#pragma clang module build`/`endbuild`): declares a module map **inline inside a C source**. It can name an absolute out-of-package header and Clang reads it, through a channel the on-disk module-map confinement stage never parses — the same escape `H03` rejects, one layer down.
- LESSON: After three rounds of hardening the *translation phases* (splicing, comments, trigraphs, Unicode white space) and the *seed set* (transitive worklist, module-map resolutions), the remaining holes were in neither. Enumerate the compiler's file-opening operations and check each against the grammar; do not enumerate the grammar and ask whether it looks complete.
- LESSON (method, third round running): every one of the three was taken to the real compiler before the implementation. `#pragma GCC dependency` looked like a fourth and is not — this clang does not implement it, `-H` shows no read. `__has_embed` is an existence oracle, not a content read. Theory alone would have reported both.

### 0530 — A compiler-faithful scanner still needs phase 1, non-ASCII white space, and the module-map header set (TASK-260811-tkurtl, review round 3)
- ROOT CAUSE: `spliceTranslationLines` (`internal/swiftpminterop/headers.go`) performs line-ending normalization and phase-2 splicing but not phase-1 **trigraph replacement**. `??=` is the trigraph for `#`, and trigraphs were removed from *C++*17 only — they are still C. Verified on this host that `clang -H` reads the header from `??=include "secret.h"` under the default std, `gnu11`, `gnu17`, and `c17`. Because the file contains no `#` byte, the new closed-grammar backstop ("any residual `#`-introduced line is a rejection") never engages. Near miss the other way: `#inc??/`⏎`lude` rejects only by accident, as the unknown directive `inc`.
- ROOT CAUSE: `directiveScanner.run` keeps `atLineStart` across ASCII horizontal space only; every other byte falls to `default:` and clears the flag. A UTF-8 BOM (`EF BB BF`) and U+00A0 both clear it, and clang reads the header in all of: BOM at file start, BOM mid-file, NBSP prefix, and the same forms compiled as Objective-C++. A BOM is the ordinary output of any UTF-8-with-BOM editor, not an exotic spelling.
- ROOT CAUSE: The include worklist seeds from declared non-Swift sources plus the public-header inventory. A custom module map may name `header "../hidden.h"` — outside the public-header root, inside the package — which `confineModuleMapReferences` admits and nothing ever opens. Verified with `clang -fmodules -fmodule-map-file=... -H` that building that module does read the file and follow its includes.
- LESSON: "Run the translation the compiler runs" is a phase list, not a vibe — phase 1 counts, and white space is not the ASCII subset. And a fixpoint is only as complete as its seed set: the transitive closure of *include* references still misses every file admitted by a *different* admission path (here, the module map).
- LESSON (review method): each of the three was found by taking a spelling to the real compiler *first* and only then to the implementation. Two of the three would have read as paranoid theory without `clang -H` confirming the read.

### 0412 — Clang's directive recognition needs both halves of the comment rule (TASK-260811-tkurtl, round 3)
- FINDING: Two Clang behaviors around block comments look contradictory and are both real, verified on this host with `clang -std=c17 -fsyntax-only -H`. (a) `#include /*`⏎`*/ "secret.h"` reads the header — a comment inside a directive does not end it, because phase 3 replaces the comment with one space before directives execute. (b) `int a;`⏎`/*`⏎`*/ #include "secret.h"` *also* reads it — a line-spanning comment restores start-of-line for the token that follows, so the `#` is a directive introducer.
- LESSON: A translate-then-split-on-newline scanner can satisfy only one of these. Collapsing a comment to a space gets (a) and opens a hole on (b); emitting its newlines gets (b) and falsely rejects (a). The fix is a stateful scan: track `atLineStart` across comments and literals, and read a directive body that skips block comments without ending at their newlines.
- DECISION: Classification is closed both ways. Inclusion operands must be exact literals, non-inclusion directives must be in an explicit 19-name set, and any residual `#`-introduced line is `swiftpm_header_input_undeclared`. A `# 1 "file"` line marker is rejected rather than parsed: package sources are not preprocessor output.
- DECISION: Swift sources are excluded from the C grammar entirely (Swift performs no textual inclusion, and `#if os(macOS)`/`#Preview` would otherwise reject), but *every* non-Swift file a directive reaches is scanned regardless of suffix — otherwise `#include "payload.txt"` hides directives behind an extension.

### 0411 — A neutral gate beats a `Selected` gate when the field lives in capture (TASK-260811-tkurtl, round 3)
- DECISION: The round-2 verdict offered two fixes for the conditional `.interoperabilityMode(.Cxx)` false rejection: make the declaration-level gate condition-neutral, or gate it on `boundary.Selected`. Only the first is safe. `boundary.Mode`, `ABI`, `Runtime`, `InterfaceContract`, and `CallingConvention` are all inside `InteropBoundaryPayload`, i.e. the *capture* node — so a `Selected`-gated classification would have labelled the same declared edge `cxx_interop` on Darwin and `c_abi` on Linux, trading a false rejection for a CGP05 neutrality violation.
- FIX: `cxxInteropDeclared` (condition-neutral) drives the boundary classification; `cxxInteropSelected` stays the destination verdict and now feeds only the destination-specific evidence digest. Both are recorded on `TargetInterop`.
- LESSON: When choosing where to put a selection gate, follow the field into the record it lands in. A gate on a capture-bearing value must read a condition-neutral input; only values that live in the binding or evidence overlay may be gated on `Selected`.

### 0310 — Include-directive recognition is a fail-open the operand fix did not reach (TASK-260811-tkurtl, review round 2)
- ROOT CAUSE: `literalIncludeOperand` (`internal/swiftpminterop/headers.go:246`) correctly made the *operand* fail closed, but `directivePattern` still decides whether a line is a directive at all — and a line it does not match is invisible, not rejected. Four spellings the pinned Apple Clang actually reads slip through: a backslash-newline-spliced keyword (`#inc\`⏎`lude <...>`), a comment-prefixed directive (`/* */ #include <...>`), a form-feed-prefixed directive, and the C99/C++ digraph `%:include <...>`. All four verified against the real `clang -std=c17 -H` on this host, not inferred: splicing is translation phase 2 and comment removal is phase 3, both *before* directive recognition.
- ROOT CAUSE: The scan set is `interop.Sources ++ headerPaths(interop.Headers)` (`interop.go:379`) and `interop.Headers` covers the public-header root only. A conventional private header beside the sources (`Sources/CLib/private.h`) is admitted as a resolved *reference* and then never opened, so every directive it declares — including an escaping one — is invisible. This needs no exotic spelling and is the ordinary C layout.
- FINDING: Both are the same shape as the round-1 finding they were meant to close. In `not-observed` mode (the only one reachable before TASK-260811-2qfnai) the declared static closure is the entire header proof, so a scanner that cannot *see* a directive is exactly as fail-open as one that drops its operand.
- LESSON: A "fail closed on unresolvable input" fix is only as strong as the recognizer that decides what counts as input. Match the translation phases the target compiler actually performs, then reject any residual line the grammar cannot classify — and scan the transitive closure, not just the roots.

### 0309 — Neutral boundary derivation over a marker-evaluated flag causes a false rejection (TASK-260811-tkurtl, review round 2)
- ROOT CAUSE: The finding-1 fix made `deriveBoundaries` walk declared edges condition-neutrally, but left the declaration-level C++ gate (`boundaries.go:85`) testing `consumer.CxxInteropMode`, which `cxxInteropSelected` (`platform.go:146`) resolves against destination markers. The rework's premise — "a declaration defect, not a destination fact" — does not hold for a *conditional* `.interoperabilityMode(.Cxx)`.
- FAILURE: A package declaring both its C++ dependency and its opt-in under `.when(platforms: [.macOS])` closes on Darwin and hard-rejects on Linux with `closure_interop_undeclared`, even though the entire C++ declaration is pruned there. Before the rework the selected-only walk never classified the provider on Linux, so this is a regression introduced by the fix.
- LESSON: When you make a derivation selection-neutral, audit every input it reads. A neutral walk over a destination-evaluated flag is not neutral; it just moves the leak.


### 0235 — A neutral interop boundary needs a conditional consumer side (TASK-260811-tkurtl)
- ROOT CAUSE: `selectionReachability` (`internal/closuregraph/projection.go:170`) reaches an `interop_boundary` *forward* from the consumer action and traverses `provides_interop` in *reverse* to the provider. With an unconditional `consumes_interop`, a boundary and its provider stayed reachable on every destination, so `ProjectActive` could never record a pruned verdict — emitting the boundary neutrally was not enough on its own.
- FIX: Optional `Condition` on `ConsumesInteropPayload` (`internal/closuregraph/edge.go`), mirroring `RequiresPayload`. `requires` was previously the only condition-bearing edge kind, and `requires` cannot point at an `interop_boundary` (`validation.go:1165`), so no in-package workaround existed.
- FINDING: The field is absent-by-default in the canonical map, so unconditional consumer sides canonicalize byte for byte as before; the 53 labeled goldens and the Ruby verifier still pass unchanged.
- FINDING: Binding edges are unconditionally members of `selectedEdges` and both endpoints must be selected (`validation.go:445`). A pruned capture node therefore must get *no* binding overlay — `uses_tool`/`targets`/boundary `requires` are now emitted only for selected targets and boundaries.
- DECISION: Destination-profile gates (C++ interop profile, Objective-C runtime, restricted-language profile, unsafe settings) apply only to the destination-selected subset; declaration-level gates (missing `.interoperabilityMode(.Cxx)` against a C++ public interface, mixed Swift/C-family target) stay unconditional. Otherwise a Linux close would reject a Darwin-only Objective-C target that it merely prunes.
- SCOPE: `internal/closuregraph/edge.go`, `internal/swiftpminterop/{interop,boundaries,graph,headers}.go`, `internal/swiftpmsource/{types,executor_runtime,manifest}.go`.

### 0234 — `publicHeadersPath` must be carried and bound, not defaulted (TASK-260811-tkurtl)
- ROOT CAUSE: `swiftpmsource.Target` dropped SwiftPM's `publicHeadersPath`, so `publicHeaderRoot(target.Path, "")` hardcoded `include`. A package with a custom public-header directory got a *generated* module map for the wrong directory while the real escaping map was never parsed — H03 defeated by a layout the model could not represent.
- FIX: `Target.PublicHeadersPath`, decoded from `dump-package`, and bound into `manifestDigest` under `public_headers_path`. `targetDeclarationDigest` already folds in `manifest_digest`, so binding it once in the manifest digest also binds it into every target-unit node identity — a silent public-header relocation is now manifest-replay drift.
- DECISION: A layout this profile cannot represent exactly fails closed (`swiftpm_target_platform_unsupported`) rather than falling back to the default; and any `module.modulemap` in the admitted target tree outside the resolved public-header root is `swiftpm_modulemap_escape`. Substituting a default silently inspects the wrong directory.
- NOTE: An inclusion directive whose operand is not an exact literal (`#include SOME_MACRO`) is now `swiftpm_header_input_undeclared`. In the only mode reachable before TASK-260811-2qfnai (`not-observed`), the static scan *is* the header proof, so a permissive scanner is a fail-open.
- STATUS: 62 tests (114 with subtests), 86.0% coverage. All gates exit 0. Each new vector proved non-vacuous by reverting its fix and observing the failure. Handed off to review; not committed.

### 0144 — SwiftPM C-family interop validation delivered (TASK-260811-tkurtl)
- SCOPE: New `internal/swiftpminterop` (10 production files, 7 test files); `internal/swiftpmsource/manager.go` gained 3 additive accessors only.
- DECISION: Interop is a separate package consuming `*swiftpmsource.Capture` and republishing a *new* capture graph + binding. The accepted source-closure graph digest is never mutated, so upstream evidence stays valid.
- FINDING: `closuregraph` allows only `targets`, `uses_tool`, toolchain-scoped `requires`, and `provides_interop` in a binding table (`internal/closuregraph/validation.go:1146`). SDK/system header reads therefore cannot be `reads` edges; they are attested by toolchain-scoped `requires` from the boundary plus the resolution record in `ReadSetEvidence`.
- FINDING: `validateInteropBoundaries` requires a *distinct* provider-before-consumer action pair, so every selected C-family/Swift target needs a compile `action` node with exactly-once-bound tool/read/write slots. Without it the boundary rejects with `closure_interop_undeclared`.
- DECISION: Interop boundary payloads carry selection-neutral contract IDs (`c-abi-v1`, `itanium-cxx-abi-v1`, `cxx-standard-library-v1`, `objc-runtime-v1`). Putting the destination ABI/runtime in the capture node broke CGP05 selection neutrality across Darwin/Linux.
- DECISION: Provider containing Objective-C/Objective-C++ binds `objc_runtime` regardless of which symbols the consumer calls — strictly more conservative than inferring a C-only edge. C++-only public header extension + Swift consumer without `.interoperabilityMode(.Cxx)` = `closure_interop_undeclared`.
- ROOT CAUSE: Early false-negative on H04/H05 — `roots.resolve` returned `selected_binding` for any path *spelled* under an SDK root, even a nonexistent one. FIX: `presentNode` existence check in `internal/swiftpminterop/containment.go`.
- FINDING: Shared closure services raise stable codes as the leading token of a plain `fmt.Errorf`, not as a typed `*swiftpmsource.Failure`. `swiftpminterop.ErrorCode` resolves both encodings against a closed set.
- STATUS: 58 tests (96 with subtests), 86.1% coverage. Focused + race + golangci-lint v2.12.2 (package and repo-wide) + repository suite minus `cmd/curator` + bounded `cmd/curator` subset + canonical golden verifier all exit 0. Handed off to review; not committed.
- NOTE: `go list -test -deps ./cmd/curator` contains no `swiftpm*` package, so the monolithic suite was not required to cover this delta.

## 2026-08-24 — SwiftPM interop review (TASK-260811-tkurtl, RUN-260823-fee71e)

Independent acceptance review of `internal/swiftpminterop` returned **changes
requested**. Three defects confirmed by executed probes, not inference:

1. **Republished interop capture is not selection-neutral.** `classifyTargets`
   walks only `capture.TargetNodeIDs` and `directTargetDependencies` filters
   conditional edges by destination markers, so a `.when(platforms:)` dependency
   makes the interop capture digest destination-dependent (`b1316468` Darwin vs
   `55c2a21a` Linux) while the upstream `swiftpmsource` capture correctly stays
   identical (`3560a162` both). The CGP05 vector cannot see it because its
   fixture has no conditional edge — a reminder that a selection-neutrality
   golden is only as strong as the conditionality in its fixture.

2. **The C-family include scanner fails open where the module-map parser fails
   closed.** `includePattern` only matches literal operands, so
   `#include SOME_MACRO` produces no reference and no diagnostic. This matters
   *today* specifically because no `ReadSetProvider` exists until
   TASK-260811-2qfnai, so `not-observed` is the only reachable mode and the
   static scan is the entire header-closure proof.

3. **A non-default `publicHeadersPath` is silently defaulted to `include/`.**
   `swiftpmsource.Target` carries no such field, so the declaration is dropped
   before the interop stage, and a `module.modulemap` in the real public-header
   directory is never parsed — the H03 absolute-header escape survives in a
   layout the model cannot represent. An unmodeled shape should fail closed, not
   get a substituted default.

General lesson: for a fail-closed profile, "the model cannot represent this
declaration" must itself be a rejection. Findings 2 and 3 are both instances of
the same shape — a construct the grammar/contract cannot resolve exactly is
quietly dropped instead of raising a diagnostic.

## 2026-08-24 — Trigraph replacement is mode-dependent; a review table said otherwise (TASK-260811-tkurtl, round 4)

Round-3 review reported that Apple clang 21.0.0 replaces `??=` with `#` in the
default and `-std=gnu17` modes as well as `-std=c17`, and asked for
unconditional phase-1 trigraph replacement in the C-family include scanner.
Re-probing with an exact match on the `-H` output line — instead of a grep that
also matched the echoed source line — shows the opposite for GNU modes:

| Mode | `??=include "x.h"` |
| --- | --- |
| default, `-std=gnu17` | ignored (`-Wtrigraphs: trigraph ignored`) |
| `-std=c89/c99/c11/c17`, `-std=c++14` | replaced, header is read |
| `-std=c++17`, Objective-C++ default | ignored |

Unconditional replacement would have opened the mirror-image hole: under the
GNU default, `int a;??/`⏎`#include "x.h"` **does** read the header, and
translating `??/` to a backslash splices that directive away. Both readings are
real, the mode cannot be bound per file (one target may compile C and
Objective-C under one declared standard and share headers with C++ translation
units under another), so a source containing any trigraph now rejects.

Two general lessons worth keeping:

1. **A compiler probe needs a control and an exact match.** `grep secret.h` over
   `clang -H` output matches the echoed source line of the failing directive, so
   every negative reads as a positive. The reliable forms are an anchored match
   on the `-H` dependency line, or making the target header an `#error MARKER`
   so a read is unambiguous. A verdict's compiler table is evidence, not fact —
   it can be wrong the same way.
2. **"Translate faithfully" is only safe when the translation is unconditional.**
   Where a translation phase is mode-dependent and the mode is not bound, the
   closed-grammar move is to reject the construct, not to pick a mode.

The same round's self-check found three further holes of the class the reviewer
was hunting, each verified with `-fmodules` against a module whose only header
is an `#error` marker: `@ import M;`, `@/*c*/import M;`, and `@`⏎`import M;` all
import, and a C++14 digit separator (`int x = 1'0; @import M;`) hid the import
from a scanner that consumed an unterminated quote to end of line. Module import
recognition is token-level now, not byte-adjacency.

## 2026-08-24 — Clang's integrated assembler is a second file-reading stage

Found while reviewing `TASK-260811-tkurtl` round 5. Five rounds of review
enumerated the C preprocessor's file-reading channels (`#include`, `#embed`,
`#pragma clang module import`, `_Pragma`, `__pragma`, trigraphs, BOM/white
space, module-map seeding) and each round found more members of the same class.
The reason the class kept reopening is that every round enumerated **directives**
when the right axis is **compiler stages**.

`clang -c` runs the integrated assembler after the preprocessor, and its grammar
has two directives that open arbitrary files: `.include` and `.incbin`. Both are
reachable from source SwiftPM admits, with no flags, in the default mode:

- `__asm__(".incbin \"/etc/passwd\"");` at file scope in a plain `.c` file.
  Verified on Apple clang 21.0.0: exit 0, the bytes land in the object at
  `0x188`, a missing file is a hard error, and `clang -fsyntax-only -H` reports
  no read — so it is invisible to header-read verification too.
- `.s` / `.S` sources, which `swiftpmsource.swiftPMSourceExtension` admits
  (`internal/swiftpmsource/executor_runtime.go:513`) and the interop stage then
  scans with the C preprocessor grammar, which models no assembler directive.

Curator accepted all four vectors with `/etc/passwd` in no include set.

Generalization worth keeping: when a closure proof is "the compiler must not read
anything we did not declare", the enumeration has to be per-stage, and each stage
needs an explicit disposition — including the stages that provably read nothing.
A longer list within one stage never terminates.

Second, unrelated defect found in the same review: `sourceLanguage` lowercases
the extension, so `.C` classifies as C and `.M` as Objective-C, while clang's
driver maps them to C++ and Objective-C++ (`clang -### -c up.C` → `-x c++`).
That makes `implementationCxx` false for a C++ provider, so the
`.interoperabilityMode(.Cxx)` gate never fires.

## Round 8 — finding K: delimited numeric escapes desynchronized the scanner from the compiler

The layer beneath every assembler/preprocessor channel closed in rounds 2–7 was
the *lexical* one. `decodeStringLiteral` in `internal/swiftpminterop/headers.go`
modelled the C11/C++11 escape set and fell through on anything else. Apple clang
21.0.0 also accepts Clang's **delimited numeric escapes** `\x{…}` and `\o{…}` as
an extension in *every* language mode with no diagnostic — verified at `-std=`
c89/c99/c11/c17/gnu17/c23, c++98/gnu++17/c++20/c++23, and `-x objective-c` /
`objective-c++`, all producing the same bytes.

So `\x{2e}` is one `.` to the compiler and four characters to the scanner. Six
`__asm__` templates were accepted by Curator while reading a marker payload into
the object, including a composite that re-entered the round-7 macro layer through
`\x{5c}` → `\` with no literal `.macro` and no literal `\` in the source.

FIX: decode `\x{…}`/`\o{…}` (new `decodeDelimitedEscape`, range-checked after
every digit so a long body cannot wrap into an in-range byte), reject malformed
delimited content, reject every universal-character-name form (`\uXXXX`,
`\UXXXXXXXX`, `\u{…}`, `\N{…}`) and `\o` without a brace. `decodeStringLiteral`
now returns `(string, bool)`; `asmTemplateText` returns `(text, reason, ok)`.
`IncludeGrammarID` → `c-family-include-scanner-v7`.

Generalization worth keeping, and it is the same shape as the round-6 one a level
down: a scanner that claims parity with a compiler owes an explicit disposition
for *every branch of the grammar it reproduces*, not just the ones in the
standard. A `\` in a C string literal is followed by exactly one of eight things
(simple escape, octal digit, `x`, `o`, `u`, `U`, `N`, undefined) — enumerate the
branches, not the spellings, and the axis terminates. "Undefined escapes yield
their own character" was true of the standard and false of the implementation.

---

## 2026-08-24 — TASK-260811-tkurtl round 8 review: a whitespace line splice dissolves the token-level channel keywords (finding L)

Round 8 closed the escape row of the lexical axis (finding K) correctly, and the
parity table re-derives cleanly against the pinned Apple clang 21.0.0. The
blocker sits one row *above* it, in translation phase 2.

`spliceTranslationLines` removes only the exact byte pair `\` `\n`, with an
explicit doc comment claiming that `\` + white space + `\n` "fails closed
instead of being resolved on a guess". The pinned compiler *does* splice that
form — warning `-Wbackslash-newline-escape` and then performing the splice — with
space, tab, vertical tab, form feed, mixed runs, and `\r`, in all twelve verified
language modes. Unlike trigraphs, it is not mode-dependent.

That splice reconstitutes a *token*. The fail-closed claim holds only for
channels recognized by line position: `#inc\ ⏎lude </etc/passwd>` leaves the
residual `#inc`, an unclassified directive that rejects, and a split inside an
asm string literal leaves an unterminated literal that rejects. It does not hold
for the three channels recognized by token prefix at an arbitrary column —
`__asm__`/`__asm`/`asm`, `_Pragma`/`__pragma`, and `@import`. Splitting the
keyword leaves no residual at all; the fragments are ordinary content and the
statement is never entered.

Compiler-proven, Curator-accepted:

- `__as\ ⏎m__(".incbin \"payload.bin\"");` — 528 B object, marker hit 1,
  byte-size identical to the direct `.incbin` control (baseline 504 B); the
  missing-file variant errors `Could not find incbin file`.
- `__as\ ⏎m__(".include \"inc_src.s\"");` — `nm` shows the included symbol.
- `_Pra\ ⏎gma("clang module import SecretKit")` and `@imp\ ⏎ort SecretKit;` —
  the module header's `#error` marker fires under `-fmodules`.

FIX direction: splice `\` + a `horizontalSpace` run + `\n` in
`spliceTranslationLines`, exactly as the compiler does. Nothing to bind per
file, so rejection would cost a real source shape for no security gain.

Generalization: the round-7/8 lesson ("enumerate the branches of the grammar you
reproduce, not the spellings") applies one level up too. A scanner that
reproduces translation phases owes a disposition for every *phase*, and the
phase that decides whether a recognizer runs at all is more dangerous than the
phases inside it. A fail-closed argument that reasons about "the residual" is
only valid for recognizers that leave one — position-anchored ones do,
prefix-anchored ones do not.

---

## 2026-08-24 — phase-4 residuals: an allowlist is a delivery vehicle, and `##` has a second spelling

TASK-260811-tkurtl, reviewer RUN-260824-d10094, Apple clang 21.0.0
(`clang-2100.1.1.101`), `arm64-apple-darwin25.5.0`.

The phase-4 fix (reject a `##` paste whose fragment comes from a call site;
reject `@` + an identifier outside Objective-C's closed `@`-keyword set) closes
the four probes that motivated it and leaks the class three more ways. All three
were reproduced end-to-end through `Close()` with **no rejection at all**.

1. **An allowlist of identifiers is an allowlist of macro names.**
   `objcAtKeywords` exists so `@` + an unknown identifier can fail closed. But
   `protocol`, `class`, `selector`, `end`, `YES` are ordinary identifiers to the
   preprocessor: `#define protocol import` + `@ protocol SecretKit;` builds and
   reads an undeclared module. The allowlist that makes the rule fail closed is
   the way through it. Also reachable as `#define class im##port` — the paste
   layer collapses it correctly, then the `@` layer waves the result through.
   Each half is right; the composition leaks.

2. **`@import NAME;` macro-expands NAME; `#pragma clang module import NAME` does
   not.** Verified both directions. The scanner records and gates the
   *pre-expansion* spelling, so `#define CLib SecretKit` + `@import CLib;`
   satisfies `moduleDeclared` on a module the compiler never resolves, and the
   recorded closure evidence names the wrong module. This is not a
   channel-keyword bug at all — it is evidence integrity — which is why a proof
   framed entirely around keywords could not reach it.

3. **`%:%:` is `##`.** `collapseMacroPastes` short-circuits on
   `strings.Contains(body, "##")`. `#define A __as%:%:m__` and
   `#define J(a,b) a%:%:b` embed arbitrary file bytes (object byte-identical to
   the direct `__asm__` `.incbin` control); `#define A _Prag%:%:ma` reads an
   undeclared module. `readDirective` already handles `%:` as a directive
   introducer, so the grammar knew about digraphs — the paste layer didn't.

Two premises of the cross-layer argument were checked and **hold**, and are
worth keeping: macro output is not re-scanned for directives (`#define INC
#include "x"` → `expected identifier`, no read), and adjacent expanded tokens do
not merge (`#define A __as` + `#define B m__`, `A B(...)` → `unknown type name
'__as'`, no payload). So the argument's method is sound.

Generalization, one level up from the round-9 entry: a fail-closed proof over a
*keyword set* must be re-derived over **preprocessing tokens including their
alternative spellings**, and separately over **every identifier position the
scanner records or gates on**, asking of each whether the compiler expands it.
The first re-derivation catches `%:%:`. The second catches the `@import` module
name, which no amount of keyword reasoning would have surfaced. And any
identifier allowlist inside a reject-by-default grammar needs the question
"can a macro be named this?" asked of it explicitly.

## 2026-08-24 — phase-4 residuals closed: where a macro-expansion rule can be *decided* (TASK-260811-tkurtl, round 11 rework)

- DECISION (N1): the `@`-keyword allowlist is closed by rejecting the **`#define`**,
  not the `@`. Both directions were available; only one is decidable where it is
  written. "Is this identifier macro-defined?" asked at the `@` needs the whole
  translation unit — and the realistic vector binds the macro in a header and
  uses it in a `.c` file. "Does this definition bind a name the compiler expands
  after `@`?" is answerable from the definition alone, and is answered wherever
  the definition sits, because every admitted file of the closure is scanned.
  General shape: when a rule needs cross-file knowledge at the use site, look for
  the equivalent rule at the *definition* site — it is usually local.
- FINDING (not in the verdict): `__pragma(clang module import NAME)` **is**
  macro-expanded under `-fms-extensions`. The verified asymmetry is not
  "`@import` expands, pragmas do not" — it is `@import` and `__pragma` on one
  side, `#pragma` and `_Pragma` on the other. Encoding the coarser statement
  would have left the Microsoft operator open.
- FINDING (not in the verdict): `#define import protocol` + `@import SecretKit;`
  compiles and imports **nothing**, while the scanner records a module import.
  Same evidence-integrity class as N2 reached from the opposite end, so the two
  import spellings joined the rejected `#define` name set.
- NOTE: the N1 narrowing is real — `#define interface struct` (Windows COM) and a
  package-local `#define true 1` C89 shim now reject. Recorded rather than
  worked around; neither is needed by an admitted SwiftPM C-family shape.
- METHOD: every new vector was run against the code with the three fixes
  *disabled in place* before being trusted. 17 of 22 subtests fail without them;
  the one that passes either way (`#define A @im%:%:port`) passes incidentally
  via a different rule, which is exactly what the reviewer predicted. A rejection
  vector that never observed the hole is not a regression guard.

## 2026-08-24 — the same channel, one input down: a `-D` value is a macro body (TASK-260811-tkurtl, round 13, finding Q)

- CONTEXT: round 12 (finding P) closed the SwiftPM `.define` build setting's
  macro NAME into both oracles and enumerated the build-setting KIND axis. The
  reviewer accepted both and then found the level below: the `define`
  disposition read `splitLeadingIdentifier(value)` and threw away everything
  after `=`.
- FINDING: a build-setting define BODY is a channel, exactly as a source
  `#define` body is (round 9 finding M). Re-derived on the pinned Apple Clang
  21.0.0: `clang -c -D'A=__asm__' d.c` with `A(".incbin \"payload.bin\"");`
  produces an object **byte-identical** to the direct `__asm__(…)` control
  (sha256 `70777455…`) with the named file's bytes inside it, and the same source
  with no `-D` fails `expected parameter declarator` — the setting is the entire
  vector. `-D'A=_Pragma'` with `A("clang module import SecretKit")` reads a module
  the target never declared.
- ROOT CAUSE (the reusable one): this is the *third* time the same defect class
  landed one input down. M closed the source `#define` body. N1/N2 closed the
  identifier positions. P closed the build-setting NAME. Q closed the
  build-setting BODY. Each time the analysis was correct and the new INPUT was
  simply not wired to it. The pattern: when a stage grows a second way to supply
  an input the compiler already honored one way, the risk is never the grammar —
  it is that the second route reimplements or skips the first route's analysis.
- FIX SHAPE: factor, do not fork. `analyzeMacroBody` and `readMacroParameters`
  came out of `readMacroDefinition`; both routes now call them, so
  `readMacroDefinition` has no body logic of its own left to drift from. The
  tests assert every vector TWICE — once bound by the setting, once by a source
  `#define` — and require the SAME diagnostic code. Equal codes are the proof
  that it is one analyzer; a pair of "both reject" assertions would not be.
- DECISION: a define body that resolves to a Clang module import is REJECTED,
  not confined. A build setting is not an admitted source file — the reference
  belongs to no scanned unit and has no directory to resolve against — so under
  reject-by-default it is refused rather than attributed to something it did not
  come from. Deliberate narrowing, recorded, not worked around.
- TERMINATING ARGUMENT: the macro-INPUT surface for an admitted C-family target
  is exactly {source `#define`, build-setting `define`}. A macro reaches the
  pinned `clang -c` from a preprocessing directive or from `-D`, and `-D` is
  reachable only through the `define` kind (the accepted kind axis proves no
  other kind spells `-D`, and portable mode passes no response file, `@file`,
  `-include`, or environment define). Name and body of both are now routed
  through the same reject logic, so the macro layer is closed across input
  surfaces, not just spellings and positions.
- METHOD (again worth it): the fix was stubbed out at the call site and the
  suite rerun before the vectors were trusted. All six `bound by the build
  setting` subtests plus the module-import and operand-separation cases fail
  without it; all six source controls, the pruned case, and the benign positive
  pass either way — which is what they are for.
