# THE ONLY CURRENT INSTRUCTION — platform documentation audit: book, specs and diagrams (researcher, read-only)

## Goal (owner, 2026-10-10)
Make sure that:
- everything the Swarma book says is covered by specifications and documents in the repositories;
- the needed diagrams exist everywhere;
- the overall architecture, each module's own internal architecture, and how the modules play together are explained clearly and covered by diagrams;
- all of it is of good quality and understandable, in the repositories and in the book itself.

This task is the audit and the fix plan. The fixes are later tasks.

## Sources (read-only via `gh`)
- **The book:** relux-works/wiki `book/swarma/ru/*.md`, `book/swarma/src/*.puml` and the renders in `book/swarma/diagrams/`. Also the platform architecture `session-host/architecture.ru.md`, `glossary.ru.md`, `naming.ru.md`.
- **Platform modules** (public): `relux-works/swarma-credential-broker`, `swarma-user-manager`, `swarma-dispatcher`, `swarma-session-host`. Each has a README, `spec/` and `diagrams/`.
- **Curator:**
  - curator-spec PR #134 (branch `cip-remote-worker-donor`): CIP-0008…0011 and `cips/diagrams/`;
  - curator-spec PR #136 (branch `cip-operator-decisions-1010`): CIP-0002…0007;
  - `relux-works/curator-agent-launcher` (`curator-run`).

## Produce one research document with
A. **Coverage matrix.** For every decision or normative statement in the book, by chapter and section: the specification that defines it (repository, file, section and line), or MISSING, or CONFLICTING with the exact contradiction. Stale names count as findings: the final names are `swarma-*` components, `swarma-session-runner`, `curator-run` for one-shot runs and `worker-<label>` accounts.
B. **Diagram inventory.** For every repository and for the book: each existing diagram, its kind (C4 context, container or component; sequence; state; deployment or layout), what it shows, and whether it is current. Then the missing diagrams, each specified (kind, participants, the flow or states, the file to create):
   - per module: its internal components; its main and failure/recovery sequences; the state machines of its records; its on-disk layout;
   - across modules: the platform container map with the final names, and end-to-end sequences for spawning a one-shot agent, a long session with goals across accounts, a credential lease and its revocation, waking an agent, and a remote-worker turn.
C. **Explanation quality.** For each document: can a new engineer and the owner understand its purpose, boundaries, interfaces and interplay? List concrete problems: undefined terms, missing "why", contradictions between documents, walls of text that need a figure or a table.
D. **Fix plan.** Concrete, sized tasks per repository (files and diagrams to write or change), and a separate list for the book. Order by value to a reader.

## Rules
Read-only; cite file and line for every claim; no execution on this host. Deliver the study as a `.research/` file in the Change Request and attach it as the resource `platform-docs-audit.md`. Never edit `LOGBOOK.md`. No secrets, personal paths or host names. Then `task-board handoff <TASK> --role researcher` and END YOUR TURN.
