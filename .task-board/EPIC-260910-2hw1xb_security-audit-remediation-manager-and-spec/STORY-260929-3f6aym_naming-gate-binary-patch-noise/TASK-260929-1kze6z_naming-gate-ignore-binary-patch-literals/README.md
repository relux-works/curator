# TASK-260929-1kze6z: naming-gate-ignore-binary-patch-literals

## Description
Make the Employer name gate in .github/workflows/ci.yml ignore git binary-patch base85 literal lines (noise) while still catching real mentions anywhere, including real text inside .patch files. Move the gate into a script under .github/ci with gate-selftest coverage.

## Scope
(define task scope)

## Acceptance Criteria
1. Naming gate lives in .github/ci/naming-gate.sh, called from ci.yml; no committed file spells either name. 2. Base85/literal/blank lines inside git binary-patch blocks are exempt; every other line (incl. non-base85 lines inside a block and text lines in .patch files) is still scanned. 3. gate-selftest rows a-e (binary-noise pass, patch text fail, non-base85-in-block fail, full name fail, clean pass) with failure-reason match. 4. naming-gate.sh passes on the current tree. 5. Mutants: no-skip and no-shape-check each killed by a row.
