# TASK-260827-1nauj3 Rework Rev2 Outcome Report

## Applied Fixes Summary

1. **G1: `curator global add` flag fix**: Removed `- \`--project string\`: project alias or path.` from `curator global add` in `docs/cli.md`.
2. **G2: `curator add` flag fix**: Added `- \`--project string\`: project alias or path.` to `curator add` in `docs/cli.md`.
3. **G3, G4, G5: `--all` flag additions**: Added `- \`--all\`: operate on all configured projects.` to `curator install`, `curator upgrade`, and `curator status` in `docs/cli.md`.
4. **G6: `README.md` one-liner fix**: Corrected line 137 in `README.md` to: `curator project add      # register a project alias and path in the machine configuration`.
5. **Shared flags**: Added `--all` and `--project` to `## Shared flags` in `docs/cli.md`.
6. **Troubleshooting citations**: Fixed `session.go:503` to `session.go:521` for `toolchain_executable_mismatch` and noted `GOROOT/bin/%s` formatting derivation for `untrusted_go_executable` in `docs/troubleshooting.md`.

## Empirical Verification Outputs

### Binary Help Outputs (-h run from /tmp/curtest)

### curator bootstrap -h
```
Usage of bootstrap:
  -default-agents string
    	comma-separated default agents (default "codex_cli")
  -force
    	overwrite an existing configuration
  -if-missing
    	create configuration only when absent
  -non-interactive
    	fail instead of prompting for missing values
  -preferred-locale string
    	preferred locale
  -skills-root string
    	directory containing skill repositories
```

### curator add -h
```
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

### curator remove -h
```
Usage of remove:
  -project string
    	project alias or path
curator: remove requires a skill name
```

### curator install -h
```
Usage of install:
  -all
    	operate on all configured projects
  -audit
    	run the audit gate in advisory or strict mode
  -build-ssh-agent string
    	agent socket for external SSH build repositories, or "auto" for your own agent (or CURATOR_BUILD_SSH_AGENT)
  -build-ssh-identity string
    	identity file for external SSH build repositories (or CURATOR_BUILD_SSH_IDENTITY)
  -build-ssh-known-hosts string
    	host keys external SSH build repositories are verified against (or CURATOR_BUILD_SSH_KNOWN_HOSTS)
  -dry-run
    	plan work without modifying files
  -fix-gitignore
    	append missing managed gitignore entries
  -strict-tags
    	fail if an installed tag moved to another commit
  -verbose
    	print detailed progress
curator: flag: help requested
```

### curator upgrade -h
```
Usage of install:
  -all
    	operate on all configured projects
  -audit
    	run the audit gate in advisory or strict mode
  -build-ssh-agent string
    	agent socket for external SSH build repositories, or "auto" for your own agent (or CURATOR_BUILD_SSH_AGENT)
  -build-ssh-identity string
    	identity file for external SSH build repositories (or CURATOR_BUILD_SSH_IDENTITY)
  -build-ssh-known-hosts string
    	host keys external SSH build repositories are verified against (or CURATOR_BUILD_SSH_KNOWN_HOSTS)
  -dry-run
    	plan work without modifying files
  -fix-gitignore
    	append missing managed gitignore entries
  -strict-tags
    	fail if an installed tag moved to another commit
  -verbose
    	print detailed progress
curator: flag: help requested
```

### curator status -h
```
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

### curator audit -h
```
Usage of audit:
  -all
    	audit all configured projects and global skills
  -allow string
    	pin trust for a content hash
  -global
    	audit global skills
  -json
    	machine-readable output
  -publish string
    	signed audit record (JSON file) to submit
  -reason string
    	reason for --allow
  -registry string
    	registry base URL for --publish
  -token string
    	auditor token for --publish (or CURATOR_REGISTRY_TOKEN)
```

### curator skill check -h
```
Usage of skill check:
  -json
    	machine-readable output
  -locale string
    	validate against a locale
```

### curator project add -h
```
Usage of project add:
  -agents string
    	comma-separated target agents
curator: project add requires <alias> <path>
```

### curator global add -h
```
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

### curator global status -h
```
Usage of global status:
  -check
    	exit non-zero unless every skill and compiled command is current
  -json
    	machine-readable output
```

### curator hybrid add -h
```
Usage of hybrid add:
  -branch string
    	git branch
  -git string
    	git clone URL
  -revision string
    	git revision
  -tag string
    	git tag
  -target string
    	target alias, absolute path, or glob
  -targets string
    	comma-separated targets (alias, absolute path, or glob)
curator: hybrid add requires a name and --target or --targets
```

### curator shell-init -h
```
Usage of shell-init:
  -install
    	cache the hook and print its optional profile source command
  -no-global
    	skip global env sourcing
curator: shell-init accepts at most one shell: auto, zsh, bash, powershell
```

### curator config build-ssh add -h
```
Usage of config build-ssh add:
  -agent
    	select an SSH agent, optionally by socket path
  -identity string
    	identity file offered to the destination
  -known-hosts string
    	known-hosts file for this scope
```

### curator config build-https add -h
```
Usage of config build-https add:
  -git-credentials
    	use the operator's own Git HTTPS credential for this scope's host
  -keyring
    	use the token already stored for this scope (see login)
  -token-env string
    	environment variable read for the token at process entry
  -username string
    	username sent alongside the resolved token
```

### curator config build-https login -h
```
Usage of config build-https login:
  -username string
    	username sent alongside the stored token
```

### Doc Guard Test Verification

```
$ go test ./cmd/curator -run "TestEveryCurrentnessCodeIsDocumented|TestInputCausesAreDistinctAndDocumented" -count=1 -timeout 10m
ok  	github.com/relux-works/curator/cmd/curator	0.392s
```
