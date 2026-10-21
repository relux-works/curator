# THE ONLY CURRENT INSTRUCTION — TASK-261018-ujp9nh: round 2, the exact 1.4.10 binary is available (developer)

Your round-1 stop was correct. At 00:01 local the Muse auto-updater replaced `~/.local/bin/muse-bin-1.4.10-R6427.1` with 1.4.11. The exact 1.4.10 build is still on this host, as a read-only pin kept by the task-board Muse manager:

- path: `$HOME/Library/Application Support/tb-muse/muse-pins/muse-bin-1.4.10-R6427.1`
- `--version` (with `MUSE_NO_AUTO_UPDATE=1`): `Muse Code 1.4.10 (1.4.10-R6427.1)`
- SHA-256: `11cc1ae887b4975bbbe1de7358ee9260a612444df9ce4dbfffa1dd785826dbcf`

## Do this

1. Set `MUSE1410_BINARY` to that path. First verify the digest yourself; the harness must refuse any other digest.
2. **Never write to, move, chmod or delete anything in the `tb-muse` directory.** It belongs to another manager. If the harness needs a copy (for example, for sandbox path rules), copy the binary into your task-local `.temp/`, verify that the copy has the same digest, and use the copy.
3. Then do the whole original brief `gw-MUSE1410-brief.md` unchanged: the harness, the P1–P4 replay per form, the record, the controls, the tests. Every invocation keeps `MUSE_NO_AUTO_UPDATE=1`.
4. Replace the round-1 "blocked intake" content of `TASK-261018-ujp9nh_gaps.md` with the real results; keep one line noting the round-1 intake stop and its cause. Record the binary locator only as `$MUSE1410_BINARY` plus "task-board Muse pin store", never the absolute path.
5. 1.4.11 is out of scope for this leaf; mention it in the gaps file as the next build to qualify.

## Hand-off

Run `task-board handoff TASK-261018-ujp9nh --role developer`, then END YOUR TURN.
