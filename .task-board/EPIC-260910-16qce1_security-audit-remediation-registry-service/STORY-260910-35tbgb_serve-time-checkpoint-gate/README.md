# STORY-260910-35tbgb: serve-time-checkpoint-gate

## Description
Findings R3+P2 (Medium): restore-time high-water checkpoint comparison exists only as the CLI verify-backup; serve never refuses a restored-but-older database, and the profile leaves the enforcement point ambiguous.

## Scope
curator-spec profiles/registry-service.md + curator-skill-registry cli.py/app.py

## Acceptance Criteria
Profile names the enforcement point; serve supports --checkpoint and fails closed when live state is below or inconsistent with it
