# Review note — TASK-260924-1ytz0u launcher v0.1.0 release prep, CR revision 1 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Small release-prep change (CHANGELOG, README, cmd/curator-run/main.go + test + help golden). Verify in a disposable clone:
1. CHANGELOG has a `0.1.0` release section (no "unreleased"/"no tag is created" wording left for 0.1.0); README install instructions
   name `go install github.com/relux-works/curator-agent-launcher/cmd/curator-run@v0.1.0` and are correct for the module path.
2. The binary reports 0.1.0 when built from the v0.1.0 tag (ldflags or constant — check the mechanism works for `go install @tag`,
   where ldflags are NOT applied: a `debug.ReadBuildInfo` module version or a constant must yield 0.1.0) and still reports the SPEC
   version; the help golden matches.
3. The results justify the decision on a release workflow (binaries or not) for M0 onboarding.
4. `make check` evidence + the clean-cache install evidence; validation log green. No stray artifacts/task documents in the tree.
accept_cr or changes requested with file:line. No LOGBOOK.md.
