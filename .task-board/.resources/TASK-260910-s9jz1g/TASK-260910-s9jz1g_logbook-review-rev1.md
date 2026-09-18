# Review logbook — TASK-260910-s9jz1g revision 1

2026-09-18: Independent reviewer requests rework. Activation with a nonempty configured passphrase promotes previously staged plain PEM without encryption (`keys.py:210`). Empty configured passphrase silently writes plain PEM (`keys.py:106`). Both reproduced through real installed CLI; see review verdict for corrections and transcripts.

Validation anomaly: required pinned dced9b8 fixtures lack checkpoint_cases used by the unchanged baseline test; full independent suite yields 184 passed / 1 failed. Actual untouched CI pin is 47c3c8c; hosted gate green. This is a brief/fixture mismatch, not a new R6 regression. Coordinator should align validation instructions; do not remove existing checkpoint coverage or alter pin under R6.

Strict mypy and all 19 feature tests pass. Two narrowing mutants caught (2/2 selected; not exhaustive). Candidate left unchanged.
