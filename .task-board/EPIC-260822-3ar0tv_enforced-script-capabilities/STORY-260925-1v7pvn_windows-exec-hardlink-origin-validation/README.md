# STORY-260925-1v7pvn: windows-exec-hardlink-origin-validation

## Description
Own the two remaining curator failures in the dcc7f015 executable identity family: windows-exec-noncomponent-store-hardlinks and windows-exec-unowned-file-hardlinks. The production resolver still accepts both. Prior STORY-260822-2h0v9j and STORY-260924-2tyzhh are done; BUG-260924-5p8b0z covers only uncaptured SystemRoot and now passes.

## Scope
Curator internal/scriptworker Windows System32 executable hard-link resolution and production-entry conformance coverage.

## Acceptance Criteria
At curator-spec dcc7f015, the production resolver rejects non-component-store and unowned-file hard-link cases while preserving the platform-owned component-store exception and uncaptured-SystemRoot rejection. Tests reach the production resolver and count all eight published cases. If platform APIs cannot prove link origin and file ownership, record the constraint and options before changing behavior.
