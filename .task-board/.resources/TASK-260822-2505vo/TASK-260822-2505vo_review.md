# TASK-260822-2505vo — reviewer verdict: ACCEPTED

Reviewer run `RUN-260822-13d94b` (not goal-bound). Read-only; no code modified.

## What was reviewed

Branch `task/TASK-260822-2505vo-per-repo-credentials` in
`.temp/TASK-260822-2505vo/worktree`, cut from `origin/main` at `6a9b201`.
14 files, +1937/-9.

The board patch artifact was verified to be the branch, not a description of it:

```
diff <(git apply --stat <worktree diff vs 6a9b201>) \
     <(git apply --stat TASK-260822-2505vo_final.patch)   -> identical
```

The branch applies the accepted `TASK-260822-96m5pj` deliverable verbatim: the
worktree's `internal/config/buildssh.go` differs from
`TASK-260822-96m5pj_build-ssh-config.patch` only by this task's own addition of
`MatchBuildSSH`. Nothing from the accepted config task was dropped. (The extra
`SetBuildSSH` / `RemoveBuildSSH` / `ValidBuildSSHPath` symbols sitting
uncommitted in the primary checkout are *not* part of the 96m5pj artifact —
they are later, untracked work for `TASK-260822-3pkc80`, and their absence here
is correct.)

## AC, item by item

| AC clause | verdict | evidence |
|---|---|---|
| flags/env beat config | met | `TestRunWideSelectionCoversEveryRepositoryAheadOfConfiguredScopes`, `TestCaptureBuildSSHSelectionPrefersFlagsOverTheEnvironment`, `TestInstallFlagsCarryTheRunWideBuildSSHSelection` |
| config scope selected per repository | met | `TestConfiguredScopesAreSelectedPerRepository` — longest-scope wins, and `portals-evil` proves the segment boundary is not a string prefix |
| https / local substitution skip | met | `TestRepositoriesThatNeedNoCredentialSkipSelection`; `needsBuildSSH` keys on the **effective** state, so `TestSubstitutionMovesARepositoryOffAndOntoTheSSHTransport` covers both directions |
| empty ssh selection fails closed with the protocol code | met | `TestUnselectedSSHRepositoriesFailClosedWithTheProtocolCode`; backstopped a second time at the admission boundary by `TestAcquireNetworkRefusesAnSSHRepositoryWithoutAnOperatorSelection` |
| pinned-agent, agent-only, identity-only all reach SSHPolicy | met | `TestEveryConfiguredShapeReachesTheWrapperPolicy` asserts the three `ExactSSHCommand` tails end-to-end from a configured scope, and runs `config.ValidateBuildSSH` on each scope first so no case proves a shape an operator may not write |
| `~` expansion for config paths | met | `TestConfiguredPathsExpandTheOperatorHome`, `TestBuildSSHExpandedResolvesHome` |
| go test green | met | see below |

## Independent verification (run by the reviewer on the final tree)

| check | result |
|---|---|
| `gofmt -l cmd internal` | 0, no output |
| `go vet ./internal/... ./cmd/...` | 0 |
| `golangci-lint run` (v2.12.2, CI's pin) | `0 issues.` |
| `go test ./internal/install ./internal/buildrepo ./internal/config -count=1 -run 'BuildSSH\|SSH\|Credential\|Shape\|KnownHosts\|Scope'` | ok, ok, ok |
| `bash .github/ci/gate-selftest.sh` | `75 passed, 0 failed` |
| attached `TASK-260822-2505vo_full-suite-final.log` | 41 `ok`, 0 `FAIL`/`panic` |

## Fit with the architecture

- Credentials are bound to `GitTool` **per repository** inside
  `externalPipelineRequest`, not per run. A closure spanning two hosts cannot
  offer either host the other's key. This is the right seam: the tool value is
  copied, so `deps.GitTool` stays credential-free.
- `resolveBuildSSH` runs once for the whole closure before the first repository
  is reached, so a run holding one unselected private repository names *every*
  one of them instead of dying part way through the network.
- Matching is on `effective.Identity` — the host actually contacted — and
  `effectiveRepository` sets exactly the transport the acquire closure later
  hands to `AcquireNetwork`. The skip predicate and the fetch cannot disagree.
- Package data never reaches the selection (Spec §12.2). Confirmed by reading
  the inputs of `resolveBuildSSH`: canonical identity and operator-owned
  flags/env/config only.
- `AcquireNetwork` refusing an unselected SSH source is a genuine second gate,
  not duplication: it catches a caller that forgot to resolve.
- No production regression from the new refusal. `productionGitTool()` never
  sets `SSHWrapper`, and `AcquireNetwork` already refused SSH without the exact
  manager wrapper, so external SSH repositories were not reachable before this
  change either.
- The two new `.github/ci/skip-classes.tsv` rows are narrow and correct. The
  other two skip reasons the new helpers introduce ("this host cannot create a
  short scratch directory", "... an agent socket") already match the
  pre-existing `host-capability  this host cannot create` class — the skips were
  phrased to land there rather than widening the table. Gate self-test agrees.

## Non-blocking observations (for the follow-up tasks, not rework here)

1. **Known-hosts precedence is asymmetric with identity/agent.** Run-wide
   flags/env beat config for identity and agent, but a scope's `known_hosts`
   beats a run-wide `--build-ssh-known-hosts` (`BuildSSHSelection.knownHosts`
   tries `named` first). Defensible — host keys are a fact about the
   destination, not about who authenticates — and it is documented and tested.
   Worth one sentence in the docs task `TASK-260822-4p3dcq` so the asymmetry is
   deliberate in the operator's mind too.
2. **`~` is not expanded for the env spelling.** A shell expands
   `--build-ssh-identity ~/.ssh/id`, but `CURATOR_BUILD_SSH_IDENTITY=~/.ssh/id`
   reaches validation verbatim and fails with "SSH identity path must be
   absolute". The AC only asks for config expansion, so this is in scope as
   written; it is a sharp edge for the docs task.
3. **`install.Options.BuildSSH` is write-only inside package `install`.** Flag
   parsing fills it and `cmd/curator` reads it back to build the selection;
   nothing in `install` consumes it. A library caller that sets it and calls
   `install.Project` gets a silent no-op. Harmless today, a footgun later.
4. **`operatorKnownHosts` uses `Lstat` + `IsRegular`**, so a symlinked
   `~/.ssh/known_hosts` is not adopted as the default, while
   `ValidateOperatorSSHCredentials` does resolve symlinks. The inconsistency
   fails in the safe direction (no default -> fail closed rather than trusting
   an unverified host).

None of these change behaviour the AC specifies, and none justify another
producer cycle.

## Verdict

**Accepted.** The work is uncommitted on a task worktree branch; per the
reviewer contract this run supplies no `commit_ack`. The commit-owning mover
should land `TASK-260822-2505vo_final.patch` (or the branch) and make the final
transition itself.
