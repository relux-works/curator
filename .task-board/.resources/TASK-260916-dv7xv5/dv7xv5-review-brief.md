# Review brief — TASK-260916-dv7xv5, review round 2 (rev2)

You are the independent reviewer of a **research** task, round 2. Round 1
(`TASK-260916-dv7xv5_review-verdict-rev1.md`, attached as an outcome) returned
rework. The producer answered with **`verify-e-findings-rev2.md`** (outcome
resource) and updated the seven sibling story descriptions. Review **rev2**;
`verify-e-findings.md` (rev1) stays attached only as history.

The findings E1–E6 (+ the "Minor" residuals the task calls E7) are defined in
the precondition resource `security-audit-2026-09-spec-supplement.md`.

## What to verify (all read-only; change no code, run no tests)

1. **Each rev1 finding is closed in rev2.** Check, one by one, that rev2:
   - cites the correct E1 entry point (`cmdProfileUpdate` → `UpdateWithPolicy`
     in `cmd/curator/profile.go`, not the import path) and bounds the signer
     claim to the context/profile resolution path;
   - replaces every former "returns nothing" assertion (E1, E2, E3) with the
     exact command, its real output and a disposition of every match
     (registry signature envelopes, build-repo signer policy, swiftpm
     `cat-file`, `internal/mcp/mcp.go` `mcp_servers`, `closure.go`,
     `gitsource.go` …) — re-run the quoted commands yourself at the pinned
     revisions and compare;
   - gives E7 an explicit strict-MCP asymmetry verdict with citations
     (`internal/envregistry/envregistry.go:192`, `:215`, `:219`;
     `internal/envprofile/managed.go` seed path);
   - states the E5 limit (remove-then-write is not `O_NOFOLLOW`/atomic;
     `writeStoreDocument` still uses `WriteFile`) and the E6 split
     (path-kind MCP dependency not applicable; system-module admission and
     directory boundary bounded to the inspected path-ingestion flow).
2. **Evidence check per finding (E1–E7).** For each row of the rev2 table open
   the cited `file:line` sites at the pinned revisions
   (`git -C ~/Developer/ReluxWorks/curator/curator show 80483355:<path>`,
   `git -C ~/Developer/ReluxWorks/curator/curator-agent-launcher show b34e1e27:<path>`;
   read through `git show`, do not check anything out and do not modify the
   trees). Confirm the cited code says what the row claims and that the verdict
   follows from it. Flag any verdict stronger or weaker than its evidence.
3. **Acceptance criteria.** (a) the outcome carries the per-finding table with
   both pins and `file:line` evidence; (b) the READMEs of
   `STORY-260916-ioemse` (E1), `-2d9coh` (E2), `-1i1gfo` (E3), `-2otjbn` (E4),
   `-73a5zg` (E5), `-wgt8vz` (E6), `-33vuzm` (E7) under `EPIC-260910-2hw1xb`
   now record the verified verdict, the pins and the narrowed scope (read them
   with `task-board q 'get(STORY-…) { overview }'` or the README files);
   (c) no code or test changes anywhere — research only.
4. **Scope discipline.** Do not start or suggest starting any remediation
   story; remediation is a separate goal.

## Checklist

The task checklist mirrors the acceptance criteria plus the role baseline. You
check an item (`task-board m 'check_item(TASK-260916-dv7xv5, item=N)'`) only
once you verified it yourself; "Tests green" is not applicable to this
read-only research task — say so in the verdict instead of running suites.

## Verdict

Record the verdict as a task-scoped outcome resource named
`TASK-260916-dv7xv5_review-verdict-rev2.md`: per-rev1-finding closure status,
per-finding evidence check (agree / disagree + why, with file:line), the
acceptance-criteria check, and the final verdict — `accept` or `rework` with
the concrete findings the producer must address. Then hand off with the
reviewer role's normal transition (`done` on accept; `to-dev`/`analysis` on
rework). Do not set `blocked` unless a human-only decision is genuinely
required.

Budget: a bounded read-only check of one 7-row table, the rev1 findings and
seven READMEs; finish in one pass.
