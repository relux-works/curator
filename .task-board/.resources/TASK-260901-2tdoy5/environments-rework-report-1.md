# Rework report 1: protocol/environments.md (TASK-260901-2tdoy5)

Rework of draft 1 per `review-findings-environments-1.md` (reviewer run
RUN-260901-b7cddd). Worktree
`~/Developer/ReluxWorks/.worktrees/curator-spec-environments-normative`,
branch `draft/environments-protocol`.

- Base of rework: `eddd509` (reviewed head)
- Rework commit: **`c3b29b1f7f37829fd4d0c50b2023efa2feb4c615`** — one signed
  commit, `protocol/environments.md` only (+48/−16)
- Signature: `git verify-commit HEAD` against the repo's
  `maintainers.allowed_signers` → Good "git" signature for oparin@me.com,
  ECDSA `SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM` (the same key
  that signed main's `2a861e5` and draft head `eddd509`), exit 0.

## Finding → disposition

| Finding | Severity | Disposition |
| --- | --- | --- |
| M1 — managed `opencode.json` bytes unspecified but hashed as MUST-equal | major | **Applied as suggested.** §5.3 opencode bullet now pins the exact bytes: the CCJ-1 bytes (registry.md §1) of the single-member object `{"instructions": [...]}` — the ordered module reference list, no other member — followed by exactly one trailing LF. The surface stays inside the §5.6 hash set unchanged, and the §13 "referenced-form layout" vector surface is now producible: for given inputs exactly one byte string satisfies the rule (demonstrated below). Trailing-LF choice: exactly one, consistent with every other file byte rule in the document. |
| M2 — "collision-free by construction" false on case-insensitive filesystems; core §2 guard uncalled on the write path | major | **Applied as suggested, dedicated code.** New normative "Platform-path collisions" paragraph in the §5 opening: any materialization or provisioning step that would write two protocol paths mapping to one platform path MUST detect the collision and fail with the new stable code `environment_path_collision` before writing anything — the core §2 extraction rule extended to every materialization and provisioning write path, explicitly naming composed-profile module trees, managed homes below the environments root (§8.1), and backup paths (§8.3). §5.3 now says "collision-free in protocol-path space" with a pointer to the §5 rule; §8.1's managed-home bullet names the same failure for folding profile names; `environment_path_collision` added to the §5.7 table (single owning table, other sections reference §5). |
| m1 — §7.2 failure path had no stable code | minor | **Applied as suggested.** §7.2 names the error `environment_form_unsupported`; row added to the §7.7 table ("configured form not supported by the adapter"). Distinct from `environment_form_unavailable` (warning + fallback), as the reviewer required. |
| m2 — §7.3 flag spellings stated as closed fact without the pre-freeze caveat | minor | **Applied as suggested.** §7.3 now carries the §7.6-style sentence: the `flag` spellings are recorded from vendor documentation and verify against the pinned tool releases before the conformance vectors freeze. (Spellings were not re-verified against live binaries in this headless run; the caveat now rides where the claim rides, which was the finding.) |
| n1 — §6 and §11 inline unnumbered diagnostics tables | nit | **Applied.** Now `### 6.1 Diagnostics` and `### 11.1 Diagnostics`, matching every other section. |
| n2 — §8.4 reused `environment_marker_invalid` for a surface-file read failure | nit | **Applied, with a dedicated code** (the reviewer offered "name the intended code" as an option). §8.4 now separates the two: a failed marker read is `environment_marker_invalid`; a failed read of a recorded surface file is the new `environment_surface_unreadable` (row non-current, currency reported as unknown), and no absence-shaped outcome — `environment_surface_missing` included — may fire on either. Row added to §8.5. §12's "unreadable is reported as unreadable, never as absence" wording already agrees. |

No finding was declined; the only deviations from the literal suggested fixes
are the two dedicated diagnostic codes (`environment_path_collision`,
`environment_surface_unreadable`), both of which the review explicitly
offered as an acceptable branch.

## Evidence

- `make validate` at `c3b29b1` (scratch venv from `requirements-dev.txt`,
  since the ambient python3 lacks `jsonschema`): **exit 0** — "validated 53
  schemas and 691 vector files", 134 python unit tests OK, `go test
  ./tools/...` ok.
- Signature verification: exit 0 against `maintainers.allowed_signers`
  (see above).
- M1 producibility attack: generated the pinned bytes twice from the sample
  module list — byte-identical
  (`{"instructions":["...companyA/00-base.md","...personal/10-style.md"]}` +
  LF); the reviewer's failure scenario (an indented serialization of the
  same object) now violates the pinned rule instead of hashing differently
  while conforming.
- Diff scope: `git diff eddd509..c3b29b1 --stat` = `protocol/environments.md`
  only. No push, no tag, no other file touched.

## For the reviewer

- The two new codes join the `environment_*` family; both have exactly one
  defining condition and one table row (§5.7, §8.5).
- §7.3 spellings remain documentation-sourced pending the pre-freeze
  verification the caveat now promises — same open item as §7.6's Xcode
  paths.
