# Contract table, deciding review B, CR rev1 — verdict: CHANGES REQUESTED

Task: TASK-261010-1ypla7. Candidate: CR-TASK-261010-1ypla7-1 rev 1, base `67f736f5e45f562df892674cb53c97b0e6d732e2`, tree `645ad50d014a3fe443a9d6b9d50d35a39f6bc679`. One file changed: `.research/261010_platform-contract-table.md` (64,667 bytes, 869 under the 64 KiB cap).

Order followed: checks 1–4 were done and the verdict drafted before `contract-table-review-A.md` was read. A was then reconciled; every A finding was re-verified against the pinned text, not taken on trust.

## 1. Citations — pass

Pinned files re-fetched read-only with `gh api ...?ref=<full commit>` (14 files, all exit 0). Rows checked line by line: N1–N6, N7, W1, W2, W3, I1, I2, I3, I4, L1, L2, P1–P6, Q1, Q2, C1, C3, C4, C5, C8 (partly), C9, G1, G2, R1, R2, X1, X2, D1, D2, D3, D4, D5, V1 (status headers of CIP-0002…0007) — about 35 rows. Every canonical value matched its cited text:
- names and socket directories: B02 L29–40, CB L465–473, DP L54–59, SH L103–123;
- sizes: helper 16 KiB / launcher 80 KiB / 5 s (UM L30–48), plan 48 KiB / 64 KiB / 96 KiB / 128 KiB (DP L253–263), broker frame 64 KiB (CB L425–427);
- domains: `swarma-dispatcher request/1` (DP L73–86), effect id 32 hex of SHA-256 over `swarma-dispatcher effect/1` + CCJ-1 (DP L145–151; the text indeed names no newline), `curator-broker grant/1` / `revocation/1` (CB L164–216);
- quota defaults (DP L88–101); ticket payload and epochs (DP L270–276, SH L125–139); PATH recipe (C07 L16–28 sample line, C02 L161–169); C09 variants, trust bootstrap, CAs (C09 L72–81, L156–158, L221–229; B11 L49–56).
The owner precondition digest `0fab6869…28bd` matches the board resource, and D-GOALS lines 3–7 / D-IDENTITY lines 9–12 are cited at the right lines. No miscitation found.

## 2. Decided versus open — pass

- G1/G2 reflect D-GOALS exactly: variant (b) now, only board/orchestrator or the launching caller at start sets a goal, no agent goal capability on the host socket, host works without goals/board/orchestrator, board goals an optional client of a generic interface, (a)-versus-(b) deferred to a separate SH2+ item. The WHY is explicitly labelled a design explanation, not an owner quote.
- R1/R2 reflect D-IDENTITY exactly: one-shot flag recorded at registration, dispatcher revokes as a separate operation after retirement; standing identities revoked only by the operator with a signed record.
- OWNER DECISION NEEDED rows (N6/P1, W2, W3, I4, C4, C7, C9, X1, D1, D5, G2) each state real options and none turns the recommendation into a decision; C9 labels its recommendation as not a decision. Decided rows cite real decisions (C10 rev 2.2 L24–31, C09 L221–229, DP §3.5 L88).
- Provenance exception for the two newer decisions is disclosed honestly; retained.

## 3. Coverage of the audit D1 P0 items — FAILS (P1)

The audit's R-UM1 and R-SH1 name contradictions that have no canonical row anywhere in the table (grep of the candidate: no `archive`, `ledger`, `locator`, `sweep`, `key class`, no A03.9 / A04.1 / A07.5 / A07.8 / A07.10 / A09.7).

- **F1 (P1) — R-UM1 `keep` enum (audit A03.9).** The audit says R-UM1 must "resolve `keep` enum". The helper's retire request admits archive/delete while the archive policy prose adds an operator-only `keep` (UM L121–150). The table has no row stating the canonical retire-fate set, who may select `keep`, or which documents must match (dispatcher cleanup, book ch03). Add a row (decided from the spec, or OWNER DECISION NEEDED with options: admit `keep` as an operator-only request value, or drop it from the policy prose).
- **F2 (P1) — R-SH1 A07.5, A07.8, A07.10.** R-SH1's basis lists A07.1–A07.10. Three CONFLICTING audit rows have no row: A07.8 (book puts all three SH1 gate lanes in public module CI, but source/consumer lanes are private — SG L3–20, SH L28–97), A07.10 (public tagged-module consumption versus the private-fork mandate in book ch12 — SH L202–209), A07.5 (exact-version native resume versus qualified-version range — AR L476–488). Add rows for gate lane ownership/evidence limits, public-module provenance tense, and the native-resume compatibility rule, each with owner section and must-match set; mark any unresolved one OWNER DECISION NEEDED.

Both are cheap to repair; the table is the only thing the P0 authors follow, so a gap here reappears as drift.

## 4. Hygiene — pass

Only the research file changed (`git status`: one untracked file; the diff has no `LOGBOOK.md` change). Scan for first names, local absolute paths, secrets and session links: none. `board-resource://` references are neutral. Size cap met.

## 5. Reconciliation with review A

- A-F1 = my F1: **agree, P1**.
- A-F2 = my F2: **agree, P1** (A07.5/8/10 verified in the audit rev3 patch lines for those rows).
- A-F3 (R-CB1 helper-ledger locator and key classes, W2 software-root vs keeper adapter, R-DP1 sweep entry): **agree on the gap, downgrade to P2**. These items are named in R-CB1/R-DP1 text, but the brief's required contract list does not include them and the audit rows A04.1 / A05.1 already carry them. Fix them in the same revision because the revision is open anyway; I do not gate on them.
- A-N1 (alias `LR` used throughout "must match" cells but absent from the source index; `LS` is defined): **confirmed, P2**. Define LR (launcher README at the LS pin) or replace it by LS.
- A-N2 (provenance exception faithful): agree.

## 6. Additional P2 notes (tb-R226, not gating)

- G1 mismatch cell calls the old text only a "namespace/phase gap". SH §3.8 steps 1–4 (L169–184) route every goal through the board's binding; under D-GOALS the section must be reframed as one optional client of a generic goal interface. Say so in the mismatch cell so R-SH1 does not keep the board-centred wording.
- The evidence baseline is labelled audit rev2; the newest candidate (rev3, which added the owner-decision flags on R-LR1/R-CS2) now exists on the board. State the revision actually used, or re-point.
- Headroom is 869 bytes. Rows for F1/F2 (and the P2 items) need compression elsewhere (e.g. the repeated "Source and propagation index" prose) to stay under 64 KiB.

## Routing

Changes requested, route to `analysis` (research rework by the same producer role). Required for rev2: F1 and F2 rows; P2 items recommended. Everything else in rev1 may be kept as is.
