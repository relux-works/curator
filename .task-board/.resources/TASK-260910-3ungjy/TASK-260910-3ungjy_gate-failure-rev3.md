# TASK-260910-3ungjy — hosted gate failure on Change Request revision 3 (run 35251729913)

Extracted by the orchestrator from the CI evidence. Linux and macOS lanes are
green; windows-latest fails exactly four tests, all new posture tests in
`cmd/curator/hook_posture_test.go`:

```
TestStatusRecordedButMissingStaysInInventory      hook_posture_test.go:115 "status --check --json omits the missing row"
TestStatusUnreadableCandidatesKeepRecord/recorded
TestEnvStatusMissingAndUnreadableKeepRecord
```

The JSON documents in the failure output DO carry the rows (`"path":
"C:\\Users\\runneradmin\\...\\.agents\\env.sh", "state": "approved",
"approved_by": "operator", "file": "missing"`). The assertions fail because
they substring-match the raw JSON text against the native path
(`strings.Contains(stdout, envPath)`): on Windows the JSON encoder escapes
every backslash, so `C:\Users\...` is not a substring of `"C:\\Users\\..."`.
Production behaviour is correct; the harness is wrong on Windows.

Fix: in every `--json` assertion decode the document (the file already has
`decodeTrustDoc`) and compare the row's `path`/`file`/`state` fields to the
expected values, exactly as the first JSON block of
`TestStatusRecordedButMissingStaysInInventory` already does; keep the text-mode
substring checks (text output is not escaped). Do not fold or lower-case
paths, do not skip on Windows.

Re-run `go test -count=1 ./cmd/curator/...` and hand off again; the runtime
re-runs the hosted gate.
