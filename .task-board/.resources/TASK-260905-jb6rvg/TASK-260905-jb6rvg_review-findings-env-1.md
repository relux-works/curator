# Review findings: environments.md revision 1.1, cycle 1 (TASK-260905-jb6rvg)

Subject: Change Request `CR-TASK-260905-jb6rvg-1` revision 1, commit `4492b7e` on
`task-board/story/STORY-260905-1xwg3d` (base `ec695ba`, candidate tree `eb9fa4d`).
One file, `protocol/environments.md`. Reviewer run `RUN-260905-85a2ae`, 2026-09-05.

## Verdict

**Changes requested** (three major findings, no blocking). The rewrite is structurally
complete: every 0012 impact-table row is applied, every M4–M16 and N1–N14 item is at its
anchor with its diagnostic tabled, Decision 0013 and the 0010 erratum are honored, and every
verified tool fact reproduces on the installed binaries. The three majors are each a short
edit; they concern a fragment that the next batch's schema cannot represent (F1), an
internally contradictory normative row (F2), and a hot-path lock rule that the text leaves
open to the exact cost M10 was raised for (F3).

## Gates re-run by this reviewer

| Gate | Result | Evidence |
|---|---|---|
| Commit signature | Good `git` signature, author Ivan Oparin, ECDSA `SHA256:V6JiKG…` | `git log --show-signature -1` (the `allowed_signers` warning is a local verifier-config artifact, not a bad signature) |
| Scope | exactly one path changed, `protocol/environments.md` (+1411/−407) | `git diff --stat ec695ba eb9fa4d` |
| `make validate` | exit 0: 57 schemas, 780 vectors, 152 unittests OK, go test ok | `.temp/review-env-1/make-validate-02.log` (a first run failed only because my own scratch copies of the base text sat under `.temp/` inside the tree — same trap the producer hit; moved out and re-run) |
| Cross-references | every `section N.N` reference resolves to an existing heading; bare `§` references all carry a document name | script over the text; 57 headings, 0 unresolved |
| Byte identity | identical to `ec695ba`: §1.2, §5.2, §7 body, §7.2, §8 body, §8.4, §9 body, §10 body | heading-split diff |
| claude 2.1.261 | `--system-prompt-file`, `--append-system-prompt-file`, `--mcp-config`, `--strict-mcp-config` present | `claude --help` |
| codex 0.153.2 | `-p, --profile`, `-c, --config <key=value>` present; brief said 0.151.0, the text records 0.153.2 and says so | `codex --help` |
| pi 0.84.2 | `--system-prompt <text>`, `--append-system-prompt <text>` ("Append text or file contents"); `--system-prompt-file` and `--append-system-prompt-file` rejected "Unknown option"; no MCP flag | `pi --help`, direct invocation |
| opencode | not installed; every opencode fact is labeled docs-confidence in the text | `command -v opencode` |

## Findings

### F1 — major — §7.3 / §7.8 / §10.2: the codex MCP descriptor needs a `name` member the closed descriptor grammar does not define

Quote (§7.8 codex row): "`flag` `-p` with `argument: name`, `name: "curator-mcp"`".
Quote (§7.3): "A `flag` descriptor carries `argument`, exactly `path`, `contents`, or `name` … A `flag` descriptor MAY carry `with`". Quote (§10.2): "`channels` reproduces the adapter's section 7.3 descriptors (`flag` with `flag`, `argument`, and OPTIONAL `with`; …)" and "readers MUST reject unknown fields".

What is wrong: `argument: name` means "the launcher passes a fixed reserved name", but no descriptor member carries that name. The §7.8 row writes `name: "curator-mcp"` as if it were a member; §7.3 and §10.2 enumerate the flag-descriptor members without it, and §10.2 makes the fragment closed. A `codex_cli` fragment with an MCP set is therefore unrepresentable under the text as written, and the next batch's `launch-env-fragment-v1` schema has no rule to carry it. Decision 0012 Decision 6 writes the same `name: "curator-mcp"` spelling, so the intent is clear; the grammar just never admits it.

Fix: in §7.3 add one sentence — "A `flag` descriptor with `argument: name` additionally carries `name`, the fixed reserved name the launcher passes; `name` is absent for `path` and `contents`." In §10.2 extend the member list to "`flag` with `flag`, `argument`, `name` when `argument` is `name`, and OPTIONAL `with`". Optionally add the codex fragment shape to the §10.2 example or state that the example is the `claude_code` shape only.

### F2 — major — §7.4 passthrough table: the `claude_code` Linux row is labeled `directory` but describes a per-file symlink

Quote: "Linux: `directory` — the managed home's `.credentials.json` is a symlink to the native file, re-checked by the liveness row". Quote (same section): "a `directory` strategy keeps the link one level above the rewritten file".

What is wrong: the row names the strategy M4(b) reserves for rename-over tools and then defines it as exactly the per-file link that the following paragraph says a rename-over write severs. `.credentials.json` lives at the home root, so "one level above" would be the whole home, which cannot be a link. The row is internally contradictory in the one table M4(b) exists for.

Fix: either relabel the Linux entry `file-link` (with the liveness row and the `unverified` write-behavior label it already carries), or, if a directory-level link is intended, name the directory that is linked. Also add the case to the strategy paragraph: a `file-link` under a rename-over tool is expected to detach and is caught by the liveness row and `--repair`.

### F3 — major — §10.1: under `--repair` the text takes the mutation lock on every launch, current home or not

Quote: "Under `--repair`, resolve takes the **repair lock** — the manager-home mutation lock of manager §2.5, the same lock `profile use` holds — with a bounded wait … provisions or repairs the home … and then emits the fragment." Quote (§10.1 `curator run`): "it resolves the fragment first with `--repair`".

What is wrong: read together, every `curator run` launch acquires the exclusive manager-home mutation lock even when the lock-free verification would have found the home current. That is the M10 hot-path cost story in a new coat: a launch during `profile sync` fails `environment_lock_unavailable` although nothing needed repairing, and every launch serializes behind GC. M10's resolution is "verification is lock-free … repair takes the mutation lock" — the lock belongs to the repair, not to the flag.

Fix: one sentence in §10.1 — "Under `--repair` the lock-free verification runs first; the repair lock is taken only when that verification finds the home stale, and a current home emits its fragment without touching the lock." Keep the recorded residual as is (a home that goes stale between verification and exec is already the residual).

Launcher follow-up (recorded, not blocking, as the brief asked): the producer's fail-closed "no fragment without `--repair`" does not contradict Decision 0013 — D6.1 only orders "resolve the fragment before building the plan" — but the launcher specification `0.2.0-draft` §4.3 says resolution repairs. The launcher spec's next revision needs the sentence: "The launcher invokes `env resolve <env-id> --repair --format json`; `environment_home_stale` cannot occur on that path, and `environment_lock_unavailable` and `environment_repair_failed` are terminal launcher errors that never fall back to a launch in the native home."

### F4 — minor — §1.3: `required_by` "empty … for a machine overlay" contradicts §6 rule 4

§6 rule 4 covers "a package that is both an overlay and a requirement"; such a member has requirers. Fix: "empty for the root; for an overlay, the sorted names of any members that also require it".

### F5 — minor — §9.2 vs §9.4: shim re-pointing on scoped switches

§9.2 step 1 says `profile use` "re-points the machine command shims … to the selected profile" unconditionally; §9.4 says the shims follow the machine-current profile and re-point "on every machine-scope switch". A scoped `profile use --env pi` must not move the machine shims. Fix in §9.2: "re-points the machine command shims (section 9.4) when the switch is machine-scope".

### F6 — minor — §9.2: "skill-scope `update` and `upgrade` … act on project declarations" is wrong for `global update|upgrade`

`cli/curator.md` row 28 keeps `curator global … update|upgrade [--profile <name>|--all-profiles]` as profile-scoped verbs under this capability. The rule "never move a profile's lock" is right; the justification is not. Fix: state what `global update|upgrade` do under this capability (recommendation: fetch only, report that pins move through `profile update`) so the next batch's CLI row can copy it.

### F7 — minor — §5.8 codex bytes: the inline-array spelling of `args` is not fixed

"`args = [...]` … one key per line, TOML basic strings … and no other bytes" leaves the separator between array elements (`", "` vs `","`) and empty-array spelling unspecified, although the row claims byte exactness. Fix: fix the spelling (recommendation: `args = ["a", "b"]`, elements separated by `", "`, empty as `[]`), or say the exact bytes are fixed by the section 13 MCP byte vectors and the next batch. Same for `[mcp_servers.<name>]`: state that `<name>` is emitted as a TOML bare key (the core §2 identifier grammar fits) so no quoting rule is needed.

### F8 — minor — §5.7 hosts `mcp_command_unresolved`, a resolution-time warning

The command check runs at resolution/audit (§2.2, §9.1), not at materialization. Not a duplicate — it is tabled once — but a reader looking at §9.7 will not find it. Fix: move the row to §9.7 or add "(reported at resolution)" and leave it; either is fine.

### F9 — nit — §10.2 example vs Decision 0012 §9 worked example

The text makes `argument` a required member of every `flag` descriptor and shows it on the system-prompt descriptors; 0012's worked example omits it there. The text is normative and right; note it for the schema batch so the `launch-env-fragment-v1` schema requires `argument` on every `flag` descriptor and the 0012 example is read as pre-revision.

### F10 — nit — §9.2 `profile update` step 5 wording

"managed homes of this profile are marked stale (`environment_home_stale` at their next resolve …)" — since `curator run` always passes `--repair`, an operator will only see `environment_home_stale` from a bare `env resolve`; worth one clause so nobody expects the launcher to surface it.

## Decision 0012 impact table → verification

| Section | Disposition | Result |
|---|---|---|
| §1 | rewritten | verified: root + lock + overlays; `git` `range|tag|revision`, `directory`, no `branch`; `path` root `agent-context.json`; `local` unchanged |
| §1.1 | bytes change | verified: ref-form row; `context_range_conflict`, `context_version_mismatch`, `context_weights_duplicate`, `context_weight_unknown` (the other two weight diagnostics in §6.1 — disclosed, acceptable) |
| §1.2 | (M3) | byte-identical |
| §1.3, §1.4 | new | verified verbatim against 0012 D2/D3 (seed, highest-satisfying, downward re-selection, termination, final check, `\|\|` rule, prerelease rule, lock order (`kind`,`name`)); F4 on `required_by` |
| §2, §2.1, §2.2 | rewritten / new | verified against 0012 D1/D6: bare command, https url, `env_names` grammar and reserved set, allowlist over package identities, `mcp_declaration_invalid`, `mcp_package_not_allowed` |
| §3, §3.1 | rewritten / bytes | verified: inline `context.modules`, no `version: 1`, `context_manifest_invalid` |
| §4 | rewritten | verified |
| §5 body | rewritten | verified: tuple, emitted order under both primitives with tie rule, chapter part bytes `## Context: <name> <version>`, no-chapter rule; M5 advisory |
| §5.1 | rewritten | verified line by line against 0012 D8 |
| §5.2 | unchanged | byte-identical |
| §5.3 | bytes change | verified: `<package-name>`; N14 docs-confidence; N1 consequence (justified extension) |
| §5.4, §5.5, §5.6 | bytes change | verified |
| §5.7 | unchanged → two rows | justified (M5 needs a table); F8 |
| §5.8 | new | verified against 0012 D6; F7 |
| §6, §6.1 | rewritten | verified: rules 1–4, two primitives, joint resolution, withdrawal; N10, N4 |
| §7 body, §7.2 | unchanged | byte-identical |
| §7.1 | unchanged → changed | adapter table byte-identical; XDG paragraph and split-brain note are M4e/M14/N2 anchors — justified |
| §7.3 | bytes change | verified: `argument`, `with`, `semantics` system-prompt-only, M5 pi row, admission rule; F1 |
| §7.4–§7.7 | unchanged → changed | each change is an M4/M13/N5/N2/N3 anchor the brief placed there — justified; F2 |
| §7.8, §7.9 | new | verified; codex row F1 |
| §8.1 | bytes change | verified; M4d, N6 |
| §8.2 | rewritten | verified: `profile` name/root/kind/`lock_sha256`, `members`, `precedence` object, `mcp` surface |
| §8.3, §8.5 | unchanged → changed | M11 backups — justified |
| §8.4 | unchanged | byte-identical |
| §9.1 | rewritten | verified: one root, flags, `--use` bare, detector scope, M9, N7 |
| §9.2 | bytes change → extended | verified: M11; F5, F6 |
| §9.3 | unchanged → changed | N11 — justified |
| §9.4 | rewritten | verified: lock's skills, direct declarations, `default` 0.0.0, M6, N4 |
| §9.5 | unchanged → changed | N8, N9 — justified |
| §9.6 | rewritten | verified: `agent-context.json` 1.0.0, `revision`-pinned skills |
| §9.7 | bytes change | verified: `profile_index_ambiguous` withdrawn, new rows |
| §10.1 | bytes change → extended | M10, M16, N7, N12, 0013 sentences; F3 |
| §10.2 | rewritten | verified: `lock_sha256`, no `composition`, `precedence` object, `mcp`, `path_prepend`; F1 |
| §10.3 | bytes change | verified: `env_names` paragraph matches 0013 D6.3 without restating the collision rule divergently |
| §10.4 | unchanged → two rows | M10 — justified |
| §11, §11.1 | unchanged → changed | M6 — justified |
| §12, §12.1, §12.2 | bytes change → extended / new | verified: lock hash, GC live roots incl. retained lock, status rows, 20-knob closed list, M15 phasing |
| §13 | rewritten | verified: complete surface list |

## Review items → verification

| Item | Section(s) | Result |
|---|---|---|
| M4a seeds | §7.4 | verified: one-time, never hashed, marker-recorded, per-adapter table, docs-confidence labels |
| M4b passthrough strategy | §7.4, §7.7, §12 | verified with F2 (Linux claude row mislabeled) |
| M4c `isolated` unsupported | §7.4, §7.7 | verified |
| M4d fresh-home paragraph, first-resolve notice | §8.1 | verified |
| M4e XDG allowlist, DATA/STATE ambient | §7.1, §12.1 | verified |
| M5 pi row, admission rule, advisory | §7.3, §7.9, §5, §5.7 | verified on the binaries; codex global cap docs-confidence |
| M6 commands | §9.4, §10.2, §11, §11.1, §9.7 | verified: "no commands in managed homes" decided and stated, `path_prepend` reserved, `curator-*` reserved, `subcommand_provider_untrusted` |
| M9 detector | §9.1, §9.7, §12.1, §13 | verified: unpinnable, waiver `{pin,file,span,reason}`, classes with vector names, `context-system-module-present` |
| M10 resolve | §10.1, §10.4 | verified except F3 |
| M11 lifecycle | §9.2, §8.3, §9.7 | verified; F5, F6 |
| M12 knobs | §12.1 | verified: closed list, bootstrap shape |
| M13 `--check` | §7.5, §7.7, §12, §12.1 | verified |
| M14 XDG reconciliation | §7.1, §7.7 | verified |
| M15 lockability | §12.2 | verified: six keys, phasing statement |
| M16 CCJ-1 digest | §10.1 | verified, matches 0013 D6.4 |
| N1 | §5.3 | decided and stated |
| N2 | §7.4, §7.1, §12 | verified |
| N3 | §7.9, §7.7, §12 | verified |
| N4 | §9.4, §6 | verified |
| N5 | §7.6, §7.7, §12 | verified |
| N6 | §8.1 | decided (loud split) and consistent with §10.3 |
| N7 | §9.1, §10.1 | verified |
| N8 | §9.5 | verified |
| N9 | §9.5, §9.7 | verified |
| N10 | §6 | decided and stated |
| N11 | §9.3 | verified |
| N12 | §10.1 | verified, matches 0013 D6.4 provider column |
| N13 | §7.3, §7.4 | erratum items 1 and 3 carried |
| N14 | §5.3 | verified |

Decision 0013: D4 has no environments surface (correct); D6.3 collision rule referenced, not restated; D6.4 pin = lock hash, digest base, provider column all consistent. Decision 0010 erratum items 1 and 3 corrected in §7.3/§7.4.

## Diagnostics hygiene

63 tabled diagnostics. Duplicate rows are exactly the five the report lists and all pre-exist at `ec695ba` (`context_manifest_invalid` was `profile_context_manifest_invalid` in both §2.1 and §3.1 there). No new duplicate. Every diagnostic named in prose is tabled except the withdrawn names, which are named only in their withdrawal sentences. Absence-vs-unreadable kept (`environment_seed_unreadable`, `environment_surface_unreadable`, `profile_source_path_unreadable`, the §9.6 loss rule).

## Held under attack (review §5)

Profile-influence boundary (§10.3 unchanged in substance), closed registry, no templating (§3), byte-exact determinism (§5/§5.6), always-strict audit (§9.1), inert system prompt (§5.5/§7.3), fire-vs-manage (§2.2 manager never launches), two modes (§8.1), onboarding import (§9.6) — all still stand.
