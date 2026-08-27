## Status
done

## Review
required

## Task Class
code

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
(empty)

## Notes
cocoaskills side of the qualification, routed from the cocoaskills board (EPIC-260822-38gkry / STORY-260822-27ze8z / TASK-260824-31y75t). Qualified identity: relux-works/curator-spec 6001dc33281b94a4ec7442ab15278550dd0f51d9, manifest sha256:803918bf8672f76cf990985e51db213b826674cd5bb54fbf47731b8404b44403, tree sha256:d2c9d3df5c71d3f38cbf5eadb4ddc3e3e347a66540adfb7289fa7cbc93ed2656, protocol 1.0.0-rc.9. cocoaskills released suite pin is still 0c81c1f8d5321d822be2a2817b05aea03e656e15 and was not touched by the qualification -- advancing it is TASK-260824-1n98b3, and the tip it must bump from is now b0d7499, not 3ecca1d. Two precision points for anyone citing this evidence: the dispatch run 32756144649 resolved its candidate from the committed .github/ci/candidate-suite.json (candidate_source committed-descriptor on all six candidate jobs), not from workflow_dispatch inputs -- same identity, different mechanism than earlier notes claimed; and that run is 16 success / 5 skipped, so all jobs green is only true of the three runs combined (PR 32756139445, dispatch 32756144649, post-merge 32757585637). Consumption is proven by removal, not asserted: eight negative controls against the real candidate root, all red, covering both a family dropped from the manifest and disk and a family that stays published but stops generating cases. All six candidate jobs recorded csk-schema8-results.xml at tests=183 failures=0 errors=0 skipped=0. cocoaskills main advanced to b0d7499 (PR 45) with the same identity re-qualified on all three platforms; that PR only reattributed a declared gap and added the assertion that keeps it honest. Full evidence: cocoaskills board resources EPIC-260822-38gkry_schema8-qualification-record.md and TASK-260824-31y75t_qualification-evidence.md.

## Precondition Resources
(none)

## Outcome Resources
(none)

## Created
2026-08-24T18:07:40Z

## Last Update
2026-08-24T21:22:11Z

## Assigned To
orchestrator-inline
