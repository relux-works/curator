# Native-home test boundary

Only test code imports this package. Existing package `TestMain` functions call
`Start` before any test runs, then use `Finish` for every normal or explicit
helper exit. `Start` fails closed if its protected preimage or isolation cannot
be established. It pins HOME, USERPROFILE and all XDG roots to fresh temporary
directories and clears inherited Claude, Codex and Pi home overrides.

The protected snapshot includes the OS-account home and inherited native-home
overrides. It records lstat identity, mode, size, modification time, instruction
symlink targets, hashes of regular CLAUDE.md, AGENTS.md and environment markers,
and skill-entry identities. It reads no credential files and does not hash
instruction symlink destinations. A changed or unreadable postimage fails the
package. Only the newly allocated isolation directory is cleaned up.

`Control` and `Child` test the real package exit boundary using synthetic
original HOME directories. The negative child restores only that synthetic
HOME, invokes the production native-home resolver and writes one context
witness. Acceptance requires exit 1 and the fingerprint of that exact synthetic
root; a nearby write through the isolated HOME must exit 0. The plant does not
select the OS-account home.

Run these package controls with a second temporary HOME:

```sh
runner_home=$(mktemp -d)
env -u GIT_DIR -u GIT_WORK_TREE -u SSH_AUTH_SOCK -u GOFLAGS HOME="$runner_home" \
  go test -count=1 -v ./internal/envprofile -run '^TestNativeHomeGuard'
env -u GIT_DIR -u GIT_WORK_TREE -u SSH_AUTH_SOCK -u GOFLAGS HOME="$runner_home" \
  CURATOR_TEST_CLI_HELPER=1 go test -count=1 -v ./cmd/curator -run '^TestNativeHomeGuard'
```

Fixture-specific overrides remain useful for independence between tests. They
must always select synthetic paths. The postimage check detects changes within
the observed process window; it does not attribute unrelated concurrent changes
to a particular process or replace filesystem access controls.
