# THE ONLY CURRENT INSTRUCTION — TASK-261012-13lcu6: launcher release prep for v0.3.0 (Muse 1.4.6 and `muse-wonder` admission, fail-closed `--hosted`) (W-LRM1, developer)

Repository: the agent launcher (`curator-run`). The owner approved a v0.3.0 launcher so that `curator run muse --profile P --model muse-wonder` works natively on the operator host. The tag itself is cut later by the orchestrator, not in this task.

**Do this:**
1. **Pin bump.**
   - Move `github.com/relux-works/skill-agents-management` from v0.5.53 to **v0.5.62**:
     - since v0.5.58 there is no Muse release table: any well-formed build the binary attests is admitted;
     - the frozen-copy tuple is optional, and without it `muse` resolves on the handed PATH;
     - v0.5.62 adds the `muse-wonder` alias and accepts Curator launch-env-fragment v3.
   - Adapt every API drift in `cmd/curator-run` and `internal/` (for example `ProbeToolRelease`, the echoed `ToolRelease` claim and the refusal on a mismatch, permission-mode resolution), so the launcher builds and its existing tests keep their meaning.
   - Update tests and fakes that pinned Muse 1.4.1 or 1.4.2 behaviour only where the module change requires it. Never weaken a refusal test to make it pass.
2. **Owner decision 2a (2026-10-12): `--hosted` must not look working in this release.**
   - `--hosted` keeps failing closed with the existing refusal `session_host_missing`, before any handoff, and with no new refusal code.
   - `--help`, README, SPEC and CHANGELOG state that hosted launch is unavailable until the swarma dispatcher path ships.
   - Machine and operator hosted defaults must never start a hosted launch either.
   - Keep the native path, `--native`, `--untracked` and direct `--network` unchanged.
3. **CHANGELOG.** Turn "Unreleased" into a `0.3.0` section marked "(pending tag)" that lists exactly what ships: the pin bump with Muse admission and `muse-wonder`, direct `--network` as already landed, and the fail-closed `--hosted` notice.
4. **SPEC.** Update the version line and the affected sections only.

**Rules:**
- Public repository: no first names, host names, local absolute paths, session links or board-resource links.
- On this host, compile-only checks (`go build ./...`, `go vet ./...`) are allowed. Never run `go test` or any suite locally: the configured hosted gate at handoff is the arbiter.
- Never write LOGBOOK.md.
- Do not tag, release or push to `main`.

**Hand-off:**
1. Attach `TASK-261012-13lcu6_gaps.md` (what changed, the drift you adapted, and what stays open, such as Muse `--network` tuples and BUG-261011-sr6fvi).
2. Run `task-board handoff TASK-261012-13lcu6 --role developer`.
3. END YOUR TURN.
