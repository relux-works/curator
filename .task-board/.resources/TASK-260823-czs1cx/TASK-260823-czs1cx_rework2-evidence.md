# TASK-260823-czs1cx — rework cycle 2 evidence (RUN-260823-bd9292)

Addresses review verdict RUN-260823-3e7189 (CHANGES REQUESTED). The
`fixed_environment` half was accepted in cycle 1 and is untouched here; this
cycle fixes only the still-red toolchain link case.

## Diagnosis, independently confirmed

The reviewer's diagnosis holds on both points, verified from primary sources
rather than taken on trust:

1. **PR #29 was a no-op on Windows.** `$GOROOT/src/os/file_windows.go`,
   `func Symlink`, first statement: `oldname = filepathlite.FromSlash(oldname)`.
   Removing the explicit `filepath.FromSlash` from the materializer changes
   nothing, so the byte-exact round-trip assertion it added was guaranteed to
   fire.

2. **The vector is correct; the implementation was guilty.** Decoding
   `preimage_base64` for `unsorted-directories-files-and-internal-link` from
   candidate `edd0721`:

   ```
   ...L\x00..\x0dpkg/tool-link\x00\x00\x00\x00\x00\x00\x00\x09../bin/go...
   ```

   Nine bytes, `../bin/go`, no backslash anywhere in the preimage. The published
   `content_sha256` is
   `sha256:baf7c5f3b9c3f1fae3da4c356381bf74442aa7f8f0b6fb2304c9c10833d6032e`.

   `internal/godriver/fingerprint.go` stored the raw `os.Root.Readlink` output in
   `record.link` and hashed it, so on Windows the `curator-go-toolchain-v1`
   preimage carried `..\bin\go` and the toolchain content digest — which is
   supposed to name a tree — became a property of the host that walked it.

## Change

| File | Change |
| --- | --- |
| `platform_windows.go` | `protocolLinkTarget` = `filepath.ToSlash` |
| `platform_unix.go` | `protocolLinkTarget` = verbatim |
| `fingerprint.go` | normalize after `Readlink`, **before** validation and hashing |
| `fingerprint_equivalence_test.go` | same normalization in the reference traversal |
| `builddriver_positive_conformance_test.go` | round-trip assertion compares through the normalizer; `fixedEnvironmentForHost` skips instead of failing |
| `fingerprint_test.go` / `fingerprint_unix_test.go` | pinned-digest test moved out from under the unix build tag |

Build-tagged rather than a blanket `ToSlash`: on unix a backslash is an ordinary
filename character, and rewriting it would change a legitimate unix digest.

The equivalence gate's reference traversal had to move too — it encodes the same
protocol, and normalizing only the real walk would make the two diverge on
Windows and turn that gate red.

## The fix is load-bearing

Proved on the native Windows host by neutralizing **only** the normalizer
(`protocolLinkTarget` → verbatim on Windows) and changing nothing else:

| Windows run (DESKTOP-3PBO632, go1.25.5 windows/amd64) | `unsorted-…-internal-link` | pinned-digest test |
| --- | --- | --- |
| normalizer neutralized | FAIL — `..\bin\go` vs `../bin/go` | FAIL — `sha256:e9b9a60b…` |
| with the fix | PASS | PASS |

That `e9b9a60b…` is the host-forked digest the bug produced, against the
published `baf7c5f3…`.

## Validation

| Gate | Host | Real exit |
| --- | --- | ---: |
| focused godriver tests vs candidate `edd0721` | macOS | 0 |
| focused godriver tests vs candidate `edd0721` | native Windows | 0 |
| **full `internal/godriver`** vs candidate `edd0721` | native Windows | 0 |
| `make candidate-test` (`CI_REQUIRE_FULL_ROOT=1`) | macOS | 0 |
| `make check-ci` (committed `SPEC_PIN` 00b1688) | macOS | 0 |
| `go vet ./...` | darwin / linux / windows | 0 |
| `golangci-lint run` | darwin / linux | 0 |
| `GOOS=windows golangci-lint run` | darwin | 1 — see below |

`GOOS=windows golangci-lint run` reports 10 findings (5 errcheck, 4 gosec,
1 revive), **all pre-existing** in `*_windows.go` files this change does not
touch: `internal/buildcache/protection_windows.go`,
`internal/buildrepo/protection_windows.go`,
`internal/managerlock/identity_windows.go`(+test),
`internal/transaction/durability_windows.go`,
`internal/godriver/controls_windows.go`. Reported red honestly rather than
presented as passing; not introduced by this task and not in its scope.

Both owned cases against candidate `edd0721` on native Windows:

```
--- PASS: TestFixedEnvironmentAndFiveDirectArgvFormsVector/fixed_environment
--- PASS: TestToolchainIdentityVectors/unsorted-directories-files-and-internal-link
ok  github.com/relux-works/curator/internal/godriver  (full package, exit 0)
```

## Candidate identity

No regeneration for this half — the toolchain preimage was already correct and
platform-neutral. Candidate `edd07210d4f3db34fd60238cb14b90f837de03cb` on
`candidate/schema-8-rc.9` and manifest
`803918bf8672f76cf990985e51db213b826674cd5bb54fbf47731b8404b44403` stay valid,
and the identity recorded on TASK-260822-c0rxj7 in cycle 1 is unchanged.

## Delivery

Signed commit `695c041` on `fix/TASK-260823-czs1cx-toolchain-link-normalize`,
PR #31.

## Note for the rerun task (TASK-260823-1l1p8q)

Dispatch the candidate rerun from `main` **after** PR #31 merges. The ordinary
`Test (windows-latest)` lane cannot catch this class of bug: the pinned default
root publishes no `vectors/build-drivers.json`, so both owned conformance tests
skip there. The newly cross-platform `TestFingerprintImplementationMatchesRC4ToolchainVector`
does now run on that lane and pins the same digest, which closes that gap.
