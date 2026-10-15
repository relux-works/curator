# Brief UUT: buildrepo tests assume /usr/bin/false is a regular file (TASK-261015-1yj9x7)

Repository: curator, tests only. On Linux distributions that ship uutils coreutils (for example Ubuntu 26.04), `/usr/bin/false` and `/usr/bin/true` are symlinks into a multicall binary. `ValidateGitTool` (internal/buildrepo/admission.go) correctly refuses a credential broker or SSH wrapper that is not a regular file (Lstat, no symlinks). Three tests hard-code `/usr/bin/false` as such a file, so on those hosts 28 buildrepo tests fail with `build_repository_identity_invalid: credential broker is not an admitted absolute regular file` (and follow-on `argv.log: no such file`):
- internal/buildrepo/admission_test.go:773 `realTool.AskPass = "/usr/bin/false"`
- internal/buildrepo/transport_test.go:683 `realTool.AskPass = "/usr/bin/false"`
- internal/buildrepo/transport_test.go:684 `realTool.SSHWrapper = "/usr/bin/false"`

Fix: a test helper that writes a regular executable stub (`#!/bin/sh` then `exit 1`, mode 0755) into `t.TempDir()` and returns its absolute path; use it in all three places. Search the whole repository's tests for other hard-coded `/usr/bin/true`, `/usr/bin/false`, `/bin/true`, `/bin/false` used where a regular file is required, and fix those too. Do not change production code or weaken the symlink refusal.

Prove it: add a test that points the broker at a symlink to the stub and expects the refusal (it must stay refused). Run `go test ./internal/buildrepo/...` and the repository lint. Leave the change uncommitted for the Change Request flow. No host names in code or text.
