# Document Curator compiled-build authoring

## Description
Update Curator-facing documentation for schema v6 compiled commands without duplicating the normative protocol or claiming unlanded releases.

## Scope
Own README.md and the Curator repository authoring or implementation documentation that already explains managed commands and CI. Include one complete mixed script and build manifest example, build_roots context exclusion, vendor-only native go-v1 prerequisites, explicit trusted toolchain selection precedence including CURATOR_GO, cache and marker currentness, dry-run outcomes, install and upgrade repair, locked GC, and Unix or Windows shim invocation. Explain that the output remains untrusted, is never run during install, and that hooks, package argv or environment, cgo, workspaces, downloads, external linking, root modules, and future generic drivers are unsupported. Link the authoritative rc.4 protocol docs rather than copying all vectors. Do not edit protocol-spec or claim release acceptance before real pins exist.

## Acceptance Criteria
A skill author can create a valid schema 6 package and understand every prerequisite and failure class without reading source; operator docs state how Curator selects Go without using user PATH, where implementation-specific cache state conceptually lives, and how dry-run, status, install or upgrade repair, and gc behave; docs preserve schema 1 through 5 guidance and distinguish portable logical identity from Curator local paths; tool and development sections list exact verification commands and artifact locations; all links resolve and documentation examples are exercised by tests or JSON parsing.
