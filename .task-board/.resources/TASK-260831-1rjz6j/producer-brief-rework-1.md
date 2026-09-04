# Producer brief: rework 1 for Decision 0010 draft

## Subject

- Worktree: `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-draft-environments`
- Branch: `draft/agent-environment-profiles` (do not switch branches, do not rebase, do not push)
- Document to edit: `decisions/0010-agent-environment-profiles.md` — the ONLY file you may modify
- Current head: `3fd5617`
- Findings to apply: board resource `review-findings-1.md` on TASK-260831-1rjz6j (13 findings: 3 major, 7 minor, 3 nit). Read it in full before editing.

## Task

Apply ALL 13 findings — majors 1–3 are mandatory; minors 4–10 and nits 11–13 ride the same pass. Follow each finding's "Suggested fix" unless you can justify a strictly better resolution consistent with the document's style; record any deviation and its rationale in the rework report.

Resolution guidance per major:
- **Finding 1 (security overclaim)**: reword to what manager §7 actually delivers; name a new secret-detection detector class over context modules as authorized normative work; prefer additionally making profile installation always-strict as an explicit new rule (state it in Decision 2 and Security impact coherently).
- **Finding 2 (builtin default profile)**: adopt the suggested `local`/builtin source kind with §8 content-hash store keying; touch Decision 8 plus one sentence each in Decisions 4 and 9 so list/sync/status/use are uniform.
- **Finding 3 (IR determinism)**: adopt validation-and-reject (UTF-8, LF-only, exactly one trailing LF per module; reject otherwise), output = modules joined by one empty line, no transformation. Keep the conformance-vector sentence true.

## Constraints

- House style: match the existing document's tone and density; keep English; keep the decisions/ section structure intact.
- Do not renumber existing findings-relevant sections unless a fix requires it; keep the document self-consistent after edits (cross-references, phasing table, open questions).
- Verify claims you change against the repo checkouts when needed: `~/Developer/ReluxWorks/curator-spec` (protocol/core.md, profiles/manager.md), `~/Developer/ReluxWorks/curator` (reference implementation), `~/Developer/ReluxWorks/agent-session-manager-spec` (read-only).
- Use shell tooling for all file operations; confirm the worktree file tree exists before editing.

## Deliverables

1. Edited `decisions/0010-agent-environment-profiles.md` with all findings applied.
2. One signed commit on the branch (`git commit -S`), message in the repo's imperative sentence style, explaining the rework and citing review-findings-1; include the trailer `Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>`. Verify the commit is signed (`git log --show-signature -1`).
3. Board: update resource `decision-0010-agent-environment-profiles.md` on TASK-260831-1rjz6j with the new document content (update_resource, path form). Add resource `rework-report-1.md`: finding-by-finding disposition table (finding # → applied fix summary or justified deviation) + new commit hash.
4. Handoff: set task status to `to-review` and state in the handoff that rework 1 is complete and which findings were applied/deviated.

## Do not

- Do not touch any other file in the worktree or any other repo.
- Do not push, tag, or open PRs.
- Do not mark the task done.
