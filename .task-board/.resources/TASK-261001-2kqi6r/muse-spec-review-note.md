# Review note — TASK-261001-2kqi6r curator-spec: the muse environment adapter (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev1 (curator-spec, base add50233, tree f72405d2; 57 files, +2908/-13; validators green) against `muse-spec-brief.md` and issue
curator-spec#117. This is on the critical path: task-board will launch Muse as a ROOT session through curator. Verify:
1. The adapter row:
   - XDG_CONFIG/DATA/STATE/CACHE_HOME set to `<home>/{config,data,state,cache}`;
   - HOME is never replaced (MUST);
   - the settings.json and trust.json seeds;
   - the skills and root-context targets marked verified or unverified, honestly.

   It must be consistent with the opencode row's XDG model.
2. Passthrough: `config/muse/auth.json` is a file-link. The refresh probe result is NOT available (the producer says so), so the spec
   must stay safe either way: resolve detects a replaced or forked link (temp+rename) and repairs or refuses; it never silently uses a
   forked credential. Check the 16 link-state × resolve/repair combinations for soundness. The concurrent-refresh locking gap is
   recorded as a known gap.
3. The permission interface in Decision 0018 (exec --yolo; serve --disable-sandbox / --trust-workspace / approvalMode), and the
   foreign-personal-context isolation gap, are recorded accurately, with the normative/informative split right.
4. Schemas:
   - launch-env-fragment-v3.schema.json is a NEW version, and v2 and every released schema are byte-identical (the validate.py
     guard);
   - say why v3 is needed;
   - release/1.0.0-rc.13.json changes only manifest digests.
5. The vectors are generated (environments_muse.go). The validate.py additions and test_muse.py are sound, not overfitted. Run
   tools/validate.py and make regenerate-check yourself, with the real exit codes.
6. No LOGBOOK.md; CHANGELOG under Unreleased; no stray files.
accept_cr, or changes requested with file:line. Never spell any employer name.
