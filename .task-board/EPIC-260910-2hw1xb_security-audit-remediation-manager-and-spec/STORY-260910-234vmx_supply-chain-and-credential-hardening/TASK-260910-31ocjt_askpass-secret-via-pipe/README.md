# TASK-260910-31ocjt: askpass-secret-via-pipe

## Description
curator: deliver the HTTPS askpass secret to the fetch child through a pipe or inherited fd instead of CURATOR_BUILD_HTTPS_ASKPASS_SECRET in the environment, so grandchildren cannot inherit it.

## Scope
(define task scope)

## Acceptance Criteria
Broker test proves the secret is absent from the fetch child environment
