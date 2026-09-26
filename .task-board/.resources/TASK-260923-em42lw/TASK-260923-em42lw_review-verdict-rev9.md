# TASK-260923-em42lw review verdict rev9 — ACCEPTED
Reviewer: claude-opus-5-5 low. Candidate: worktree tree = 0b7c2c38 (temp-index write-tree match), base 316438cc.
- Lattice at production entry: envprofile/managed.go buildFragment (lock→native/locked/global; profile knob→profile; else native/default). CLI rows TestEnvResolveEmitsPermissionsV2AtProductionEntry (yolo/native/absent/lock incl. §1 warning naming system file).
- Reran (zsh, pipefail): go test ./internal/envfragment ./internal/config ok; go test -run 'Permission|EnvResolve' ./cmd/curator ok (57s).
- Own mutant: locked branch mode := profile knob (breaks "mode=native when global") → KILLED at env_permissions_test.go:250. Restored, tree verified.
- No source-side pre-0018 accommodation found (grep). Gap ledger: overlay rows re-owned by ioemse only; schema2-permissions vector remains owned by 1cenbr (EffectiveJSON defaults, outside this scope) — residual.
- Windows: platform-control skip for full-schema validation only (POSIX absolutePath), lattice assertions still run (platform-cases.tsv:462-467, skip-classes.tsv:55).
- No CHANGELOG/LOGBOOK in delta (per rev9 note). Hosted gate green per orchestrator; full suite not rerun by me.
