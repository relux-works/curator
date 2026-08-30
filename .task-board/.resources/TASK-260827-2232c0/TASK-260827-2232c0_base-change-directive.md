# Base change directive (round 2)

The trunk moved: origin/main now carries a 429-line README (the old base
was 408) and docs/build-https.md. The orchestrator rebased the story
worktree onto origin/main and re-applied your four products on top; your
restructured README.md currently OVERWRITES the upstream additions.

Do, in this order:
1. Read the attached readme-upstream-delta.diff. Integrate every
   addition it shows into your restructured README (the operator
   credentials bullet with both docs links, and the suite-consumption
   material: decide its home per the spec: the CI/gates destination
   document you create for the deleted Gates section is the natural
   place; README keeps one line plus a link).
2. Apply the cycle-1 verdict in full (rework-instructions resource):
   restore the ten dropped facts into docs/compiled-commands.md, give
   the 62-line Gates and tooling reference a destination document, fix
   the global-shim PATH claim per internal/globalbins/globalbins.go,
   correct the false LOGBOOK claim.
3. Verify every command against go run ./cmd/curator ... --help from
   the CURRENT tree (it now contains build-https).
