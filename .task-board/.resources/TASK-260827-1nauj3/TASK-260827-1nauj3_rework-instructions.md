# TASK-260827-1nauj3 review verdict: changes requested

Reviewer run, single pass, 2026-08-28. CR `CR-TASK-260827-1nauj3-1` revision 1.
Delta reviewed: `git diff 41ab53cd 8b9f91b3`.

Verification binary: `bin/curator`, rebuilt in this worktree with `make build`
(`go build -ldflags '-X .../version.value=dev' -o bin/curator ./cmd/curator`).
Source of truth for flag sets: `cmd/curator/main.go`.

## Verdict

Changes requested. The acceptance criterion "every synopsis and flag matches the
tree binary help verbatim" does not hold: six command entries in `docs/cli.md`
document flags and positionals the binary does not accept, and one
`docs/troubleshooting.md` entry maps two Git admission errors to a Go toolchain
cause and a remedy that cannot fix either failure.

`docs/cli.md` opens with "Every synopsis and flag in this reference was verified
verbatim against `./bin/curator` built via `make build`". That claim is false as
written, which is the reason this is a rework rather than a note.

## Blocking findings

### F1. `curator project add` is documented as a different command

`docs/cli.md:246` documents `curator project add <name> [options]` with
`--branch`, `--git`, `--project`, `--revision`, `--source`, `--tag`.

Real definition, `cmd/curator/main.go:1002`:

```
flags := flag.NewFlagSet("project add", flag.ContinueOnError)
agentsRaw := flags.String("agents", "", "comma-separated target agents")
positional, err := parseInterspersed(flags, args[1:])
if err != nil || len(positional) != 2 {
        fmt.Fprintln(os.Stderr, "curator: project add requires <alias> <path>")
```

The documented example run verbatim:

```
$ ./bin/curator project add helper --git https://github.com/example/helper.git
flag provided but not defined: -git
Usage of project add:
  -agents string
    	comma-separated target agents
curator: project add requires <alias> <path>
exit=2
```

Every documented flag is rejected, the real flag `--agents` is absent from the
document, and the arity is `<alias> <path>`, not `<name>`.

### F2. `curator global add --project` does not exist

`docs/cli.md:341` lists `--project string: project alias or path.`
`cmd/curator/main.go:1107` defines only `git`, `tag`, `revision`, `branch`,
`source`.

```
$ ./bin/curator global add helper --project someproj --git https://github.com/example/helper.git
flag provided but not defined: -project
Usage of global add:
  -branch string
```

### F3. `curator hybrid status --check` and `--json` do not exist, and the
document sells `--check` as a drift gate

`docs/cli.md:545` documents:

```
Flags:
- `--check`: exit non-zero unless every skill is up to date.
- `--json`: machine-readable output.

Check hybrid state status:
    curator hybrid status --check
```

`cmd/curator/main.go:1546` has no flag set for this subcommand. It loads the
hybrid declarations, prints one line per entry, and returns `exitOK`
unconditionally; `args[1:]` is never read.

```
$ ./bin/curator hybrid status --check
exit=0
```

This is the worst of the six: an operator who follows the document puts
`curator hybrid status --check` in CI and gets a gate that always passes. A
documented fail-closed check that is silently fail-open is a negative-evidence
failure, not a typo.

### F4. `--all` is documented on `curator global install` and `curator global upgrade`, which reject it

`docs/cli.md:417` and `docs/cli.md:461` both list
`--all: operate on all configured projects.`

`cmd/curator/main.go:1305`:

```
opts, positional, all, auditMode, err := installFlags(args)
if err != nil || len(positional) != 0 || all {
        if err == nil {
                err = fmt.Errorf("global install accepts flags only")
        }
```

```
$ ./bin/curator global install --all
curator: global install accepts flags only
```

The flag parses (the global path reuses `installFlags`) and is then an explicit
usage error. Listing it as a supported flag of the global scope is wrong.

### F5. `curator skill check --locale` type and default are invented

`docs/cli.md:298` documents
"`--locale code`: language locale for validation messages (default: `en`)."

`cmd/curator/main.go:1063`: `localeValue := flags.String("locale", "", "validate against a locale")`.
The flag is `-locale string`, its default is empty, and
`internal/skillcheck/skillcheck.go:28` records that locale consistency is checked
"when a locale is given". There is no `en` default, and `code` is not the flag's
value type.

### F6. `curator list [path]` and `curator ui [path]` take no positional argument

`docs/cli.md:226` documents `curator list [path]` with the example `curator list .`
and "List declared skills in a project".
`docs/cli.md:642` documents `curator ui [path]` with the example `curator ui .`.

`cmd/curator/main.go:973` is `func cmdList() int` and
`cmd/curator/main.go:2387` is `func cmdUI() int`. Neither takes arguments;
`main.go:143` calls `cmdUI()` with the arguments dropped. `curator list` always
reports every configured project, so the documented example does not do what the
surrounding sentence says it does.

### F7. The "Missing or untrusted Go toolchain" troubleshooting entry is wrong

`docs/troubleshooting.md:222` reads:

```
Symptom: error `trusted Git version probe failed` or `Git release family is not operator-pinned`.
Cause: installed Git or Go toolchain binaries fail security probes or version family constraints.
Source location: internal/buildrepo/admission.go:203 and internal/buildrepo/admission.go:211.
Remedy: install an operator-pinned Go toolchain release family and set CURATOR_GO.
```

Both cited lines are inside `ValidateGitTool` (`internal/buildrepo/admission.go:180`).
Line 203 fires when `git --version` fails or answers with more than 256 bytes;
line 211 fires when that version string matches no entry in `tool.AllowedVersions`.
Neither involves the Go toolchain, and setting `CURATOR_GO` cannot resolve
either. The heading, the cause, and the remedy all contradict the evidence they
cite.

The task named toolchain preflight mismatches as one of the highest-value
failures to cover, and the real ones are absent from the document:
`untrusted_go_executable` (`internal/godriver/session.go:489`,
"CURATOR_GO must name an absolute GOROOT/bin/go") and
`toolchain_executable_mismatch` with its shipped remedy, asserted verbatim by
`cmd/curator/toolchain_remedy_test.go:51`:

```
go-v1 toolchain_executable_mismatch: selected Go executable is not the regular
executable under the derived GOROOT; put the real GOROOT/bin first on PATH,
e.g. PATH="$(go env GOROOT)/bin:$PATH"
```

That is the entry an operator hitting a goenv/asdf/mise wrapper needs, and it is
the section this document promised.

## Non-blocking notes for the rework

1. Three SSH messages in `docs/troubleshooting.md:277` do not grep as literals
   because they are format-string derived: `internal/buildrepo/credentials.go:63`
   emits `"%s is unavailable"` over the labels `SSH identity`, `SSH agent socket`,
   `SSH known hosts`. The rendered text is correct and the line reference is
   correct; no change needed beyond noting the derivation. The cause prose is
   imprecise though: "permissions are invalid" does not produce that message. A
   wrong file mode reaches `credentials.go:73`, `"%s is not an admitted %s"`.
2. `internal/closureexec/acquisition.go:549` is the generic
   `source acquisition command failed:` wrapper, not a preflight check. The
   "Language source-closure adapter preflight checks" entry cites it while
   quoting no error string at all, so it carries no verifiable symptom.
3. The delta carries two artifacts that are side effects of running the binary
   inside the repository during verification: the new empty `Skillfile.json` and
   the `# Curator` block appended to `.gitignore`. Reproduced here:
   `./bin/curator init -h` writes `Skillfile.json` into the working directory
   (`init` has no flag set, so `-h` is treated as a path argument and the command
   runs). The `.gitignore` block ignores `.claude/skills/` and `.codex/skills/`,
   which is a live decision for this repository rather than a documentation
   change. The orchestrator should decide whether both belong in the docs commit.
4. `README.md:207` misaligns the `curator config build-https` comment column by
   one space against its two siblings.

## What is correct

- Every `cmd/curator/builds.go` line reference in `docs/troubleshooting.md` is
  accurate: 45, 47, 50, 55, 57, 61, 64, 67, 70, 72, 75, 77, 80 each land on the
  named constant.
- The six admission and credential identifiers that are literal strings all grep
  in `internal/`: `trusted Git version probe failed` (admission.go:203),
  `Git release family is not operator-pinned` (211),
  `build_repository_ssh_credential_missing` (credentials.go:13),
  `HTTPS credential host does not match protected source` (admission.go:332),
  `HTTPS requires a manager credential broker` (259),
  `cannot materialize HTTPS credential broker` (337).
- The `unusable-build-toolchain` remedy is right: `internal/godriver/session.go:53`
  admits exactly `CURATOR_GO` and `GOROOT`, in that order.
- The `curator gc` 24-hour grace claim matches
  `internal/buildcache/collect.go:26`, `DefaultGrace = 24 * time.Hour`.
- `curator add`, `curator remove`, `curator install`, `curator upgrade`,
  `curator status`, `curator hybrid add`, `curator audit`, `curator shell-init`,
  `curator config build-ssh add`, and `curator config build-https add|login`
  flag sets match the source exactly, including the `--username` default `token`
  (`internal/config/buildhttps.go:40`).
- The README `## Commands` section covers all eighteen top-level groups the
  binary's usage block lists, and links to `docs/cli.md`. Every link in the delta
  resolves: `docs/cli.md`, `docs/troubleshooting.md`, `docs/compiled-commands.md`.
- Style guide holds. No em-dash or en-dash, no guillemets, no blacklisted opener
  or marketing adjective, no closing summary paragraph, reasoning in prose and
  lists reserved for parallel flag enumerations.
- No regression in the README documentation guards:

```
$ go test ./cmd/curator -run 'TestEveryCurrentnessCodeIsDocumented|TestInputCausesAreDistinctAndDocumented' -count=1 -timeout 10m
ok  	github.com/relux-works/curator/cmd/curator	0.700s
```

## Rework instruction

Fix F1 through F7 against the source, not against memory. For every command
group, read the `flag.NewFlagSet` block in `cmd/curator/main.go` and the
positional handling that follows it, and copy the flag names, value types, and
defaults from there. Where a subcommand has no flag set (`hybrid status`,
`hybrid list`, `global list`, `global update`, `list`, `ui`, `project resolve`),
state that it takes no flags rather than inventing a plausible pair. Then replace
the Go-toolchain troubleshooting entry with the two real toolchain diagnostics
and give the Git admission errors their own entry with a Git remedy.
