# TASK-260827-1nauj3 review verdict rev3: accepted

Reviewer run, single pass, 2026-08-28. CR `CR-TASK-260827-1nauj3-3` revision 3.
Delta reviewed: `git diff 41ab53cd e5202bc0`. Rework delta: `git diff 208c9647 e5202bc0`.

Verification binary: `bin/curator`, rebuilt in this worktree with `make build`
(`go build -ldflags '-X .../version.value=dev' -o bin/curator ./cmd/curator`).
All per-command `-h` probes were run from `/tmp/curtest3`, outside the repository
tree, so no `curator init` side effect could touch the worktree. Source of truth
for flag sets: the `flag.NewFlagSet` blocks and positional handling in
`cmd/curator/main.go`.

## Verdict

Accepted. All six rev2 rework items landed, each verified against the binary
rather than against the diff. The opening claim of `docs/cli.md:3` now holds:
every documented synopsis, flag name, value type, and default matches the tree
binary.

## Rev2 findings: closure evidence

| Finding | State | Evidence |
| --- | --- | --- |
| G1 `global add --project` documented but rejected | fixed | line removed at `docs/cli.md:334`; binary still rejects, so the removal is the correct direction |
| G2 `curator add` omitted `--project` | fixed | `docs/cli.md:82` |
| G3 `curator install` omitted `--all` | fixed | `docs/cli.md:129` |
| G4 `curator upgrade` omitted `--all` | fixed | `docs/cli.md:177` |
| G5 `curator status` omitted `--all` | fixed | `docs/cli.md:207` |
| G6 README one-liner described the wrong command | fixed | `README.md:137` |
| rev2 note 4 (`Shared flags` omission) | fixed | `--all` at `docs/cli.md:9`, `--project` at `:16` |
| rev2 note 6 (`session.go:503` citation) | fixed | `docs/troubleshooting.md:243` now cites `:521`; `:224` carries the `GOROOT/bin/%s` derivation note |
| rev1 note 4 (README comment column) | fixed | the three `curator config ...` comments align at column 29 |

### Binary probes, verbatim

G2 through G5, the four flags the rework added:

```
$ ./bin/curator add -h
Usage of add:
  -branch string
    	git branch
  -git string
    	git clone URL
  -project string
    	project alias or path
  -revision string
    	git revision
  -source string
    	source directory under skills_root
  -tag string
    	git tag

$ ./bin/curator install -h
Usage of install:
  -all
    	operate on all configured projects
  -audit
    	run the audit gate in advisory or strict mode
  -build-ssh-agent string
  -build-ssh-identity string
  -build-ssh-known-hosts string
  -dry-run
  -fix-gitignore
  -strict-tags
  -verbose

$ ./bin/curator upgrade -h
Usage of install:            # upgrade routes through installFlags, same set incl. -all

$ ./bin/curator status -h
Usage of status:
  -all
    	operate on all configured projects
  -attest
    	re-check installed skills against trusted registries
  -check
    	exit non-zero unless every skill is up to date
  -json
    	machine-readable output
```

G1, the flag the rework removed, is still absent from the binary:

```
$ ./bin/curator global add -h
Usage of global add:
  -branch string
  -git string
  -revision string
  -source string
  -tag string
```

The four rev1 negative cases still behave as the document now states:

```
$ ./bin/curator global install --all
curator: global install accepts flags only
$ ./bin/curator global upgrade --all
curator: global install accepts flags only
$ ./bin/curator hybrid status --check
exit=0                      # no flag set; documented as taking no flags
$ ./bin/curator list .
                            # no positional; documented without one
```

### Remaining flag sets, re-verified this pass

Every other documented flag list matches its `-h` output exactly:
`bootstrap` (6 flags, `-default-agents` default `codex_cli`), `remove`
(`-project`), `audit` (8 flags), `skill check` (`-json`, `-locale string`),
`global status` (`-check`, `-json`), `hybrid add` (6 flags including
`-target`/`-targets`), `shell-init` (`-install`, `-no-global`),
`config build-ssh add` (3), `config build-https add` (4, `--username` default
`token` from `internal/config/buildhttps.go:40`), `config build-https login`
(`-username`).

Positional arity re-checked against source, not help text:
`cmdRemove` (`main.go:453`) reads `positional[1:]` as root args, so
`curator remove <name> [path]` is right; `cmdList` (`:973`), `cmdUI` (`:2387`),
and `cmdGC` (`:1729`) take no arguments; `cmdInit` (`:374`) takes a path;
`project add` requires `<alias> <path>`.

## Troubleshooting: citation audit

All thirteen `cmd/curator/builds.go` line references land on the named constant:
45 `buildCommandDrift`, 47 `buildContextExposed`, 50 `buildSourceDrift`,
55 `buildInputDrift`, 57 `buildUnsupportedDriver`, 61 `buildUnusableToolchain`,
64 `buildMissingArtifact`, 67 `buildCorruptReceipt`, 70 `buildArtifactDrift`,
72 `buildCorruptCache`, 75 `buildUntrustedCache`, 77 `buildUnsupportedPlatform`,
80 `buildStateChanged`.

The two Go toolchain diagnostics are exact:

```
internal/godriver/session.go:489: diagnostic("untrusted_go_executable", "CURATOR_GO must name an absolute GOROOT/bin/%s", platformGoName)
internal/godriver/session.go:521: diagnosticRemedy("toolchain_executable_mismatch", toolchainSelectionRemedy,
cmd/curator/toolchain_remedy_test.go:51-53: const want = "go-v1 toolchain_executable_mismatch: " + ... verbatim
```

Admission and credential strings:

```
internal/buildrepo/admission.go:203: "trusted Git version probe failed"
internal/buildrepo/admission.go:211: "Git release family is not operator-pinned"
internal/buildrepo/admission.go:259: "HTTPS requires a manager credential broker"
internal/buildrepo/admission.go:332: "HTTPS credential host does not match protected source"
internal/buildrepo/admission.go:337: "cannot materialize HTTPS credential broker"
internal/buildrepo/credentials.go:13: CodeSSHCredentialMissing = "build_repository_ssh_credential_missing"
internal/buildrepo/credentials.go:63: "%s is unavailable"
internal/buildrepo/credentials.go:73: "%s is not an admitted %s"
```

## Style, links, guards

No em-dash, en-dash, or guillemet anywhere in `README.md`, `docs/cli.md`, or
`docs/troubleshooting.md`. No blacklisted opener, marketing adjective, antithesis
construction, or closing summary paragraph. Reasoning stays in prose; lists carry
parallel flag enumerations only.

Every relative link in the delta resolves, checked mechanically across
`README.md`, `docs/cli.md`, `docs/troubleshooting.md`, and `docs/ci-gates.md`
(including the `../.github/...` and `../internal/...` targets the sibling
ci-gates edit added).

The README `## Commands` section sits at `README.md:112`, before
`## An open protocol`, links to `docs/cli.md`, and covers all eighteen top-level
groups the binary's usage block lists.

Documentation guards pass with no regression:

```
$ go test ./cmd/curator -run 'TestEveryCurrentnessCodeIsDocumented|TestInputCausesAreDistinctAndDocumented' -count=1 -timeout 10m
ok  	github.com/relux-works/curator/cmd/curator	0.583s
```

Working tree is clean of verification side effects: no stray `Skillfile.json`,
`git status --short` shows only the seven intended paths.

## Notes carried to the orchestrator, not blocking

1. The `# Curator` block appended to `.gitignore` (`.agents/`, `.claude/skills/`,
   `.codex/skills/`, `.cursor/rules/`, `.gemini/skills/`, `Skillfile.dev.json`)
   is a live repository decision riding inside a documentation delta. Raised in
   rev1 and rev2 and still unresolved. The orchestrator should decide whether it
   belongs in the docs commit.
2. `docs/troubleshooting.md:272`, the language adapter entry, still quotes no
   error string and cites `internal/closureexec/acquisition.go:549`, which is the
   generic `source acquisition command failed:` wrapper rather than a preflight
   check. The symptom and remedy are consistent with what that wrapper reports,
   so the entry is loose rather than false. Third cycle as non-blocking; leaving
   the ruling consistent with rev1 and rev2.
3. `docs/cli.md:673` and `:702` omit the `string` type suffix on the
   `config build-ssh` and `config build-https` flags, unlike every other section.
   The synopsis lines carry `PATH`, `SOCKET`, and `NAME` placeholders, so the
   value type reaches the reader; cosmetic only.
4. `docs/ci-gates.md` and `docs/implementation-plan.md` in this delta belong to
   sibling story tasks, not to this task's scope. Their links were checked and
   resolve; their content was not reviewed here.
