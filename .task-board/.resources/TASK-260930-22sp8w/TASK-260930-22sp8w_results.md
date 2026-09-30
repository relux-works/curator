# TASK-260930-22sp8w — isolate test git config: results

## Change
- `.github/ci/test-gate.sh`: exports `GIT_CONFIG_NOSYSTEM=1` and `GIT_CONFIG_GLOBAL=$EVIDENCE/isolated.gitconfig` (an empty file the gate creates) before every go test stage, and logs both (`test-gate: GIT_CONFIG_GLOBAL=... GIT_CONFIG_NOSYSTEM=1`).
- New `internal/testgitenv.Isolate()`: sets the same two variables process-wide and returns a cleanup func. If it cannot create the empty config it exits the process (it never falls back to the ambient config).
- TestMain calls `Isolate()` in every package whose tests run git. That is 22 new `gitenv_main_test.go` files plus the existing TestMains in cmd/curator (before the CLI-helper branch), rustsource, godriver, scriptworker and install. A sweep of `*_test.go` for git exec finds no git-running package without the call.
- Tests that set their own global config still work because t.Setenv overrides the process value. In `buildHTTPSGitHome` (cmd/curator), buildrepo httpsbroker, envprofile network fixture, gitcred and status_test, tests that used to rely on `HOME/.gitconfig` now set `GIT_CONFIG_GLOBAL` to that file explicitly.
- Regression row `cmd/curator/gitenv_hostile_test.go` `TestSignerRowsSurviveHostileGlobalGitConfig`:
  - Hostile config: commit.gpgsign, tag.gpgsign, gpg.format=ssh, user.signingkey set to a generated throwaway key. It is planted in GIT_CONFIG_GLOBAL, HOME/.gitconfig and XDG, and system config is re-enabled.
  - Precondition: a probe commit made under that config is signed.
  - The row then re-execs the test binary for `TestSourceSignerVectorsAtProfileInstall` and `TestRevisionDoesNotBorrowTagSignature`, and requires `--- PASS` for `unsigned-refused` and for `TestRevisionDoesNotBorrowTagSignature`.
- `gate-selftest.sh`: the synthetic go test runs under a hostile XDG global config, with both variables unset. The selftest checks three things:
  - go test received `NOSYSTEM=1`;
  - go test received an empty gate-owned `GIT_CONFIG_GLOBAL`, not the hostile one;
  - the gate logged the line.

## Evidence (real exit codes; CURATOR_CONFORMANCE_ROOT = local curator-spec/conformance/v1)
- `go build ./...` → 0; `go vet ./...` → 0; `gofmt -l cmd internal` → empty.
- `go test -count=1 -run '^TestSignerRowsSurviveHostileGlobalGitConfig$' -v ./cmd/curator` → 0 (PASS, 19.7s).
- Mutant: in cmd/curator TestMain, replace `testgitenv.Isolate()` with a no-op, then run the same row → exit 1. The child run fails `TestSourceSignerVectorsAtProfileInstall/unsigned-refused` (`context_source_signer_rejected`) and `TestRevisionDoesNotBorrowTagSignature`, the same failures as on the runner. KILLED. The file was restored afterwards.
- `bash .github/ci/gate-selftest.sh` → 0 (295 passed, 0 failed, including the 3 new rows).
- Package runs, all exit 0:
  - `go test -count=1` over buildrepo, gitcred, audit, closure, config, gitignore, gitops, identity, manifest, marker, snapshot, sourcelock, skillspec, scriptpolicy, envmarker, devsub, contextpkg → all ok.
  - That same run → exit 1 overall, but only because envprofile hit the default 10m timeout. The rerun `go test -count=1 -timeout 40m ./internal/envprofile` → 0 (1833s).
  - `go test -count=1 -timeout 40m` over godriver, install/atomicity, interop/environments, scriptworker, swiftpmbuild, swiftpminterop, swiftpmsource, yarnmodernsource, crossconformance, rustsource → 0.

## Not run locally
- The full `./cmd/curator` and `./internal/install` package suites (multi-hour on this host). Only the new row and the mutant ran in cmd/curator. The hosted gate is the arbiter for those suites.
- The full `test-gate.sh`. Its export behaviour is covered by the gate-selftest rows.

Out of scope: macbook-iv Rust failures (no /var/db/xcode_select_link). That is host setup for the operator.

## Revision 3 — rev2 re-applied unchanged on b4b08a1993d240b5dd24d6929cafe2b51e72637c

- `git diff bdb77413 refs/campaign/12oimr-rev2-20260930 -- . ':!.task-board' > $TMPDIR/gitiso.patch && git apply --3way $TMPDIR/gitiso.patch` → exit 0 (gate-selftest.sh and test-gate.sh "Applied patch ... cleanly")
- `git diff --name-only origin/main -- . ':!.task-board' | wc -l` → `35`
- Sorted +/- line identity diff (rev2 vs worktree over origin/main) → empty output, exit 0
- `go build ./...` → exit 0 (not a gate; I did not re-run any test suites. The runner's hosted gate is the arbiter.)
- No other content change, no CHANGELOG/LOGBOOK edit.
