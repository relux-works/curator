# TASK-260907-187z6x revision 1 review verdict: changes_requested

Candidate: 77cdaf87a6b12a86eca101f421f92bd1c692cc46; base 09b25ef6629b41455d91dcb252ab4e4034e12750. No repository source changed. Reviewed modified files match candidate blobs, including untracked test blob f4abd7475351a098cb9f5d558775391c43da13d9. Run is not goal-bound (`task-board spawn goal` checked before verdict).

## R1 — medium: partial-activation reinstall reports installation; row 2 omits required operator-line assertion

At cmd/curator/profile_git_reinstall_test.go:134, row 2 discards stdout. Unlike the seven success rows, it never checks the operator line. On the exact candidate, this row prints `installed profile groot (lock sha256:...)`, not `updated profile groot (lock sha256:...)`. cmd/curator/profile.go:128-133 prints installed unconditionally on activation errors, ignoring the returned `updated` flag. The newly reachable Git activation failure therefore violates the pinned environments §9.1 reinstall reporting rule and the report's claim that every case is reported as an update. The normal eight-row suite stays green.

Reproduction: use a Go test overlay (no production mutation) on cmd/curator/profile_git_reinstall_test.go, replacing:

```go
code, _, stderr := runProfile(t, source, "profile", "install", operand, "--use")
```

with:

```go
code, stdout, stderr := runProfile(t, source, "profile", "install", operand, "--use")
t.Logf("stdout:\n%s\nstderr:\n%s", stdout, stderr)
updatedLine(t, stdout)
```

Run `go test -overlay .temp/TASK-260907-187z6x-review/row2.json ./cmd/curator -run 'TestProfileGitReinstallHonoursUseAndTakeover/not-current/use_without' -count=1 -v`. Exit 1: `operator line "installed profile groot (lock sha256:...)", want the reinstall reported as an update`. Overlay JSON maps the absolute original test path to row2.go, the modified copy. Full observed output attached in the review checks resource.

Required rework: assert the partial-failure operator output, make Git reinstall reporting honour the update result even when activation is partial, and preserve the requested path-root behaviour boundary. Keep exit 1 and both diagnostics. Add the missing no-backup assertions to the two neither-flag rows so the eight-row backup claims are actually checked. Re-run narrow sibling tests and return for review.

## R2 — evidence correction: “nothing written, no backup” is only demonstrated for the blocked Claude home

The producer results describe row 2 as no writes/no backup without qualification. Actual stdout reports codex_cli, opencode and pi switched. This is expected by pinned environments §9.2: attempt every entry, retain successful switches, leave recorded current unchanged if any entry failed. Do not implement global rollback to satisfy the incorrect prose. Correct the report/table to distinguish the blocked Claude surface (unchanged, no backup) from other adapters; assert their partial-switch outcomes or explicitly bound the claim. The new tests currently check only the fresh Claude directory for backup absence, not every home.

## Verified implementation and normative interpretation

Read curator-spec at CI SPEC_PIN 87a0d0060bad64ab883d007dcdf35df7485368bf, protocol/environments.md §§8.3, 9.1, 9.2, 9.5 and cli/curator.md install row. There is no reinstall carve-out from --use/--takeover. Parity is correct. Plain --use must fail on the unmanaged Claude surface; no valid contract promises the previous silent success. No repository caller dependency on that silence identified; external callers not audited. The stale first DoD wording is interpreted through the binding review note and spec: explicit takeover permits the retry to converge, not unconditionally fail.

The Git same-source branch runs updateLocked then the existing reinstallActivation/activateReinstall seams. Path implementation and activation helper bodies are byte-unchanged; only the helper comment changes. CHANGELOG is under Unreleased/Fixed. Test runProfile calls production run(), and the Git operand is https://example.com/groot with a local insteadOf rewrite. All 8/8 requested flag/currency rows execute. Successful operator-line checks cover 7/8 rows; failed row 2 is the gap above. Hash text uses prefix/suffix matching, not a full exact lock comparison.

Takeover safety: switch.go:503 opens backup before writes; :529 removes the Claude directory entry before WriteFile, so it does not open the old symlink target for writing. openBackup reads old bytes into generation 1. Row 4 asserts original backup bytes, materialized groot context and notice. Existing foreign-symlink test independently passes. No new platform-specific feature or skip shape requires a ledger row here; adjacent Git fixture cases are likewise not registered. No cross-platform execution performed by this reviewer.

## Independent verification (zsh, set -o pipefail)

- `go test ./internal/envprofile ./cmd/curator -run 'Reinstall' -count=1`: exit 0; envprofile 23.519s, CLI 67.680s. Includes path sibling and eight Git rows.
- `go test ./internal/envprofile -run 'TestUseWithoutTakeoverRefusesUnmanaged|TestUseTakeoverBacksUpAndNotifies|TestForeignSymlinkStopsSwitch' -count=1`: exit 0, 10.993s.
- `go vet ./internal/envprofile ./cmd/curator`: exit 0. `gofmt -l` changed Go files and `git diff --check`: clean.
- Producer mutant independently rerun via Go overlay: delete only Git branch activation (path branch remains). `go test -overlay .../drop.json ./cmd/curator -run 'TestProfileGitReinstallHonoursUseAndTakeover/not-current/(use|takeover)' -count=1 -v`: exit 1. Both named use rows fail (exit 0 instead of 1; missing switch), takeover-without-use control passes. Thus 2/2 affected rows kill the Git-only narrowing, 1/1 control survives.
- Reviewer mutant: replace options.Use with true only at Git reinstallActivation call. `go test -overlay .../force.json ./cmd/curator -run 'TestProfileGitReinstallHonoursUseAndTakeover/not-current' -count=1 -v`: exit 1. Neither-flag and takeover-without-use fail; both use rows pass. Tests detect unauthorized activation, 2/2 targeted rows.
- Row-2 output assertion overlay: exit 1 as above, with production untouched.

Reused, did not rerun: TASK-260907-187z6x_change-request_rev1-validation.log, remote run https://github.com/relux-works/curator/actions/runs/35722834632, exit 0. Verified local Git object d022e96515c82ab3542ae4cf4ea1d664ca28a25d has tree 77cdaf87a6b12a86eca101f421f92bd1c692cc46. Attached evidence reports hosted macOS/Windows/Ubuntu tests, lint, naming, interop, gate self-tests and macOS/Ubuntu race green. Candidate suite and rose-air skipped: unverified, not passing. Accepted this runtime-attached gate evidence, not an independent remote rerun. Producer build/lint/package subset claims were read but not relabelled as reviewer reruns.

## Logbook entry

Review discovered a reachable operator-reporting mismatch hidden by row 2 discarding stdout, plus overbroad no-write evidence. Core activation and takeover semantics are supported by the spec. Ordinary implementation/test/evidence rework; route to to-dev, not blocked. Issue #73 remains open; no control-root LOGBOOK modified. No acceptance or integration performed.
