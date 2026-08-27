Replaces #14 with a clean extraction of `d345420` onto current `main`.

The extraction preserves the multi-project dry-run binding and anti-vacuity coverage from PR 14. The only conflict resolution keeps current-main platform selection through `installPlatform()` instead of the old hard-coded `"unix"` value.

Validation:

- Baseline `origin/main`: named candidate test fails with `published dry-run scope "multi-project" has no executable binding` (exit 1, expected red).
- Extraction: the same named candidate test passes (exit 0).
- Full `make candidate-test` against the candidate conformance root passes (exit 0; 41 packages served, none deferred/excluded).
- `golangci-lint run` passes (exit 0).
- `go vet ./...` passes (exit 0).
- `go build ./...` passes (exit 0).

Board: TASK-260823-37vckg
