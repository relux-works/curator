### 5.8 Activation boundary for `launch-delegated/1`

**Q1 is signed off** (Ivan, 2026-10-21, relayed by `ivmbp_curator_orch`):
`launch-delegated/1` is accepted as the interim route, for Claude Code and
Codex now and for Muse once its task-board consumer qualifies (§7). The
launcher therefore applies the **After sign-off** column below to every
launch.

`delegated_trial` (system-layer `launcher.json`, §3.1) authorised a live
trial before the sign-off. It is now withdrawn: the launcher still validates
it as a boolean, so existing machine files stay valid, but it changes no
route. A later schema revision drops it. The launcher no longer emits
`session_host_delegated_unauthorized`. The code stays reserved, because
historical operation outcomes may carry it.

The matrix below applies whenever the launcher **would use the
`launch-delegated/1` route**, whatever else the manager offers. That
includes a manager that also offers `launch-exact/2`, in every case where
§5.2 selects delegated. For example, `werkstatt up` and every detached
launch use delegated, because exact cannot detach:

| Invocation | Before Q1 sign-off, `delegated_trial` absent or false | Before sign-off, `delegated_trial: true` | After sign-off |
| --- | --- | --- | --- |
| `--hosted` | Refuse `session_host_delegated_unauthorized` (16) | Hosted | Hosted |
| `werkstatt up` | Refuse `session_host_delegated_unauthorized` (16) | Hosted | Hosted |
| No flag, `host_mode: auto` | Native with `notice-unselected` | Native with `notice-unselected` | Hosted |
| No flag, `host_mode: hosted` | Refuse `session_host_missing`, `reason=no manager selected` | Refuse, the same | Hosted |
| `--native`; or `host_mode: native` with neither `--hosted` nor `werkstatt up` (`werkstatt up` is always hosted, and its row above applies) | Native | Native | Native |

- The trial flag never changed the two implicit rows. Since the sign-off it
  changes no row at all.
- The two pre-sign-off columns stay as history. The decision function still
  covers them with unit vectors, but no launch takes them.
- Launches that use the `launch-exact/2` route are outside this matrix.

The paired test vectors are the 15 cells of the matrix, each with its
expected outcome.

