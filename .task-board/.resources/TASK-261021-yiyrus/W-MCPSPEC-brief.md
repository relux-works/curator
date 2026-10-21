# Brief: spec text, MCP server declarations in the project Skillfile (curator-spec, new CIP)

**Repository:** curator-spec. **Branch:** `cip-mcp-in-project-skillfile`, from `origin/main`.

## Context and owner decisions (2026-10-21)

- The design agreed with another Skillfile implementation is attached as `mcp-design-agreed-v1.md`. Treat it as **input**: the product-named member in it is replaced, per the decisions below.
- **Independence.** The Curator spec names no other implementation and reserves no names for one. The writer-local extension beside `server` is a neutral member `install` (headers, auth, scope, git_policy). The spec defines it neutrally as "writer-local install options that a manager may honour or ignore and names in its plan", or leaves it undefined.
- **Environments.** One neutral vocabulary of environment ids, Curator's adapter ids. Absent means every environment the manager can write. A listed environment the manager cannot write is reported per pair as `skipped: unsupported`, never refused. An unknown id is invalid.
- **Two stages.** (a) First, the `package` form: a git `agent-mcp.json` declaration package resolved under the existing allowlist and audit. (b) Inline project declarations, accepted by the owner as a CIP-0002 project contribution: each one approved by the operator, the approval bound to the digest of the normalized `server` object, and any edit re-approved. The allowlist key is the project identity plus NAME.
- The owner wants to discuss the spec text before it is final, so the status is **Draft**.

## Do this

1. Read GOVERNANCE.md and the CIP template and process. Read environments §2.2 (`agent-mcp.json`, the `server` object, URL rules), the Skillfile spec (schema 1/2, the source extension), manager §12, and CIP-0002 (project contribution, approval, ceilings).
2. Write a new CIP, numbered as the next free one: "MCP server declarations in the project Skillfile". It covers:
   - Skillfile schema 3 = schema 2 plus `mcp_servers`, project scope only; a machine-global Skillfile keeps rejecting 3;
   - the grammar: NAME, the URL shorthand, the object form `{server, install?}` with `server` exactly the closed `agent-mcp.json` server object;
   - the URL and secret rules;
   - the environments rule;
   - the `package` form;
   - the inline stage as a CIP-0002 contribution with the approval binding;
   - digests over the normalized model, so the shorthand equals its expansion;
   - per-pair reporting;
   - typed refusal codes;
   - what a manager does with `install`;
   - security considerations, including a token in a URL as a blocking finding and no process ever started at install;
   - compatibility (rc.7 refuses schema 3 with the upgrade error);
   - conformance vectors to add;
   - the open questions for the owner.
3. Add conformance vector stubs only if the repository's process says CIP drafts carry them. Otherwise list them in the CIP.
4. Register the CIP wherever the repository indexes CIPs (README, cips/README, CHANGELOG) as its rules require.

## Hand-off

Results file: the summary, the open questions and the files changed. Branch pushed, signed.
