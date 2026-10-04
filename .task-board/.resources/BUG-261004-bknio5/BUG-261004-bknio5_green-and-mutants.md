# BUG-261004-bknio5 — green and mutation evidence

Temporary fixture prefixes are redacted. This records tests executed by the developer on the current candidate; no previously attached test result was accepted in lieu of these local commands.

## Restored green regression

Command: `env -u GOROOT GOFLAGS=-work GOMAXPROCS=2 go test -p 2 ./cmd/curator ./internal/scopes -run '^(TestGCPreservesRuntimeOnUncertainMarks|TestCollectStaysFailSafe)' -count=1 -v`

Observed exit: **0** (passing).

```text
=== RUN   TestGCPreservesRuntimeOnUncertainMarks
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/truncated_registry
    gc_test.go:92: pass 1: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: consumer registry [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry1747710911/001/manager home/consumers.json cannot be trusted: manager_state_unreadable: [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry1747710911/001/manager home/consumers.json: EOF; no consumer was pruned and no build entry was swept\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: consumer registry [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry1747710911/001/manager home/consumers.json cannot be trusted: manager_state_unreadable: [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry1747710911/001/manager home/consumers.json: EOF; no consumer was pruned and no build entry was swept\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/invalid_marker
    gc_test.go:92: pass 1: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker821457416/002/.agents/skills/skill-a/.csk-install.json is invalid (install_marker_invalid): install_marker_invalid: [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker821457416/002/.agents/skills/skill-a/.csk-install.json: malformed protocol JSON object; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker821457416/002/.agents/skills/skill-a/.csk-install.json is invalid (install_marker_invalid): install_marker_invalid: [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker821457416/002/.agents/skills/skill-a/.csk-install.json: malformed protocol JSON object; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/unreadable_marker
    gc_test.go:92: pass 1: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker4193351718/002/.agents/skills/skill-a/.csk-install.json cannot be inspected (manager_state_unreadable): manager_state_unreadable: [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker4193351718/002/.agents/skills/skill-a/.csk-install.json: read [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker4193351718/002/.agents/skills/skill-a/.csk-install.json: is a directory; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker4193351718/002/.agents/skills/skill-a/.csk-install.json cannot be inspected (manager_state_unreadable): manager_state_unreadable: [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker4193351718/002/.agents/skills/skill-a/.csk-install.json: read [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker4193351718/002/.agents/skills/skill-a/.csk-install.json: is a directory; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/complete_references
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\ngc: 1 runtime entry removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 25187: invalid argument; all builds retained\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 25218: invalid argument; all builds retained\n"
--- PASS: TestGCPreservesRuntimeOnUncertainMarks (5.09s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/truncated_registry (2.14s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/invalid_marker (1.28s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/unreadable_marker (0.94s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/complete_references (0.73s)
PASS
ok  	github.com/relux-works/curator/cmd/curator	6.462s
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses/registry_of_an_unknown_schema
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses/registry_repeating_a_known_member
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses/project_with_only_an_invalid_marker
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses/unreadable_installed_skill_directory
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses/installed_skill_replaced_by_a_file
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses/skill_root_replaced_by_a_file
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses/corrupt_consumer_registry
--- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses (0.03s)
    --- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses/registry_of_an_unknown_schema (0.00s)
    --- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses/registry_repeating_a_known_member (0.00s)
    --- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses/project_with_only_an_invalid_marker (0.01s)
    --- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses/unreadable_installed_skill_directory (0.00s)
    --- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses/installed_skill_replaced_by_a_file (0.01s)
    --- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses/skill_root_replaced_by_a_file (0.00s)
    --- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses/corrupt_consumer_registry (0.00s)
=== RUN   TestCollectStaysFailSafeOnRedirectedUnixScopes
=== RUN   TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_project_skill_root
=== RUN   TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_installed_skill
=== RUN   TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_global_scope_root
=== RUN   TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_hybrid_scope_root
--- PASS: TestCollectStaysFailSafeOnRedirectedUnixScopes (0.23s)
    --- PASS: TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_project_skill_root (0.01s)
    --- PASS: TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_installed_skill (0.01s)
    --- PASS: TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_global_scope_root (0.22s)
    --- PASS: TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_hybrid_scope_root (0.00s)
PASS
ok  	github.com/relux-works/curator/internal/scopes	2.873s

```

## Mutant 1: check uncertainty after sweeping runtime

Command: `env -u GOROOT GOFLAGS=-work GOMAXPROCS=2 go test -p 2 ./cmd/curator -run '^TestGCPreservesRuntimeOnUncertainMarks$' -count=1`

Observed exit: **1** (expected failing mutant; caught by broken shim and runtime deletion).

```text
--- FAIL: TestGCPreservesRuntimeOnUncertainMarks (1.48s)
    --- FAIL: TestGCPreservesRuntimeOnUncertainMarks/truncated_registry (0.38s)
        gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\nremoved runtime skill-a/1111111111111111111111111111111111111111\ngc: 2 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: consumer registry [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry4275809945/001/manager home/consumers.json cannot be trusted: manager_state_unreadable: [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry4275809945/001/manager home/consumers.json: EOF; no consumer was pruned and no build entry was swept\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
        gc_test.go:96: shim = "", fork/exec [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry4275809945/002/.agents/bin/tool: no such file or directory; want runtime-ok
        gc_test.go:104: pass 1: uncertain marks removed an unreferenced runtime too: stat [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry4275809945/001/manager home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
        gc_test.go:89: shim = "", fork/exec [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry4275809945/002/.agents/bin/tool: no such file or directory; want runtime-ok
        gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: consumer registry [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry4275809945/001/manager home/consumers.json cannot be trusted: manager_state_unreadable: [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry4275809945/001/manager home/consumers.json: EOF; no consumer was pruned and no build entry was swept\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
        gc_test.go:96: shim = "", fork/exec [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry4275809945/002/.agents/bin/tool: no such file or directory; want runtime-ok
        gc_test.go:104: pass 2: uncertain marks removed an unreferenced runtime too: stat [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry4275809945/001/manager home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
    --- FAIL: TestGCPreservesRuntimeOnUncertainMarks/invalid_marker (0.35s)
        gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\nremoved runtime skill-a/1111111111111111111111111111111111111111\ngc: 2 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker242940096/002/.agents/skills/skill-a/.csk-install.json is invalid (install_marker_invalid): install_marker_invalid: [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker242940096/002/.agents/skills/skill-a/.csk-install.json: malformed protocol JSON object; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
        gc_test.go:96: shim = "", fork/exec [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker242940096/002/.agents/bin/tool: no such file or directory; want runtime-ok
        gc_test.go:104: pass 1: uncertain marks removed an unreferenced runtime too: stat [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker242940096/001/manager home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
        gc_test.go:89: shim = "", fork/exec [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker242940096/002/.agents/bin/tool: no such file or directory; want runtime-ok
        gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker242940096/002/.agents/skills/skill-a/.csk-install.json is invalid (install_marker_invalid): install_marker_invalid: [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker242940096/002/.agents/skills/skill-a/.csk-install.json: malformed protocol JSON object; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
        gc_test.go:96: shim = "", fork/exec [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker242940096/002/.agents/bin/tool: no such file or directory; want runtime-ok
        gc_test.go:104: pass 2: uncertain marks removed an unreferenced runtime too: stat [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker242940096/001/manager home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
    --- FAIL: TestGCPreservesRuntimeOnUncertainMarks/unreadable_marker (0.35s)
        gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\nremoved runtime skill-a/1111111111111111111111111111111111111111\ngc: 2 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2002261790/002/.agents/skills/skill-a/.csk-install.json cannot be inspected (manager_state_unreadable): manager_state_unreadable: [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2002261790/002/.agents/skills/skill-a/.csk-install.json: read [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2002261790/002/.agents/skills/skill-a/.csk-install.json: is a directory; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
        gc_test.go:96: shim = "", fork/exec [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2002261790/002/.agents/bin/tool: no such file or directory; want runtime-ok
        gc_test.go:104: pass 1: uncertain marks removed an unreferenced runtime too: stat [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2002261790/001/manager home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
        gc_test.go:89: shim = "", fork/exec [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2002261790/002/.agents/bin/tool: no such file or directory; want runtime-ok
        gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2002261790/002/.agents/skills/skill-a/.csk-install.json cannot be inspected (manager_state_unreadable): manager_state_unreadable: [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2002261790/002/.agents/skills/skill-a/.csk-install.json: read [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2002261790/002/.agents/skills/skill-a/.csk-install.json: is a directory; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
        gc_test.go:96: shim = "", fork/exec [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2002261790/002/.agents/bin/tool: no such file or directory; want runtime-ok
        gc_test.go:104: pass 2: uncertain marks removed an unreferenced runtime too: stat [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2002261790/001/manager home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	2.020s
FAIL

```

## Mutant 2: omit only consumer-registry read uncertainty

Command: `env -u GOROOT GOFLAGS=-work GOMAXPROCS=2 go test -p 2 ./cmd/curator -run '^TestGCPreservesRuntimeOnUncertainMarks/truncated_registry$' -count=1`

Observed exit: **1** (expected failing mutant; caught by broken shim and runtime deletion).

```text
--- FAIL: TestGCPreservesRuntimeOnUncertainMarks (0.44s)
    --- FAIL: TestGCPreservesRuntimeOnUncertainMarks/truncated_registry (0.44s)
        gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\nremoved runtime skill-a/1111111111111111111111111111111111111111\ngc: 2 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 89184: invalid argument; all builds retained\n"
        gc_test.go:96: shim = "", fork/exec [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry659616833/002/.agents/bin/tool: no such file or directory; want runtime-ok
        gc_test.go:100: pass 1: missing warning "cannot be trusted"
        gc_test.go:100: pass 1: missing warning "runtime sweep skipped"
        gc_test.go:100: pass 1: missing warning "live reference set could not be proven complete"
        gc_test.go:104: pass 1: uncertain marks removed an unreferenced runtime too: stat [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry659616833/001/manager home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
        gc_test.go:89: shim = "", fork/exec [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry659616833/002/.agents/bin/tool: no such file or directory; want runtime-ok
        gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 89184: invalid argument; all builds retained\n"
        gc_test.go:96: shim = "", fork/exec [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry659616833/002/.agents/bin/tool: no such file or directory; want runtime-ok
        gc_test.go:100: pass 2: missing warning "cannot be trusted"
        gc_test.go:100: pass 2: missing warning "runtime sweep skipped"
        gc_test.go:100: pass 2: missing warning "live reference set could not be proven complete"
        gc_test.go:104: pass 2: uncertain marks removed an unreferenced runtime too: stat [temporary fixture]/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry659616833/001/manager home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	1.053s
FAIL

```

Both mutants were restored from a saved green copy, with byte equality asserted. The restored tracked worktree matches hosted snapshot `b78a1b2a358babffd8499416ceca5a24a844aaad` (`git diff --exit-code` returned 0).

