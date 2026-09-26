# TASK-260923-2elcdc review verdict rev1 — ACCEPTED
Reviewer: claude-opus-5-5 low. Candidate tree 18d8491a verified identical to worktree (git diff --quiet); only 8 product/test paths, no CHANGELOG/LOGBOOK.
## Reruns (zsh, pipefail, real rc)
- go build ./... ; go vet cmd/curator internal/config internal/envregistry: clean
- go test ./internal/config ./internal/envregistry: ok rc=0
- go test -run 'TestEnvResolve|TestEnvMigrate|TestEnvStatus' ./cmd/curator: ok (292s) rc=0
## Rules via CLI entry (cmd/curator env resolve / env migrate)
- isolated admitted in system-config-v2 (parseIsolation systemOnly removed; config.go applySystem) — TestEnvResolveIsolatedSystemLockUsesDirection
- silence → locked isolated (no auth passthrough) — same test
- explicit shared under lock → environment_isolation_lock_conflict, exit fail, names system file (envregistry.go EffectiveIsolation) — TestEnvResolveExplicitSharedConflictsWithIsolatedSystemLock
- provisioned shared passthrough fails closed, names F-C2 `env migrate --plan/--apply`, link+native bytes preserved; migrate now uses machineFromConfig (envmigrate.go:65) — TestEnvResolveLockedIsolationRequiresMigration
- pinned vectors: valid-isolation-{shared,isolated}-direction.json driven through Load (isolation projection; bound: other locks in those cases out of scope)
## Mutants (disposable rsync copy)
- M1 explicit shared admitted (lock check disabled): KILLED (ExplicitShared test FAIL)
- M2 silence resolves to shared (machine.Isolation = UserIsolation): KILLED (2 tests FAIL)
- silent-migration mutant not run by reviewer (bound; test asserts link preserved on refusal).
## Residuals
- Only whole-map lock (manager §1 locked list granularity) supported; Windows/linux platform lanes unverified locally — hosted gate arbiter.
