# STORY-260712-3aad42: machine-config

## Description
User configuration and the enforced system configuration with locked keys (Spec 7), including env var overrides and lockable key validation.

## Acceptance Criteria
- Locked key overrides user value with a warning
- Locked-but-unset key is an error
- Registry entry validation (name, unique http(s) url, keys)
