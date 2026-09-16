# TASK-260910-24cuys review verdict, CR-TASK-260910-24cuys-1 rev 1 — ACCEPTED

Reviewer run RUN-260915-8abb7f (claude-fable-5-1). Shell: zsh, `set -o pipefail`, worktree `.temp/STORY-260910-197y84/worktree`.

## Candidate identity
- Base 4f27ccb2, candidate tree ab04d5366fc7b62e6b7d1e38a2e8121701d0d4b0.
- Independently re-derived working-tree OID (HEAD index + `git add -A internal`, `git write-tree`) = ab04d536… (exact match).
- Fixtures byte-identical to curator-spec main (3535d63 / a4fcaf0): 41/41 schema cases, skillfile-v2.schema.json, v1 common + skillfile-v1 schemas (`diff -r` clean).

## Independent reruns (all rc=0)
- `go build ./...`, `go vet ./internal/manifest ./internal/identity`, `golangci-lint run` on both packages: 0 issues.
- `go test -count=1 ./internal/manifest ./internal/identity ./internal/protocoljson ./internal/skillspec`: ok.
- `-run TestDraft -v`: 7/7 top-level tests pass (41 published fixture subtests via `manifest.LoadWithOptions`, the production read entry point).

## Contract conformance (skillfile-sources.md §1, §5; repository-transport.md)
- Schema 2 admitted only via reader-owned `ParseOptions{DraftSourcesV1}`; `Load`/`ParseBytes`/`Parse` still reject v2 (tested). Unsupported versions 0/3/2.5/"2"/null rejected.
- `sources` union: path (no expansion, no other fields) | git (closed 6.3 endpoint grammar via buildrepo.ParseSource) | repository (must equal its own canonical identity: no uppercase host, no .git, no creds/ports) + exactly one tag/branch/revision. Revision = 40/64 lowercase hex; tag/branch = git ref-name grammar with 255-rune bound.
- Selector arms: legacy / individual / collection are mutually exclusive; unknown alias -> `source_alias_unknown`; duplicate names -> `source_name_conflict` (also across legacy+selector); directory "." or portable contained path with glob chars rejected; include non-empty unique literals or `*`, exclude literals only.
- `sources` rejected on schema 1; every v1 field byte-for-byte preserved on v2 (DeepEqual regression incl. legacy `source` = configured-root meaning).
- All checks happen before any filesystem/network I/O (tests use /nonexistent paths).

## Additional probes (17 extra malformed payloads, all rejected)
v1 entry with `from`; selector missing directory; selector+source; selector+tag; non-string from/directory/name/path/git; include `[5]`, `**`, `a/b`; exclude `*`; NUL in path; uppercase revision; missing skills; legacy/selector name clash.

## Mutant attack (narrowing, 18 mutants, gates in manifest.go / sources.go / draft_sources.go)
Killed 16/18: opt-in bypass, git+repository both, non-canonical repository, multiple refs, directory glob chars, unknown alias, individual-selector mixed arms, empty include, wildcard exclude, duplicate members, ref `..`, ref length >255, path+ref mix, duplicate selector name, extra top-level field, `sources` on schema 1.
Survived 2/18 (committed-test gaps only; production rejects both, confirmed by direct probe):
- M5: uppercase hex revision admitted by validSourceRef.
- M17: control characters in `path` admitted (containsControl dropped).
These are non-blocking coverage bounds for a follow-up test addition, not defects.

## Bounds
- Parsing only. No production caller passes `DraftSourcesV1` yet; resolution, containment, enumeration, locking and install are later tasks per the contract and README.
- No JSON Schema engine executes the local schema copies; filename labels are the oracle (documented).
- Landing suite left to the runtime per campaign rules; not rerun manually.
