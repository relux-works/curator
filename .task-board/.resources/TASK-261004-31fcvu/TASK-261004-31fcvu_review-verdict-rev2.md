# TASK-261004-31fcvu — rc3-release-notes: revision 2 review verdict

Verdict: accepted. Accept CR revision 2 and route to integrating; this review does not land the change or mark the task done.

Reviewed base `876127f7c714e01092f43c8950dc879421c52461`, previous candidate `92a37bbd92b9686b6bd384ac95114d7f8d6e2a74`, and exact revision 2 candidate `6c30a1ff1e35ceda69bdaf86207f75e763cab835`. Fresh `git ls-remote origin refs/heads/main` returned the base, exit 0. Worktree CHANGELOG bytes equal the reviewed candidate. No repository files were modified by the reviewer. `task-board spawn goal` confirms this run is not goal-bound.

## Findings and swept surfaces

No outstanding findings. The prior two findings are resolved by mechanism, as below.

| Surface | Result | Repeat-of / disposition |
| --- | --- | --- |
| Historical entry preservation | PASS: 27/27 byte-identical entries, in original order, under the required final historical subsection and exact attribution sentence; per-entry proof below | rev1 P1: tag Unreleased incorrectly treated as released-section evidence; resolved |
| Public wording | PASS: read the entire rc.3 section, including restored historical entries; known internal identifier removed; no internal host/machine name, personal user path or employer name identified | rev1 P2: internal runner label disclosure; resolved |
| Validator regression | PASS: exact rev1 candidate exits 1 naming all 27 missing entries; exact rev2 exits 0; three regression tests pass including all 27 individual omissions | rev1 P1/P2: corrected predicate and full-section scan exercised |
| Delta confinement | PASS: independent reconstruction proves rev2 equals rev1 plus precisely the historical subsection insertion and the neutral runner-label replacement | No new finding |
| Repository scope and older releases | PASS: CHANGELOG.md is the only changed path; LOGBOOK.md unchanged; all bytes from rc.2 heading through EOF match base | Prior PASS retained and independently rerun |
| Structure, pin/write policy, known issues | PASS: empty Unreleased, rc.3 heading/date and groups; rc.14 pin, v1 writes, fp8vx7 risk, rc.4 writer deferral, N1–N4 unfixed with #106, B3 exclusion retained | Prior PASS retained; validator and direct prose inspection rerun |
| History identity/accounting | PASS: validator rerun verifies 54/54 subject mappings and 258/258 board-only commits | Prior content review retained under binding delta-review instruction |
| Architecture/implementation | PASS: documentation-only revision follows required historical attribution; no product behavior changes | No new finding |

The public-wording check combines the producer's known-label/private-path pattern scan with manual reading of the full rc.3 section (lines 7–508). Generic documented paths and environment variables are product instructions, not personal paths. No private identifier is reproduced in this outcome.

## Executed evidence and limits

- Independent byte-level Python comparison: exit 0. Parsed base bullets separately from the producer validator; compared 27 entries, original ordering and exact attribution; proved older-release equality and exact two-fix reconstruction against rev1. Entry hashes below cover the original bullet bytes excluding separating blank lines. Inter-entry spacing matches the required subsection reconstruction.
- Updated producer validator with exact rev1 candidate: exit 1, expected rejection for entries 5–20 and 22–32.
- Same validator with exact rev2 candidate: exit 0; 32/32 dispositions, release preservation, subject ledger, CI pin, disabled v2 writer, known issues and scope pass.
- Validator `--regression`: exit 0, 3 tests; rejects 27/27 individual entry removals, rejects moving an entry back into Unreleased, and detects the internal label in both fresh prose and the historical subsection.
- `git diff --check BASE CANDIDATE`: exit 0.
- Exact changed-path and working-tree identity checks: exit 0.

Accepted prior evidence: revision 1 verdict's 11 content patch spot checks (Muse, live-process sweep, rc.14 pin, Windows hard-link origin, marker cross-field checks, Codex seed A, posture A, Unix askpass, fragment v2, restore-backups, and versioned hash carriers). Those patches were not rerun in this delta review; independent reconstruction proves their rc.3 prose unchanged except the authorized neutral CI wording. The producer's subject map remains attached in TASK-261004-31fcvu_results.md and TASK-261004-31fcvu_map.json; its identity/accounting was rerun here. No Go builds, unit/platform/conformance suites or hosted stress runs were rerun or claimed as new evidence. These are documentation checks, not release qualification.

## Per-entry byte proof

| Source entry | Base line | Rev2 line | SHA-256 of exact entry bytes | Result |
| --- | --- | --- | --- | --- |
| 5 | 15 | 134 | `f6d4bf925f89817250ee3d3c7491acaa88917c66264f495ef7ae3f36fdd7db9f` | Exact |
| 6 | 33 | 153 | `1199da030e18c3ba4ea9cc87921d5916376306b715c4a1b0ac8231e235e8f59f` | Exact |
| 7 | 50 | 171 | `3a26f4b02debaff348206e75fd9c72114b9f233d5bd1f1b6ff3c9334b0470d02` | Exact |
| 8 | 65 | 187 | `b112cfb7d1147866834eac30397c44ab35cc2b81ed9f93f13177e9872c137d8e` | Exact |
| 9 | 101 | 224 | `8ac42ee6783788826ef1f7ec828b96015ecefb06cd7778dc2e52eb3efd1934ee` | Exact |
| 10 | 120 | 244 | `910943d29ddf278d80411a7dc44ea09e2b9cb8a9de3b990eb303c512ac4cdbe3` | Exact |
| 11 | 137 | 262 | `83a89b0acfba98d3701d1d9bca801993c524928ce3ce32b3c09b784b6995ce0b` | Exact |
| 12 | 146 | 272 | `1f981e13e332c2fa83f86fd72bf6d16cb083994b5a9cb29fd55dc58de8c5ce93` | Exact |
| 13 | 152 | 279 | `883e32b2b6ba9c037c028f236ccb8a649f807b72ce6175d0618df664dca76859` | Exact |
| 14 | 156 | 284 | `27ea97f1caf7d305702e7eb9f13bd50871376854295a0d41a0d8ebd4f83de5fc` | Exact |
| 15 | 183 | 312 | `60d1c026ead15317d7c2ffd5dcd16266095b8eb5bc58c85a980c9e77da66f81c` | Exact |
| 16 | 193 | 323 | `c0c5f50735232d427ea9cef1ecbe99ee04627adbca64a5cd4b20997238da83d1` | Exact |
| 17 | 206 | 337 | `72374927886f76cc2b6f86af78a6078c7222cb4840bec514216e709a0869d01f` | Exact |
| 18 | 224 | 356 | `a739cc374b67826a6246380e11db92c758339c03d09c518a75ac9987169ca48e` | Exact |
| 19 | 234 | 364 | `d19a57ffee6105f823f1b1bb4832c9c47f8bd723e12082477c5970481dca53d6` | Exact |
| 20 | 241 | 372 | `0bec3536b6f9481f927b240d348b26a84ba9a253521dd0318a336a69ae21858d` | Exact |
| 22 | 254 | 381 | `386b4265ebbf677827ec406556e749acd88440fdfc65e8d2f0a0cb2565edacea` | Exact |
| 23 | 266 | 394 | `7f137182cdf52ae8d58e4535cdd2ad13767c030b8d4b925587d41180f344ad70` | Exact |
| 24 | 282 | 411 | `c5ccdd8548be6f089a451c8428efd859d80b91b48f873824cd19c5d541461ef5` | Exact |
| 25 | 293 | 422 | `d3acacae0caf33328e002de8b7da71c5d74631dad4e6a3ce582eac61cf4030fb` | Exact |
| 26 | 301 | 431 | `4ff0e556815c16af0fe7dea82d67a918eafe4ed726c770e8cbb41e15de989cc0` | Exact |
| 27 | 305 | 436 | `f4726364d172690c05e390f3bc39bc3ccb526676b14a186bd122be5250b04889` | Exact |
| 28 | 309 | 441 | `16a9dbc987f440613646bf8774ee931e6b0016ab1cf90b7b6ac9a533b0b78788` | Exact |
| 29 | 314 | 447 | `76c11b46eecf2fdd16a3c41bb2bdc7e1fa3d54d73c20178ea427bfca820c71fb` | Exact |
| 30 | 329 | 463 | `a5d5bd164cf1f2f6c51899a747e525014518673040e52c809bc3dc5f4414fb3a` | Exact |
| 31 | 351 | 486 | `4769009b143643f4e49a0f954488e4512006b23d8d89f317d10ef656c7d62f5c` | Exact |
| 32 | 364 | 500 | `9fbc35c908c9b92f08799c97d9fc4d6052ccee5e1db81e5a14ed8c129847fbb4` | Exact |

## Task logbook

Both revision 1 findings are resolved. Preserve this corrected historical-attribution mechanism and its regression gate; whole-tag presence must not again be used as released-note evidence. LOGBOOK.md remains untouched per the explicit task constraint; this task-scoped outcome records the review decision and evidence before acceptance.
