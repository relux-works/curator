# THE ONLY CURRENT INSTRUCTION — TASK-261013-vtmjnu: apply the wave-5 owner decisions (worker code transfer, untrusted class, abuse controls) (W-CS6, developer)

Repository: the Curator specification (`cips/CIP-0008-remote-worker-launch-mode.md`, `cips/CIP-0009-donor-side-deployment-and-bridge.md`, `protocol/remote-worker.md`; CIP-0010 only where it is affected). The preconditions are:
- `gw5-decisions.md`: the owner decisions of 2026-10-13 and the shared names;
- `gw4b-decisions.md` and `gw4-decisions.md`: the earlier names.

The owner decisions take precedence over any older text.

**Do this:**
1. **R-TRANSFER** (CIP-0009 D5 `swarma-workplace-exec/1`, tasks and results, the publication register).
   - Replace "output review and export of changes as reviewed patches" with the R-TRANSFER round:
     - per-round materialization from the admitted base or from the last candidate;
     - round end and the write fence;
     - export by the approved executor;
     - candidate, review and rework on our side, with rework sent as signed task letters;
     - an intermediate export at a turn boundary.
   - Adopt the RWP gate list.
   - Register the design labels as unpublished.
2. **R-NET-TOOL** (CIP-0009 network rows, CIP-0008 lockdown).
   - Workplaces have zero egress: remove the proxy references for workplaces.
   - **Default: no web.** The donor's loop reaches only its model-provider endpoints and our host. No search or fetch tool is offered, and provider-side search stays disabled in the lockdown. The lockdown plan is the only control for it, because it travels inside the provider API, where no network rule can see it.
   - State the rationale: web content mixes contexts and invites injection.
   - **Task resources.** Tasks carry the resources they need as attachments of the signed task letter, materialized read-only next to the workplace's workspace, with digests. The worker asks for more by letter; it never fetches.
   - **Trusted circle.** Describe web search as an opt-in, only for a remote worker that the operator admits to the trusted circle by an operator act at admission. Never for the untrusted class. Two options:
     - (a) provider-side search per a qualified tuple;
     - (b) the donor-side `swarma-web` MCP server holding the donor's own search key, with the custody rules from the decisions.
   - Option (b) runs under its own donor-side identity in variant B, or its own sandbox profile in variant C. It never runs in the harness identity, whose egress is the provider allowlist only, and never in the bridge identity. It holds the donor's key there and reaches only the search provider's API endpoints.
   - The lockdown is otherwise unchanged. Both options are pending qualification.
3. **R-EXEC and R-VERIFY** (CIP-0009 tool set and D5).
   - Workplace `exec` calls run under the exec profile.
   - State the excluded categories.
   - Heavy work runs only in the verification phase.
   - Dependency caches are read-only.
4. **R-STANDING and R-UNTRUSTED** (CIP-0009 admission and lifecycle).
   - A workplace is untrusted-class unless the operator admits that worker to the trusted circle. That admission enables only donor-side opt-ins, and every boundary on our host stays.
   - A standing workplace is allowed, but never an agent loop on our host.
   - Admission follows the class requirements and the host capability matrix.
   - No subscription lease.
5. **R-ALLSTOP and R-WATCH** (CIP-0009 project-side lifecycle).
   - The node-wide stop and watcher narrowing apply to workplaces: stop the jobs, freeze, recover explicitly.
   - On the donor side, use the existing channel-down and suspended states.

**Rules:**
- Public repository: no first names, host names, local absolute paths, session links or board-resource links.
- Docs only: spec text, README, CIP text and PlantUML diagrams. Use conservative PlantUML syntax; this host has no renderer.
- Do not run tests or builds yourself. The configured CR validation at handoff is sanctioned.
- Never write LOGBOOK.md.
- Apply the owner decisions in `gw5-decisions.md` exactly.
- Use its shared names (R-TRANSFER, R-UNTRUSTED, R-NET-TOOL, R-EXEC, R-VERIFY, R-CEILINGS, R-WATCH, R-ALLSTOP, R-RANGE-DENY, R-STANDING, R-SOCKETS, R-SHARED) unchanged, together with the earlier names in `gw4b-decisions.md` and `gw4-decisions.md`.
- Do not invent wire fields, refusal codes or numeric values. Where a format is unpublished, say so and add or extend an open-question entry ("would need …").
- Remove superseded text; do not leave it as an alternative.
- Update every figure that shows changed behaviour, so no contradicting copy remains.
- Add a revision note at the top of the spec in its existing style.
- If an item is already satisfied by the current text, report that with section numbers instead of rewriting it.

**Hand-off:**
1. Attach `TASK-261013-vtmjnu_gaps.md`. In it, give for each shared name you applied the changed sections and figures, and list what stays open.
2. Run `task-board handoff TASK-261013-vtmjnu --role developer`.
3. END YOUR TURN.
