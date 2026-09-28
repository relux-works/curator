# Delta review — TASK-260928-3ed9m3 rev2 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Rev1 ACCEPTED (your verdict). Rev2 re-applies it on trunk d8e87bac (E5 managed-write nofollow 19shmj, E6 path-source boundary yvxbs1 landed):
base = trunk, tree 9f6f20f3, gate green, same 7 paths. Orchestrator line check vs (trunk ∪ rev1): 6 paths clean. Review ONLY
internal/envprofile/switch.go: 70 lines in neither side and 17 TRUNK lines dropped, e.g. `linkTarget, readErr := os.Readlink(full)` and
a comment `// in-place homes only. Root context and profile skill directories are …`.
1. Show `git diff d8e87bac 9f6f20f3 -- internal/envprofile/switch.go` and decide, hunk by hunk, whether each is (a) 3ed9m3's lock-first
   ordering, (b) a legitimate merge of both sides, or (c) a REGRESSION of trunk behaviour — in particular E5's §8.3.1 nofollow discipline
   (never write/traverse through a link; replace entries via the managed helpers; the Readlink-based detection of manager-owned vs planted
   links) and E6's path-source preflight. Any (c) is a blocking finding with file:line.
2. Run `go test ./internal/envprofile -run 'Nofollow|Link|Switch|Global|Takeover|Path|Guarded'` and `go test ./cmd/curator -run
   'Global|Takeover|Profile'` with real exit codes, and re-run one E5 mutant (Lstat→Stat in the parent walk) to prove the nofollow rows
   still bite on the candidate.
accept_cr or changes requested with file:line. No LOGBOOK.md.
