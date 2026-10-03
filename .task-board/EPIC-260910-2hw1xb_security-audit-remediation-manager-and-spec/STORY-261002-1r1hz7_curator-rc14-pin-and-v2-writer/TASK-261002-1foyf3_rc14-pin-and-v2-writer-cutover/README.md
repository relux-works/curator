# TASK-261002-1foyf3: rc14-pin-and-v2-writer-cutover

## Description
Promote SPEC_PIN to rc.14 and enable v2 writers; clear dependent rc.14 gap rows.

## Scope
(define task scope)

## Acceptance Criteria
SPEC_PIN = final rc.14 peeled commit (digest 6f832d81 verified); EnableV2Writers stays false; rc.14 byte-exact-snapshot gap row re-owned to TASK-261003-1uzji7 with exact counts; hosted gate green at the rc.14 default pin.
