# Revision-2 validation evidence

Task: TASK-261013-2borgq — wave6-launcher-review-findings.

- git diff --check: exit 0.
- Revision-2 source audit: exit 0; 9/9 scoped files, 3/3 exact pins, 7/7 hygiene patterns, 20/20 consistency clauses, 101/101 local file targets, 4/4 figure-source structures.
- sh scripts/remote-gate.sh: exit 0, invoked directly without a pipe or tee.
- Required hosted lanes: Test (ubuntu-latest), Race (ubuntu-latest), Test (macos-latest), Race (macos-latest): 4/4 success.
- Test includes hosted gofmt, build, vet and test steps. Optional self-hosted lane skipped; not claimed as passing.
- Hosted result: https://github.com/relux-works/curator-agent-launcher/actions/runs/38372728944
- Validated snapshot commit: 9fa7acd22cd75e6066295c4a42f7eed94b274579.
- No local Go test/build or rendering. Prior Java probe exit 1 retained only as predecessor evidence.
- Release-pin changes are exactly lines 415, 416 and 418 of cmd/curator-run/main_test.go; build/spec pins unchanged.
- Eight-file documentation delta preserved. Work left uncommitted. No production string, schema or LOGBOOK.md edit.

CR handoff will run the configured gate on its frozen candidate and attach its
own bound log. This pre-handoff artifact does not claim that later result.
