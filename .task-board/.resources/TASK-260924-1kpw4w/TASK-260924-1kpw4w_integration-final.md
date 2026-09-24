run_write_boundary_uncleared: delivery of element STORY-260924-1ckno7 is gated on 2 run(s) under warn policy
  [BLOCKED] run RUN-260924-b28bdd verdict=violated terminal=violated: the terminal assessment is violated
  [BLOCKED] run RUN-260924-0e49e4 verdict=violated terminal=violated: the terminal assessment is violated
clear a violating run with: task-board spawn write-boundary-clear <RUN-ID> --reason "..."
integration_base_moved: unpublished Story prefix is not proven: a same-Story predecessor and acceptance against the current protected base are required
  head: a12ad1784eacf6e30413251cf45d2c9e339cf891
  protected_oid: 948ae7c9e4a71a4026968913e1ff646aa21e0d52
  remedy: if the local commits were already landed under rewritten identities, run task-board worktree reconcile-trunk; unique local commits must be published through a pull request
