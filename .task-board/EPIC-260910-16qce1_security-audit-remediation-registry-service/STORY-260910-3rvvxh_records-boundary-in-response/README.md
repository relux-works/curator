# STORY-260910-3rvvxh: records-boundary-in-response

## Description
Finding R1 (High, cross-repo): the server evaluates record pages at a committed boundary but never tells the client which one, enabling a fresh-snapshot/stale-pages withholding attack against clients that do not replay the log.

## Scope
curator-skill-registry app.py/store.py + spec task TASK-260910-1b1ens

## Acceptance Criteria
Records and log responses carry the boundary they were evaluated at in the shape the spec revision defines; conformance vector added
