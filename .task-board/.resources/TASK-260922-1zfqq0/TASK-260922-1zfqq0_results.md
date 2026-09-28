# TASK-260922-1zfqq0 — results

Status: ready for review. This is the F-L1a specification/documentation leaf; permission runtime behavior and its behavior tests remain with F-L1b.

## Section-by-section changes

| Area | Before | After |
|---|---|---|
| SPEC §3 | No typed permission mode | Added `--permissions <native\|yolo>`, exact `--yolo` alias, same-increment semantics, and pre-boundary rejection of `-d` / `--danger`. |
| SPEC §4.1 | Consumed only the v1 fragment | Cites F-S2 (`launch-env-fragment-v2`) as the minimum transport token; defines the required closed `permissions` object and rejects contradictory fields. Legacy v1 can proceed only when the resolved mode is native. |
| SPEC §4.3 | Defaults v1 covered model and effort | Documents defaults v2 and v1 compatibility; adds flag > profile > launcher-global > interactive built-in > headless built-in precedence, `source=default` silence, and the untracked provenance line. Clarifies that provenance `source=global` means launcher `defaults.json`, while fragment `permissions.source=global` means Curator's lock. |
| SPEC §§4.4–4.5 | Interactive plan carried model/effort only | Names agents-management v0.5.18 request members and `permission-grammar-v1`; passes the resolved mode through `LaunchRequest.PermissionMode`, leaving mapping and permission argv grammar with agents-management. |
| SPEC §4.6 | No permission-specific detector or refusal rules | Defines the single closed/versioned headless marker set, native headless/CI/tracked silence, the force-native lock rule, tracked refusal without fallback, and the exact effective-native-policy stderr line and tracked record key. |
| SPEC §4.7 | File family omitted the new mode | Records defaults-v2 ownership of the launcher-global permission default and that Curator's per-profile knob travels only through the fragment. |
| SPEC §6 | Permission diagnostics absent | Adds `permission_policy_unsupported`, `permission_mode_tracked_unsupported`, `permission_mode_unsupported`, and their exit-1 conditions; usage cases state exit 2. |
| README / CHANGELOG | SPEC 0.4.1-draft and no permission options | Updates the options/configuration/diagnostic documentation, 0.5.0-draft references, and unreleased CHANGELOG entry. States that F-L1b updates the implementation dependency to v0.5.18; this documentation-only leaf does not change `go.mod`. |
| Version pins | 0.4.1-draft | Bumps `specVersion`, its pinned test expectation, and the help golden header to 0.5.0-draft. No permission implementation code changed. |

## Choice 4 contract

- Stderr line: `curator-run: effective-native-policy: relaxation=<selector[,selector...]> source=<settings-source>`.
- Tracked launch-record key: `works.relux.curator.effective-native-policy`; it records the same sorted selectors and source identifier.
- Untracked mode has stderr provenance only and creates no persistent record.

## Source contracts cited

- F-S2, curator-spec `ec8dc656`: minimum token `launch-env-fragment-v2`; required `permissions` object has `mode`, `locked`, and `source`; `locked` iff source is `global`; mode is `native` for `global` and `default`; `source=default` means the profile level is silent.
- F-M1, skill-agents-management v0.5.18 (`149569d`): `LaunchRequest.PermissionMode`, `LaunchRequest.ToolRelease`, `LaunchRequest.NativeArgs`, `permission-grammar-v1`, and `ErrPermissionModeUnverifiedRelease`.
- Decision 0018 choices 1, 4, 5, and 7 and Decision 0013 D5: the launcher resolves and passes the mode while agents-management owns provider mapping and permission grammar.

## Evidence

| Check | Command / result |
|---|---|
| Documentation absence grep | `rg -n -- '--dangerously' SPEC.md README.md CHANGELOG.md` — exit 1; no matches (expected grep no-match status). |
| Headless marker count | `rg -c -- 'GITHUB_ACTIONS' SPEC.md` — exit 0; one occurrence. |

- Focused test (zsh): `set -o pipefail; go test ./cmd/curator-run -run 'TestSpecVersionPinned|TestRunHelpGolden' -count=1` — exit 0.
- `make build` — exit 0.
- `make vet` — exit 0.
- `make fmt-check` — exit 0.
- `git diff --check` — exit 0.
- `make check` was not run locally. Campaign producer rules route the full landing suite to hosted CI exactly once at handoff; local verification used the focused tests and build/validation commands above.

The managed Story workspace records the protected `main` authority with advertised, fetched, and selected base OID `8c5b0495252d742d1ee27cc76b7f569ece79ec48`; the upstream preflight was fresh and the worktree was clean before edits.

## Revision 2 — diagnostics declarations and conformance gate

Revision 1's landing gate exposed a real SPEC-to-registry mismatch: SPEC §6 contained 21 normative codes while `internal/diagnostics.Codes()` and `TestGateCoverageCounts` still expected 18. The revision-1 brief had incorrectly excluded the declaration changes required by this repository's conformance gate.

- Added `permission_policy_unsupported`, `permission_mode_tracked_unsupported`, and `permission_mode_unsupported` to the diagnostics declarations and `Codes()` in SPEC §6 order. They use the existing operational exit rule (exit 1); no permission resolution, refusal, or transport behavior and no new `CodeOf` error owner were added. Those runtime owners remain F-L1b's work.
- Updated `TestGateCoverageCounts` to require 21 normative codes. The five existing typed owner rows and their 36 direct/wrapped/joined positives remain intact; foreign-code rejection now covers 53 normative owner/code pairs and 195 cases across the three forms.
- Preserved the revision-1 SPEC, README, CHANGELOG, version pins, and help golden.

| Check | Command / result |
|---|---|
| Focused diagnostics tests | `go test ./internal/diagnostics -count=1` — exit 0. |
| Full local validation | `make check` — exit 0; build, vet, full tests, and race tests passed. |
| Documentation absence grep | `rg -n -- '--dangerously' SPEC.md README.md CHANGELOG.md` — exit 1; no matches (expected grep no-match status). |
| Headless marker count | `rg -c -- 'GITHUB_ACTIONS' SPEC.md` — exit 0; one occurrence. |
| Choice-4 spelling | `rg -n 'effective-native-policy|works\.relux\.curator\.effective-native-policy' SPEC.md` — exit 0; line and key present. |
| Diff whitespace validation | `git diff --check` — exit 0. |
