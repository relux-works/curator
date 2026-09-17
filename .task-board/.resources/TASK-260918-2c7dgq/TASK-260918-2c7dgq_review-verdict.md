# TASK-260918-2c7dgq — landing review

Verdict: ACCEPTED. Scope: exact curator-spec commit `4a2fa3ec428b5021e9e8a80f29fa7fc5a88e4990`, parent `e8b53a0`, E3 revision 2 against `684c9f1`. This review approves landing fidelity; it does not claim a manager implementation was exercised.

## Identities and method

Delivery worktree was read-only and clean. Both attachment digests match the assignment:
- union: `651809cf63fd94f11b6e41b97a3e8211703d64b33d47b87a0fc66d403372bcc8`
- accepted E3: `dd6d113236ac0137285e050ee8bb8b51e786530e961e392530db20b18e433feb`

The attached union is byte-identical to `git diff e8b53a0 4a2fa3e`. Accepted E3 was reconstructed by applying its patch to a disposable base tree. For each changed path, compare accepted bytes directly, or three-way merge base / landed / accepted. The complete changed-path sets are equal: 17/17; no additional files. Both conflict arms in CHANGELOG and validate.py concatenate exactly to delivery. In environments.md there are exactly 2/2 conflict regions; concatenation plus four connective edits equals delivery exactly: remove the earlier list's `and`; join the store row ending to the codex row beginning; change the write-nofollow ending to `afterwards;` and remove the next `and`; join the store block ending to `and the section 7.4 codex-seed`. No normative content is added or lost by those edits.

## Per-hunk review

All locations below refer to commit 4a2fa3e. Exact full merged excerpts follow the table.

| Hunk | Landed content retained | Accepted E3 content retained | Verdict |
| --- | --- | --- | --- |
| protocol/environments.md:2870–2883, §12 | “the store-trust row per installed profile”; “boundary and pin-hash verdict”; “naming the failing check”; “boundary (environments root or store root)” | “the active codex-seed revision”; “per managed `codex_cli` home the `codex_seed_record`”; all three native-server diagnostics and the non-empty A-under-B condition | PASS. One additive status list; shared recompute/report/no-mutation rule remains. Store-untrusted is non-current while seed diagnostics are warnings, so their combined reporting is consistent. |
| protocol/environments.md:3128–3168, §13 | E5 “write-discipline vectors”, both symlinked-backup refusal shapes and byte-identical former target; S5 “protected-boundary cases”, all refusal, repair, posture and negative cases | “section 7.4 codex-seed cases”, both revisions, no-server case, pre-rule and A-under-B posture, marker shape | PASS. All three conformance blocks retained in order; no scope weakened. Non-blocking punctuation nit at line 3157: `non-conforming.;` should read `non-conforming;`; it introduces no semantic contradiction. |
| tools/validate.py:6023–6745, constant/helper/validator conflict | `S5_DIAG_UNTRUSTED = "environment_store_untrusted"`, all S5 constants, pins, helpers and validator through line 6427 | `E3_DIAG_UNGOVERNED = "mcp_native_servers_ungoverned"`, other diagnostics, revision descriptions, both branch maps, helpers and validator | PASS. Exact concatenation of both conflict arms, not merely a name-presence check; S5 and E3 blocks use distinct names and retain all checks. |
| tools/validate.py:7805–7806, registration conflict | `validate_environments_store_boundary_vectors,` | `validate_environments_codex_seed_vectors,` | PASS. Both remain in the real main() checks list; E5 registration also remains at line 7811. |
| CHANGELOG.md:10–119, additive conflict | All E5/R3/P2/S5 intervening entries, ending “semantically by `tools/validate.py`.” | Entire E3 entry beginning “the `codex_cli` provisioning seed stops inheriting native” | PASS. Exact concatenation of both conflict arms; no landed entry removed or third entry invented. |

The §12 rows agree with §7.4 provisioning-only seed ownership and §8.2 recorded revision semantics. §13 preserves complete capability conformance while adding the separate codex-seed revision claim. Existing versioned schema/vector/generator structure is followed. Previously accepted E3 behavior was not redesigned in this landing review.

## File-by-file evidence

```
CHANGELOG.md: exact concatenation of BOTH conflict arms; all nonconflict bytes match
conformance/v1/manifest.json: generated pins; checked by regeneration
conformance/v1/schema-cases/agent-environment-marker-v1/invalid-codex-seed-record-names-member.json: accepted bytes identical
conformance/v1/schema-cases/agent-environment-marker-v1/invalid-codex-seed-record-names-not-array.json: accepted bytes identical
conformance/v1/schema-cases/agent-environment-marker-v1/invalid-codex-seed-record-on-linked-home.json: accepted bytes identical
conformance/v1/schema-cases/agent-environment-marker-v1/invalid-codex-seed-record-revision.json: accepted bytes identical
conformance/v1/schema-cases/agent-environment-marker-v1/invalid-codex-seed-record-unknown-field.json: accepted bytes identical
conformance/v1/schema-cases/agent-environment-marker-v1/valid-codex-seed-record-empty-snapshot.json: accepted bytes identical
conformance/v1/schema-cases/agent-environment-marker-v1/valid-codex-seed-record.json: accepted bytes identical
conformance/v1/schema-cases/index.json: accepted bytes identical
conformance/v1/vectors/environments-codex-seed.json: accepted bytes identical
protocol/environments.md: exact union after four enumerated connective/punctuation replacements, 2/2 conflict regions
release/1.0.0-rc.9.json: generated pins; checked by regeneration
schemas/v1/agent-environment-marker-v1.schema.json: accepted bytes identical
tools/generate-vectors/environments.go: accepted bytes identical
tools/test_validate.py: exact automatic three-way merge
tools/validate.py: exact concatenation of BOTH conflict arms; all nonconflict bytes match
```

## Exact union excerpts

### protocol/environments.md:2870

```text
2870: value, the store-trust row per installed profile — the section 4
2871: boundary and pin-hash verdict for the profile's lock, marker, and
2872: named store entries, naming the failing check and, for an enclosing
2873: failure, the boundary (environments root or store root) when the profile
2874: is `environment_store_untrusted`, the active codex-seed revision (`A` or `B`, section 7.4) with its
2875: behaviour, and per managed `codex_cli` home the `codex_seed_record`
2876: revision with its native-server rows (`mcp_native_servers_ungoverned`
2877: for the native names an `A`-record home carries,
2878: `mcp_native_servers_not_inherited` for a `B`-record home's stripped
2879: names, `mcp_seed_unstripped` with the re-provision hint when the marker
2880: predates the rule or when an `A`-record home with a non-empty snapshot
2881: is served by a revision-`B` manager). Both commands follow
2882: the manager §10 discipline exactly: recompute and report, never mutate — no
2883: fetch, no repair, no adoption, no channel application, no onboarding.
```

### protocol/environments.md:3128

```text
3128: `--all` confirming every profile of the run; the section 8.3.1
3129: write-discipline vectors
3130: (`vectors/environments-write-nofollow.json`) — the symlinked-target
3131: takeover (replaced with backup under authorization, stopped with
3132: `environment_foreign_manager_detected` without), the symlinked-parent
3133: refusal with `environment_write_would_follow_link` under both
3134: authorization states, the post-provisioning planted-link repair, the
3135: manager-owned-link replace, the symlinked-backup-destination refusals (a traversed parent link and a directly symlinked target),
3136: the inside-pointing-link ledger refusal, and the clean-path and
3137: recorded-file positives — every foreign-link case asserting the link's
3138: former target is byte-identical afterwards;
3139: the section 4
3140: protected-boundary cases
3141: (`vectors/environments-store-boundary.json`) — the intact resolve that
3142: emits a fragment; the swapped system-prompt bytes, swapped root-context
3143: bytes, symlinked entry root, wrong ownership, wrong permissions,
3144: containment escape, non-regular component, and pin-hash mismatch cases
3145: that refuse with `environment_store_untrusted` and emit no fragment; the
3146: environments-root and store-root enclosing-boundary cases that refuse with
3147: no rebuild; the intact-updated-store with old marker case that reports
3148: `environment_home_stale` and repairs, and the swapped-updated-store with
3149: old marker case that refuses with `environment_store_untrusted` and is
3150: never adopted; the unprovisioned-home intact and swapped cases; the
3151: unreadable-marker case that reports
3152: `environment_marker_unreadable`; the dry-run `would-rebuild-untrusted-store`
3153: entry-class case and the enclosing no-rebuild case; the repair rebuild,
3154: entry-rebuild, enclosing-refusal, stale-repair, and unprovisioned cases;
3155: the `env status` non-current posture rows naming the failing check and the
3156: boundary; and the negative cases whose fragment-emitting,
3157: current-reporting, or re-applying observation is non-conforming.;
3158: and the section 7.4 codex-seed
3159: cases (`vectors/environments-codex-seed.json`) — a native `config.toml`
3160: with `mcp_servers` tables under revision A (copied whole,
3161: `mcp_native_servers_ungoverned` names the entries, posture lists them
3162: ungoverned) and under revision B (the seeded file lacks them,
3163: `mcp_native_servers_not_inherited` names them), a native file without
3164: `mcp_servers` entries (no warning under either revision, the record still
3165: written), the pre-rule-home `mcp_seed_unstripped` posture, and the
3166: revision-A home under a revision-B manager (`mcp_seed_unstripped` with
3167: the re-provision hint, the recorded names still ungoverned) — with the
3168: `codex_seed_record` marker shape. The nine
```

### tools/validate.py:6023

```text
6023: # Sections 4, 8.4, 10.1, 10.4, 12: protected-boundary contract (S5)
6024: 
6025: 
6026: S5_DIAG_UNTRUSTED = "environment_store_untrusted"
6027: S5_OUTCOME_WOULD_REBUILD = "would-rebuild-untrusted-store"
6028: S5_DIAG_REPAIR_FAILED = "environment_repair_failed"
6029: S5_DIAG_HOME_STALE = "environment_home_stale"
6030: S5_DIAG_MARKER_UNREADABLE = "environment_marker_unreadable"
6031: 
6032: S5_BOUNDARY_CHECKS = ("ownership", "permissions", "containment", "regular_types", "link_safety")
6033: S5_OBJECTS = ("environments-root", "store-root", "lock-file", "marker-file", "store-entry")
6034: S5_ENCLOSING_OBJECTS = ("environments-root", "store-root")
6035: S5_SURFACES = ("system-prompt", "root-context")
6036: S5_ENTRY_KINDS = ("git", "path", "local")
6037: S5_HOME_STATES = ("current", "stale-old-marker", "unprovisioned", "unreadable-marker")
```

### tools/validate.py:6423

```text
6423:         if store_trusted:
6424:             if case.get("names_failing_check") is not None:
6425:                 raise ValidationFailure(f"store-boundary case {name}: a trusted row names no failing check")
6426:         elif case.get("names_failing_check") not in failing:
6427:             raise ValidationFailure(f"store-boundary case {name}: the status row names the failing check")
6428: # Sections 7.4, 7.7, 7.8, 8.2, 12: codex seed mcp_servers handling (E3)
6429: 
6430: 
6431: E3_DIAG_UNGOVERNED = "mcp_native_servers_ungoverned"
6432: E3_DIAG_NOT_INHERITED = "mcp_native_servers_not_inherited"
6433: E3_DIAG_UNSTRIPPED = "mcp_seed_unstripped"
6434: 
6435: E3_REVISION_A = "warning release: the codex_cli seed is still copied whole, but provisioning warns mcp_native_servers_ungoverned naming every inherited native mcp_servers entry with the migration hint"
6436: E3_REVISION_B = "flip release: the codex_cli seed strips the mcp_servers table and every mcp_servers.* sub-table; provisioning reports the stripped names once with mcp_native_servers_not_inherited"
```

### tools/validate.py:7793

```text
7793: def main() -> int:
7794:     checks = [
7795:         validate_schemas,
7796:         validate_repository_descriptor_identity,
7797:         validate_manifest,
7798:         validate_review_evidence,
7799:         validate_shared_fixture_markers,
7800:         validate_vector_semantics,
7801:         validate_assurance_vectors,
7802:         validate_environment_vectors,
7803:         validate_environments_env_passthrough_vectors,
7804:         validate_environments_source_signers_vectors,
7805:         validate_environments_store_boundary_vectors,
7806:         validate_environments_codex_seed_vectors,
7807:         validate_context_version_vectors,
7808:         validate_context_detector_vectors,
7809:         validate_snapshot_acquisition_vectors,
7810:         validate_shell_hook_trust_vectors,
7811:         validate_environments_write_nofollow_vectors,
7812:         validate_registry_page_boundary_vectors,
```

### CHANGELOG.md:10

```text
10: - E5: nofollow write discipline for managed-surface writes (environments
11:   §8.3.1, with pointers from §5/§5.8/§7.5/§8.1/§8.4/§9.5/§10.1/§12/§13
12:   and the manager profile §12.2): every materialization, takeover,
13:   repair, or backup write to a managed surface in any mode, a backup,
14:   the marker, or the adapter ledger replaces the directory entry —
15:   operation-private temp file plus rename — and MUST NOT follow a
16:   symlink at the target path or at any path component below the managed
17:   root that the manager did not create in this operation
18:   (`O_NOFOLLOW`-class open, `lstat`-class inspection). A manager-owned
19:   target link is replaced as an entry; the §9.5 foreign-manager stop
20:   keeps its disposition, with an authorized takeover backing up the
21:   link itself (same link text, never dereferenced) before replacing
22:   the entry; any other unowned target link follows the ledger rule
23:   (`environment_surface_unmanaged_conflict` unless a takeover
24:   authorization covers the path). One new diagnostic,
25:   `environment_write_would_follow_link`, refuses writes that would
26:   traverse a non-manager link below the managed root, or open through
27:   one at a backup, marker, or ledger destination — no takeover flag
28:   authorizes traversal. `env status` reports link-blocked paths as
29:   non-current rows naming the path. Direct rollout, under the hood: a
30:   tampered or symlinked target now yields a refusal or an
31:   entry-replacement, never a write through. Conformance vectors in
32:   `vectors/environments-write-nofollow.json` (authorized-takeover
33:   replace, unauthorized stop, symlinked-parent refusals under both
34:   authorization states, planted-link and manager-owned-link repairs,
35:   backup-destination refusal, inside-link ledger refusal, clean-path
36:   and recorded-file positives — every foreign-link case asserting the
37:   link's former target is byte-identical afterwards), checked
38: - S5: protected-boundary contract for the environments root and profile
39:   store (environments §4/§1.3/§8.2/§8.4/§8.5/§10.1/§10.4/§12/§13,
40:   manager §12.2/§12.5/§12.7), mirroring core §9.3: the environments root
41:   and every store entry are protected state — manager-created,
42:   manager-protected, resolved independently of package input. On every
43:   `env resolve`, and again under the manager-home mutation lock for every
44:   mutating profile operation (install, update, use, sync, repair,
45:   garbage collection), the manager MUST verify ownership, private
46:   mutation permissions or DACL, containment, regular file types, and link
47:   safety (`lstat`, no symlink at the root, the entry root, or any
48:   component the manager did not create) for the environments root, the
49:   profile store root, the lock and marker files, and every store entry
50:   the lock names; link-target identity is necessary but no longer
51:   sufficient currency at resolve. Store integrity is verified against the
52:   pin, not the marker: resolve recomputes every named entry's tree hash
53:   from its bytes and requires equality with the pin (`git` tree identity,
54:   `path`/`local` `state_sha256`), with no marker required and before any
55:   provisioning or repair, at O(store entry bytes named by the lock) —
56:   missing hashes never pass — so a same-user byte swap of a
57:   system-prompt or root-context file (or any entry file) is detected even
58:   for an unprovisioned home. Home currency is separate: a marker that
59:   belongs to another lock is `environment_home_stale` repaired from the
60:   verified store, never `environment_store_untrusted`; an absent marker
61:   is unprovisioned while an unreadable or malformed marker is
62:   `environment_marker_unreadable` (never "absent"). Two failure classes
63:   in order (enclosing boundary → entries → pin hashes → home currency):
64:   an enclosing boundary (environments root or store root) that cannot be
65:   proven refuses every resolve and every mutating operation before its
66:   first write with nothing rebuilt (operator repairs out of band), while
67:   an entry-class failure (store entry, lock, marker) inside a proven
68:   enclosing boundary is `environment_store_untrusted` (error): resolve
69:   emits no fragment, status is non-current, `env status` reports the row
70:   naming the failing check and the boundary; a real operation rebuilds a
71:   `git` entry from the revalidated snapshot into newly established
72:   protected state via operation-private staging and atomic publication (a
73:   `path` or `local` entry has no second copy, so repair fails with
74:   `environment_repair_failed` and the operator reinstalls); dry-run
75:   evaluation of an entry-class failure reports
76:   `would-rebuild-untrusted-store` and mutates nothing, while dry-run of
77:   an enclosing failure reports `environment_store_untrusted` with no
78:   rebuild planned. Repair re-applies only from entries that passed the
79:   contract and the pin hash, so `env resolve --repair` is not persistence
80:   for a tampered store (audit note E7). Rollout is direct, not warn-first
81:   (impact row "S5"): no knob — the contract is not configurable.
82:   Conformance: `vectors/environments-store-boundary.json` (intact
83:   resolve, swapped bytes, symlinked root, ownership, permissions,
84:   containment, non-regular, and pin-hash refusals, enclosing-root and
85:   store-root refusals with no rebuild, intact-old-marker stale repair and
86:   swapped-old-marker untrusted, unprovisioned intact/swapped,
87:   unreadable-marker, the entry-class dry-run outcome and the enclosing
88:   no-rebuild case, repair rebuild/entry-rebuild/enclosing-refusal/stale/
89:   unprovisioned, the non-current posture rows, and negatives), checked
90:   semantically by `tools/validate.py`.
91: - E3: the `codex_cli` provisioning seed stops inheriting native
92:   `mcp_servers` (environments §7.4/§7.7/§7.8/§8.2/§12/§13): a whole-copy
93:   seed runs every native MCP server outside the profile lock and the
94:   §2.2 allowlist while `claude_code` runs only the profile set, so the
95:   seed rule ships in two labelled revisions — revision A (warning
96:   release, ships first) still copies `config.toml` whole but provisioning
97:   warns `mcp_native_servers_ungoverned` naming every inherited entry with
98:   the migration hint (declare the server in the profile's MCP set, or
99:   accept the loss), and revision B (flip release) copies every top-level
100:   member except `mcp_servers` and reports the stripped names once with
101:   `mcp_native_servers_not_inherited`. Both revisions write the marker
102:   `codex_seed_record` (`revision` exactly `A` or `B`, `native_mcp_servers`
103:   the provisioning-time name snapshot, names only); a managed `codex_cli`
104:   home whose marker predates the rule — and a revision-A home now served
105:   by a revision-B manager — reports `mcp_seed_unstripped` (warning) with
106:   the re-provision hint, and existing homes keep their bytes. `env status`
107:   reports the active codex-seed revision with its behaviour and, per
108:   managed `codex_cli` home, the record with the native entries listed as
109:   ungoverned (revision A, including an A-record home under a B manager)
110:   or not inherited (revision B); all three rows are warnings and never
111:   make a row non-current. §7.8 gains the closed per-adapter residual
112:   table (what each channel does to the home's own MCP configuration).
113:   Conformance vectors in `vectors/environments-codex-seed.json`
114:   (provisioning under both revisions, the no-server and empty-table
115:   negatives, the pre-rule-home and the A-home-under-B-manager
116:   `mcp_seed_unstripped` postures, and the non-codex negative), pinned by
117:   the `validate_environments_codex_seed_vectors` gate. Manager/README follow-up
118:   (`TASK-260916-33abdk`): a managed codex home runs only the profile MCP
119:   set; native `~/.codex/config.toml` servers are not inherited.
```

## Verification method and bounded claims

Fresh gates were run by this reviewer; no prior pass was substituted. The disposable delivery tree was verified byte-for-byte against all 1,390 tracked commit blobs and the read-only delivery worktree, with a temporary Git baseline and local `core.autocrlf=false`. Commands used the requested repository venv on PATH. `make regenerate-check` ran the generator and compared all configured conformance/release outputs; whole-tree Git status was also clean afterwards. This checks the generated pins, not only the manually merged prose.

An initial `git archive` preparation was invalid: export-subst expanded the literal placeholder in fixtures/byte-exact/subst.txt, causing regenerate-check and validate to exit 2. Those results are setup failures, not evidence about the delivered tree. The replacement byte copy preserved all fixture bytes. No delivery files were modified. The accepted-patch reconstruction had one initial absolute-path git-apply invocation rejected, then succeeded from the disposable directory; it affected no source tree.

Additional adversarial checks drive real validate.main() with only the selected vector load substituted in memory: two internally consistent E3 branch-narrowing replacements and an S5 swapped-byte case falsely emitting a fragment. All other gates, including the actual main registration list, remain intact. The transcript gives the rejection ratio. This is validator-entry evidence, not a manager-runtime launch/materialize proof. Full normative clause mutation coverage is not claimed.

## Gate transcripts

Changed-file stat (`git diff e8b53a0..4a2fa3e --stat`):

```text
 CHANGELOG.md                                       |  29 ++
 conformance/v1/manifest.json                       |  34 ++-
 .../invalid-codex-seed-record-names-member.json    |  83 ++++++
 .../invalid-codex-seed-record-names-not-array.json |  80 +++++
 .../invalid-codex-seed-record-on-linked-home.json  |  82 ++++++
 .../invalid-codex-seed-record-revision.json        |  82 ++++++
 .../invalid-codex-seed-record-unknown-field.json   |  85 ++++++
 .../valid-codex-seed-record-empty-snapshot.json    |  80 +++++
 .../valid-codex-seed-record.json                   |  83 ++++++
 conformance/v1/schema-cases/index.json             |  35 +++
 .../v1/vectors/environments-codex-seed.json        | 246 ++++++++++++++++
 protocol/environments.md                           | 125 +++++++-
 release/1.0.0-rc.9.json                            |   4 +-
 schemas/v1/agent-environment-marker-v1.schema.json |  18 +-
 tools/generate-vectors/environments.go             |  11 +
 tools/test_validate.py                             | 289 +++++++++++++++++++
 tools/validate.py                                  | 321 +++++++++++++++++++++
 17 files changed, 1673 insertions(+), 14 deletions(-)
```

Commands executed in the verified disposable byte copy with `PATH=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH`:

### make regenerate-check — exit 0

```text
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
EXIT=0
```

### make validate — exit 0

```text
python3 tools/validate.py
validated 62 schemas and 1103 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
.......................................................................................................................................................................................................................................................................................................................................................................................................................................................
----------------------------------------------------------------------
Ran 439 tests in 905.491s

OK
go test ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	2.977s
EXIT=0
```

### python3 ../entry_checks.py — exit 0

```text
environments-codex-seed.json/b-strips-servers-keeps-rest: validate.main exit=1; validation failed: codex-seed case b-strips-servers-keeps-rest: native top-level members are not the pinned set (['mcp_servers', 'model', 'projects', 'tui'])
environments-codex-seed.json/a-home-unstripped-under-b: validate.main exit=1; validation failed: codex-seed case a-home-unstripped-under-b: revision_shipped does not match the pinned one (B)
environments-store-boundary.json/swapped-system-prompt-bytes-untrusted: validate.main exit=1; validation failed: store-boundary case swapped-system-prompt-bytes-untrusted: fragment verdict is not the §10.1 rule
3/3 targeted invalid cases rejected through main; input-loader substitution only, no gate or registration mocked; no manager runtime claim
EXIT=0
```

Final checks: disposable tracked tree clean after the complete suite; delivery tree clean and still at exact 4a2fa3ec428b5021e9e8a80f29fa7fc5a88e4990. Full make validate observed exit 0, including 439/439 Python tests (905.491 seconds) and the Go generator tests. The long-lived command was observed to completion via bounded output waits; no command remains backgrounded. This fresh run replaces the invalid archive-based setup attempt, and no previously attached green was used as a substitute.

32 CodexSeedVectorTests and 25 StoreBoundaryVectorTests are included in the 439 tests. Supplementary main-entry invalid substitutions: 3/3 rejected. This ratio covers those three selected attacks only; it is not an exhaustive coverage claim.

Run goal queried before verdict: RUN-260917-cc8986 is not goal-bound. No Change Request Under Review was supplied for this landing-review task; the accepted E3 source record is untouched. Acceptance is recorded in this task-scoped verdict; task routing follows the supplied landing-review brief. No commit_ack supplied, no source edit, no merge, no delivery commit.
