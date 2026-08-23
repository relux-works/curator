# TASK-260728-1skseh: run-linux-native-external-repository-qualification

## Description
Provision and use the later dedicated Linux host to run the accepted rc.5 external-build-repository suite against released Curator and csk binaries. Validate native filesystem, permissions, process, PATH, cache, offline, repair, rollback, and Git transport behavior under a clean non-root account before proposing Linux claim evidence.

## Scope
Record distribution/kernel/architecture, filesystem and shell, SSH access, non-root user, pinned Git/Go/Python toolchains, manager binaries and spec/fixture pins; execute project/global user install, activation, protected caches, network/offline and adversarial cases. Root/system installation is included only if the product explicitly supports it. No new protocol or driver design.

## Acceptance Criteria
A reproducible Linux host manifest and setup commands are attached; both released managers pass shared canonical cases and Linux-native project/global user workflows; exact-tag, inaccessible source, independent audit ordering, compiler containment, cache hit/corruption/offline reuse, status/repair/GC, shim/PATH, collision, crash/rollback, uninstall and permissions evidence passes; failures leave no unauthorized mutation; reviewer confirms results are native rather than simulated; only then is a Linux claim-v3 addendum proposed with exact pins, otherwise Linux remains unclaimed with typed gaps recorded.
