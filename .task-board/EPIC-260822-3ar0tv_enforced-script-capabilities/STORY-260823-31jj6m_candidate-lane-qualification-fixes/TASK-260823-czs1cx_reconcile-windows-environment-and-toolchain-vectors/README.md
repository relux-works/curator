# TASK-260823-czs1cx: reconcile-windows-environment-and-toolchain-vectors

## Description
Windows candidate lane, two failures that may be CANDIDATE VECTOR bugs rather than implementation bugs — judge each on the merits: (1) internal/godriver TestFixedEnvironmentAndFiveDirectArgvFormsVector/fixed_environment observed GOARCH=amd64 while the candidate expects a closed arm64 value — if the schema-8 vector generator hardcoded the authoring host GOARCH, the vector must become host-conditional or parameterized, fixed ON the candidate branch candidate/schema-8-rc.9, producing a NEW immutable candidate commit and manifest digest (never rewrite the old one; record the new identity on TASK-260822-c0rxj7); (2) TestToolchainIdentityVectors/unsorted-directories-files-and-internal-link digest mismatch — determine whether the Windows implementation hashes wrongly or the vector encodes platform-dependent bytes; fix the guilty side. Implementation fixes land on main via PR; vector fixes land on the candidate branch with double regeneration. Fully autonomous per the 2026-08-22 pre-authorization.

## Scope
(define task scope)

## Acceptance Criteria
Both cases green on windows against the (possibly superseded) candidate; any new candidate identity recorded on TASK-260822-c0rxj7 with regeneration proof.
