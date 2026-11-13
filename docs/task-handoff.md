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
Board directory: <canonical absolute board directory from local.board_dir, for --board-dir>
Board access: <authorized host/tool context, or receiver-side mapping to this same board>
Outcome resource: <resource name or supported resource locator on that board>
Resource binding: <same board directory, owning element ID, outcome resource name>
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

For a local board, `Board directory` names the configured `local.board_dir`,
resolved against the owning project root, rather than the project root itself.
Set `BOARD_DIR` to that directory, `TASK_ID` to the owning element ID and
`RESOURCE_NAME` to the outcome resource name. Read both with the explicit binding:

```sh
task-board --board-dir "$BOARD_DIR" q "get($TASK_ID)"
task-board --board-dir "$BOARD_DIR" resource get "$TASK_ID" "$RESOURCE_NAME" -o -
```

Both reads must resolve on the named board; compare the returned resource bytes
with the supplied digest.

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
