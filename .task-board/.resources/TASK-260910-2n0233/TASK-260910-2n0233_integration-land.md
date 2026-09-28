board_publication_pending: STORY-260910-25yc0h is landed and its board state is committed as eca2bf27edaec03227ce8485953d96df9d7ef48c on the local trunk, but the publication push did not land (board_publish_local_trunk_unproven); the landing stands — run `task-board board publish` to publish it
  board_commit_oid: eca2bf27edaec03227ce8485953d96df9d7ef48c
  cause_code: board_publish_local_trunk_unproven
  post_landing_steps: ["publish the landed commits as a non-default branch and open a pull request against the protected default branch","review on the hosting platform, wait for the required checks, and merge the exact reviewed head","in the control root, after the hosted merge, prove the landed commits delivered under their rewritten identities and move local trunk (a unique local commit refuses): task-board worktree reconcile-trunk"]
  remedy: task-board board publish
  story_commit_oid: 890d598a76499b28bff41193a80dc316a669967f
  story_id: STORY-260910-25yc0h
  cause: board_publish_local_trunk_unproven: the local trunk holds an unpublished link eca2bf27edaec03227ce8485953d96df9d7ef48c that is not this repository's own board-state record (the signer cannot be bound to the repository identity: the configured signing key holds 1 whitespace-separated fields and is not exactly one SSH public key); a stable content digest is never ownership proof — nothing was projected and nothing was pushed
  link_oid: eca2bf27edaec03227ce8485953d96df9d7ef48c
  reason: the signer cannot be bound to the repository identity: the configured signing key holds 1 whitespace-separated fields and is not exactly one SSH public key
  remedy: land the unproven commits through integrate/reconcile-trunk, not through this command
