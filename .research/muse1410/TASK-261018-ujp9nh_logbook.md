# TASK-261018-ujp9nh qualification logbook

- Round 1 correctly stopped intake after the auto-updater replaced the named
  1.4.10 executable with 1.4.11. Round 2 uses the task-board Muse pin store,
  with the exact expected digest verified before a disposable local copy.
  The manager's pin store was not modified.
- Current spec main was advertised and fetched as
  `e41f96387287f5e76604627f9e63662f7d6d86d4`. Environments §§7.1, 7.9,
  10.2, CIP-0017 and the Muse v4 conformance README were reread.
- Exec P1–P4 loader replay passes on 1.4.10-R6427.1. Interactive required
  exclusion still retains one foreign skill with either a linked or frozen
  regular root, reproducing 1.4.8. Its §7.9 gate remains pending/refused.
- The brief anticipated an exec refusal for interactive `-c`; the actual
  binary accepts it, exit 0. The first assertion failed, exit 1; the record
  retains that evidence and a separately named exact-build observation test.
  The extra key's exec exclusion semantics remain unmeasured.
- Reused the 1.4.8 fixture bytes and original reviewer regressions. The
  current Curator install/resolve entrypoints reproduce the frozen root.
  Snapshot metadata is reused representation evidence, not new provenance.
- Narrowing controls kill a real exec exclusion omission, a one-skill
  checker exception, contradictory count/outcome exceptions, snapshot extra
  member/link omissions, a synthetic foreign-digest allowance and a
  localhost-only network allowance. Interactive omission survives because
  the vendor control already fails; it proves no interactive protection.
- Production/default tests/configuration remain unchanged from the green
  checkpoint CI run 39193852508. Hosted evidence is reused for those exact
  inputs; this is not an external-CI mirror or a synthesized status. The
  opt-in module and published reference CLI are exercised locally.
- See the qualification record, sanitized observations, command logs and
  task gaps record for boundaries, measured ratios and individual exits.
  1.4.11 is the next exact build to qualify. Consumer changes remain outside
  this leaf.
- Local unpartitioned default test exited 1: three packages exceeded the
  default 10-minute budget, and three CLI status tests discovered ambient
  untrusted providers. A controlled-PATH replay of those three tests exited
  0. Full real Go-discovered test lists are partitioned into bounded slices;
  remaining default packages run separately. The failing initial command
  is retained as failing evidence, with no production/test changes to hide it.
- Concurrent local slices encountered the repository's shared host-GOROOT
  lock. CLI slice 0 timed out, slices 1–3 refused before their bodies;
  install slice 0 passed 93 entries. Other owned waiters were terminated
  by recorded PIDs and remain incomplete/unrun. No guard was bypassed.
- Downloaded and checked the exact checkpoint's original macOS CI artifact:
  85 of 85 default packages accounted for (80 pass, 5 no tests), zero fail
  events; job log records Go stage/overall exit 0 and platform gate exit 0.
  Reuse that green evidence for unchanged default-source/test/config inputs
  and its original hosted environment, per standing order 10. Do not claim
  local full-suite success. New opt-in harness remains separately green.
