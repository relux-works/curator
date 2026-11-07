# Linux process evidence during build-cache collection

A non-root sweep reads the four `Uid` fields of each visible process's
`/proc/<pid>/status`: real, effective, saved-set and filesystem UID. If any
matches the manager's effective UID, that process stays in the evidence set.
Unreadable or malformed credentials are uncertainty, not a foreign-user match.
An unreadable live same-user executable retains every build and reports why.
Exited processes and proven no-image processes retain the existing treatment.

Other-UID processes do not make an owner-only publication uncollectable just
because Linux denies access to their executable links. When any foreign
process was excluded, a candidate must additionally have no group/other
permission bits and no setuid/setgid bits. The existing descriptor-based
inspection still proves the complete cache boundary, artifact owner, regular
file type, single hard link, receipts, contents and publication age. Shared
or special-mode candidates are retained. The same proof is required before
finishing removal of a retired entry; unprovable wreckage is retained.

This is discretionary protection for an unprivileged user's store, not a
claim against root or another privileged actor bypassing file permissions.
An ordinary foreign UID cannot execute the owner-only inode. A setuid
process whose real, saved or filesystem UID is still this user remains in
scope even if its effective UID or its proc-directory owner is root. Special
artifact modes never earn the foreign-process exception. An actor capable of
arbitrary credential replacement or privileged filesystem mutation is outside
this unprivileged-store argument. A root manager keeps the previous strict
all-process enumeration instead of applying the exception.

Linux snapshots also retain executable device/inode identities obtained
through `/proc/<pid>/exe`. Path spelling alone cannot prove that a build is
unused: a process may execute an external hard-link alias. A still-linked
alias is already refused by the single-link check; an alias unlinked after
launch can leave one link in the cache, so the live inode match still retains
that build. No extra process is executed by the collector.

The collector still uses a process snapshot under the manager-home mutation
lock. This change does not create a global barrier against arbitrary new
executions after that snapshot, change process visibility outside the current
procfs view, or weaken entry validation. macOS and Windows keep their existing
native enumeration and path matching.

The kernel's [process status documentation](https://man7.org/linux/man-pages/man5/proc_pid_status.5.html)
defines the UID tuple; [executable-link documentation](https://man7.org/linux/man-pages/man5/proc_pid_exe.5.html)
defines ptrace-controlled access and deleted-link spelling;
[execve](https://man7.org/linux/man-pages/man2/execve.2.html) describes credential
transitions at execution.

## Qualification

`Linux unprivileged process sweep` runs on hosted Ubuntu and fails immediately
if the qualification account is root. The job creates a dedicated unprivileged
UID, so non-dumpable processes belonging to the runner account cannot prevent
the collection controls. It requires a complete baseline. Ordinary suite runs
still exercise these rows: an existing unreadable same-user process must retain
all entries, including the unused control, rather than make a false collection
claim. No skip substitutes for that assertion.

The isolated job runs real cached copies of the test binary,
checks a real other-UID PID, and makes its own child non-dumpable to obtain a
real same-user permission refusal. Each retained publication has a nearby
positive control: after stopping the child, the same entry is collected.
UID parser and permission-boundary controls cover mixed setuid credentials,
malformed metadata and shared/special artifact modes. No root-only fixture or
new skip class substitutes for these controls. The same hosted step runs
`python3 .github/ci/linux-sweep-control.py`: it temporarily omits the owner
UID and requires the real running-build assertion to fail by name, then
restores the exact source bytes. A compilation error or unrelated timeout is
not a detected plant.

Run the focused Linux controls with a private HOME and cleared Git, SSH and
agent selectors:

```sh
go test -count=1 -timeout=600s ./internal/buildcache \
  -run '^TestLinux(UnprivilegedSweep|ProcessUIDControls|ForeignImagePermissions|ExecutablePaths)$'
```
