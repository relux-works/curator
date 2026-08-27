# TASK-260822-4p3dcq: docs-build-ssh-surface

Ready for review.

## Where the work lives

Worktree `.temp/TASK-260822-4p3dcq/worktree`, branch
`task/TASK-260822-4p3dcq-docs-build-ssh`, based off `origin/main` **6a9b201**.

The branch carries the accepted predecessor chain, applied verbatim from
`TASK-260822-b0wg3a_final.patch` (applies clean, `git apply --check` exit 0),
and staged as the baseline. The delta charged to this task is the unstaged
working-tree change and is exactly two files:

| File | State |
| --- | --- |
| `docs/build-ssh.md` | new, 494 lines |
| `README.md` | one bullet added under "What Curator manages", linking the new page |

Nothing under `cmd/`, `internal/`, `.github/` was touched by this task.

The checkpoint branch `handoff/cocoaskills-parity-20260731` has no
`internal/buildrepo` and no `internal/install/external.go`, so `origin/main` is
the only base this story can be documented against, the same base every other
task in the story used.

## What the page documents

`docs/build-ssh.md`, sections in order:

1. **What needs a selection**: effective state, not declared (`Spec §6.4`); an
   HTTPS or local-path substitution moves a repository off the SSH transport and
   it then needs nothing.
2. **Where a selection comes from**: the four-source precedence table (flags >
   `CURATOR_BUILD_SSH_*` > `build_ssh` scopes > interactive precheck > fail
   closed), the field-by-field flag/env merge, and why a bare
   `--build-ssh-known-hosts` is not a selection.
3. **The three selection shapes**: identity-only, agent-only, pinned agent,
   mapped to config entry, CLI spelling, and the exact authentication tail.
4. **The `build_ssh` configuration field**: entry grammar, path rules, Windows
   absolute forms, why `"agent": false` is rejected, and the full fail-closed
   parse table with the real error strings.
5. **Scope matching**: canonical `host/path` identity (`Spec §6.3`),
   segment-aware matching as in `Spec §6.1`, longest scope wins, and why the
   identity is the only matching key (`Spec §12.2`).
6. **`curator config build-ssh`**: add/list/remove transcripts, the
   optional-valued `--agent`, the added/replaced distinction, the empty-listing
   stderr line, and the refusal cases.
7. **Run-wide flags and environment**: the three flags with their real help
   text, `--build-ssh-agent auto`, and the tilde trap (config entries expand a
   leading `~/`, flag and environment values do not).
8. **The install-time precheck**: discovery, the interactive transcript, the
   non-interactive and dry-run diagnostic, and the dry-run provenance line.
9. **Host keys**: `StrictHostKeyChecking=yes`, the three-step known-hosts
   resolution, and the refusal when none yields a file.
10. **Limits worth knowing**: the three items below.

## Spec citations

Every normative claim cites `curator-spec` and no other implementation.

| Cited | Used for |
| --- | --- |
| `Spec §12.2` Operator credentials and signing | credentials are operator-owned; no manifest, descriptor, repository, substitution, or marker may select them |
| `Spec §6.3` External repository source identity and exact lock | canonical `host/path` identity; the SSH raw-path ASCII restriction |
| `Spec §6.1` Canonical source identity | segment-aware matching (`h/a` matches `h/a/b`, never `h/a-evil`) |
| `Spec §6.4` Declared and effective external source state | effective state decides whether a selection is needed |
| `Spec profiles/manager.md §11.3` SSH isolation | the authentication tails and the "no ambient anything" rule |

**Flagged honestly in the page:** `profiles/manager.md §11.3` on `curator-spec`
main today admits only *two* tails. The third, pinned-agent tail Curator
implements (accepted here in 6a9b201) lives on the unmerged spec branch
`spec/pinned-agent-authentication-tail`, commit `38232d3`. The page says so
rather than presenting three tails as settled protocol. CI `SPEC_PIN` is
`00b1688`, which also predates that change; the conformance corpus does not
encode the tail, so no pin or checksum is affected.

## Checklist item 5 is documented, with a correction

The item says: *"a repository namespace can contain characters the build_ssh
scope grammar rejects (PortableComponent vs scopeSegmentRE), in which case the
suggested default scope must be widened to the host."*

That is `TASK-260822-b0wg3a` review finding 3. Verified at first hand and it does
not reproduce as worded. Evidence, from a throwaway test run against the real
`buildrepo.ParseSource` and `config.ValidBuildSSHScope` in this worktree:

| Source | Result |
| --- | --- |
| `ssh://git.example.com/team+infra/app.git` | parse error: `repository SSH path must be portable ASCII` |
| `git@git.example.com:team+infra/app.git` | parse error: `repository source must use HTTPS or SSH` |
| `ssh://git.example.com/team@infra/app.git` | parse error: `repository SSH path must be portable ASCII` |
| `https://git.example.com/team+infra/app.git` | identity `git.example.com/team+infra/app`, transport `https`, default scope invalid, host scope valid |
| `ssh://git.example.com./team/app.git` | identity `git.example.com./team/app`, default scope invalid, **host scope also invalid** |
| `ssh://git..example.com/team/app.git` | identity `git..example.com/team/app`, default scope invalid, **host scope also invalid** |
| `ssh://git-.example.com/team/app.git` | identity `git-.example.com/team/app`, default scope invalid, **host scope also invalid** |

Two corrections follow:

1. The reviewer's `PortableComponent` reference is the `internal/identity`
   path, which governs skill sources. External build repositories parse through
   `buildrepo.ParseSource`, which for SSH restricts the raw path to
   `^[A-Za-z0-9._/-]+$` (`Spec §6.3`) before Git or SSH starts. That is exactly
   the alphabet `scopeSegmentRE` admits, so the **path** half of the divergence
   cannot occur for the only transport `build_ssh` ever selects for. An HTTPS
   identity can carry such a namespace, but an HTTPS repository never reaches
   the scope suggestion.
2. The divergence that does reproduce is the **host**: `hostRE`
   (`^[A-Za-z0-9][A-Za-z0-9.-]*$`) admits `git.example.com.`, `git..example.com`
   and `git-.example.com`, which `scopeHostRE` rejects. In that case widening to
   the host does **not** rescue the suggestion, because the host is itself not a
   valid scope. No `build_ssh` entry can cover such a repository at all; the
   covering path is `--build-ssh-agent` / `--build-ssh-identity` or the
   `CURATOR_BUILD_SSH_*` environment.

The page documents the real behaviour, including the HTTPS contrast, rather than
the wording of the item.

## Checklist item 4 is documented as written

The agent-less identity tail is
`-o IdentitiesOnly=yes -o IdentityAgent=none -i <path>`
(`internal/buildrepo/admission.go`). With a `*.pub` path and no agent there is
nothing to sign with, and `ValidateOperatorSSHCredentials` proves only that the
path resolves to a regular file. The page says the identity-only menu entries
and the `--identity ~/.ssh/<key>.pub` diagnostic lines are selectable choices
that fail at authentication time, and names the sound shapes: a private key for
identity-only, or `--agent --identity <key>.pub` for the pinned-agent form,
which is the default menu entry.

A third limit is documented as well (`b0wg3a` review finding 1): a prompted
credential is written to the configuration file but not folded into the loaded
in-memory configuration, so `curator install --all` re-asks per target.

## Examples match actual output

No transcript in the page is hand-written. Provenance for each:

| Example | Produced by |
| --- | --- |
| `config build-ssh add/list/remove` transcripts, added vs replaced, the empty-listing stderr line, the two refusal cases | a binary built from this worktree, run against a scratch `CURATOR_CONFIG` |
| the `build_ssh` JSON block | that same scratch config file after the four `add` invocations, verbatim |
| the seven fail-closed parse messages | seven hand-written malformed configs, each run through `curator config build-ssh list` |
| the three `-build-ssh-*` flag help lines | `curator install -h` |
| interactive prompt transcript, both fail-closed diagnostics, the provenance line | a throwaway `docsample_test.go` in `internal/install` driving `resolveBuildSSH` and `InteractiveBuildSSHResolver` directly, captured and then deleted |

The one construction: the provenance example shows the line `resolveBuildSSH`
returned, without the project-alias prefix `credentialReport(label)` prepends,
because producing a real prefixed line needs a full project fixture. The prefix
is described in prose instead of shown in a fabricated console transcript.

Raw capture: `TASK-260822-4p3dcq_docsamples.log`.

The prompt transcript shows one question pair for two uncovered repositories
under one namespace, which is the "one scope covers every sibling" behaviour
rather than a truncated capture.

## Gates, each run as a standalone process in the worktree

| Command | Exit | Result |
| --- | --- | --- |
| `gofmt -l .` | 0 | no output |
| `go build ./...` | 0 | clean |
| `go vet ./...` | 0 | clean (needed `git submodule update --init` first; the worktree had no `tuitestkit`, and vet failed 1 on the missing replacement directory until it was checked out) |
| `golangci-lint run ./...` | 0 | `0 issues.` |
| naming gate (`.github/workflows/ci.yml` job `naming-gate`, run inline) | 0 | `naming gate: clean (one README line)` |
| em dash / guillemet check (`grep -n '-\|–\|«\|»' docs/build-ssh.md README.md`) | 1 | no match, which is the pass condition for the prose rule of `docs/implementation-plan.md` section 3 |
| relative link and in-page anchor resolution over `README.md`, `docs/build-ssh.md`, `CONTRIBUTING.md` | 0 | 13 relative targets and 10 anchors all resolve, 0 misses |
| `go test -count=1 -timeout 30m ./internal/config/ ./internal/install/... ./internal/buildrepo/ ./cmd/curator/` | see `TASK-260822-4p3dcq_gotest.log` | unchanged by this task, run to prove the tree the page describes still behaves |

**Not run, and why:** the repository has no markdown linter, no link checker,
and no docs build step. There is no `make docs` target and no docs job in
`.github/workflows/ci.yml`. "Docs build/lint green" was therefore satisfied by
the gates that do exist and apply to prose: the naming gate, the prose style
rule, and an explicit link/anchor resolution pass. The full `go test ./...` was
not run for this task; the delta is two markdown files and touches no Go code,
and the predecessor chain's full suite was already re-run green by the
`b0wg3a` reviewer at `RUN-260822-fad8ab`.
