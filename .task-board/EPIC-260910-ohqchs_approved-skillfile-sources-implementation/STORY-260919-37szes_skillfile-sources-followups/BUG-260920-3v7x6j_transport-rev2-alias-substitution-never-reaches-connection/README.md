# BUG-260920-3v7x6j: transport-rev2-alias-substitution-never-reaches-connection

## Description
From TASK-260910-1xya7x review rev3 (corpus case v2-alias-resolution): config.ResolveRepositoryEndpoints carries ResolvedHost/ResolvedPort but the executor (internal/buildrepo/transport.go:502, 931) fetches sources[index] parsed from attempt.URL and only records the resolved host as provenance; CLI resolve clones the listed URL. The transport-revision-2 alias feature never changes the connection in production (touches STORY-260916-v58b5y's accepted claims).

## Scope
internal/buildrepo transport executor connection target, cmd/curator resolve clone target

## Acceptance Criteria
Alias substitution changes the actual connection target (host:port) while the declared identity and recorded provenance stay as specified; corpus case v2-alias-resolution flips to driven-pass; provenance still sanitized; narrowing mutant (connect to the declared host) killed.
