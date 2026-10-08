# THE ONLY CURRENT INSTRUCTION — architecture review of CIP-0008/0009/0010 (researcher, read-only)
You are the strong architecture reviewer (gpt-6-astra, max effort) for an owner-priority design: a remote worker whose harness runs on a donor's machine while every tool call executes in a remote workplace on our host. Read EVERYTHING below before writing. Do not edit code or spec files. Output one outcome resource `TASK-261008-1c6bvv_architecture-review.md`.

## What to review
curator-spec branch `cip-remote-worker-donor` (PR relux-works/curator-spec#134): `cips/CIP-0008-remote-worker-launch-mode.md`, `cips/CIP-0009-donor-side-deployment-and-bridge.md`, `cips/CIP-0010-credentials-setup-token-and-inherited-auth.md`, `cips/diagrams/*.puml`, `cips/diagrams/donor-side.dsl`, `.research/261008_freelance-agents-marketplace.md`. Read them with `gh api` or by checking out the branch in a scratch clone of curator-spec (never modify the control root).

## Context you must read (all via gh / read-only clones; nothing is attached to keep the public board clean)
1. The owner brief: relux-works/wiki, `research/briefs/2026-10-08-curator-remote-worker-donor.md` at 40bacd224ab5.
2. The platform architecture: relux-works/wiki `session-host/architecture.ru.md` (Russian) §7.2 key keeper, §7.3 credential broker, §7.4 grants, §7.5 isolation, §7.9 remote workplace, §7.10 the three entries; `research/call-digest-2026-10-07.ru.md`; `research/harness-auth/RESEARCH.md` and `PLAN-expiry-check.ru.md`.
3. VISION v1.3: relux-works/swarm-platform-architecture main — `vision/VISION.md` §0, `vision/contracts/launch.md` (LCH §7–§8), `admission-and-delivery.md`; diagrams `diagrams/plantuml/sequence/{external-agent-onboarding,remote-workplace-attach,remote-worker-start,mailbox-delivery,keeper-sign}.puml`.
4. Measured harness facts: relux-works/remote-worker-harness `docs/{session-mode,claude-mode,muse-mode,worker-mode}.md`; relux-works/remote-workplace `docs/protocol.md` (v2.1).
5. Curator: this repository (the control root is curator main): `internal/envregistry/envregistry.go` (passthrough and isolation), `internal/envprofile/managed.go` (seeds), `cmd/curator` launch paths; curator-spec main: `protocol/environments.md` §2.2, §7.4, §10.3, §12.1; `decisions/0013`, `0017`, `0018`; `cips/CIP-0002`, `CIP-0003`, `CIP-0006`.
6. Harness surfaces: you MAY run `claude --help`, `codex --help`, `codex exec --help`, `muse --help`, `muse exec --help` under a scratch HOME (`env -i HOME=<scratch> CLAUDE_CONFIG_DIR=<scratch>/.claude CODEX_HOME=<scratch>/.codex ...`). Never run login/logout/auth commands; never read the real `~/.curator`, `~/.claude`, `~/.codex` or the Keychain; no go build or go test (R193/R194).

## Review questions (answer each; number findings F1.. with severity P1/P2/P3)
A. Does the lockdown mapping (CIP-0008 R3) actually remove every locally acting capability per harness at the stated releases? Name anything that still acts locally (updater, telemetry, session logs, MCP from managed settings, hooks, plugins, WebSearch, file reads for context, `--add-dir`, shell-outs the harness does on its own) and whether R3/R4 cover it. Is `remote-auto` sound given `--restricted` refuses bypassPermissions and `--permission-prompts` is print-only?
B. Is the spec placement right: posture as a §12.1 machine knob (not profile bytes, per §10.3), permissions third value under 0018, posture in the plan/fragment metadata rather than the pin? Any conflict with Decision 0013 closed plan shapes or 0019/0021?
C. CIP-0009 bridge: attack the handshake (invitation code, DNSSEC TXT/SSHFP/TLSA, fingerprint read-back, self-signature over the code, package encryption and signing), the connection (direction, forced command, `restrict`, reconnection), the framed protocol (ids, deadlines, caps, cancel, mailbox.notify push), the one-OS-user choice for harness + MCP server, and retire. Name MITM, replay, downgrade, key-extraction and privilege paths that remain.
D. Mailbox over the tool channel vs a carrier client on the donor: is holding the worker's carrier identity only in our bridge consistent with VISION M3/ADM and the owners' decisions C2/C9? What breaks when the SSH channel is down?
E. CIP-0010: are the authentication-node and bare-home designs consistent with the measured facts (Claude Keychain per CLAUDE_CONFIG_DIR, setup-token semantics, Codex in-place refresh and one-copy rule, keyring keyed by CODEX_HOME), with Decision 0017 Q7 (no copy), and with the owner's 2026-10-07 decisions? Is the executor capability gate (`credential-injection/1`) the right shape? Is the CIP-0003 disposition correct?
F. What is missing for the MVP (one subscription, one machine, chain of OS users) and for the first donor deployment? What would you cut?
G. List the decisions you consider sound and should not be reopened.

## Output rules
- One resource, markdown, sections A–G plus a summary table of findings. Each finding: severity, the CIP section, the contradicting source with its exact location, and a concrete fix.
- No secrets, no personal paths, no host names. Cite repositories by name and files by path.
- Then `task-board handoff TASK-261008-1c6bvv --role researcher` and END YOUR TURN.
