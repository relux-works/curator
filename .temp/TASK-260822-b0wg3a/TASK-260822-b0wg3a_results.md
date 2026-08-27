# TASK-260822-b0wg3a — install-precheck-and-candidates

Handed to review.

## Where the work lives

Worktree `.temp/TASK-260822-b0wg3a/worktree`, branch
`task/TASK-260822-b0wg3a-install-precheck`, based off `origin/main` **6a9b201**.

The checkpoint branch `handoff/cocoaskills-parity-20260731` has no
`internal/install/external.go` and no `internal/buildrepo`, so the named target
files only exist on main — the same base every task in this story used.

The branch carries the three predecessor patches and then adds the precheck:

| Source | Files | State on this branch |
| --- | --- | --- |
| TASK-260822-96m5pj | `internal/config/buildssh.go`, `config.go` | verbatim (plus the `ValidBuildSSHPath` export 3pkc80 added) |
| TASK-260822-2505vo | `internal/install/buildssh.go`, `external.go`, `internal/buildrepo/credentials.go`, `admission.go`, `.github/ci/skip-classes.tsv` | verbatim; `buildssh.go` extended, the rest byte-identical to the accepted patch |
| TASK-260822-3pkc80 | `cmd/curator/main.go`, `internal/config/write.go` | verbatim |
| **this task** | `internal/install/buildsshcandidates.go`, `buildsshprompt.go` + tests, additions to `buildssh.go`, `global.go`, `install.go` | new |

Verified by rebuilding 6a9b201, applying `TASK-260822-2505vo_final.patch`
(applies clean), and diffing every shared file: `external.go`,
`credentials.go`, `admission.go`, `buildssh_test.go`, and `skip-classes.tsv`
show no delta.

## What was built

### Discovery (`internal/install/buildsshcandidates.go`)

Lists the operator's own material and nothing else:

- the live `SSH_AUTH_SOCK`, with a key count from `ssh-add -l`. Every failure
  path — tool absent, stale socket, agent refuses, 5s timeout — degrades to
  "key count unavailable" rather than dropping the agent as a candidate. Exit
  status 1 is a real answer (an agent holding nothing), not a failure to reach.
  The tool is invoked with a resolved path and a deliberately minimal
  environment: only `SSH_AUTH_SOCK`, plus `SYSTEMROOT`/`PATH` so a process can
  start at all on Windows.
- `*.pub` files directly under `~/.ssh`, regular files only, sorted, spelled
  with the leading `~/` the config grammar accepts, capped at 8 with the
  remainder reported as a count rather than dropped silently.

Discovery runs **after** the run-wide selection has already failed to cover a
repository, reads only operator-owned state, and never turns what it finds into
a selection.

### Precheck (`resolveBuildSSH`, `internal/install/buildssh.go`)

`planExternalBuilds` resolves credentials for the whole run before the first
repository is reached. A closure holding one uncovered private repository fails
closed naming **every** uncovered repository at once, rather than part way
through the network.

The scope set the run matches against is a copy of the configured map, so a
prompted answer cannot mutate the loaded configuration behind the run.

### Interactive prompt (`internal/install/buildsshprompt.go`)

Two questions per uncovered repository, in order:

1. **credential** — numbered menu over the discovered material. Entry 1 is the
   default: the agent pinned to the first discovered key, the only shape that
   both reuses a loaded key and stops the agent offering every other key to the
   destination. `m` takes a free-form identity path; `q` aborts.
2. **scope** — defaulting to the repository namespace, so a sibling repository
   of the same group is covered without naming every repository by hand.

`persist` is reached only after both questions are answered. An abort, or end
of input, returns `ErrBuildSSHAborted`, which carries
`build_repository_ssh_credential_missing` so the run fails closed exactly as it
would have without a terminal. A scope the operator just chose suppresses the
question for the sibling repositories it already covers.

The resolver is supplied (`cmd/curator/main.go`) only when both stdin and
stderr are real terminals — `term.IsTerminal`, not a character-device test, so
`< /dev/null` fails closed instead of blocking on a prompt nobody can answer.
A dry run never gets one either: it reports what a run would do, and persisting
a credential mid-report would make a read-only surface write.

### Fail-closed diagnostic (`missingBuildSSHError`)

Built from the same candidates the menu offers: every uncovered repository, the
material detected on this host, and ready-to-run
`curator config build-ssh add` commands — one set per uncovered namespace,
deduplicated. A host with nothing detected says so and emits a
`~/.ssh/<key>.pub` placeholder the operator must replace, rather than inventing
a path.

### Dry-run provenance

One line per repository that needs credentials, naming the repository and where
its selection came from (`operator flags/env`, or `config scope "…"`),
populated only on a dry run and labelled by the scope it belongs to on the way
into `Result.Messages`.

## Change made to the predecessor code

The two identical `plan.messages → Result.Messages` loops in `install.go` and
`global.go` were the one hop in the chain with no test: a dropped loop would
have left the tests green while the operator saw nothing. Extracted as
`externalPlan.credentialReport(label)` and covered by
`TestTheCredentialReportIsLabelledByTheScopeThatProducedIt`.

## Acceptance criteria

| Clause | Evidence |
| --- | --- |
| Prompt flow, default selection | `TestPromptDefaultSelectionPinsTheAgentToTheFirstDiscoveredKey` |
| Prompt flow, abort | `TestPromptAbortFailsClosedWithoutPersistingAnything`, `TestAbortCarriesTheProtocolCodeSoTheRunStillFailsClosed`, `TestAnAbortedResolverFailsTheRunWithItsOwnError` |
| Prompt flow, manual path | `TestPromptManualIdentityPathCoversAKeyDiscoveryDidNotList`, `TestPromptRejectsAMalformedManualPathAndAsksAgain` |
| Fail-closed message asserts candidate commands | `TestUnselectedRepositoriesFailClosedWithCommandsBuiltFromTheCandidates`, `TestAHostWithoutCandidatesSaysSoRatherThanInventingAPath` |
| Dry-run reports per-repository credential source | `TestADryRunReportsThePerRepositoryCredentialSource`, `TestProvenanceNamesEverySourceOnePerRepository`, `TestTheCredentialReportIsLabelledByTheScopeThatProducedIt` |
| Precheck before any fetch | `TestTheCredentialPrecheckRunsBeforeAnyFetch` (asserts zero fetches) |
| Persistence only after an explicit scope choice | `TestOneScopeChoiceCoversEverySiblingRepository`, `TestPromptedScopesDoNotMutateTheLoadedConfiguration`, `TestThePromptPersistsThroughTheOrdinaryConfigWriter` (cmd/curator) |
| go test green | see Gates |

## Gates

All run as standalone processes in the worktree.

| Command | Result |
| --- | --- |
| `go build ./...` | exit 0 |
| `go vet ./...` | exit 0 |
| `gofmt -l .` | no output |
| `golangci-lint run ./...` (v2.12.2) | `0 issues.`, exit 0 |
| `.github/ci/gate-selftest.sh` | `75 passed, 0 failed`, exit 0 |
| 40 named buildssh/precheck/prompt/candidate tests | all PASS, exit 0 |
| `go test -count=1 -timeout 30m ./...` | see `TASK-260822-b0wg3a_full-suite-final.log` |

## Note for whoever runs the suite next

`cmd/curator` alone takes ~10 minutes on this host (`godriver` fingerprints the
whole GOROOT). The predecessor's own log records `588.652s` for it. A run with
`-timeout 5m` panics with `test timed out` inside `godriver.digestToolchainRecords`
and reads as a hang — it is not one. Use `-timeout 30m`.

## Not done here

`internal/install` has no end-to-end fixture that drives `Project()` with a
real external build repository; the authoritative dry-run conformance test is
gated behind `CURATOR_CONFORMANCE_ROOT` and skips locally. The credential
report is therefore asserted at the plan boundary and at the labelling hop, not
through a live `Project()` call. Building that fixture tier was out of scope
for this task.

## Observation for review and for the docs task (TASK-260822-4p3dcq)

Discovery lists `*.pub` files, as the task description prescribes. Those feed
two different menu entries, and only one of them is unconditionally sound:

| Entry | Wrapper flags (`internal/buildrepo/admission.go:150-168`) | A `.pub` path here |
| --- | --- | --- |
| agent, pinned to `<key>.pub` (default) | `IdentitiesOnly=yes IdentityAgent=<socket> -i <path>` | correct — a public key is exactly what pins the agent to one identity |
| identity `<key>.pub` (agent-less) | `IdentitiesOnly=yes IdentityAgent=none -i <path>` | cannot authenticate — no agent, and the file is not a private key |

Nothing rejects the second combination: `ValidateOperatorSSHCredentials` proves
the path resolves to a regular file, not that it is a usable private key. An
operator who picks an agent-less identity from a discovered `.pub` gets an SSH
authentication failure at fetch time rather than a selection-time refusal.

Left as-is deliberately — the default entry is the sound one, discovering
`*.pub` is what the task specifies, and narrowing the identity-only entry to
private keys is a product call about what discovery may read from `~/.ssh`,
not a bug in this precheck. Worth a sentence in the build-ssh docs.
