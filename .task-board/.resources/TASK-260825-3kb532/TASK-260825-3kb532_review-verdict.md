# TASK-260825-3kb532 — Review Verdict: ACCEPTED

Reviewer run: RUN-260825-18497d. Scope reviewed: the install precheck and
candidate prompt for HTTPS build repositories, the shared scope/persistence
question, the SSH prompt fix, and their wiring — as delivered in the Story
worktree `.temp/STORY-260825-32bopo/worktree` (uncommitted delta over 903af23).
Sibling-task bytes (gitcred, config build-https CLI, broker/fetch wiring) were
already reviewer-accepted on their own tasks and were read here only as
context.

## Acceptance criteria — all satisfied

1. **Prompt default, this-run-only, and abort covered by tests.**
   - Default: `TestHTTPSPromptDefaultSelectsExistingCredentialAndNarrowestSavedScope`
     — Enter/Enter selects the existing host credential and persists the
     narrowest (repository-namespace) scope. `defaultBuildSSHScope` supplies
     that namespace as the scope-question default on both surfaces.
   - This-run-only: `TestHTTPSPromptThisRunOnlyNeverReachesPersistenceOnEitherCredentialSurface`
     (existing-credential and entered-token subtests) and
     `TestPromptThisRunOnlyReturnsSSHCredentialWithoutPersisting` (SSH). Both
     assert the run-local material IS returned while the persist callback is
     never called — the negative is paired with positive twins through the
     same harness, so it cannot pass vacuously.
   - Abort: `TestHTTPSPromptAbortNeverPersistsOrSelects` covers `q` at every
     question plus end-of-input; `TestPromptedBuildHTTPSAbortStopsTheProductionPlanBeforeAnyFetch`
     drives the production entry point `planExternalBuilds` with a fetch
     counter and proves the abort lands before the first acquisition.

2. **A this-run-only answer never lands in the saved config.**
   `TestHTTPSThisRunOnlyPromptNeverReachesConfigOrCredentialStore` uses the
   real production callback `persistPromptedBuildHTTPS(cfg, access)`, a real
   isolated Git credential store, and byte-for-byte before/after comparison of
   both the config file and the credential-store file, on both HTTPS
   surfaces. `TestSSHThisRunOnlyPromptNeverReachesTheSavedConfig` does the
   byte comparison for SSH through the real `config.SetBuildSSH` writer.

3. **SSH latent persistence bug checked and fixed.** The bug was real: the old
   `InteractiveBuildSSHResolver` called `persist(credential)` unconditionally
   once the scope question was answered — there was no run-only choice and
   every accepted answer reached the config. Fixed by the shared
   `readCredentialScope` (save flag; `r` = this run only; persist gated on
   save). Structural check: `config.SetBuildSSH`/`SetBuildHTTPS` are called
   only from the two per-choice persist callbacks and the explicit CLI
   commands — no accumulator-wide save exists; `resolveBuildSSH` matches
   against an explicit copy of the configured scope map, so a prompted answer
   cannot mutate the loaded configuration behind the run.

4. **go test green.** Orchestrator's full-suite evidence is attached
   (`TASK-260825-3kb532_orchestrator-full-suite.log`, 42 packages ok, exit 0).
   Reran myself in the Story worktree: `go build ./...` ok; `gofmt -l cmd
   internal` empty; `go vet` clean on the five affected packages;
   `golangci-lint run` on them — 0 issues; `go test -count=1
   ./internal/gitcred ./internal/config ./internal/buildrepo
   ./internal/install` all ok; targeted `./cmd/curator -run
   'TestConfigBuildHTTPS|…ThisRunOnly…|…PromptIsWired…|…PromptPersists…'` ok.
   The `-v` run confirms every run-only subtest executes (no skips).

## Task-description checks

- Resolution runs before the first fetch and covers every declared HTTPS
  repository; precedence is captured override → longest configured scope →
  prompt → anonymous, with provenance lines for the dry run.
- Candidates are exactly the two required: the operator's existing credential
  for the host (presence-only discovery, username shown, secret never read
  until selected — pinned by
  `TestBuildHTTPSDiscoveryOnlyListsAndNeverSelectsAHostCredential`, which
  asserts the only reader call is `discover:<host>`), and entering a token now.
- Nothing persists without the explicit scope answer; the narrowest offered
  default is the repository namespace; a chosen scope must actually cover the
  repository identity.
- Off a terminal (and on dry run) the resolver stays nil and unmatched HTTPS
  repositories continue anonymously
  (`TestTheCredentialPromptIsWiredOnlyWhereAnOperatorCanAnswerIt`, plus the
  nil-resolver anonymous branch in the resolution tests).
- The prompt mirrors the SSH shape: same ask/say helpers, same abort token,
  the same shared scope question; docs/build-ssh.md transcript updated to the
  new question wording, README credential summary updated.

## Non-blocking observations (no rework required)

1. `persistPromptedBuildHTTPS`'s save path — keyring `StoreScoped` →
   `SetBuildHTTPS` with rollback `DeleteScoped` when the config write fails —
   has no direct test. Both building blocks are independently tested against
   real stores (login round-trip test; write_test), and the run-only test
   proves the callback is never reached without a save. The rollback branch
   itself is the only untested line of glue.
2. The new "scope must cover <identity>" re-ask loop in `readCredentialScope`
   has no direct test (the malformed-scope re-ask is tested). It is a
   hardening beyond the AC.
3. A this-run-only answer always applies at the default namespace scope (`r`
   is one token, no custom run-only scope). Consistent with the documented
   design.

## Verdict

Accepted → `done`. Implementation matches the AC, fits the existing
build_ssh/credential architecture (shared scope grammar, longest-scope match
extracted to one place, fail-closed admission), and the persistence boundary
is proven by negative tests that would fail if the gate admitted what it must
reject. Commit/integration into trunk remains the orchestrator's step; no
commit was made by this reviewer run.
