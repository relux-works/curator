run_write_boundary_uncleared: delivery of element TASK-260922-3bbvrs is gated on 2 run(s) under warn policy
  [BLOCKED] run RUN-260923-7c2fcf verdict=violated terminal=violated: the terminal assessment is violated
  [ok] run RUN-260924-555a29 verdict=violated terminal=violated: assessed
clear a violating run with: task-board spawn write-boundary-clear <RUN-ID> --reason "..."
TASK-260922-3bbvrs: checkpointed as 3761705d4a228081ecc38d95c8d0a3f8071b96e6 on task-board/story/STORY-260922-2goxjs
TASK-260922-3bbvrs: status integrating
