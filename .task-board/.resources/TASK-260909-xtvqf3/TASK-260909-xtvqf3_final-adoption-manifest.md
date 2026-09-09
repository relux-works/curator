# TASK-260909-xtvqf3 final adoption manifest (accepted successor package, no new implementation)

Finalizes this artifact-only gate leaf on the checkpointed accepted successor
package from sibling TASK-260909-3d1589 (CR3 accepted, RUN-260909-18798b,
checkpoint-outcome.md). No code was written, no test added, no runner
rewritten in this task. Original xtvqf3 rev1/rev2 resources and the original
diagnostics candidate are preserved untouched.

## 1. Canonical accepted package (sibling board resources)

All bytes verified read-only on 2026-09-09 via `task-board resource get`
(public API) and `sha256sum`; mirrors below hash identically.

| Board resource (TASK-260909-3d1589) | sha256 | bytes |
| --- | --- | --- |
| TASK-260909-3d1589_rev2_gate-conformance_test.go | 8495946dce980b9e6a69a17f20fd40703917da51880f4aaa50d7b43f0da754e9 | 25080 |
| TASK-260909-3d1589_rev2_gate-framing_test.go | 5031d27fdcdb65c3a69ee723baa202d5f928335a26ac225a935b3763810b5f9e | 4394 |
| TASK-260909-3d1589_rev2_vectors.json | a404a9b703ea8f3fa585b4de45d04f5d8e9079b19129c44fc28164e286ddb76d | 14047 |
| TASK-260909-3d1589_rev2_run-gate.sh | 6b278b35604507f741ffd03f1f1e3b8ccf57d8c841f69573ab93e479686b224d | 23715 |
| TASK-260909-3d1589_rev3_adoption.md | 861431c62047aa5968813689a88afada9867bd7089e299260b724786fffb3845 | 7641 |

## 2. Self-contained mirrors on this task (byte-identical, basenames preserved)

Copied through the public resource API only
(`resource get` from sibling to /tmp, `resource add` onto this task).
Basenames are unchanged so the runner's sibling-filename guards keep working.
Re-fetching any mirror and hashing it reproduces the sibling hash in §1.

- TASK-260909-3d1589_rev2_gate-conformance_test.go (mirror on TASK-260909-xtvqf3)
- TASK-260909-3d1589_rev2_gate-framing_test.go (mirror on TASK-260909-xtvqf3)
- TASK-260909-3d1589_rev2_vectors.json (mirror on TASK-260909-xtvqf3)
- TASK-260909-3d1589_rev2_run-gate.sh (mirror on TASK-260909-xtvqf3)
- TASK-260909-3d1589_rev3_adoption.md (mirror on TASK-260909-xtvqf3)

## 3. Provenance / identity (narrow verification, not a rerun)

- Gate source tree (diagnostics CR2): `fbe90d5e60593a3a069721b2ad9e53cd071d8c02`
  (runner `TREE`; distinct from the artifact-only CR candidate tree below).
- Artifact-only Story base: `3ff66a9421ff6ddf675a49fc0c2868309f6e3de3`
  (worktree observed clean at this tip; `git status --porcelain` empty).
- Artifact-only CR candidate tree: `ff61be4a8bd43fa4ffb179d31aa38e41891d4313`
  (empty repository delta is the correct scope for this leaf).
- SPEC.md sha256: `5a7ccf0ba95708cb573a586977eb0ba4bad1e46233a540dc99784c1d4d922d48`.
- Eight pinned blob OIDs enforced by the runner before any test runs:
  `.scripts/diagnostics-mutants.sh 5ff00da1…`, `README.md 2fee2dee…`,
  `cmd/curator-run/main.go f2a23748…`, `cmd/curator-run/main_test.go 849e6aae…`,
  `internal/diagnostics/diagnostics.go 4baf919a…`,
  `internal/diagnostics/diagnostics_test.go dcfbc7e1…`,
  `internal/diagnostics/helpers_test.go a46ad729…`, `SPEC.md 997f0065…`
  (full OIDs in runner lines 83–93).
- Runner entry contract: the five §1 files must sit adjacent to the runner
  under their exact basenames (`CONF_SRC`/`FRAM_SRC` guards); zero selected
  tests, missing overlays, stale non-empty destination, baseline failure and
  surviving mutant all exit nonzero with a `GATE-FAIL` marker.

## 4. Supersession map for original findings (TASK-260909-xtvqf3_review-verdict-rev2.md)

Original verdict hash on this task:
`933c5af70d09da19bd3d48dc24882ba4a760be1b09011339cbbf5d8e0b6c82bb`;
evidence `3ce85369aa52357bc4e967b57d1c000b170ad18f1f4d0bf1902bd202bb97a353`.
Both stay attached; neither is edited.

- **R2 (medium, repeat: fixed-code owners lack wrapped/joined positives).**
  Superseded by the accepted package: single owner registry (5 owners) x
  single form registry (direct/wrapped/joined) = 36 positive assertions
  (30 mutable + 6 fixed) in `TestGateOwnerFormPositives`; 23 nil cases in
  `TestGateOwnerFormNil`; arithmetic pinned in `TestGateCoverageCounts`;
  runner kills four fixed-owner narrowings with named assertions
  (`usage joined lost`, `usage wrapped lost`, `axconfig joined lost`,
  `axconfig wrapped lost`) — the exact rev2 reviewer survivor plus its three
  siblings, not only the remembered example.
- **R1 (low, repeat: manual adoption block deletes its destination).**
  Superseded by the accepted adoption doc: no `rm -rf` anywhere; the manual
  entry resolves `GITROOT` explicitly and hands a fresh task-local
  `manual-work` destination the runner refuses when non-empty; the automatic
  entry adds the `|| exit $?` short-circuit so a primary-runner refusal
  cannot be masked by a passing `--self-check` (sibling rev3 R1 repair;
  manual block byte-identical rev2→rev3).
- Confirmed repairs carried over unchanged: published filenames preserved,
  fail-closed setup/identity/freshness, eight-owner… (10 mutable-Code slots
  + 2 fixed) foreign matrix with direct/wrapped/joined plus extras
  (44 pairs / 168 cases), exact single-detail framing mutant through
  `run -> Resolver.Resolve (ExecRunner) -> Emit/Line` with independent
  byte-exact companions, restoration byte-equal after every mutant.

## 5. Reused independent evidence (no full-gate rerun per finalization bounds)

- `TASK-260909-3d1589_review-verdict-rev3.md`
  (`9fc81165544de5438f8d5ad94baf4f311f1e3843553d6a83d1f27401c0989f14`,
  verdict **accepted**): exact automatic block exit 0 in a fresh clone at
  base 3ff66a9 (8 + 2 named gate tests, nine named narrowing mutants exit 1,
  self-check 5/5); populated destination exit 1 with data preserved;
  short-circuit-removal mutant unmasked (exit 0) proving the guard kills it.
- `TASK-260909-3d1589_review-evidence-rev3.txt`
  (`e65eddb0cf8b8dfdad5c3db43c3823ade834be063c17de29efbbc8a2a0995216`,
  55806 bytes): exact blocks, doc diff, hashes, commands, full outputs,
  preservation maps, replay script.
- `TASK-260909-3d1589_rev2_gate-evidence.log`
  (`b8fa6db4d62babd52fb3773bdd59a7a12903b63e46a244c0342066c76527b904`):
  rev2 runner behavior/mutant/negative evidence inherited by rev3 review.
- `TASK-260909-3d1589_checkpoint-outcome.md`
  (`4d201850bf43daf4b3d1a8a4b2b7fc6ffc3d7a5b1430cc5f49cbc47f31e8ef7c`):
  CR3 checkpointed, branch tip unchanged 3ff66a9, tree clean, no source commit.
- Prior rev1/rev2 verdicts+evidence on both tasks stay attached and are
  inherited, not rerun here.

## 6. Downstream instructions for TASK-260908-1wr53w (next diagnostics revision)

1. Stage the five §1 files (or the §2 mirrors — identical bytes, identical
   basenames) into `<git-root>/.temp/<task>/gate/` and `chmod +x` the runner.
2. Run the adoption blocks in `TASK-260909-3d1589_rev3_adoption.md` verbatim
   (manual entry for a fresh `WORK` dir; automatic entry with the `|| exit $?`
   short-circuit before `--self-check`). Expected: baselines exit 0
   (8 diagnostics + 2 framing named tests + package suites); all nine
   narrowings exit 1 on their named assertions; `--self-check` 5/5 exit 0.
3. Adopt into the candidate per that doc's checklist: replace hand-selected
   `strangers` with the derived foreign sets, add joined owned positives,
   owner/form registry positives and joined typed-nil cases, keep the framing
   matrix plus the fixed-binary main entry with independent companions.
4. Re-run gate + `--self-check` before requesting review; attach the log.
   CodeOf stays API-only classification in tree fbe90d5 (no non-test
   production caller); full-main launch integration is a separate obligation
   and is not claimed by this gate.
5. Gate acceptance here does NOT accept diagnostics CR2 and does not waive
   later code adoption/review/main integration or defaults/plan/Pi work.

## 7. Bounds respected

No production edits, no new tests, no rewritten runner, no second workflow,
no installs/tags/hosted CI/real ax-model calls/private-record writes,
no LOGBOOK entry, no commits or branch changes, no writes to the original
diagnostics candidate. Worktree left clean for an artifact-only CR.
Review routing: independent Astra review of this final leaf; developer does
not accept or Complete.
