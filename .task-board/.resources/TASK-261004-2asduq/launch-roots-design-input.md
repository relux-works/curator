# Design input from the cocoaskills orchestrator (a2a #curator launch-project-roots-idea, 2026-10-04T00:21Z)

Read-only design input, not a scheduling request; the operator has not decided. Pinned to curator 876127f, spec 43bf0a2, launcher 27cc242. The full document stays with its author.

Scenario: work on project X, launch with `curator run` and profile Y. X's admitted project skills and commands are layered over Y and callable as ordinary commands. Nothing is sourced from the repository.

Recommendation: a persistent manager-owned project × profile home, `environments/project-views/<checkout-id>/<profile>/<env>/`, with a composed skill view.
- Composition: the admitted project generation P overlays Y's lock. The same pin deduplicates. The project version replaces the whole same-identity skill. Different skills exporting one command refuse. An incompatible replacement of a profile-required provider refuses.
- Y's root context, prompt, MCP and permissions are kept. The project adds only skills and commands.
- A derived composition record, never a new lock; Y is never mutated.
- Leased for the session's lifetime. Two projects on the same profile get different homes. Claude on macOS may need a first login per new home, and that is an open owner decision.

Commands: ONE manager dispatcher dir, selected through a typed fragment channel.
- A closed `command_environment` object: its own version, the dispatcher dir, a manager context ref and digest, and the registered project root or null.
- The operation is fixed to append, applied on the admitted filtered child PATH.
- The checkout `.agents/bin` is never an append root.

Spec and launcher work:
- the next unused outer fragment revision; v3 is already Muse, so v4 is the candidate;
- project-view home layout, composition and lease rules (environments §§7-10, manager §§3/12);
- a revised §9.4 singleton-current command limitation;
- the dispatcher dir added to the provider-discovery refusal set, so a skill cannot supply `curator-run`.

Tracked mode needs a destination-side PATH-append operation in the launch plan (today: own literals only, no PATH transform). Until then, refuse tracked requests for this capability.

The shared model for both managers: manager identity → checkout/context → admitted generations → whole-skill overlay → command owner → protected activation. The fallback differs: csk falls back to the global install, Curator to the selected profile Y, never to machine-current.
