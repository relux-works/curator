# TASK-260924-1fnikm — review external acceptance PR relux-works/curator-spec#96 (THE ONLY CURRENT INSTRUCTION)

The cocoaskills orchestrator authored the acceptance PR (Decision 0022): head 5746367 (signed ivanopcode) on main 7b06bb3. We review and
release; we do NOT push to that branch. Work in a disposable clone ($TMPDIR), never the control root. Produce a review report resource
`TASK-260924-1fnikm_pr96-review.md` with a verdict APPROVE / CHANGES-REQUESTED and file:line findings. Binding criteria (operator memo
`skillfile-operator-memo-20260924.md`, handoff `skillfile-default-on-handoff-20260924.md`, addendum `skillfile-lock-replay-addendum-20260924.md`):
1. Corpus in its own directory and pin (`conformance/skillfile-sources-v1`, `schemas/skillfile-sources-v1`), not merged into conformance/v1;
   frozen `schemas/v1/*.schema.json` untouched (only README prose may change).
2. rc.10 independence: every `$ref` from schemas/skillfile-sources-v1 into schemas/v1 resolves to a definition byte-identical at tag
   v1.0.0-rc.10; no prose dependency on post-rc.10 core/registry/manager/environments clauses. Show the check (script + output).
3. Content equals what was reviewed/accepted: the move from draft-sources-v1 is a rename (per-file blob identity or patch-id) except the
   named additions — list every non-rename change. #90 (2cp9w9) content present unchanged. v9 manifest material ABSENT.
4. Global-scope rule (project scope only; global stays schema 1) with positive and negative vectors.
5. §3 lock-replay amendment: missing snapshot re-materialized from the declared source (exact locked commit by object id for
   git/repository; current bytes for path); accept only on identity + content_sha256 match else `source_snapshot_changed`;
   `source_snapshot_unavailable` only for an unreachable source; lock byte-identical; no re-resolution; five vectors.
   Declared tag NOT verified during replay — repository-transport scopes exact-tag verification to resolution and refresh; check the text
   is consistent everywhere (no remaining sentence requiring tag verification at install-from-lock).
6. Curator conformance: Curator's replay (internal/install/draftsources.go replayGitDraftSnapshot / fetchLockedCommitFromDeclaredSource on
   curator Story STORY-260924-3eywt2, landing now) fetches the locked commit by object id, verifies repository identity and content_sha256,
   does no tag check — confirm, and state whether it verifies the object FORMAT (sha1/sha256) as the new text requires; if not, name it.
   Run curator's crossconformance suite against the PR corpus in a disposable curator clone and list every case curator fails, grouped by
   owner (known: #90 rows → TASK-260924-20o9dk; replay/global vectors → new rows).
7. Release notes/CHANGELOG/COMPATIBILITY/decision record coherent; regenerate/validate recipes green in your clone (bounded).
No LOGBOOK.md. Attach the report, check DoD, `task-board handoff TASK-260924-1fnikm --role developer`.
