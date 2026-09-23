# TASK-260906-2b3nar — fold directory components in the gitops platform-path gate (curator)

Control root: /Users/administrator/Developer/ReluxWorks/curator/curator; work only in your assigned Story
worktree `.temp/STORY-260905-2qvzwk/worktree`. Read `campaign-producer-rules.md` first. The board's landing
gate is the hosted CI (runtime runs it once at handoff); run the narrow package tests yourself.

Defect (review N1 of TASK-260905-3r30t1, pre-existing, still present): `internal/gitops/gitops.go`
`planWrites` (≈ line 542) keys the case-folding collision gate on `strings.ToLower(target)` of the FULL
path only. Two tree entries whose directory components fold together but whose basenames differ —
`Dir/x.txt` + `dir/y.txt` — produce distinct keys, are admitted, and land in ONE physical directory on a
case-folding destination, so the extracted tree no longer reproduces the committed tree (the second
directory's identity is lost) and the acquisition byte-exactness claim (environments.md §1.2) is violated
without a diagnostic.

Fix: fold per path component — track folded directory prefixes (every ancestor of every planned target,
lower-cased) and, when a planned entry introduces a directory prefix that folds onto an already-planned
prefix with a different exact spelling, probe the destination once (existing `destinationFoldsCase`) and
refuse on a case-insensitive filesystem with the EXISTING `duplicate platform path in git snapshot: %q`
class (same message shape; name the entry that collides). A file and a directory that fold together
(`Dir/x.txt` + `dir` regular file) must also be refused. Case-sensitive destinations keep coexisting
exactly as today; exact duplicates keep the `duplicate path` message.

Tests (production entry = the exported extraction used by internal/snapshot / internal/closure, not the
helper alone): (a) `Dir/x.txt` + `dir/y.txt` extracted into a case-insensitive destination → refusal with
the duplicate-platform-path diagnostic and NO partial write; (b) the same tree into a case-sensitive
destination → both files present with exact bytes; (c) file-vs-directory fold refused; (d) nested prefix
fold (`A/B/x` + `a/b/y`) refused. Use the repository's existing case-folding test fixture approach (probe
the temp dir; skip with a named reason only where the platform cannot provide a case-insensitive
directory — do not silently pass). Register any new platform-case rows in the platform-cases ledger per
lane if the gate demands it. Narrowing mutants: (m1) fold only the basename, (m2) skip the prefix probe,
(m3) refuse only exact prefix duplicates — each executed and killed.

CHANGELOG (unreleased, Fixed). Attach `TASK-260906-2b3nar_results.md` (defect, fix, test table with
platform evidence, mutant table, narrow `go test ./internal/gitops ./internal/snapshot -count=1` exit code)
and hand off with `task-board handoff TASK-260906-2b3nar --role developer`.
