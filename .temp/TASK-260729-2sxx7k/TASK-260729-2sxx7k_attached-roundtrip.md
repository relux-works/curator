# TASK-260729-2sxx7k — native Linux validation evidence

**Observed:** 2026-07-29T01:04:42Z–01:09:58Z  
**Host:** `ssh lev`  
**Classification:** non-gating, non-release validation of the independently
accepted `TASK-260720-1ljev5` snapshot only. This is not evidence for the
concurrently changing `TASK-260720-1nlmvv` tree and is not a Linux release or
qualification claim. Replay every native gate on the final integrated
candidate after an operator supplies a manager-approved Go root and trusted
tree identity.

## Outcome

The deterministic accepted source snapshot was transferred into a private
remote temporary directory, verified byte-for-byte, extracted, inspected, and
fully removed. Native build, vet, test, race, coverage, and lint gates were
**not run** because preflight found no installed Go executable or approved Go
root. No exit code is reported for commands that were not invoked.

This is the precise preflight-failure branch permitted by the task acceptance
criteria. Ambient PATH, an automatically downloaded toolchain, or Ubuntu's
unaccepted Go 1.26 package would violate the accepted `go-v1` trust boundary.

## Accepted source identity

- Board source: `TASK-260720-1ljev5`, status `done`.
- Independent source verdict: review cycle 3 `ACCEPTED`.
- Preserved worktree:
  `.temp/TASK-260720-1ljev5/worktree`.
- Git base:
  `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`.
- Source manifest: 243 paths from
  `git ls-files --cached --others --exclude-standard`, sorted bytewise.
- Explicit exclusions required by the accepted task provenance:
  `.task-board/`, untracked `task-board.config.json`, and the absent
  `agents/skills/skill-go-testing-tools` gitlink. Product source, tests,
  module files, workflows, docs, and the two tracked skill symlinks remained.
- Deterministic archive controls: copied staging tree; normalized mtimes;
  numeric owner/group normalization; `gzip -n`; and
  `COPYFILE_DISABLE=1 tar --no-mac-metadata --no-xattrs --no-acls
  --no-fflags`.
- Two independently produced corrected archives compared byte-identically:
  `cmp -s ...clean-a.tar.gz ...clean-b.tar.gz` → exit `0`.
- Source-manifest/archive-list comparison:
  `diff -u TASK-260729-2sxx7k_source-files.txt
  TASK-260729-2sxx7k_archive-files.txt` → exit `0`.
- Corrected archive gzip integrity:
  `gzip -t ...accepted-snapshot-clean-a.tar.gz` → exit `0`.
- Canonical archive SHA-256:
  `f65b4d85f76ef06b09863206dc95ba48bceac92f0573a54aee8c2c25f3d4ee2a`.

## Platform and trusted-root preflight

The task consumed the accepted `TASK-260729-1vpytz` inventory before making a
fresh minimal read-only probe. That prerequisite recorded Ubuntu 26.04 LTS
x86_64, no Go on PATH, no conventional Go root, no installed `golang*`
package, and an unaccepted Ubuntu Go 1.26 package candidate.

Fresh standalone command:

```sh
ssh -o BatchMode=yes -o ConnectTimeout=10 lev '<read-only OS, filesystem, PATH, and conventional-root probe>'
```

Exit: `0`.

Observed:

```text
observed_utc=2026-07-29T01:04:42Z
user=lev uid=1000
Linux 7.0.0-15-generic x86_64 GNU/Linux
os_release=Ubuntu 26.04 (Resolute Raccoon)
home_fs=ext4 /dev/sda2 rw,relatime
tmp_fs=tmpfs tmpfs rw,nosuid,nodev,nr_inodes=1048576,inode64,usrquota
path_go=absent
candidate_root=/usr/local/go absent
candidate_root=/opt/curator/toolchains/go1.25.12 absent
candidate_root=/usr/lib/go absent
candidate_root=/snap/go/current absent
```

There is therefore no exact absolute launcher that can be admitted, no clean
`GOROOT` equality to probe, no Go 1.25 family/version to verify, and no
manager-approved `curator-go-toolchain-v1` tree digest. Preflight does not
permit any native Go gate.

## Native gate disposition

The following commands were intentionally **not invoked**, so they have no
exit codes:

```text
<approved-goroot>/bin/go build ./...
<approved-goroot>/bin/go vet ./...
<approved-goroot>/bin/go test -count=1 ./internal/buildcache ./internal/scopes ./cmd/curator
<approved-goroot>/bin/go test -count=1 ./...
<approved-goroot>/bin/go test -count=1 -race ./internal/buildcache ./internal/scopes ./cmd/curator
<approved-goroot>/bin/go test -count=1 -cover ./internal/buildcache ./internal/scopes ./cmd/curator
golangci-lint run ./internal/buildcache/... ./internal/scopes/... ./cmd/curator/...
```

If replayed, each Go invocation must use the already admitted absolute native
launcher under an empty manager-controlled environment with `GOENV=off`,
`GOTOOLCHAIN=local`, fixed locale, private HOME/TMPDIR, and no ambient
`GOROOT`. No test file or product file was changed in this validation task.

## Transfer and remote identity evidence

### Rejected first archive

The first locally reproducible archive had SHA-256
`a1ef43bea46adb4331ddde53f753c95f476999af3e8b36a36f785e7992f73d83`.
Transfer and SHA verification exited `0`, but GNU tar extraction exposed
macOS AppleDouble/xattr records as `._*` files. Extraction itself exited `0`
with warnings. This archive was rejected as a source transfer.

The exact private directory `/tmp/TASK-260729-2sxx7k.OROnhh` was removed and
`test ! -e /tmp/TASK-260729-2sxx7k.OROnhh` exited `0`.

### Accepted corrected archive

Private remote directory creation:

```text
remote_dir=/tmp/TASK-260729-2sxx7k.O7eoSd
owner=lev:lev mode=700 type=directory
```

Command exit: `0`.

Transfer:

```sh
scp -q ...accepted-snapshot-clean-a.tar.gz \
  lev:/tmp/TASK-260729-2sxx7k.O7eoSd/accepted-snapshot.tar.gz
```

Exit: `0`.

Remote archive identity:

```sh
ssh -o BatchMode=yes lev \
  'sha256sum /tmp/TASK-260729-2sxx7k.O7eoSd/accepted-snapshot.tar.gz'
```

Exit: `0`.

```text
f65b4d85f76ef06b09863206dc95ba48bceac92f0573a54aee8c2c25f3d4ee2a
```

Remote `gzip -t` exited `0`. Remote extraction exited `0` without output.
One standalone assertion then verified all of the following and exited `0`:

- `go.mod`, `cmd/curator/main.go`, `internal/buildcache/cache.go`, and
  `internal/scopes/gc.go` exist;
- `.task-board`, `task-board.config.json`, and every `._*` path are absent;
- extracted representative hashes are:

```text
go.mod
3b5e489ef304bbccd82706f7d19598e081667285c7b798903c0831ee99e157e2
internal/buildcache/cache.go
680c3afb6498285926d55d3613076253b0e2520e7930d57bd408a4bfafc5b32c
internal/scopes/gc.go
fab4eadda0e57db1eba353a028b9f2f3d1a762232c8a10499c4fc55acfca7e6b
```

The standalone local `shasum -a 256` over those same three source files exited
`0` and returned identical hashes.

## Cleanup and no-mutation proof

Cleanup command:

```sh
ssh -o BatchMode=yes lev \
  'rm -rf -- /tmp/TASK-260729-2sxx7k.O7eoSd &&
   test ! -e /tmp/TASK-260729-2sxx7k.O7eoSd'
```

Exit: `0`.

Independent postflight command exited `0` and reported:

```text
observed_utc=2026-07-29T01:09:58Z
path=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/games:/usr/local/games:/snap/bin
path_go=absent
task_temp_leftovers=none
```

It asserted both exact remote temporary directories absent, no
`/tmp/TASK-260729-2sxx7k.*` entry, and all probed Go roots still absent. The
only remote writes were inside the two mode-0700 task temporary directories;
both were removed. No package, PATH, profile, registry, system root, product,
release, stage, commit, publication, or pin mutation was attempted.

## Tool readiness and evidence caveats

Local readiness commands exited `0` for:

- `task-board version 0.23.0`;
- `/usr/bin/ssh`, OpenSSH 10.2p1;
- `/usr/bin/tar`, bsdtar 3.5.3;
- `/usr/bin/shasum`;
- `/usr/bin/rsync`, openrsync protocol 29;
- `/usr/bin/gzip`, Apple gzip 479.

`tar --help | rg 'xattr|mac-metadata|copyfile'` exited `1` because the compact
help omitted those options; `man tar | col -b | rg ...` then exited `0` and
confirmed `--no-mac-metadata`/`--no-xattrs`. This was option discovery, not a
product validation gate.

The result is deliberately limited: source transfer/identity and teardown are
green, while native Go qualification is unavailable due to the precise
trusted-toolchain preflight failure. It must not be promoted into a gating or
release claim.
