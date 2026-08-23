# TASK-260720-2g0e3b implementation evidence

## Provenance

- Worktree: /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-2g0e3b/worktree.
- Exact base: origin/main 17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8.
- Imported the accepted TASK-260720-3pwg2w product state and verified it remains byte-identical outside internal/skillspec. Board/config, temp state, diagrams, binaries, caches, and unrelated files were excluded.
- Candidate input: /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-3ag6pi/worktree/conformance/v1, manifest SHA-256 70cd274150d629f16ca04f2073a8eae5a0f3fff4584f905d807775040af21eae. It was consumed only as candidate conformance data.
- No files were committed or staged.

## Task-only delta

- internal/skillspec/types.go models schema 6, BuildRoots, Driver, and SourceDir.
- internal/skillspec/parse.go accepts schema 6, parses the closed go-v1 build command, and validates real link-free roots/source directories, root uniqueness/disjointness/use, runtime overlap, exact root containment, and nearest direct go.mod without invoking Go. Schema 1-5 and runtime fallback branches remain version-gated.
- internal/skillspec/build_test.go adds canonical/legacy mixed-command coverage and stable negative diagnostics for schema 5, runtime fallback build objects, unknown/missing drivers and source directories, mixed/extra fields, unsafe/unused/overlapping roots, escapes, links, missing go.mod, and nested modules.
- internal/skillspec/conformance_test.go consumes all 17 canonical and 17 legacy authoritative v6 schema cases and the go-build fixture through CURATOR_CONFORMANCE_ROOT, with a fake-go sentinel proving parsing does not launch it.

## Verification

- make check: PASS, including go vet ./..., go test ./..., and repository gofmt gate.
- go test -race ./...: PASS.
- go build ./...: PASS.
- Focused rc.4 v6 schema cases and skill-manifest-resolution vectors: PASS.
- Parser stress, 25 runs: PASS.
- internal/skillspec coverage: 89.9 percent.
- Windows amd64 and Linux amd64 internal/skillspec test binaries compile: PASS.
- git diff --check, task-scope gofmt, no os/exec or exec.Command in internal/skillspec, and predecessor byte comparison: PASS.

## Integration anomaly

A full repository test with CURATOR_CONFORMANCE_ROOT set to the rc.4 candidate reaches an unrelated existing mismatch in internal/interop TestManagerLifecycleVectors because that consumer does not yet model the new compiled-cache dry-run vector. The parser-owned v6 schema and manifest-resolution consumers pass, baseline full tests pass, and this task does not own manager lifecycle consumption.