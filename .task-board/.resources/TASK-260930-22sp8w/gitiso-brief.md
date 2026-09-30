# TASK-260930-22sp8w — tests must not inherit ambient git config (THE ONLY CURRENT INSTRUCTION)

## Problem

Main bdb77413 is red on Test (rose-air), runner "macbook-iv". That runner copies the host user's ~/.gitconfig into HOME, and the host
config enables commit and tag signing. Fixture commits that should be unsigned come out signed, so these fail:
- cmd/curator TestSourceSignerVectorsAtProfileInstall/unsigned-refused, which expects context_source_unsigned but gets
  context_source_signer_rejected;
- TestRevisionDoesNotBorrowTagSignature.

Isolation exists today only in scattered places (draft_transport_test.go:121, envconfig_test.go, global_lock_publication_test.go,
main_test.go:1110).

## Required change

1. `.github/ci/test-gate.sh` exports, for every go test stage:
   - `GIT_CONFIG_NOSYSTEM=1`;
   - `GIT_CONFIG_GLOBAL` pointing to an empty file it creates under the evidence or temp dir.

   Log both in the gate output.
2. Defence in depth, for local runs: add a shared helper, e.g. in internal/testcli or a small internal test-support package, that
   TestMain in each package that runs git calls (at least cmd/curator, internal/rustsource, internal/crossconformance and any other
   package whose tests shell out to git). It sets the same two variables unless a test sets its own. Tests that already set
   GIT_CONFIG_GLOBAL on purpose, such as insteadOf rewrites, must keep working. Their t.Setenv overrides the TestMain value.
3. A regression row that runs the two failing tests (or an equivalent fixture path) under a hostile global config and shows they pass:
   commit.gpgsign=true, tag.gpgsign=true, gpg.format=ssh, and user.signingkey pointing to a generated throwaway key. The row sets the
   hostile config and then relies on the helper to neutralise it.
4. Add a gate-selftest.sh row asserting that test-gate.sh exports both variables.
5. Mutant: remove the helper call from cmd/curator's TestMain → the regression row fails. Give real exit codes.

Out of scope: the Rust failures on macbook-iv ("closed native Apple developer registry is unavailable", i.e. no
/var/db/xcode_select_link). That is host setup and goes to the operator.

No CHANGELOG/LOGBOOK edit. Never spell any employer name. Update the results, then run `task-board handoff TASK-260930-22sp8w --role developer`, then
END YOUR TURN.
