# Review note — BUG-260928-2bfhgj rose-air scriptworker tamper fixture (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev1 (base 2252ebee, tree 3db1bec3, 1 path internal/scriptworker/worker_test.go, gate green) against `2bfhgj-brief.md`. The producer's
root cause: on darwin/arm64 the kernel SIGKILLs a Mach-O whose code signature no longer matches its bytes, so the byte-flipped fixture
never runs and the test reads EOF; the fix re-signs the tampered copy ad hoc under another identifier on darwin.
Scrutinise:
1. The diagnosis must also explain why hosted macos-latest (also arm64) passed with the byte-flip (e.g. which bytes the flip hits —
   inside the signature blob / LINKEDIT vs a signed page; hosted vs self-hosted AMFI/SIP state; Go build flags). If the explanation does
   not hold, say so — a fixture that only works by accident on hosted runners is itself a finding.
2. The tests still prove the product refusals: the tampered/substituted binary IS launchable and DIFFERS in identity (bytes and
   code-signing identifier), and the manager rejects it with the specified refusal — not EOF. Mutant: disable the manager's identity check →
   both tests must fail on macOS locally (real exit codes).
3. Non-darwin behaviour unchanged; `codesign` absence handled (skip vs fail?) per the repo's platform-case ledger rules; the failure message
   now names the worker exit status/stderr as the brief asked.
accept_cr or changes requested with file:line. No LOGBOOK.md.
