# Mac mini host rules (binding for every go build/test you run here)

- **-work experiment (tb-R136 addition, 2026-10-02).** Run every `go build`, `go test`, `go vet` and `go run` with `-work`, or set `GOFLAGS=-work` in the environment of your shell commands. This keeps the WORK dir after exit. Hypothesis: syspolicyd crashes when it scans a freshly exec'd binary that go has already deleted.
- **Crash counts.** Before and after each long build or test, record `launchctl print system/com.apple.security.syspolicy | grep successive`. Put the counts and the duration in your results.
- **When syspolicyd is down.** If it is not "running", or the crash count just changed, wait about 5 minutes before starting the next build; do not retry in a tight loop.
- **Builds go through the FIFO wrapper (tb-R193): `~/.local/bin/mini-build-lock run <your-run-id> -- go test -work ...`. Never spin on `~/.mini-build-lock` with bare mkdir.**
- **Never edit LOGBOOK.md. Never spell any employer name.**
