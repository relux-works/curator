# BUG-260922-k6eypp story integrate (run 2)

Command: `task-board worktree integrate STORY-260923-11vn9k --cr BUG-260922-k6eypp --revision 3` (control root, SSH_AUTH_SOCK exported). Exit code: 1.

Summary: Story landed on local trunk (story_commit 1308f025, board commit 2252ebee); board publication push refused (board_publish_local_trunk_unproven: configured signing key is not exactly one SSH public key). Hosted PR/merge + reconcile-trunk remain for the orchestrator.

```
board_publication_pending: STORY-260923-11vn9k is landed and its board state is committed as 2252ebee05b5d93c982b288aa6de57dcc4dea6a9 on the local trunk, but the publication push did not land (board_publish_local_trunk_unproven); the landing stands — run `task-board board publish` to publish it
  board_commit_oid: 2252ebee05b5d93c982b288aa6de57dcc4dea6a9
  cause_code: board_publish_local_trunk_unproven
  post_landing_steps: ["publish the landed commits as a non-default branch and open a pull request against the protected default branch","review on the hosting platform, wait for the required checks, and merge the exact reviewed head","in the control root, after the hosted merge, prove the landed commits delivered under their rewritten identities and move local trunk (a unique local commit refuses): task-board worktree reconcile-trunk"]
  remedy: task-board board publish
  story_commit_oid: 1308f02575308801f97db54fc5f1d55cbd16cc99
  story_id: STORY-260923-11vn9k
  cause: board_publish_local_trunk_unproven: the local trunk holds an unpublished link 2252ebee05b5d93c982b288aa6de57dcc4dea6a9 that is not this repository's own board-state record (the signer cannot be bound to the repository identity: the configured signing key holds 1 whitespace-separated fields and is not exactly one SSH public key); a stable content digest is never ownership proof — nothing was projected and nothing was pushed
  link_oid: 2252ebee05b5d93c982b288aa6de57dcc4dea6a9
  reason: the signer cannot be bound to the repository identity: the configured signing key holds 1 whitespace-separated fields and is not exactly one SSH public key
  remedy: land the unproven commits through integrate/reconcile-trunk, not through this command
exit=1
```
