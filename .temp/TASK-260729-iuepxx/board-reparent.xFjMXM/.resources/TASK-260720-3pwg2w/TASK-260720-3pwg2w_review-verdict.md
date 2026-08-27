# Review verdict: changes requested

## Blocking finding

Windows protected-DACL validation is not an effective-rights check. In internal/buildcache/protection_windows.go:159-183, every ACCESS_DENIED_ACE_TYPE is skipped and an owner ACCESS_ALLOWED_ACE_TYPE is counted even when ACE flags make it inherit-only. Windows evaluates applicable allow and deny ACEs in sequence. Therefore a protected DACL with an owner or owner-group deny for a required mutation right followed by owner GENERIC_ALL, or a DACL whose apparent owner allow is INHERIT_ONLY, can pass validateWindowsSecurity even though the owner lacks the required effective mutation rights. This violates the AC that Windows verify owner and DACL mutation rights and permits reuse of state that should fail closed. Microsoft references: https://learn.microsoft.com/en-us/windows/win32/secauthz/dacls-and-aces and https://learn.microsoft.com/en-us/windows/win32/secauthz/access-control-lists.

Existing Windows tests cover another-principal mutation grants, hard links, directory-as-special-file, and reparse points, but not owner/group deny ACEs, inherit-only owner allows, missing effective owner mutation rights, or owner mismatch.

## Required rework

Compute the owner token effective mutation rights with Windows access-check semantics, or enforce a stricter canonical owner-only protected DACL that cannot contain applicable deny/inherit-only ambiguity. Continue rejecting mutation grants to any other principal. Add Windows tests proving fail-closed behavior for an owner or applicable-group deny, an inherit-only owner allow, missing required mutation rights, and wrong owner; run them on Windows when available and retain the compile/link gate elsewhere.

## Passing evidence

Task-only delta versus accepted TASK-260720-3mrm4z is exactly internal/buildcache. POSIX no-follow ownership/mode/link/executable review is sound. Independent commands passed: make check; go test -race ./...; candidate-focused buildcache/buildmeta/buildsource/protocoljson/registry tests; Windows and Linux full compile/link; Plan 9 unsupported compile; git diff --check; gofmt; 20x race stress for publication and forged-cache tests. Native buildcache coverage is 81.6%. Windows runtime tests were not executed on the Darwin host. No human decision is required; this is focused implementation and test rework.