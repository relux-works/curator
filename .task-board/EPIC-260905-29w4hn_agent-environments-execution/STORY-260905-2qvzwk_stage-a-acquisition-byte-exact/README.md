# STORY-260905-2qvzwk: stage-a-acquisition-byte-exact

## Description
Implementation stage (a), item 1: replace git-archive extraction in internal/gitops with object-database extraction that reproduces exact committed blob bytes (environments.md §1.2, vector snapshot-acquisition.json at curator-spec ec695ba); switch internal/snapshot and internal/closure to it; byte-exact testdata + scratch-repo conformance test + CURATOR_CONFORMANCE_ROOT-driven test that skips on the rc.9 root. Worktree ~/Developer/ReluxWorks/.worktrees/curator-acquisition-byte-exact, branch feat/byte-exact-acquisition.

## Scope
(define story scope)

## Acceptance Criteria
Extraction never consults working-tree conversion, attributes, or filters; the spec vector hash reproduces under core.autocrlf=true and false; links/gitlinks/escapes/size bounds refused with tests; go build/vet/test green; PR reviewed and landed on curator main by fast-forward of the reviewed head.
