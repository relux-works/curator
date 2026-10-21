# Rework brief r6: hosted CI gates (BUG-261021-2qv4zj, curator)

Review r5 accepted the change. The orchestrator then rebased it onto current main, patch-identical, as branch `fix/audit-markdown-scope-r2` at head `04fab6de`. Hosted CI on PR #159 (run 39600035060) fails two repository gates that the local runs did not exercise. Work on **`fix/audit-markdown-scope-r2`** (fetch it, and base your commits on `04fab6de`), not on the old branch.

## 1. Lint (golangci-lint, job "Lint")

```
internal/audit/markdown_alias_test.go:25:17: Error return value of `os.Remove` is not checked (errcheck)
internal/envprofile/credential_projection_test.go:203:18: Error return value of `op.close` is not checked (errcheck)
internal/envprofile/skill_projection.go:95:20: Error return value of `os.RemoveAll` is not checked (errcheck)
internal/envprofile/skill_projection.go:170:20: Error return value of `os.RemoveAll` is not checked (errcheck)
internal/envprofile/skill_projection_repair_test.go:315:16: Error return value of `op.close` is not checked (errcheck)
internal/envprofile/skill_projection_test.go:519:16: Error return value of `op.close` is not checked (errcheck)
internal/envprofile/skill_projection.go:46:16: G304: Potential file inclusion via variable (gosec)
```

Fix each one the way the surrounding code does. Handle or join the errors in product code. For G304, follow the repository's existing pattern for a path the program constructs itself (look at how other `os.ReadFile`/`Open` calls on store paths are handled); do not suppress it blanket. Run the repository's lint configuration locally (`golangci-lint run` with the repo's config, version as in `.github/workflows`).

## 2. Platform-case skip gate (Linux "Test" and "Race")

`.github/ci/platform-cases.tsv` and `.github/ci/skip-classes.tsv` require every skip to carry a recognised reason class. Your case-alias tests skip on case-sensitive Linux with unregistered reasons:
- "ruleset-6 alias allowance requires a case-insensitive filesystem";
- "alias control requires a case-insensitive filesystem";
- the `TestProfileInstallMarkdownCaseAliasesCLI/*` cases.

There are 13 failing entries in total (see the run log). Register them correctly: either reuse an existing class for "requires a case-insensitive filesystem", or add one with the platforms where it is expected. Do not simply widen the gate. Where a case can run meaningfully on Linux, for example the case-sensitive counterpart asserting no alias publication, make it run instead of skipping.

## Run and hand-off

Run lint, the gate script the CI job uses (see the workflow for the exact command; run it locally with GOOS matching where possible), build, vet, and the affected packages. A full suite is not required again unless product code beyond the errcheck handling changes. Push signed commits to `fix/audit-markdown-scope-r2`, attach `BUG-261021-2qv4zj_results-r6.md`, hand over in `audit-md-scope` with the head, and delete scratch.
