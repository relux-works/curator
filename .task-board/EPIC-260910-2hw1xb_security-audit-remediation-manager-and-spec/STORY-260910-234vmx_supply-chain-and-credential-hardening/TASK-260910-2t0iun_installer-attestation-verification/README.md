# TASK-260910-2t0iun: installer-attestation-verification

## Description
curator: install.sh verifies the GitHub artifact attestation (gh attestation verify) or an independent minisign/cosign signature of checksums.txt before installing.

## Scope
(define task scope)

## Acceptance Criteria
Installer fails on tampered or unattested artifacts; tested manually and documented in README
