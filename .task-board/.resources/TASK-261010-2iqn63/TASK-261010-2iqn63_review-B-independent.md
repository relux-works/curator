# Reviewer B (cross-provider, deciding) — independent verdict, written before reading reviewer A

Task TASK-261010-2iqn63, Change Request rev1 (base c53ba4b9, candidate tree e819ceae). Study: `.research/261010_platform-docs-audit.md` (same as the outcome resource). Method: every source was fetched read-only with `gh api` at the pinned commits named in the study's register, then compared with the row text. No builds, tests or renderers were run.

## Verdict (independent): ACCEPT — no P0/P1 finding

Checks 1–5 hold. Six P2 notes follow (tb-R226). None changes the order or dependencies of the repair plan.

## Check 1 — CONFLICTING rows (17 verified, 3 over-labelled)

Real, with the exact contradiction as the row states it:

| Row | Chapter | Texts compared | Result |
|---|---|---|---|
| A00.1 | 00/11 | B00 front matter "8 октября … v0.4, CIP-0008–0010 вторая редакция" vs B11:5 "10 октября … третья редакция" | real |
| A02.3 | 02 | B02:29–40 names table (`swarma-dispatcher`, `swarma-credential-broker`, `swarma-session-host`) vs SH:1–6 `curator-session-host` + `curator-*` companions, CB:447 `curator-broker`, DP:240 `swarma-dispatcherd`, NM `agent-session-host` | real (stale names) |
| A03.2 | 03 | B03:22 "не больше 16 КиБ" for both binaries vs UM:36 16 KiB `helper-op/1` / 80 KiB `launch-op/1`, DP:255–263 | real |
| A03.4 | 03 | B03:36 "Один номер запроса связывает создание, привязку и запуск" vs DP:147 "create, bind, launch, stop, unbind and retire use distinct ids" | real |
| A03.5 | 03 | B03:40–48 drops privileges (step 5), then writes the start receipt (step 6) vs UM:200–207 receipt (6) before privilege drop (7) | real |
| A03.9 | 03 | UM:121 `home: "archive" \| "delete"` vs UM:147 `keep` | real, internal to the helper spec |
| A04.2 | 04 | B04:27 `revoke` "владелец или диспетчер" vs AR:937 `revoke(index)` "только оператор" | real |
| A05.6 | 05 | B05:97 domain tag `curator-dispatcher request/1` vs DP:76 `swarma-dispatcher request/1` | real, wire constant |
| A07.1 | 06/07 | B06:11–15 "псевдотерминал принадлежит процессу хоста сессий" vs B07:13 and SH:115 runner holds the PTY | real, book vs book |
| A07.8 | 07 | B07:119 "Все полосы … в публичном CI модуля" vs SG:8–20 lane 1 private self-hosted, lane 2 private workflow, lane 3 public | real |
| A09.1 | 09 | B09:24: "`curator-run` исполняет план" and, in the same section, "Сам он ничего не исполняет" | real, internal |
| A09.10 | 09 | B09:106–109 dispatcher runs `curator-run` headless vs D21:135 "There is no headless `curator run`"; AR §5.1.8 names the amendment as a pending owner decision (D-R3) | real |
| A10.3 | 10 | B10:33 "в связке ключей у человека" vs C10:22 and C1 "first release ships only a 0600 file backend… No Keychain" | real |
| A10.9 | 10 | B10:65 "Первым исполнителем стал `curator-run`" vs C10:148 task-board spawn runner first, C11:7 same, C11:118 Curator's executor recommended | real |
| A11.5 | 11 | C09:130 "DNS check, mandatory… nothing downgrades", C09:207 mandatory DNSSEC not cut, vs C09:222 invitation pins are the fallback for a domain without DNSSEC | real, inside the CIP |
| A12.5 | 12 | B12:49 "Каждая замена строится в приватном форке" vs SH:205 "cutover through the private fork — superseded" | real |
| A13.1 | 13 | B13:17 cleanup "привязка и ключ отзываются" vs B04:35 and DP:292 retirement keeps identity | real |

Over-labelled or mis-described:

- **A09.9 (P2).** The row restates the book as "capacity refusal has no queue". B09:95 actually says "Нет места — запрос ждёт, а не падает", and only the quota refusal is "а не очередь" (B09:105). The book therefore matches DP v1 (queue on gate closure, no queue on quota) and disagrees only with v0 (DP:160 "refuse (v0) / queue (v1)"). The finding stands, but as an unlabelled phase, and the text "CONFLICTING if applied to v1" has it backwards.
- **A02.2 (P2).** "Agent holds no secret" is qualified by the book itself (B02:23 "Всё, чего агент не может сделать физически…", B03:140 token readable in the agent process). That is a scope/ordering gap, not a contradiction with CB:403–413. The "board optional / single admission writer" half is fine.
- **A07.6 / A08.3 (P2).** SH:182 "An agent cannot set its own goal" sits in §3.8 (board goals across accounts) and B07:81–90 is the same board-goal context; B08:26–28 and AR:586–630 split board and board-free goals. This is a missing namespace qualifier in SH §3.8, not an unlabelled contradiction. D1 R-SH1 already words it as "namespace disposition", which is the right size.

Spread: chapters 00, 02, 03, 04, 05, 06/07, 09, 10, 11, 12, 13 are covered. The count is above the required 12.

## Check 2 — MISSING / ARCH-ONLY rows (7 verified, all real)

- A07.3: `command grep -i approv SH.md` finds nothing. The SH2 spec has no approval/unknown-outcome wire or state schema. MISSING stands.
- A08.1 / A06.2: the only public bridge source is `agent-session-bridge` with a README (design, trust gate list). Its tree has no spec or diagram files, so "no standalone bridge module contract" holds. The study does not mention that stub README; P2 note below.
- A04.1 / A04.6: the keeper appears only as "later" in CB:104/488 and DP:292, and as the AR §7.2 design. ARCH-ONLY stands.
- A11.7: join verbs exist (C09 D4 steps 3–6) but C09:194 itself lists the schemas as future work. MISSING a standalone contract stands.
- A12.1: board module specs are outside the audited set; the row says so.
- A05.5, A13.2, A15.x: consistent with the bounded wording ("in the audited corpus").

## Check 3 — Diagrams (9 claims verified)

- Book tree at 2e30edbd: 26 PNGs, 7 `.puml`, the 7 stems all match a PNG, so 19 of 26 have no local source. ✔
- Module trees: broker 5 puml, user-manager 2, dispatcher 5, session-host 0, launcher 0; curator-spec PR134 `cips/diagrams/` 5 files, PR136 none. Matches B1/B2 and the "43 identities" count. ✔
- UM `launch.puml`: "Executor (agent account, uid 612)" is outside the 30000–59999 range of UM:86; its note "sudo stays the parent and relays signals" contradicts UM:209. ✔ Receipt is correctly before the drop in the figure, so the book is the odd one out.
- CB `agent-registration.puml`: `launch.start` carries no execution_id/epoch (UM:194 requires them). ✔
- DP `run-states.puml`: run/1 collapses failed/expired into the outcome note and sends queue timeout to `cancelled`; DP:132–142 lists `failed`/`expired` as states. ✔
- Book `containers.puml`: no runner; `HOST --> HARN : держит процесс`, `HOST --> CRED : учётные данные`. ✔
- No proposed module diagram (M-CB1…M-LR2) exists already: confirmed by the trees above (no components/state/layout files in broker, user manager, dispatcher except the five named; none in session host or launcher).
- Not verified: "five S pairs match embedded PlantUML metadata exactly" (would need PNG metadata decoding). Recorded as unchecked, not as wrong.

## Check 4 — Repair plan D1/D2

Order follows from the findings: contract reconciliation (P0) before the new figures (P1) before the index (P2). Dependencies are stated per row. No item asks a producer to invent behaviour: the schema items say "or explicitly label those schemas pending", and the helper-extension item says it needs a decision. Gaps (P2):

- R-LR1 depends on the pending owner decision D-R3 (AR §5.1.8, launcher owner) to amend Decisions 0019/0021 for the one-shot mode. D1 says "Requires explicit enactment" but never names the decision or tags the item OWNER DECISION. The same applies to A10.9 (first consumer) inside R-CS2. Add an explicit tag column.
- R-CB1 (P0) depends on R-CS2 (P0), which is listed after it, so the P0 list is not in dependency order.
- B-F1 depends on all R-* P0 items; the row says so.

## Check 5 — Scope and hygiene

- Delta is only `.research/261010_platform-docs-audit.md` (`git diff --name-only`). No `LOGBOOK.md` change; the three LOGBOOK mentions are statements that it was not touched.
- Scan for `/Users`, `/home`, tokens, key blocks, e-mail, session links, local worktree paths: clean.
- P2: the heading "7.8 Решения звонка 7 октября 2026 (Иван и Алексей)" is quoted five times in link titles. It is a public repository; replace with "§7.8" (tb-R226 hygiene).

## Check 6 — Readability

Each D item names files and diagrams. P2: R-SH1, R-CB1 and B-F1 send the engineer to ranges of A-rows ("A07.1–A07.10", "Apply A's CONFLICTING rows") instead of listing them. Add the row IDs of the concrete contradictions per item.

## Findings block

```yaml
findings:
  - id: F-B1
    severity: note
    severity_reason: "P2 — row A09.9 misstates the book clause (capacity waits; quota refuses) and inverts the phase; the finding itself (unlabelled v0/v1 difference) stands"
    location: "A09.9"
  - id: F-B2
    severity: note
    severity_reason: "P2 — A02.2 and A07.6/A08.3 are scope/qualifier gaps the book or architecture already partly label, not contradictions"
    location: "A02.2, A07.6, A08.3"
  - id: F-B3
    severity: note
    severity_reason: "P2 — R-LR1 and the first-consumer item in R-CS2 need an explicit OWNER DECISION tag (AR §5.1.8 D-R3, A10.9)"
    location: "D1 R-LR1, R-CS2"
  - id: F-B4
    severity: note
    severity_reason: "P2 — P0 list is not in dependency order (R-CB1 depends on later R-CS2)"
    location: "D1"
  - id: F-B5
    severity: note
    severity_reason: "P2 — D items cite A-row ranges instead of listing the contradictions (R-SH1, R-CB1, B-F1)"
    location: "D1, D2"
  - id: F-B6
    severity: note
    severity_reason: "P2 — first names from a private document's heading appear five times in a public file; bridge README stub not mentioned for A06.2/A08.1"
    location: "link titles for AR:L1412; A06.2, A08.1"
```
