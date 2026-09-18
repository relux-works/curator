# TASK-260917-2ecpjv: qualify-curator-spec-rc12

## Description
Read-only qualification of the curator-spec v1.0.0-rc.12 release: resolve the annotated signed tag to its full commit (expected dced9b8317e0e8af79edf2d0539b32bd22b6c85b), verify the tag signature against the maintainer key, run make validate and the regeneration check at that commit in a disposable checkout, record the conformance manifest digest and the release/1.0.0-rc.9.json pin content, and confirm the vector families the wave-1 manager tasks consume (manager-config-v2, environments, umbrella-provider-resolution, environments-env-passthrough, shell-hook-trust, registry-client page_boundary_cases) are published there.

## Scope
(define task scope)

## Acceptance Criteria
Outcome resource records tag, full commit, tagger identity and signature verification result, make validate exit 0 transcript at the tag, manifest sha256, and the presence of every named vector family; no repository mutated
