board_publication_pending: STORY-260923-3qwrnl is landed and its board state is committed as ab34556ebf17ab95532a3f795aa677b2234ecd8a on the local trunk, but the publication push did not land (board_publish_local_trunk_unproven); the landing stands — run `task-board board publish` to publish it
  board_commit_oid: ab34556ebf17ab95532a3f795aa677b2234ecd8a
  cause_code: board_publish_local_trunk_unproven
  post_landing_steps: ["publish the landed commits as a non-default branch and open a pull request against the protected default branch","review on the hosting platform, wait for the required checks, and merge the exact reviewed head","in the control root, after the hosted merge, prove the landed commits delivered under their rewritten identities and move local trunk (a unique local commit refuses): task-board worktree reconcile-trunk"]
  remedy: task-board board publish
  story_commit_oid: 16eec15e6a9672e081be2310e47f9e5db2d6fdb6
  story_id: STORY-260923-3qwrnl
  cause: board_publish_local_trunk_unproven: the local trunk holds an unpublished link ab34556ebf17ab95532a3f795aa677b2234ecd8a that is not this repository's own board-state record (the signer cannot be bound to the repository identity: the configured signing key holds 1 whitespace-separated fields and is not exactly one SSH public key); a stable content digest is never ownership proof — nothing was projected and nothing was pushed
  link_oid: ab34556ebf17ab95532a3f795aa677b2234ecd8a
  reason: the signer cannot be bound to the repository identity: the configured signing key holds 1 whitespace-separated fields and is not exactly one SSH public key
  remedy: land the unproven commits through integrate/reconcile-trunk, not through this command
