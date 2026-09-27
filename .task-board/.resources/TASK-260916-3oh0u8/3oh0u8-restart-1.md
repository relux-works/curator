# TASK-260916-3oh0u8 — restart on current trunk (THE ONLY CURRENT INSTRUCTION, with 3oh0u8-sec-brief.md)

The old Story workspace sat on 1de6f8e1 (123 commits behind trunk) and was snapshotted then discarded. Your Story worktree is now fresh on
trunk (≥ d41da0fb: E2 direct-only system modules and 1wc76r just landed). The previous work — CR rev1 (changes_requested) and the full
snapshot `refs/campaign/2otjbn-full-20260927` (15 files: cmd/curator/umbrella*.go, envstatus.go, internal/config/environments*.go,
internal/envprofile/status.go, provider_schema_conformance_test.go …) — is REFERENCE ONLY: read `git diff refs/campaign/2otjbn-full-20260927^
refs/campaign/2otjbn-full-20260927` and the rev1 review verdict, then re-implement against current trunk (many of those files moved a lot; never
apply it blindly, never revert trunk code). Spec: rc.13 provider resolution trust roots (TASK-260916-1x0ogh landed the spec side) — umbrella
provider lookup restricted to the declared trust roots, refused fail-closed outside them; cite clauses; drive vectors; gap rows leave the ledger.
`task-board m 'set_status(TASK-260916-3oh0u8, status=development)'` first. New manager-state reads via internal/stateread (guard test).
Before handoff: the hosted gate on your published revision must be green (see "Hosted gate before handoff" in campaign-producer-rules.md);
VERIFY `git diff --name-only origin/main -- . ':!.task-board'` lists only your paths. Write only inside your Story worktree. No CHANGELOG/LOGBOOK.
