# TASK-260917-2vapkz — rework 3: split out the skillfile-sources half (THE ONLY CURRENT INSTRUCTION)

PR #116 (rev3 merged onto 2b649c4, plus lockstep commit 8785a12) is green on the Go line. The Python line fails because csk (pinned
4a88aa0e) pins the digest of conformance/skillfile-sources-v1/index.json:
`draft sources digest mismatch for index.json`. csk is not ours. So the skillfile-sources half moves out of this task into
TASK-260930-3ny11n (Story STORY-260930-3kchby), which will land later in lockstep with a csk commit.

The full rev3 content is preserved at `refs/campaign/2vapkz-rev3-full-526a9aa` in this repo, and the new task re-lands the removed part
from there.

Your Story worktree has rev3. Make exactly this change:
1. Restore these to byte-identical 4ad8042b, removing the added ones:
   - everything under conformance/skillfile-sources-v1/;
   - everything under schemas/skillfile-sources-v1/;
   - protocol/skillfile-sources.md;
   - the skillfile-sources manifest_sha256 field in release/1.0.0-rc.13.json. The main-suite manifest_sha256 fields keep the generator's
     value.
2. Remove any normative text in core.md, registry.md or the CHANGELOG that names skillfile-sources carriers (skillfile-lock-v2,
   install-marker-v6, source-audit-v2). Replace it with one sentence saying the skillfile-sources carriers adopt hash_version in a
   follow-up (name TASK-260930-3ny11n). Everything else from rev3 stays byte-identical.
3. Regenerate and run `tools/validate.py` and `make regenerate-check`, recording the real exit codes. Then show
   `git diff --stat 4ad8042b -- conformance/skillfile-sources-v1 schemas/skillfile-sources-v1 protocol/skillfile-sources.md` is empty.
4. In the results, give the new conformance/v1 manifest sha256. Curator must pin it next.

No LOGBOOK.md. Never spell any employer name. Set status development, update the results, run
`task-board handoff TASK-260917-2vapkz --role developer`, then END YOUR TURN.
