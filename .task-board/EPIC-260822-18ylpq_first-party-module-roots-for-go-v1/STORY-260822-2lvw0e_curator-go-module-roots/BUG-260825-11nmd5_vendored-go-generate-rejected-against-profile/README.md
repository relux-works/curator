# BUG-260825-11nmd5: vendored-go-generate-rejected-against-profile

## Description
go-v1 rejects any active Go file containing //go:generate, including audited third-party vendored code. curator-spec profiles/manager.md is normative that the directive is inert and that its presence in vendored GoFiles does not fail preflight, and decision 0005-vendored-go-boundary-relaxation names clipperhouse/displaywidth and skill-project-management as the motivating case. Commit 9ba552f Restore Go generator rejection dropped the vendor carve-out from internal/godriver/graph.go. Discovered while adopting schema-8 module roots in skill-project-management (its board TASK-260822-hje0ya): task-board-tui cannot build because bubbletea to charmbracelet/x/ansi to clipperhouse/displaywidth ships a bare //go:generate in gen.go, and no released version drops it.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
validatePackageGraph exempts //go:generate only for packages below the build root vendor tree; build-root and first-party packages are still rejected with go_generator_forbidden; the cgo_import_dynamic allowlist is unchanged; conformance suite against SPEC_PIN stays green.
