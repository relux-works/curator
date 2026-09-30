# TASK-260930-3vni9d — carrier: content-hash v2 spec without the skillfile-sources half (THE ONLY CURRENT INSTRUCTION)

TASK-260917-2vapkz rev3 was ACCEPTED: three review rounds, and the framing hashes were recomputed independently. Its full content, rev3
merged onto spec main 2b649c4, is commit `refs/campaign/2vapkz-rev3-full-526a9aa` (526a9aa0, parent 2b649c4). It cannot land as-is:
the spec Implementations job pins csk, which pins the digest of conformance/skillfile-sources-v1/index.json. So the skillfile-sources
half moves to TASK-260930-3ny11n, to be landed later in lockstep with csk. Your Story worktree is fresh on spec main (2b649c4).
1. `task-board m 'set_status(TASK-260930-3vni9d, status=development)'`.
2. `git diff 2b649c4 refs/campaign/2vapkz-rev3-full-526a9aa -- . ':!.task-board' > $TMPDIR/hv2.patch && git apply $TMPDIR/hv2.patch`.
3. Then restore these to byte-identical 2b649c4, removing the added ones:
   - everything under conformance/skillfile-sources-v1/;
   - everything under schemas/skillfile-sources-v1/;
   - protocol/skillfile-sources.md;
   - the skillfile-sources manifest_sha256 field in release/1.0.0-rc.13.json.
4. Edit only the text in core.md, registry.md or the CHANGELOG entry that names the skillfile-sources carriers (skillfile-lock-v2,
   install-marker-v6, source-audit-v2). Replace it with one sentence saying those carriers adopt hash_version in follow-up
   TASK-260930-3ny11n. Change nothing else.
5. Regenerate the vectors, then run `tools/validate.py` and `make regenerate-check`, recording the real exit codes. Show that
   `git diff --stat 2b649c4 -- conformance/skillfile-sources-v1 schemas/skillfile-sources-v1 protocol/skillfile-sources.md` is empty.
   Report the new conformance/v1 manifest sha256.
6. In the results, list every path that differs from the rev3-full commit, with the reason for each.

No LOGBOOK.md. Never spell any employer name. Update the results, run `task-board handoff TASK-260930-3vni9d --role developer`, then END YOUR TURN.
