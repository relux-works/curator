# Review note — TASK-260924-1nh93t F-M1e revision 4 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Revision 2 was CHANGES_REQUESTED only for the missing SPEC §4.3 `mapped=` API. Revision 3's content failed validation only on a
host-load package timeout; revision 4 = same content, green. Verify in a disposable clone:
1. Public release-versioned permission-mapping API (e.g. `Registry.PermissionMapping(system, toolRelease, mode)`): native → none;
   yolo → the module-owned flag for that verified release; unknown system / unverified release / unsupported mode → existing typed
   errors; callable before admission; reuses the same verified-release rows as the grammar (no second table).
2. The audit now has the §4.3 row and states the result of one more full pass over §4 ("none further" or new rows with APIs).
   Do your OWN final pass over launcher SPEC 0.5.0-draft §4 for any "supplied by agents-management" / module-returned value —
   this is the last chance before F-L1b resumes.
3. Rows per system × mode × verified/unverified release; one mutant re-applied by you (flag for native, or for an unverified
   release) → killed; classifier and everything verified in rev2 unchanged (diff rev2→rev4 patches).
4. Spelling guard green; README + the same unreleased CHANGELOG bullet extended; validation log green.
accept_cr on revision 4, or changes requested with file:line. No LOGBOOK.md.
