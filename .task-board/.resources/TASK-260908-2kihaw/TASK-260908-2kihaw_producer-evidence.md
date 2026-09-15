# B2 producer evidence — TASK-260908-2kihaw: skill manifests in owning repos

Producer (Muse Spark xhigh, developer). Status: work handed to review;
independent Astra-medium review, landing, and v-tagging are routed by the
orchestrator after this handoff.

## 1. Delivered repositories (all PRIVATE, org relux-works)

| Skill | Repo | main (import) | b2/manifest head (for review) | PR |
|---|---|---|---|---|
| pdf | https://github.com/relux-works/skill-pdf | 9e3e72a1509823998ab08586b6c981689f2245f2 | 6d2392a861cdf389c9aa0245c46e8207975a2bcd | https://github.com/relux-works/skill-pdf/pull/1 |
| skill-creator (standalone successor reused) | https://github.com/relux-works/skill-creator | 0a49a505096831a07d6971d1a5120ea5ba8a432e | ea8fd665486233ccaccce2d8508ea8caa9378bcd | https://github.com/relux-works/skill-creator/pull/1 |
| agents-attachments (CLI skill) | https://github.com/relux-works/skill-agents-attachments | a8156f1d7df66ad4fc7e50e881a9a5d5bfacd911 | 240f0292a6484a743ede98fb9af0097f4a875480 | https://github.com/relux-works/skill-agents-attachments/pull/1 |

All six commits carry `[skip ci]`, are signed `Good "git" signature for
ivan@relux.works` (ED25519 SHA256:Ng99XGF2pboYgFVfWJhYI2JRi0PyYsV9UwsJ70NBYd0),
and use identity Ivan Oparin <ivan@relux.works>. Each PR asks for a
comment-review verdict on its exact head; a producer comment on each PR
records the amended head (shellcheck-clean gate idiom, no behavior change).
Nothing is landed; no tags or releases were created (tags target the exact
verified reviewed commit, post-landing).

## 2. Successor decision (brief: reuse standalone repo if it exists and the plan permits)

Reused. Agents-infra STORY-260903-2pxzvb records the owner decision
2026-09-03 that the standalone build is canonical, and its SKILL.md is
byte-identical to `dee5403:.skills/skill-creator/SKILL.md`
(sha256 93b5880f24b5ffd05995a13413d37dedf83330b2f0d130b0ba68fa4031722811
both sides). The unpublished working tree at
`/Users/iv/agents/skills/skill-creator` was published as-is (plus manifest
wiring) to `relux-works/skill-creator`; the author's local dir was not
modified.

## 3. Source fingerprints (read-only extraction, `git show dee5403:<path>`)

Byte-faithful files (staged sha256 == source sha256):

- pdf: SKILL.md 02be9010…, assets/template.html5 9ae6d627…,
  assets/themes/prose-classic.css e7a8f7f3…,
  assets/themes/report-clean.css 193149e0…,
  scripts/render-pdf.sh 29a7b5ae… (mode 100755 preserved)
- skill-creator: full standalone tree incl. SKILL.md 93b5880f…,
  references/, scripts/ (100755 preserved), agents/openai.yaml,
  locales/metadata.json, .skill_triggers/, setup.sh, tests/*.py, LICENSE
- agents-attachments: internal/attachments/attachments.go 85abb9fa…,
  internal/attachments/attachments_test.go f1cf5946…;
  SKILL.md body == INSTRUCTIONS_ATTACHMENTS.md 73ddf388… (only YAML
  frontmatter added); LICENSE == agents-infra MIT LICENSE text

New files (manifests are new per brief): per repo `agent-skill.json`
(schema 8), `tests/manifest_gate.sh`, repo `README.md` (+Makefile `check`
target; skill-creator Makefile/README extended, hook `make test` runs the
gate first), attachments `cmd/agents-attachments/main.go` + `go.mod` +
`.gitignore`s + gate fixtures.

Manifest surfaces: pdf = 1 script command (render-pdf; exec
pandoc/weasyprint); skill-creator = 3 script commands
(init/validate/package-skill; exec git/zip); agents-attachments = 1 go-v1
build command (nested module at cmd/agents-attachments, no `modules`
replacements; exec sips/magick/convert for HEIC only; env_read the three
documented vars). Design facts found by probing (not guessing):
build `source_dir` must contain the nearest go.mod directly, each `modules`
entry must contain its own go.mod, dangling paths fail `skill.manifest_invalid`.

## 4. Umbrella range contract (for B1 `relux-root-context-ivan`)

Initial versions: **v0.1.0** for all three skills (tags post-landing).
Documented `requires.skills` entries (also in each repo README):

- "pdf": {"git": "git@github.com:relux-works/skill-pdf.git", "range": "^0.1"}
- "skill-creator": {"git": "git@github.com:relux-works/skill-creator.git", "range": "^0.1"}
- "agents-attachments": {"git": "git@github.com:relux-works/skill-agents-attachments.git", "range": "^0.1"}

`umbrella_range_contract.py` (attached): range/identifier grammars loaded
from the live curator-spec schemas; caret admission per Decision 0012 §2
(`^0.1` ≡ >=0.1.0 <0.2.0-0, prereleases excluded); §6.1 canonicalization
checked. Result: 24/24 green (exit 0). Bound: live resolution against real
tags happens at B1 umbrella install time.

## 5. Validation (local only, no hosted CI; exits recorded in the log)

| Command | Exit |
|---|---|
| make test in skill-pdf (curator check + gate) | 0 |
| make test in skill-agents-attachments (curator check + gate + go vet + go test, 13 tests) | 0 |
| make test in skill-creator (curator check + gate + 5 unit tests) | 0 |
| python3 umbrella_range_contract.py (24 checks) | 0 |
| shellcheck -S warning on render-pdf.sh + 3 gate scripts | 0, no findings |
| python3 -m py_compile on skill-creator python files | 0 |

Full transcript: attached `TASK-260908-2kihaw_validation.log` (27 GATE PASS
lines, 0 failures). One advisory warning is present in every
`curator skill check` run (`skill.command_resolution_contract_missing`);
it is also emitted for the exemplar skill-project-management and is left
as-is because skill bodies stay byte-faithful. Source zsh scripts are
excluded from shellcheck (zsh dialect, byte-faithful) and are covered
behaviorally instead.

## 6. Mutant table (production entry point: `curator skill check` + the repo behavioral suite)

| Mutant | What it narrows the gate to | Named failing probe | Survivors / bound stated |
|---|---|---|---|
| N1 per repo: drop required `capabilities` | required-field presence, not file existence | tests/manifest_gate.sh "mutant N1 killed" | none — killed in all 3 repos |
| N2 pdf/creator: `unix_path` → missing sibling; attachments: `source_dir` → cmd/missing | path/root containment, not string shape | tests/manifest_gate.sh "mutant N2 killed" | none — killed in all 3 repos |
| B1 per repo: emptied command body, every manifest token intact | static-vs-behavioral split | behavioral suite only (pdf render produces no PDF; creator validator admits broken fixture; attachments build fails) | SURVIVES the static checker in all 3 repos (recorded blind spot: `curator skill check` exits 0 on empty/non-executable scripts, verified by probe). Bound: static check proves shape, never behavior; the gate verdict for B1 comes from executing the behavioral suite. |

Delete-only mutants were not used as evidence; every gate above is a
narrowing mutant.

## 7. AC coverage: 7 of 8 rows driven

1. pdf manifest in owning repo — `curator skill check` / `make test` in skill-pdf ✓
2. skill-creator successor manifest in owning repo — same in skill-creator ✓
3. agents-attachments manifest in owning repo — same in skill-agents-attachments ✓
4. umbrella `requires.skills` ranges — `umbrella_range_contract.py`, 24/24 ✓
5. bodies byte-faithful — cmp + sha256 table in §3 ✓
6. strict local validation (parse + schema + ranges) — curator check + schema-loaded grammar checks ✓
7. signed scoped PR delivery (`[skip ci]`, private repos) — §1 table ✓
8. independent reviewer acceptance + landing + v0.1.0 tags — NOT driven (pending orchestrator routing to Astra-medium review). Tags will target the exact verified reviewed commit; never retag.

## 8. Gaps, bounds, follow-ups

- Review verdicts outstanding on all three PR 1s; landing (fast-forward of
  the exact reviewed head) and signed `v0.1.0` tags + GitHub Releases follow
  the verdicts. Repos stay private until the operator flips visibility.
- `NEVER` touched: ~/.agents, ~/.claude, ~/.codex, real ax, agents-infra
  launchers, agents-infra source (read-only `git show`), the Story worktree
  (clean: still at cb232a1 on task-board/story/STORY-260908-sd6xkr, `git
  status` empty).
- Staging checkouts live at /tmp/b2/* (outside the worktree and board);
  the durable record is the three GitHub repos + this evidence.
- Preflight recorded: origin HEAD symref refs/heads/main =
  cb232a120c9a04c56ae5037c82347921f020e688; origin/main is an ancestor of
  the worktree HEAD (identical OIDs).
