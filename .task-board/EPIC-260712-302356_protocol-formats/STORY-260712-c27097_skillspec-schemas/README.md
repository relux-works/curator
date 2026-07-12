# STORY-260712-c27097: skillspec-schemas

## Description
csk-skill.json parsing and validation for schemas 1 through 5: schema gating, unknown-field rejection (v2+), runtime_roots, script and system commands, capabilities, dependencies.commands incl. legacy skill type, dependencies.skills, dependencies.mcp_servers, no install hooks, legacy agents/runtime.json fallback (Spec 5).

## Acceptance Criteria
- One table-driven test per MUST of Spec 5
- Valid/invalid fixture corpus under testdata
- Legacy runtime.json fallback covered
