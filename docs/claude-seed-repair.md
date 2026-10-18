# Claude provisioning seed and tool state

Environments §7.4 defines `.claude.json` as a provisioning seed owned by
Claude after the first run. Repair owns only the launch directory's
`projects` entry: `hasTrustDialogAccepted`, plus
`hasClaudeMdExternalIncludesApproved` for the referenced form.

An absent file retains the original CCJ-1 seed encoding and trailing LF.
For an existing file, repair reads ordinary JSON with `UseNumber` and
preserves number text, including fractions, exponents and integers beyond
the floating-point precision range. It preserves unknown values and merges
only the required flags into the launch entry. As before, a missing or
non-object `projects` member becomes an object, and a non-object launch
entry becomes an object. A satisfied entry leaves the file bytes, inode
and mtime unchanged even if other managed surfaces need repair.

A necessary merge may reorder keys, compact whitespace and normalize
string escapes through the standard JSON encoder. Number lexemes and
decoded values are preserved. This is a semantic merge, not a lossless
JSON text editor. Verification uses the same number-preserving reader.
Existing invalid-object, symlink, non-regular and unreadable-file refusals
remain in force.

Run the regression and overlay plant controls with:

```sh
go test ./internal/envprofile -count=1 -v -run 'TestClaude(Repair|Absent|SeedNarrowing)'
```

`claude_seed_test.go` drives `Resolve` with temporary homes for both forms.
`claude_seed_mutants_test.go` uses Go source overlays, without editing the
candidate, to require a named behavioral failure and exit code 1 for each
plant. Controls cover whole-object CCJ-1 encoding, number normalization,
incorrect flag satisfaction, rewriting satisfied state and admitting a
null object. Test output can be captured under `.temp/`.
