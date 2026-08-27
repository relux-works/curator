# TASK-260728-1g0z69 reviewer verdict — cycle 4

Verdict: changes requested. Route: `analysis`.

The cycle-3 rework closes the four prior ordering and payload gaps: Stage A declaration presence and origin classification read disjoint inputs; host applicability precedes relpath and probe construction; Stage B rejects both empty and non-empty late descriptor narrowing before cache or compiler work; repeated and malformed Go files have a typed file-shape outcome. The submitted static and release gates are green. The Go metadata classifier remains non-canonical and not implementation-ready.

## Blocking findings

1. **`GoVersionRE` is not the complete semantic Go version language.** The reference at lines 611-619 correctly says `cmd/go/internal/gover.Parse` rejects a prerelease suffix after an explicit patch, but applies that semantic parser only to `toolchain` names. Lines 663-676 and vector 122 instead permit `go 1.23.4rc1` solely because `modfile.GoVersionRE` matches it. Current upstream behavior disproves that outcome: both Go 1.25.1 and Go 1.25.5 accept the line at the modfile-regex layer, then `GOTOOLCHAIN=local go mod tidy` and `go list -mod=mod ./...` panic with `go: internal error: missing go root module`. The upstream semantic parser explicitly returns invalid when a prerelease follows a patch. Thus package metadata that the Go command cannot represent passes Stage B and reaches compiler-side work, reopening the exact cycle-3 blocker. Define the `go`-directive classifier over the complete upstream acceptance pipeline, not `GoVersionRE` alone; classify patch-prerelease values with a stable typed pre-compiler outcome; replace vector 122; and add an executable probe covering the regex-versus-semantic-parser boundary. Primary sources: https://go.dev/src/cmd/vendor/golang.org/x/mod/modfile/rule.go and https://go.dev/src/internal/gover/gover.go.

2. **Vector 126 states an impossible equality.** It requires the set classified as `compared` to equal the set upstream accepts for both directives. Upstream explicitly accepts non-standard names such as `go1.21.0-custom`, while this contract intentionally classifies custom-distribution names as `forbidden` package influence. Both choices can be sound, but the sets cannot be equal. Scope the no-widening property to values remaining after the security-forbidden partition, or express separate grammar-completeness and selector-exclusion properties. The official toolchain-name contract is https://go.dev/doc/toolchain#name.

## Independent evidence

- Candidate versus accepted predecessor differs only in `CHANGELOG.md`, decision 0007, and the toolchain reference. Board decision/reference resources are byte-identical to those files.
- `tools/validate.py`: exit 0, 42 schemas and 422 vector files.
- Python unit suite: exit 0, 29 tests.
- `go test ./tools/...`, `go vet ./tools/...`, `gofmt -l tools`, and `git diff --check`: exit 0; formatter and diff checks produced no output.
- Exact clean probe `41e8250d87fa43c928fa17e8658ee85d18e2631d` is byte-identical to the candidate. `make regenerate-check`: exit 0. rc.5 release gate: exit 0.
- `conformance/v1` and `release/1.0.0-rc.5.json` remain byte-identical to the accepted predecessor.
- Adversarial runtime probe: Go 1.25.1 and 1.25.5 both panic on the submitted permitted `go 1.23.4rc1` case before a normal package result.

No task-worktree source, schema, vector, release artifact, stage, commit, publication, pin, or platform claim was changed by this review.