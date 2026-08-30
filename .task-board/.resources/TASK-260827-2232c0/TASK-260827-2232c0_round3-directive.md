# Round-3 directive: restructure on the adapter-landed base

The adapter branch landed (merge 2bb54a25). README.md on the story
worktree is now the upstream 659-line version; your previous 116-line
restructure is attached as a resource (our-readme-restructure.md). Redo
the restructure ON TOP of the new upstream README:

1. Start from the current worktree README.md (659 lines). Apply the same
   structural moves as before, now covering the adapter additions:
   reference dumps go to docs/ (compiled-commands.md and ci-gates.md
   already exist in the worktree from your earlier rounds; extend them
   or add a new docs file where the adapter material warrants one, e.g.
   the language/source-closure sections likely belong with
   docs/authoring-language-adapters.md and
   docs/source-closure-adapter-conformance.md which now exist upstream:
   link, do not duplicate).
2. Apply EVERY item of the cycle-2 verdict (rework-instructions
   resource): the ten restored facts stay restored, the Gates reference
   keeps its destination, the shim PATH claim stays per globalbins.go,
   no false LOGBOOK claims.
3. Target: README under 260 lines given the grown upstream; every
   command verified against go run ./cmd/curator ... --help from this
   tree (it now includes the adapter commands).
4. Shell-only edits, grep verification, literal outputs in the outcome.
