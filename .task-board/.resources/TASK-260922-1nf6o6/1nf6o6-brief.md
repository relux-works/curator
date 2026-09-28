# TASK-260922-1nf6o6 — F-S1: marker credential record (curator-spec)

Control root: /Users/administrator/Developer/ReluxWorks/curator/curator-spec; work only in your
assigned Story worktree `.temp/STORY-260922-188t6n/worktree`. Read `campaign-producer-rules.md`
first. Gate = the board's validation command (the runtime runs it once at handoff).

## Source of truth
Decision 0017 is ADOPTED (`decisions/0017-environment-credential-modes.md`, adoption choice 5 and
the Compatibility section): the marker-schema extension is a NAMED follow-up spec revision — this
leaf. `protocol/environments.md` §7.4 already carries the normative "Credential record" paragraph
(≈ lines 1496–1516: every passthrough record carries `isolation`, `strategy`, `source_role`
(`native`|`managed`), `backend` (`file`|`keychain`|`ambient`), `backend_version`, `provenance`
(`provisioned`|`repaired`|`migrated`); linkless strategies record without a path; a schema-1 marker
records `path` and `strategy` only and is never rewritten to add the record; publication under the
manager-home mutation lock through same-directory temp + atomic rename with journal protection,
rollback restores the preceding marker; backup discovery and backups never follow auth symlinks and
never archive credential bytes). What is MISSING is the schema that admits it.

## What exists
`schemas/v1/agent-environment-marker-v1.schema.json` — `"title": ".agent-environment.json schema 1"`,
`version: {const: 1}`, `passthrough` = array of `{path, strategy}` with
`additionalProperties: false` and `required: [path, strategy]`; strategies
`per-home-keychain|file-link|keyring-preferred|ambient|in-place`.
Cases live in `conformance/v1/schema-cases/agent-environment-marker-v1/` (72 files today, named
`valid-*.json` / `invalid-*.json`). Follow the existing naming and structure exactly; check
`conformance/v1/manifest.json` for how a schema family and its cases are registered.

## Deliverable
1. `schemas/v1/agent-environment-marker-v2.schema.json`: schema 1 plus the credential record on each
   `passthrough` entry — `isolation` (`shared`|`isolated`), `strategy` (unchanged enum),
   `source_role` (`native`|`managed`), `backend` (`file`|`keychain`|`ambient`), `backend_version`
   (non-empty string: the verified tool release), `provenance` (`provisioned`|`repaired`|`migrated`),
   and `path` REQUIRED only for linkable strategies — `per-home-keychain` under `isolated` and
   `ambient` record WITHOUT a path (encode that conditionality in the schema, e.g. allOf/if-then,
   and keep `additionalProperties: false`). `version: {const: 2}`. Everything else identical to
   schema 1 (diff the two and say so in results.md).
2. Schema cases under `conformance/v1/schema-cases/agent-environment-marker-v2/`: valid record per
   strategy (file-link/native/file; keyring-preferred; per-home-keychain isolated without path;
   ambient without path; in-place); invalid: missing any of the six members, unknown `backend`,
   unknown `provenance`, unknown `isolation`, `path` present on a linkless strategy, `path` absent
   on a linkable one, unknown extra field. Keep the schema-1 family and its 72 cases UNCHANGED, and
   add (or keep) the case proving a schema-1 marker is valid without the record and invalid WITH it.
3. Register the new family wherever the manifest/generator requires it; `make validate` and the
   regenerate-check must be green.
4. `protocol/environments.md` §7.4/§8.4.1: point the "Credential record" paragraph at schema 2 by
   name and state that a schema-1 marker is never rewritten to add the record (upgrade happens only
   when the manager rewrites the marker for another reason — state the exact rule you choose and
   keep it consistent with §12.4/§12.5).
5. CHANGELOG (unreleased). results.md MUST name the curator follow-up leaf that has to PUBLISH the
   record (manager side) with the exact member names, the schema id, and the upgrade rule, so the
   orchestrator can create that leaf verbatim.

## Boundaries
No frozen v1 protocol schema changed — you ADD schema 2, you do not edit schema 1. No curator code.
No fleet-policy work (that is F-S3, TASK-260922-1ejkxv) and no fragment work (F-S2,
TASK-260922-1hla8q). Hand off with `task-board handoff TASK-260922-1nf6o6 --role developer` after
attaching `TASK-260922-1nf6o6_results.md` (schema diff, case table with pass/fail evidence, gate
command + exit code, the named curator follow-up, bounds).
