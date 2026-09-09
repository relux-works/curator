# TASK-260909-xtvqf3 republish provenance — ordinary producer after public release (RUN-260909-ee8fbe)

Role: developer (implementer), Muse Spark xhigh. Brief: gate-republish-after-pr199.md
(supersedes release-only and gate-development briefs). No gate redevelopment, no rerun
of gate research. Runtime publication validation belongs to the configured runtime at
handoff; the whole suite was deliberately NOT run manually first.

## 1. Release outcome read first

Read `TASK-260909-xtvqf3_release-outcome.md` (sha256
489b6500fdb8595d8db0dc10067093f85704ed79bcc8015b611e2c540d67017e):
public invalidate-acceptance in terminal RUN-260909-cdae9b returned
`{"change_request_state":"accepted","revision":3,"status":"to-dev"}` (exit 0),
original accepted rev3 bytes unchanged (sha256
693e8797d0383a4b50e42c6b34c8783bdd27e08fee79abf3345761786bc8ddcc),
sibling TASK-260909-3d1589 CR3 checkpointed empty. Installed binary
`task-board version 0.24.3-335-gac9044a5 (commit ac9044a5, built 2026-09-09T16:35:07Z)`
re-verified here via `task-board --version` (exit 0). Matches reviewed PR199 source.

## 2. Accepted history preserved byte-for-byte (read-only, all exit 0)

Re-hashed frozen control root
`.temp/changerequests/TASK-260909-xtvqf3/` — all identical to release-outcome §1/§3:

- rev-000003.json: 693e8797d0383a4b50e42c6b34c8783bdd27e08fee79abf3345761786bc8ddcc (1526 bytes)
- current.json: 1e073642b852f1acd9632acc462e5f5a340b6df05e5a57e5436ae22bbc7ae1c7 (105 bytes, still revision 3)
- events.jsonl: 889f55e0038372eccae7457235f73fe02360349c67495aa8f7b8fee9498be7b8 (1009 bytes, still ends accepted by RUN-260909-2fc911)
- rev-000001.json: fe3af263e187d840426bf5105a10d5ae992357d1519badd08238be8e5e8d5c5d
- rev-000002.json: 28a435125cc264ebb3595080055fa81d6580b3f3bd74c6f1aabab4d34bca7179

Commands: `sha256sum <frozen-root>/*.json*` (exit 0), `wc -c` (exit 0).
No CR kind set, no private-record create/edit, no acceptance transfer.

## 3. Gate package verified against accepted evidence (read-only, all exit 0)

Five self-contained mirrors on this task are byte-identical to the canonical
sibling TASK-260909-3d1589 resources and to final-adoption-manifest.md §1
(`sha256sum`, `wc -c`, exit 0):

| file | sha256 | bytes |
| --- | --- | --- |
| TASK-260909-3d1589_rev2_gate-conformance_test.go | 8495946dce980b9e6a69a17f20fd40703917da51880f4aaa50d7b43f0da754e9 | 25080 |
| TASK-260909-3d1589_rev2_gate-framing_test.go | 5031d27fdcdb65c3a69ee723baa202d5f928335a26ac225a935b3763810b5f9e | 4394 |
| TASK-260909-3d1589_rev2_vectors.json | a404a9b703ea8f3fa585b4de45d04f5d8e9079b19129c44fc28164e286ddb76d | 14047 |
| TASK-260909-3d1589_rev2_run-gate.sh | 6b278b35604507f741ffd03f1f1e3b8ccf57d8c841f69573ab93e479686b224d | 23715 |
| TASK-260909-3d1589_rev3_adoption.md | 861431c62047aa5968813689a88afada9867bd7089e299260b724786fffb3845 | 7641 |

Manifest `TASK-260909-xtvqf3_final-adoption-manifest.md`: ba6a3fd47ac36c499e64905c179bde5897cb0cb514a9dfa4191a47b8334ff02f
(8155 bytes) — matches the hash quoted in review-verdict-rev3.
CR3 verdict: accepted (final adoption of checkpointed sibling package; five-owner
registry, fail-closed runner, exact staging path and typo clarifications).
Behavioral evidence (baseline exit 0, nine named narrowing mutants exit 1,
five runner negatives, framing companions, restoration) is REUSED from sibling
rev3 acceptance and earlier rev1/rev2 evidence per finalization bounds — no
full-gate rerun here, no Go toolchain invocation, no `make check` by hand.

## 4. Historical vs actual revision identity (must not be conflated)

- Historical artifact verification: diagnostics CR2 source tree
  fbe90d5e60593a3a069721b2ad9e53cd071d8c02 (runner TREE pin); artifact-only CR
  candidate tree ff61be4a8bd43fa4ffb179d31aa38e41891d4313 with empty repository
  delta (zero-byte patch e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855);
  artifact-only Story base 3ff66a9421ff6ddf675a49fc0c2868309f6e3de3 (CR rev3 base_oid).
- Actual derived revision for THIS run: produced by the configured runtime at
  handoff/publication from the current Story branch tip 289ff42f037b
  (worktree clean, `git status --porcelain` empty, exit 0; HEAD 289ff42f037b;
  merge-base with origin/main is 289ff42 itself — branch replayed onto trunk per
  progress record). Worktree HEAD tree: 489e695df7ada7598347233c6553b791dacebb60.
  `worktree status` (exit 0) shows CR-TASK-260909-xtvqf3-3 revision 3 accepted
  task_delta empty carried forward, lease held by this run RUN-260909-ee8fbe
  (own lease, not a foreign block; no gc needed, no lease files touched).
- Actual validation: `make check` runs EXACTLY ONCE under the configured runtime
  during publication — not executed in this shell. Prior rev3 validation log
  (exit 0) is historical evidence only.

## 5. Bounds respected

No source edits or commits (`git status --porcelain` empty before and after;
`git diff HEAD --stat` empty). No installs, restarts, hosted CI, tags, releases,
real ax/model calls, runtime-home edits, LOGBOOK or control-root source writes.
No board-state push, no checkpoint/complete, no generic handoff. Explicit
resource scope only; diagnostics candidate untouched. Gate remains artifact-only;
diagnostics is already delivered — this acceptance path does not accept
diagnostics CR2 nor waive adoption/review/main integration.

## 6. Handoff and routing

Normal developer handoff (`task-board handoff TASK-260909-xtvqf3 --role developer`)
publishes the new derived story_final revision through the runtime. Parent routes
a NEW independent Astra medium review, then bound Complete. Developer does not
accept or close. Checklist (12/12 done, incl. CR3 mirror/evidence/adoption item)
and this task-scoped outcome satisfy handoff evidence.
