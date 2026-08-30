# TASK-260827-1nauj3 Rework Results

## Rework Summary

All blocking findings (F1–F7) and non-blocking notes from CR-TASK-260827-1nauj3-1 revision 1 have been resolved against the source of truth in `cmd/curator/main.go` and `internal/`.

### F1. `curator project add` Synopsis and Flags Corrected
- **Synopsis**: `curator project add <alias> <path> [flags]`
- **Flag**: `--agents string` (comma-separated target agents).
- **Fix**: Removed obsolete `--branch`, `--git`, `--project`, `--revision`, `--source`, `--tag` flags. Corrected argument arity to `<alias> <path>`.

### F2. `curator global add` Flags Corrected
- **Flags**: `--branch`, `--git`, `--revision`, `--source`, `--tag`.
- **Fix**: Removed `--project` flag which is not accepted by `global add`.

### F3. `curator hybrid status` Synopsis and Flags Corrected
- **Synopsis**: `curator hybrid status`
- **Fix**: Removed `--check` and `--json` flags. `cmd/curator/main.go:1546` defines no flag set for `hybrid status`.

### F4. `curator global install` and `curator global upgrade` Flags Corrected
- **Fix**: Removed `--all` flag from `global install` and `global upgrade` listings as both commands accept flags only and explicitly reject `--all`.

### F5. `curator skill check` Flag Types Corrected
- **Flag**: `--locale string` (validate against a locale, default empty `""`).
- **Fix**: Corrected flag type from `code` to `string` and default from `en` to empty string.

### F6. `curator list` and `curator ui` Positionals Corrected
- **Synopsis**: `curator list` and `curator ui`
- **Fix**: Removed positional `[path]` arguments as neither subcommand accepts positional parameters.

### F7. Toolchain Preflight and Git Admission Entries Restructured
- **Go Toolchain**: Replaced improper Git admission mapping with two authentic Go toolchain preflight entries:
  1. `untrusted_go_executable`: `CURATOR_GO must name an absolute GOROOT/bin/go` (`internal/godriver/session.go:489`).
  2. `toolchain_executable_mismatch`: `go-v1 toolchain_executable_mismatch: selected Go executable is not the regular executable under the derived GOROOT; put the real GOROOT/bin first on PATH, e.g. PATH="$(go env GOROOT)/bin:$PATH"` (`internal/godriver/session.go:503`, `cmd/curator/toolchain_remedy_test.go:51`).
- **Git Admission**: Created dedicated entry for `trusted Git version probe failed` and `Git release family is not operator-pinned` (`internal/buildrepo/admission.go:203, 211`).

### Additional Notes Handled
- **Note 1**: Imprecise cause prose for SSH identity/socket/known_hosts unavailable updated in `docs/troubleshooting.md`.
- **Note 4**: Aligned comment column for `curator config build-https` in `README.md`.

## Verification Output

### Test Suite Execution
```
$ go test ./cmd/curator -run "TestEveryCurrentnessCodeIsDocumented|TestInputCausesAreDistinctAndDocumented" -count=1 -timeout 10m
ok  	github.com/relux-works/curator/cmd/curator	0.427s
```

### Spot-Check Outputs against `./bin/curator`
```
$ ./bin/curator project add -h
Usage of project add:
  -agents string
    	comma-separated target agents
curator: project add requires <alias> <path>
exit=2

$ ./bin/curator global add -h
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
exit=2

$ ./bin/curator skill check -h
Usage of skill check:
  -json
    	machine-readable output
  -locale string
    	validate against a locale
exit=2
```

### Error String Grep Evidence
```
$ grep -n "CURATOR_GO must name" internal/godriver/session.go
489:			return "", "", nil, diagnostic("untrusted_go_executable", "CURATOR_GO must name an absolute GOROOT/bin/%s", platformGoName)

$ grep -n "trusted Git version probe failed" internal/buildrepo/admission.go
203:		return admissionError(CodeIdentityInvalid, "trusted Git version probe failed")
```
