# TASK-260822-4p3dcq: docs-build-ssh-surface — review

**Verdict: accepted.**

Reviewed independently in the producer's worktree
`.temp/TASK-260822-4p3dcq/worktree`, branch
`task/TASK-260822-4p3dcq-docs-build-ssh`, off `origin/main` **6a9b201**. Delta
charged to this task is two files: `docs/build-ssh.md` (new, 494 lines) and one
`README.md` bullet. Nothing under `cmd/`, `internal/`, `.github/` was touched by
this task; the rest of the branch is the accepted
`96m5pj`+`2505vo`+`3pkc80`+`b0wg3a` chain carried as the staged baseline.

Review was verification, not trust. Every console transcript in the page was
re-executed against a binary I built from the worktree myself
(`go build -o review-curator ./cmd/curator`, exit 0), and every spec citation
was read in `/Users/iv/Developer/ReluxWorks/curator-spec`.

## Acceptance criteria

| AC | Result |
| --- | --- |
| Docs build/lint green | Met by the gates that exist and apply. Verified below. |
| Examples match actual CLI output | Verified by re-execution. Byte-for-byte. |
| Links resolve | 0 misses over `README.md`, `docs/build-ssh.md`, `CONTRIBUTING.md`, `docs/implementation-plan.md`. |

## Independent reproduction of every transcript

Built `review-curator` from the worktree, pointed `CURATOR_CONFIG` at a fresh
scratch config, and replayed the page. Output matched the page exactly,
including exit codes:

| Page transcript | Reproduced | Notes |
| --- | --- | --- |
| four `build-ssh add` lines (`added ...`) | yes | identical text and ordering |
| `replaced build_ssh scope ...` | yes | replace vs add distinction is real |
| the `build_ssh` JSON block | yes | scratch config after the four adds, verbatim, key-for-key |
| `build-ssh list` (four tab-separated lines, sorted) | yes | operator spelling preserved, `~/` not expanded |
| empty listing on stderr, exit 0 | yes | `curator: no build_ssh scopes are configured`, exit 0 |
| `removed build_ssh scope ...` | yes | |
| removing an absent scope | yes | exit 1, path named in the message |
| the two refusal cases | yes | both exit 2 |
| three `-build-ssh-*` help lines | yes | `curator install -h`, identical wording |
| seven fail-closed parse messages | yes | all seven, exact strings, exit 1 |

The seven parse errors were reproduced from seven hand-written malformed
configs of my own, not the producer's:

```
build_ssh: must be an object
build_ssh: scope "Git.Example.COM" must be a lowercase host optionally followed by ...
build_ssh.git.example.com: must be an object
build_ssh.git.example.com: has unsupported field(s): key, zzz
build_ssh.git.example.com.agent: must be true for the operator's own agent socket, or an agent socket path
build_ssh.git.example.com.identity: must be an absolute path or start with '~/' and carry no control character
build_ssh.git.example.com: requires 'agent', 'identity', or both
```

Claims about `add` that I checked separately and that hold:

- `add` validates before the configuration is read: a malformed scope with
  `CURATOR_CONFIG` pointing at a nonexistent file still exits 2 with the scope
  rule, not with `global config not found`.
- optional-valued `--agent`: `add --agent git.example.com/portals2` keeps the
  scope positional; `add --agent /run/agent2.sock git.example.com/x` records the
  socket; `add --agent ./relative git.example.com/y` is a usage error, so the
  next token is claimed only when it reads as a socket path.
- the three flags are present on all four of `install`, `upgrade`,
  `global install`, `global upgrade`, as the page claims.

The prompt transcript, both fail-closed diagnostics and the provenance line
match `TASK-260822-4p3dcq_docsamples.log` verbatim, and I confirmed each against
its generator in source rather than against the log alone:

- diagnostic layout and ordering: `missingBuildSSHError`, `internal/install/buildssh.go:330`
- the suggestion ordering (agent+identity, agent, then each identity) and the
  `~/.ssh/<key>.pub` placeholder when discovery is empty:
  `buildSSHAddCommands`, `internal/install/buildsshcandidates.go:177`
- menu labels, `[default]`, `credential [1-4, m, q] (default 1)`, `m`, `q`:
  `internal/install/buildsshprompt.go:122,128,158,172`
- provenance line format: `buildSSHProvenance`, `internal/install/buildssh.go:236`

## Behavioural claims checked against source

Every load-bearing statement in the page was checked, not sampled:

| Page claim | Source | Holds |
| --- | --- | --- |
| the three authentication tails, exactly as spelled | `internal/buildrepo/admission.go:150,155,168` | yes |
| precedence flags > env > scopes > prompt > fail closed | `CaptureBuildSSHSelection` / `resolveBuildSSH`, `internal/install/buildssh.go:78,147` | yes |
| flags and env merge field by field, flags first | same, the three-field loop | yes |
| bare `--build-ssh-known-hosts` is not a selection | `Selected()`, `internal/buildrepo/credentials.go:35` | yes, ignores `KnownHosts` |
| `--build-ssh-agent auto` fails when `SSH_AUTH_SOCK` is unset | `runWideCredentials`, `internal/install/buildssh.go:270` | yes, with `build_repository_ssh_credential_missing` |
| known-hosts resolution order: scope, run-wide, `~/.ssh/known_hosts` | `knownHosts`, `internal/install/buildssh.go:292` | yes |
| no host key file at all is a refusal | `SSHPolicyFor`, `internal/buildrepo/credentials.go:125` | yes |
| `build_repository_identity_invalid: SSH identity path must be absolute` | `ValidateOperatorSSHCredentials` + `AdmissionError.Error()`, `internal/buildrepo/credentials.go:59`, `admission.go:52` | yes, that is the exact `Error()` string |
| symlink resolved, then must be a regular file / socket | `internal/buildrepo/credentials.go:62-74` | yes |
| longest matching scope wins; non-canonical value matches nothing | `MatchBuildSSH`, `internal/config/buildssh.go:153` | yes |
| segment-aware matching via the same helper as allowlists | `identity.MatchesPrefix` | yes |
| paths capped at 4096 Unicode scalar values | `maxCredentialPathLength` + `utf8.RuneCountInString`, `internal/config/buildssh.go:34,291` | yes |
| faulty scopes reported in sorted order | `sort.Strings(scopes)`, `internal/config/buildssh.go:221` | yes |
| Windows absolute forms recognised without consulting the platform | `windowsAbsoluteRE`, `internal/config/buildssh.go:53` | yes |
| prompt only when stdin and stderr are real terminals and not a dry run | `operatorBuildSSHResolver`, `cmd/curator/main.go:1320` | yes |
| a prompted answer is written through the same writer as `add` | `config.SetBuildSSH` at `main.go:1327` and `main.go:1915` | yes, same function |
| unparsable menu answer re-asks; `1)` and `1abc` select entry 1 | `Sscanf` branch, `buildsshprompt.go:186` | yes |
| Enter on a rejected default scope re-asks with the scope rule | `readBuildSSHScope`, `buildsshprompt.go:216` | yes |
| provenance lines populated only on a dry run | `internal/install/external.go:131` | yes |
| default scope is the repository namespace | `defaultBuildSSHScope`, `internal/install/buildssh.go:382` | yes |
| `install --all` captures the selection once, so a prompted answer is not folded back | `main.go:532` sits before the target loop at `main.go:535` | yes |

## Spec citations

Read at first hand in `curator-spec`. All five are accurate and none cites
another implementation:

| Citation | Verified against | Verdict |
| --- | --- | --- |
| `Spec §12.2` | `protocol/core.md:1271` | faithful; the page's list of forbidden selectors is the spec's own list |
| `Spec §6.3` | `protocol/core.md:665` | faithful; the SSH raw path is restricted to ASCII letters, digits, `.`, `_`, `-`, `/` for `go-repository-v1`, and HTTPS retains the Unicode grammar, exactly as the page contrasts them |
| `Spec §6.1` | `protocol/core.md:628` | faithful; the spec's own example is `h/a/b` matches `h/a` but not `h/a-evil`, which is the page's example |
| `Spec §6.4` | `protocol/core.md:739` | faithful; effective state is what decides |
| `Spec profiles/manager.md §11.3` | `profiles/manager.md:919` | faithful, and the drift note is correct |

**The drift note is correct and I confirmed it.** `profiles/manager.md §11.3` on
`curator-spec` main says the tail "is exactly either" the identity-only or the
agent-only form: two tails. The pinned-agent third tail exists only on branch
`spec/pinned-agent-authentication-tail`, head **38232d3**, where §11.3 lists
three numbered tails. Flagging that in the page rather than presenting three
tails as settled protocol is the right call and is the kind of thing a docs page
usually gets wrong silently.

## Checklist items 4 and 5

**Item 4 (agent-less `--identity` at a `*.pub` file) is documented as written.**
The page names the tail, explains there is nothing to sign with, states that
validation only proves the path resolves to a regular file, identifies the
affected surfaces (the identity-only menu entries and the
`--identity ~/.ssh/<key>.pub` diagnostic lines), and names both sound shapes.
Confirmed against `admission.go:150` and `credentials.go:45`.

**Item 5 is documented with a correction, and the correction is right.** I
re-derived it rather than taking the producer's word:

- `internal/identity.PortableComponent`, which the `b0wg3a` finding cited,
  governs skill sources. External build repositories parse through
  `buildrepo.ParseSource`, which for SSH gates the raw path on
  `sshPathRE` before Git or SSH starts (`buildrepo.go:124`,
  `validRepositoryPath:151`). That alphabet is exactly what `scopeSegmentRE`
  (`^[A-Za-z0-9._-]+$`) admits, so the **path** half of the claimed divergence
  cannot occur on the only transport `build_ssh` selects for.
- The **host** half does diverge and the page documents it:
  `hostRE = ^[A-Za-z0-9][A-Za-z0-9.-]*$` (`buildrepo.go:36`) admits
  `git.example.com.`, `git..example.com`, `git-.example.com`, which
  `scopeHostRE` (`buildssh.go:49`) rejects. Widening the suggestion to the host
  does not rescue it, because the host is not a valid scope either, so the only
  covering path is `--build-ssh-*` or `CURATOR_BUILD_SSH_*`. The page says
  precisely that, including what happens interactively (the scope question keeps
  re-asking, confirmed at `buildsshprompt.go:216`) and non-interactively (the
  printed command errors when pasted).

Correcting a review finding instead of transcribing it, with evidence, is the
better outcome. Accepted as documented.

## Gates I ran myself, in the worktree

| Command | Exit | Result |
| --- | --- | --- |
| `go build ./cmd/curator` | 0 | clean |
| `go build ./...` | 0 | clean |
| `gofmt -l .` | 0 | no output |
| `golangci-lint run ./...` | 0 | `0 issues.` |
| naming gate (`ci.yml` job `naming-gate`, replayed inline) | 0 | zero matches outside `README.md`, exactly 1 line in `README.md` |
| em dash / guillemet rule (`CONTRIBUTING.md:29`, `docs/implementation-plan.md:44`) | pass | `docs/build-ssh.md` has none; the added `README.md` bullet has none |
| link and anchor resolution (own script over four markdown files) | 0 | 0 misses, including all 10 in-page TOC anchors |
| `go test -count=1 -timeout 30m ./internal/config/ ./internal/install/... ./internal/buildrepo/ ./cmd/curator/` | 0 | all five packages `ok` |

Test detail, run by me in the worktree, log
`.temp/TASK-260822-4p3dcq/review-gotest.log`:

```
ok  github.com/relux-works/curator/internal/config             0.440s
ok  github.com/relux-works/curator/internal/install           96.279s
ok  github.com/relux-works/curator/internal/install/atomicity 104.102s
ok  github.com/relux-works/curator/internal/buildrepo         22.630s
ok  github.com/relux-works/curator/cmd/curator               376.452s
EXIT=0
```

The full `go test ./...` was not run for this task and does not need to be: the
delta is two markdown files and touches no Go code, the scoped suite above
covers every package the page describes, and the chain's full suite was green at
`b0wg3a` `RUN-260822-fad8ab`.

**"Docs build/lint green" assessed honestly.** The repository genuinely has no
markdown linter, no link checker and no docs build step: `docs/` holds two files
and no index, `Makefile` has no docs target, and `.github/workflows/ci.yml` has
no docs job. The producer said so plainly instead of claiming a green gate that
does not exist, which is correct behaviour. The AC is satisfied by the gates
that do exist and apply to prose, all of which I re-ran.

## Fit with the project

- A new page under `docs/` plus a `README.md` cross-link is right.
  `docs/implementation-plan.md` is the v0.1 phase plan and mentions no external
  build repositories at all, so there was no existing page to extend; the task
  wording "update the repository's existing docs pages" is satisfied by the
  README bullet, and inventing a home in the phase plan would have been worse.
- Citation style (`` `Spec §6.3` ``) matches the convention already used across
  `docs/` and `README.md`.
- The page documents operator-owned surface only and never presents a package as
  able to influence credential selection, which is the point of `Spec §12.2`.

## Non-blocking observations

None of these affect the verdict. Recorded so they are not rediscovered later.

1. **`list` on a host with no configuration file at all** exits 1 with
   `global config not found: <path>`, not the documented empty-listing line. The
   page's "with nothing configured" reads naturally as "no `build_ssh` scopes",
   which is what it demonstrates, so this is an ambiguity rather than an error.
2. **`defaultBuildSSHScope` on a two-segment identity** (`host/repo`) returns the
   full identity rather than the host. The page says the default is "the
   repository namespace" without covering that edge.
3. **Implementation asymmetry, not a docs defect.** A run-wide
   `--build-ssh-agent auto` with `SSH_AUTH_SOCK` unset fails with the stable code
   `build_repository_ssh_credential_missing` (`buildssh.go:274`), but a
   configured `"agent": true` scope in the same situation fails through
   `scopeCredentials` (`buildssh.go:314`) with a bare message carrying no code,
   wrapped by `matchScope`. Worth a follow-up on the code, not on this page.
4. **`CHANGELOG.md` has an empty `Unreleased` section** and this whole operator
   credential surface lands without an entry. Not charged to this task: the file
   has not been touched since 0.12.5 (2026-07-14) while schema 7, the seamless
   manager lifecycle and the pinned-agent tail all landed without entries, there
   is no CI check for it and `CONTRIBUTING.md` does not require it. A story-level
   decision, not a rework item.
5. **Presentation nit.** The dry-run provenance example is shown without the
   `<alias>: ` prefix that `credentialReport` prepends. The surrounding prose says
   the line is prefixed, and the producer flagged the construction rather than
   fabricating a prefixed transcript, which is the right trade.

## Verdict

Accepted. All five substantive checklist items are done, the AC holds under
independent verification, the page is accurate everywhere I could check it,
and where the truth was inconvenient (spec drift, a wrong review finding, gates
that do not exist) it was written down instead of papered over.

Reviewer-archetype run: no `commit_ack` supplied. This artifact is the
acceptance evidence for the commit-owning mover, which commits this scope and
then makes the final `done` transition with `commit_ack=scope_committed`.
