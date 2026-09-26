# TASK-260924-1ytz0u — prepare curator-agent-launcher v0.1.0 (THE ONLY CURRENT INSTRUCTION)

Control root: /Users/administrator/Developer/ReluxWorks/curator/curator-agent-launcher (main = 5d0d0a3, includes the 0018 permission
interface on skill-agents-management v0.5.22). Your Story worktree only. Read `campaign-producer-rules.md`. You do NOT tag — the
orchestrator cuts the signed tag `v0.1.0` after your change lands.
Purpose (operator, roadmap M0): onboard the development Mac from TAGGED releases. This repo has no tags and no release workflow yet.
1. Decide with evidence what a v0.1.0 release must contain here: at minimum a CHANGELOG release section for 0.1.0 (the unreleased
   entries), a version string the binary reports (`curator-run --version` or equivalent — add if absent, set by ldflags or a
   constant consistent with the tag), SPEC/README version references consistent (SPEC stays 0.5.0-draft unless the repo convention says
   the release pins it), and `go install github.com/relux-works/curator-agent-launcher/cmd/curator-run@v0.1.0` working (module path,
   no replace directives, no local-only deps). Check how the onboarding docs (README) tell a user to install; make them name v0.1.0.
2. Add a GitHub release workflow ONLY if onboarding needs prebuilt binaries (state why or why not; if added, mirror curator's
   release.yml conventions: tag-triggered, verify-tag, checks green).
3. Tests stay green (`make check`), CHANGELOG. Attach `TASK-260924-1ytz0u_results.md` (what a release contains and why, the exact
   install command, verification you ran), check DoD, `task-board handoff TASK-260924-1ytz0u --role developer`. A
   `run_wrote_outside_worktree … policy warn` block is a warning — verify status `to-review`.
