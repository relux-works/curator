# BUG-260928-2bfhgj — rose-air: scriptworker forged-identity tests close the channel (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md` and the bug description. Only the self-hosted `Test (rose-air)` lane fails, and only these two tests
(run 36387486082, artifact test-evidence-rose-air): internal/scriptworker TestScriptWorkerRejectsForgedWorkerIdentity (worker_test.go:546)
and TestScriptWorkerRejectsSubstitutedManager (worker_test.go:601): "cannot read a worker message: worker session channel closed". They were
hidden since 2026-09-24 because the lane failed earlier at the Rust step (fixed in 2252ebee).
1. Diagnose: read the two tests and the worker/manager handshake they exercise; list what differs on rose-air (see the lane's own diagnostics
   in the run log: uname -m/arch, macOS version, runner user, TMPDIR, codesign/quarantine, sandbox-exec availability, SIP) — the worker
   process likely dies (killed / exec refused / sandbox profile refused) before it can emit the refusal, so the test sees EOF instead of
   the expected rejection. Prove it with evidence (the lane log, or a diagnostic you add to the test failure message).
2. Fix the root cause without weakening the refusal: the manager must still reject a forged worker identity and a substituted manager.
   If the difference is an environment capability (e.g. the enforced launch path needs a host control rose-air lacks), the test must
   assert the specified fail-closed diagnostic for that case rather than EOF — cite manager.md §3.1 / core §4.1.1 for what the manager must
   report when a host control is unavailable.
3. The rose-air lane runs only on main pushes (check .github/workflows/ci.yml); if you need rose-air evidence before landing, say so in the
   results (the orchestrator decides) — do not push to main. Local macOS runs with real exit codes; improve the failure message so a future
   channel-closed failure names the worker's exit status/stderr.
No CHANGELOG/LOGBOOK edit (entry text in results). Update the results resource, handoff, END YOUR TURN. Write only inside your Story worktree.
