# Review note — TASK-260930-3vni9d carrier: content-hash v2 minus skillfile-sources (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Delta review. The content of TASK-260917-2vapkz rev3 was accepted over three rounds; full commit refs/campaign/2vapkz-rev3-full-526a9aa.
The carrier CR is at base 2b649c4d, tree 6c02b8b4, with validators and regenerate-check green. Review ONLY the difference from
526a9aa0: `git diff 526a9aa <carrier tree>`. The expected difference is:
- conformance/skillfile-sources-v1, schemas/skillfile-sources-v1 and protocol/skillfile-sources.md are byte-identical to 2b649c4. The
  orchestrator already saw an empty `git diff --stat 2b649c4d 6c02b8b4` on those paths.
- In release/1.0.0-rc.13.json only the skillfile-sources manifest_sha256 went back to its 2b649c4 value.
- Text edits in CHANGELOG.md, protocol/core.md, protocol/repository-transport.md and schemas/v1/README.md: every mention of the
  skillfile-sources v2 carriers is replaced by the one-sentence follow-up reference (TASK-260930-3ny11n), and nothing else changed.
  Show each hunk.

Confirm that no main-suite content changed versus rev3: every conformance/v1, schemas/v1 and other protocol file is byte-identical to
526a9aa0 except the listed text edits. Run tools/validate.py and make regenerate-check yourself, with the real exit codes.
accept_cr, or changes requested with file:line. Never spell any employer name.
