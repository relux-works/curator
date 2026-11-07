# Task and outcome handoff

An element ID alone does not locate its board. A task in an isolated worktree
can belong to a different board from the repository root. Every Curator task
or outcome handoff must name the board that actually owns the element and bind
the outcome resource to that board and element.

Use this template in the authorized coordination channel. Replace every
placeholder; keep private paths, private task IDs and private operational
evidence out of public source and PR metadata.

```text
Task: <title and element ID>
Board root: <canonical absolute project root used by the board tool>
Board access: <authorized host/tool context, or receiver-side mapping to this same board>
Outcome resource: <resource ID or supported resource locator on that board>
Resource binding: <same board root, owning element ID, outcome resource ID>
Outcome SHA-256: <digest of the exact resource bytes>
Read-back evidence: <supported tool, explicit board binding, actual result and time>
Source: <repository and full candidate commit, when applicable>
Status: <produced, reviewed, landed or blocked, with actual evidence>
Next owner and step: <accountable recipient and concrete next action>
Reply by: <absolute timestamp with timezone>
```

For a handoff before an outcome exists, mark the outcome fields as pending
and name the owner and checkpoint for producing them. Do not invent a resource
or digest. Such a handoff is an assignment, not accepted outcome evidence.

Before sending outcome evidence, use the supported native board tooling with
the stated board binding to read the element and its outcome resource. Record
the actual read result and verify the resource digest. If the receiver uses a
different host, agree the mapping to the same board rather than reusing a
sender-local path as though it existed there.

The receiver must use that explicit board binding for the supported element
and resource reads, then compare the owning element and exact resource digest.
A not-found result from the repository root does not disprove an element on a
different board. A missing binding, inaccessible board, mismatched element or
digest leaves the evidence unaccepted: preserve the failure and route the
missing binding or access to the sending orchestrator, with a named next owner
and checkpoint.

A handoff supplies a locator, not filesystem or credential permission. Do not
search ignored worktrees, private homes, histories or ledgers to reconstruct a
missing binding. Do not move boards, copy task history, change custody or lift
an existing takeover hold. Review, marker checks and source-owned landing
remain separate gates; publication of this template does not establish its
adoption or close a historical delivery failure.
