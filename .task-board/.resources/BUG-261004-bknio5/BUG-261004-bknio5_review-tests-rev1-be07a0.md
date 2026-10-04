# BUG-261004-bknio5 — independent review transcripts

Temporary filesystem paths redacted. All Go commands used GOFLAGS=-work.

## base (exit 1)

```text
WORK=<temporary-path>
=== RUN   TestGCPreservesRuntimeOnUncertainMarks
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/truncated_registry
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\nremoved runtime skill-a/1111111111111111111111111111111111111111\ngc: 2 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: consumer registry <temporary-path> home/consumers.json cannot be trusted: manager_state_unreadable: <temporary-path> home/consumers.json: EOF; no consumer was pruned and no build entry was swept\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:96: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:100: pass 1: missing warning "runtime sweep skipped"
    gc_test.go:104: pass 1: uncertain marks removed an unreferenced runtime too: stat <temporary-path> home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
    gc_test.go:89: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: consumer registry <temporary-path> home/consumers.json cannot be trusted: manager_state_unreadable: <temporary-path> home/consumers.json: EOF; no consumer was pruned and no build entry was swept\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:96: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:100: pass 2: missing warning "runtime sweep skipped"
    gc_test.go:104: pass 2: uncertain marks removed an unreferenced runtime too: stat <temporary-path> home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/invalid_marker
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\nremoved runtime skill-a/1111111111111111111111111111111111111111\ngc: 2 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary-path> is invalid (install_marker_invalid): install_marker_invalid: <temporary-path> malformed protocol JSON object; repair or remove that installation\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:96: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:100: pass 1: missing warning "runtime sweep skipped"
    gc_test.go:104: pass 1: uncertain marks removed an unreferenced runtime too: stat <temporary-path> home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
    gc_test.go:89: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary-path> is invalid (install_marker_invalid): install_marker_invalid: <temporary-path> malformed protocol JSON object; repair or remove that installation\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:96: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:100: pass 2: missing warning "runtime sweep skipped"
    gc_test.go:104: pass 2: uncertain marks removed an unreferenced runtime too: stat <temporary-path> home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/unreadable_marker
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\nremoved runtime skill-a/1111111111111111111111111111111111111111\ngc: 2 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary-path> cannot be inspected (manager_state_unreadable): manager_state_unreadable: <temporary-path> read <temporary-path> is a directory; repair or remove that installation\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:96: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:100: pass 1: missing warning "runtime sweep skipped"
    gc_test.go:104: pass 1: uncertain marks removed an unreferenced runtime too: stat <temporary-path> home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
    gc_test.go:89: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary-path> cannot be inspected (manager_state_unreadable): manager_state_unreadable: <temporary-path> read <temporary-path> is a directory; repair or remove that installation\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:96: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:100: pass 2: missing warning "runtime sweep skipped"
    gc_test.go:104: pass 2: uncertain marks removed an unreferenced runtime too: stat <temporary-path> home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/complete_references
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\ngc: 1 runtime entry removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 54559: invalid argument; all builds retained\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 54559: invalid argument; all builds retained\n"
--- FAIL: TestGCPreservesRuntimeOnUncertainMarks (2.11s)
    --- FAIL: TestGCPreservesRuntimeOnUncertainMarks/truncated_registry (0.50s)
    --- FAIL: TestGCPreservesRuntimeOnUncertainMarks/invalid_marker (0.43s)
    --- FAIL: TestGCPreservesRuntimeOnUncertainMarks/unreadable_marker (0.64s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/complete_references (0.55s)
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	2.666s
FAIL
```

## green (exit 0)

```text
WORK=<temporary-path>
=== RUN   TestGCPreservesRuntimeOnUncertainMarks
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/truncated_registry
    gc_test.go:92: pass 1: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: consumer registry <temporary-path> home/consumers.json cannot be trusted: manager_state_unreadable: <temporary-path> home/consumers.json: EOF; no consumer was pruned and no build entry was swept\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: consumer registry <temporary-path> home/consumers.json cannot be trusted: manager_state_unreadable: <temporary-path> home/consumers.json: EOF; no consumer was pruned and no build entry was swept\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/invalid_marker
    gc_test.go:92: pass 1: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary-path> is invalid (install_marker_invalid): install_marker_invalid: <temporary-path> malformed protocol JSON object; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary-path> is invalid (install_marker_invalid): install_marker_invalid: <temporary-path> malformed protocol JSON object; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/unreadable_marker
    gc_test.go:92: pass 1: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary-path> cannot be inspected (manager_state_unreadable): manager_state_unreadable: <temporary-path> read <temporary-path> is a directory; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary-path> cannot be inspected (manager_state_unreadable): manager_state_unreadable: <temporary-path> read <temporary-path> is a directory; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/complete_references
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\ngc: 1 runtime entry removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 43198: invalid argument; all builds retained\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 43198: invalid argument; all builds retained\n"
--- PASS: TestGCPreservesRuntimeOnUncertainMarks (2.21s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/truncated_registry (0.40s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/invalid_marker (0.41s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/unreadable_marker (0.75s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/complete_references (0.64s)
PASS
ok  	github.com/relux-works/curator/cmd/curator	2.809s
```

## mutant1 (exit 1)

```text
WORK=<temporary-path>
=== RUN   TestGCPreservesRuntimeOnUncertainMarks
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/truncated_registry
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\nremoved runtime skill-a/1111111111111111111111111111111111111111\ngc: 2 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: consumer registry <temporary-path> home/consumers.json cannot be trusted: manager_state_unreadable: <temporary-path> home/consumers.json: EOF; no consumer was pruned and no build entry was swept\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:96: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:104: pass 1: uncertain marks removed an unreferenced runtime too: stat <temporary-path> home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
    gc_test.go:89: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: consumer registry <temporary-path> home/consumers.json cannot be trusted: manager_state_unreadable: <temporary-path> home/consumers.json: EOF; no consumer was pruned and no build entry was swept\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:96: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:104: pass 2: uncertain marks removed an unreferenced runtime too: stat <temporary-path> home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/invalid_marker
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\nremoved runtime skill-a/1111111111111111111111111111111111111111\ngc: 2 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary-path> is invalid (install_marker_invalid): install_marker_invalid: <temporary-path> malformed protocol JSON object; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:96: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:104: pass 1: uncertain marks removed an unreferenced runtime too: stat <temporary-path> home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
    gc_test.go:89: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary-path> is invalid (install_marker_invalid): install_marker_invalid: <temporary-path> malformed protocol JSON object; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:96: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:104: pass 2: uncertain marks removed an unreferenced runtime too: stat <temporary-path> home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/unreadable_marker
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\nremoved runtime skill-a/1111111111111111111111111111111111111111\ngc: 2 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary-path> cannot be inspected (manager_state_unreadable): manager_state_unreadable: <temporary-path> read <temporary-path> is a directory; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:96: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:104: pass 1: uncertain marks removed an unreferenced runtime too: stat <temporary-path> home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
    gc_test.go:89: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary-path> cannot be inspected (manager_state_unreadable): manager_state_unreadable: <temporary-path> read <temporary-path> is a directory; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:96: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:104: pass 2: uncertain marks removed an unreferenced runtime too: stat <temporary-path> home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/complete_references
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\ngc: 1 runtime entry removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 54559: invalid argument; all builds retained\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 54559: invalid argument; all builds retained\n"
--- FAIL: TestGCPreservesRuntimeOnUncertainMarks (1.92s)
    --- FAIL: TestGCPreservesRuntimeOnUncertainMarks/truncated_registry (0.44s)
    --- FAIL: TestGCPreservesRuntimeOnUncertainMarks/invalid_marker (0.57s)
    --- FAIL: TestGCPreservesRuntimeOnUncertainMarks/unreadable_marker (0.56s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/complete_references (0.35s)
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	2.420s
FAIL
```

## mutant2 (exit 1)

```text
WORK=<temporary-path>
=== RUN   TestGCPreservesRuntimeOnUncertainMarks
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/truncated_registry
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\nremoved runtime skill-a/1111111111111111111111111111111111111111\ngc: 2 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 43198: invalid argument; all builds retained\n"
    gc_test.go:96: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:100: pass 1: missing warning "cannot be trusted"
    gc_test.go:100: pass 1: missing warning "runtime sweep skipped"
    gc_test.go:100: pass 1: missing warning "live reference set could not be proven complete"
    gc_test.go:104: pass 1: uncertain marks removed an unreferenced runtime too: stat <temporary-path> home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
    gc_test.go:89: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 43198: invalid argument; all builds retained\n"
    gc_test.go:96: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:100: pass 2: missing warning "cannot be trusted"
    gc_test.go:100: pass 2: missing warning "runtime sweep skipped"
    gc_test.go:100: pass 2: missing warning "live reference set could not be proven complete"
    gc_test.go:104: pass 2: uncertain marks removed an unreferenced runtime too: stat <temporary-path> home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/invalid_marker
    gc_test.go:92: pass 1: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary-path> is invalid (install_marker_invalid): install_marker_invalid: <temporary-path> malformed protocol JSON object; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary-path> is invalid (install_marker_invalid): install_marker_invalid: <temporary-path> malformed protocol JSON object; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/unreadable_marker
    gc_test.go:92: pass 1: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary-path> cannot be inspected (manager_state_unreadable): manager_state_unreadable: <temporary-path> read <temporary-path> is a directory; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary-path> cannot be inspected (manager_state_unreadable): manager_state_unreadable: <temporary-path> read <temporary-path> is a directory; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/complete_references
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\ngc: 1 runtime entry removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 43198: invalid argument; all builds retained\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 43198: invalid argument; all builds retained\n"
--- FAIL: TestGCPreservesRuntimeOnUncertainMarks (1.71s)
    --- FAIL: TestGCPreservesRuntimeOnUncertainMarks/truncated_registry (0.58s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/invalid_marker (0.53s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/unreadable_marker (0.29s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/complete_references (0.31s)
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	2.210s
FAIL
```

## mutant3 (exit 1)

```text
WORK=<temporary-path>
=== RUN   TestGCPreservesRuntimeOnUncertainMarks
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/truncated_registry
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\nremoved runtime skill-a/1111111111111111111111111111111111111111\ngc: 2 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: consumer registry <temporary-path> home/consumers.json cannot be trusted: manager_state_unreadable: <temporary-path> home/consumers.json: EOF; no consumer was pruned and no build entry was swept\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 43198: invalid argument; all builds retained\n"
    gc_test.go:96: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:100: pass 1: missing warning "runtime sweep skipped"
    gc_test.go:100: pass 1: missing warning "live reference set could not be proven complete"
    gc_test.go:104: pass 1: uncertain marks removed an unreferenced runtime too: stat <temporary-path> home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
    gc_test.go:89: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: consumer registry <temporary-path> home/consumers.json cannot be trusted: manager_state_unreadable: <temporary-path> home/consumers.json: EOF; no consumer was pruned and no build entry was swept\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 43198: invalid argument; all builds retained\n"
    gc_test.go:96: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:100: pass 2: missing warning "runtime sweep skipped"
    gc_test.go:100: pass 2: missing warning "live reference set could not be proven complete"
    gc_test.go:104: pass 2: uncertain marks removed an unreferenced runtime too: stat <temporary-path> home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/invalid_marker
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\nremoved runtime skill-a/1111111111111111111111111111111111111111\ngc: 2 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary-path> is invalid (install_marker_invalid): install_marker_invalid: <temporary-path> malformed protocol JSON object; repair or remove that installation\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 43198: invalid argument; all builds retained\n"
    gc_test.go:96: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:100: pass 1: missing warning "runtime sweep skipped"
    gc_test.go:100: pass 1: missing warning "live reference set could not be proven complete"
    gc_test.go:104: pass 1: uncertain marks removed an unreferenced runtime too: stat <temporary-path> home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
    gc_test.go:89: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary-path> is invalid (install_marker_invalid): install_marker_invalid: <temporary-path> malformed protocol JSON object; repair or remove that installation\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 43198: invalid argument; all builds retained\n"
    gc_test.go:96: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:100: pass 2: missing warning "runtime sweep skipped"
    gc_test.go:100: pass 2: missing warning "live reference set could not be proven complete"
    gc_test.go:104: pass 2: uncertain marks removed an unreferenced runtime too: stat <temporary-path> home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/unreadable_marker
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\nremoved runtime skill-a/1111111111111111111111111111111111111111\ngc: 2 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary-path> cannot be inspected (manager_state_unreadable): manager_state_unreadable: <temporary-path> read <temporary-path> is a directory; repair or remove that installation\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 43198: invalid argument; all builds retained\n"
    gc_test.go:96: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:100: pass 1: missing warning "runtime sweep skipped"
    gc_test.go:100: pass 1: missing warning "live reference set could not be proven complete"
    gc_test.go:104: pass 1: uncertain marks removed an unreferenced runtime too: stat <temporary-path> home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
    gc_test.go:89: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary-path> cannot be inspected (manager_state_unreadable): manager_state_unreadable: <temporary-path> read <temporary-path> is a directory; repair or remove that installation\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 43198: invalid argument; all builds retained\n"
    gc_test.go:96: shim = "", fork/exec <temporary-path> no such file or directory; want runtime-ok
    gc_test.go:100: pass 2: missing warning "runtime sweep skipped"
    gc_test.go:100: pass 2: missing warning "live reference set could not be proven complete"
    gc_test.go:104: pass 2: uncertain marks removed an unreferenced runtime too: stat <temporary-path> home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/complete_references
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\ngc: 1 runtime entry removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 43198: invalid argument; all builds retained\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 43198: invalid argument; all builds retained\n"
--- FAIL: TestGCPreservesRuntimeOnUncertainMarks (1.36s)
    --- FAIL: TestGCPreservesRuntimeOnUncertainMarks/truncated_registry (0.39s)
    --- FAIL: TestGCPreservesRuntimeOnUncertainMarks/invalid_marker (0.31s)
    --- FAIL: TestGCPreservesRuntimeOnUncertainMarks/unreadable_marker (0.33s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/complete_references (0.33s)
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	1.873s
FAIL
```

## final-green (exit 0)

```text
WORK=<temporary-path>
=== RUN   TestAuthoritativeGarbageCollectionRootsAreRetained
=== RUN   TestAuthoritativeGarbageCollectionRootsAreRetained/status-current
=== RUN   TestAuthoritativeGarbageCollectionRootsAreRetained/status-missing-snapshot
=== RUN   TestAuthoritativeGarbageCollectionRootsAreRetained/status-unreadable-protected-state
=== RUN   TestAuthoritativeGarbageCollectionRootsAreRetained/repair-reacquires-exact-source
=== RUN   TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots
=== RUN   TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots/artifact-receipts
=== RUN   TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots/in-flight-journals
=== RUN   TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots/install-markers
=== RUN   TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots/protected-snapshots
=== RUN   TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots/uncertain-entries
=== NAME  TestAuthoritativeGarbageCollectionRootsAreRetained
    coverage.go:236: published cases external-repository-lifecycle/status-repair-gc-cases: 1 driven, 0 known-gap, 4 bound, 0 skipped, 5 total
--- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained (0.57s)
    --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/status-current (0.00s)
    --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/status-missing-snapshot (0.00s)
    --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/status-unreadable-protected-state (0.00s)
    --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/repair-reacquires-exact-source (0.00s)
    --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots (0.57s)
        --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots/artifact-receipts (0.27s)
        --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots/in-flight-journals (0.22s)
        --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots/install-markers (0.00s)
        --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots/protected-snapshots (0.00s)
        --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots/uncertain-entries (0.07s)
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses/registry_repeating_a_known_member
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses/project_with_only_an_invalid_marker
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses/unreadable_installed_skill_directory
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses/installed_skill_replaced_by_a_file
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses/skill_root_replaced_by_a_file
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses/corrupt_consumer_registry
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses/registry_of_an_unknown_schema
--- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses (0.14s)
    --- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses/registry_repeating_a_known_member (0.00s)
    --- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses/project_with_only_an_invalid_marker (0.03s)
    --- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses/unreadable_installed_skill_directory (0.05s)
    --- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses/installed_skill_replaced_by_a_file (0.01s)
    --- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses/skill_root_replaced_by_a_file (0.05s)
    --- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses/corrupt_consumer_registry (0.01s)
    --- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses/registry_of_an_unknown_schema (0.00s)
=== RUN   TestNonDirectoryScopeMembersDoNotBlockMaintenance
--- PASS: TestNonDirectoryScopeMembersDoNotBlockMaintenance (0.01s)
=== RUN   TestRecordingAConsumerNeverOverwritesAnUntrustedRegistry
--- PASS: TestRecordingAConsumerNeverOverwritesAnUntrustedRegistry (0.00s)
=== RUN   TestParseConsumersAcceptsOnlyTheCanonicalRegistry
--- PASS: TestParseConsumersAcceptsOnlyTheCanonicalRegistry (0.00s)
=== RUN   TestARepeatedRegistryMemberNeverEmptiesTheRegistry
--- PASS: TestARepeatedRegistryMemberNeverEmptiesTheRegistry (0.00s)
=== RUN   TestWritersRefuseARepeatedRegistryMember
--- PASS: TestWritersRefuseARepeatedRegistryMember (0.00s)
=== RUN   TestStructDecodingWouldTrustARepeatedRegistryMember
--- PASS: TestStructDecodingWouldTrustARepeatedRegistryMember (0.00s)
=== RUN   TestCollectStaysFailSafeOnRedirectedUnixScopes
=== RUN   TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_installed_skill
=== RUN   TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_global_scope_root
=== RUN   TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_hybrid_scope_root
=== RUN   TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_project_skill_root
--- PASS: TestCollectStaysFailSafeOnRedirectedUnixScopes (0.02s)
    --- PASS: TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_installed_skill (0.01s)
    --- PASS: TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_global_scope_root (0.00s)
    --- PASS: TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_hybrid_scope_root (0.00s)
    --- PASS: TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_project_skill_root (0.00s)
=== RUN   TestCollectRetainsDraftSourceV1Runtime
--- PASS: TestCollectRetainsDraftSourceV1Runtime (0.00s)
=== RUN   TestCollectRetainsLiveProcessBuild
=== RUN   TestCollectRetainsLiveProcessBuild/in-use
=== RUN   TestCollectRetainsLiveProcessBuild/enumeration-error
--- PASS: TestCollectRetainsLiveProcessBuild (0.28s)
    --- PASS: TestCollectRetainsLiveProcessBuild/in-use (0.14s)
    --- PASS: TestCollectRetainsLiveProcessBuild/enumeration-error (0.14s)
=== RUN   TestCollectSweepsOnlyUnreferencedProtectedEntries
--- PASS: TestCollectSweepsOnlyUnreferencedProtectedEntries (0.14s)
=== RUN   TestCollectRetainsAJournalOwnedEntry
--- PASS: TestCollectRetainsAJournalOwnedEntry (0.07s)
=== RUN   TestCollectRetainsBuildEntriesWhenAMarkerCannotBeRead
--- PASS: TestCollectRetainsBuildEntriesWhenAMarkerCannotBeRead (0.07s)
=== RUN   TestCollectMarksBuildKeysFromEveryBuildBearingMarkerSchema
--- PASS: TestCollectMarksBuildKeysFromEveryBuildBearingMarkerSchema (0.01s)
=== RUN   TestCollectMarksBuildKeysFromEveryLiveScope
--- PASS: TestCollectMarksBuildKeysFromEveryLiveScope (0.01s)
=== RUN   TestCollectMarksRuntimeFromEverySupportedMarkerSchema
--- PASS: TestCollectMarksRuntimeFromEverySupportedMarkerSchema (0.01s)
=== RUN   TestCollectSkipsTheBuildSweepOnUnprovableReferences
=== RUN   TestCollectSkipsTheBuildSweepOnUnprovableReferences/invalid_marker
=== RUN   TestCollectSkipsTheBuildSweepOnUnprovableReferences/corrupt_consumer_registry
=== RUN   TestCollectSkipsTheBuildSweepOnUnprovableReferences/unreadable_global_skills_directory
--- PASS: TestCollectSkipsTheBuildSweepOnUnprovableReferences (0.01s)
    --- PASS: TestCollectSkipsTheBuildSweepOnUnprovableReferences/invalid_marker (0.00s)
    --- PASS: TestCollectSkipsTheBuildSweepOnUnprovableReferences/corrupt_consumer_registry (0.00s)
    --- PASS: TestCollectSkipsTheBuildSweepOnUnprovableReferences/unreadable_global_skills_directory (0.00s)
=== RUN   TestCollectRequiresTheHomeLock
--- PASS: TestCollectRequiresTheHomeLock (0.00s)
=== RUN   TestCollectPrunesConsumersInsideTheSamePass
--- PASS: TestCollectPrunesConsumersInsideTheSamePass (0.00s)
=== RUN   TestCollectReportsSweepFailuresWithoutLosingRuntimeWork
--- PASS: TestCollectReportsSweepFailuresWithoutLosingRuntimeWork (0.00s)
=== RUN   TestCollectForwardsTheSweepClockAndGrace
--- PASS: TestCollectForwardsTheSweepClockAndGrace (0.00s)
=== RUN   TestCollectResolvesTheRealStoreByDefault
--- PASS: TestCollectResolvesTheRealStoreByDefault (0.00s)
PASS
ok  	github.com/relux-works/curator/internal/scopes	1.801s
=== RUN   TestGCPreservesRuntimeOnUncertainMarks
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/truncated_registry
    gc_test.go:92: pass 1: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: consumer registry <temporary-path> home/consumers.json cannot be trusted: manager_state_unreadable: <temporary-path> home/consumers.json: EOF; no consumer was pruned and no build entry was swept\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: consumer registry <temporary-path> home/consumers.json cannot be trusted: manager_state_unreadable: <temporary-path> home/consumers.json: EOF; no consumer was pruned and no build entry was swept\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/invalid_marker
    gc_test.go:92: pass 1: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary-path> is invalid (install_marker_invalid): install_marker_invalid: <temporary-path> malformed protocol JSON object; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary-path> is invalid (install_marker_invalid): install_marker_invalid: <temporary-path> malformed protocol JSON object; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/unreadable_marker
    gc_test.go:92: pass 1: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary-path> cannot be inspected (manager_state_unreadable): manager_state_unreadable: <temporary-path> read <temporary-path> is a directory; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary-path> cannot be inspected (manager_state_unreadable): manager_state_unreadable: <temporary-path> read <temporary-path> is a directory; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/complete_references
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\ngc: 1 runtime entry removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 43198: invalid argument; all builds retained\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 43198: invalid argument; all builds retained\n"
--- PASS: TestGCPreservesRuntimeOnUncertainMarks (1.65s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/truncated_registry (0.49s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/invalid_marker (0.41s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/unreadable_marker (0.43s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/complete_references (0.32s)
=== RUN   TestGCRetainsLiveProcessBuild
    gc_test.go:193: native enumeration could not inspect every process; exercised fail-safe retention: warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default
        curator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 43198: invalid argument; all builds retained
--- PASS: TestGCRetainsLiveProcessBuild (26.36s)
=== RUN   TestGCPrunesDeadConsumersUnderTheHomeLock
=== PAUSE TestGCPrunesDeadConsumersUnderTheHomeLock
=== RUN   TestGCFailsClosedForUntrustedCurrentPathSource
--- PASS: TestGCFailsClosedForUntrustedCurrentPathSource (4.68s)
=== RUN   TestGCWaitsForTheHomeLock
=== PAUSE TestGCWaitsForTheHomeLock
=== RUN   TestGCRunsSerializedAcrossConcurrentInvocations
=== PAUSE TestGCRunsSerializedAcrossConcurrentInvocations
=== RUN   TestGCRetainsAndReportsReferencedCompiledState
=== PAUSE TestGCRetainsAndReportsReferencedCompiledState
=== CONT  TestGCPrunesDeadConsumersUnderTheHomeLock
=== CONT  TestGCRunsSerializedAcrossConcurrentInvocations
=== CONT  TestGCRetainsAndReportsReferencedCompiledState
=== CONT  TestGCWaitsForTheHomeLock
--- PASS: TestGCPrunesDeadConsumersUnderTheHomeLock (0.05s)
--- PASS: TestGCRunsSerializedAcrossConcurrentInvocations (0.15s)
--- PASS: TestGCWaitsForTheHomeLock (0.31s)
--- PASS: TestGCRetainsAndReportsReferencedCompiledState (27.51s)
PASS
ok  	github.com/relux-works/curator/cmd/curator	60.730s
```
