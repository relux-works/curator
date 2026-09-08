# TASK-260908-fdg3gn publication recovery

Ready for review; canonical independent acceptance and signed delivery remain next.

Base: 84747c326eee9863ddfd7e86ac65be1056718fbc, managed Story branch.
The three preserved producer code/test/harness files match the recovery manifest
byte-for-byte (SHA-256 below). Only README content was integrated: retained the
composition API section and tool row from PR9, restored the systemprompt API
section/tool row, and clarified that main API integration remains pending.
No code redevelopment, commits, main/SPEC changes, registry/home writes, installs,
CI, releases, actual launches, LOGBOOK or private/control-record writes.

## Checklist disposition

The supplied publication instruction supersedes independent acceptance as a
producer prerequisite; review is still required next. Two conditional rows were
still present in the live checklist. Row 11 is acknowledged as inapplicable:
no production gate inspects source text (the mutation injector is not such a
gate). Row 12 is acknowledged as inapplicable under the explicit LOGBOOK-write
prohibition; this attached outcome is persistent evidence. Neither check claims
that a token-preserving static-gate mutant or LOGBOOK write was performed.

## Evidence provenance and bounds

Accepted from attached TASK-260908-fdg3gn_producer.md and TASK-260908-fdg3gn_logs.zip
from RUN-260908-ac40f7: 11 of 12 AC rows driven, exact production call sites and
named tests in the producer table, 11 of 11 narrowing mutants killed with named
failing tests and expected-red exit 1. No mutants were rerun in recovery.
Main launch integration remains 0 of 2 modes by scope. Reviewed signed delivery
is the remaining AC bound. The full mutant table and survivor bounds remain in
the attached producer report; no new mutant result is claimed here.

Executed during recovery: manifest assertions exit 0; git diff --check exit 0.
Tool readiness: task-board 0.24.3-330-g5ec20de4, Go 1.25.5 darwin/arm64,
Git 2.50.1, Python 3.14.7, Make 3.81. README diff self-reviewed.
Configured make check is delegated to the immediately following normal runtime
handoff, which must validate the integrated candidate before publication.
This pre-handoff artifact does not pre-claim its exit code; runtime handoff
validation evidence is authoritative. No redundant manual broad suite was run.

## Preserved SHA-256

- `.scripts/systemprompt-mutants.py`: `36a5c69d50af75dd0e40ad64a66949521b0b06d383b11f36a0c9363a0ecf6a9c`
- `internal/systemprompt/systemprompt.go`: `f6b62900581d8abb3be2df0bf58d7098d43abdb7eae345bcbb1aaa2e9e578528`
- `internal/systemprompt/systemprompt_test.go`: `4ebb898e292d5eff788b1f7bb2ea69af642534ee39cf80087d4b1c4a37b660d9`
