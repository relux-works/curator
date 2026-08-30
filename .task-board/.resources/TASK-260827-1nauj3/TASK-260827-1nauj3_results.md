# Verification Results for TASK-260827-1nauj3

## 1. Document Integrity
- `docs/cli.md`: 739 lines (verified intact)
- `docs/troubleshooting.md`: 321 lines (verified intact)
- `README.md`: Commands section present before protocol section with collapsible details groups linking to `docs/cli.md`.

## 2. Synopsis Spot-Checks

### Check 1: `curator bootstrap --help`
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
Verbatim match in `docs/cli.md`.

### Check 2: `curator project add --help`
```
Usage of project add:
  -agents string
    	comma-separated target agents
curator: project add requires <alias> <path>
```
Verbatim match in `docs/cli.md`.

### Check 3: `curator config build-https --help`
```
curator config build-https: operator HTTPS tokens for external build repositories

Usage:
  curator config build-https add <scope> (--git-credentials | --keyring | --token-env NAME) [--username NAME]
  curator config build-https login <scope> [--username NAME]
  curator config build-https list
  curator config build-https remove <scope>
```
Verbatim match in `docs/cli.md`.

## 3. Error String Grep Evidence in `internal/`

1. `trusted Git version probe failed`:
```
internal/buildrepo/admission.go:203:		return admissionError(CodeIdentityInvalid, "trusted Git version probe failed")
```

2. `HTTPS credential host does not match protected source`:
```
internal/buildrepo/admission.go:332:			return nil, admissionError(CodeIdentityInvalid, "HTTPS credential host does not match protected source")
```

3. `build_repository_ssh_credential_missing`:
```
internal/buildrepo/credentials.go:13:const CodeSSHCredentialMissing = "build_repository_ssh_credential_missing"
internal/install/buildsshprompt_test.go:193:	if !strings.HasPrefix(ErrBuildSSHAborted.Error(), "build_repository_ssh_credential_missing") {
```

## 4. Verification Summary
- Code compilation: `go build ./...` passed (exit code 0).
- All synopses and error codes match codebase exactly.
- Task ready for review handoff.
