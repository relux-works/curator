# TASK-260924-2v4v2m: close-marker-v5-local-go-v1-raw-shape-and-tag-validation

## Description
In marker.Read v5 local arm reject nullable external-only members (repository, substituted, substitution present as null) and validate declared_tag with the canonical Git ref-name validator; schema versions 1-4 unchanged (gap matrix TASK-260924-3re9jo; BUG-260920-2eg8nv review section 8). Fail-closed reading; the matching spec clarification runs in parallel.

## Scope
(define task scope)

## Acceptance Criteria
production-entry marker.Read rejects repository:null, substituted:null, substitution:null, empty and malformed declared_tag; valid local v5 and accepted external records still parse; v1-v4 controls green; mutants killed
