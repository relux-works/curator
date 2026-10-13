# THE ONLY CURRENT INSTRUCTION — TASK-261013-zaqbpr: apply the wave-5 owner decisions (worker code transfer, untrusted class, abuse controls) (W-LR6, developer)

Repository: the agent launcher (`SPEC.md`, `diagrams/`, README). The preconditions are:
- `gw5-decisions.md`: the owner decisions of 2026-10-13 and the shared names;
- `gw4b-decisions.md` and `gw4-decisions.md`: the earlier names.

The owner decisions take precedence over any older text.

**Do this:**
1. **R-TRANSFER.**
   - The composed plan carries the workspace section (base commit, manifest digest, write scope) and the exec-profile reference, all covered by the plan digest. Exact fields are unpublished and owned with the intake contract.
   - Replace the "materialization detail unspecified" note in `diagrams/artifact-layout.puml` with a reference to the dispatcher's workspace contract.
2. **R-STANDING.**
   - The one-shot executor verifies harness surfaces by digest before exec, or uses root-owned surfaces outside the home.
   - Drift refuses the launch (code unpublished).
3. **R-EXEC.**
   - The protected manager dispatcher directory (D-PATH) carries the light-check wrappers with fixed subcommand allowlists. Heavy entry points are not exposed.
   - The OS-enforced allowlist belongs to the user manager.
4. **R-NET-TOOL.**
   - A managed launch sets proxy variables only for the bound provider proxy. No other egress.
   - No web tool by default. Tasks carry their resources, and the plan covers their read-only materialization by digest.
   - The launch plan must switch off the harness's provider-side (server-executed) search unless the identity holds the operator-granted opt-in. No network rule can see that search, because it travels inside the model-provider API, so the plan is the only control.
   - Provider-side search, per a qualified tuple, or the platform web-tool service is only an operator-granted opt-in for trusted-circle identities.
5. **Figures.** Update every figure that shows these.

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
1. Attach `TASK-261013-zaqbpr_gaps.md`. In it, give for each shared name you applied the changed sections and figures, and list what stays open.
2. Run `task-board handoff TASK-261013-zaqbpr --role developer`.
3. END YOUR TURN.
