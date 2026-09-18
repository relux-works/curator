# Brief — TASK-260916-3l60rn: which source kinds may carry MCP declarations and system modules; path overlays pass the store boundary (E6)

Story `STORY-260916-wgt8vz` (path-kind-admission-and-boundary), wave 3 of the
2026-09 security-audit remediation; spec first, then the manager task
`TASK-260916-yvxbs1`.
Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-wgt8vz/worktree`
(branch `task-board/story/STORY-260916-wgt8vz`, forked from curator-spec `main` `e8b53a0`, which carries S5 §4).
Rules: `remediation-spec-producer-rules.md` (attached). Role: doc-writer of
normative spec text; no implementation.

## Finding (read it first)
`docs/security-audit-2026-09.md` E6 (Medium) and Appendix B (E6 partially
confirmed: dependencies are git-only so a `path`-kind MCP package cannot
enter a closure — not applicable; a `path`-kind root or overlay still
carries `class: system` modules with no directory boundary). `path`-kind
sources (Decision 0012 §5 overlays, environments §6, §9.6 onboarding
imports) are pinned by state hash and have no canonical source identity, so
the MCP package allowlist and the E1 signer allowlist cannot name them, and
the store boundary of S5 applies to the directory itself. Text to revise:
Decision 0012 §5 (amendment section, like the E1 amendment), environments
§2.2 (MCP declaration packages), §3 (system modules — the E2 admission
rule), §6 (composition/overlays), §9.6 (onboarding import), §4 (S5 contract,
extended to `path` source directories), §1.1/§2.1/§3.1 diagnostics tables,
§12 posture, §13 conformance surfaces.

## Settled decisions (do not reopen)
- **MCP declarations: `git` sources only.** A `path`-kind package (root,
  overlay, onboarding import) MUST NOT carry an MCP declaration; a
  declaration found there is refused at resolution with
  `mcp_declaration_path_source_refused` (error) naming the package and the
  declaration — never admitted, never warned-through. This closes the "MCP
  half" the audit left open; it matches the implementation (dependencies
  are git-only) and makes the §2.2 allowlist total over canonical
  identities.
- **`class: system` modules from a `path` source: admitted only for a
  directly named `path` root or overlay (E2 direct-naming rule applies as
  is), AND only after the directory passes the §4 protected-boundary
  contract** (ownership by the operator, private permissions/DACL,
  containment below the declared directory, regular file types, link
  safety) at every resolve and before any materialization, exactly like a
  store entry; a `path` source that fails it is `environment_store_untrusted`
  for that profile (no fragment, non-current, posture row naming the path
  and the failing check) — the S5 outcomes apply, but there is no rebuild
  (the source is the operator's directory; the operator fixes it). Its
  `state_sha256` pin remains the integrity baseline as S5 defines.
- **`path` overlays and onboarding imports without system modules** are
  admitted as today but pass the same boundary contract (the check is on
  the directory, not on the content class).
- The E1 signer allowlist stays over canonical identities; a `path` source
  is never verified (E1 already says so) — say explicitly that this is why
  `path` sources may not carry MCP declarations and may carry system modules
  only under the boundary contract with direct naming.
- Rollout is direct (impact row "E6": under the hood).
- Closed diagnostics: `mcp_declaration_path_source_refused` new; reuse
  `environment_store_untrusted` (S5) and `context_system_module_transitive`
  (E2); spelled identically everywhere.

## Deliverable
1. Decision 0012 §5 amendment (dated 2026-09-18, E6) stating the kind
   admission rule; environments §2.2 kind restriction + diagnostic row,
   §3 system-module admission clause for `path` sources, §4 extension of
   the boundary contract to `path` source directories, §6/§9.6 pointers,
   §12 posture, §13 conformance surfaces.
2. Vectors: extend `environments.json` (or the family that carries closure
   admission cases — find it) with: a `path` overlay declaring an MCP server
   → refused; a directly named `path` overlay with a system module in a
   directory that passes the contract → admitted; the same with a
   world-writable directory → untrusted, no fragment; a symlinked component
   below the declared directory → untrusted; a transitive `path` package
   (if expressible) → refused by E2's rule; registered in the manifest;
   validator gate pins scenarios (rule 7); existing vectors byte-identical.
3. `CHANGELOG.md` Unreleased entry "E6: …".

## Out of scope
Implementation (`TASK-260916-yvxbs1`), S5 itself (landed), E1/E2 rules
(reference only), E7.

## Checklist and handoff
Tick the checklist items you satisfy; attach
`TASK-260916-3l60rn_spec-patch_rev1.patch` (`git diff HEAD` of the worktree
with new files via `git add -N`; NOT `git diff origin/main`) and
`TASK-260916-3l60rn_evidence.md` (with the `make validate` and regeneration
transcripts), then `task-board handoff TASK-260916-3l60rn --role doc-writer`.
