# Contract table, deciding delta review B, CR rev2 — verdict: CHANGES REQUESTED

Task: TASK-261010-1ypla7. Candidate: CR-TASK-261010-1ypla7-2 rev 2, base `67f736f5e45f562df892674cb53c97b0e6d732e2`, tree `3a9b0d864af6b1398d07759277d90597344050f0`. One file changed: `.research/261010_platform-contract-table.md` (76,851 bytes; 81,920 allowed). The working-tree blob equals the candidate tree blob (`69c65e61…`).

Order followed: the rev1→rev2 delta was diffed against the rev1 patch and checked against the pinned sources first (9 files re-fetched read-only with `gh api …?ref=<full commit>`, all exit 0: UM, SH, SG, AR, B07, B12, CB, DP, LR). The verdict below was formed before `contract-table-delta-A.md` was read; A was then reconciled.

## 1. Rev1 findings F1 / F2 and the P2 items

| Item | Result |
|---|---|
| F1, P7 retirement fate | **Present and correct.** Request enum `archive`/`delete` (UM L121), `keep` only in §4.4.1 prose (L143 `keep_requires`, L147), role table (L55–63) and defaults/bounds (L132–150) match. Marked OWNER DECISION NEEDED with two real options (admit `keep` as an operator-only request value, or drop it from the policy). No decision invented. |
| F2, P9 gate lanes (A07.8) | **Present and correct.** Lane table SG L3–20 and SH §2.4–2.6 (L69–97) match; B07 L119 is the conflicting text. See N-B (note). |
| F2, P10 public-module consumption (A07.10) | **Present and correct.** SH L50–52, L202–209 (row 205 "superseded"), L78–97, L3–5; owner as-is direction at SH L5; B12 L49 is the conflicting text. Scope limited to this module. |
| F2, P8 resume compatibility (A07.5) | **Present, but the status is wrong — see F1 below.** |
| P2: LR alias, G1 rewrite, baseline wording | Handled (LR → launcher README, G1 mismatch cell now names SH §3.8 L169–184 board binding). |
| P2: I5 ledger locator, W4 key classes, P11 sweep | Handled and correct against UM L70–74/L152–154, CB L92–104/L231–235, DP L75–81/L209–232 (`sweep` is indeed absent from the `dispatch-op/1` table). |
| Nothing else changed in meaning | Confirmed: the diff against rev1 is the seven new rows, G1 mismatch text, the source-index compression, the LR row, the rev2 section and evidence wording. |
| Size, hygiene | 76,851 bytes ≤ 80 KiB. No `LOGBOOK.md` change. Scan for personal first names, local paths, secrets, session links: none. |

## 2. Findings

```yaml
findings:
  - id: DB-F1
    severity: regression
    severity_reason: "P1: check 2 (decided vs open) — a recommendation from an explicitly open owner-decision register is presented as the specified canonical value; the P0 authors would rewrite the book to it."
    repeat_of: null
    location: ".research/261010_platform-contract-table.md row P8 (line 106), status cell"
    mechanism: "P8 states the 'qualified-version-range' rule as canonical and labels it 'Specified architecture target, not a new owner decision'. The cited source does not support that status: AR §5.3 (heading, L413) carries the footnote [^предложено] ('Proposed … not canon until joint review', footnote L1973–1974; [^открыто] at L1977); D-R5 sits in AR §10.2 'Решения для владельцев D-R1…D-R8' marked [^открыто] ('no decision'), whose text says 'Рекомендации — совет, не решение' (L1814–1817); the D-R5 row is L1825. The book (B07 L73, 'той же версии') says same version; the architecture proposes a qualified range after a probe. Neither is decided."
    recommendation: "Change the P8 status to OWNER DECISION NEEDED with the two options taken from the sources: (a) qualified version range with a successful probe for the adapter version (AR L476, D-R5 recommendation, flagged proposed/open); (b) same version only (B07 L73, current book). Cite AR L413, L1814–1817, L1825, L1973–1977. Keep the rest of the preconditions (same family, managed home, profile pin, transcript digest, exclusive lock, auth, grants, no transport conversion) as the common part. Say that until decided, R-SH1 and the book chapter must not state either rule as settled, and that both documents keep the A07.5 mismatch open."
  - id: DB-N1
    severity: note
    severity_reason: "P2: formatting only, readability of the F1/F2 rows (A: DA-N1, verified)."
    repeat_of: null
    location: ".research/261010_platform-contract-table.md lines 88 and 104"
    mechanism: "A blank line precedes I5 (line 89) and P7 (line 105); under GFM the preceding table ends, so I5/W4 and P7–P11 render as pipe-text paragraphs, not rows."
    recommendation: "Delete both blank lines (lines 88 and 104)."
  - id: DB-N2
    severity: note
    severity_reason: "P2: stale owner-section label (A: DA-N2, confirmed)."
    repeat_of: null
    location: ".research/261010_platform-contract-table.md row W4, owner cell"
    mechanism: "Owner cell says CB '§§4.3/6.5'; the cited text at CB L231–235 sits under §6.6 and the migration procedure under §4.3."
    recommendation: "Use CB §§4.3/6.6."
  - id: DB-N3
    severity: note
    severity_reason: "P2: completeness of the P9 mismatch cell, not gating."
    repeat_of: null
    location: ".research/261010_platform-contract-table.md row P9, mismatch cell"
    mechanism: "B07 L119 ends 'Владелец принял, что куски кода доски могут попасть в публичные логи' — an owner acceptance of public-log exposure. The row replaces the book wording with the SG/SH private-lane layout but does not say that this sentence is an owner-attributed acceptance and has no decision text beside it in SH/SG."
    recommendation: "Add one clause: the B07 sentence must be removed or restated together with the lane correction (it is moot while lanes 1–2 are private), and the SG/SH layout is the module-owner text, not a new owner decision."
  - id: DB-N4
    severity: note
    severity_reason: "P2: P11 owner question may be over-asked."
    repeat_of: null
    location: ".research/261010_platform-contract-table.md row P11"
    mechanism: "DP §7.2 L211 already states the timer runs `swarma-dispatcher sweep` as the service account; the closed `dispatch-op/1` table covers caller requests only. The recorded options (internal CLI entry versus a public op) are real, but the first reading needs no decision."
    recommendation: "State that the default reading is an internal service-only entry documented outside `dispatch-op/1`, and that a public operation would be a new decision."
```

## 3. Reconciliation with delta review A

- A reports no remaining P0/P1. I disagree on one point: A marks P8 "Fixed" because the row exists and is faithful to the AR text, but A does not test its **status**. Check 2 of the rev1 brief — "no row turns a recommendation into a decision" — applies to P8. That is DB-F1 (P1).
- A: DA-N1 → DB-N1, confirmed on lines 88 and 104. A: DA-N2 → DB-N2, confirmed. DB-N3 and DB-N4 are mine and are not in A.
- A's other conclusions (F1, P9, P10, I5/W4, P11, hygiene, size) agree with my checks.

## 4. Routing

Changes requested; research rework by the same producer role, route to `analysis`. Required for rev3: DB-F1. Recommended in the same revision: DB-N1…DB-N4 (cheap). Everything else in rev2 stays as is. Not accepted, so `accept_cr` was not called.
