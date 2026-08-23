# EPIC-260822-18ylpq: first-party-module-roots-for-go-v1

## Description
Adoption feature: let go-v1 build multi-module repositories with lockstep first-party local modules. Today core.md 4.2 fixes one module per build root, vendor-only versioned dependency resolution, and the driver rejects Module.Replace outright — third-party tool repos that use local replace directives cannot be packaged without restructuring. Proposal shape: an explicit declared surface (e.g. modules list on the build command) validated one-to-one against go.mod replace directives — directory-form only, strictly inside the snapshot, no module-to-module redirects; declared module dirs are hashed and scanned exactly like the main module. Snapshot-wide curator-build-source-v1 identity (8.1) already keeps cache keys sound. Origin analysis on skill-project-management board TASK-260822-1gs27d.

## Scope
(define epic scope)

## Acceptance Criteria
Spec decision accepted; core.md 4.2 prose, schema bump, and positive/negative conformance vectors (incl. Windows path honesty) released; curator and cocoaskills implement and pass cross-implementation CI; skill-project-management switches its manifest from system back to go-v1 build with declared module roots without repo restructuring.
