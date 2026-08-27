## Status
done

## Review
required

## Task Class
docs

## Estimate
estimated(fibonacci(3))

## Blocked By
- TASK-260822-b0wg3a

## Blocks
- TASK-260823-qwr5w9

## Checklist
- [x] Config grammar, precedence, three canonical selection shapes documented citing curator-spec
- [x] CLI subcommand and precheck/candidates behaviour documented; examples match actual output
- [x] Links/lint green
- [x] Document that an agent-less --identity pointing at a *.pub file cannot authenticate (admission.go emits IdentityAgent=none -i <path>); the sound discovered shape is --agent --identity <key>.pub. From TASK-260822-b0wg3a review finding 2.
- [x] Note that a repository namespace can contain characters the build_ssh scope grammar rejects (PortableComponent vs scopeSegmentRE), in which case the suggested default scope must be widened to the host. From TASK-260822-b0wg3a review finding 3.
- [x] Code written per task description and AC
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
Producer note carried over from install-precheck-and-candidates: the interactive discovery lists *.pub files; the agent-pinned default entry is correct, but the agent-less identity-file menu entry with a .pub path yields a selection that cannot authenticate (no private key material is validated). Document the three selection shapes accordingly and state that an identity-file-only selection needs the private key path; whether discovery should offer private-key paths is a product decision, cite the spec canonical tails. Patches for the whole surface live as artifacts on the four sibling done tasks; the CLI/help texts to document are in their patches, not on origin/main yet.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260822-e767cb, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260822-e767cb)
Docs delta only: docs/build-ssh.md (new, 494 lines) + one README bullet. Branch task/TASK-260822-4p3dcq-docs-build-ssh in .temp/TASK-260822-4p3dcq/worktree, off origin/main 6a9b201, carrying the accepted 96m5pj+2505vo+3pkc80+b0wg3a chain as the staged baseline.

Checklist item 5 is documented WITH A CORRECTION, verified at first hand. b0wg3a review finding 3 cited identifiers.PortableComponent, which governs skill sources via internal/identity. External build repositories parse through buildrepo.ParseSource, and for SSH it restricts the raw path to ^[A-Za-z0-9._/-]+$ (Spec 6.3) - exactly the alphabet scopeSegmentRE admits - so ssh://git.example.com/team+infra/app.git is refused at parse time and the PATH half of the divergence cannot occur on the only transport build_ssh selects for. What does reproduce is the HOST: hostRE admits git.example.com., git..example.com, git-.example.com, which scopeHostRE rejects, and widening the suggestion to the host does NOT help because the host is not a valid scope either. Such a repository can only be covered by --build-ssh-* flags or CURATOR_BUILD_SSH_*. The page documents that, plus the HTTPS contrast (git.example.com/team+infra/app does carry a rejected namespace but never reaches the scope suggestion).

Spec drift flagged in the page: profiles/manager.md 11.3 on curator-spec main admits only TWO authentication tails. The pinned-agent third tail this repo accepted in 6a9b201 is on the unmerged spec branch spec/pinned-agent-authentication-tail (38232d3); CI SPEC_PIN 00b1688 also predates it. The corpus does not encode the tail, so no pin or checksum is affected.

Every transcript in the page came out of the implementation: CLI output from a binary built in the worktree, the JSON block is the scratch config verbatim, the seven parse errors from seven malformed configs, and the prompt/diagnostic/provenance text from a throwaway docsample_test.go driving resolveBuildSSH and InteractiveBuildSSHResolver, since deleted. Raw capture attached.

Gates (standalone, real exit codes): gofmt -l . 0 empty; go build ./... 0; go vet ./... 0; golangci-lint run ./... 0 (0 issues.); naming gate clean (one README line); em dash/guillemet grep no match; link+anchor pass 13 targets and 10 anchors, 0 misses. Note: a fresh git worktree has no submodule, so go vet exits 1 on the missing tuitestkit replacement directory until git submodule update --init --recursive is run.

NOT RUN, stated plainly: the repository has no markdown linter, no link checker and no docs build step (no make docs target, no docs job in ci.yml), so "docs build/lint green" was met by the gates that exist and apply to prose. Full go test ./... was not run for a two-markdown-file delta; a scoped go test over internal/config, internal/install/..., internal/buildrepo and cmd/curator was run instead, and the chain full suite was already green at b0wg3a RUN-260822-fad8ab.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-e767cb, pid=98602, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [reviewer] reviewer (claude) (run=RUN-260822-3dc3b7, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260822-3dc3b7)
REVIEW VERDICT: ACCEPTED (RUN-260822-3dc3b7, reviewer, claude). Evidence: TASK-260822-4p3dcq_review.md + TASK-260822-4p3dcq_review-gotest.log.

Verification was independent, not trust. Built curator from the producer worktree and re-executed every transcript in docs/build-ssh.md against a fresh scratch CURATOR_CONFIG: four add lines, the replaced-vs-added distinction, the sorted tab-separated list, the empty-listing stderr line at exit 0, both remove cases, both refusals at exit 2, the three -build-ssh-* help lines from install -h, and all seven fail-closed parse messages regenerated from my own malformed configs. Byte-for-byte matches including exit codes. The build_ssh JSON block is exactly the scratch config after the four adds. The prompt transcript, both diagnostics and the provenance line match the attached docsamples capture and each was cross-checked against its generator in source (missingBuildSSHError, buildSSHAddCommands, buildsshprompt.go, buildSSHProvenance).

Twenty-one behavioural claims checked against source, all hold: the three tails at admission.go:150/155/168; precedence and the field-by-field flags-first merge; Selected() ignoring known_hosts so a bare --build-ssh-known-hosts is not a selection; auto failing closed without SSH_AUTH_SOCK; the three-step known-hosts order; the exact build_repository_identity_invalid Error() string; longest-scope-wins and non-canonical matching nothing; the 4096 scalar cap; sorted fault reporting; platform-independent Windows forms; the prompt gated on both TTYs and not dry-run; the prompt and CLI sharing config.SetBuildSSH; Sscanf leniency and re-ask on unparsable; the rejected-default scope re-ask loop; provenance only on dry run; and capture-before-loop at main.go:532 proving limit 3.

All five spec citations read at first hand in curator-spec. The drift note is correct: profiles/manager.md 11.3 on main says the tail is exactly either of TWO forms; the third pinned-agent tail exists only on spec/pinned-agent-authentication-tail head 38232d3. Documenting that rather than presenting three tails as settled protocol is the right call.

Checklist item 5 correction re-derived independently and confirmed. PortableComponent governs skill sources; build repos gate SSH raw paths through validRepositoryPath/sshPathRE (buildrepo.go:124,151), exactly the scopeSegmentRE alphabet, so the path half cannot diverge. The host half does: hostRE (buildrepo.go:36) admits git.example.com., git..example.com, git-.example.com, which scopeHostRE (buildssh.go:49) rejects, and widening to the host does not rescue it. Page documents the real behaviour including the interactive re-ask loop and the HTTPS contrast.

Gates re-run by me in the worktree: go build ./... 0; gofmt -l . clean; golangci-lint run ./... 0 issues.; naming gate clean (zero outside README, exactly one README line); em dash/guillemet rule (CONTRIBUTING.md:29, implementation-plan.md:44) satisfied by the delta; own link/anchor script over four markdown files 0 misses including all ten TOC anchors; go test -count=1 -timeout 30m over config, install/..., buildrepo, cmd/curator all ok, EXIT=0.

Docs build/lint green assessed honestly: the repo really has no markdown linter, link checker or docs build step (two files under docs/, no index, no make docs, no docs job in ci.yml). The producer said so instead of claiming a gate that does not exist. AC met by the gates that exist and apply to prose, all re-run.

Non-blocking observations recorded in the artifact, none charged as rework: (1) list on a host with no config file at all exits 1 with global config not found, not the documented empty-listing line; (2) defaultBuildSSHScope returns the full identity for a two-segment host/repo; (3) implementation asymmetry worth a follow-up on the CODE not this page - run-wide agent auto fails with the stable build_repository_ssh_credential_missing code but a configured agent:true scope fails through scopeCredentials with a bare uncoded message; (4) CHANGELOG.md Unreleased is empty and this surface lands without an entry, but the file has been untouched since 0.12.5 while schema 7, seamless manager and the pinned-agent tail all landed without entries, there is no CI check and CONTRIBUTING does not require it - a story-level decision; (5) the dry-run provenance example omits the alias prefix credentialReport prepends, described in prose instead of fabricated.

Reviewer-archetype run: no commit_ack supplied. This is the acceptance evidence for the commit-owning mover, which commits docs/build-ssh.md plus the README bullet and then makes the final done transition with commit_ack=scope_committed.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-3dc3b7, pid=16566, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260822-4p3dcq_spawn-log_-implementer--developer--claude-_RUN-260822-e767cb.log](file://TASK-260822-4p3dcq/TASK-260822-4p3dcq_spawn-log_-implementer--developer--claude-_RUN-260822-e767cb.log) — System spawn log captured by task-board
- [TASK-260822-4p3dcq_results.md](file://TASK-260822-4p3dcq/TASK-260822-4p3dcq_results.md) — Implementation notes, spec citations, the finding-3 correction with evidence, example provenance, and the gate table
- [TASK-260822-4p3dcq_docsamples.log](file://TASK-260822-4p3dcq/TASK-260822-4p3dcq_docsamples.log) — Raw capture of the prompt transcript, both fail-closed diagnostics and the provenance line, produced from the implementation
- [TASK-260822-4p3dcq_docs-only.patch](file://TASK-260822-4p3dcq/TASK-260822-4p3dcq_docs-only.patch) — The delta charged to this task: docs/build-ssh.md plus one README bullet, against origin/main 6a9b201
- [TASK-260822-4p3dcq_final.patch](file://TASK-260822-4p3dcq/TASK-260822-4p3dcq_final.patch) — Full branch state: the accepted 96m5pj+2505vo+3pkc80+b0wg3a chain plus this task's docs, against origin/main 6a9b201
- [TASK-260822-4p3dcq_spawn-log_-reviewer--reviewer--claude-_RUN-260822-3dc3b7.log](file://TASK-260822-4p3dcq/TASK-260822-4p3dcq_spawn-log_-reviewer--reviewer--claude-_RUN-260822-3dc3b7.log) — System spawn log captured by task-board
- [TASK-260822-4p3dcq_review.md](file://TASK-260822-4p3dcq/TASK-260822-4p3dcq_review.md) — Reviewer verdict: accepted. Independent re-execution of every CLI transcript, spec citations verified in curator-spec, gates re-run, scoped test suite green.
- [TASK-260822-4p3dcq_review-gotest.log](file://TASK-260822-4p3dcq/TASK-260822-4p3dcq_review-gotest.log) — Reviewer-run scoped test suite: config, install, install/atomicity, buildrepo, cmd/curator all ok, exit 0

## Created
2026-08-22T16:12:06Z

## Last Update
2026-08-22T22:33:57Z

## Assigned To
[reviewer] reviewer (claude)
