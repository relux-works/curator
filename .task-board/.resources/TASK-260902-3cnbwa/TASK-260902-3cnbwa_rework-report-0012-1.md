# Rework report 0012-1: Decision 0012 after review findings 0012-1

Producer run RUN-260902-223ac2 (claude-fable-5-1), 2026-09-03. Subject:
`decisions/0012-context-packages-and-semver-locks.md` on
`draft/decision-0012-context-packages` in worktree
`~/Developer/ReluxWorks/.worktrees/curator-spec-decision-0012`, base
`a25dc67`. The earlier producer run RUN-260902-f93b9f exited on a provider
session limit before its first tool call and made no edits.

## Commit

`84447064233a44467f356ad2a349a54ce2433c56` — one signed commit on top of
`a25dc67`, author Ivan Oparin <oparin@me.com>, SSH signature verified
("Good git signature", key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM;
git reports validity `U` because the configured allowed-signers file is a
stale temp path — the base commit `a25dc67` shows the same state). Only
the decision file changed (+529/−167, 433 → 796 lines). Not pushed, not
tagged, not marked done.

Note on location: the story worktree
`.temp/STORY-260902-le61cp/worktree` (branch
`task-board/story/STORY-260902-le61cp`) does not contain the decision file
at all — it sits at main `4d55698`. The producer brief names the decision
worktree explicitly, so the edit and commit live there; the story branch
is untouched.

## Finding → disposition

| # | Sev | Disposition | Where in the reworked document |
|---|---|---|---|
| F1 | blocking | applied as decided | Compatibility impact: per-section table over environments §1–§13 (subsection granularity where dispositions differ), manager §12.1–§12.7, `cli/curator.md`, the two withdrawn schemas and schema-cases, the two in-place schema rewrites, the nine vector sets. Verified unchanged: §5.2, §5.7, §7 body, §7.1, §7.2, §7.4–§7.7, §8.3–§8.5, §9.2, §9.3, §9.5, §10.4, §11, §11.1, manager §12.4 |
| F2 | major | applied as decided | Decision 2 "Resolution": seed → select-and-expand (lexicographic pending order) → downward re-selection with dropped contributions → final check; monotone non-increasing selections give termination; no cross-name backtracking; `\|\|` satisfied by highest satisfying candidate; also a new Rejected alternative "A backtracking resolver" |
| F3 | major | applied as decided | Decision 2 "Versions": `+build` tags are not candidates; uniqueness argument replaces "ties are impossible" |
| F4 | major | applied as decided | Decision 1 bullet on `version`; Decision 2: exact forms fix the candidate; contexts/MCP use manifest `version`, skills use the highest version tag peeling to the commit or none; `path` authoritative; `context_version_mismatch` named |
| F5 | major | applied as decided | Decision 4 rules 1–4: declared edge weights from distinct direct requirers must agree (`context_weight_conflict`) unless the root map names the member; root edge weights are map entries; both → `context_weights_duplicate`; overlay weight outranks (machine config over repository content) |
| F6 | major | applied as decided | Decision 4: "Weights order chapters and nothing else in this revision"; collision sentence deleted |
| F7 | major | applied as decided | Status ("changes nothing in the frozen core"), Context (semver honored wherever a lock exists), Decision 2 "Skills in the closure", Compatibility impact opening, Rejected alternative "Ranges in project `Skillfile.json` now", Open question 6 (project lock) |
| F8 | major | applied as decided | Decision 2 "Ranges" and "Prereleases": node-semver cited by README section; comparator sets, partial coercion (`>=2.1`, `>1.2`, `<3`, `<=1.2`), caret on 0.x/0.0.x/partials, tilde partials, x-ranges, `-0` upper bounds, hyphen ranges and in-range `v` excluded, `latest` as Curator spelling. Every stated coercion re-verified against node-semver 7.7.4 `validRange`/`satisfies` on this machine |
| F9 | major | applied as decided | Decision 6: allowlist over MCP package canonical source identities (`mcp_package_not_allowed`), bare `command` on `PATH`, `https` `url` without userinfo/query/fragment, `args`/`url` inside `context-secret-material`, `env_names` under §2 grammar, never manager §3.1 reserved, optionally bounded by a lockable passable-names list; Security impact; Rejected alternative "An MCP allowlist over launcher binaries" |
| F10 | major | applied as decided | Decision 6 (fragment `mcp.env_names` union; `curator-run` as single composer adds them to the plan allowlist) and Decision 8; worked example shows both |
| F11 | major | applied as decided | environments.md stays at revision 1; type line only bumps; Status, Compatibility, Open questions 2 and 4 reworded; no "revision 2" remains |
| F12 | major | applied as decided | Context cites the review's M1 resolution, Option A, "to be recorded as its own decision under the next free number after reconciliation with the swift-driver draft's 0011"; no "Decision 0011" remains |
| F13 | major | applied as decided | New Decision 9 "Worked example": lock excerpt (sort key (kind, name), pins, weights, `required_by`, overlay), `curator-root-context-v2` header in emitted order under the default policy, fragment with `profile.lock_sha256` and `mcp`, CCJ-1 bytes of `.agent-context/mcp/claude_code.json`, the resulting exec |
| F14 | minor | applied as decided | `required_in` → `environments` (environments §3 selector; absent = every adapter) |
| F15 | minor | applied as decided | Decision 4 and 8: one chapter per context member with applicable modules; one `member:` line per context member; skills/MCP in the lock only; `monolithic-composed-empty-chapter` re-cut as the no-chapter case (also in the table) |
| F16 | minor | applied as decided | Decision 1: `path` on context and MCP requirements only, counted as new wire; skills keep §4.4 addressing |
| F17 | minor | applied as decided | Decision 4: ties keep §7 order regardless of `placement`; root as ordinary node; "then by name" dropped |
| F18 | minor | applied as decided | Decision 4: `context_weights_not_root` at resolution time; reuse as dependency allowed exactly when the map is empty or absent |
| F19 | minor | applied as decided | Decision 3: pin is `commit` or `state_sha256`; sort key (`kind`, `name`) bytewise; `path` member's source path stays out of the lock |
| F20 | minor | applied as decided | Decision 7 and Rejected alternative "Branch-tracking profile roots" |
| F21 | minor | applied as decided | Decision 6 channel list: fixed codex layer `curator-mcp`, `-p` collision and last-wins unverified (Open question 3); opencode merge order recorded; claude `--strict-mcp-config` disabling `.claude.json` servers recorded as intended |
| F22 | minor | applied as decided | Consequences: MUST items apply with M9 scope and M12 knobs re-targeted |
| F23 | nit | applied as decided | Compatibility impact closing paragraph and table rows name the two schemas, their schema-cases, and the nine vector sets |
| F24 | nit | applied as decided | Decision 1: non-negative integer at most 2147483647 |

No deviation from the author decisions. Two additions beyond the letter of
the brief, both forced by applying it: the MCP channel descriptor grammar
(`argument: path | contents | name`, `with`) had to be stated for the
example's fragment to be well-formed, and the lock member list gained
`path` and `required_by` so that F5's "distinct direct requirers" is
checkable from the lock.

## Evidence gathered

| Claim | Source | Result |
|---|---|---|
| node-semver coercions quoted in Decision 2 | `semver` 7.7.4 under the global npm, `validRange`/`satisfies`/`maxSatisfying` | `^0.2.3`→`>=0.2.3 <0.3.0-0`; `^0.0.3`→`>=0.0.3 <0.0.4-0`; `^0.1`→`>=0.1.0 <0.2.0-0`; `^0`→`<1.0.0-0`; `~1.2`→`>=1.2.0 <1.3.0-0`; `~1`→`>=1.0.0 <2.0.0-0`; `>=2.1 <3`→`>=2.1.0 <3.0.0-0`; `>1.2`→`>=1.3.0`; `<=1.2`→`<1.3.0-0`; `1.2`→`>=1.2.0 <1.3.0-0`; `x`→`*`; `latest`→null; hyphen range accepted by npm (excluded here); `2.0.0-rc.1` satisfies `^2.0.0-rc.0`, `2.1.0-rc.1` does not; `3.0.0-rc.1` does not satisfy `<3`; `eq('1.2.3+a','1.2.3+b')` true (hence F3) |
| `claude --mcp-config`, `--strict-mcp-config` | `claude --help`, 2.1.259 installed | both present with the quoted help text |
| codex `-p, --profile` layering | `codex --help`, 0.151.0 | "Layer $CODEX_HOME/<name>.config.toml on top of the base user config" |
| `pi` has no MCP channel | `pi --version` 0.84.2; reviewer's help check | accepted from review evidence |
| opencode merge order | reviewer's docs check (opencode not installed) | accepted from review evidence, recorded verbatim |
| skill manifest carries no `version` | `schemas/v1/agent-skill-v8.schema.json` properties | confirmed — drives the skill branch of the F4 rule |
| manager §3.1 reserved set, §1 `locked`, §6; core §2, §4.4, §6.1–§6.3, §7, §10; environments §1–§13; manager §12.1–§12.7; `cli/curator.md` rows; vector and schema-case listings | read on this checkout at `4d55698` | every citation in the reworked text resolves |

## Checks run

- `grep` for `revision 2`, `Decision 0011`, `required_in`, `schema 9`,
  `Skillfile schema 2`, `Ties are impossible`, `mechanical collision`,
  `then by name`: no matches (exit 1 from grep, expected).
- All four fenced JSON blocks parse; the materialized MCP bytes equal
  `json.dumps(sort_keys=True, separators=(',',':'))` of the object.
- Prose wrap: no line over 79 columns outside code and tables except the
  unchanged title line.
- No test suite applies to a decisions-only change; nothing was built.
