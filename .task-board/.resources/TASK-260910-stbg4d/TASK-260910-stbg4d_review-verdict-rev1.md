# TASK-260910-stbg4d — review verdict, revision 1: CHANGES_REQUESTED

Reviewer run RUN-260919-a9273d (claude-opus-5, independent). Read-only:
no repository file was modified; mutants ran in a disposable clone under
/tmp, never in the Story worktree. Shell: zsh; exit codes are the
binary's own (`$pipestatus[1]` / `echo $?`), no unverified pipes.

## Candidate identity

- CR-TASK-260910-stbg4d-1, base `97a161da43b1…`, candidate tree
  `071b60a5a999acddc26be5b8840e678deb153135`.
- Working tree of `.temp/STORY-260910-1cnwwp/worktree` re-hashed through a
  temporary index (`git read-tree HEAD; git add -A; git write-tree`) →
  `071b60a5…` before and after the review: exact candidate, unchanged.
- Hosted gate: `gh run view 35439578509` → conclusion `success`, headSha
  `15a38507896294d17f1803826a15d9342ef91ee9`; `git cat-file -p 15a3850` →
  `tree 071b60a5…`, `parent 97a161da…`. The green gate ran on the exact
  candidate tree (ubuntu/macos/windows tests, race lanes, lint, naming,
  interop, gate self-tests; rose-air and candidate suite skipped, not
  passing).

## Independently rerun on the exact candidate (compiled test binary, `-count=1`)

- Unit/help: `TestWithDraftRemediationTable|TestDraftAttemptTransportTable|TestDraftAttemptClauseShape|TestDraftDocsPinExamples|TestEndpointExhaustionCarriesRemediation|TestDraftProjectResolveHelp|TestDraftInstallStatusHelpGated` → 7/7 PASS, exit 0.
- `TestDraftRemediationThroughCLI` → 15/15 rows PASS (4.6 s), exit 0.
- `TestDraftDocumentedLocalShapesThroughCLI` (5 rows) + `TestDraftMachinePolicySetupThroughCLI` + `TestDraftDocumentedGitRevisionThroughCLI` + `TestDraftFetchFailureSanitizedThroughCLI` → PASS (24.1 s / 7.2 s / 12.1 s / 0.1 s), exit 0.
- `TestDraftLaunchConsumesFrozenRuntimeThroughCLI` + `TestDraftRepairRestoresDriftedContentThroughCLI` → PASS (26.5 s / 11.3 s), exit 0.
- Legacy/sibling regression, switch unset (`env -u CURATOR_DRAFT_SOURCES_V1 -u CURATOR_CONFIG`): mask `TestUsage|TestRun|TestStatus|TestInstall` → 27 tests, PASS, exit 0; mask `TestProjectResolve|TestProjectRefresh|TestDraftStatus|TestClassifyDraftMember` → 41/41 PASS, exit 0.
- `go build ./cmd/curator/ ./internal/buildrepo/` exit 0; `go vet` same packages exit 0; `gofmt -l cmd internal` empty; `git diff --check base..candidate` exit 0. Lint: accepted from the hosted Lint job on the exact tree (not rerun locally).

## Switch-off byte identity, base binary vs candidate binary (v1 project, `CURATOR_DRAFT_SOURCES_V1` unset)

12 invocations compared on stdout+stderr+exit: `install -h`, `status -h`,
`upgrade -h`, `install --bogus`, bare `curator`, `project resolve`,
`project refresh app`, `project`, `status --bogus` → 9/12 byte-identical.
`project resolve -h`, `project resolve --help`, `project resolve help` →
3/12 DIFFER: base prints the frozen v1 path report (alias/path/skillfile/
skills/bin, exit 0); candidate prints the 40-line help golden including
the draft workflow and example sections (exit 0). See finding F2.

## Narrowing mutants (disposable clone at the gate commit; each attributed by its failure line)

| id | mutant | killer | result |
|---|---|---|---|
| A | `cmdInstallMode` runs `cmdProjectResolve` before every non-dry-run install (install silently rescans/re-resolves live inputs) | `TestDraftLaunchConsumesFrozenRuntimeThroughCLI` — `draft_diagnostics_test.go:819: shim after reinstall without refresh = "mutated\n", want frozen "original\n"` | KILLED, exit 1 |
| B | `draftAttemptClause` appends `attempt.URL` | `TestDraftFetchFailureSanitizedThroughCLI` (`:758 stderr leaks endpoint detail "https://invalid.invalid/no-such-repo.git"`) and `TestDraftAttemptClauseShape` (`:130`) | KILLED, exit 1 |
| C | `printFailures` prints `message` without `withDraftRemediation` (remediation dropped at the install/status print site) | `TestDraftRemediationThroughCLI` rows snapshot-changed, audit-unavailable, audit-rejected (`:577 stderr misses remediation …`) | KILLED, exit 1 |
| D | docs/cli.md heading "(opt-in, unreleased)" → "(opt-in, released)" | `TestDraftDocsPinExamples` (`:943 docs/cli.md misses "opt-in, unreleased"`) | KILLED, exit 1 |
| D' | README heading only → "(opt-in)" | README pin still satisfied by the link anchor `#draft-skillfile-sources-opt-in-unreleased` | SURVIVES (bound: the README "unreleased" pin is a URL fragment, not the label) |

## Findings

### F1 (must fix — this leaf's diagnostics surface): draft-lane clone failures render as `unknown: unclassified failure`, and the documented `availability-auth` fallback never advances on a fresh clone

`fetchDraftRepoAllowingFallback` (cmd/curator/project_resolve.go:214-223)
classifies `gitFailureDetail(cloneErr)`, i.e. the full stderr of
`git clone -- <url> <dir>` (internal/gitops/gitops.go:71). git prefixes
that stderr with the progress line `Cloning into '<dir>'...` even when
stderr is not a terminal; the closed line table
(internal/buildrepo/transport.go:303, "any unmatched line … a progress line
… FailureUnknown") therefore reports every clone failure as `unknown`,
which is fail-closed. The strict build lane never sees this line (it does
not run `git clone`), so the sibling classifier is not wrong — the input
this leaf feeds it is.

Reproduction (real binary built from the candidate, bootstrap + `project
add app`, `CURATOR_DRAFT_SOURCES_V1=1`, `GIT_TERMINAL_PROMPT=0`):

```
Skillfile.json: {"schema_version":2,"sources":{"kit":{"git":"https://invalid.invalid/no-such-repo.git","tag":"v1"}},"skills":[{"name":"review","from":"kit","directory":"skills/review"}]}
$ curator project resolve app            # exit 1
warning: kit: endpoint 1 (https): unknown: unclassified failure
curator: repository_endpoint_unavailable: kit: identity invalid.invalid/no-such-repo: endpoint 1 (https): unknown: unclassified failure; verify the network path …

source-policy.json: {"schema_version":1,"repositories":{"invalid.invalid/no-such-repo":{"endpoints":[{"url":"https://invalid.invalid/no-such-repo.git","authentication":"team-https"},{"url":"git@invalid.invalid:no-such-repo.git","authentication":"team-ssh"}],"fallback":"availability-auth"}}}
$ curator project resolve app            # exit 1
warning: kit: endpoint 1 (https, provider "team-https"): unknown: unclassified failure
curator: repository_endpoint_unavailable: kit: identity invalid.invalid/no-such-repo: endpoint 1 (https, provider "team-https"): unknown: unclassified failure; endpoint 2: not attempted; verify the network path …
```

git's actual stderr for that clone is two lines: `Cloning into
'…'...` then `fatal: unable to access '…': Could not resolve host:
invalid.invalid`. Probe against `buildrepo.ClassifyFetchOutput`: with the
first line → `unknown`; without it → `availability` (probe test in the
disposable clone, `ok … 0.372s`).

Why it is this leaf's finding: revision 1 replaced the raw `cloneErr`
(which leaked the URL but showed the cause) with the sanitized class — for
the most common failures (DNS, connection refused, 401/403) the operator
now gets "unknown: unclassified failure", which is not an actionable
diagnostic, and docs/cli.md "Machine policy setup" (this revision)
documents fallback semantics the wired lane cannot deliver ("endpoint 2:
not attempted" above). `TestDraftFetchFailureSanitizedThroughCLI` avoids
asserting the class, so the suite cannot see it.

Requested: skip git's fixed clone framing line (`^Cloning into '[^']+'\.\.\.$`)
before classification — either in `fetchDraftRepoAllowingFallback`
(cmd/curator) or as an additive framing entry in the closed line table —
and add production-entry rows through `project resolve`: (a) a fake-git
arm emitting the DNS stderr → clause reads `endpoint 1 (https):
availability: endpoint unavailable (…)`; (b) a two-endpoint
`availability-auth` policy with arm 1 DNS-failing and arm 2 succeeding →
resolve exits 0 with the lock bound to the alternate; (c) a narrowing
mutant re-poisoning the first line reported killed. Assert the class in
the existing sanitized-failure test as well (a fake-git arm also removes
its live DNS lookup of `invalid.invalid`).

Bound (sibling logic, not requested here): after a fetch failure the
"recloning from the alternate endpoint" warning is followed by a clone
loop that starts at endpoint 1 again; only the wording is this leaf's.

### F2 (must fix — acceptance clause "released v1 behaviour retained byte-identical with the switch off"): `project resolve|refresh -h|--help|help` changes switch-off output and prints draft text without the switch

cmd/curator/main.go:1174-1177 intercepts `-h`, `--help` and the bare word
`help` regardless of `CURATOR_DRAFT_SOURCES_V1`. At base those spellings
resolved the current project (`projectRootArg` skips dash arguments;
`help` walks up from `./help`), so with the switch off the candidate
changes the v1 output of three spellings (measured above: base = v1 path
report, candidate = help golden with the draft workflow and examples),
prints draft-lane instructions to an operator who has not opted in, and
is inconsistent with the leaf's own rule for `install -h`/`status -h`
(`appendDraftUsage` gates the draft section on the switch). The bare
`help` positional additionally hijacks a legitimate project alias or
directory named `help` (`curator project add help ./help; curator project
resolve help` → help text, not a resolve).

Requested: drop the bare `help` word; with the switch off print only the
frozen-v1 part (or leave the frozen behaviour untouched) and append the
draft workflow/examples only when `install.DraftSourcesEnabled` is true —
the same rule as `appendDraftUsage` — and cover both halves the way
`TestDraftInstallStatusHelpGated` does. Document the chosen rule in
docs/cli.md ("Run `curator project resolve -h` …").

### F3 (should fix — draft labelled accurately): docs point Skillfile-source operators at `source-providers.json` for a lane that never reads it

docs/cli.md "Machine policy setup": "Configure the named authentication
providers through the manager's existing operator mechanism
(`source-providers.json` beside the same configuration)". The Skillfile
source alias path (`fetchDraftRepoAllowingFallback` → `gitops.Clone/Fetch`)
binds no provider: authentication is the operator's ambient Git
credentials, and the `authentication` identifier is validated but unused
on this lane (the producer's own bound in results.md). An operator
following this section would configure providers that have no effect on
`project resolve`. State it explicitly (providers apply to the external
build lane under `CURATOR_DRAFT_TRANSPORT_RESOLUTION`; Skillfile sources
use ambient Git credentials) so the opt-in support is labelled accurately.

### Nits (non-blocking)

- `TestDraftRemediationThroughCLI` has two rows named `alias-unknown`
  (source vs repository) → `alias-unknown#01` in evidence; rename.
- README "unreleased" pin is satisfied by the link anchor (mutant D').
- `TestDraftFetchFailureSanitizedThroughCLI` performs a live DNS lookup of
  `invalid.invalid`; prefer the existing fake-git arm helper.

## What is accepted as-is (do not re-open in the rework)

Remediation table and its wiring at `printFailures`, `cmdProjectResolve`,
status drift (sanitized, idempotent, already-guided classes untouched,
v1 messages untouched — mutant C killed); the revision-2 exhaustion shape
in `endpointExhaustion`/`fetchExhaustion`/`draftAttemptClause` (no URL,
scheme, `.git`, raw tool output or secret in user-facing text — mutant B
killed; a userinfo token in a declared URL is refused at parse without
echoing the URL); the `install -h`/`status -h`/`upgrade -h` gating
(byte-identical with the switch off against the base binary); documented
local/absolute/Git/collection shapes and the machine-policy setup exercised
through `run()` (positive rows) with the `repository_policy_invalid` bad-pin
negative row; frozen-launch proof through an installed shim (mutant A
killed); repair row; troubleshooting sections for all 15 classes with
remedy keywords pinned; `FailureClass.Reason()` additive accessor; Windows
skip reasons from the declared vocabulary (`test transport wrapper is
POSIX-only`, `executes POSIX skill commands`; `git is not available`
follows existing precedent and never fires on hosted runners).

## Verdict

CHANGES_REQUESTED → `to-dev`. F1 and F2 are required for revision 2 (F1
in cmd/curator/project_resolve.go plus tests; F2 in cmd/curator/main.go
and draft_help.go plus a gating test); F3 is a docs sentence. Run goal
queried: run is not goal-bound; no directives recorded.
