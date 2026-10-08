# Review note — base-refresh revision (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

The previous revision of this task was ACCEPTED on content; its integrate refused only because trunk moved
on a path it also changes. The new revision is a BASE REFRESH. Review the refresh only:
1. Diff the new revision's patch against the last ACCEPTED revision's patch (both are resources on this task): every
   non-merged path must be byte-identical (per-file `git patch-id --stable` or a two-way diff); the merged paths
   (CHANGELOG.md, ledgers such as `.github/ci/platform-cases.tsv`, workflow files) must keep BOTH sides — trunk's new
   entries/rows AND this task's — with nothing dropped or duplicated.
2. The runtime's validation log for the new revision is green.
Anything beyond a faithful combination is a finding. Record exactly one verdict (accept_cr on the new revision, or
changes requested with file:line). No LOGBOOK.md.
