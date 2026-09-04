# TASK-260901-32j97g: bootstrap-launcher-repo-and-spec

## Description
Bootstrap the curator-agent-launcher repository locally: compiling Go skeleton (cmd/curator-run stub), Apache-2.0 licensing, and the in-repo SPEC.md draft covering CLI surface, three-plane composition algorithm, ax always-when-configured, system-prompt opt-in warnings, diagnostics, and versioning. Remote already created empty; orchestrator pushes after review.

## Scope
(define task scope)

## Acceptance Criteria
Local repo at ~/Developer/ReluxWorks/curator-agent-launcher on main; go build/vet green; SPEC.md covers scope/non-goals, flags, composition, ax handoff, system-prompt, diagnostics, versioning; LICENSE+NOTICE+README+Makefile+.gitignore present; signed commits; notes resource; handoff to-review.
