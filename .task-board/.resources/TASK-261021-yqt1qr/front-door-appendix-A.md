## Appendix A. Proposed CIP-0016 amendment: delegated local launches

Proposed text, for `ivmbp_curator_orch` to file in curator-spec once Ivan
signs off Q1.

> **Amendment (proposed 2026-10-18): delegated local launches.** Until a
> session manager advertises the exact-plan successor capability, a
> registered local session manager MAY host interactive sessions through
> `launch-delegated/1` of the curator-run session-manager protocol.
>
> **Division of work.**
> - `curator-run` remains the front door. It resolves the profile and the
>   fragment, the model and effort, the permissions, the system-prompt
>   opt-in, the typed resume intent and the credentials.
> - It hands the manager a closed request that names them.
> - The manager's caller-side component composes the process through the
>   same agents-management construction site (Decision 0019), against the
>   profile's managed home.
> - The manager MUST refuse when its fragment digest differs from the
>   request's.
> - The manager MUST NOT add profile content, change any resolved value, or
>   enable Remote Control.
>
> **Stated limits.**
> - There is no session snapshot. A running session uses the live managed
>   home, so repair, profile switching and skill sync can change it
>   underneath the session. The CIP-0014 isolation does not hold.
> - "A host never composes" holds for the daemon only, not for the
>   manager's caller-side component.
> - The CIP-0016 digest-covered provenance is replaced by the request's
>   profile pin and fragment digest.
>
> **Retained rules.** Typed resume never creates a new session. Hosted
> refusals never fall back to native.
>
> **Activation.** It is per runtime: Claude Code and Codex first, Muse once
> the manager's v4 consumer is qualified. The joint three-runtime rule
> continues to govern the exact-plan release.
>
> **Retirement and backend selection.** The mode is declared per manager
> and per runtime. For a given runtime it is retired once the manager
> advertises the exact-plan capability for that runtime, together with a
> detached form; curator-run then uses exact plan for attached launches of
> that runtime. Until the exact form supports detach, detached launches
> keep the delegated route. Exact support for one runtime retires nothing
> for another. The backend-selection
> question of section 2 is answered by the curator-run session-manager
> registry. An absent backend selection becomes "no registered manager",
> and task-board's installer registers task-board.

