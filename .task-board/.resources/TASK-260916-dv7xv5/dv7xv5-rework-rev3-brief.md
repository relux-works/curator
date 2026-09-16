# Rework brief — TASK-260916-dv7xv5, rev3 (answers review verdict rev2)

Role: researcher (read-only research; no code, no tests, no remediation).
Task class: research. Task is in `analysis`; you end with
`task-board handoff TASK-260916-dv7xv5 --role researcher` at `to-review`.

## Plan contract
- **Decision this unblocks:** acceptance of the E1–E7 implementation
  verification so the remediation stories (`STORY-260916-*` under
  `EPIC-260910-2hw1xb`) start from accurate `file:line` evidence.
- **Inputs:** `verify-e-findings-rev2.md` (current artifact),
  `TASK-260916-dv7xv5_review-verdict-rev2.md` (the findings you answer; it
  carries independent literal grep transcripts you may reuse verbatim),
  `security-audit-2026-09-spec-supplement.md` (E1–E7 definitions).
- **Budget:** one pass; four corrections; no new investigation beyond
  confirming each cited line with `git show` at the pins.
- **Exit criteria:** every "Required producer correction" of verdict rev2 is
  applied and visibly answered; a reviewer can re-run each quoted command and
  get the literal output you show.
- **Consumer:** review round 3, then the E-series stories.

## Pins and method
Read the sources only through
`git -C ~/Developer/ReluxWorks/curator/curator show 80483355:<path>` and
`git -C ~/Developer/ReluxWorks/curator/curator-agent-launcher show b34e1e27:<path>`
(`| nl -ba` for line numbers). Do not check anything out, do not modify either
tree, do not run tests.

## Corrections to apply (from verdict rev2, "Required producer corrections")
1. **E1:** `cmdProfileUpdate` starts at `cmd/curator/profile.go:252` (not
   `:260`); range selection is `internal/contextresolve/contextresolve.go:554-575`
   (include `:489` for candidate ordering); `:511-515` is the exact-commit
   branch — drop it from the range claim. Propagate the corrected citations to
   the `STORY-260916-ioemse` README description.
2. **E2:** replace "declarations only" with the literal four-line output of
   `grep -rn 'context_system_module_transitive\|SystemModules' internal --include='*.go' | grep -v _test.go`
   and classify `internal/envprofile/envprofile.go:1138-1139` as the
   production caller that surfaces `SystemModules` as warnings (warning
   surfacing, not admission). Verdict stays confirmed.
3. **E3:** the seed writes are `internal/envprofile/managed.go:887-893` (loop
   over `seeds.files`, `WriteFile` with the payload); `:378-388` renders
   root-context documents and proves nothing about seeds — replace it. Include
   the literal 15-line `mcp_servers` grep output and disposition
   `internal/skillspec/types.go:98` (McpServer type comment). Propagate the
   seed-write citation to the `STORY-260916-1i1gfo` README description.
4. Everywhere a command is quoted, show the literal output block (command,
   output, `exit=`), not a paraphrase. Keep the E5/E6 limits and the E7
   strict-MCP asymmetry exactly as rev2 states them; keep E4–E7 verdicts.

## Deliverables
- New outcome resource **`verify-e-findings-rev3.md`** (attach with
  `task-board resource add TASK-260916-dv7xv5 <file> --type outcome --name verify-e-findings-rev3.md -d "…"`),
  headed by a short "Rev3 answers verdict rev2" paragraph listing the four
  corrections; keep the full per-finding table and the sibling-consequences
  section. Leave rev1/rev2 attached as history.
- Sibling README updates for `STORY-260916-ioemse` (E1) and
  `STORY-260916-1i1gfo` (E3) through `task-board m 'set_details(STORY-…, description="…")'`
  (read the current description first with `task-board q 'get(STORY-…) { overview }'`
  and change only the citations; never edit `.task-board/` files directly).
- Checklist: leave items as they are; the reviewer checks them. Then hand off.
