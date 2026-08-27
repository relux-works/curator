# Finalization run (previous run timed out after writing the document)

docs/build-https.md (285 lines) already exists in the story worktree
with an outcome resource attached. Do NOT rewrite it. Within 15 minutes:
1. Verify the document intact and the build-ssh cross-link present
   (grep both files).
2. Spot-check three claims against internal/ with grep (one credential
   source, one env variable, one error identifier) and append the
   literal outputs to the outcome resource.
3. Register the change request revision if the CR plane requires it,
   complete the checklist, and hand off with task-board handoff.
