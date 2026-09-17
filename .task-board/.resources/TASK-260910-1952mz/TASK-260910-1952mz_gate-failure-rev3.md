# TASK-260910-1952mz — hosted gate failure on Change Request revision 3 (run 35189457421)

Extracted by the orchestrator from the CI evidence. The POSIX/dash repair
and the record validation now pass on every lane; the ONLY failing test is
the new symlink-identity test's PowerShell subtest, on Linux, macOS and
Windows alike:

```
internal/shell TestShellHookTrustResolvesSymlinkedProject/powershell
shell_hook_trust_test.go:900: stdout lacks "sourced1=1"
  stdout: sourced1=no / sourced2=no
  stderr: curator: shell_hook_env_unapproved: <tmp>/alias-project/.agents/env.ps1 is not approved; ...
```

So the PowerShell hook still looks up the record under the alias spelling
(`.../alias-project/.agents/env.ps1`) instead of the resolved real path, while
the record was written for the real path. The POSIX subtest passes, i.e.
`realpath` resolution was implemented for the POSIX hook but not (or not
correctly) for the PowerShell hook: resolve the candidate through the
symlink chain before the lookup (e.g. `(Get-Item -LiteralPath $p).ResolvedTarget`
/ `[IO.Path]::GetFullPath` on the link target, or `Resolve-Path` on the
directory's `LinkTarget` for the `.agents` parent), and make the Go
canonicalization and the PowerShell resolution agree on the same string
form (Windows drive-letter case, separators). Note the test runs on hosts
with `pwsh` on all three lanes, so keep the host-capability skip only for
hosts without `pwsh`.

Re-run `go test -count=1 ./internal/shell/...` locally (with `pwsh`
installed here if available) and hand off again; the runtime re-runs the
hosted gate.
