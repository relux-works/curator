# Wave 4, part b (2026-10-12): F-23, operator custody. Owner decision and orchestrator resolutions

The precondition `gw4b-f23.md` holds the finding text. The owner accepted it with the amendment below, which takes precedence. The wave-4 shared names in `gw4-decisions.md` still apply.

## Owner decision (operator, 2026-10-12): apply exactly

1. **A separate operator account.** It never hosts a harness session. It, not the person's interactive account, is:
   - the configured broker operator and account owner;
   - the user manager's `operator` UID;
   - the holder of the operator signing authority, including the broker v0 root and the operator records (`team/1`, standing retirement, maintenance authority, explicit home retention, standing re-grants).
2. **The main path is a key in the phone.**
   - The operator key lives in the phone's secure element (Secure Enclave or StrongBox). It never leaves the phone and is usable only after biometric confirmation.
   - The phone app reaches the machine by SSH over the tailnet, into the operator account. Login is by that key only, and there is no shell: a forced, narrow operator command accepts operator acts only.
   - Every act is signed on the phone after the app shows its exact content (who, which actions, until when). The machine verifies the signature against the pinned phone key. The signature is bound to that one act's canonical content and cannot be reused for another act.
   - Revocation is a single biometric confirmation.
3. **The fallback path is the terminal.** From a computer: `sudo` into the operator account with a password, no `NOPASSWD` and no cached ticket (`timestamp_timeout=0`).
4. **A read-only account is mandatory.** Reads (status, grants, accounts, bindings) go through it without any confirmation.
5. **Password prompts.**
   - Widening acts need authentication plus confirmation of the act's content: grant issue, subscription enrolment, `team/1`, standing retirement or re-creation, explicit home retention, maintenance stop authority, renewal of standing grants. The operator CLI refuses them from inside an agent session.
   - Narrowing acts (revocation, disabling an account) need authentication only.
   - `NOPASSWD` stays only for the dispatcher service account's helper and launcher rules.

## Orchestrator resolutions: shared names

- **R-OPERATOR.**
  - The operator account is `swarma-operator`; the read-only account is `swarma-observer`. Both follow the service naming rule.
  - The forced command is the **operator command**; its binary name and wire format are unpublished.
  - The phone key is a new key class: the **operator device key** (hardware-backed, presence-gated).
  - An operator act is a signed operator record over its canonical content. The signature domain and the canonical bytes are unpublished; where a spec already names a signed operator record (`team/1`, standing retirement), the operator device key is one admitted signer.
  - Key enrolment, rotation and loss of the phone are an open question, with a terminal re-enrolment under the fallback path as the proposal.
- No new hardware is mandated beyond the operator's phone; the terminal fallback works without it. No signature-algorithm change is mandated.
- **Scope.** Docs only: spec text, README, CIP text and PlantUML diagrams. Update every figure that shows operator acts. Add a revision note at the top of each changed spec in its existing style.
