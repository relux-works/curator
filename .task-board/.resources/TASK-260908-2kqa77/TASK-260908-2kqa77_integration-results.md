run_write_boundary_uncleared: delivery of element STORY-260908-g7o5zw is gated on 2 run(s) under warn policy
  [BLOCKED] run RUN-260923-cbc7df verdict=violated terminal=violated: the terminal assessment is violated
  [ok] run RUN-260923-30160c verdict=violated terminal=violated: assessed
clear a violating run with: task-board spawn write-boundary-clear <RUN-ID> --reason "..."
board_publication_pending: STORY-260908-g7o5zw is landed and its board state is committed as 1511b345c143acfd78b5db0ab4f3176f5ce6ce94 on the local trunk, but the publication push did not land (board_publish_local_trunk_unproven); the landing stands — run `task-board board publish` to publish it
  board_commit_oid: 1511b345c143acfd78b5db0ab4f3176f5ce6ce94
  cause_code: board_publish_local_trunk_unproven
  post_landing_steps: ["publish the landed commits as a non-default branch and open a pull request against the protected default branch","review on the hosting platform, wait for the required checks, and merge the exact reviewed head","in the control root, after the hosted merge, prove the landed commits delivered under their rewritten identities and move local trunk (a unique local commit refuses): task-board worktree reconcile-trunk"]
  remedy: task-board board publish
  story_commit_oid: b1e296ef1b7e09da61683e4a5e3e691cf9b1449e
  story_id: STORY-260908-g7o5zw
  cause: board_publish_local_trunk_unproven: the local trunk holds an unpublished link 1511b345c143acfd78b5db0ab4f3176f5ce6ce94 that is not this repository's own board-state record (the signer cannot be bound to the repository identity: the configured signing key holds 1 whitespace-separated fields and is not exactly one SSH public key); a stable content digest is never ownership proof — nothing was projected and nothing was pushed
  link_oid: 1511b345c143acfd78b5db0ab4f3176f5ce6ce94
  reason: the signer cannot be bound to the repository identity: the configured signing key holds 1 whitespace-separated fields and is not exactly one SSH public key
  remedy: land the unproven commits through integrate/reconcile-trunk, not through this command
