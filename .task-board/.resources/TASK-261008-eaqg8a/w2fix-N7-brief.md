# THE ONLY CURRENT INSTRUCTION — TASK-261008-eaqg8a: fix wave-2 finding N7 (developer)
Read section N7 of `.research/261004_inline-audit-wave-2.md` on main: it names the promised invariant, the production entry, the reproduction and the expected fix. The production boundary is install.Global with globalbins.StageForwarding/ownedTarget.
**Do:**
1. Implement the fix the report describes at that boundary, keeping every negative control the report lists (they must still pass).
2. Regression: the report's probe for N7 (attached as `N7-probe_test.go`, originally at `internal/install/wave2_probe_test.go`) asserts the invariant and was expected-red; adapt it into a permanent test through the production entry with a fitting name (no "wave2" in the name), so it fails without your fix and passes with it.
3. One CHANGELOG line under Unreleased (Fixed or Security), operator-visible wording.
4. R223: NO local `go test` on the mini; compile-only (`go vet ./...`, `go build ./...`). The hosted gate is the arbiter.
5. Results resource (plain text, no archives): the change, the test, and the compile-only tail.
Then `task-board handoff TASK-261008-eaqg8a --role developer` and END YOUR TURN. No LOGBOOK edits.
