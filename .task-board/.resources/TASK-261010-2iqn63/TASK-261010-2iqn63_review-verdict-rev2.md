# Reviewer B (deciding, cross-provider) — verdict on Change Request rev2 of TASK-261010-2iqn63

Candidate tree ffb71052, base c53ba4b9; study `.research/261010_platform-docs-audit.md` (560 lines; rev1 reconstructed from the rev1 patch, 517 lines). Read-only; CIP-0009 fetched at 1604d402, plus A09.9/A09.4/bridge sources at their pinned commits, via `gh api`. No builds or tests. Written BEFORE reading `docs-audit-delta-A.md`.

## Verdict: CHANGES REQUESTED — one hygiene item, everything else holds

Every row of the rev1 checklist is fixed and verified. One first name quoted from the private document still stands in two link titles, which delta check 3 ("no first names") does not allow in a public file.

## Check 1 — listed rows

P1:
- **A09.9** fixed. Book B09:95 "Нет места — запрос ждёт" and B09:105 "постоянный отказ `quota_exceeded`, а не очередь" are now stated correctly. DP:160 "refuse (v0) / queue (v1)" and DP:251 "permanent refusals (quota, authority) never queue" support the new text: the book disagrees with v0 only. R-DP1 basis matches (A09.6–A09.10, with the qualifier note).
- **A06.1, A08.4, A11.7, A11.8** fixed. At 1604d402: L145 carrier clause ("What it never sees: … the carrier, carrier credentials") ✓; D9 L160 ✓; D10 L162 ✓; D11 L164–169 ✓; L144–145 disconnected behaviour plus carrier clause ✓. The old L149–158 locators are gone.
- **CIP-0009 sweep** done for every C09 locator in the study (matrix rows, section B tables, section C and D). A11.6 (L128–133, L144–145: D4 steps 1–6 and carrier clause) ✓. A11.10 (L139–145, L164–169) ✓. A11.11 (L149–154: D7/D8, previously L143–146 = D5 delivery) ✓. A04.5, A11.1–A11.5, A11.9: unchanged locators still match the text. A04.4 now says "Ed25519 signing/SSH keys", which is accurate: L76 gives variant A P-256 Secure Enclave signing/SSH keys, L129 gives Ed25519 `K_sig`/`K_ssh`.

P2:
- **A09.4** relabelled "AMBIGUITY: default registry versus shipped trust bundle"; matches C07 L16–28 (separate release-shipped trust bundle) versus C07:L241 "built-in registry only". Clarification action kept. ✓
- **A02.2, A07.6, A08.3** relabelled SCOPE GAP / QUALIFIER GAP; wording no longer asserts incompatibility. ✓
- **D1:** R-LR1 and R-CS2 carry OWNER DECISION; R-CS2 now sits before R-CB1 and R-CB1 names the dependency. ✓
- **Row IDs:** R-SH1 (A07.1…A07.10), R-CB1 (A10.1…A10.9), B-F1 (30 IDs) are listed concretely. I recomputed: B-F1 contains every Ch02–13 CONFLICTING row plus the five relabelled gap rows; the only CONFLICTING rows absent are A00.1, A00.3 and A15.10, outside the Ch02–13 scope. CONFLICTING count fell 37 → 32 = the five relabelled rows. ✓
- **A06.2, A08.1** now cite the public `agent-session-bridge` README stub (L1–21) and its pinned tree; I confirmed the tree holds only .gitignore, LICENSE, NOTICE and README.md, and the README is a design stub. ✓
- **Names:** "Иван"/"Алексей" no longer appear in the five L1412 link titles (now "§7.8"). **But see finding B2-F1.**

## Check 2 — delta scope
rev1→rev2 differs in 18 hunks: the listed rows, the sweep rows (A04.4, A11.6, A11.10, A11.11), the five title replacements (A03.12, A13.1, keeper-start in B1, and the two titles in the architecture row of section C), the D1 rows R-LR1, R-DP1, R-SH1, R-CS2, R-CB1, B-F1 (Ch02–07/09–13 → Ch02–13, matching its A08.4 basis), and the appended "Rev2 changes" section. Nothing else differs; the Rev2 changes table lists every content change and mentions the titles generically.

## Check 3 — hygiene
- Delta is the one research file; `git diff --name-only` lists only `.research/261010_platform-docs-audit.md`; LOGBOOK untouched.
- No secrets, local paths, host names or session links found (searched for user paths, `~/`, `.temp/`, session ids, tokens, key headers).
- **First names: not clean.** `Алексея` remains twice, in the title of the link to `architecture.ru.md` L1922–1935 ("12. Диаграммы (репозиторий Алексея, `diagrams/`)"): once in the B2 legend paragraph (study line 326) and once in the B4 preamble (study line 385). Same class as the L1412 titles — a first name quoted from the private document's heading — and unchanged from rev1. The rev1 checklist named only the L1412 titles, so the producer did what was asked; the residual is a miss in my own rev1 checklist.

Why I do not wave it through as a P2 note: the file lands in a public repository, a name published there stays in history, and the owner asked for exactly this class of text to be removed. The fix is two link titles.

## Findings block
```yaml
findings:
  - id: B2-F1
    severity: robustness
    severity_reason: "P1 — hygiene: first name 'Алексея' in two link titles (L1922–1935, study B2 legend and B4 preamble) of a file that lands in a public repo; delta check 3 requires none; irreversible once published"
    location: "B2 legend paragraph and B4 preamble link titles for architecture.ru.md#L1922-L1935"
  - id: B2-N1
    severity: note
    severity_reason: "P2 — A04.4 wording change and A11.6/A11.10/A11.11 locator changes come from the requested CIP-0009 sweep and are accurate; the Rev2 changes table lists them, but names the five title replacements only generically (A03.12, A13.1, keeper-start, two in the section C architecture row)"
    location: "Rev2 changes"
```

## Rows to fix in rev3 (only this)
- Replace the title text `12. Диаграммы (репозиторий Алексея, `diagrams/`)` with `§12` in both link titles (study lines 326 and 385; the URLs stay unchanged), then grep the whole file for any other first name (Cyrillic and Latin), and add one line to "Rev2 changes" (or a "Rev3 changes" line). Nothing else is to change.

## Reconciliation with reviewer A (`docs-audit-delta-A.md`, read after the sections above)
- **Agree** with A on every checklist surface: prior B-F1…B-F6 resolved, C09 sweep held, plan ordering and OWNER DECISION tags present, bridge stub accurate, scope held, delta is the one file. A's independent fetches match mine (D9 L160, D10 L162, D11 L164–169, carrier L145; A09.9 phases).
- **Disagree on one point.** A reports "removed names: none found" and `findings: []`. The candidate blob still contains `Алексея` twice (study lines 326 and 385, title of the link to `architecture.ru.md` L1922–1935; the heading is in the private wiki, checked with `gh api repos/relux-works/wiki --jq .private` = true). A's hygiene check evidently covered the L1412 titles only. This is not a repeat of an earlier finding; it is the same class, found by the full-file grep. A raised no P0/P1, so there is nothing of A's to carry over.
- Result: B2-F1 stands. Reject for rev3 with the one-item fix above.
