# TASK-260908-1bpra2 producer evidence

Ready for independent review. Changes remain uncommitted; no tags, installs,
home edits, real ax calls, or Safari launches. Signed PR/integration and
independent reviewer acceptance remain orchestrator responsibilities.

## Scope and provenance
- B3 epic goal read; newer b3-mcp-brief supersedes its stale LLDB requirement.
- agents-infra source 459742ea67e3c6b84169520b92d74fe7f73e3002: figma HTTP,
  Safari absolute driver with --mcp. Two packages delivered; LLDB dropped upstream.
- Story base HEAD and freshly advertised/fetched origin main both
  3125e491275c23c0e8b7f4da5ebfd3fe900220a0.
- Curator oracle source checkout HEAD: f750344f5e6fe3ef539485981c12e24240cca1ed.
  The brief's internal/mcp parser path is stale: declaration parser is
  internal/contextpkg.ParseMCP; internal/mcp verifies existing configuration.
  Oracle uses a throwaway module under this worktree .temp/oracle, replace to
  the read-only Curator checkout, and imports the actual ParseMCP entry point.
- Source identities and machine-current shim integration documented in README.
  MCP-only packages do not install skill command shims; deployment owner must
  expose the shipped wrapper. No path_prepend workaround.

## Package results
| Package | Fields | release gate | Curator parser |
| --- | --- | --- | --- |
| figma | schema_version=1, name=figma, version=1.0.0; http, pinned HTTPS URL, env_names=[] | 0 | 0 ACCEPT |
| safari | schema_version=1, name=safari, version=1.0.0; stdio, safaridriver-mcp, args=[--mcp], env_names=[] | 0 | 0 ACCEPT |

No token is declared by the pinned upstream configuration. Empty env_names
preserves that fact rather than inventing a credential requirement.

## Direct validation commands (zsh, no pipelines)
| Command | Exit |
| --- | --- |
| scripts/validate.sh | 0 |
| python3 -m unittest discover -s tests -v | 0; 6 tests |
| sh -n bin/safaridriver-mcp scripts/validate.sh | 0 |
| git diff --check | 0 |
| .temp/oracle: go run -mod=mod . ../../packages/figma/agent-mcp.json ../../packages/safari/agent-mcp.json | 0 |
| .temp/oracle: go run -mod=mod . absolute.json | 1, expected refusal |
| .temp/oracle: go run -mod=mod . token.json | 0, parser accepts argument text |
| .temp/oracle: go run -mod=mod . unknown.json | 1, expected refusal |

Oracle output:
    ../../packages/figma/agent-mcp.json: ACCEPT
    ../../packages/safari/agent-mcp.json: ACCEPT
    absolute.json: REJECT mcp_declaration_invalid: server.command must be a bare executable name without a path separator
    exit status 1
    token.json: ACCEPT
    unknown.json: REJECT mcp_declaration_invalid: unknown field server.token
    exit status 1

Parser is structural, not the secret-audit entry point: literal-token survival
is explicitly bounded. The repository gate rejects it through exact args.

## Narrowing mutant table
Tests launch scripts/validate.sh in isolated copies. Each original attack exits
1; narrowed implementation exits 0, defeating the normal rejection expectation.
| Attack | Narrowing | Normal | Narrowed | Result |
| --- | --- | --- | --- | --- |
| /tmp/driver command | path guard only rejects ./; command identity limited to transport | 1 | 0 | killed |
| --mcp --token=literal args | compare only first argument | 1 | 0 | killed |
| server.token literal field | require known fields as subset, allow extras | 1 | 0 | killed |

Measured 3/3 targeted narrowing shapes; 2/2 manifest inventory entries reached.
Not exhaustive protocol conformance or arbitrary secret detection. Additional
negative cases cover relative commands, credential URL query, top-level unknown
field, reserved environment name, wrong schema type, missing version, invalid
and duplicate-key JSON, null and missing manifest.

Wrapper tests substitute only the fixed driver path in a sandbox copy:
missing driver exits 127 with diagnostic; fake driver preserves --mcp,
space-containing and empty arguments and exit 23. Real /Applications execution
is unverified. POSIX syntax checked; Linux/Windows and Safari integration unrun.
No compiled application in this repository; Python test execution, shell syntax
and successful go run oracle provide the relevant validation/build evidence.
No full landing suite manually run; runtime owns it at handoff.

## Findings for review
The parser-location correction, LLDB drop, and parser token-audit bound above
are the task findings. Recorded here and in board notes; campaign instructions
explicitly prohibit LOGBOOK.md edits. No evidence reused from other producers.
