# Rework brief r7: Windows test fixtures (BUG-261021-2qv4zj, curator)

Hosted CI for PR #160 is green except "Test (windows-latest)". The branch is `fix/audit-markdown-scope-r3` at head `08fa312e`, rebased onto main `08fa2a01`. Work on **`fix/audit-markdown-scope-r3`**.

The Linux and macOS jobs, lint, Race and the platform-case gate pass. All 13 alias cases run on Windows. Two test-only defects remain.

## 1. Source-anchor test breaks on Windows checkouts

```
internal/audit TestMarkdownScopeInvalidatesRulesetSevenAllowance
  markdown_kind_test.go:145: ruleset-7 anchor not unique: "\tif data.Kind != subject.Kind {\n\t\treturn nil, false\n\t}"
```

The mutant `TestNetworkContactNarrowingMutants/markdown-reuse-ruleset-seven` fails for the same reason. The Windows runner checks out with CRLF, so a `\n` text anchor does not match. Make the anchor matching independent of line endings: normalise CRLF before searching and preserve the file's endings when writing the mutant, or use the repository's existing mutant/anchor helper if one exists. The count must still be exactly 1, so a missing anchor keeps failing loudly.

## 2. CLI context fixtures fail the Windows permission boundary

```
internal/envprofile TestProfileInstallMarkdownContextCLI/{clean,reserved,undeclared}
  curator: environment_store_untrusted: path "...\Temp\TestProfileInstallMarkdownContextCLIclean...\003" fails permissions boundary check ... DACL grants mutation rights to another identity
```

The product refusal is correct; the fixture is wrong. The fixture's store and home directories must be created owner-only on Windows, the way the other envprofile CLI tests do. Find and reuse the existing helper, and do not weaken the product check.

## Run and hand-off

On macOS, run the two tests, the mutant, and `GOOS=windows go vet ./internal/audit/... ./internal/envprofile/...`. Hosted Windows CI on the landing PR is the real check, so also say which existing helper you reused. Push signed commits to `fix/audit-markdown-scope-r3`, attach `BUG-261021-2qv4zj_results-r7.md`, hand over with the head in `audit-md-scope`, and delete scratch.
