# TASK-260720-3pwg2w implementation and rework evidence

## Provenance

- Worktree: /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-3pwg2w/worktree
- Exact base: origin/main 17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8.
- Imported predecessor product state remains byte-identical to the accepted TASK-260720-3mrm4z worktree outside internal/buildcache.
- Candidate conformance input: /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-3ag6pi/worktree/conformance/v1. Manifest SHA-256: 70cd274150d629f16ca04f2073a8eae5a0f3fff4584f905d807775040af21eae. It was used only as vector input.
- No files were committed or staged.

## Implementation

internal/buildcache implements immutable manager-protected go-v1 entries at cache/build/go-v1/<64-hex-key>, canonical Curator receipts, bin artifacts, exact key/input/target/toolchain/path/hash/size/canonical-byte validation, explicit hit/miss/corrupt/untrusted-provenance/unsupported outcomes, caller-held-lock publication and quarantine, strictly read-only inspection, private staging, atomic winner publication, identical-loser reuse, conflicting-byte errors, POSIX no-follow ownership/mode/link/executable checks, Windows owner/DACL/reparse/link checks, and fail-closed unsupported-platform behavior. The forged self-consistent cache vector remains untrusted and is rebuilt rather than adopted.

## Reviewer-directed Windows DACL rework

The reviewed validator previously skipped deny ACEs and credited inherit-only owner allows. The rework now accepts only the unambiguous DACL shape Curator creates: valid effective owner, protected DACL, direct owner-only ACCESS_ALLOWED entries with zero inheritance flags, no deny or unsupported entries, no other-principal entries, and owner rights containing GENERIC_ALL or the full concrete mutation-right set. This is a strict canonical-policy check, so deny-order and inherit-only effective-right ambiguity fail closed.

Windows regression coverage now includes integrated Inspect rejection for owner deny, applicable World-group deny, inherit-only owner allow, insufficient owner mutation rights, and another-principal grants. Direct policy tests also cover wrong owner and the accepted canonical owner-only DACL, alongside existing reparse, special-file, and hard-link cases.

## Verification

- make check: PASS, including go vet ./..., go test ./..., and repository gofmt gate.
- go build ./...: PASS.
- go test -race ./...: PASS.
- Candidate-focused tests for buildcache, buildmeta, buildsource, protocoljson, and registry: PASS.
- Race stress count 20 for atomic publication and forged-cache rejection: PASS.
- Native internal/buildcache coverage: 81.6 percent.
- Linux amd64 full compile/link: PASS.
- Windows amd64 full compile/link: PASS.
- Windows buildcache test binaries compile for amd64 and arm64: PASS.
- Windows amd64 go vet for internal/buildcache: PASS.
- Focused buildcache compile matrix for FreeBSD, NetBSD, OpenBSD, AIX, DragonFly BSD, illumos, Solaris, and unsupported Plan 9: PASS.
- git diff --check and gofmt: PASS.
- Predecessor product provenance comparison: PASS.

Windows runtime tests were not executed because the available Darwin host has neither a Windows runner nor Wine. Windows production and test sources compile and link successfully; Darwin executes the POSIX protection suite. This is the only unavailable runtime gate.