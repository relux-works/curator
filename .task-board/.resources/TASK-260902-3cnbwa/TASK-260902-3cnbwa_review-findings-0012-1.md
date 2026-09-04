# Review findings 0012-1: Decision 0012 draft (context packages, semver locks, launch-channel MCP)

Reviewer run RUN-260902-ffb20b (claude-fable-5-1), 2026-09-02. Subject:
`decisions/0012-context-packages-and-semver-locks.md` at `a25dc67` on
`draft/decision-0012-context-packages` (worktree
`~/Developer/ReluxWorks/.worktrees/curator-spec-decision-0012`), base
`4d55698` = main. Read-only review; no edits, commits, or pushes.

## Verdict

**Changes requested → `development`.** One blocking finding (F1: the
environments.md impact list is wrong for ten sections, and acceptance
authorizes rewriting exactly the named sections) and twelve major findings,
most of them soundness gaps in the resolution and weight rules that can be
read two ways or are contradicted by a counter-example. The model itself —
context package as a second kind on the closure engine, lock as identity,
umbrella as convention, weights never resolving versions, declaration-only
MCP into managed homes, two closed precedence primitives — held under attack
and should be kept. The document needs one more producer pass, not a
redesign.

## Evidence gathered (verified on this machine)

| Claim in draft | Source checked | Result |
|---|---|---|
| `claude --mcp-config <path>` + `--strict-mcp-config` | `claude --help`, 2.1.258 | Both exist: `--mcp-config <configs...>` "Load MCP servers from JSON files or strings (space-separated)"; `--strict-mcp-config` "Only use MCP servers from --mcp-config, ignoring all other MCP configurations". Holds. |
| codex `-p <name>` layers `$CODEX_HOME/<name>.config.toml` | `codex --help`, 0.151.0 | `-p, --profile <CONFIG_PROFILE_V2>  Layer $CODEX_HOME/<name>.config.toml on top of the base user config`. Holds; whether an `mcp_servers`-only layer composes cleanly is unverified (draft OQ3 says so — correct). |
| `OPENCODE_CONFIG` names a config file whose `mcp` member is read | opencode.ai/docs/config (opencode not installed locally) | Documented ("Specify a custom config file path using the `OPENCODE_CONFIG` environment variable"); configs are **merged**, precedence remote → global → custom (`OPENCODE_CONFIG`) → project `opencode.json` → `.opencode/` → `OPENCODE_CONFIG_CONTENT` → managed. `mcp` key documented. Holds, with a precedence caveat (F21). |
| `pi` — no MCP channel | `pi --help`, 0.84.2 | No MCP flag or subcommand; extensions only. Holds. |
| npm prerelease rule | node-semver 7.7.4 README "Prerelease Tags" + `satisfies()` | `2.0.0-rc.1` does not satisfy `*`, `>=1.0.0`, `<2.0.0`; satisfies `>=2.0.0-rc.0`, `^2.0.0-rc.0`; `2.1.0-rc.1` does **not** satisfy `^2.0.0-rc.0`; `maxSatisfying([1.9.0, 2.0.0-rc.1], '*')` = `1.9.0`. Draft's rule matches npm. |
| `latest` ≡ `*` | node-semver 7.7.4 | `validRange('latest')` = `null`: `latest` is an npm dist-tag, not range grammar. The draft's spelling is a Curator extension (F8). |
| caret/tilde/partial semantics | node-semver 7.7.4 `validRange` | `^0.2.3`→`>=0.2.3 <0.3.0-0`, `^0.0.3`→`>=0.0.3 <0.0.4-0`, `^3.0`→`>=3.0.0 <4.0.0-0`, `~1.2`→`>=1.2.0 <1.3.0-0`, `>=2.1 <3`→`>=2.1.0 <3.0.0-0`, `1.2.x`→`>=1.2.0 <1.3.0-0`, `1.2.3 - 2.3.4` accepted (hyphen range). The draft's grammar text does not define any of these (F8). |
| build metadata ignored for equality | node-semver 7.7.4 | `eq('1.2.3+build','1.2.3')` = true, `compare` = 0. Consequence: F3. |
| Decision 0011 exists | repo `decisions/`, main and all worktrees, board resources | No `0011-*execution*` document anywhere. `decisions/0011-swift-driver-pair.md` exists in worktree `curator-spec-draft-swift`. The only text is pre-implementation-review-v3 §M1 "decide as Decision 0011" with Option A recommended; story notes say "Pending operator". (F12) |
| every § citation | core.md, environments.md rev 1, manager.md, registry.md | Verified: core §2, §4, §4.4, §6.1, §6.3, §7, §10; environments §1, §3, §5.5, §9.1; manager §1, §6; `registry.md` §1 (file is `protocol/registry.md`, cited the same way by environments.md). One misattribution: "the skill `source` convention" (F16). |

## Findings

Severity: blocking | major | minor | nit. Section = draft section.

### F1 — blocking — Compatibility impact — environments.md impact list is wrong for ten sections

> "`protocol/environments.md` moves to revision 2: §2 (repository shape), §6 (composition), §9.1 (installation), and §9.4 (skills) are rewritten to this decision; §5.1 and the chapter part change bytes under the `curator-root-context-v2` type line; every other section stands."

Checked section by section against the landed revision 1. The following
sections cannot stand as written and are not named:

- **§1** — "an OPTIONAL profile-scoped `Skillfile.json` … and no other surface in revision 1. Surfaces not defined by this revision (MCP servers, …) MUST NOT be declared, materialized, or inferred from profile data"; `git` "carries exactly one of `tag`, `branch`, or `revision`" (no `range`); `path` "names a directory whose root contains `Profilefile.json`". Every clause contradicts Decisions 1, 2, 5, 6, 7.
- **§3** — is titled and specified as `context/context.json`, "a strict schema-1 object" with `version`, own diagnostics `profile_context_manifest_invalid`. Decision 1 moves the manifest inline into `agent-context.json`; the section's object, location, and diagnostics change even if the entry shape is "unchanged".
- **§4** — "exactly one store entry below the manager home, keyed by its effective pin: the resolved commit for a `git` profile". Decision 3: per-package commit-keyed entries, profile identity = set of entries the lock names.
- **§5** (body, not only §5.1) — the pure-function tuple "(profile store entry, composition chain, environment identifier, form)" becomes (lock, …); part sequence item 3 "for each profile of the composition chain in chain order — a chapter part … emitted for every composed profile, including one whose applicable module set is empty" becomes one chapter per closure member ordered by weight under `placement` (and Decision 4 says only members "with modules"); the chapter part bytes `## Profile: <name>` → `## Context: <name> <version>`; **§5.3** layout `.agent-context/modules/<profile-name>/<module-path>` is grouped per *profile*, now per package name; **§5.4** "the header followed by the empty chapters" changes with the chapter rule; **§5.5** "of every profile in chain order" becomes weight order; **§5.6** hash tuple changes. This is a rewrite of §5, not a byte change in §5.1.
- **§7.3** — channel descriptors gain the MCP channels of Decision 6; the `flag` descriptor shape has one `flag` per descriptor, while `claude_code` needs two flags together and `codex_cli`'s `-p <name>` argument is a name, not a path (review M5 asked for `argument: path | contents`; a third kind is needed). §7 changes.
- **§8.2** — the marker records "the effective pin (`commit` for `git`, `state_sha256` for `local` and `path`)" and `composition` with member pins. Decision 8 replaces this with root + lock hash + member list. The Compatibility paragraph admits "the environments schemas for the marker … revise", so the section list is internally inconsistent.
- **§9.6** — reassembly emits "`Profilefile.json`, version 1, declaring exactly one profile", `context/<env-id>.md` under a `context.json` manifest, and "one `Skillfile.json` declaration per mapping skills entry". All three identifiers are withdrawn by Decision 7; the import must reassemble an `agent-context.json` with `requires.skills` (pinned by `revision`).
- **§10.2** — fragment `profile` carries `commit`/`state_sha256`, `composition` members carry pins, and there is no `mcp` section. Decision 8 changes all three; again admitted in the Compatibility paragraph but absent from the list.
- **§12** — `profile list` "declared ref, effective pin"; `env status` "materialized pin" — pin becomes lock hash; declared ref becomes a range.
- **§13** — names "`Profilefile.json` schema 1, `context.json` schema 1" as conformance surfaces; they are withdrawn.

Also affected outside environments.md: manager §12.3 (install flags, "every profile the snapshot declares"), §12.6 (detector scope "`Profilefile.json`, `context.json`, and `PROFILE.md`"), §12.7; and `schemas/v1/profilefile-v1`, `context-manifest-v1`, their schema-cases, and the nine `conformance/v1/expected/environments/*` vector sets (all carry `## Profile:` chapters and `curator-root-context-v1` headers).

**Fix:** replace the sentence with an honest per-section table (rewritten / bytes change / unchanged) covering §1–§13 and the manager §12 subsections; keep "every other section stands" only for the sections that truly do (§9.2, §9.3, §9.5, §11, and the §7 subsections other than 7.3 as far as I can see — §7.4/7.5/7.6 stand). The AC for the story explicitly demands this list; acceptance authorizes rewriting exactly what it names.

### F2 — major — Decision 2 — resolution algorithm is circular as stated

> "Resolution is the §7 closure with one change to its admission rule. For each package name across the whole closure … the effective constraint is the intersection of every declared range … The manager … selects the highest version satisfying the effective constraint".

Core §7 processes fixed commits: "Processing a provider adds its skill requirements". With ranges, a package's requirement set depends on which version was selected, and the set of constraints on a name depends on which packages (at which versions) are in the closure. "Intersection of every declared range across the whole closure" therefore names a fixpoint the text never defines: a requirement discovered later can exclude an already-selected version, whose manifest contributed the very requirements that led there. Nothing says whether the manager re-selects and re-expands, fails, or terminates. `||` disjunctions make a greedy pass order-dependent.

**Fix:** state the algorithm. Recommended minimal shape: (1) seed with root + overlays; (2) for each name, select the highest tag satisfying the current intersection and expand *that version's* manifest; (3) when a newly added constraint excludes a selected version, re-select the highest remaining candidate for that name, drop the requirements its previous version contributed, and re-expand; (4) terminate because each name's selection only ever decreases; (5) final check: every requirement in the lock is satisfied by the locked version, else `context_range_conflict`. Say that no backtracking across names is performed and that a `||` disjunction is satisfied by its highest satisfying member. Alternatively state the simpler rule "a later-discovered constraint that excludes an already-selected version is `context_range_conflict`" — but say which.

### F3 — major — Decision 2 — "ties are impossible" has a counter-example

> "Ties are impossible (versions are a total order after discarding build metadata)."

Tags `v1.2.3+a` and `v1.2.3+b` are two refs, two commits, one version after discarding build metadata (node-semver: `compare('1.2.3+build','1.2.3') = 0`). Both are candidates; the "highest" is not unique.

**Fix:** either forbid build metadata in version tags (simplest, consistent with "strict"), or define `context_version_ambiguous` when two candidate tags of one source parse to an equal version and point at different commits (equal commits unify per §7).

### F4 — major — Decision 2 — exact forms have no version to intersect

> "the effective constraint is the intersection of every declared range with every exact `tag`/`revision` form treated as a single-version range"

A `revision` is a commit, not a version. A `tag` that does not parse "is not a version candidate … it remains addressable by the exact `tag` form" — so it has no version either. Intersecting either with `^1.4` is undefined. Related: "`version` … MUST equal the version of the tag the package was resolved from" is unevaluable for a `revision`-resolved package, a non-semver tag, and a `path` overlay.

**Fix:** define: a `revision` or non-version `tag` requirement contributes the version of the manifest at that commit (already required to be present) as its single-version range, and the resolved commit MUST be the commit that the source's tag `v<version>` peels to, else `context_range_conflict` (or drop the equality check for non-tag resolution and say so). For `path` packages state that the manifest `version` is authoritative and no tag check applies.

### F5 — major — Decision 4 — effective weight is ambiguous with several direct requirers

> "its manifest `weight`, overridden by the `weight` its **direct requirer** declares on the requirement edge, overridden by the root package's `weights` map"

In the example, `developers-ios` may be required by the umbrella (weight 60) and by `developers-core` (say weight 20): two direct requirers, two edge weights, one member. The singular "its direct requirer" does not pick one. Second ambiguity: the root is itself a direct requirer; its edge weight and its `weights` map can both name a package — which wins, and why have both?

**Fix:** one rule, e.g.: edge weights from distinct direct requirers MUST agree, else `context_weight_conflict`, unless the root `weights` map names the package (root's final word); a root edge weight is treated as a `weights` entry (or forbid duplicating a package in both).

### F6 — major — Decision 4 — the only named "mechanical collision" cannot occur

> "Mechanical collisions — two members declaring the same MCP server name with differing declarations — are resolved for the member that `winner` favors and reported."

An MCP server name is the MCP package `name` (`agent-mcp.json` has no separate server name). Two requirements for one name must agree on source identity and resolve to one commit under §7 as amended, so two *differing* declarations for one name are a `context_range_conflict`/§7 failure, never a weight question. Skills are now range-resolved too. "Weights decide mechanical collisions" is thus a rule with no defined domain.

**Fix:** either enumerate the real collision classes weights decide (I could not find one in revision 2 — referenced-form paths are per-package, MCP and skills are name-unified) or delete the sentence and state plainly that in revision 2 weights order chapters and nothing else.

### F7 — major — Compatibility impact — project `Skillfile.json` gains ranges with no lock

> "core §4.4 `dependencies.skills` and `Skillfile.json` admit `ref.kind: "semver"` with `range`, delivered as skill-manifest schema 9 and Skillfile schema 2"

The lock is defined for profiles only (Decision 3: "The result of resolution is the profile lock … below the manager home"). A project `Skillfile.json` schema 2 with `^4` and no project lock re-resolves on every operation — exactly the "Range resolution without a lock" alternative the draft rejects. The same applies to a schema-9 skill manifest with a range dependency installed by a project.

**Fix:** either (a) scope semver refs to context-package `requires` and the profile lock in this decision and drop the Skillfile schema 2 / manifest schema 9 widening (recommended: it is the minimal correct compatibility path and keeps core frozen), or (b) define the project lock (identifier joining core §1.1, location, marker v5 relationship, `Skillfile.dev.json` interaction).

### F8 — major — Decision 2 — range grammar is "closed" but undefined for the forms the example uses

> "The range grammar is closed and npm-shaped: an exact version (`1.2.3`, meaning `=1.2.3`); caret `^1.2.3`; tilde `~1.2.3`; comparator sets … the wildcards `*` and `x` in a component; and the spelling `latest`, equivalent to `*`."

The example uses `^3.0`, `^1.4`, `^4`, `~1.2`, `>=2.1 <3`, `^1` — partial versions — which the grammar text does not admit or define. Undefined and needed: partial-version coercion (`>=2.1`→`>=2.1.0`, `<3`→`<3.0.0-0`), caret on 0.x (`^0.2.3`→`<0.3.0-0`, `^0.0.3`→`<0.0.4-0`), tilde with fewer components, whether hyphen ranges (`1.2.3 - 2.3.4`, npm grammar) are excluded, whether a `v` prefix is admitted inside a range, and the `-0` upper-bound spelling that makes `3.0.0-rc.1` fall outside `<3`. Also `latest` is not npm range grammar (`validRange('latest')` = null; it is a dist-tag), so "npm-shaped" plus `latest` misleads.

**Fix:** define each (cite node-semver README "Caret Ranges", "Tilde Ranges", "X-Ranges", "Prerelease Tags" as the reference semantics), state that hyphen ranges are excluded and `v` is not admitted in ranges, and call `latest` a Curator spelling.

### F9 — major — Decision 6 / Security impact — the MCP allowlist is bypassable as designed; "values never appear" is not enforced by the shape

> "an allowlist of `stdio` commands and `http` hosts, lockable, so an organization can bound what its profiles may declare" / "`env_names` lists environment-variable **names** … values never appear in any package"

Attacks: (1) the draft's own example is `npx -y figma-developer-mcp`; allowlisting `npx` (or `uvx`, `node`, `python`, `sh`) admits any package or script through `args`, so the allowlist bounds nothing an organization cares about; (2) `command` admits an absolute path, so `/bin/sh` is in scope; (3) `args` and `url` are free text — `--token=…` or `https://user:secret@host/?key=…` carry values through a package, lock, and materialized file; (4) `env_names` lets package bytes choose which operator environment variables reach the child (`NODE_OPTIONS`, `PATH`, `LD_PRELOAD`); under Option A's SpawnPlan name-allowlist this is package influence over how a process is launched, which environments §10.3 forbids.

**Fix:** allowlist over (command, leading-args) patterns or per-package identity, or state honestly that the allowlist bounds the launcher binary only; define `url` grammar (https only, no userinfo, no query) and put `args`/`url` under `context-secret-material`; bound `env_names` by grammar and by a manager/system allowlist of passable names, and record that a `system` config can lock it.

### F10 — major — Decision 6 — `env_names` do not reach an ax-launched child under Decision 0011 Option A

> "the operator's session environment supplies them"

Under the Option A the draft relies on, `ax` spawns from a SpawnPlan with an environment-name allowlist and literal values (Decision 0010 Context; review M1). A direct exec inherits the session environment; an ax-tracked launch does not unless curator-run adds the names. The fragment's `mcp` section as described names only the file and the channel.

**Fix:** the fragment `mcp` section carries the sorted union of `env_names`; the launcher specification adds them to the plan's allowlist; say that curator-run is the actor (single composer, Option A).

### F11 — major — Decision 6, Security impact — revision numbering contradicts itself

> "Revision-1 channels: …" / "never into a native in-place home … in revision 1" / "MCP declarations … never reach a native home in revision 1" vs Compatibility impact "`protocol/environments.md` moves to revision 2".

If the document moves to revision 2, every "revision 1" in Decisions 6 and Security impact is wrong or means something else. The operator's "MCP in revision 1" answer meant the first shipped revision; the draft chose to renumber, so it must follow through.

**Fix:** either keep environments at revision 1 (defensible: "claimed by no implementation and carried by no tag") and change only the type line, or say "revision 2" everywhere.

### F12 — major — Context, Status — Decision 0011 is cited as an existing decision; it does not exist and the number is contested

> "the execution plane composes a launch through a caller-supplied plan into `ax` (Decision 0011, Option A)"

No `decisions/0011-*` execution-ownership document exists on main, in this worktree, or on the board; the only text is review v3 §M1 ("decide as Decision 0011", Option A recommended, story notes "Pending operator"). `decisions/0011-swift-driver-pair.md` already exists in the `curator-spec-draft-swift` worktree, so the number may collide.

**Fix:** cite "the pre-implementation review's M1 resolution, Option A (to be recorded as its own decision)" or land that decision first; reconcile numbering with the swift-driver draft.

### F13 — major — whole document — the story AC's end-to-end companyA example is missing

The AC reads "end-to-end companyA example". The draft shows one manifest. There is no lock excerpt, no `curator-root-context-v2` header for the example, no fragment with `profile.lock_sha256` and an `mcp` section, and no `.agent-context/mcp/` file bytes for one adapter. A reader cannot check Decisions 3, 4, 6, 8 against each other without them.

**Fix:** add a short worked example: lock members with weights and chain, the header lines in emitted order under the default policy, one fragment.

### F14 — minor — Decision 6 — `required_in` reuses a core §4.4 name with different semantics

Core §4.4: `required_in` is `any` (default) or `all`. The draft's example gives a list of adapter identifiers. **Fix:** rename to `environments` (the environments §3 selector convention) and define absence as every adapter.

### F15 — minor — Decisions 4 and 8 — chapter and `member:` line rules disagree

Decision 4: "emits one chapter per closure member with modules" and "lists every member with its weight"; Decision 8: "one `member:` line per closure member with modules". Revision 1 emits a chapter for every chain member even when empty (§5, `monolithic-composed-empty-chapter` vector). **Fix:** pick one rule for chapters and one for `member:` lines; say whether skill and MCP members appear in the header.

### F16 — minor — Decision 1 — "the skill `source` convention" is misattributed

> "an OPTIONAL `path` naming a directory within the snapshot (the skill `source` convention)"

Core §5: `source` "is a portable relative path below the manager's configured source root" — a checkout location, not a subdirectory of a git snapshot. Neither `skillfile-v1` nor `dependencies.skills` has a subdirectory field, although core §3 admits "a directory within one". `path` on `requires.skills` is therefore new wire, which the Compatibility section must count (or restrict `path` to contexts and MCP).

### F17 — minor — Decision 4 — tie-break under `placement` and the root's place in §7 order

"ties broken by the §7 topological order and then by name": §7 order is already name-tie-broken, so "then by name" is vacuous; unstated whether the tie order inverts under `winner-first`; §7 never emits the synthetic root while here the root is a member with modules. **Fix:** "ties keep §7 order regardless of `placement`; the root participates as an ordinary node".

### F18 — minor — Decision 4 — `weights` legality depends on position, but is checked at snapshot validation

"a non-root manifest carrying `weights` is rejected at snapshot validation" makes one immutable snapshot valid as a root and invalid as a dependency; snapshot validation is position-independent everywhere else in core. **Fix:** resolution-time diagnostic (`context_weights_ignored` warning or `context_weights_not_root` error), and say whether a root with `weights` may be reused as a dependency.

### F19 — minor — Decision 3 — lock member pin for `path` overlays and the canonical order are unspecified

The lock records "canonical source identity, resolved version, full commit"; a `path` overlay has none of the three (environments §1: state hash). "Sorted canonical order" names no key. **Fix:** member pin is `commit` or `state_sha256` (the rev-1 pin shape), sort key = (kind, name).

### F20 — minor — Decision 7 — branch tracking is withdrawn silently

Revision 1 admits `--branch` and defaults to tracking the remote default branch (§9.1). The draft's install grammar has `--range | --tag | --revision` only and defaults to `latest`; a repository with no `v*` tags can be installed only by `--revision`. **Fix:** say so in Decision 7 and Rejected alternatives.

### F21 — minor — Decision 6 — channel caveats worth recording

codex: `-p` is also the operator's profile flag; an operand after `--` (`-p work`) collides — reserve a fixed layer name (e.g. `curator-mcp`) and note last-wins is unverified. opencode: `OPENCODE_CONFIG` sits below project `opencode.json`, `.opencode/`, and `OPENCODE_CONFIG_CONTENT` in the documented merge order, so a project entry with the same server name overrides the managed one; state it. claude: `--strict-mcp-config` also disables the managed home's own `.claude.json` servers — intended, but say it.

### F22 — minor — Consequences — "sixteen MUST items apply unchanged" is overstated

M9's detector scope names `Profilefile.json`, `context.json`, `PROFILE.md`; M12's knob list grows (overlay default weight, precedence primitives, per-overlay weight, MCP allowlist). **Fix:** "apply, with M9's detector scope and M12's knob list re-targeted".

### F23 — nit — Compatibility impact — "`Profilefile.json` and `context.json` leave the identifier list"

They never entered core §1.1 or COMPATIBILITY.md. What actually leaves: `schemas/v1/profilefile-v1.schema.json`, `context-manifest-v1.schema.json`, their schema-cases, and the nine `expected/environments/*` vector sets, replaced by the v2 sets. Say that.

### F24 — nit — Decision 1 — `weight` bounds

"an integer" — say non-negative or give a range within CCJ-1 integer bounds; negative weights interact oddly with the overlay default.

## What held under attack (do not reopen)

- Umbrella as convention (any package with `requires.contexts`; any package installs as a root).
- Lock as identity, commit-keyed store entries shared across profiles, nothing re-resolves implicitly.
- Weights never resolve version constraints; empty intersection fails regardless of weights.
- `winner` × `placement`: closed, orthogonal, default reproduces `later-overrides-earlier` (overlay default weight above roots, higher wins, winner last).
- Prerelease rule matches npm exactly (verified against node-semver 7.7.4).
- MCP: declaration-only, no execution, managed homes only, launcher applies the channel; claude/codex/opencode/pi channel facts verified as stated; OQ3 correctly hedges the codex layer composition.
- Overlays join the closure (no second copy), package cannot declare overlays.
- House style matches 0009/0010 (section set, tone, density); English throughout; every § citation resolves except F16's misattribution.
- M4 seeds, M6 `path_prepend`, M10 read-only resolve: no regression found; M16 CCJ-1 alignment is consistent with the lock hash.
