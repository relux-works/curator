# TASK-260908-3bxxzi — recovered candidate validation

## Candidate

Recovered the task-scoped revision 2 patch into the clean assigned Story worktree on base 1cb41aa. All 16 paths applied successfully, preserving the newer gate/** workflow trigger. No commits, installs, real ax calls, tags, releases, or control-root/home configuration edits.

README covers trusted PATH installation, umbrella discovery, defaults family/locking, Pi preference and no-MCP channel, tracked configuration and diagnostic exits. CHANGELOG 0.1.0 (unreleased) records PRs 4–15 and agents-management v0.5.13. Help and version report 0.1.0-dev against SPEC 0.3.0-draft. Hosted Ubuntu/macOS jobs retained; gated rose-air setup-go/make check lane added.

## Fresh validation

Every command below ran directly as a standalone process in zsh, without pipes. Personally rerun on the recovered candidate:

| Command | Actual exit |
| --- | --- |
| make test | 0; all 11 packages, including help and production goldens |
| make build | 0 |
| make fmt-check | 0 |
| make vet | 0 |
| git diff --check | 0 |

TestRunHelpGolden drives production run via runNoResolve for 2/2 aliases, checking exact stdout and empty stderr with resolution forbidden. Existing production forbidden-flag tests cover 9/9 shapes in direct/tracked modes. This change adds no runtime admission gate; no mutation coverage is claimed. Existing test evidence from prior reports was not used in place of fresh tests. Earlier PR-number mapping is inherited from the attached producer report; the implementation sequence and dependency were checked against local history/go.mod.

## Bounds and handoff

The full landing suite is reserved for the handoff runtime and was not run manually. This base configures a remote GitHub landing gate; its result is runtime-owned and not preclaimed here. Hosted Ubuntu/macOS, rose-air and Windows execution are not verified by this local run. Independent reviewer acceptance, signed scoped PR and integration remain with their owning runs. No real ax or installed umbrella test was run; fake subprocesses only.

## Logbook finding

The resumed worktree was clean despite earlier checked board items and stored candidate patches. Restored revision 2 through git apply and revalidated rather than treating historical checkmarks as current evidence. The newer remote gate trigger is preserved. Campaign rules prohibit LOGBOOK.md edits; this resource and board notes preserve the finding. Current task brief supersedes the old epic installation example: environments §11 rejects the user-bin shim directory, so README directs installation to /usr/local/bin or a dedicated trusted PATH directory.
