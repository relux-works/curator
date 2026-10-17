# Producer brief: BUG-261017-ygfdal, Landlock probe recursion in test binaries

Read the bug description first. Code: `internal/scriptworker/landlock_linux.go`
(`probeScriptLandlock`, `RunLandlockProbe`, `ScriptLandlockProbeMode`),
`internal/scriptworker/main_test.go` (has a TestMain dispatch),
`cmd/curator/main.go` (production dispatch).

Observed: an `install.test` binary was re-executed as
`install.test __curator-script-landlock-probe <dir>`; with no dispatch it ran the
whole `internal/install` suite again, which probed again. The process tree grew
until the host was saturated, and survived the CI job.

Required:

1. No test binary in this module may recurse through the probe. Find every test
   package whose tests can reach `probeScriptLandlock` (directly or through the
   script worker or install paths) and make the hidden mode terminate there.
   Prefer one shared mechanism over per-package copies where the code allows it,
   for example a small exported helper the TestMain functions call, or a guard
   in the probe itself.
2. Defence in depth in the product code: the probe child must never be able to
   start another probe. A child that does not recognise the mode must not run
   unrelated work; pick a mechanism that keeps the existing contract (hidden
   mode, no user-visible surface, sanitized environment) and state it.
3. The probe must not leave descendants behind on timeout: run the child in its
   own process group and kill the group on timeout or cancellation (Linux), or
   show why it cannot have descendants after the fix.
4. Keep the production behaviour of the probe unchanged: present on hosts with
   Landlock, absent otherwise, no error on expected absence.

Negative evidence (required):
- A test that fails on current main: re-executing a test binary of an affected
  package in probe mode must exit promptly with the probe result and must not
  run tests (for example, assert on runtime and on the absence of test output).
  Show it red on main, green with the fix.
- A test for the timeout path: the probe child's descendants are gone after the
  timeout (Linux only; skip with a stated reason elsewhere).
- Positive control: the probe still reports present where Landlock works and
  absent where it does not (existing tests stay green).

Run the full gate (`sh scripts/remote-gate.sh`). Linux CI runs on the
self-hosted runners; report run id and the per-job runner names. No host names,
IPs, personal names or local paths in committed text.

Handoff: Change Request with summary, the list of affected test packages and how
each is covered, the red-on-main evidence, and the gate run.
