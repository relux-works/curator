# TASK-260729-1bf72u: qualify-csk-macos-windows-runner-readiness

## Description
Inventory and qualify the current macOS primary and ssh win runner prerequisites for CocoaSkills Go parity, including Python, approved Go, filesystem/security primitives, shell behavior, temp paths, and exact CI commands.

## Scope
Read-only local and remote inspection. No downloads, package installation, registry/system mutation, source edits, service changes, or destructive cleanup.

## Acceptance Criteria
Outcome records exact host/tool versions and command exits, identifies ready and missing prerequisites, provides an operator-safe Go setup recommendation for Windows if absent, defines process/disk/temp cleanup barriers, and supplies the native macOS/Windows validation matrix.
