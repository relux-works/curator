# Finalization run (previous run wrote the documents but did not hand off)

docs/cli.md (739 lines) and docs/troubleshooting.md (321 lines) already
exist in the story worktree. Do NOT rewrite them. Within 15 minutes:
1. Verify both files intact; verify the README Commands section exists
   (grep). If the Commands section is missing, add it per the task
   description (collapsible groups + docs/cli.md link) - that is the
   only writing allowed.
2. Spot-check three synopses against go run ./cmd/curator ... --help
   and two error strings against internal/, literal outputs into the
   outcome resource.
3. Register the CR revision, complete the checklist, hand off.
