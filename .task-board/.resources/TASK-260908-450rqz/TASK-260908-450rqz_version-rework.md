# F1 version rework — TASK-260908-450rqz

Ready for review; independent reviewer acceptance remains mandatory.

This rework supersedes the original version choice. Reviewer RUN-260908-c1ad07 withheld acceptance only for F1: the settled goal requires 0.3.0-draft. Exactly six current occurrences changed in SPEC.md, README.md, main.go and main_test.go. Historical rows and the Pi E1/E2 explanation remain intact. All seven candidate paths were compared against CR1 tree 9ff5c7588e3ef9705be5bfa92d20d321f4619e07; only these six substitutions differ. No new behavior, dependencies, stages or commits.

## Checks run in this rework

| Command | Real exit | Evidence |
|---|---:|---|
| go test ./cmd/curator-run ./internal/mapping -run 'TestSpecVersionPinned|TestRunInformationalFlags|TestRunMapping|TestRunUnknownResolvedMapping|TestRunUnknownFragmentStillRefusesResolution|TestRunResolveFailurePrecedesMapping|TestResolve|TestKnownUnsupportedIsNotUnknown' -count=1 -v | 0 | focused-01.log |
| go run ./cmd/curator-run --version | 0 | version-01.log: curator-run 0.0.0-dev (specification 0.3.0-draft) |
| git diff --check | 0 | diff-check-01.log |
| Python exact CR1 comparison and executable-version assertion | 0 | scope-check-01.log |

Configured make check is delegated to the mandatory publication handoff once; its actual result will be captured in the new CR validation resource. No separate broad suite or mutant rerun was warranted by the six literal replacements. The accepted prior behavioral evidence below is retained, not claimed as rerun. Historical command results describe the original producer run. Behavioral coverage remains 9 of 10 AC rows driven, with the native-tail downstream bound unchanged. Tests remain uncommitted candidate files for parent-owned signed delivery.

Initial discovery included rg against absent optional skill directories (exit 2); this was not a validation gate. No failed read was used as evidence of passing validation. No LOGBOOK/control-root writes per explicit scope.

## Retained original producer evidence (version statement corrected)

