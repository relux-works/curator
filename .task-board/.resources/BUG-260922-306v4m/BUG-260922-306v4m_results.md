# BUG-260922-306v4m results — rose-air rustup probe reports no evidence

## Change

`.github/ci/install-rust-toolchain.sh`: the failure branch now calls
`print_rustup_diagnostics` (stderr) before the unchanged remedy sentence.
Resolution loop, success path, exit codes, probe set: untouched.
Also: `gate-selftest.sh` rows, one docs paragraph in
`docs/self-hosted-runner-setup.md`, one CHANGELOG entry under
Unreleased/Fixed.

No probe path was added: the evidence in hand proves no missing one, and the
brief boundary for this revision is evidence-first (AC 4 fires on the next
red main push, which has not happened yet).

## Failure block exact shape (stderr, exit 1, remedy last)

Observed with `RUNNER_NAME=rose-air RUNNER_OS=macOS` and empty fixtures:

```text
rust-pin: rustup not found; runner diagnostics:
rust-pin:   RUNNER_NAME=rose-air
rust-pin:   RUNNER_OS=macOS
rust-pin:   hostname=e11-1.macminivault.com
rust-pin:   whoami=administrator
rust-pin:   HOME=/Users/administrator
rust-pin:   CARGO_HOME=/tmp/.../empty-cargo (set)
rust-pin:   HOMEBREW_PREFIX=/tmp/.../empty-brew
rust-pin:   PATH=<searched PATH, $CARGO_HOME/bin first>
rust-pin:   candidate <cargo>/bin/rustup: absent
rust-pin:   candidate <brew-prefix>/bin/rustup: absent
rust-pin:   candidate /opt/homebrew/bin/rustup: absent
rust-pin:   candidate /usr/local/bin/rustup: absent
rust-pin:   listing <cargo>/bin: no names containing rust or cargo
rust-pin:   listing /opt/homebrew/bin: absent
rust-pin:   listing /usr/local/bin:
rust-pin:     <names matching rust|cargo, at most 20, if any>
rust-pin:   command -v rustup: (no rustup on PATH)
rust-pin:   type -a rustup:
rust-pin:     <type -a output, one line each>
rust-pin: rustup is not installed on this runner; install it once per docs/self-hosted-runner-setup.md, then re-run this lane
```

States per candidate: `executable` / `exists-not-executable` / `absent`.
`CARGO_HOME` carries `(set)` or `(defaulted)`. Unset `RUNNER_NAME` /
`RUNNER_OS` / `HOMEBREW_PREFIX` print as `unset`. Missing `hostname` /
`whoami` print as `unknown`. Only named variables are printed.

## The three gate-selftest rows (all run on this host, exit codes real)

`bash .github/ci/gate-selftest.sh` → exit 0, 224 passed, 0 failed, 0 skipped.

1. rustup absent everywhere → block shown, remedy last, exit 1:
   - `a runner without rustup fails` (exit 1) plus 18 content rows:
     block marker, RUNNER_NAME/RUNNER_OS/hostname/whoami/HOME/CARGO_HOME
     (set)/HOMEBREW_PREFIX/PATH, 4 candidate lines (incl. the
     `exists-not-executable` middle state via a non-exec stub under the
     fake prefix), 3 listing lines, `command -v`, `type -a`, remedy-last.
   - `a runner without rustup and without CARGO_HOME fails` (exit 1) plus
     5 rows: `(defaulted)` spelling, unset spellings, defaulted candidate
     absent, exactly 3 candidate lines when `HOMEBREW_PREFIX` is unset.
   - `a runner with many rust-named files still fails` (exit 1): 25 seeded
     names → exactly 20 entry lines (bound proven), remedy still last.
2. rustup only under a probed candidate → success, no block, exit 0:
   - `the installer finds rustup under CARGO_HOME/bin without PATH help`
     (exit 0) + `... success prints no diagnostics block` + exact
     GITHUB_PATH bytes (one line) + `using rustup at` line still present.
   - Same no-block pin on the Homebrew-prefix fixture (exit 0) + exact
     two-line GITHUB_PATH order (cargo bin, then brew bin).
3. Narrowing mutant killed (production call site
   `print_rustup_diagnostics` in `.github/ci/install-rust-toolchain.sh`,
   reached from the `Install pinned Rust toolchain via rustup` lane step
   in `.github/workflows/ci.yml`):
   - mutant drops the `/opt/homebrew/bin/rustup` line from the diagnostic
     enumeration only (middle loop line, so `; do` survives);
   - `the narrowing mutant drops exactly the diagnostic Apple-silicon
     candidate` (production 2 occurrences, mutant 1);
   - `the narrowing mutant is syntactically valid` (`bash -n`);
   - `the narrowed mutant still exits 1` (resolution untouched);
   - `the named Apple-silicon-candidate row kills the narrowing mutant`
     (dropped line absent from mutant output);
   - `the narrowed mutant keeps its other candidate lines`.

## Success path byte-identical (independent proof, not the selftest)

Same shared fixtures against `HEAD` script vs new script (`cmp`):

- cargo-only stdout+stderr, GITHUB_PATH, rustup invocation log: IDENTICAL
- brew-only stdout+stderr, GITHUB_PATH, rustup invocation log: IDENTICAL
- absent-fixture GITHUB_PATH write: IDENTICAL; exit codes 0/0/1 both sides;
  remedy sentence byte-equal old vs new.

## Anomaly (pre-existing, out of scope, not changed)

While probing I found bash `command -v` reports a PATH entry even when the
file is not executable: a non-exec `rustup` stub under `$CARGO_HOME/bin`
is therefore *found* by resolution (PATH is prepended before probing) and
fails later at exec with `Permission denied`, never reaching the new
diagnostics. The `exists-not-executable` diagnostic state is reachable via
the prefix/fixed candidates, which is what the selftest fixtures. No
behavior change made; resolution semantics are exactly as before.

## Notes for review

- `references/negative-evidence.md` (named in the assignment) exists
  neither in the worktree nor the control root, so no shape from it could
  be followed; negative coverage is the exit-1 asserts plus the committed
  narrowing mutant above, with the production call site named.
- Lint: repo configures no shell linter (Makefile `lint` is
  golangci-lint; no Go files touched; shellcheck/shfmt not installed).
  `bash -n` passes on both edited scripts; the code is bash-3.2 and
  Git-Bash compatible (no assoc arrays, no mapfile; guarded pipelines).
- Full landing suite not run manually per campaign rules (runtime runs it
  once at handoff); narrow gate `gate-selftest.sh` run twice with real
  exit codes (first run 215/2-fail exposed two test bugs, fixed; second
  run 224/0 green).
- Work left uncommitted in the story worktree as required.

revision 2 = revision 1 unchanged; republished under the new board binary so the validation evidence is tree-bound (validation_not_bound_to_tree).

## Revision 4 (final-leaf republish)

Revision 3 was ACCEPTED (review-verdict-rev3: identity review) but remained the
Story's only open leaf, which task-board can land neither by checkpoint nor by
integrate. The orchestrator ran `worktree converge` onto trunk `fad88136`; the
accepted delta was carried uncommitted into this worktree. This revision
republishes that delta unchanged as the Story's final revision.

- Identity: `git status --short` shows exactly the 4 rev3 paths
  (`.github/ci/gate-selftest.sh`, `.github/ci/install-rust-toolchain.sh`,
  `CHANGELOG.md`, `docs/self-hosted-runner-setup.md`). Per-file content-line
  comparison against `BUG-260922-306v4m_change-request_rev3.patch` (ignoring
  index hashes and hunk headers): all 4 files IDENTICAL, including CHANGELOG
  (no combination drift). No conflict markers found. No file changed in this
  revision.
- Verification (this run, real exit codes): `bash .github/ci/gate-selftest.sh`
  run directly as a standalone process, exit 0, 224 passed, 0 failed — the
  rustup-absent diagnostics rows, the under-candidate success rows (no block,
  exact GITHUB_PATH bytes), and the narrowing-mutant kill all green.
  `bash -n` passes on both edited scripts.

## Revision 5 (refresh republish)

Revision 4 was ACCEPTED (review-verdict-rev4: base-refresh review, content
unchanged from rev3) but trunk moved again (2kqa77 landed, CHANGELOG only),
so the orchestrator ran `worktree converge` onto trunk `1511b345`; the
accepted delta is carried uncommitted in this worktree. This revision
republishes that delta unchanged as the Story's final revision. No file was
changed in this revision.

- Identity: `git status --short` shows exactly the 4 rev4 paths
  (`.github/ci/gate-selftest.sh`, `.github/ci/install-rust-toolchain.sh`,
  `CHANGELOG.md`, `docs/self-hosted-runner-setup.md`). Per-file content-line
  comparison of the live `git diff` against
  `BUG-260922-306v4m_change-request_rev4.patch` (ignoring index hashes and
  hunk headers): all 4 files IDENTICAL, including CHANGELOG
  (install-rust-toolchain.sh 72 content lines, gate-selftest.sh 144, docs
  11, CHANGELOG 11). The CHANGELOG hunk is purely additive (11 insertions,
  0 deletions): trunk's 2kqa77 entry is retained, this task's entry appears
  exactly once (grep count 1). No conflict markers found (grep count 0).
- Verification (this run, real exit codes): `bash -n` passes on both edited
  scripts (exit 0 each). `bash .github/ci/gate-selftest.sh` run directly as
  a standalone process, exit 0, 237 passed, 0 failed — the rustup-absent
  diagnostics rows, the under-candidate success rows (no block, exact
  GITHUB_PATH bytes), and the narrowing-mutant kill all green. (Count grew
  from rev4's 224 because trunk added gate rows since; the touched
  `.github/ci` scope is fully green.)
- DoD: all checklist items were already checked from the accepted revisions
  and remain satisfied; no code changed, so no new tests were required
  beyond the carried gate rows. Work left uncommitted as required.

Revision 5 (refresh republish): carried delta equals revision 4 except the
CHANGELOG combination, which is purely additive with both sides kept.

## Revision 6 (carry-forward republish)

Revision 5 was ACCEPTED on content. Trunk moved to `a48f584c`, so the
orchestrator ran `worktree converge STORY-260915-3w11un`; the accepted delta
is carried uncommitted in this worktree. Per-path verification against
`BUG-260922-306v4m_change-request_rev5.patch`, re-measured this run
(`git rev-parse HEAD:<path>` vs the patch `index` base hashes;
`git hash-object <worktree file>` vs the patch after hashes):

- `.github/ci/gate-selftest.sh`: trunk HEAD blob `3def74a6` equals the rev5
  base, so trunk did NOT touch it; worktree blob `0ea1a3e9` equals the rev5
  after-blob, so byte-identical to revision 5.
- `.github/ci/install-rust-toolchain.sh`: trunk HEAD blob `f45e0123` equals
  the rev5 base, so trunk did NOT touch it; worktree blob `be073ac3` equals
  the rev5 after-blob, so byte-identical to revision 5.
- `docs/self-hosted-runner-setup.md`: trunk HEAD blob `5b97209c` equals the
  rev5 base, so trunk did NOT touch it; worktree blob `38cafdf0` equals the
  rev5 after-blob, so byte-identical to revision 5.
- `CHANGELOG.md`: intersecting path (trunk HEAD blob `171ec1f8` differs from
  rev5 base `17773d62`). No conflict markers in any of the 4 files
  (`grep -rn '<<<<<<<'` over all four: no matches, exit 1). Nothing else
  changed.

CHANGELOG policy (orchestrator 2026-09-24): this task's CHANGELOG hunk
reverted entirely — `bash -c 'cmp CHANGELOG.md <(git show
HEAD:CHANGELOG.md)'` exit 0 (file equals trunk), and the entry sentence
occurs 0 times in `CHANGELOG.md` (`grep -c`, exit 1: absent, as required).
Entry text copied verbatim into "## CHANGELOG entry (for release prep)"
below for the release-prep leaf (verified character-identical to the rev5
patch hunk). No stray root TASK-*/BUG-*.md, test/ or ledger/ paths:
`git status --short --untracked-files=all` shows only the 3 remaining
modified files, no untracked entries.

Focused bounded run (this run, real exit codes, standalone processes):
`bash -n .github/ci/install-rust-toolchain.sh` exit 0;
`bash -n .github/ci/gate-selftest.sh` exit 0. Full `gate-selftest.sh` NOT
rerun in this revision: no code changed (carry-forward republish only,
scripts hash-identical to rev5), so rev5's tree-bound green run (237 passed,
0 failed, recorded in the Revision 5 section of this same results file)
stands for the identical script content. (Correction: that count lives in
the rev5 section here, not in
`BUG-260922-306v4m_change-request_rev5-validation.log`, which is the
remote-gate log and carries no local pass counts.)

DoD: all 13 checklist items were already checked from the accepted revisions
and remain satisfied (re-read this run: all `done:true`); the CHANGELOG
revert is policy-directed, content otherwise unchanged. Work left
uncommitted as required.

Retry note: run RUN-260924-65f74f performed this same Revision 6 publish,
then failed on runner heartbeat expiry; run RUN-260924-d6c00d
independently re-measured every hash, count, and exit code above and
corrected two citations (the unobserved intermediate CHANGELOG grep count
and the validation-log attribution). No file changed in that run.

Retry note (RUN-260924-826b7c, autonomous recovery after rev6 remote-gate
validation failed): the CR-6 validation log
(`BUG-260922-306v4m_change-request_rev6-validation.log`) shows the remote
gate failing only on `Race (ubuntu-latest)`: `go test -race` served stage
exit 1 plus `FAIL required case never ran on linux: internal/install ::
TestDryRunTouchesNothing` — a trunk Go-test/platform-case matter untouched
by this leaf (3 files: two bash scripts + one doc; no Go code). All other
remote lanes passed, including Gate self-test on all three OSes. This run
re-verified the carry-forward from scratch: trunk HEAD blobs
`3def74a6`/`f45e0123`/`5b97209c` equal the rev5 bases (trunk untouched),
worktree blobs `0ea1a3e9`/`be073ac3`/`38cafdf0` equal the rev5 after-images,
live `git diff` of the 3 files is byte-equal to the rev6 patch (`cmp` exit
0), no conflict markers, `CHANGELOG.md` equals trunk (`cmp` exit 0, entry
grep count 0), `git status --untracked-files=all` shows only the 3 modified
files. Focused bounded run with real exit codes, all standalone: `bash -n`
exit 0 on both scripts; full `bash .github/ci/gate-selftest.sh` exit 0 with
237 passed, 0 failed (rerun in THIS run, not inherited). DoD: all 13 items
remain checked. No file changed in this run; work left uncommitted.

## CHANGELOG entry (for release prep)

Verbatim entry text reverted from `CHANGELOG.md` (Unreleased / Fixed) for
the release-prep leaf to write:

```text
- The `Install pinned Rust toolchain via rustup` lane step now prints a
  runner diagnostics block before its remedy note when no probed location
  holds `rustup`: `RUNNER_NAME`, `RUNNER_OS`, `hostname`, `whoami`, `HOME`,
  `CARGO_HOME` (set or defaulted), `HOMEBREW_PREFIX`, the searched `PATH`,
  one executable/exists-not-executable/absent line per probed candidate, a
  bounded rust/cargo listing of the searched bin directories, and
  `command -v` / `type -a` for `rustup`. The bare remedy sentence could not
  distinguish a different runner, a different service user, or a `rustup`
  outside the probed paths (rose-air runs 35663049586, 35725359745); the
  next red run names its evidence. Exit code, remedy sentence, and the
  success path are unchanged.
```


## Revision 8 (carry-forward republish)

Revision 7 was ACCEPTED on content (review-verdict-rev7: carry-forward
without CHANGELOG; whole-patch patch-id `0ec91428`, per-file patch-ids
gate-selftest `2fbf268f`, install-rust-toolchain `b5651d36`, docs
`2aa4d062`). Trunk moved to `0a628621`, so the orchestrator ran `worktree
converge STORY-260915-3w11un`; the accepted delta is carried uncommitted in
this worktree. Per-path verification against the rev7 base (`a48f584c`) and
candidate tree (`47474653`), re-measured this run (`git rev-parse
<commit>:<path>` vs `git hash-object <worktree file>`):

- `.github/ci/install-rust-toolchain.sh`: trunk HEAD blob `f45e0123` equals
  the rev7 base, so trunk did NOT touch it; worktree blob `be073ac3` equals
  the rev7 after-blob, so byte-identical to revision 7.
- `docs/self-hosted-runner-setup.md`: trunk HEAD blob `5b97209c` equals the
  rev7 base, so trunk did NOT touch it; worktree blob `38cafdf0` equals the
  rev7 after-blob, so byte-identical to revision 7.
- `.github/ci/gate-selftest.sh`: INTERSECTING path (trunk HEAD blob
  `a8c67e96` differs from rev7 base `3def74a6`; trunk rewrote the ci.yml
  timeout-pin rows, renamed the pin to `committed_conformance_pin`, and added
  a Go timeout-regression row). Converge kept both sides: the worktree-minus-
  rev7 diff is body-identical to the trunk-minus-base diff (`cmp` exit 0
  after dropping the `index` header line; 216 insertions, 52 deletions both
  ways), no conflict markers in any of the 3 files (`grep` exit 1), rev7's
  narrowing-mutant rows (7 `narrowing mutant` lines) and trunk's
  `committed_conformance_pin` / `candidate-conformance` rows all present.
  Worktree blob `dc9e6148`. Nothing dropped or duplicated.
- `CHANGELOG.md`: base, rev7-after, trunk HEAD, and worktree blobs are all
  `171ec1f8` — the file equals trunk (`cmp` exit 0), no hunk to revert, and
  the verbatim 11-line entry remains preserved under "## CHANGELOG entry
  (for release prep)" below for the release-prep leaf. No stray root
  TASK-*/BUG-*.md, test/ or ledger/ paths.

Worktree staleness check: `git diff --name-only HEAD -- . ':!.task-board'`
lists ONLY the 3 rev7 paths above (CHANGELOG.md clean); `git status --short
--untracked-files=all` shows exactly those 3 modified files, no untracked
entries. No file changed in this revision.

Focused bounded run (this run, real exit codes, standalone processes):
`bash -n` exit 0 on both scripts; full `bash
.github/ci/gate-selftest.sh` exit 0 with 251 passed, 0 failed (rerun in THIS
run, not inherited — the count grew from rev7's 237 because trunk added gate
rows since; the rev7 rustup-absent, under-candidate success, and
narrowing-mutant-kill rows are all green in the log). DoD: all 13 checklist
items remain checked. Work left uncommitted as required.
