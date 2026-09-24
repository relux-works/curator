# TASK-260916-2ok97n brief (orchestrator, binding) — R5 qualify script-worker runtime conformance (FINAL leaf)

Story STORY-260822-2h0v9j; R1–R4 checkpointed on the Story branch. This is the Story's LAST leaf:
on ACCEPT the orchestrator integrates the whole Story (base refresh onto current trunk happens at
your spawn; the runtime replays the checkpoints). Read the four results.md files and every review
verdict (R1 rev2/rev4, R2 rev3, R3 rev5/6/7, R4 rev1): their residual lists are your input
(R-e stream model / pass-through binding; R-d Windows qualification with real `python.exe`/
`node.exe` and a declared `exec`; R2 R-b/R-c/R-h rows if still open; test hygiene R-m; the
`internal/snapshot` Windows flake note).

## Scope (R5 exactly, per the reconciliation R5 and the Story AC)
- Replace every remaining "unreachable"/"not implemented" classification in
  `internal/scriptpolicy/conformance_test.go` with real consumers: all 12 vector top-level keys of
  `script-host-execution-policy.json` classified as consumed (`execution_policy`, `interpreters`,
  `opt_in_cases` 6 preserved, `capability_derivation_cases` 4, `mandatory_controls` 11,
  `native_control_inventory`, `capability_evidence_record`, `capability_evidence_cases` 14,
  `preflight_cases` 5, `audit_label_cases` 4, `protocol_version`, `schema_version`), each pointing
  at the production-entry row that drives it — a ratio line 33/33 named behavioural cases + 11
  controls, no "helper-level only" entries.
- Register every real launch/control/evidence/audit test in `.github/ci/platform-cases.tsv` with
  the must/skip shape per lane (linux/darwin/windows) and the ledger vocabulary; the Linux
  cgroup/Landlock/netns rows `must=linux`, Windows job-object rows `must=windows`, macOS rows
  `must=darwin`; no new skip class.
- Narrowing negative tests for the refusal/attestation gates that still lack one (from the
  verdicts' residual lists), each killing a named mutant.
- Host evidence: the three hosted lanes (ubuntu/macos/windows) prove the rows through the gate;
  rose-air (self-hosted macOS arm64, Xcode present since 2026-09-21) — the gate runs it only on
  main pushes, so report it as "to be observed on the landing run" and name the rows it will add;
  Windows qualification with REAL interpreters where the runner has them (`python.exe`,
  `node.exe` on windows-latest) for at least the happy path + one refusal, else state the exact
  bound (R-d).
- Stream model (R-e): implement pass-through stdio binding if it fits the budget as a bounded
  protocol change with rows; otherwise document the bound precisely in docs (interactive/large
  output commands) — decide early and say which.

## Rulings
R1 No proof weakened, no test deleted/skipped to reach 33/33; classifications may only change from
unreachable/not-implemented to consumed when a production-entry row exists.
R2 Admission honesty end state: production launches enforced node-v1/python3-v1 scripts only
where the probe finds every mandatory control available; the seam of R2/R3 is gone (row proves
production cannot open the path otherwise).
R3 Docs: `docs/script-interpreters.md` + troubleshooting complete for every refusal class the
worker emits; CHANGELOG `### Added` summarising script-worker-v1 support and its platform
matrix (which controls are available/host-conditional/unavailable per OS).
R4 Budget: results.md in the first 10 minutes and updated as you go; publish an honest partial if
the 150-minute budget runs out (the orchestrator continues you).
results.md: classification table (12 keys → rows), ledger delta, mutant table, platform matrix,
rose-air expectation, bounds. Publish only on a green gate.

## Addendum 2026-09-23 (orchestrator)
- Trunk is now `48da2690` (rc.12 conformance pin: `SPEC_PIN` = `dced9b8`). Your base refresh replays
  R1–R4 onto it; `CHANGELOG.md` carries `merge=union` in `.git/info/attributes`. If a vector row that
  previously skipped now EXECUTES because of the rc.12 root, name it in results.md.
- The rose-air lane is currently red on main for an unrelated runner reason (BUG-260922-k6eypp, being
  diagnosed by BUG-260922-306v4m); do not treat a rose-air failure as yours — report it as unverified.
- The Windows `internal/managerlock` tiny-deadline flake (BUG-260922-6chzf9) is known and unrelated;
  if the gate fails only on it, say so with the test name and the run id.
