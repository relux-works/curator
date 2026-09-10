# STORY-260910-9484i4: registry-key-management

## Description
Finding R6 (Low/Operational): the signing key is plain unencrypted PKCS8 on the same volume as the database; no passphrase or KMS path exists.

## Scope
curator-skill-registry keys.py/cli.py

## Acceptance Criteria
Optional passphrase-protected key loading (env-provided secret) and a documented KMS hook point; tests cover the encrypted-key path
