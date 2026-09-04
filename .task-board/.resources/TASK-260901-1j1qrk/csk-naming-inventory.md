# csk surface-naming inventory — TASK-260901-1j1qrk

Classified inventory of every case-insensitive `csk` occurrence in
`relux-works/curator-spec` and `relux-works/curator`, per Decision 0010 D4
(operator-narrowed scope): surface prose/docs/diagnostics stop spelling `csk`
except where they name a frozen §1.1 wire identifier. Wire identifiers,
schema files/ids, vector fixture bytes, conformance expected outputs stay
untouched.

- Inventory command: `grep -rn -i csk` (excluding `.git`; in curator also
  `.task-board` and `.temp` checkout artifacts).
- curator-spec: worktree `~/Developer/ReluxWorks/.worktrees/curator-spec-csk-naming`,
  branch `draft/csk-surface-naming`, base `origin/main` = `f8d7e7a`.
  **712 hits.**
- curator: story worktree `.temp/STORY-260901-2zdg81/worktree`, branch
  `task-board/story/STORY-260901-2zdg81`, base `main` = `979fa36e`.
  **171 hits.**

Frozen §1.1 identifiers (spelling preserved everywhere): `.csk-install.json`,
`.csk-managed.json`, `CSK_PROJECT_ROOT`, legacy alias `csk-skill.json`, schema
filenames `csk-skill-v*.schema.json` (and their `$id` URLs / fixture
directories `schema-cases/csk-skill-v*`).

## Result

Exactly **one surface hit required rewriting**; every other hit names a frozen
wire identifier, a functional/wire artifact, an external-implementation
identifier that this repo cannot rename, or is a frozen historical record.

### Rewritten (surface)

| Repo | Location | Before | After | Rationale |
|------|----------|--------|-------|-----------|
| curator-spec | `protocol/core.md:1644` | ``…portable paths such as `build-cache` or `.csk-build.json`.`` | ``…`build-cache` or `.agent-build.json`.`` | `.csk-build.json` is a purely illustrative hypothetical path (sole occurrence across both repos, confirmed by `grep -rn csk-build`), NOT a frozen §1.1 identifier. Rewritten into the agent-* family per D4. Signed commit `4d55698`. |

## curator-spec — wire/untouched dispositions (711 hits)

| Category | Files | Hits | Disposition |
|----------|-------|-----:|-------------|
| Conformance fixture index + manifest | `conformance/v1/schema-cases/index.json`, `conformance/v1/manifest.json` | 589 | WIRE: generated/pinned fixture paths and schema names (`csk-skill-v*` families, `fixtures/go-build-skill/.csk-install.json`). Digest-pinned; untouched. |
| Conformance vectors | `conformance/v1/vectors/skill-manifest-resolution.json` (6), `conformance/v1/vectors/build-drivers.json` (5) | 11 | WIRE: vector bytes exercising the legacy `csk-skill.json` read alias and `.csk-install.json` marker (incl. `//go:embed .csk-install.json` directive bytes). Untouched. |
| Schema files | `schemas/v1/csk-skill-v{1..8}.schema.json` (`$id` + `title`), `schemas/v1/install-marker-v{1..4}.schema.json` (titles `.csk-install.json schema N`), `schemas/v1/adapter-ledger-v1.schema.json` (title `.csk-managed.json schema 1`) | 22 | WIRE: frozen schema filenames, `$id` URLs, and titles naming frozen marker/ledger filenames. Untouched. |
| Vector generator + tests | `tools/generate-vectors/main.go` (13), `tools/generate-vectors/main_test.go` (21) | 34 | FUNCTIONAL/WIRE: generates and asserts the frozen fixture bytes above (legacy filename, schema names, pinned digests, agent-/csk-skill parity assertions). Renaming breaks generation determinism. Untouched. |
| Validation tooling | `tools/validate.py` (6), `tools/release_gate.py` (6), `tools/test_validate.py` (4) | 16 | FUNCTIONAL/WIRE: match/require frozen schema filenames `csk-skill-v*.schema.json`. Untouched. |
| Protocol text naming frozen identifiers | `protocol/core.md:36,38,158,160,165,1442,1445,1695,1799` | 9 | KEEP: sentences that NAME frozen §1.1 identifiers (`.csk-install.json`, `.csk-managed.json`, legacy `csk-skill.json`, `csk-skill-v8.schema.json`); identifier spelling preserved, no csk-flavored prose beyond the identifiers themselves. (10th core.md hit was line 1644, rewritten above.) |
| Profiles | `profiles/manager.md:88,588,718` | 3 | KEEP: name frozen `.csk-install.json` / `CSK_PROJECT_ROOT`. |
| Repo docs naming frozen identifiers | `COMPATIBILITY.md:40-44,47,59` (6), `README.md:24,25` (2), `SECURITY.md:21` (1), `schemas/v1/README.md:7,36` (2), `CHANGELOG.md:268,275` (2) | 13 | KEEP: the frozen-identifier list itself and sentences naming frozen identifiers/schema filenames; CHANGELOG additionally is a historical release record. |
| External implementation name | `README.md:64` (`[csk](…/cocoaskills)` open-protocol section), `decisions/0001-normative-boundaries.md:23` | 2 | KEEP: `csk` here is the proper name of the independent Python implementation (cocoaskills). This repo cannot rename an external product; the README open-protocol section is the sanctioned single reference. |
| Decision records | `decisions/0004:59`, `decisions/0005:12,28`, `decisions/0010:20,324,325,326,330` | 8 | KEEP: historical decision records. 0004/0010 hits name frozen identifiers; 0010:324-330 is the D4 policy text that deliberately spells `csk` to define this very rule; 0005 quotes real external paths/commands (`src/csk/builds/go_v1.py`, `csk skill check`). |
| CI plumbing | `.github/ci/implementation-coverage.tsv:42,50,51` (3), `.github/workflows/implementations.yml:117` (`CSK_REQUIRE_FULL_CANDIDATE_ROOT`) | 4 | FUNCTIONAL/WIRE: fixture-directory paths (`schema-cases/csk-skill-v8`) and an env var consumed by the manager implementation's candidate CI lane. Renaming changes CI behavior. Untouched. |
| Checkout artifact | `.git` gitdir pointer line | 1 | Not repo content (worktree gitdir contains the branch name `curator-spec-csk-naming`). |

## curator — wire/untouched dispositions (171 hits, zero rewrites)

| Category | Files | Hits | Disposition |
|----------|-------|-----:|-------------|
| Wire-identifier constants + code | `internal/hashing/hashing.go:22` (`MarkerName = ".csk-install.json"`), `internal/marker/marker.go:1,28`, `internal/adapters/adapters.go:35,127` (`LedgerName = ".csk-managed.json"`), `internal/skillspec/types.go:12` (`LegacyManifestName = "csk-skill.json"`), `internal/skillspec/parse.go:32`, `internal/audit/audit.go:311` | 9 | WIRE: constants holding frozen §1.1 filenames plus the comments naming them. `audit.go:311` scopes the audit to the literal legacy manifest filename — behavior, not prose. Untouched. |
| Env-file emission | `internal/envfiles/envfiles.go` (10 hits: emits `CSK_PROJECT_ROOT`, `CSK_GLOBAL_ROOT` into generated `env.sh`/`env.ps1`, + comment) | 10 | WIRE/FUNCTIONAL: `CSK_PROJECT_ROOT` is frozen; `CSK_GLOBAL_ROOT` is the functional global-env counterpart written into user environments — renaming breaks installed environments. Untouched. |
| Go tests exercising wire behavior | `internal/{closure,install,install/atomicity,skillspec,skillcheck,audit,adapters,buildsource,staging,envfiles,shell}` test files (55 hits) | 55 | FUNCTIONAL/WIRE: write `csk-skill.json` / `.csk-install.json` / `.csk-managed.json` fixtures and assert `CSK_PROJECT_ROOT`/`CSK_GLOBAL_ROOT` behavior — they exercise the frozen legacy read alias and marker/ledger/env contracts. Test comments name those identifiers. Untouched (brief: code identifiers/tests out of scope; none leak csk wording into user-facing output beyond frozen names). |
| Docs naming frozen identifiers | `CONTRIBUTING.md:26`, `CHANGELOG.md:69`, `README.md:60`, `docs/ci-gates.md:49`, `docs/implementation-plan.md:36,44,61,90` | 8 | KEEP: sentences naming frozen wire filenames / schema-case families; CONTRIBUTING/implementation-plan explicitly document the wire-format used-as-is rule this task implements. CHANGELOG is a historical record. |
| CI plumbing | `.github/ci/platform-cases.tsv:229`, `.github/ci/root-artifacts.tsv:37` | 2 | FUNCTIONAL/WIRE: declare `schema-cases/csk-skill-v8` fixture family as a required root artifact. Untouched. |
| LOGBOOK.md | 63 hits across findings/records | 63 | KEEP: append-only historical record. Hits quote frozen identifiers, real external cocoaskills identifiers (`src/csk/*` paths, `csk install`/`csk gc` CLI, `CSK_CONFIG`, `CSK_PROBE_ROOT`, `.csk-materialization-plan-*`, `__csk-go-worker-v1`), and past code states. Rewriting would falsify the record. |
| .research/* | `260811_inventory-language-and-reference-surfaces.md` (24), `260823_schema-8-impact-analysis.md` (3), `260823_vendor-inert-text-audit-policy.md` (2), `260822_decision-0008-open-questions.md` (2) | 31 (approx; exact per-line list in grep dump) | KEEP: dated research evidence records quoting live estate paths, real `csk` CLI transcripts, external repo paths (`src/csk`), and frozen filenames. Historical evidence; rewriting would break its evidentiary value. |

Note on external-name policy: both repos already enforce/observe the rule that
the alternative implementation's product name appears once in README; `csk` in
LOGBOOK/.research/decisions is that implementation's real CLI/package name and
real filesystem paths, i.e. identifiers of an external system, not this
repo's surface vocabulary.

## Gates (real exit codes)

curator-spec (`~/Developer/ReluxWorks/.worktrees/curator-spec-csk-naming`, after the rewrite):
- `tools/validate.py` (the exact command `make validate` runs), executed with a
  jsonschema-equipped venv python because system `python3` lacks `jsonschema`
  (bare `make validate` exits 2 on `ModuleNotFoundError` before touching repo
  content): **exit 0**, "validated 57 schemas and 773 vector files".

curator (story worktree at `979fa36e` + this task's zero-file delta; submodule
`agents/skills/skill-go-testing-tools` initialised first per repo requirement):
- `go build ./...`: **exit 0**
- `go vet ./...`: **exit 0**
- `go test $(go list ./... | grep -v 'cmd/curator$') -count=1`: **exit 0** (57 ok)
- `go test ./cmd/curator -count=1 -timeout 30m` split into three bounded
  sequential slices to respect the ~10-minute shell bound (masks together cover
  all 96 tests, first letters A-U): `-run '^Test[A-C]'` **exit 0** (187s),
  `-run '^Test[D-R]'` **exit 0** (43s), `-run '^Test[S-Z]'` **exit 0** (27s).

## Commits

- curator-spec `draft/csk-surface-naming`: `4d55698` (signed, ECDSA
  `SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`, good signature; not
  pushed per brief).
- curator: **no commit — zero surface hits required rewriting**; the working
  tree is unchanged (`git status` clean). All 171 curator hits are classified
  wire/functional/historical above.

Raw grep dumps (712 + 171 lines) preserved in the session scratchpad
(`spec-hits.txt`, `curator-hits.txt`); the tables above account for every line.
