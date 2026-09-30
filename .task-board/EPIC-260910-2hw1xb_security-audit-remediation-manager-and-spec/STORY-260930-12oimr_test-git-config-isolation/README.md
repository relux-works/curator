# STORY-260930-12oimr: test-git-config-isolation

## Description
Test (rose-air) on runner macbook-iv: the runner copies the host ~/.gitconfig (commit/tag signing enabled) into HOME, so fixture commits expected unsigned come out signed; TestSourceSignerVectorsAtProfileInstall/unsigned-refused and TestRevisionDoesNotBorrowTagSignature fail (main bdb77413). Tests must not inherit ambient git config.

## Scope
(define story scope)

## Acceptance Criteria
(define acceptance criteria)
