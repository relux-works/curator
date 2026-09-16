# STORY-260916-1on1d2: launch-credential-and-permission-modes

## Description
Operator request 2026-09-16: (a) a credential mode per environment — shared (managed home reuses the native login, as codex_cli does via auth.json file-link) or isolated (per-profile login reused across launches, as claude_code on macOS is today because its OAuth token lives in the Keychain and no darwin passthrough exists; pi links an empty native auth.json) — selectable rather than platform-accidental; (b) a first-class curator run permission interface (e.g. --permissions standard|yolo) mapped to each environment native flag (Claude --dangerously-skip-permissions, Codex --dangerously-bypass-approvals-and-sandbox/--yolo, Pi per its docs), explicit opt-in only, coordinated with --ax-profile. Process: Astra research report → Fable independent review → two spec-gap proposals filed as curator-spec decision drafts (proposed, not adopted) via PR plus tracking issues; implementation only after adoption.

## Scope
(define story scope)

## Acceptance Criteria
Astra report accepted by Fable review; two decision drafts landed on curator-spec main through signed PR with independent review; GitHub issues reference them; no normative or implementation change.
