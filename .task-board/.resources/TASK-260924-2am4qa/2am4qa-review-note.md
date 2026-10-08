# Review note — TASK-260924-2am4qa core §4.4 `directory` amendment, CR revision 1 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review against `2am4qa-brief.md` and the operator memo `skillfile-operator-memo-20260924.md`; disposable clone of curator-spec.
0. BLOCKING by rule: the candidate contains the root file `TASK-260924-2am4qa_results.md` — task documents are board resources, never
   repository files. Record it as a finding (removal required) and continue the full review so the rework fixes everything at once.
1. core §4.4: `dependencies.skills[].directory` optional; grammar/containment exactly the Skillfile schema 2 individual-selector
   `directory` rules (referenced, not restated differently); absent = repository root, unchanged meaning.
2. New manifest schema revisions (agent-skill-v9 / csk-skill-v9): justified as a new revision rather than an edit of a frozen schema;
   frozen v1 schemas untouched; placement in `schemas/draft-sources-v1/` — note that another author is preparing the promotion of that
   namespace to released; judge whether this placement merges cleanly with that move (flag, don't block).
3. Identity/closure/lock/audit record the directory; diamond (two folders of one repo at one ref) and same-folder-twice rules normative.
4. Vectors: valid (absent/root/subfolder) and invalid (absolute/backslash/empty/escape/glob/parent) for both manifests, plus the
   semantic `manifest-dependency-directories.json` (missing folder, no SKILL.md, diamond) — each validated; regenerate-check clean.
5. Validation evidence green (runtime log + results table).
accept_cr or changes requested with file:line. No LOGBOOK.md.
