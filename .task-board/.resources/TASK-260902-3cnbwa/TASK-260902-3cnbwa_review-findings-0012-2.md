# Review findings 0012-2: Decision 0012 after rework 1

Reviewer run RUN-260902-7a5ae9 (claude-fable-5-1), 2026-09-03. Subject:
`decisions/0012-context-packages-and-semver-locks.md` at `8444706` (rework
of `a25dc67`) on `draft/decision-0012-context-packages`, worktree
`~/Developer/ReluxWorks/.worktrees/curator-spec-decision-0012`, base
`4d55698` = main. Inputs: `review-findings-0012-1.md`,
`producer-brief-0012-rework-1.md`, `rework-report-0012-1.md`. Read-only
review; no edits, commits, or pushes.

## Verdict

**ACCEPT.** All 24 cycle-1 findings are resolved in the text, and every
resolution matches the orchestrator's author decision in the producer brief
(the rework report claims no deviation; I found none). The two additions the
producer made beyond the brief — the MCP channel-descriptor grammar
(`argument`, `with`) and the lock's `path`/`required_by` members — are both
forced by the decisions they serve and are correct. The attack pass on the
new text (resolution algorithm, range grammar, MCP policy, worked example,
impact table) produced no blocking or major finding. It produced seven
minor findings and five nits, listed below for the normative-authoring
pass that acceptance authorizes; none changes a rule, a diagnostic's
meaning, or a decision, and none is a contradiction a reader cannot resolve
from the surrounding text. Under the verdict contract (blocking/major →
development, else ACCEPT) this is an acceptance.

Commit `8444706` carries a good SSH signature by the configured author
(`git verify-commit`: "Good git signature"; principal lookup fails only
because the configured allowed-signers path is a stale temp file, exactly as
on `a25dc67` and `4d55698`). Only the decision file changed (+530/−167).

## Evidence gathered (verified on this machine, 2026-09-03)

| Claim in draft | Source checked | Result |
|---|---|---|
| node-semver coercions in Decision 2 | `semver` 7.7.4 (`/opt/homebrew/lib/node_modules/npm/node_modules/semver`), `validRange`/`satisfies`/`maxSatisfying` | `1.2`→`>=1.2.0 <1.3.0-0`; `>=2.1`→`>=2.1.0`; `>1.2`→`>=1.3.0`; `<3`→`<3.0.0-0`; `<=1.2`→`<1.3.0-0`; `^1.2.3`→`>=1.2.3 <2.0.0-0`; `^0.2.3`→`>=0.2.3 <0.3.0-0`; `^0.0.3`→`>=0.0.3 <0.0.4-0`; `^1.4`→`>=1.4.0 <2.0.0-0`; `^0.1`→`>=0.1.0 <0.2.0-0`; `^0`→`<1.0.0-0` (draft's `>=0.0.0 <1.0.0-0` is equivalent); `~1.2.3`→`>=1.2.3 <1.3.0-0`; `~1.2`→`>=1.2.0 <1.3.0-0`; `~1`→`>=1.0.0 <2.0.0-0`; `1.x`, `1.2.x`, `1` as stated; `*`/`x`/`X`→`*`; `latest`→`null` (dist-tag, not grammar — correctly labeled a Curator spelling). **Every coercion in the draft matches.** |
| README section names cited | semver 7.7.4 `README.md` | "Prerelease Tags", "Hyphen Ranges", "X-Ranges", "Tilde Ranges", "Caret Ranges" all exist as headings. Holds. |
| prerelease rule | same | `2.0.0-rc.1` satisfies `^2.0.0-rc.0` and `>=2.0.0-rc.0`; `2.1.0-rc.1` satisfies neither; `2.0.0-rc.1` satisfies none of `*`, `>=1.0.0`, `<3`; `3.0.0-rc.1` does not satisfy `<3`; `maxSatisfying([1.9.0, 2.0.0-rc.1], '*')` = `1.9.0`. Holds. |
| exclusions | same | `1.2.3 - 2.3.4` and `^v1.2.3` are accepted by npm; the draft excludes both explicitly and says so. Holds. |
| `\|\|` highest-satisfying-member | same | `maxSatisfying([1.5.0,2.0.0,3.1.0], '1.x \|\| 3.x')` = `3.1.0`. Matches step 2. |
| total order | same | `1.0.0-alpha < 1.0.0-alpha.1 < 1.0.0-beta < 1.0.0`. Holds. |
| build metadata | same | `eq('1.2.3+a','1.2.3+b')` = true — the F3 counter-example; the draft now excludes `+build` tags as candidates, which removes it. Holds. |
| worked-example intersections | same | `3.2.1` satisfies `^3.0 ^3.1` (= `>=3.1.0 <4.0.0-0`); `4.3.0` satisfies `^4 ^4.2` (= `>=4.2.0 <5.0.0-0`); `1.2.5` ~ `~1.2`; `1.6.0` ~ `^1.4`; `2.4.2` ~ `>=2.1 <3`; `1.1.0` and `1.0.4` ~ `^1.0`; `1.2.0` ~ `^1`. Every lock version in §9 satisfies its stated constraint. |
| `claude --mcp-config` / `--strict-mcp-config` | `claude --help`, **2.1.259** installed | `--mcp-config <configs...>` "Load MCP servers from JSON files or strings (space-separated)"; `--strict-mcp-config` "Only use MCP servers from --mcp-config, ignoring all other MCP configurations". Holds (draft cites 2.1.258; see N11). |
| codex `-p` layering | `codex --help`, 0.151.0 | `-p, --profile <CONFIG_PROFILE_V2>  Layer $CODEX_HOME/<name>.config.toml on top of the base user config`. Holds; composition of an `mcp_servers`-only layer and double-`-p` precedence remain unverified, as Open question 3 says. |
| `pi` has no MCP channel | `pi --help`, 0.84.2 | No MCP flag or subcommand. Holds. |
| opencode merge order | opencode.ai/docs/config (opencode not installed) | "Config sources are loaded in this order (later sources override earlier ones)": remote, global, custom (`OPENCODE_CONFIG`), project `opencode.json`, `.opencode` directories, inline (`OPENCODE_CONFIG_CONTENT`), managed config files, macOS managed preferences; "merged together, not replaced". The draft's order is a correct prefix of this list. Holds. |
| "revision 1 … claimed by no implementation and carried by no tag" | `git tag --contains eddd509` (the commit that added environments.md) | No tag contains it; latest tag `v1.0.0-rc.10`. Holds. |
| `works.relux.curator.profile-pin` (ax PR #1) | `agent-session-manager-spec`, PR #1 branch `draft/curator-environment-integration`, `SPEC.md` | Key defined there alongside `profile-name` and `fragment-digest`. Holds. |
| `Profilefile.json`/`context.json` never in COMPATIBILITY.md | `COMPATIBILITY.md` at `4d55698` | No mention. Holds. |
| every § citation | core §2, §4.4, §6.1, §6.3, §7, §10; environments §1, §3, §5.5, §7.3, §9.1; manager §1, §3.1, §6; `registry.md` §1; Decision 0010 Decisions 2, 5, 8; review M1, M5, M9, M10, M12, M15 | Every citation resolves to text that says what the draft attributes to it. |
| nine vector sets named in the table | `conformance/v1/expected/environments/` | Exactly the nine directories listed. Holds. |

## Cycle-1 findings: resolution check

| # | Sev | Resolved | Matches author decision | Where verified |
|---|---|---|---|---|
| F1 | blocking | yes | yes | Compatibility impact table, §1–§13 plus manager §12.1–§12.7, CLI rows, schemas, schema-cases, nine vector sets. Rechecked row by row against the landed revision 1 (see N3, N4 for the two residues) |
| F2 | major | yes | yes | Decision 2 "Resolution" steps 1–4; walked below |
| F3 | major | yes | yes | Decision 2 "Versions": `+build` not a candidate; uniqueness argument |
| F4 | major | yes | yes | Decision 1 `version` bullet; Decision 2 exact-constraint sentence; `path` authoritative; `context_version_mismatch` |
| F5 | major | yes | yes | Decision 4 rules 2–3, `context_weight_conflict`, `context_weights_duplicate`, root edge = map entry |
| F6 | major | yes | yes | "Weights order chapters and nothing else in this revision"; collision sentence gone |
| F7 | major | yes | yes | Status, Context, Decision 2 "Skills in the closure", Compatibility opening, Rejected "Ranges in project Skillfile.json now", Open question 6; `grep 'schema 9'` empty |
| F8 | major | yes | yes | Decision 2 "Ranges"/"Prereleases"; every coercion verified above; hyphen and in-range `v` excluded; `-0` bound; `latest` labeled |
| F9 | major | yes | yes | Decision 6: package-identity allowlist, bare `command`, `https` grammar, `args`/`url` in detector scope, `env_names` bounds; Security impact; Rejected "allowlist over launcher binaries" |
| F10 | major | yes | yes | Decision 6 last paragraph before policy; Decision 8 fragment `mcp`; §9 fragment shows `env_names` |
| F11 | major | yes | yes | `grep 'revision 2'` empty; Status, Compatibility, `generated:` line all say revision 1 |
| F12 | major | yes | yes | `grep 'Decision 0011'` empty; Context carries the M1/Option A/next-free-number wording |
| F13 | major | yes | yes | Decision 9: lock, header, fragment, MCP file bytes, exec line; internally consistent (checked below) |
| F14 | minor | yes | yes | `environments` selector on `agent-mcp.json`; absent = every adapter |
| F15 | minor | yes | yes | Decision 4 (chapter per member with applicable modules), Decision 8 (`member:` per context member), vector note |
| F16 | minor | yes | yes | Decision 1 `requires` bullet |
| F17 | minor | yes | yes | Decision 4 tie sentence |
| F18 | minor | yes | yes | Decision 4 `context_weights_not_root` at resolution time; reuse when map empty/absent |
| F19 | minor | yes | yes | Decision 3 pin shape and (`kind`, `name`) sort |
| F20 | minor | yes | yes | Decision 7 and Rejected "Branch-tracking profile roots" |
| F21 | minor | yes | yes | Decision 6 channel bullets |
| F22 | minor | yes | yes | Consequences |
| F23 | nit | yes | yes | Compatibility closing paragraph and table rows |
| F24 | nit | yes | yes | Decision 1 `weight` bullet |

## Attack results on the new text

**F2 algorithm.** Walked on: root R requires A `^1`, B `^1`; A@1.5 requires
B `<1.3`; B@1.2 requires A `~1.4`; candidates A {1.4.2, 1.5.0}, B {1.0.0,
1.2.0, 1.5.0}. Seed: A, B pending. Take A: select 1.5.0, expand → B gains
`<1.3`. Take B: `^1 ∩ <1.3` → 1.2.0, expand → A gains `~1.4`, which
excludes 1.5.0 → step 3: A re-selected to the highest candidate ≤ 1.5.0
satisfying `^1 ∩ ~1.4` = 1.4.2; A@1.5.0's `<1.3` on B is dropped; A@1.4.2
re-expanded (say, B `^1`). B's constraint set changed → pending; B's
selection stays 1.2.0 by the "never increases" rule even though 1.5.0 now
satisfies. Step 4: R: A ✓ B ✓; A@1.4.2: B `^1` ✓; B@1.2.0: A `~1.4` ✓. Lock
{A 1.4.2, B 1.2.0}. Deterministic, terminates, every constraint satisfied.
The non-maximal B is the stated cost of no backtracking and is predictable
from the manifests. Termination: with selections fixed, constraint sets only
grow (attributions are set-valued, re-expansion at the same (name, version)
adds nothing); selections only decrease; both finite. Sound. `||`: "highest
candidate satisfying any member" equals `maxSatisfying` semantics
(verified). The final check makes "no backtracking cannot produce a lock
that violates a requirement" true by construction: any residual violation
is `context_range_conflict`. One gap in step 2's failure rule: N5.

**F1 impact table.** Every environments §1–§13 row and every manager,
schema, and vector row rechecked against the landed text. Two residues:
§9.2 is marked unchanged but carries a sentence the new model changes (N3),
and the two MCP diagnostics have no diagnostics-table home (N4). Everything
else is correctly classified; in particular §5.2, §5.7, §7 body, §7.1,
§7.2, §7.4–§7.7, §8.3–§8.5, §9.3, §9.5, §10.4, §11, §11.1, manager §12.4
are genuinely untouched by the model.

**F8 grammar.** Verified line by line (table above). No deviation from
node-semver remains; the two exclusions are stated as exclusions.

**F9/F10 MCP.** Attacks: (1) allowlist over `npx` — closed, allowlist is
over package identities and the rejected alternative names why; (2)
absolute `command` — closed, `mcp_declaration_invalid`; (3) values in
`args`/`url` — closed, detector scope plus `https`-only/no-userinfo/no-
query/no-fragment; (4) `env_names` choosing `NODE_OPTIONS`/`LD_PRELOAD` —
closed by the reserved-set exclusion, with one precision gap (N6); (5) an
overlay from an operator-local directory requiring an MCP package —
requirements carry a canonical git source, so the package allowlist still
applies (local sources bypass only the *source* allowlist for themselves);
(6) MCP into a native or secondary-target home — closed by "managed home
only"; (7) `--strict-mcp-config` semantics — matches help text. The
fragment `mcp` section carries the `env_names` union and curator-run is
named as the actor in Decisions 6 and 8. One shape gap on the descriptor
(N2) and one location contradiction (N1).

**F13 worked example.** Lock: members sorted by (`kind`, `name`) bytewise —
verified (`context` < `mcp` < `skill`; `…-core` < `…-developers-core` <
`…-developers-figma` < `…-developers-ios` < `…-ios-developer-umbrella` <
`…-organizational-structure` < `personal`; `pdf` < `swiftui`); `required_by`
sorted; pins `commit` for git members and `state_sha256` for the `path`
overlay; `source` absent for the `path` overlay; `path` present only on the
figma subdirectory package; every version satisfies its constraint
(verified above). Header: emitted in ascending weight (0, 10, 20, 40, 60,
100, 1000) with the heaviest last under `winner=higher-weight
placement=winner-last`, root as an ordinary node, `overlay` suffix on
`personal`, `precedence:`/`lock:`/`generated:`/`notice:` lines in the
Decision 8 order. Fragment: `profile.lock_sha256` equals the header `lock:`
hash; `precedence` object; `system_prompt` present (the umbrella has a
`class: system` module for `claude_code`); `mcp` with path, `env_names`,
channel. MCP file bytes: CCJ-1 key order `args` < `command` < `type`,
no `env`. Internally consistent. Two nits (N8, N9).

**F11/F12.** `revision 2` and `Decision 0011` absent (grep). The `generated:`
line, Status, Compatibility, and Open questions all read "revision 1". The
numbering note is present in Context.

**Regression sweep.** Untouched sections re-read; nothing regressed. M4
seeds, M6 `path_prepend`, M9 unpinnable detector (draft keeps "always-strict
audit … blocking finding"), M10 (`managed homes are marked stale for
explicit repair`), M12 knob list, M15 lockability (OQ4), M16 CCJ-1 (lock
hash over CCJ-1) — all consistent. House style matches 0009/0010 (same
section set, tone, density); English throughout; lines ≤ 79 columns outside
code, tables, and the title.

## New findings (for the normative-authoring pass)

Severity: minor | nit. Section = draft section.

### N1 — minor — Decision 6 — MCP file location contradicts the codex channel

> "materializes as one inert, hashed, marker-recorded file per adapter
> format below `<home>/.agent-context/mcp/`" vs. "`codex_cli` — … the
> manager writes `<home>/curator-mcp.config.toml`"

Codex fixes the layer location at `$CODEX_HOME/<name>.config.toml`, which is
`<home>/curator-mcp.config.toml`, not below `.agent-context/mcp/`. The
general sentence and the codex bullet cannot both hold. **Fix:** "below
`<home>/.agent-context/mcp/`, except where the tool fixes the location — the
codex layer file at `<home>/curator-mcp.config.toml`"; the §5.8 row and the
§8.2 `mcp` surface entry then record a tool-fixed path for that adapter.

### N2 — minor — Decision 6, §9 fragment — MCP channel descriptor has no `semantics`

> `{ "kind": "flag", "flag": "--mcp-config", "argument": "path", "with": ["--strict-mcp-config"] }`

Environments §7.3: "its `semantics` is exactly `append` or `replace`";
§10.2: readers reject unknown semantics values. The draft says the MCP
descriptor "reuses the environments §7.3 descriptor grammar" and then emits
one without `semantics`. **Fix:** state that `semantics` is a system-prompt
member and is absent on an MCP channel descriptor (or give MCP a value), and
add that to the §7.3 and §10.2 impact rows.

### N3 — minor — Compatibility impact — §9.2 is marked unchanged but changes bytes

> §9.2: "re-materializes every in-place surface … from the selected
> profile's **store entry**, atomically per entry"

Decision 3: a profile's store identity is the *set* of entries its lock
names. The table changes this same phrase in §8.1 ("the same store entry"
→ "the same lock's store entries") but leaves §9.2 at "unchanged".
**Fix:** §9.2 → bytes change, "store entry" → "the store entries its lock
names".

### N4 — minor — Compatibility impact — MCP diagnostics have no table home

`mcp_declaration_invalid` and `mcp_package_not_allowed` are introduced in
Decision 6 but no row of the impact table says which diagnostics table
receives them (the §1.1 row lists the range, version, and weight
diagnostics; the §2.1 row the manifest ones). **Fix:** name the table —
§2.1 alongside `context_manifest_invalid`, or the new §5.8/§9.1 — so
acceptance authorizes exactly the rows that will be written.

### N5 — minor — Decision 2, step 2 — an unsatisfiable non-empty constraint names no outcome

> "Compute its effective constraint; if empty, fail `context_range_conflict`.
> Select the highest candidate satisfying it"

"Empty" reads as the abstract intersection (`>=2 <2`). A non-empty
constraint that no candidate satisfies — `^5` against a source whose
highest tag is `v4.9.0`, or a source with no version tags at all — is the
common failure and falls between the two sentences. **Fix:** "if no
candidate satisfies it, fail `context_range_conflict` naming every requirer
with its range or exact form and the candidate versions considered" (one
diagnostic covers both), or a distinct `context_range_unsatisfied`.

### N6 — minor — Decision 6 — the `env_names` reserved set is under-specified

> "MUST NOT name a manager-reserved variable (the manager §3.1 reserved set
> — `PATH`, `HOME`, the `LD_`/`DYLD_`/`NODE_` families, and the rest)"

Manager §3.1 reserves per platform and per interpreter identifier: `NODE_`
only for `node-v1`, `PYTHON` only for `python3-v1`, `DYLD_` only on macOS.
An MCP declaration has no interpreter identifier and is platform-portable,
so "the §3.1 reserved set" does not pick one set. The parenthetical implies
the union. **Fix:** say it: "the union of every §3.1 reserved set across
platforms and interpreter identifiers".

### N7 — minor — Decisions 1, 3, 5, 7 — `path` names three different things

Source kind `path` (environments §1: an operator-local directory), the
requirement member `path` (a subdirectory within a git snapshot), and the
install flag `--path <dir>`. Decision 3 uses both in one sentence: "its
`path` within the snapshot … A `path` member's source path". The lock
schema will have to carry a subdirectory member and a member kind that are
spelled identically. The requirement spelling was the orchestrator's F16
decision, so this is recorded, not required. **Fix (recommended):** rename
the subdirectory member (`directory`, or `subdir`) in `requires` and the
lock, or add one sentence in Decision 1 defining the two spellings side by
side.

### N8 — nit — Decision 7/8 — the `default` root's version and surface

`root: <name> <version> <pin>` and the marker record a root version; the
synthesized `local` root of `default` has no manifest and therefore no
version. Revision 1 also wrote no root-context file for a profile without
`context/`; the draft's "a package with no modules of its own is a pure
umbrella" plus "a pure umbrella contributes a `member:` line and no
chapter" leaves it unstated whether a root with no `context` writes the
header-only file (§5.4) or no file (§2). **Fix:** say `default` records
version `0.0.0` (or `-`) and which of the two surfaces rules applies to a
context-less root.

### N9 — nit — §9 example — `developers-core` weight `20` is unexplained

The narrative explains the core (manifest `0`), the organizational
structure (`weights` map `10`), and the overlay (`1000`); `20` for
`developers-core` must be its manifest weight but is not said. One clause.

### N10 — nit — Decision 5 — "initially above any root weight"

`weight` admits up to 2147483647 and Open question 1 recommends the
constant `1000`; the default cannot be above *any* root weight. **Fix:**
"above the weights roots use in practice, so that…".

### N11 — nit — Decision 6 — verified release

"both verified in 2.1.258 help": cycle-1 verified on 2.1.258, the rework
and this cycle on 2.1.259; both carry the flags. Record the release the
vectors will freeze against, per the §7.3 discipline.

### N12 — nit — Decision 2 — `X` omitted once

"`*`, `x`, and `latest` select the highest stable version" — the preceding
sentence admits `X` too.

## Change Request emptiness

`CR-TASK-260902-3cnbwa-1` has `repository_delta=empty` on
`task-board/story/STORY-260902-le61cp` (tree `9bf1cc5` = base `4d55698`).
That is the correct outcome for this leaf: TASK-260902-3cnbwa is a review
task whose deliverables are board resources (`review-findings-0012-1.md`,
`rework-report-0012-1.md`, and now this document), and the reviewed
document lives by the orchestrator's design on
`draft/decision-0012-context-packages` in a separate worktree, named in
every brief. No file on the story branch was supposed to change. Note for
integration: the story branch does not carry `8444706`; landing Decision
0012 is a separate orchestrator step from this task.

## Board note

The story description still says "environments.md revision 2 impact"; the
F11 author decision (keep revision 1, bump only the type line) is what the
document now says. Worth aligning the description when the story closes.
