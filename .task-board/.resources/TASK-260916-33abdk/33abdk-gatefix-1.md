# TASK-260916-33abdk — gate fix (THE ONLY CURRENT INSTRUCTION, with 33abdk-sec-brief.md)

Revision 1 (tree ade7c3f8) FAILED the hosted gate on every Test/Race lane (run 36299079208). The single failing test (from the gate's
go-test.json artifacts, all 5 lanes): cmd/curator TestEnvResolveKeepsSchema1BytesForMetadataOnly — env_credential_marker_test.go:186
"metadata-only upgrade changed schema-1 marker bytes": your change writes `"codex_seed_record": {"revision": "B", "native_mcp_servers": []}`
into the marker even when nothing was stripped, so a metadata-only resolve rewrites a schema-1 marker. Fix so that the marker bytes stay
unchanged when there is nothing to record (omit the field when empty / only write it on a real strip, per the rc.13 marker rules — cite the
clause that allows the new field and when), and keep the strip/report behaviour and your mutant evidence. Run `go test ./cmd/curator -run
'EnvResolve|Marker|Credential|Seed|Mcp'` and `go test ./internal/envprofile -run 'Seed|Mcp|Codex|Status|Guarded'` with captured exit codes.
Also: trunk moved to d41da0fb — `git fetch origin main`, combine keeping both sides; VERIFY `git diff --name-only origin/main -- . ':!.task-board'`
lists only your paths. `task-board m 'set_status(TASK-260916-33abdk, status=development)'` first; append "Revision 2 — marker bytes stable";
handoff; the hosted gate on the published revision must be green before you claim anything (campaign-producer-rules.md). If the handoff
command runs long, it is the gate — wait for it, do not interrupt it. No CHANGELOG/LOGBOOK edit.
