# TASK-260827-19aqkr: build-https-contract

## Description
Create docs/build-https.md: the operator HTTPS credential contract for external build repositories, mirroring docs/build-ssh.md in shape and register. Source of truth: the merged implementation in internal/ (find the https credential broker: token sources git-credentials, keyring, token_env; scope grammar and longest-prefix; install-time precheck with detected candidates; environment override variables; fail-closed off-TTY; the broker answers only the two Git prompts for the pinned host). Cross-check against the CocoaSkills sibling doc attached as a precondition, but every claim must be verified against the Curator code, not assumed from the sibling; where the implementations differ, the Curator code wins. Cross-link with docs/build-ssh.md both ways.

## Scope
docs/build-https.md (new), one cross-link line in docs/build-ssh.md.

## Acceptance Criteria
Every credential source, scope rule, env variable, prompt behavior, and error identifier verified against internal/ with grep evidence in the outcome; shape mirrors build-ssh.md; cross-links present; style guide holds.
