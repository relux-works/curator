# Independent N1 review execution logs

Temporary and home paths redacted. Exit codes and interpretation are in the verdict.

## base

```text
WORK=<temporary>/go-build1320288955
=== RUN   TestGCPreservesRuntimeOnUncertainMarks
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/truncated_registry
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\nremoved runtime skill-a/1111111111111111111111111111111111111111\ngc: 2 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: consumer registry <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry2561655828/001/manager home/consumers.json cannot be trusted: manager_state_unreadable: <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry2561655828/001/manager home/consumers.json: EOF; no consumer was pruned and no build entry was swept\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:96: shim = "", fork/exec <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry2561655828/002/.agents/bin/tool: no such file or directory; want runtime-ok
    gc_test.go:100: pass 1: missing warning "runtime sweep skipped"
    gc_test.go:104: pass 1: uncertain marks removed an unreferenced runtime too: stat <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry2561655828/001/manager home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
    gc_test.go:89: shim = "", fork/exec <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry2561655828/002/.agents/bin/tool: no such file or directory; want runtime-ok
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: consumer registry <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry2561655828/001/manager home/consumers.json cannot be trusted: manager_state_unreadable: <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry2561655828/001/manager home/consumers.json: EOF; no consumer was pruned and no build entry was swept\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:96: shim = "", fork/exec <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry2561655828/002/.agents/bin/tool: no such file or directory; want runtime-ok
    gc_test.go:100: pass 2: missing warning "runtime sweep skipped"
    gc_test.go:104: pass 2: uncertain marks removed an unreferenced runtime too: stat <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry2561655828/001/manager home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/invalid_marker
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\nremoved runtime skill-a/1111111111111111111111111111111111111111\ngc: 2 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker119380404/002/.agents/skills/skill-a/.csk-install.json is invalid (install_marker_invalid): install_marker_invalid: <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker119380404/002/.agents/skills/skill-a/.csk-install.json: malformed protocol JSON object; repair or remove that installation\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:96: shim = "", fork/exec <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker119380404/002/.agents/bin/tool: no such file or directory; want runtime-ok
    gc_test.go:100: pass 1: missing warning "runtime sweep skipped"
    gc_test.go:104: pass 1: uncertain marks removed an unreferenced runtime too: stat <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker119380404/001/manager home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
    gc_test.go:89: shim = "", fork/exec <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker119380404/002/.agents/bin/tool: no such file or directory; want runtime-ok
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker119380404/002/.agents/skills/skill-a/.csk-install.json is invalid (install_marker_invalid): install_marker_invalid: <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker119380404/002/.agents/skills/skill-a/.csk-install.json: malformed protocol JSON object; repair or remove that installation\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:96: shim = "", fork/exec <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker119380404/002/.agents/bin/tool: no such file or directory; want runtime-ok
    gc_test.go:100: pass 2: missing warning "runtime sweep skipped"
    gc_test.go:104: pass 2: uncertain marks removed an unreferenced runtime too: stat <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker119380404/001/manager home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/unreadable_marker
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\nremoved runtime skill-a/1111111111111111111111111111111111111111\ngc: 2 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker645755069/002/.agents/skills/skill-a/.csk-install.json cannot be inspected (manager_state_unreadable): manager_state_unreadable: <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker645755069/002/.agents/skills/skill-a/.csk-install.json: read <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker645755069/002/.agents/skills/skill-a/.csk-install.json: is a directory; repair or remove that installation\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:96: shim = "", fork/exec <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker645755069/002/.agents/bin/tool: no such file or directory; want runtime-ok
    gc_test.go:100: pass 1: missing warning "runtime sweep skipped"
    gc_test.go:104: pass 1: uncertain marks removed an unreferenced runtime too: stat <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker645755069/001/manager home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
    gc_test.go:89: shim = "", fork/exec <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker645755069/002/.agents/bin/tool: no such file or directory; want runtime-ok
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker645755069/002/.agents/skills/skill-a/.csk-install.json cannot be inspected (manager_state_unreadable): manager_state_unreadable: <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker645755069/002/.agents/skills/skill-a/.csk-install.json: read <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker645755069/002/.agents/skills/skill-a/.csk-install.json: is a directory; repair or remove that installation\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:96: shim = "", fork/exec <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker645755069/002/.agents/bin/tool: no such file or directory; want runtime-ok
    gc_test.go:100: pass 2: missing warning "runtime sweep skipped"
    gc_test.go:104: pass 2: uncertain marks removed an unreferenced runtime too: stat <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker645755069/001/manager home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/complete_references
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\ngc: 1 runtime entry removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 31052: invalid argument; all builds retained\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 33390: invalid argument; all builds retained\n"
--- FAIL: TestGCPreservesRuntimeOnUncertainMarks (6.19s)
    --- FAIL: TestGCPreservesRuntimeOnUncertainMarks/truncated_registry (0.87s)
    --- FAIL: TestGCPreservesRuntimeOnUncertainMarks/invalid_marker (3.26s)
    --- FAIL: TestGCPreservesRuntimeOnUncertainMarks/unreadable_marker (0.99s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/complete_references (1.07s)
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	8.878s
FAIL
```

## green

```text
WORK=<temporary>/go-build394131746
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
--- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained (0.49s)
    --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/status-current (0.00s)
    --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/status-missing-snapshot (0.00s)
    --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/status-unreadable-protected-state (0.00s)
    --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/repair-reacquires-exact-source (0.00s)
    --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots (0.49s)
        --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots/artifact-receipts (0.21s)
        --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots/in-flight-journals (0.20s)
        --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots/install-markers (0.00s)
        --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots/protected-snapshots (0.01s)
        --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots/uncertain-entries (0.07s)
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses/project_with_only_an_invalid_marker
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses/unreadable_installed_skill_directory
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses/installed_skill_replaced_by_a_file
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses/skill_root_replaced_by_a_file
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses/corrupt_consumer_registry
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses/registry_of_an_unknown_schema
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses/registry_repeating_a_known_member
--- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses (0.03s)
    --- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses/project_with_only_an_invalid_marker (0.00s)
    --- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses/unreadable_installed_skill_directory (0.00s)
    --- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses/installed_skill_replaced_by_a_file (0.01s)
    --- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses/skill_root_replaced_by_a_file (0.00s)
    --- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses/corrupt_consumer_registry (0.00s)
    --- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses/registry_of_an_unknown_schema (0.00s)
    --- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses/registry_repeating_a_known_member (0.00s)
=== RUN   TestNonDirectoryScopeMembersDoNotBlockMaintenance
--- PASS: TestNonDirectoryScopeMembersDoNotBlockMaintenance (0.00s)
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
=== RUN   TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_project_skill_root
=== RUN   TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_installed_skill
=== RUN   TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_global_scope_root
=== RUN   TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_hybrid_scope_root
--- PASS: TestCollectStaysFailSafeOnRedirectedUnixScopes (0.02s)
    --- PASS: TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_project_skill_root (0.01s)
    --- PASS: TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_installed_skill (0.01s)
    --- PASS: TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_global_scope_root (0.01s)
    --- PASS: TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_hybrid_scope_root (0.01s)
=== RUN   TestCollectRetainsDraftSourceV1Runtime
--- PASS: TestCollectRetainsDraftSourceV1Runtime (0.01s)
=== RUN   TestCollectRetainsLiveProcessBuild
=== RUN   TestCollectRetainsLiveProcessBuild/in-use
=== RUN   TestCollectRetainsLiveProcessBuild/enumeration-error
--- PASS: TestCollectRetainsLiveProcessBuild (0.31s)
    --- PASS: TestCollectRetainsLiveProcessBuild/in-use (0.15s)
    --- PASS: TestCollectRetainsLiveProcessBuild/enumeration-error (0.16s)
=== RUN   TestCollectSweepsOnlyUnreferencedProtectedEntries
--- PASS: TestCollectSweepsOnlyUnreferencedProtectedEntries (0.15s)
=== RUN   TestCollectRetainsAJournalOwnedEntry
--- PASS: TestCollectRetainsAJournalOwnedEntry (0.07s)
=== RUN   TestCollectRetainsBuildEntriesWhenAMarkerCannotBeRead
--- PASS: TestCollectRetainsBuildEntriesWhenAMarkerCannotBeRead (0.08s)
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
--- PASS: TestCollectPrunesConsumersInsideTheSamePass (0.03s)
=== RUN   TestCollectReportsSweepFailuresWithoutLosingRuntimeWork
--- PASS: TestCollectReportsSweepFailuresWithoutLosingRuntimeWork (0.00s)
=== RUN   TestCollectForwardsTheSweepClockAndGrace
--- PASS: TestCollectForwardsTheSweepClockAndGrace (0.00s)
=== RUN   TestCollectResolvesTheRealStoreByDefault
--- PASS: TestCollectResolvesTheRealStoreByDefault (0.01s)
PASS
ok  	github.com/relux-works/curator/internal/scopes	2.132s
testing: warning: no tests to run
PASS
ok  	github.com/relux-works/curator/internal/runtimestore	1.519s [no tests to run]
=== RUN   TestGCPreservesRuntimeOnUncertainMarks
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/truncated_registry
    gc_test.go:92: pass 1: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: consumer registry <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry1467755767/001/manager home/consumers.json cannot be trusted: manager_state_unreadable: <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry1467755767/001/manager home/consumers.json: EOF; no consumer was pruned and no build entry was swept\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: consumer registry <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry1467755767/001/manager home/consumers.json cannot be trusted: manager_state_unreadable: <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry1467755767/001/manager home/consumers.json: EOF; no consumer was pruned and no build entry was swept\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/invalid_marker
    gc_test.go:92: pass 1: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker1845384221/002/.agents/skills/skill-a/.csk-install.json is invalid (install_marker_invalid): install_marker_invalid: <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker1845384221/002/.agents/skills/skill-a/.csk-install.json: malformed protocol JSON object; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker1845384221/002/.agents/skills/skill-a/.csk-install.json is invalid (install_marker_invalid): install_marker_invalid: <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker1845384221/002/.agents/skills/skill-a/.csk-install.json: malformed protocol JSON object; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/unreadable_marker
    gc_test.go:92: pass 1: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker1965791215/002/.agents/skills/skill-a/.csk-install.json cannot be inspected (manager_state_unreadable): manager_state_unreadable: <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker1965791215/002/.agents/skills/skill-a/.csk-install.json: read <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker1965791215/002/.agents/skills/skill-a/.csk-install.json: is a directory; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker1965791215/002/.agents/skills/skill-a/.csk-install.json cannot be inspected (manager_state_unreadable): manager_state_unreadable: <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker1965791215/002/.agents/skills/skill-a/.csk-install.json: read <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker1965791215/002/.agents/skills/skill-a/.csk-install.json: is a directory; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/complete_references
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\ngc: 1 runtime entry removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 99560: invalid argument; all builds retained\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 99560: invalid argument; all builds retained\n"
--- PASS: TestGCPreservesRuntimeOnUncertainMarks (6.93s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/truncated_registry (2.93s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/invalid_marker (1.60s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/unreadable_marker (1.31s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/complete_references (1.08s)
=== RUN   TestGCRetainsLiveProcessBuild
    gc_test.go:193: native enumeration could not inspect every process; exercised fail-safe retention: warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default
        curator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 99560: invalid argument; all builds retained
--- PASS: TestGCRetainsLiveProcessBuild (73.27s)
=== RUN   TestGCPrunesDeadConsumersUnderTheHomeLock
=== PAUSE TestGCPrunesDeadConsumersUnderTheHomeLock
=== RUN   TestGCFailsClosedForUntrustedCurrentPathSource
--- PASS: TestGCFailsClosedForUntrustedCurrentPathSource (2.80s)
=== RUN   TestGCWaitsForTheHomeLock
=== PAUSE TestGCWaitsForTheHomeLock
=== RUN   TestGCRunsSerializedAcrossConcurrentInvocations
=== PAUSE TestGCRunsSerializedAcrossConcurrentInvocations
=== RUN   TestGCRetainsAndReportsReferencedCompiledState
=== PAUSE TestGCRetainsAndReportsReferencedCompiledState
=== CONT  TestGCPrunesDeadConsumersUnderTheHomeLock
=== CONT  TestGCRunsSerializedAcrossConcurrentInvocations
--- PASS: TestGCPrunesDeadConsumersUnderTheHomeLock (0.01s)
=== CONT  TestGCWaitsForTheHomeLock
--- PASS: TestGCRunsSerializedAcrossConcurrentInvocations (0.05s)
=== CONT  TestGCRetainsAndReportsReferencedCompiledState
--- PASS: TestGCWaitsForTheHomeLock (0.27s)
--- PASS: TestGCRetainsAndReportsReferencedCompiledState (54.27s)
PASS
ok  	github.com/relux-works/curator/cmd/curator	138.847s
```

## rc14

```text
WORK=<temporary>/go-build4234818334
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
--- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained (1.95s)
    --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/status-current (0.00s)
    --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/status-missing-snapshot (0.00s)
    --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/status-unreadable-protected-state (0.00s)
    --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/repair-reacquires-exact-source (0.00s)
    --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots (1.95s)
        --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots/artifact-receipts (1.23s)
        --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots/in-flight-journals (0.26s)
        --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots/install-markers (0.01s)
        --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots/protected-snapshots (0.01s)
        --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots/uncertain-entries (0.43s)
PASS
ok  	github.com/relux-works/curator/internal/scopes	4.459s
```

## reorder

```text
WORK=<temporary>/go-build1133227876
=== RUN   TestGCPreservesRuntimeOnUncertainMarks
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/truncated_registry
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\nremoved runtime skill-a/1111111111111111111111111111111111111111\ngc: 2 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: consumer registry <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry76252208/001/manager home/consumers.json cannot be trusted: manager_state_unreadable: <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry76252208/001/manager home/consumers.json: EOF; no consumer was pruned and no build entry was swept\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:96: shim = "", fork/exec <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry76252208/002/.agents/bin/tool: no such file or directory; want runtime-ok
    gc_test.go:104: pass 1: uncertain marks removed an unreferenced runtime too: stat <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry76252208/001/manager home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
    gc_test.go:89: shim = "", fork/exec <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry76252208/002/.agents/bin/tool: no such file or directory; want runtime-ok
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: consumer registry <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry76252208/001/manager home/consumers.json cannot be trusted: manager_state_unreadable: <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry76252208/001/manager home/consumers.json: EOF; no consumer was pruned and no build entry was swept\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:96: shim = "", fork/exec <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry76252208/002/.agents/bin/tool: no such file or directory; want runtime-ok
    gc_test.go:104: pass 2: uncertain marks removed an unreferenced runtime too: stat <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry76252208/001/manager home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/invalid_marker
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\nremoved runtime skill-a/1111111111111111111111111111111111111111\ngc: 2 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker3057807409/002/.agents/skills/skill-a/.csk-install.json is invalid (install_marker_invalid): install_marker_invalid: <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker3057807409/002/.agents/skills/skill-a/.csk-install.json: malformed protocol JSON object; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:96: shim = "", fork/exec <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker3057807409/002/.agents/bin/tool: no such file or directory; want runtime-ok
    gc_test.go:104: pass 1: uncertain marks removed an unreferenced runtime too: stat <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker3057807409/001/manager home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
    gc_test.go:89: shim = "", fork/exec <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker3057807409/002/.agents/bin/tool: no such file or directory; want runtime-ok
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker3057807409/002/.agents/skills/skill-a/.csk-install.json is invalid (install_marker_invalid): install_marker_invalid: <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker3057807409/002/.agents/skills/skill-a/.csk-install.json: malformed protocol JSON object; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:96: shim = "", fork/exec <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker3057807409/002/.agents/bin/tool: no such file or directory; want runtime-ok
    gc_test.go:104: pass 2: uncertain marks removed an unreferenced runtime too: stat <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker3057807409/001/manager home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/unreadable_marker
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\nremoved runtime skill-a/1111111111111111111111111111111111111111\ngc: 2 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2311242562/002/.agents/skills/skill-a/.csk-install.json cannot be inspected (manager_state_unreadable): manager_state_unreadable: <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2311242562/002/.agents/skills/skill-a/.csk-install.json: read <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2311242562/002/.agents/skills/skill-a/.csk-install.json: is a directory; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:96: shim = "", fork/exec <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2311242562/002/.agents/bin/tool: no such file or directory; want runtime-ok
    gc_test.go:104: pass 1: uncertain marks removed an unreferenced runtime too: stat <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2311242562/001/manager home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
    gc_test.go:89: shim = "", fork/exec <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2311242562/002/.agents/bin/tool: no such file or directory; want runtime-ok
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2311242562/002/.agents/skills/skill-a/.csk-install.json cannot be inspected (manager_state_unreadable): manager_state_unreadable: <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2311242562/002/.agents/skills/skill-a/.csk-install.json: read <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2311242562/002/.agents/skills/skill-a/.csk-install.json: is a directory; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:96: shim = "", fork/exec <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2311242562/002/.agents/bin/tool: no such file or directory; want runtime-ok
    gc_test.go:104: pass 2: uncertain marks removed an unreferenced runtime too: stat <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2311242562/001/manager home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/complete_references
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\ngc: 1 runtime entry removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 43666: invalid argument; all builds retained\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 43666: invalid argument; all builds retained\n"
--- FAIL: TestGCPreservesRuntimeOnUncertainMarks (2.87s)
    --- FAIL: TestGCPreservesRuntimeOnUncertainMarks/truncated_registry (0.65s)
    --- FAIL: TestGCPreservesRuntimeOnUncertainMarks/invalid_marker (0.73s)
    --- FAIL: TestGCPreservesRuntimeOnUncertainMarks/unreadable_marker (0.72s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/complete_references (0.77s)
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	3.740s
FAIL
```

## drop-registry

```text
WORK=<temporary>/go-build3033963012
=== RUN   TestGCPreservesRuntimeOnUncertainMarks
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/truncated_registry
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\nremoved runtime skill-a/1111111111111111111111111111111111111111\ngc: 2 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 43666: invalid argument; all builds retained\n"
    gc_test.go:96: shim = "", fork/exec <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry4215805948/002/.agents/bin/tool: no such file or directory; want runtime-ok
    gc_test.go:100: pass 1: missing warning "cannot be trusted"
    gc_test.go:100: pass 1: missing warning "runtime sweep skipped"
    gc_test.go:100: pass 1: missing warning "live reference set could not be proven complete"
    gc_test.go:104: pass 1: uncertain marks removed an unreferenced runtime too: stat <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry4215805948/001/manager home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
    gc_test.go:89: shim = "", fork/exec <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry4215805948/002/.agents/bin/tool: no such file or directory; want runtime-ok
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 43666: invalid argument; all builds retained\n"
    gc_test.go:96: shim = "", fork/exec <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry4215805948/002/.agents/bin/tool: no such file or directory; want runtime-ok
    gc_test.go:100: pass 2: missing warning "cannot be trusted"
    gc_test.go:100: pass 2: missing warning "runtime sweep skipped"
    gc_test.go:100: pass 2: missing warning "live reference set could not be proven complete"
    gc_test.go:104: pass 2: uncertain marks removed an unreferenced runtime too: stat <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry4215805948/001/manager home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/invalid_marker
    gc_test.go:92: pass 1: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker2728174845/002/.agents/skills/skill-a/.csk-install.json is invalid (install_marker_invalid): install_marker_invalid: <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker2728174845/002/.agents/skills/skill-a/.csk-install.json: malformed protocol JSON object; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker2728174845/002/.agents/skills/skill-a/.csk-install.json is invalid (install_marker_invalid): install_marker_invalid: <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker2728174845/002/.agents/skills/skill-a/.csk-install.json: malformed protocol JSON object; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/unreadable_marker
    gc_test.go:92: pass 1: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2792281620/002/.agents/skills/skill-a/.csk-install.json cannot be inspected (manager_state_unreadable): manager_state_unreadable: <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2792281620/002/.agents/skills/skill-a/.csk-install.json: read <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2792281620/002/.agents/skills/skill-a/.csk-install.json: is a directory; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2792281620/002/.agents/skills/skill-a/.csk-install.json cannot be inspected (manager_state_unreadable): manager_state_unreadable: <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2792281620/002/.agents/skills/skill-a/.csk-install.json: read <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2792281620/002/.agents/skills/skill-a/.csk-install.json: is a directory; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/complete_references
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\ngc: 1 runtime entry removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 43666: invalid argument; all builds retained\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 43666: invalid argument; all builds retained\n"
--- FAIL: TestGCPreservesRuntimeOnUncertainMarks (3.37s)
    --- FAIL: TestGCPreservesRuntimeOnUncertainMarks/truncated_registry (0.72s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/invalid_marker (0.97s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/unreadable_marker (0.96s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/complete_references (0.72s)
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	4.644s
FAIL
```

## narrow

```text
WORK=<temporary>/go-build2203652691
=== RUN   TestGCPreservesRuntimeOnUncertainMarks
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/truncated_registry
    gc_test.go:92: pass 1: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: consumer registry <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry3355157822/001/manager home/consumers.json cannot be trusted: manager_state_unreadable: <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry3355157822/001/manager home/consumers.json: EOF; no consumer was pruned and no build entry was swept\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: consumer registry <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry3355157822/001/manager home/consumers.json cannot be trusted: manager_state_unreadable: <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry3355157822/001/manager home/consumers.json: EOF; no consumer was pruned and no build entry was swept\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/invalid_marker
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\nremoved runtime skill-a/1111111111111111111111111111111111111111\ngc: 2 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker2998265127/002/.agents/skills/skill-a/.csk-install.json is invalid (install_marker_invalid): install_marker_invalid: <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker2998265127/002/.agents/skills/skill-a/.csk-install.json: malformed protocol JSON object; repair or remove that installation\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 47855: invalid argument; all builds retained\n"
    gc_test.go:96: shim = "", fork/exec <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker2998265127/002/.agents/bin/tool: no such file or directory; want runtime-ok
    gc_test.go:100: pass 1: missing warning "runtime sweep skipped"
    gc_test.go:100: pass 1: missing warning "live reference set could not be proven complete"
    gc_test.go:104: pass 1: uncertain marks removed an unreferenced runtime too: stat <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker2998265127/001/manager home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
    gc_test.go:89: shim = "", fork/exec <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker2998265127/002/.agents/bin/tool: no such file or directory; want runtime-ok
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker2998265127/002/.agents/skills/skill-a/.csk-install.json is invalid (install_marker_invalid): install_marker_invalid: <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker2998265127/002/.agents/skills/skill-a/.csk-install.json: malformed protocol JSON object; repair or remove that installation\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 99560: invalid argument; all builds retained\n"
    gc_test.go:96: shim = "", fork/exec <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker2998265127/002/.agents/bin/tool: no such file or directory; want runtime-ok
    gc_test.go:100: pass 2: missing warning "runtime sweep skipped"
    gc_test.go:100: pass 2: missing warning "live reference set could not be proven complete"
    gc_test.go:104: pass 2: uncertain marks removed an unreferenced runtime too: stat <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker2998265127/001/manager home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/unreadable_marker
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\nremoved runtime skill-a/1111111111111111111111111111111111111111\ngc: 2 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2762803856/002/.agents/skills/skill-a/.csk-install.json cannot be inspected (manager_state_unreadable): manager_state_unreadable: <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2762803856/002/.agents/skills/skill-a/.csk-install.json: read <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2762803856/002/.agents/skills/skill-a/.csk-install.json: is a directory; repair or remove that installation\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 99560: invalid argument; all builds retained\n"
    gc_test.go:96: shim = "", fork/exec <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2762803856/002/.agents/bin/tool: no such file or directory; want runtime-ok
    gc_test.go:100: pass 1: missing warning "runtime sweep skipped"
    gc_test.go:100: pass 1: missing warning "live reference set could not be proven complete"
    gc_test.go:104: pass 1: uncertain marks removed an unreferenced runtime too: stat <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2762803856/001/manager home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
    gc_test.go:89: shim = "", fork/exec <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2762803856/002/.agents/bin/tool: no such file or directory; want runtime-ok
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2762803856/002/.agents/skills/skill-a/.csk-install.json cannot be inspected (manager_state_unreadable): manager_state_unreadable: <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2762803856/002/.agents/skills/skill-a/.csk-install.json: read <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2762803856/002/.agents/skills/skill-a/.csk-install.json: is a directory; repair or remove that installation\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 99560: invalid argument; all builds retained\n"
    gc_test.go:96: shim = "", fork/exec <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2762803856/002/.agents/bin/tool: no such file or directory; want runtime-ok
    gc_test.go:100: pass 2: missing warning "runtime sweep skipped"
    gc_test.go:100: pass 2: missing warning "live reference set could not be proven complete"
    gc_test.go:104: pass 2: uncertain marks removed an unreferenced runtime too: stat <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2762803856/001/manager home/runtime/orphan/2222222222222222222222222222222222222222/tool: no such file or directory
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/complete_references
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\ngc: 1 runtime entry removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 99560: invalid argument; all builds retained\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 99560: invalid argument; all builds retained\n"
--- FAIL: TestGCPreservesRuntimeOnUncertainMarks (2.60s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/truncated_registry (0.66s)
    --- FAIL: TestGCPreservesRuntimeOnUncertainMarks/invalid_marker (0.73s)
    --- FAIL: TestGCPreservesRuntimeOnUncertainMarks/unreadable_marker (0.75s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/complete_references (0.46s)
FAIL
FAIL	github.com/relux-works/curator/cmd/curator	3.547s
FAIL
```

## restored

```text
WORK=<temporary>/go-build3617486034
=== RUN   TestGCPreservesRuntimeOnUncertainMarks
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/truncated_registry
    gc_test.go:92: pass 1: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: consumer registry <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry1401264818/001/manager home/consumers.json cannot be trusted: manager_state_unreadable: <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry1401264818/001/manager home/consumers.json: EOF; no consumer was pruned and no build entry was swept\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: consumer registry <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry1401264818/001/manager home/consumers.json cannot be trusted: manager_state_unreadable: <temporary>/TestGCPreservesRuntimeOnUncertainMarkstruncated_registry1401264818/001/manager home/consumers.json: EOF; no consumer was pruned and no build entry was swept\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/invalid_marker
    gc_test.go:92: pass 1: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker1987441701/002/.agents/skills/skill-a/.csk-install.json is invalid (install_marker_invalid): install_marker_invalid: <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker1987441701/002/.agents/skills/skill-a/.csk-install.json: malformed protocol JSON object; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker1987441701/002/.agents/skills/skill-a/.csk-install.json is invalid (install_marker_invalid): install_marker_invalid: <temporary>/TestGCPreservesRuntimeOnUncertainMarksinvalid_marker1987441701/002/.agents/skills/skill-a/.csk-install.json: malformed protocol JSON object; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/unreadable_marker
    gc_test.go:92: pass 1: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2304365179/002/.agents/skills/skill-a/.csk-install.json cannot be inspected (manager_state_unreadable): manager_state_unreadable: <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2304365179/002/.agents/skills/skill-a/.csk-install.json: read <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2304365179/002/.agents/skills/skill-a/.csk-install.json: is a directory; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: install marker <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2304365179/002/.agents/skills/skill-a/.csk-install.json cannot be inspected (manager_state_unreadable): manager_state_unreadable: <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2304365179/002/.agents/skills/skill-a/.csk-install.json: read <temporary>/TestGCPreservesRuntimeOnUncertainMarksunreadable_marker2304365179/002/.agents/skills/skill-a/.csk-install.json: is a directory; repair or remove that installation\ncurator: warning: runtime sweep skipped: the live reference set could not be proven complete\ncurator: warning: build cache sweep skipped: the live reference set could not be proven complete\n"
=== RUN   TestGCPreservesRuntimeOnUncertainMarks/complete_references
    gc_test.go:92: pass 1: gc exit=0 stdout="removed runtime orphan/2222222222222222222222222222222222222222\ngc: 1 runtime entry removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 99560: invalid argument; all builds retained\n"
    gc_test.go:92: pass 2: gc exit=0 stdout="gc: 0 runtime entries removed, 0 build entries removed\n" stderr="warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default\ncurator: warning: build cache sweep skipped: live process executables could not be enumerated: read executable of process 99560: invalid argument; all builds retained\n"
--- PASS: TestGCPreservesRuntimeOnUncertainMarks (3.50s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/truncated_registry (0.80s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/invalid_marker (0.79s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/unreadable_marker (0.92s)
    --- PASS: TestGCPreservesRuntimeOnUncertainMarks/complete_references (0.98s)
PASS
ok  	github.com/relux-works/curator/cmd/curator	5.288s
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
--- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained (0.47s)
    --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/status-current (0.00s)
    --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/status-missing-snapshot (0.00s)
    --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/status-unreadable-protected-state (0.00s)
    --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/repair-reacquires-exact-source (0.00s)
    --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots (0.47s)
        --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots/artifact-receipts (0.19s)
        --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots/in-flight-journals (0.19s)
        --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots/install-markers (0.00s)
        --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots/protected-snapshots (0.01s)
        --- PASS: TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots/uncertain-entries (0.08s)
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses/corrupt_consumer_registry
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses/registry_of_an_unknown_schema
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses/registry_repeating_a_known_member
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses/project_with_only_an_invalid_marker
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses/unreadable_installed_skill_directory
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses/installed_skill_replaced_by_a_file
=== RUN   TestCollectStaysFailSafeAcrossConsecutivePasses/skill_root_replaced_by_a_file
--- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses (0.03s)
    --- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses/corrupt_consumer_registry (0.00s)
    --- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses/registry_of_an_unknown_schema (0.00s)
    --- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses/registry_repeating_a_known_member (0.00s)
    --- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses/project_with_only_an_invalid_marker (0.01s)
    --- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses/unreadable_installed_skill_directory (0.00s)
    --- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses/installed_skill_replaced_by_a_file (0.01s)
    --- PASS: TestCollectStaysFailSafeAcrossConsecutivePasses/skill_root_replaced_by_a_file (0.00s)
=== RUN   TestCollectStaysFailSafeOnRedirectedUnixScopes
=== RUN   TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_installed_skill
=== RUN   TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_global_scope_root
=== RUN   TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_hybrid_scope_root
=== RUN   TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_project_skill_root
--- PASS: TestCollectStaysFailSafeOnRedirectedUnixScopes (0.02s)
    --- PASS: TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_installed_skill (0.01s)
    --- PASS: TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_global_scope_root (0.00s)
    --- PASS: TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_hybrid_scope_root (0.01s)
    --- PASS: TestCollectStaysFailSafeOnRedirectedUnixScopes/redirected_project_skill_root (0.00s)
PASS
ok  	github.com/relux-works/curator/internal/scopes	1.322s
```
