board_publication_pending: STORY-260924-txgta4 is landed and its board state is committed as 58669b1d27721d13eb752dc987fe4528ca685ea1 on the local trunk, but the publication push did not land (board_publish_local_trunk_unproven); the landing stands — run `task-board board publish` to publish it
  board_commit_oid: 58669b1d27721d13eb752dc987fe4528ca685ea1
  cause_code: board_publish_local_trunk_unproven
  post_landing_steps: ["publish the landed commits as a non-default branch and open a pull request against the protected default branch","review on the hosting platform, wait for the required checks, and merge the exact reviewed head","in the control root, after the hosted merge, prove the landed commits delivered under their rewritten identities and move local trunk (a unique local commit refuses): task-board worktree reconcile-trunk"]
  remedy: task-board board publish
  story_commit_oid: f03da5bc45a3746eaf967bfa231181c978c39ec2
  story_id: STORY-260924-txgta4
  cause: board_publish_local_trunk_unproven: the local trunk holds an unpublished link 58669b1d27721d13eb752dc987fe4528ca685ea1 that is not this repository's own board-state record (the signer cannot be bound to the repository identity: the configured signing key holds 1 whitespace-separated fields and is not exactly one SSH public key); a stable content digest is never ownership proof — nothing was projected and nothing was pushed
  link_oid: 58669b1d27721d13eb752dc987fe4528ca685ea1
  reason: the signer cannot be bound to the repository identity: the configured signing key holds 1 whitespace-separated fields and is not exactly one SSH public key
  remedy: land the unproven commits through integrate/reconcile-trunk, not through this command
