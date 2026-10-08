# Mac mini host rules (binding for every go build/test you run here)

- **-work experiment (tb-R136 addition, 2026-10-02).** Run every `go build`, `go test`, `go vet` and `go run` with `-work`, or set `GOFLAGS=-work` in the environment of your shell commands. This keeps the WORK dir after exit. Hypothesis: syspolicyd crashes when it scans a freshly exec'd binary that go has already deleted.
- **Crash counts.** Before and after each long build or test, record `launchctl print system/com.apple.security.syspolicy | grep successive`. Put the counts and the duration in your results.
- **When syspolicyd is down.** If it is not "running", or the crash count just changed, wait about 5 minutes before starting the next build; do not retry in a tight loop.
- **Builds go through the FIFO wrapper (tb-R193): `~/.local/bin/mini-build-lock run <your-run-id> -- go test -work ...`. Never spin on `~/.mini-build-lock` with bare mkdir.**
- **Never edit LOGBOOK.md. Never spell any employer name.**
- **Sandboxed Muse runs (tb-keeper, R193 gap, 2026-10-05):** if your sandbox cannot write the mini-build-lock queue, you must NOT run `go test` unqueued. Stop before the tests, write in your results exactly which test commands still need to run, and hand off. The orchestrator then runs them through mini-build-lock.
- **tb-R194 (2026-10-05):** do NOT run locally on the Mac mini any test that creates and deletes executables: fake claude/codex/muse/git binaries or scripts, script-worker or launcher fakes, install or shim tests that write executables. In curator that includes `cmd/curator`, `internal/scriptworker`, `internal/install`, `internal/runtimestore`, `internal/globalbins` and launcher fakes. Those tests run on the hosted CR gate, which is the arbiter. Narrow pure-logic tests stay local, through mini-build-lock.
- **Do NOT edit `scripts/remote-gate.sh`.** Its snapshot-message privacy fix is already on main (1fc0a93a). Any edit to it conflicts with other leaves.
