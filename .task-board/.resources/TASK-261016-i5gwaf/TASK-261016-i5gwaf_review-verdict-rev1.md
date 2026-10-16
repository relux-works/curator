# Review verdict — TASK-261016-i5gwaf: M2 Muse 1.4.8 exclusion replay

**REJECT / changes_requested.** Two P2 evidence-attestation bypass mechanisms
require localized rework. Both have `repeat-of: none`; there is no previous
review verdict for this CR revision in the task outcomes. The measured vendor
interactive failure is correctly documented and is not the reason for rejection.

- CR: `CR-TASK-261016-i5gwaf-1`, revision 1.
- Base: `0d00a109ad1da77a5506747c89f24af9442eb341`.
- Exact candidate tree: `0f67bfb2d484e4e1f32711d4d8b76012aca05680`.
- All 44 changed worktree files matched the candidate blobs before execution.
- Fresh upstream main equals the CR base. Normative spec main remains
  `d2a942cf17e4eb017c45475442da1456d8dfbe6c`; environments §§7.1, 7.9, 10.2,
  CIP-0017 leaf 2/minimal probe rows, and Muse v4 conformance were reread.
- Scope remains evidence/probe/documentation only. Candidate code was not edited.

## Swept surface and AC coverage

The brief supplies one declared surface, no catalog-keyed row table, and no
wall-clock budget. Its result was persisted before supplemental free hunt.

| Supplied surface | Result | Named attacks |
|---|---|---|
| Vendor-build evidence authorizing form-specific exclusion | broken | Duplicate/conflicting count and outcome fields; listed symlink and unlisted dangling link; missing-record, retained-skill, extra-entry, offline-loopback and exclusion-omission controls |

**6 of 6 in-scope AC observation rows have named driving tests**, checked before
reading implementation. All six were rerun through the real vendor/Curator
entrypoints (44 cases), with these limits:

| AC observation | Executed tests | Review result |
|---|---|---|
| Parser | TestMuse148ParserRefusals; LoaderMatrix positive controls | Both independent exit-2 refusals and accepted spellings reproduced |
| Native root / foreign rules / persistent skills | TestMuse148LoaderMatrix | Both forms, all root states, separate foreign sources reproduced |
| Required exclusion | LoaderMatrix; InteractiveExclusionRegression | Exec removal and interactive retained-skill FAILURE reproduced on both roots |
| Interactive boolean without native root | LoaderMatrix/interactive/absent | Rule fallback retained as recorded |
| Curator link and frozen snapshot | TestFrozenSnapshot; matrix; reference CLI | Real loader results reproduced; independent closure attestation broken by snapshot-link-attestation |
| Workspace residual | TestMuse148WorkspaceResidual | Real loader and /rules observations reproduced; model-visible precedence remains unknown |

Free hunt added the unlisted dangling-link variant of the same snapshot
mechanism; findings are deduplicated. No additional blocking mechanism.

## Validation

- `go test ./.research/muse148 -run '^(TestObservationChecker|TestFrozenSnapshot)$' -count=1 -v`: exit 0.
- Exact-build disposable offline `go test ./.research/muse148 -count=1 -v -timeout=5m`: exit 0, all 44 vendor cases; sandbox controls also pass.
- `go build -o .temp/review-muse148/curator ./cmd/curator`: exit 0. Fresh binary is byte-identical to the Curator executable used in replay.
- `go vet ./.research/muse148`: exit 0.
- Checker, snapshot-extra-entry and offline-loopback narrowing mutants: each exit 1 at the intended assertion.
- Exec omission on linked and snapshot roots: exit 1 with retained-skill assertion; both no-exclusion positive controls pass.
- Interactive omission on linked and snapshot roots: exit 0, expected survivor; no successful exclusion claim.
- Both review findings reproduced against exact candidate copies. The portable reproduction driver runs every case and reports actual/expected exit codes; its own exit 0 means the driver finished, not that candidate checks are sound.
- Full `sh scripts/remote-gate.sh` was started on signed snapshot
  `308faef12e3f6a43811fe5fb3b42a516ba0b3861`, whose tree equals the CR candidate.
  A task-local Git wrapper preserves the configured human identity and forces
  `commit-tree -S` because the existing gate script otherwise creates an unsigned
  snapshot; signature verification returned `G`. No Story commit/ref changed.
  [Reviewer CI run](https://github.com/relux-works/curator/actions/runs/38844928424)
  was intentionally stopped after the completed review sweep found two blockers.
  Local waiter exit 137 is an intentional stop, not a test failure or a pass.
  This fresh full gate is **unverified**, never green. Its terminal state is
  recorded in the attached evidence. The owned temporary gate branch was deleted.
- Producer hosted run 38835589308 was independently read as completed/success;
  default source/tests/config are byte-identical to candidate. That is reused
  evidence only, not a claim that the reviewer full gate passed or CI ran the hidden probe.

## Required verdict schema

```json
{
  "verdict": "changes_requested",
  "candidate_revision": 1,
  "candidate_tree": "0f67bfb2d484e4e1f32711d4d8b76012aca05680",
  "findings": [
    {
      "id": "contradictory-loader-attestation",
      "row": "vendor-build evidence authorizing form-specific exclusion",
      "invariant": "Missing, failed, truncated or contradictory loader observations must not attest successful exclusion (record Coverage map and replay_test.go:48-49).",
      "mechanism": ".research/muse148/replay_test.go:53-65 uses substring outcome matching and the first matching count field, admitting contradictory duplicate count/outcome fields. Production callers are TestMuse148LoaderMatrix and TestMuse148InteractiveExclusionRegression.",
      "reproductions": [
        {
          "test_file": "review_attack_test.go (embedded in TASK-261016-i5gwaf_review-reproduce-rev1.py)",
          "command": "go test ./<scratch>/checker -run '^Test(ObservationChecker|ReviewContradictoryObservation)$' -count=1 -v",
          "expected_failure": "TestReviewContradictoryObservation fails with contradictory record admitted for both user_skills=0 user_skills=1 and outcome=\"failed\" outcome=\"ready\". TestObservationChecker remains green.",
          "log": "TASK-261016-i5gwaf_review-evidence-rev1.md: reproduce-attacks-02.log"
        }
      ],
      "severity": "bypass",
      "priority": "P2",
      "repeat-of": "none",
      "required_change": "Reject conflicting/duplicate relevant fields rather than treating one successful token as sufficient. Add the two named cases to the existing checker test with a valid positive control and a narrowing mutant for this class. Keep actual vendor records and regression behavior unchanged."
    },
    {
      "id": "snapshot-link-attestation",
      "row": "vendor-build evidence authorizing form-specific exclusion",
      "invariant": "M2 requires an independently checked regular-file snapshot closure; linked, missing or unlisted entries cannot be attested as that closure.",
      "mechanism": ".research/muse148/reference-snapshot-check.py:10 follows symlinks via is_file/read_bytes/stat and omits dangling links before check_bundle. TestFrozenSnapshot at replay_test.go:437-440 builds a new root from testdata/root.md instead of validating testdata/snapshot physical entries, so that lane does not detect the malformed input either.",
      "reproductions": [
        {
          "test_file": "TASK-261016-i5gwaf_review-reproduce-rev1.py",
          "command": "SPEC_PYTHON=<spec-python> PINNED_SPEC=<pinned-spec> python3 TASK-261016-i5gwaf_review-reproduce-rev1.py",
          "expected_failure": "Regular positive must exit 0; replacing the listed root with a symlink or adding an unlisted dangling link must refuse. Actual reference CLI exits 0 and prints complete regular-file closure for both invalid inputs. TestFrozenSnapshot also exits 0 with the listed-link input.",
          "log": "TASK-261016-i5gwaf_review-evidence-rev1.md: reproduce-attacks-02.log"
        }
      ],
      "severity": "bypass",
      "priority": "P2",
      "repeat-of": "none",
      "required_change": "Validate the actual frozen fixture with no-follow entry classification and a complete physical inventory before passing bytes to check_bundle. Add listed-symlink and unlisted-dangling-link regressions beside the valid regular fixture and one narrowing mutant. This is a focused correction of the existing checker, not a new snapshot implementation or research task."
    }
  ],
  "notes": [
    "6 of 6 in-scope AC observations have named driving tests; all 44 exact-build vendor entry cases reran green as observation witnesses. Interactive exclusion remains a measured vendor failure, correctly retained as pending under Rework 1; it is not a blocking review finding.",
    "No catalog-keyed surface table or review/free-hunt wall-clock budget was supplied. The one declared leaf surface was swept; its result is broken. Supplemental attacks cover its observation grammar and fixture identity. No supplied row was omitted.",
    "No runtime qualification constant, gate or credential policy changes. Model-visible behavior, refresh and actual env materialize integration remain unmeasured.",
    "First portable reproduction-script run had a Python tempfile absolute-path setup error in its Go commands; that setup failure is not counted. The corrected second run reproduces both findings. Initial direct isolated reproductions independently confirm them.",
    "Reviewer code changes are confined to ignored scratch copies. LOGBOOK entry is supplied as a task outcome, not a candidate/control-root edit."
  ]
}
```
