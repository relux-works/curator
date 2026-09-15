# Verified recovery publication sequence

Both installed public recoveries succeeded: composition RUN-260909-49fa72 exported/applied current launcher3ff66a9 treeff61be4a; source RUN-260909-1b402c exported/applied current sourcea8c6abdc tree922a7856 including independently proven board bytes. Old acceptances remained immutable.

Important runtime boundary observed in real execution: a run whose immutable StartStatus is integrating skips new CR publication even after start-landed-rework and handoff. to-review alone is insufficient evidence. Start a NEW ordinary developer owner run after the authorized recovery transition. Source inspected: cmd/spawn_workspace.go resolveIntegrationSpawnBinding keys on current integrating status; internal/spawnruntime/changerequest.go publishChangeRequestForProducer skips StartStatus integrating. No private record changes or source workaround required.

Composition follow-up RUN-260909-e08d63 completed publication: public worktree status confirms CR2 ready with21paths. Independent Astra review RUN-260909-b97efe is active. Source follow-up RUN-260909-ebc75e completed child handoff and runtime publication/validation remains active; CR3 not yet confirmed. Do not repeat recovery patching or accept old revisions for new trees.

Gate completeness TASK-260909-3d1589 rev2 review RUN-260909-d3717d requested the bounded automatic-block short-circuit repair; manual block and underlying registry/runner verified, automatic block masked primary refusal with later self-check success. Muse rework RUN-260909-52361f active, using exact reviewer replay regression. Diagnostics original task remains unaccepted, full launcher/migration objective unchanged.
