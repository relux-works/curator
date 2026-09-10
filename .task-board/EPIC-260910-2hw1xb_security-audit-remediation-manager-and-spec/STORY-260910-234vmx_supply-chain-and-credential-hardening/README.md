# STORY-260910-234vmx: supply-chain-and-credential-hardening

## Description
Findings I2+I3+S7 (Low/Medium): install.sh verifies only same-origin checksums; the HTTPS askpass secret travels through the child environment; the no-kernel-sandbox posture needs a top-level statement.

## Scope
install.sh + internal/buildrepo/httpsbroker + README/SECURITY

## Acceptance Criteria
Installer verifies release attestation or an independent signature; askpass secret is delivered via pipe/fd; README states plainly that installed commands run with user privileges unless verified assurance is configured
