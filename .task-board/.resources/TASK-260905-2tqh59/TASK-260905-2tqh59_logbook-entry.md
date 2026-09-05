# Logbook entry (board copy — LOGBOOK.md must not be written into curator-spec per the producer brief)

## 2026-09-05

### 2230 — Marker positive fixture recorded a copy outside its paths
- ANOMALY: `schema-cases/agent-environment-marker-v1/valid-linked-symlink-fallback.json` carried `copies: [{path: "skills/pdf", …}]` on a skills surface whose `paths` was `[]`; nothing enforced copies ⊆ paths (batch-2 reviewer F2), so the contradiction shipped as a positive case.
- ROOT CAUSE: `tools/generate-vectors/environments.go` `environmentMarkerSchemaExamples` set only `copies` on the linked variant of `validEnvironmentMarkerV1`, whose skills surface has empty `paths`.
- FIX: fixture lists `skills/pdf` in `paths`; `tools/validate.py` wire semantics now reject a copy outside its surface's paths; generated negative `invalid-copy-outside-paths`. Commit `fcdb9ba` on `draft/environments-1-1-follow-ups`.
- STATUS: resolved; `make validate` and `make regenerate-check` exit 0.

### 2231 — Enum cross-check has a self-blind layer
- FINDING: a mutant relaxing the new §12.1 enum comparison in `tools/validate.py` to a subset test passes the bare `python3 tools/validate.py` run (exit 0); only `test_validate.ManagerConfigVectorTests.test_widened_enum_fails` (run by `make validate`) kills it.
- NOTE: the bound is stated in `TASK-260905-2tqh59_drafting-report.md` (mutant M7). `make validate`, not `validate.py` alone, is the gate to cite.
