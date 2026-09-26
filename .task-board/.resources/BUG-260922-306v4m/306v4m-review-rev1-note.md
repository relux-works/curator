# Review note for BUG-260922-306v4m revision 1 (orchestrator, binding) — rose-air diagnostics

Read the brief `306v4m-brief.md` and `BUG-260922-306v4m_results.md`. Context: the operator confirmed
rustup IS installed on the rose-air machine, yet `.github/ci/install-rust-toolchain.sh` fails every
main push with "rustup is not installed on this runner". This leaf only makes the FAILURE path print
evidence (RUNNER_NAME, RUNNER_OS, hostname, whoami, HOME, CARGO_HOME set/defaulted,
HOMEBREW_PREFIX, the searched PATH, per-candidate absent/exists/executable, bounded rust/cargo
listings, `command -v`/`type -a`) before the unchanged remedy sentence. No probe path was added.

Verify:
1. The success path is behaviourally byte-identical: same `GITHUB_PATH` writes, same PATH export
   order, same `rust-pin: using rustup at <path>` line — diff the script and reason about every
   touched line; a changed success path is a finding.
2. The block prints ONLY the named variables (no full environment dump, nothing secret) and cannot
   itself fail under `set -euo pipefail` (unset HOMEBREW_PREFIX, missing directories, `type`
   absent in `sh`) — a diagnostic that crashes before the remedy line is a finding.
3. The self-test rows exist and are executed: absent-everywhere fixture → block + remedy last +
   exit 1; rustup-under-a-candidate fixture → success, no block; a narrowing mutant (drop one
   candidate from the enumeration) killed. Fixtures must not depend on the host's real rustup.
4. Docs paragraph + CHANGELOG entry present.

Record exactly one verdict: `accept_cr(BUG-260922-306v4m, revision=1, evidence=<your outcome
resource>)` on ACCEPT, or changes-requested with file:line and reproduction. No LOGBOOK.md writes.
