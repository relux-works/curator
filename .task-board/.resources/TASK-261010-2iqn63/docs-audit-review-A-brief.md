# THE ONLY CURRENT INSTRUCTION — reviewer A of two (R138), RECORD-ONLY, for Change Request rev1 of TASK-261010-2iqn63 (read-only)

The study is a cross-repository documentation and diagram audit: book `book/swarma/ru` against the platform module specs, curator-spec CIPs and the launcher. The brief is the precondition `platform-docs-audit-brief.md`. The study is `.research/261010_platform-docs-audit.md`, the same as the outcome resource. The orchestrator will turn its repair plan (section D) into producer tasks, so the review must establish that its findings are real.

Check, by fetching the cited sources read-only at their pinned commits (`gh api repos/OWNER/REPO/contents/PATH?ref=COMMIT`):
1. **CONFLICTING rows.** Take at least 12, spread across chapters 02–11. For each, the two cited texts really disagree in the way the row says. Flag any row where the "conflict" is only a phase difference that the texts already label, or a misreading.
2. **MISSING and ARCH-ONLY rows.** Take at least 6. The gap is real: no cited or obvious defining source covers the clause.
3. **Diagrams.** Take at least 5 claims from B1–B4. Check the existing files, the "19 of 26 book figures have no local source" claim, and that each proposed module diagram is not already present.
4. **Repair plan D1/D2.** The P0/P1/P2 order and dependencies follow from the findings. No item asks a producer to invent product behaviour. Every item that needs an owner decision is labelled as such.
5. **Scope and hygiene.** The delta is only the research file, with no `LOGBOOK.md` change, no secrets, no personal or local paths and no session links.
6. **Readability.** An engineer can act on each D item without re-reading the whole study.

No tests or builds on this host. RECORD-ONLY: attach your verdict as the resource `docs-audit-review-A.md` on TASK-261010-2iqn63. Use the usual verdict structure with a findings block; `severity` is one of `bypass|regression|robustness|note`, and the P-level goes in `severity_reason`. Do NOT call accept_cr or reject_cr: reviewer B (cross-provider) decides. Then END YOUR TURN.
