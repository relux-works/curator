# TASK-260827-1nauj3 review verdict rev2: changes requested

Reviewer run, single pass, 2026-08-28. CR `CR-TASK-260827-1nauj3-2` revision 2.
Delta reviewed: `git diff 41ab53cd 208c9647`.

Verification binary: `bin/curator`, rebuilt in this worktree with `make build`
(`go build -ldflags '-X .../version.value=dev' -o bin/curator ./cmd/curator`).
Source of truth for flag sets: the fifteen `flag.NewFlagSet` blocks in
`cmd/curator/main.go`, cross-checked against per-command `-h` output run from an
empty scratch directory (`/tmp/curtest`) so no repository file was touched.

## Verdict

Changes requested. Five of the seven rev1 blocking findings are fixed. F2 is not
fixed at all, and the flag-set sweep this rework was supposed to perform surfaced
four more real flags the document omits. The opening claim of `docs/cli.md:3`,
"Every synopsis and flag in this reference was verified verbatim against
`./bin/curator` built via `make build`", still does not hold.

## Rev1 findings: what landed

| Finding | State | Evidence |
| --- | --- | --- |
| F1 `project add` wrong command | fixed | `docs/cli.md:241` is now `curator project add <alias> <path> [flags]` with only `--agents string` |
| F2 `global add --project` | **NOT fixed** | see G1 |
| F3 `hybrid status --check/--json` | fixed | `docs/cli.md:533` documents `curator hybrid status` with no flags |
| F4 `--all` on global install/upgrade | fixed | both flag lists now stop at `--verbose` |
| F5 `skill check --locale` type/default | fixed | `docs/cli.md:288` is `--locale string`: validate against a locale |
| F6 `list`/`ui` positionals | fixed | `docs/cli.md:221` and `:625` take no argument |
| F7 Go-toolchain troubleshooting entry | fixed | `docs/troubleshooting.md:222` and `:239` are the two real Go diagnostics; `:256` is a separate Git admission entry with a Git remedy |

## Blocking findings

### G1. `curator global add --project` is still documented and still rejected

This is rev1 F2 verbatim, unchanged. `docs/cli.md:331` still reads
`- `--project string`: project alias or path.`

`cmd/curator/main.go:1107` defines `git`, `tag`, `revision`, `branch`, `source`
and nothing else:

```
$ ./bin/curator global add helper --project someproj --git https://github.com/example/helper.git
flag provided but not defined: -project
Usage of global add:
  -branch string
    	git branch
  -git string
    	git clone URL
  -revision string
    	git revision
  -source string
    	source directory under skills_root
  -tag string
    	git tag
```

An explicit blocking finding carried through a full rework cycle untouched is the
reason this is a second rework rather than a note.

### G2. `curator add` omits its real `--project` flag

`docs/cli.md:76` lists `--branch`, `--git`, `--revision`, `--source`, `--tag`.
`cmd/curator/main.go:402` also defines `-project string`:

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
```

The flag is live, not vestigial. Run from an unrelated directory it selects the
target project:

```
$ /tmp/curtest $ curator add helper --project /tmp/curtest --git https://example.com/x.git --tag v1
curator: Skillfile.json not found at /tmp/curtest/Skillfile.json; run 'curator init' first
```

`docs/cli.md:104` already documents the same flag on `curator remove`, so the
omission is inconsistent within the document.

### G3. `curator install` omits its real `--all` flag

`docs/cli.md:124` lists eight flags and stops. `installFlags`
(`cmd/curator/main.go:483`) opens with
`all := flags.Bool("all", false, "operate on all configured projects")`:

```
$ ./bin/curator install -h
Usage of install:
  -all
    	operate on all configured projects
  -audit
...
$ ./bin/curator install --all --dry-run
curator: --all requested but no projects are configured
```

The flag parses and drives a distinct code path; the message above is the
multi-project branch reporting an empty machine configuration, not a usage
rejection.

### G4. `curator upgrade` omits its real `--all` flag

`docs/cli.md:171`. `curator upgrade` routes through the same `installFlags`
(`cmd/curator/main.go:449`, `cmdInstallMode(args, true)`), so its help is the
same `Usage of install:` block carrying `-all`.

Rev1 F4 asked for `--all` to be removed from the two global entries, where the
binary refuses it. The rework removed it from the project-scope entries as well,
where the binary accepts it. That inverted the finding.

### G5. `curator status` omits its real `--all` flag

`docs/cli.md:200` lists `--attest`, `--check`, `--json`.
`cmd/curator/main.go:629`:

```
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
$ ./bin/curator status --all
curator: --all requested but no projects are configured
```

`curator status --all --check` is the natural machine-wide drift gate, and the
document gives an operator no way to find it.

### G6. The README one-liner for `curator project add` describes the wrong command

`README.md:137`:

```
curator project add      # add or replace a skill declaration in a project manifest
```

`cmd/curator/main.go:1002` requires `<alias> <path>` and registers a project
mapping in the machine configuration; `curator add` is the command that adds a
skill declaration. `docs/cli.md:236` states this correctly, so the README
contradicts the reference it links to, and it repeats the exact confusion rev1 F1
was raised about.

## Non-blocking notes for the rework

1. `docs/troubleshooting.md:243` cites `internal/godriver/session.go:503` for
   `toolchain_executable_mismatch`. Line 503 is inside a comment. The two real
   emission sites are `session.go:521` (`diagnosticRemedy`) and `session.go:669`
   (`diagnosticErrRemedy`). The companion citation
   `cmd/curator/toolchain_remedy_test.go:51` is correct: that file exists and
   lines 51 to 53 assert the quoted string verbatim.
2. `docs/troubleshooting.md:224` quotes `CURATOR_GO must name an absolute
   GOROOT/bin/go`, which does not grep as a literal:
   `session.go:489` emits `"CURATOR_GO must name an absolute GOROOT/bin/%s"` over
   `platformGoName`. The rendered text is right on darwin and linux. Add the same
   one-line derivation note the SSH entry deserves (rev1 note 1).
3. `docs/troubleshooting.md:272`, the language adapter entry, still quotes no
   error string and still cites `internal/closureexec/acquisition.go:549`, which
   is the generic `source acquisition command failed:` wrapper rather than a
   preflight check. Repeat of rev1 note 2; the entry carries no verifiable
   symptom.
4. The `## Shared flags` section (`docs/cli.md:9`) omits the two flags that are
   genuinely shared across groups: `--all` (install, upgrade, status, audit) and
   `--project` (add, remove). Adding them there is the cheapest way to keep
   G2 through G5 from recurring.
5. The `# Curator` block appended to `.gitignore` (`.agents/`, `.claude/skills/`,
   `.codex/skills/`, `.cursor/rules/`, `.gemini/skills/`, `Skillfile.dev.json`)
   is a live repository decision riding inside a documentation delta. The stray
   `Skillfile.json` from rev1 is gone. The orchestrator should decide whether the
   gitignore block belongs in the docs commit.

## What is correct

- All thirteen `cmd/curator/builds.go` line references in
  `docs/troubleshooting.md` land on the named constant: 45, 47, 50, 55, 57, 61,
  64, 67, 70, 72, 75, 77, 80.
- Every cited credential and admission line is exact:
  `credentials.go:13` `CodeSSHCredentialMissing`, `credentials.go:63`
  `"%s is unavailable"`, `credentials.go:73` the file-mode branch,
  `admission.go:203` `trusted Git version probe failed`, `:211` `Git release
  family is not operator-pinned`, `:259` `HTTPS requires a manager credential
  broker`, `:332` `HTTPS credential host does not match protected source`,
  `:337` `cannot materialize HTTPS credential broker`.
- `session.go:489` is exactly the `untrusted_go_executable` site the new entry
  claims.
- Beyond G1 through G5, every remaining flag set matches the source: `bootstrap`,
  `remove`, `audit`, `skill check`, `global add` (minus G1), `global status`,
  `hybrid add`, `shell-init`, `config build-ssh add`, `config build-https add`,
  `config build-https login`. The `--username` default `token` is the effective
  default from `internal/config/buildhttps.go:40`
  (`BuildHTTPSDefaultUsername`), correctly stated.
- The seven subcommands with no flag set (`project resolve`, `global init`,
  `global remove`, `global list`, `global update`, `hybrid remove`,
  `hybrid list`, `hybrid status`, `list`, `ui`, `gc`, `config show`) are
  documented as taking no flags rather than carrying invented pairs.
- The README `## Commands` section sits before `## An open protocol`, links to
  `docs/cli.md`, and covers every top-level group the binary's usage block
  lists. Every link in the delta resolves: `docs/cli.md`,
  `docs/troubleshooting.md`, `docs/compiled-commands.md`, `docs/ci-gates.md`,
  `docs/authoring-language-adapters.md`,
  `docs/source-closure-adapter-conformance.md`, `CONTRIBUTING.md`, `LICENSE`,
  `NOTICE`, and the relative `../.github/...` and `../internal/...` targets added
  to `docs/ci-gates.md`.
- Style guide holds across `docs/cli.md`, `docs/troubleshooting.md`, and
  `README.md`. No em-dash or en-dash, no guillemets, no blacklisted opener, no
  marketing adjective, no closing summary paragraph, reasoning in prose and lists
  reserved for parallel flag enumerations.
- No regression in the README documentation guards:

```
$ go test ./cmd/curator -run 'TestEveryCurrentnessCodeIsDocumented|TestInputCausesAreDistinctAndDocumented' -count=1 -timeout 10m
ok  	github.com/relux-works/curator/cmd/curator	0.618s
```

## Rework instruction

Do not rewrite the documents. Six edits close this:

1. Delete `- `--project string`: project alias or path.` from the `curator global
   add` options list (`docs/cli.md:331`).
2. Add `- `--project string`: project alias or path.` to `curator add`
   (`docs/cli.md:76`), alphabetically before `--revision`.
3. Add `- `--all`: operate on all configured projects.` to `curator install`
   (`docs/cli.md:124`), `curator upgrade` (`:171`), and `curator status` (`:200`),
   first in each list.
4. Correct `README.md:137` to describe what `curator project add` does: register
   a project alias and path in the machine configuration.
5. Add `--all` and `--project` to the `## Shared flags` list (`docs/cli.md:9`).
6. Fix the `session.go:503` citation to `session.go:521` and note the
   `GOROOT/bin/%s` derivation on the `untrusted_go_executable` symptom.

Then re-verify by running each documented command with `-h` from a scratch
directory outside the repository, and diff the printed flag names, value types,
and defaults against the document list. Do not run `curator` subcommands inside
the repository tree: `curator init` has no flag set, so `curator init -h` treats
`-h` as a path and writes `Skillfile.json` into the working directory.
