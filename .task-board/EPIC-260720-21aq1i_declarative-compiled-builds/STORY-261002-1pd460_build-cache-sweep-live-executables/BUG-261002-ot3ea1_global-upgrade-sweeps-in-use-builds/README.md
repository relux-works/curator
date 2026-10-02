# BUG-261002-ot3ea1: global-upgrade-sweeps-in-use-builds

## Description
Build-cache sweep after curator global upgrade must retain any cache build whose binary is the executable of a live process (or keep the previous build per command for a grace period).

## Scope
(define bug scope / affected area)

## Acceptance Criteria
Sweep retains builds executed by live processes (portable enumeration, fail-safe on errors); tests for in-use/unused/error/daemon; mutant; go test + windows vet evidence.
