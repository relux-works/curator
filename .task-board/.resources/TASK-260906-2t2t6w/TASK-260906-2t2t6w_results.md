# TASK-260906-2t2t6w results — launcher SPEC 0.4.1-draft (four 0.2.1-review minors)

Status: ready for review. `make check` exit 0 (build, fmt-check, vet, test, race; all 11 packages incl. `-race`).

Worktree: `.temp/STORY-260905-3l0fav/worktree` on `task-board/story/STORY-260905-3l0fav`, uncommitted.
9 files, +144/−25: SPEC.md, README.md, CHANGELOG.md, cmd/curator-run/main.go,
cmd/curator-run/main_test.go, cmd/curator-run/testdata/help.golden,
internal/execution/execution.go, internal/execution/execution_test.go,
internal/composition/probe.go (comment only).

## Minor 1 — §4.1/§6 resolve pass-through for the widened --repair surface

Before: every unmapped non-zero resolve exit collapsed into
`resolve_invocation_failed`, whose §6 gloss led with "curator not startable";
no clause named the pass-through for the --repair failure surface
(environment_marker_invalid, environment_surface_unmanaged_conflict,
environment_backup_exists, environment_seed_unreadable).

After:
- SPEC §4.1 failure-mode list gains the pass-through clause: the widened
  surface is not mapped, every such exit is `resolve_invocation_failed`,
  and "Curator's own diagnostic code and message are printed verbatim with
  the launcher's code line". §6 resolve-row gloss fixed to match ("a
  non-zero exit with an unmapped diagnostic, Curator's own code and
  message passed through verbatim"); the six code/condition clauses stay
  in matching order. No new mappings added, as briefed.
- Ordering decision (deviation from the finding's letter, recorded):
  the finding asks for the foreign bytes "beneath the launcher's code
  line, exactly as ax_handoff_failed does". The implemented resolve
  transport streams Curator's stderr live *before* the launcher's line
  (fragment.ExecRunner; pinned by TestRunDiagnosticsContract
  "forwarded verbatim ahead of it" and the TestRunResolveFailuresExit1
  prefix assertions). Buffering to print after would destroy live
  streaming and contradict the pinned contract, so the SPEC clause
  states the streamed-before form explicitly and names it the resolve
  transport's form of the ax pass-through. Behavior (verbatim code and
  message reach the operator) satisfies the minor; byte position differs
  by transport and is now documented.
- Tests: TestRunResolveFailuresExit1 gains the four unmapped cases
  (verbatim prefix + exit 1 + the unmapped code named in the launcher's
  own detail via a new wantInDetail pin); TestRunDiagnosticsContract
  gains "resolve invocation failed unmapped" (exactly one diagnostic
  line). Implementation (fragment/resolve.go) unchanged.

## Minor 2 — §4.6 pre-launch check order

Before: SPEC grouped the three checks with no precedence while §6 admits
exactly one line; implementation ran §5 probe → binary → stat.

After:
- SPEC §4.6: "run in this order, immediately before the handoff or exec,
  in both modes, and the first failure is the one reported" (binary
  check first, then §4.5 stat, then §5.1 probe).
- internal/execution/execution.go Run reordered to binary →
  CheckLaunchBoundary → Boundary, with the order pinned in a comment;
  probe.go doc comment aligned. The nil-Boundary caller-contract guard
  stays first (programming error, not a launch check).
- New TestPrelaunchCheckOrder (both modes x 3 scenarios): all three
  failures at once reports exactly one line, exec_provider_missing;
  stat+probe reports mcp_layer_missing; probe-only reports
  sysprompt_file_unreadable. Single-failure suites
  (TestLateChecksBothModes, TestProductionLateChecks) still pass
  unchanged, confirming no single-check behavior moved.
- Order mutant: new test run against the pre-fix order FAILS (4
  subtests report sysprompt_file_unreadable first), passes after the
  reorder. Fix verified restored (git diff present, targeted test green).

## Minor 3 — §9 silent-absence residual

Before: the residual-window bullet bounded the wrong axis; the sibling
adapters' swallow behavior was unverified and unrecorded.

After: new §9 item — whether claude_code
(--mcp-config <missing> --strict-mcp-config) and opencode
(OPENCODE_CONFIG naming a missing file) proceed silently is unverified
at the pinned releases; if either does, the §4.5 stat rule generalizes
to it; per-adapter facts belong upstream in environments §7.8.
SPEC-only change.

## Minor 4 — TestSpecVersionPinned reads both documents

Before: the test compared specVersion to a literal; M2/M3-style doc
reversions exited 0.

After: the test reads ../../SPEC.md and ../../README.md and asserts
specVersion (now 0.4.1-draft) appears in both, then keeps the
--version dispatch assertion.

Mutant table (each mutant = one doc set fully back to 0.4.0-draft;
docs restored byte-identical, sha256 verified):

| # | Mutant | go test -run TestSpecVersionPinned | Failing assertion |
|---|---|---|---|
| D1 | SPEC.md → 0.4.0-draft | FAIL (exit 1) | "../../SPEC.md does not state specVersion 0.4.1-draft" |
| D2 | README.md → 0.4.0-draft | FAIL (exit 1) | "../../README.md does not state specVersion 0.4.1-draft" |

Both mutants killed; post-restore suite green.

## Version bump (three-file + goldens/history)

0.4.0-draft → 0.4.1-draft in SPEC.md line 3, §8, cmd/curator-run/main.go
specVersion, README.md (2 spots), help.golden line 1,
main_test.go want literal. New §8.1 row, SPEC changelog entry
(2026-09-22), CHANGELOG header + Corrected entry. Remaining
0.4.0-draft strings are historical only (§8.1 row, two changelog
lines). No §3/§4.3/§4.5 mode text touched (SPEC hunks: L1, §4.1,
§4.6, §6, §8, §8.1, §9, changelog); no 0018 permission-interface text.

## Verification log (real exit codes, this session)

- gofmt -l cmd internal: clean (after formatting the new test file)
- go build ./... + go vet ./...: ok
- Targeted: TestPrelaunchCheckOrder + TestLateChecksBothModes: PASS
  (both modes); order mutant vs pre-fix code: FAIL as expected,
  restored + green
- Targeted: TestSpecVersionPinned, TestRunResolveFailuresExit1,
  TestRunDiagnosticsContract, TestRunHelpGolden,
  TestRunInformationalFlags: PASS
- Minor-4 mutants D1/D2: both FAIL as expected, docs hash-restored
- make check: MAKE_CHECK_EXIT=0 (test + race across all packages)
