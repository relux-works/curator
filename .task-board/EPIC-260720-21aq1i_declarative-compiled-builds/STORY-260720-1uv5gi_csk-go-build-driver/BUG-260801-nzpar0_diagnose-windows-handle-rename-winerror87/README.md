# BUG-260801-nzpar0: diagnose-windows-handle-rename-winerror87

## Description
Independently diagnose repeated native Windows WinError 87 from SetFileInformationByHandle(FileRenameInfo) in CocoaSkills PR #18 exact heads 0314ab5 and 4cd1589. Use ssh win for minimal reproductions and inspect Microsoft ABI/contracts. This is diagnosis only; do not edit or commit product code.

## Scope
Own a task-scoped diagnostic artifact covering FILE_RENAME_INFORMATION ctypes layout and padding, RootDirectory handle rights/type, source DELETE/share flags, relative-name encoding and byte length, buffer sizing, Windows Server 2025 behavior, and a concrete minimal patch recommendation. Reproduce against commit 4cd158990da1ae030d1115f9d0ca4cb3310cb60e where feasible. Do not touch the active CocoaSkills implementation worktree.

## Acceptance Criteria
Provide exact commands and native Windows outputs proving the cause or ruling out each plausible cause; cite primary Microsoft documentation; recommend the smallest identity-safe correction that preserves no-replace and held-root semantics; attach findings to the board and hand them to the active developer. No product edits, merge, tag, or release.
