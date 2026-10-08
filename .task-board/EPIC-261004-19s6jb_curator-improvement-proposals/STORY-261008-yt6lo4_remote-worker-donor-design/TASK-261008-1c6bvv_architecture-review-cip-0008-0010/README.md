# Architecture review of CIP-0008/0009/0010 (remote worker on a donor machine)

## Description
Owner-priority research (2026-10-08). Review the draft CIPs 0008 (remote-worker launch mode), 0009 (donor-side deployment and bridge) and 0010 (credentials) on curator-spec branch cip-remote-worker-donor (PR #134) against the full context: the owner brief, the platform architecture (relux-works/wiki session-host/architecture.ru.md, VISION v1.3 and its diagrams in relux-works/swarm-platform-architecture), the current Curator code base and curator-spec rc.14 main, and the measured facts from relux-works/remote-worker-harness and remote-workplace. Output: a findings resource with numbered findings (severity, evidence, concrete fix) and a list of what is sound. Research only; no code or spec edits.

## Scope
read-only review; one outcome resource

## Acceptance Criteria
Numbered findings with severity and a concrete fix each; every claim cites a file/section; contradictions with curator-spec MUSTs, VISION contracts or measured harness behaviour are named; a short list of confirmed-sound decisions; no secrets, personal paths or host names in the resource.
