# TASK-260923-2elcdc — enforce the isolated environment lock (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md` and the task description. 0017 (TASK-260922-cww1ov, incl. F-C2 explicit migration) is on curator main;
build on it. Spec: curator-spec main — environments §12.2, manager §1 rule 1, system-config-v2 `environments.isolation` (F-S3, fe2d1c6).
1. system-config-v2 environments.isolation admits `isolated`; silence resolves to the locked direction; an explicit `shared` under an
   engaged isolated lock refuses with environment_isolation_lock_conflict; an already-provisioned shared passthrough fails closed toward the
   explicit F-C2 migration (never migrates silently).
2. Production-entry rows through the CLI for each rule, plus the pinned suite's (dcc7f015) isolation vectors if present (drive them).
3. Mutants: silent migration; explicit shared admitted under the lock; silence resolving to shared — each killed.
No CHANGELOG/LOGBOOK edit (entry text in results). Focused bounded runs (host memory is tight). Attach results, check DoD,
`task-board handoff TASK-260923-2elcdc --role developer`. A write-boundary `policy warn` block is a warning.
