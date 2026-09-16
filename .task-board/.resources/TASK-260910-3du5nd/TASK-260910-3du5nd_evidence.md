# Evidence — TASK-260910-3du5nd: repository-transport revision 2 (spec amendment)

Worktree: curator-spec `.temp/STORY-260916-2txa8v/worktree` (branch
`task-board/story/STORY-260916-2txa8v`). Spec only; no implementation code.
All changes uncommitted in the worktree for handoff snapshot.

## What was specified

`protocol/repository-transport.md` now carries revisions 1 and 2. Revision 1
(§§1–3) is intact except the title, a scoping paragraph, and the closing
sentence that previously left advanced mappings undecided. Revision 2
(§§4–7, new) normatively decides:

- §4 scope/versioning: `repository-transport-v2` opt-in; new file
  `schemas/draft-sources-v1/source-policy-v2.schema.json`
  (`schema_version: 2`), an additive superset of byte-unchanged schema 1
  (optional per-endpoint `mirror_of`/`alias`, optional top-level `aliases`,
  port-admitting endpoint/`pin` URL grammar). v2 readers must also accept v1
  documents; v1 readers must reject v2 documents as
  `repository_policy_invalid`. No other schema, manifest, lock, receipt, or
  lane declaration grammar changes; core 6.1/6.3 unchanged.
- §5 identity/ports/mirrors/aliases: canonical `host/path` stays the only
  portable identity (ports/mirrors/aliases never enter manifests, lock,
  receipts, markers, allowlists); URI-form ports `1`–`65535` no leading
  zeros, stripped before canonicalization, included in `pin` equality;
  scp-like spelling carries no port; mirrors admitted only with exact-key
  `mirror_of` (present iff URL host differs); aliases declared only in the
  operator `aliases` table, named only via the `alias` field (never embedded
  as URL host), single substitution, no chaining, alias auth must equal
  endpoint auth, no double port; identity always derives from the endpoint
  URL; alias routing to another host requires `mirror_of`. Package
  declarations stay canonical/logical; user ssh/git configuration remains
  never imported.
- §6 resolution/failure classes: ordered fail-before-network-I/O checks; new
  `repository_mirror_undeclared` and `repository_alias_unknown`; misuse
  variants stay `repository_policy_invalid`. Two-endpoint cap kept (max two
  total, one attempt each); revision 1 fallback table unchanged; mirrors are
  ordinary endpoints once declared (may be first).
- §7 secrets/provenance/compat/external builds: secret rules unchanged;
  sanitized provenance (listed URL+port, resolved host:port, alias,
  `mirror_of`) in machine-private diagnostics only; no port field added to
  receipt transport fields (ports don't change receipt/cache identity);
  manager §11 boundaries mandatory for every v2 attempt.

## Decisions taken (with rationale)

1. Same-file additive section (§§4–7), not a companion file: protocol/ has no
   per-revision file precedent; environments.md carries revision 1.1 in-file.
2. New `schema_version: 2` file, not optional fields under 1: the version is a
   const discriminator; forking shape under one value would split "schema 1",
   while v1 readers already fail v2 documents closed.
3. `mirror_of` attestation field name; required iff URL host differs from key
   host (spurious attestation and value mismatch are `repository_policy_invalid`).
4. Aliases referenced explicitly via `alias` field; URL-embedded alias hosts
   rejected. This makes `repository_alias_unknown` well-defined (transparent
   embedding would collapse typos into mirror errors and risk DNS lookup of
   alias names).
5. Alias auth must equal endpoint auth (no precedence rule); no alias
   chaining; URL port + alias port mutually exclusive — all fail closed.
6. Two-endpoint cap kept: endpoint properties create no new failure class
   needing more attempts; preserves rev1 deadline accounting and blast radius.
7. Ports excluded from receipt/cache identity beyond the existing
   https/ssh transport enum (avoids cache fragmentation and frozen-schema churn;
   verified content is identical).

## Conformance vectors

- 12 schema cases in `conformance/draft-sources-v1/schema-cases/source-policy-v2/`
  (4 valid: port, mirror, alias, v1-shape-under-2; 8 invalid: unknown top-level,
  port range, leading-zero port, `mirror_of` pattern, alias name, alias missing
  auth, string port, extra endpoint member), all indexed in `index.json`.
- 18 `v2-*` cases appended to `semantic-cases.json` (5 positive incl.
  mirror-first and v1-policy acceptance; 13 refusal incl. undeclared mirror,
  pin/port mismatch, unknown alias, alias-routed undeclared mirror, user
  ssh-alias/insteadOf ignored, embedded alias host, auth mismatch, chaining,
  double port, spurious/mismatched `mirror_of`, v1-reader-rejects-v2).
- Every vector id is named in §7; schema files versioned additively.

## Open questions

None left in this leaf's scope. Still open per UNRESOLVED_QUESTIONS.md:
reusable cross-machine logical alias registries and safe translation of user
`insteadOf`/SSH configuration (rejected until a future design). No manager
implementation or release qualification claimed.

## Validation (real exit codes, venv at .temp/venv, gitignored)

- Baseline before changes: `make validate` → exit 0.
- Draft conformance command verbatim from
  `conformance/draft-sources-v1/README.md` → exit 0:
  `Marker migration: 25/25 ... 18/18 detected`,
  `Schema cases: 114/114; negatives: 90/90; wire schemas: 8/8`,
  `Snapshot byte vectors: 3/3`.
- `make validate` after changes → exit 0
  (`tools/validate.py`: validated 60 schemas and 1047 vector files;
  unittest: Ran 227 tests, OK; `go test ./tools/...`: ok).
  Note: `make` hardcodes `python3`; run with the venv first on PATH
  (`PATH="$PWD/.temp/venv/bin:$PATH" make validate`) since the system
  interpreter lacks `jsonschema`. Status captured via PIPESTATUS, not masked.
- Extra probes (not committed): 10/10 port-grammar boundary probes pass
  (`:1`/`:65535` accept; `:0`/`:65536`/leading-zero/userinfo-on-https/empty-path
  reject; scp `host:2222/path` stays a path); the new operator-guide rev2
  example validates under source-policy-v2.
- `git status`: 11 modified files + 2 new paths; source-policy-v1 schema
  byte-untouched; no writes outside the Story worktree (board ops excepted).
- Logbook: not edited per campaign-producer-rules (no LOGBOOK.md edits in
  Story worktrees); findings and decisions are recorded in this artifact.

## Files changed

- `protocol/repository-transport.md` (§§4–7 added; rev1 intact)
- `schemas/draft-sources-v1/source-policy-v2.schema.json` (new)
- `schemas/draft-sources-v1/README.md`, `conformance/draft-sources-v1/README.md`
- `conformance/draft-sources-v1/index.json`, `semantic-cases.json`,
  `schema-cases/source-policy-v2/` (12 new cases)
- `UNRESOLVED_QUESTIONS.md`, `CHANGELOG.md`, `README.md`, `COMPATIBILITY.md`,
  `docs/skillfile-sources.md`, `protocol/skillfile-sources.md` (link text)

## Revision 2 (rework 1; reviewer verdict TASK-260910-3du5nd_review-verdict-rev1.md, CHANGES_REQUESTED)

Fixed exactly the three verdict findings; no other changes. All changes
remain uncommitted in the Story worktree for handoff snapshot.

1. HIGH — alias/mirror admission contradiction (§5, was ~175-207): one
   consistent rule now. The resolved connection host is the alias target
   host when an `alias` is named, else the URL host; `mirror_of` is
   required iff the RESOLVED host differs from the canonical key host and
   forbidden otherwise (missing → `repository_mirror_undeclared`;
   value mismatch or spurious attestation →
   `repository_policy_invalid`). Repository identity for an entry is
   always the entry key — never the URL host or alias target — and the
   port-stripped canonical path must equal the key path even for
   mirrors. Covered in text: canonical URL + mirror alias (requires
   `mirror_of`), same-host alias (forbids it), and mirror URL + `alias`
   (refused as `repository_policy_invalid`: the URL host would be
   neither identity nor connection address). The attestation authorizes
   the resolved connection host only. §6 ordered checks and the failure
   table now use resolved-host language. Shipped vectors already agreed
   with this predicate (`valid-alias.json`, `v2-alias-resolution`,
   `v2-alias-mirror-undeclared`, `v2-spurious-mirror-of`,
   `v2-alias-auth-mismatch`, `v2-double-port`, `v2-alias-chain`,
   `v2-embedded-alias-host` rechecked case by case); no vector edits
   needed for this finding.
2. HIGH — strict external-build lane consistency (narrower option, per
   rework brief): revision 2 does NOT extend manager §11.2 URL grammar
   or the §11.3 SSH wrapper. §4 states the carve-out; §5 port/alias
   paragraphs point to it; §7 normatively refuses port-bearing (URL or
   alias port) or alias-naming endpoints selected for the strict lane
   with `build_repository_identity_invalid` before network I/O, and
   forbids stripping the port or ignoring the alias to force admission.
   Port-free, alias-free mirror endpoints are ordinary §11.2 URLs and
   stay admissible with identical verification. Three new
   `semantic-cases.json` contract cases (all named in the §7 inventory):
   `v2-external-build-mirror-admitted` (positive),
   `v2-external-build-port-refused` and
   `v2-external-build-alias-refused` (refusals; the alias case carries a
   transport-valid `mirror_of` to pin that lane refusal is independent).
3. MEDIUM — SSH upper-bound refusal vector: new indexed
   `schema-cases/source-policy-v2/invalid-port-range-ssh.json`
   (`ssh://git@example.org:65536/...` and `:65539/...`, both in one
   invalid instance) + `index.json` entry; §7 inventory updated to nine
   negatives. HTTPS vectors untouched.

Validation (real exit codes, worktree venv `.temp/venv`, no installs):

- Mutant probe `/tmp/3du5nd/mutant_probe.py` (throwaway, not committed):
  4/4 PASS — new SSH vector invalid under the real schema, VALID under
  the SSH-only widened mutant (`6553[0-5]`→`6553[0-9]` in the SSH
  branch only; HTTPS branch untouched), old HTTPS `:70000` vector still
  invalid under the mutant (not the detector), `valid-port.json` still
  valid. Exit 0.
- Draft conformance command verbatim from
  `conformance/draft-sources-v1/README.md` (extracted to
  `/tmp/3du5nd/draft_check.py` for reviewability): exit 0 —
  `Marker migration: 25/25 ... 18/18 detected`,
  `Schema cases: 115/115; negatives: 91/91; wire schemas: 8/8`,
  `Snapshot byte vectors: 3/3`,
  `Manager semantic execution: 0 cases` (semantic cases remain contract
  examples, as the reviewer noted).
- `PATH="$PWD/.temp/venv/bin:$PATH" make validate` → exit 0
  (`tools/validate.py`: validated 60 schemas and 1047 vector files;
  unittest: Ran 227 tests, OK; `go test ./tools/...`: ok).

Files changed by rework 1 (on top of the rev-1 delta):
`protocol/repository-transport.md` (§§4–7 targeted edits),
`conformance/draft-sources-v1/semantic-cases.json` (+3 cases),
`conformance/draft-sources-v1/index.json` (+1 entry),
`conformance/draft-sources-v1/schema-cases/source-policy-v2/invalid-port-range-ssh.json`
(new). No schema-file, manager-profile, CHANGELOG, or docs edits.
