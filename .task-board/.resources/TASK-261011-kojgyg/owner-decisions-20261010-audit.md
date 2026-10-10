# Owner decisions, 2026-10-10, from the docs-audit questions

## D-GOALS: agent self-goals outside a board, variant (b) for now
- Now: only the board/orchestrator (board goals) or the launching caller at start (a goal passed through the launcher) sets a session's goal. The agent account gets no goal capability on the session host socket.
- The docs must say WHY this is the choice for now, and that the final choice between (a) agent self-goals in a separate namespace and (b) is deferred to a separate SH2+ item.
- Hard requirement (owner): no strict coupling to goal functionality, to our orchestrator or to its toolchain. In the module, task-board is an OPTIONAL component. The session host must work without a board and without goals. Board goals are one optional client of a generic goal interface, not a dependency.
- Targets: swarma-session-host spec §3.8 and README (R-SH1), wiki architecture goal-namespace section (R-WA1), book ch07/08.

## D-IDENTITY: identity revocation at worker retirement
- One-shot workers: the dispatcher revokes the identity itself at cleanup, as a separate operation after the account is retired (dispatcher §14). The dispatcher's right covers only identities it registered as one-shot, and that flag is recorded at registration.
- Standing agents (orchestrator, fixed team roles): retiring the account does not touch the identity. Only the operator revokes it, with a signed record.
- Targets: swarma-dispatcher spec (R-DP1), key-keeper/architecture (R-WA1), swarma-user-manager retire wording (R-UM1), book ch04 and ch13.

## D-EXECUTOR (2026-10-11): one-shot execution and first broker consumer
- Q1: amend Decision 0021 with a versioned, narrow agent-account one-shot mode for curator-run (variant a). Rails, proposed to the owner: under the agent account curator-run never composes or resolves profiles; it only verifies and executes a plan composed on the caller's side (a dedicated entry such as `curator-run exec --plan-fd N`, no network except the broker); one shared launch library with swarma-session-runner; a separate minimal executor binary is a later option.
- Q2: the first broker consumer is curator-run under the agent account. The board runner and the session host do not embed a broker client; they ask the dispatcher, which launches through curator-run.
- The owner confirmed (2026-10-11) the curator-run rails above (narrow exec mode, compose on the caller side only, shared launch library).

## D-PATH (2026-10-11): launch-scoped PATH recipe, accepted
In order: (1) the protected tool directory of this launch (CIP-0007); (2) the admitted base PATH of this user, in its existing order; (3) one protected manager dispatcher directory, through which project commands run only after checks. A repository `.agents/bin` is never executed in a managed launch. The resolved PATH is recorded in the launch plan and covered by its digest.

## D-DONOR-PROVISIONING (2026-10-11)
A separate small donor-side component, with its own short spec, rules and removal. It may share libraries (UID and path checks), but the binary and its privileges are separate from swarma-user-manager. Variant B stays disabled until this component is specified.

## D-WIRE-NAMES (2026-10-11): clean names now, while free
Versioned wire identifiers carry the repository name: `swarma-<component> <object>/<n>`. Examples: `swarma-credential-broker grant/1` and `swarma-credential-broker revocation/1` (replacing `curator-broker …/1`), and `swarma-dispatcher request/1`. Apply the same rule to the other versioned wire identifiers of the platform modules (user manager request kinds, launch ticket). Nothing has shipped, so no migration is needed. Record the rule once in the contract owner documents.

## D-RESUME (2026-10-11): keep it simple, trust semver
Native resume of a session on the same machine is allowed without an extra qualification row when the harness version change is non-breaking by semver: same major, and for 0.x the same minor. A different major (or 0.x minor) is refused. If the harness itself fails the resume, the session is reported as resume-failed, and the caller decides what to do next. No automatic fallback machinery on this path; the portable seed is not part of resume for now. A per-harness note records any vendor that is known not to follow semver.
