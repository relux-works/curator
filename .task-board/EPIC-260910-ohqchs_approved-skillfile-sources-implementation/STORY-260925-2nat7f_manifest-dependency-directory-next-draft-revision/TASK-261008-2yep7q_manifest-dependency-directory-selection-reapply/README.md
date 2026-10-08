# Re-apply the accepted manifest dependency-directory amendment (TASK-260924-2am4qa rev2) onto curator-spec main

## Description
TASK-260924-2am4qa revision 2 was accepted on 2026-09-24 and never landed; the board holds it integrating without a Change Request record, so it cannot be republished in place. This successor re-applies the full revision-2 content (attached patch) onto today's curator-spec main (rc.14), passes validation, and lands; the old task is then closed as superseded by this one.

## Scope
curator-spec: protocol/core.md §4.4, skillfile-sources and manager cross-references, draft-sources-v1 schemas and conformance cases, tools/validate.py recipe lines, CHANGELOG

## Acceptance Criteria
Revision-2 intent re-applied on spec main; validate.py and test_validate green; conformance cases present and indexed; reviewer accepted; landed on spec main.
