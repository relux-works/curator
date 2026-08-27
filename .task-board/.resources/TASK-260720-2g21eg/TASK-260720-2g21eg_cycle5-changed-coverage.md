# TASK-260720-2g21eg — cycle-5 changed coverage

## Scope and method

This measures the cycle-5 rework delta requested by the operator, not the whole
new `go_v1.py` module that accumulated across earlier development/review
cycles.

- Reviewed cycle-4 source:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-2g21eg/reviewer-cycle4/manager-venv/lib/python3.14/site-packages/csk/builds/go_v1.py`
- Reviewed source SHA-256:
  `4d5d6b33ed7b5597531271d37372d926cc961519abc59346792893e2f0d71283`
- Cycle-5 source:
  `src/csk/builds/go_v1.py`
- Cycle-5 source SHA-256:
  `abdf349c3ff2ebad7f17c25480655e65134378b1a402c7de9e010de3a684fcd4`
- Coverage.py: 7.10.7, branch measurement enabled.
- Focused execution:
  `python -m coverage run --branch --source=csk.builds.go_v1 -m pytest -q tests/test_builds_go_v1.py`
- Focused result: exit 0, 139 passed / 1 platform skip.
- Added-line inventory: parsed from
  `git diff --no-index --unified=0` between the two exact sources. Its exit 1
  is expected and truthfully means the files differ.
- Executable changed lines are the intersection of new-side added lines with
  Coverage.py's executed/missing statement lines.
- Changed branches are Coverage.py arcs whose source line is a new-side added
  line. Combined coverage weights executable lines and branch arcs once each.

## Result

| Metric | Covered / total | Percent |
| --- | ---: | ---: |
| Changed executable lines | 271 / 291 | 93.13% |
| Changed branch arcs | 94 / 106 | 88.68% |
| Combined changed line + branch | 365 / 397 | 91.94% |

All three measurements exceed the role-added 80% target.

The first measurement before adding narrow coverage tests was:

- changed lines: 197 / 291 (67.70%)
- changed branches: 61 / 106 (57.55%)
- combined: 258 / 397 (64.99%)

That result was not marked as satisfying the checklist. Targeted tests were
then added for Windows venv/runtime mapping and archive identity, malformed
runtime configuration, macOS framework/fallback runtime layouts, bounded
descriptor capacity including a real low-soft-limit raise/restore regression,
and macOS/Windows loaded-image proof helpers. No product behavior was changed
to raise coverage.

## Reproducibility artifacts

- Coverage JSON:
  `TASK-260720-2g21eg_cycle5-coverage.json`,
  SHA-256
  `08a5fa6ea00ffb9ac71805854f51369d666c5d0fb1e232f746798c4f86f5f0bc`
- Changed-coverage calculator:
  `TASK-260720-2g21eg_measure_changed_coverage.py`,
  SHA-256
  `04766b0c9f76c463bb89f9a12f85659745b5b0198b7d16082412e1fc6a8b3500`
- Added physical lines in the cycle-5 source diff: 805.
- Remaining uncovered changed statement lines:
  `2445, 2446, 2457, 2477, 2489, 2634, 2635, 2677, 2680, 3081, 3179,
  3180, 3862, 3863, 3864, 3912, 3913, 3915, 4503, 6818`.
- Remaining uncovered changed branch arcs:
  `2456->2457, 2473->2477, 2488->2489, 2560->2555, 2676->2677,
  2678->2680, 3174->3179, 3857->3866, 3863->3864, 3863->3865,
  3866->3869, 3911->3912`.
