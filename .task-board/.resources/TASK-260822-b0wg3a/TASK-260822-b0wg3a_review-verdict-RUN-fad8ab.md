# TASK-260822-b0wg3a — reviewer verdict: ACCEPTED

Run `RUN-260822-fad8ab`, read-only, not goal-bound.
Reviewed tree: `.temp/TASK-260822-b0wg3a/worktree`, branch
`task/TASK-260822-b0wg3a-install-precheck`, staged-but-uncommitted on
`origin/main` **6a9b201**.

## Provenance re-verified at first hand

The attached `TASK-260822-b0wg3a_final.patch` re-diffs **identical** to the
working tree (`git diff --cached origin/main`, 4485 lines, compared modulo
`index` lines). The evidence is faithful rather than asserted.

Rebuilt the predecessor baseline independently — `git archive origin/main` +
`TASK-260822-2505vo_final.patch` (applies clean) — and diffed the whole tree
against it. Every file the implementer claimed byte-identical is:

| File | vs accepted 2505vo baseline |
| --- | --- |
| `internal/install/external.go` | identical |
| `internal/buildrepo/credentials.go` | identical |
| `internal/buildrepo/admission.go` | identical |
| `internal/buildrepo/credentials_test.go` | identical |
| `internal/install/buildssh_test.go` | identical |
| `internal/config/config.go` | identical |
| `.github/ci/skip-classes.tsv` | identical |

The true delta charged to this task is exactly: new
`internal/install/buildsshcandidates.go`, `buildsshprompt.go` and three test
files; edits to `internal/install/buildssh.go`, `install.go`, `global.go`;
resolver wiring in `cmd/curator/main.go` + tests; `go.mod` promoting
`charmbracelet/x/term` from indirect to direct. `internal/config/buildssh.go`,
`write.go`, `write_test.go` and the rest of `cmd/curator/main.go` are
3pkc80's, already accepted.

## Acceptance criteria

| Clause | Verdict | Evidence re-run here |
| --- | --- | --- |
| Prompt flow — default selection | met | `TestPromptDefaultSelectionPinsTheAgentToTheFirstDiscoveredKey` |
| Prompt flow — abort | met | `TestPromptAbortFailsClosedWithoutPersistingAnything` (5 scripted abort shapes incl. EOF at each question), `TestAnAbortedResolverFailsTheRunWithItsOwnError` |
| Prompt flow — manual path | met | `TestPromptManualIdentityPathCoversAKeyDiscoveryDidNotList`, `TestPromptRejectsAMalformedManualPathAndAsksAgain` |
| Fail-closed message asserts candidate commands | met | `TestUnselectedRepositoriesFailClosedWithCommandsBuiltFromTheCandidates`, `TestAHostWithoutCandidatesSaysSoRatherThanInventingAPath` |
| Dry-run reports per-repository credential source | met | `TestADryRunReportsThePerRepositoryCredentialSource` (and asserts an install carries **no** report), `TestTheCredentialReportIsLabelledByTheScopeThatProducedIt` |
| Precheck before any fetch | met | `TestTheCredentialPrecheckRunsBeforeAnyFetch` asserts 0 `Acquire` calls and that **both** uncovered repositories are named at once |
| Persist only after an explicit scope choice | met | `persist` is unreachable before both questions answer (`buildsshprompt.go:63-73`); `TestPromptedScopesDoNotMutateTheLoadedConfiguration`; `TestThePromptPersistsThroughTheOrdinaryConfigWriter` |
| go test green | met | see gates |

Design invariants I checked by reading rather than by trusting the notes:

- **Discovery never selects.** `discoverBuildSSHCandidates` returns a list; the
  only writes are `persist(...)` from the prompt and the local `scopes` copy.
  `resolveBuildSSH` copies `selection.Scopes` before mutating, so a prompted
  answer cannot alter the loaded configuration behind the run — verified by
  test *and* by reading the copy at `buildssh.go:151-154`.
- **A nil resolver is the fail-closed path.** `operatorBuildSSHResolver`
  returns nil on a dry run and whenever either stdin or stderr is not a real
  terminal (`term.IsTerminal`, not a character-device test). `< /dev/null` and
  a pipe are both covered by `TestTheCredentialPromptIsWiredOnlyWhereAnOperatorCanAnswerIt`.
- **A resolver that answers the wrong question still fails closed**
  (`TestAResolverThatCoversNothingStillFailsClosed`): the second match pass is
  the ordinary longest-scope rule, not a trust of whatever the resolver
  returned.
- **Credentials are bound per repository at the fetch boundary.**
  `externalPipelineRequest` copies `deps.GitTool` per row before setting
  `SSHCredentials`, so no repository inherits a sibling's selection. (2505vo's
  code, unchanged; re-checked because this task is what populates the map.)
- **No new read-only regression.** `curator status`/`doctor` build plans with
  `productionExternalDeps(cfg, true)` and no resolver, so an uncovered SSH
  repository fails those surfaces — but it already did under 2505vo via
  `SSHPolicyFor`; this only moves the refusal earlier and improves the message.
- `ssh-add` is invoked with a resolved path, a fixed argument, a 5s context
  timeout and a deliberately minimal environment; exit 1 is correctly read as
  "an agent holding nothing" rather than as unreachable.
- The `plan.messages → Result.Messages` extraction into
  `externalPlan.credentialReport(label)` is behaviour-preserving — diffed both
  call sites against the baseline; output is byte-identical, now under test.

## Gates re-run at first hand, in the worktree

| Command | Result |
| --- | --- |
| `go build ./...` | exit 0 |
| `go vet ./...` | exit 0 |
| `gofmt -l .` | no output |
| `golangci-lint run ./...` (v2.12.2) | `0 issues.`, exit 0 |
| `.github/ci/gate-selftest.sh` | `75 passed, 0 failed` |
| `go test -count=1 -timeout 30m ./...` | **exit 0**, 41 packages `ok`, `cmd/curator` 486.238s |
| `go test -race` on the 36 named precheck/prompt/candidate tests | exit 0, race-clean |

Statement coverage on the new code (`internal/install`): `buildsshprompt.go`
88.9–100% per function, `buildsshcandidates.go` 78.9–100%, `buildssh.go`
83.3–100%. The only sub-90% functions are the two that shell out to a live
`ssh-add` — unavoidable without a real agent, and both degrade paths are
covered.

Confirmed the implementer's `-timeout` note independently: the log stays
0 bytes for ~8 minutes because `go test ./...` emits in package order and
`cmd/curator` is first. Not a hang.

## Findings — none blocking, all recorded for the follow-ups

**1. A prompted credential is written to disk but not to the in-memory config,
so `install --all` re-asks per target.** `operatorBuildSSHResolver`'s persist
closure calls `config.SetBuildSSH(cfg.Path, ...)`, which re-reads and rewrites
the file; `cfg.BuildSSH` is untouched. `cmdInstallMode` captures the selection
once (`main.go:531-533`) and loops over targets, so target 2's
`resolveBuildSSH` copies the same stale scope map and prompts again for a
repository the operator already covered. Answering twice is harmless — the
same scope is replaced — but it defeats, across targets, exactly the
"one scope covers every sibling" property the prompt enforces within a target.
Cheap fix when this is next touched: have the persist closure also write the
credential into `cfg.BuildSSH` (allocating the map when nil). Low severity.

**2. Identity-only candidates built from `*.pub` cannot authenticate.** The
implementer flagged this in the notes and deliberately deferred it; I confirmed
it against `internal/buildrepo/admission.go:150-168`. The agent-less shape
emits `IdentitiesOnly=yes IdentityAgent=none -i <path>`; with a public-key file
and no agent there is nothing to sign with. So menu entries 3..N, and the
`--identity ~/.ssh/x.pub` lines in the fail-closed diagnostic, are selectable
choices that fail at fetch time. `ValidateOperatorSSHCredentials` proves only
that the path is a regular file. The **default** entry (agent pinned to the
first key) is the sound one, and listing `*.pub` is what the task description
prescribes, so this is not a defect in the precheck — but a "ready-to-run"
command that cannot work is worth a sentence in `TASK-260822-4p3dcq`, and
narrowing the identity-only entry to private keys is a product call about what
discovery may read from `~/.ssh`.

**3. `defaultBuildSSHScope` can produce a scope the scope grammar rejects.**
Canonical repository paths are validated by `identifiers.PortableComponent`,
which admits any non-control rune except `:` `/` `\` — `+`, `@`, `~`, non-ASCII.
`config.ValidBuildSSHScope` restricts path segments to `^[A-Za-z0-9._-]+$`.
For an identity such as `git.example.com/team+infra/app` the derived default
scope is `git.example.com/team+infra`, which the CLI refuses. Two consequences:
the non-interactive diagnostic prints `curator config build-ssh add
git.example.com/team+infra …`, a command that errors when pasted; and
interactively, `readBuildSSHScope` offers that default and then rejects it, so
pressing Enter re-asks with the same rejected suggestion until the operator
guesses a host-only scope or aborts. Recoverable (the rule text is printed) and
half of it predates this task — 2505vo's `missingBuildSSHError` already derived
the command scope the same way — but the prompt half is new here. Worth a
narrowing pass: fall back to the host segment when the namespace scope does not
validate. Low severity; unlikely on GitHub/GitLab-shaped namespaces.

**4. Nit — `fmt.Sscanf(answer, "%d", &index)` accepts trailing garbage**
(`buildsshprompt.go:186`). Verified: `"1abc"`, `"2 x"`, `"1.5"` all parse to a
selection. That contradicts the function's own stated intent that an
unparsable answer is re-asked. In practice it is forgiving rather than
dangerous (`"1)"` does what the operator meant); `strconv.Atoi` would be exact.

## Handoff

Nothing is committed. The reviewed tree is the staged index of
`.temp/TASK-260822-b0wg3a/worktree`; `TASK-260822-b0wg3a_final.patch` on the
board is its faithful serialization. The commit-owning mover lands that patch.
