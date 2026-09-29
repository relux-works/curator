run_write_boundary_uncleared: delivery of element STORY-260928-16hi30 is gated on 2 run(s) under warn policy
  [BLOCKED] run RUN-260927-8aa45c verdict=indeterminate terminal=indeterminate: the terminal assessment is indeterminate
  [ok] run RUN-260928-133947 verdict=violated terminal=violated: assessed
clear a violating run with: task-board spawn write-boundary-clear <RUN-ID> --reason "..."
