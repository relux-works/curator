# Linux process evidence during build-cache collection

The guarantee is per kernel: retain builds that any process on this kernel
is executing, with two bounded exceptions. Linux exempts executables opened
under a fanotify pre-content (HSM) watch from write denial. Also, a new exec
can begin after the final immediate-close probe and before removal. Either
exception can remove the path without stopping the already-running process;
a later exec of that path fails until the build is rebuilt. The HSM declaration
row records that limit; it does not test or qualify HSM execution.

A home shared over NFS and executed from another host is outside this
guarantee: write denial is local kernel state. Removing the executable there
can give the remote process `ESTALE` or `SIGBUS`, rather than merely denying
a later exec. This design does not qualify cross-host shared-home collection.

## In-use evidence

Only after all other retention checks pass (including references, readable
process evidence, permissions, future publication time, grace and lock
ownership) does the sweep probe a live removal candidate. Builds still in
grace are never write-opened. A successful writable open briefly prevents
a concurrent exec of that inode with `ETXTBSY`, so its lifetime is kept to the
inode check and immediate close. Inotify observes no close-write event for
in-grace, future, referenced or shared-mode rows; an aged removal row is the
positive control.

For every executable remaining in a Linux removal candidate, the sweep
pins its regular inode without following links through the proven cache
directory handles, then reopens its own descriptor for writing through the
kernel `/proc/self/fd` link. This prevents basename replacement from
redirecting the write-open onto another inode or a device.
It does not create, truncate, chmod or write the file, and closes a successful
open immediately after checking that it names the expected regular inode.
`ETXTBSY` retains the build. Every other probe failure retains it too, with
the build or retired member, executable name and error in the warning.

Write denial belongs to the executing inode, regardless of the executor's
UID, capabilities, container or dumpability. It therefore covers ordinary
root and capability-bearing execution as well as an unreadable same-user
process. It is not a claim that privileged filesystem mutation is harmless.
A successful open alone never proves absence: readable `/proc` paths and
inode identities remain additional retention signals, including an external
hard-link alias unlinked after launch. Existing single-link, owner, regular
file, receipt, hash and size validation remains mandatory for a live entry.

The scan reads real, effective, saved-set and filesystem UIDs. Any match
keeps readable process evidence in scope. Foreign or unreadable images,
including a non-dumpable same-user `ssh-agent`-like process, no longer block
the entire sweep: mandatory per-candidate probes independently protect their
running executables. Failure to list the process table still refuses the pass.

For every non-root sweep, the mode guard also retains group/other-accessible
or setuid/setgid executables. It applies even when `hidepid=2` exposes no
foreign PID. Mode is read at sweep time, not exec time. A process that started
before its owner changed the mode to owner-only is independently protected by
the write probe and any readable inode evidence.

Interrupted `.sweep-*` removal does not need intact receipts. The sweep walks
remaining files relative to the protected directory, without following
symlinks, and checks each remaining executable before finishing deletion.
Missing receipts, artifacts or `bin` directories do not permanently orphan
an otherwise unused retired tree. Busy or probe-error remnants are retained;
external symlink targets are never traversed or removed.

After renaming a live candidate into the retired namespace, the sweep probes
the retired tree again immediately before `RemoveAll`. A busy or uncertain
remnant stays retired for a later sweep. This narrows the key-path exec window.
An exec through an already-open descriptor, bind mount or the retired name
between the final probe and deletion remains the bounded race exception
named above; the snapshot and probes are not a global exec barrier.

The manager-home mutation lock and entry-boundary checks remain in place.
macOS and Windows keep their existing native enumeration and path matching;
the Linux probe and unreadable-image treatment do not apply there.

## Kernel history and qualification boundary

[2a010c412853](https://github.com/torvalds/linux/commit/2a010c41285345da60cece35575b4e0af7e7bf44)
removed exec write denial during the Linux 6.11 cycle.
[3b832035387f](https://kernel.googlesource.com/pub/scm/linux/kernel/git/vfs/vfs/+/3b832035387ff508fdcf0fba66701afc78f79e3d)
reverted it on 2024-11-27 after a userspace regression.
[0357ef03c94e](https://kernel.googlesource.com/pub/scm/linux/kernel/git/jack/linux-fs/+/0357ef03c94ef835bd44a0658b8edb672a9dbf51)
then deliberately exempted pre-content-watched executables. The flag is fixed
for that open file's lifetime; removing a watch later does not restore denial
for that existing execution. The exception remains in
[upstream Linux 7.0](https://github.com/torvalds/linux/blob/v7.0/include/linux/fs.h#L2809-L2821)
and the [upstream revision inspected for this work](https://github.com/torvalds/linux/blob/d48723711bae9b8cdc56719c448ba42fef28cbee/include/linux/fs.h#L2861-L2873).

This design assumes the restored write-denial behavior; it does not qualify
the unreverted Linux 6.11-era regression or arbitrary vendor patches. The
hosted job records `uname -r` and checks actual `ETXTBSY` behavior on its
kernel. That result must not be transferred to an untested downstream kernel.
The accepted HSM exception applies even on otherwise qualified kernels.

## Controls

`Linux unprivileged process sweep` creates a dedicated non-root UID on hosted
Ubuntu. Real cached test binaries exercise ordinary, linked-alias, unlinked
alias and non-dumpable images. Each busy probe must return `ETXTBSY`, retain
the build without changing its bytes, and allow collection after the process
stops. An unrelated unused build must collect while the non-dumpable child
is still running. This is an `ssh-agent`-like dumpability witness, not a real
OpenSSH invocation.

A real `EACCES` write-open row retains and names a read-only executable;
restoring its write permission is the collection control. Native snapshot
rows finish partial retired trees, retain running retired aliases and probe
errors, then finish after the cause is removed.

An inotify watch observes close-write events rather than inferring probe
placement from source. A late-exec row starts a real non-dumpable image after
rename: the re-probe retains its retired tree, and stopping it permits cleanup.
A late probe-error sibling retains and names its retired tree until repaired.

An empty process-table fixture models hidden foreign PIDs for the non-root
mode guard; it does not mount a `hidepid` filesystem. The declaration row names the exceptions and per-kernel
boundary without qualifying HSM or cross-host execution.

`python3 .github/ci/linux-sweep-control.py` plants two defects separately:
a compile-valid UID misclassification and an ignored `ETXTBSY`. Each must
fail at its intended assertion, and exact source bytes are restored in a
`finally` block. UID omission alone cannot defeat the independent probe, so
its own classification row kills it. A compile failure or unrelated timeout
is never a successful kill. No native root/container/capability or HSM
execution is claimed by these controls; the executor-independent mechanism
and HSM boundary above are derived from the kernel sources.

Run the focused controls under a private absolute HOME with Git, SSH, agent
and XDG selectors cleared:

```sh
go test -count=1 -timeout=600s ./internal/buildcache \
  -run '^TestLinux(UnprivilegedSweep|ProcessUIDControls|ForeignImagePermissions|ExecutablePaths|ExecutionProbeErrors|RetiredNativeSnapshot|HSMExceptionContract|ProbeRemovalCandidates|RetiredReprobe)$'
```
