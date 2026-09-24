# TASK-260916-1l44nd rework 5 (orchestrator, binding) — Landlock write-confinement is incomplete

Verdict rev5: CHANGES_REQUESTED (TASK-260916-1l44nd_review-verdict-rev5.md; reviewer contract
tests attached). Two HIGH findings, both in `internal/scriptworker/landlock.go`; everything else
(probe design, evidence record, preflight, permit gating, Windows job objects) was verified from
the hosted artifacts and stands. Continue from the revision-5 tree (no checkout/clean/stash).

F1 — wrong UAPI bit: `landlockAccessFSTruncate = 1 << 12` is MAKE_SYM (0x1000); TRUNCATE is
`1 << 14` (0x4000). Consequence: real truncation (`truncate(path,0)`, `open(O_RDONLY|O_TRUNC)`)
outside the grants is NOT handled while evidence claims write-confinement applied; and the
object-typing rule at lines ~79–83 strips the mislabeled bit from non-directories (that was the
real cause of the earlier EINVAL), which after the constant fix would wrongly deny granted file
truncation. Fix: take every access-right constant from `golang.org/x/sys/unix`
(`LANDLOCK_ACCESS_FS_*`) instead of hand-written literals; type rules correctly (TRUNCATE valid on
regular files; MAKE_* / REMOVE_* / REFER directory-only); keep the `/dev/null` file-typed
`READ_FILE|WRITE_FILE` rule; replace tests that derive expected masks from the same wrong
constants with tests against the `unix` constants.
F2 — the handled mask omits directory mutation rights: REMOVE_FILE, REMOVE_DIR, MAKE_DIR,
MAKE_REG, MAKE_FIFO (and MAKE_SOCK/MAKE_CHAR/MAKE_BLOCK/MAKE_SYM, REFER on ABI≥2, TRUNCATE on
ABI≥3, IOCTL_DEV on ABI≥5 as supported) — Landlock permits every unhandled action, so a confined
script can unlink/rmdir/mkdir/mkfifo/rename outside its grants. Fix: handle every filesystem
right the probed ABI supports; grant directory-only rights only beneath the DERIVED writable
directories; never broaden the writable roots.
Rows (ubuntu Test + Race, through the real worker): (a) truncate + O_TRUNC of an existing outside
file denied with content unchanged; a granted individual file overwrites/truncates fine; evidence
`applied` in the same invocation; (b) unlink, rmdir, mkdir, mkfifo, same-directory rename OUTSIDE
the grant denied with state preserved, and the same operations INSIDE a granted directory
allowed; (c) narrowing mutants: omit the real TRUNCATE bit; omit each of the five mutation rights
individually → the corresponding row fails. Keep `TestLinuxLandlockConfinementMatchesProbe`
consistent with the new mask. Update `docs/script-interpreters.md` ("everything else stays
denied" must now be true) and results.md "Revision 6" with the ABI→mask table.
Republish only on a green gate.
