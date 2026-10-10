# Reviewer B (deciding, cross-provider) — verdict on Change Request rev1 of TASK-261010-2iqn63

Candidate tree e819ceae, base c53ba4b9. Study `.research/261010_platform-docs-audit.md`. Read-only; sources fetched with `gh api` at the pinned commits. No builds or tests.

## Verdict: CHANGES REQUESTED (rev2 of the study needed)

My independent record `TASK-261010-2iqn63_review-B-independent.md` (written before reading reviewer A) found no P0/P1 and six P2 notes. Reviewer A's record `docs-audit-review-A.md` raised three P1 findings. After verifying each one against the pinned sources I uphold two of them as P1 and downgrade one to P2. The upheld findings are factual errors in matrix rows that feed the P0 repair items R-DP1 and R-CS1/R-CS3, so the orchestrator cannot cut producer tasks from them as they stand.

## Reconciliation with reviewer A

### A-F1 (row A09.9) — UPHELD as P1
- Book B09:95 "Ёмкость и ворота … Нет места — запрос ждёт, а не падает": capacity shortage waits.
- Book B09:105 "Сверх лимита — постоянный отказ `quota_exceeded`, а не очередь": only quota refuses.
- DP:160 "take an owned claim (§9) or refuse (v0) / queue (v1)"; DP:251 "permanent refusals (quota, authority) never queue".
- The row states the opposite ("Capacity refusal has no queue … CONFLICTING if applied to v1"). The book matches v1 and disagrees only with v0. My independent record had this as P2 because the underlying unlabelled-phase finding stands; on reflection an inverted clause in a row named as the basis of P0 item R-DP1 ("v0/v1 capacity") is a defect in the deliverable, not a wording note. I agree with A: P1.

### A-F2 (row A09.4) — DOWNGRADED to P2
- C07:21 decision: "No trust keys compiled into the binary … A release ships a default trust bundle as a separate file".
- C07:241 recommendation: "Defaults: built-in registry only". C07:565–568 and C07:589–591 recommend pins and a pinned snapshot "distributed with the manager release".
- A release-shipped default registry plus a separately shipped default trust bundle do not contradict "no keys in the binary". The row labels this CONFLICTING without establishing compiled-in keys. The row's own requested action ("distinguish default configuration bundle from binary authority") is the right clarification and does not commission a security-contract change, so the harm is limited to an over-strong label. I disagree with A on severity: P2, fix the label in rev2.

### A-F3 (rows A06.1, A08.4, A11.7, A11.8) — UPHELD as P1
At pinned CIP-0009 (1604d402): D7 lifecycle is lines 149–152, D8 line 154, D12 lines 156–158, **D9 line 160, D10 line 162, D11 lines 164–169**. The rows cite:
- A11.7 "COVERED D9" → C09:L149–150 (D7 stop/leave/purge).
- A11.8 "COVERED D10" → C09:L151–154 (D7 project side, D8).
- A06.1 "donor CIP confirms final carrier separation" → C09:L155–158 (blank line + D12 SSH CAs); the carrier clause is C09:145 and D11 C09:168.
- A08.4 "COVERED target donor/carrier decision" → C09:L155–156 (blank + D12 header); D11 is C09:164–169.
The acceptance criterion is "every claim cited to file and line". Four locators on normative D-blocks point at unrelated sections; a producer following them lands on the wrong text. I agree with A: P1. The underlying findings (gap in A11.7, policy-admission note in A11.8, carrier-token drift in A08.4) are real and survive; only the locators must be repaired, and every CIP-0009 D-citation should be re-swept semantically since the D-blocks are out of numeric order in the file (D12 sits before D9).

### A-N1 — agree (P2)
Same as my F-B2: A07.6/A08.3 are a missing namespace qualifier, not a contradiction.

## Rows to fix in rev2 (by ID)
P1:
- **A09.9** — restate the book clause (capacity waits, quota refuses); compare to v0 refuse, not v1; adjust R-DP1 basis text.
- **A06.1, A08.4, A11.7, A11.8** — replace CIP-0009 locators with D9 L160, D10 L162, D11 L164–169, carrier clause L145; sweep all other C09 D-block citations (A04.4, A04.5, A11.1–A11.11) for the same drift.

P2 (tb-R226 notes, fix in the same revision since one is pending anyway):
- **A09.4** — relabel from CONFLICTING to "ambiguity: default registry vs. shipped trust bundle"; keep the clarification action.
- **A02.2, A07.6, A08.3** — relabel as scope/qualifier gaps.
- **D1 R-LR1, R-CS2** — add an explicit OWNER DECISION tag (AR §5.1.8 D-R3; first consumer A10.9).
- **D1** — put R-CS2 before R-CB1 or state the dependency order.
- **D1 R-SH1, R-CB1, D2 B-F1** — list the concrete A-row IDs instead of ranges.
- **Link titles for AR:L1412 (five occurrences)** — remove the first names quoted from the private document's heading; use "§7.8" (public repository hygiene).
- **A06.2, A08.1** — mention the public bridge README stub as the only bridge source.

## Checks 1–6 status
1–3 verified in the independent record (17 CONFLICTING rows, 7 gaps, 9 diagram claims; the "19 of 26" figure count holds). Check 4: order sound except the two P2 notes above. Check 5: delta is only the research file, no LOGBOOK change, no secrets/paths/session links; first-name hygiene is P2. Check 6: actionable with the row-ID note above.

## Findings block
```yaml
findings:
  - id: B-F1
    severity: robustness
    severity_reason: "P1 — A09.9 inverts the book's capacity/quota clauses and the v0/v1 phase; feeds P0 item R-DP1 (upholds A-F1)"
    location: "A09.9, D1 R-DP1"
  - id: B-F2
    severity: robustness
    severity_reason: "P1 — A06.1, A08.4, A11.7, A11.8 cite CIP-0009 D7/D8/D12 lines for D9/D10/D11 claims; AC requires file-and-line citations (upholds A-F3)"
    location: "A06.1, A08.4, A11.7, A11.8"
  - id: B-F3
    severity: note
    severity_reason: "P2 — A09.4 CONFLICTING label not established; default registry and separate trust bundle coexist (downgrades A-F2)"
    location: "A09.4"
  - id: B-F4
    severity: note
    severity_reason: "P2 — A02.2, A07.6, A08.3 are qualifier gaps, not contradictions (agrees with A-N1)"
    location: "A02.2, A07.6, A08.3"
  - id: B-F5
    severity: note
    severity_reason: "P2 — R-LR1/R-CS2 need OWNER DECISION tag; P0 list not in dependency order; D items cite A-row ranges"
    location: "D1, D2"
  - id: B-F6
    severity: note
    severity_reason: "P2 — first names from a private document heading in a public file (5 link titles); bridge README stub unmentioned"
    location: "link titles for AR:L1412; A06.2, A08.1"
```

Routing: the producer is a researcher, so after `reject_cr` the element is set to `analysis` for the rev2 study.
