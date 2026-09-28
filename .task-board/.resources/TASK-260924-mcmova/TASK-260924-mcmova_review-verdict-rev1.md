# TASK-260924-mcmova review verdict — CR rev1: ACCEPTED
Candidate tree 666d0f02 on base fe2d1c64. Checked in a disposable clone (base commit plus the candidate tree archive), shell zsh with pipefail.

## Normative review
- **Definition:** a hard-link substitution is a link that makes resolution select a file identity other than the platform-owned executable the manager intended. Stated in core.md (script policy paragraph) and in manager.md.
- **Bounds (core.md exec-capability bullet, manager.md, same conjunctive wording):**
  - it applies to Windows only and only to declared `exec` names;
  - the name must be resolved through the manager's default Windows search list only;
  - canonical `%SystemRoot%\System32` must be derived from the manager's own captured SystemRoot;
  - the target must be physically below that directory;
  - every extra link must be in the same root's `WinSxS` component store;
  - if any condition is false or cannot be established, the exception is rejected;
  - every other multiply-linked target stays rejected.
  - core.md also states that symlink and reparse-point rejection is unchanged.
- **No widening found:** no other directory, no caller or package SystemRoot, no caller search directories.
- **Interpreters:** python3-v1 and node-v1 are explicitly excluded in both documents.
- **Vector:** the script-host-execution-policy vector adds a definition and 8 `executable_identity_cases`: 1 accept and 7 rejections, one for each bound plus python and node. The generator, the Go test and validate.py pin these cases with exact equality.
- **Test mutants:** 7 mutants in test_validate.py loosen one bound each (target, links, SystemRoot, search, ownership, python, node). They are rejected.
- **Other changes:** manifest and rc.9 candidate pin hashes are regenerated; CHANGELOG (unreleased) entry added. No other normative change.
- **Minor, not blocking:** manager.md does not repeat the symlink and reparse-point sentence; core.md carries it.

## Commands rerun
| Command | Exit |
|---|---|
| `python tools/validate.py` (venv with jsonschema) | 0 (64 schemas, 1169 vectors) |
| `make regenerate-check` | 0 |
| `go test ./tools/...` | 0 |
| `unittest test_validate.WireSemanticValidationTests` (includes the new mutants) | 0 (18 tests) |

The full `unittest discover` suite (about 20 minutes) was not rerun. I accepted the producer's split-run table for it.
