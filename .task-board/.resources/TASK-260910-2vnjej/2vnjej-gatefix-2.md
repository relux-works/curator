# TASK-260910-2vnjej — Windows gate fix 2 (THE ONLY CURRENT INSTRUCTION, with 2vnjej-sec-brief.md)

Good: rev2 (tree 2bc75612) made the failure explicit. Windows now fails 11 tests (run 36460771718) with
`manager_state_unreadable: …\checkpoints\checkpoint.json: checkpoint permits group or other mutation`. Root cause: the checkpoint
private-file check uses Unix permission bits; on Windows Go reports 0666 for ordinary files, so every checkpoint looks group/other-writable.
Fix: on Windows, check the owner-only DACL instead of mode bits — reuse the repo's existing helper (internal/privatedir validateOwnerOnlyDACL,
or internal/buildcache's Windows protection helpers) behind a build-tagged function; keep the unix mode check unchanged. Test fixtures on
Windows must create the checkpoint through the same private-write helper the product uses (owner-only DACL) — do not relax the check.
Keep rev2's CRLF row and the fail-closed behaviour for present-but-unverifiable checkpoints. `GOOS=windows go vet ./internal/registry
./internal/install`; unix `go test ./internal/registry ./internal/install -run 'Checkpoint|Bootstrap|Registry'` with real exit codes.
Set status development; update results (`git diff 2bc75612` non-empty); handoff; END YOUR TURN. No CHANGELOG/LOGBOOK edit.
