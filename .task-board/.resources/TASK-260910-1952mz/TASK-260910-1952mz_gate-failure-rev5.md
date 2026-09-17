# TASK-260910-1952mz — hosted gate failure on Change Request revision 5 (run 35201254365)

Extracted by the orchestrator from the CI evidence (`test-evidence-windows-latest`,
`test/go-test.json`). Linux and macOS lanes are green; the ONLY failing test is
on Windows, in the new native/MSYS identity test, and it is a **test-harness
expectation bug, not a hook or record-identity bug**:

```
internal/shell TestShellHookTrustNativeRecordAuthorizesMSYSSpelling/{bash,sh,bash#01,sh#01}
shell_hook_trust_test.go:1123: stdout lacks "sourced1=1"
  stdout: sourced1=2 / sourced2=2
  stderr: curator: shell_hook_env_changed: C:\Users\...\project\.agents\env.sh changed since approval; run curator hook approve ... to approve the new bytes
          PROBE-1
```

That is the final subcase `native-record-msys-lookup/changed/A-warning`
(`assertTrustOutcome(t, expectA, ...)`). Both earlier subcases passed: the
native-spelling record authorized the MSYS-spelling lookup in profiles A and B
(the identity mapping via `cygpath` works), and the changed-bytes case in
profile B refused with one `shell_hook_env_changed` warning. In profile A the
hook did exactly what §8 asks for changed bytes: warn once on the first
activation (before `PROBE-1`), stay silent on the second, and source the
current bytes — which now export `CURATOR_PROJECT_ENV=2`.

The harness however compares against the default marker `1`: `expectA` is a
copy of `expectB` with `Sourced = true` but without `SourcedMarker`, and
`hookTrustCase.SourcedMarker`'s own comment says hand-built cases whose bytes
export another value must set it explicitly.

Fix: in `TestShellHookTrustNativeRecordAuthorizesMSYSSpelling` set
`expectA.SourcedMarker = "2"` (the changed bytes are
`export CURATOR_PROJECT_ENV=2`). Nothing in the hook, `hookapproval`, or the
Windows identity code needs to change for this failure. Do not weaken or skip
the subcase, and do not touch the profile-A semantics.

Also note: `bash#01`/`sh#01` show the same interpreter twice in the subtest
names (Git Bash `bash.exe` and `sh.exe` both discovered on PATH and again
under the Git install root). `posixTrustShells` claims interpreters resolving
to the same binary run once — if that dedupe is meant to hold on Windows,
compare resolved paths case-insensitively / after `EvalSymlinks`; harmless for
the gate but it doubles the Windows lane time.

Re-run `go test -count=1 ./internal/shell/...` and hand off again; the
runtime re-runs the hosted gate.
