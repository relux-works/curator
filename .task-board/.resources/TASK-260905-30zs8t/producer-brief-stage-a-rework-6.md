# Producer brief: stage (a) core — rework 6 (Windows CI: insteadOf fixtures break on backslash paths)

Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-a-core`, branch
`feat/agent-environments-stage-a`, head `59afa6ee`, PR https://github.com/relux-works/curator/pull/59.
Lint, both Race lanes, ubuntu and macOS Test, the Interop conformance gate, the Naming gate and all
three Gate self-tests are green. **`Test (windows-latest)` fails with 20 failing tests** — 17 in
`internal/envprofile` and 3 in `cmd/curator`, all introduced by the `insteadOf` fixtures the rework
cycles added:

```
TestMigrationHonoursSystemLockedRevocation, TestMigrationSucceedsWhenSystemPolicyPermits,
TestRevokedSkillMemberIsRefused, TestRevokedMCPMemberIsRefused, TestFileRequirementSourceIsRefused,
TestMachineUseSkipsScopedAdapter, TestInstallUseSkipsScopedAdapter, TestScopedUseStillSwitchesOnlyThatHome,
TestSyncWritesScopedHomeOnce, TestCanonicalIdentityUnifiesSpellings, TestRevokedSourceIsRefused,
TestMCPAllowlistRefusesOutsidePackage, TestMigratedSkillSourceIsCanonical, TestInstallGitResolvesNetworkIdentity,
TestInstallSurfacesUnresolvedMCPCommand, TestScopedUseAndClear, TestEnsureDefaultMigratesGlobalSkills,
cmd/curator: TestProfileScopedUseAndClear, TestProfileUpdatePinnedTagIsUnchanged, TestProfileMachineUseSkipsScopedAdapter
```

## The defect
The failure text shows git receiving a path whose separators vanished:

```
git clone -- https://example.com/skills/hello C:\Users\RUNNER~1\AppData\Local\Temp\Test…\003\profile-repos\example.com_skills_hello
  fatal: 'C:UsersRUNNER~1AppDataLocalTempTest…002' does not appear to be a git repository
```

The `insteadOf` value (and/or the fixture remote path) is written with native Windows backslashes into
git configuration, where a backslash is an escape: `\U`, `\A`, `\L`, `\T` are consumed and the path
collapses. On Unix this never appears, which is why local runs and the ubuntu/macOS lanes are green.

## Fix
Write every git-facing path in the fixtures in a form git parses identically on all three platforms:
`filepath.ToSlash` the temporary directory and use a `file:///C:/Users/...` URL (three slashes, forward
separators) for the local fixture remote and for the `insteadOf` target; never interpolate a raw
`t.TempDir()` into a git config value. Apply it in one helper the fixtures share, so a future fixture
cannot reintroduce it. Keep the fixtures hermetic and offline as they are today, and keep the
`file://`-operand **rejection** of F12 intact — this is about the fixture's own remote, not about
admitting `file://` operands into the product path.

Do not weaken or skip a Windows case to make the lane green. If any single test genuinely cannot run on
Windows for a host-capability reason, it goes into `.github/ci/platform-cases.tsv` with that reason and
a registered skip class — and the report must say which and why.

## Verification
You cannot run Windows locally. Verify what you can (`go build`, `go vet`, `gofmt`, `golangci-lint`, the
full `go test` on this host, the vector families, `gate-selftest.sh`, the platform-case gate for the
three GOOS values), then **push and watch `gh pr checks 59 --watch`** until `Test (windows-latest)` and
every other check is green. Iterate on the PR until it is; that is the acceptance signal for this
rework. Signed commits on top of `59afa6ee`, no rewrite, plain push (no force). Attach
`TASK-260905-30zs8t_rework-report-6.md` with the fix, the shared helper, and the final check summary;
`task-board handoff TASK-260905-30zs8t --role developer`. Never write LOGBOOK.md or anything into the
control root.
