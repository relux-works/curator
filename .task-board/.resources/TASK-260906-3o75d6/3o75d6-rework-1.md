# TASK-260906-3o75d6 rework 1 → revision 2 (orchestrator brief, binding)

Verdict on revision 1: CHANGES_REQUESTED (`TASK-260906-3o75d6_review-verdict-rev1.md`). The restructuring is
right (example moved, import/global rows clean, pins stricter); what is wrong is that the single note LOST content.

1. **Blocking — restore what the flag covers.** The old per-row clause said the flag takes over the files the
   operation would write (§9.5 notice, §8.3 backup) and that without it the write fails with
   `environment_surface_unmanaged_conflict`. The new note (`cli/curator.md:54-60`) copies environments.md only up
   to "on no other operation." — `grep -n unmanaged_conflict cli/curator.md` now finds nothing. Extend the note
   with the following §9.5 sentences (environments.md ≈ 2629-2632: same notice and backup as onboarding when the
   flag is given; without it §8.3 applies and the operation fails with `environment_surface_unmanaged_conflict`),
   and keep `env resolve`'s "`--takeover` applies only with `--repair`" wording in the note or its pointer.
2. **Pin it.** Extend the `takeover_cli_clause` check in `tools/validate.py` so the pin covers the restored
   sentence(s), with a NEGATIVE test that drops them (and one that drops the `--repair`-only wording).
3. **Minor — dangling reference.** "named above as onboarding triggers" (`cli/curator.md:57-58`) points at
   nothing in the CLI guide; say "named in environments section 9.5 as onboarding triggers" and adapt the
   validator's equality accordingly.
4. The reviewer could not run the validator locally (system python lacks jsonschema). Run it yourself through the
   project venv or `uv run --with-requirements requirements-dev.txt …`, including the 1xbrz6 negative tests and two
   mutations (drop a carrier from the note; add `--takeover` to the `profile import` row) — report exit codes.
Continue from the revision-1 tree, append "Revision 2" to the results, hand off.
