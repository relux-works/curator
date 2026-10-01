# STORY-261001-17ali8: askpass-pipe-refusal-epipe

## Description
HTTPS broker writes the secret into the askpass pipe even when the askpass child refuses (foreign user, bare prompt, wrong argv) and exits without reading; on macOS -race the write fails with EPIPE (write |1: broken pipe) and the transport reports an error.

## Scope
(define story scope)

## Acceptance Criteria
(define acceptance criteria)
