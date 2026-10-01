# BUG-261001-2772iz: askpass-pipe-refusal-epipe

## Description
TestDraftSourcesBrokerAskpassDispatch refusal subtests fail on Race (macos-latest) with broken pipe.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
Deterministic reproduction of askpass refusal-path EPIPE fails on main and passes after fix; refusal outcome unchanged, secret never leaked, real transport errors still reported; -race count=50 green on darwin; mutant kills.
