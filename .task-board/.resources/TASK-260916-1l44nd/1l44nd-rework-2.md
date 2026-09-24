# TASK-260916-1l44nd rework 2 (orchestrator, binding)

Revision 2 gate FAILED on BOTH ubuntu lanes (run 35671704019; 33 rows, one root cause):
`script_execution_capability_evidence_invalid: cannot enforce the Landlock ruleset` — i.e.
`landlock_restrict_self` now fails where rev1's `add_rule` failed. Lint is green. Nothing on trunk
uses Landlock yet, so there is no precedent to copy — implement it correctly:

1. `landlock_restrict_self` returns EPERM unless the calling THREAD has `no_new_privs`
   (`prctl(PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0)`) or CAP_SYS_ADMIN; it restricts the calling thread
   only, and Go is multi-threaded. The only correct place is the WORKER process, on a locked OS
   thread, immediately before the interpreter is spawned, from that same thread:
   `runtime.LockOSThread()` → `prctl(PR_SET_NO_NEW_PRIVS,1)` → `landlock_create_ruleset` with the
   `handled_access_fs` mask the probed ABI supports (ABI 1–4 masks differ; never request rights the
   ABI lacks) → `landlock_add_rule` per path with rights valid for the object type (directory rights
   only on directories; `/dev/null`, stdio and regular files: `READ_FILE|WRITE_FILE` at most, or
   open them BEFORE restricting and add no rule) → `landlock_restrict_self(fd, 0)` → fork/exec of
   the interpreter from the locked thread (Go's `syscall.forkExec` forks on the calling thread, so the
   child inherits the domain). Any error before restrict_self must not be reported as
   "capability evidence invalid": it is the probe/apply outcome of ONE control
   (`filesystem-write-confinement` / `descendant-exec-denial`) and belongs in the evidence record as
   applied/unavailable per the vector's rules — only a contradiction between probe and evidence is
   `capability_evidence_invalid`.
2. Never call `landlock_restrict_self` inside the test binary or the manager process: Landlock stacks
   at most 16 layers per thread (E2BIG on the 17th) and would confine the test process itself.
   Unit tests exercise ruleset/rule construction only; the enforcement rows drive the built worker
   binary at the process boundary (ubuntu lanes), asserting the confinement effect (write outside
   the derived path set refused with EACCES, inside allowed; exec denial when applied) and the
   evidence entries. Rows must pass identically under `-race`.
3. Probe: `landlock_create_ruleset(NULL, 0, LANDLOCK_CREATE_RULESET_VERSION)` for the ABI, plus a
   real `PR_SET_NO_NEW_PRIVS` capability check, decides available vs unavailable on this host; the
   hosted ubuntu runner (kernel ≥ 6.8, ABI 4) must probe available; when it does not, the vector's
   `linux-*-probe-unavailable` rows still succeed (never refuse on a fixed/unavailable control).
Continue from the revision-2 tree (no checkout/clean/stash); append "Revision 3" to results.md
(the enforcement design: thread locking, no_new_privs, ABI mask, rule typing; where the apply
happens; lane status); republish only on a green gate. No new scope.
