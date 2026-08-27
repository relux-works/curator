# TASK-260810-1veyfw — skill-facing CLI and protocol implementation inventory

Date: 2026-08-11

Role: researcher

Outcome: research decision prepared for review

Delivery input: `TASK-260810-1veyfw_skill-facing-cli-source-closure.md`

## Decision

The evidence supports the delivery classification in the task attachment, with
nine important findings:

1. **Go is the only implemented Curator build-adapter baseline.** Curator
   implements `go-v1` and `go-repository-v1`; the canonical live skill manifest
   found in the installed roots declares only `go-v1` commands.
2. **Node/TypeScript is not implemented in, derived from, or co-located with
   the Python project.** the alternative implementation is a separate Python implementation of the
   Curator Protocol. Its cited tree has no Node package manifest, lockfile, or
   JavaScript/TypeScript source. Node/TypeScript is therefore a new Curator
   adapter target. It should share protocol policy and conformance semantics
   with the Python reference where applicable, not share Python implementation
   code or packaging state.
3. **Swift has real current CLI surfaces, C has real transitive SwiftPM edges,
   and currency exchange is the estate's clearest mixed Go/Swift migration
   input.** `exchange` is a Go frontend that launches the independently built
   Swift `exchange-scraper` subprocess. The separate iOS testing graph contains
   C shim/parser targets. No current C++, Objective-C, or Objective-C++ skill
   CLI edge was found, so those remain confirmed delivery targets only where
   the eventual SwiftPM profile can prove their source boundary conservatively.
4. **Rust is a confirmed target without a current estate implementation.** No
   `Cargo.toml` or `Cargo.lock` occurred in the revision-pinned repositories or
   active skill roots. One Rust benchmark helper exists inside a generated
   swift-collections dependency checkout, but it is not a Cargo package, CLI
   product, or skill command. A purpose-built fixture, rather than an existing
   CLI migration, must seed that adapter.
5. **Python is both the independent protocol reference and a current CLI
   ecosystem baseline, not a default new adapter deliverable.** In addition to
   the alternative implementation and the legacy script fleet, `telegram-telethon` is a live
   Python/Bash CLI surface. Its exact Telethon top-level pin still permits
   transitive drift, and observed runtimes contain bytecode and Windows
   executables. That behavior cannot be admitted unchanged under the new
   source-only closure invariant.
6. **Kotlin remains outside the active investigation.** A Kotlin/JVM Gradle CLI
   exists in the Android testing repository, but it is recorded only as a
   deferred boundary. No Kotlin closure strategy is launched or recommended by
   this outcome. Dart and .NET are likewise deferred, with no current surface
   found.
7. **The current estate cannot be defined only by machine command
   declarations.** The reproducible inclusion rule also admits repository-
   shipped commands named by skill instructions or launchers and every
   the alternative implementation global/registered-project surface. This is why currency exchange
   and Telegram are in scope even though both install markers have empty
   `commands` arrays.
8. **External system commands are a distinct trust class.** `glab` and
   `sentry-cli` are declared as `type: system`; they are current skill-facing
   command dependencies but are neither source payloads nor adapter build
   products. Their provenance/version policy must remain outside the language
   source-closure graph.
9. **A live source checkout and an installed the alternative implementation surface are separate
   revision authorities.** The clean `skill-grafana` feature checkout at
   `6234e2a…` declares `analyze`, `query`, and `grafana-auth`, while the
   configured and installed `v1.1.0` state at `a557671…` declares only
   `grafana`. Both are current estate states and both are inventoried; neither
   command set may be substituted for the other merely because the repository
   name matches.

The inventory is sufficient to begin adapter design after the repository and
manifest gaps below are turned into explicit implementation tasks. It is not
evidence that the non-Go ecosystems already satisfy the source-closure policy.

## Evidence method and limits

- Discovery and revision selection are separate. The deterministic authority
  rule is applied in this order:
  1. a physical `csk-skill.json` under the the alternative implementation skills root is a
     **discovery locator**, not itself revision evidence; resolve its Git
     top-level and inspect the manifest and launch sources from the checkout's
     exact `HEAD` when the tree is clean;
  2. resolve every registered-project or global configured ref to an immutable
     commit and inspect that commit as the **configured-target state**, even
     when status says `update-available`;
  3. read each `.csk-install.json` commit, content hash, file list, and commands
     as the **installed state**; a marker never supplies commands for a
     different working-copy or configured revision;
  4. union these states by repository, commit, and command, deduplicating only
     identical revisions. If command-bearing files are dirty and cannot be
     bound to an immutable state, record the dirty state and fail the closure
     claim rather than silently using `HEAD`; all eight current manifest
     checkouts were clean during this scan;
  5. include every command declared by a Curator `agent-skill.json` under its
     cited immutable repository revision;
  6. include every repository-shipped CLI that a shipped `SKILL.md`, setup
     entrypoint, or launcher tells the agent to invoke, even when the install
     marker's `commands` array is empty;
  7. enumerate every the alternative implementation global and registered-project surface, then
     record a negative disposition for a surface that ships no CLI rather than
     silently dropping it;
  8. classify `type: system` commands separately because the skill does not
     own their source or build; and
  9. retain libraries, toolchains, services, and companion packages as explicit
     dependency edges, not standalone skill CLIs, unless another included rule
     independently admits them.
- The second-review estate scan was rerun on 2026-08-11 with `csk 0.9.0` using
  `csk list --paths`, `csk global status`, and `csk status --all`; every command
  exited 0. It covered four registered projects and four global surfaces.
  A hidden-file `rg --files` projection over `/Users/iv/.agents/skills`,
  `/Users/iv/agents/skills`, `/Users/iv/.the alternative implementation/global/skills`, and each
  registered project's `.agents/skills` found 16 installed markers, eight
  `csk-skill.json` manifests, and the one current Curator
  `agent-skill.json`. The eight physical the alternative implementation manifests resolve to 15
  distinct live/configured/installed Git revisions after exact-commit
  deduplication. The task-scoped revision table and closure ledger below account
  for every resulting command-bearing, repo-facing, docs-only, missing,
  library-only, or external-system disposition; the physical-file count is not
  presented as command-closure proof.
- Public repositories were inspected at immutable Git object IDs. Their
  `main` heads were resolved directly with `git ls-remote`; the ledger below is
  the resulting revision set.
- Repository trees were queried recursively for manifests, package-manager
  files, locks, known executable/intermediate extensions, Rust/Node sources,
  and named extensionless build outputs. Extension-only scanning was not
  treated as sufficient: the largest repository defect found is an
  extensionless Mach-O file.
- Installed state was inspected through the alternative implementation markers, commit-keyed runtime
  directories, command shims, and the manager's read-only status commands on
  2026-08-11. Installed observations are evidence about the real machine estate,
  not claims about every historical or remote version.
- “No surface found” is bounded to the cited revisions and the exact active
  roots above. It is not a universal claim about all possible skills.
- External commands named by a `system` declaration (for example `glab` or
  `sentry-cli`) are recorded separately. They are not package payloads built by
  a language adapter and need their own system-dependency trust policy.

## Authoritative revision ledger

| Authority | Revision inspected | Why authoritative here |
| --- | --- | --- |
| [Curator](https://github.com/relux-works/curator/tree/9ba552f04bacb91c4b643378cac928ed90bfb229) | `9ba552f04bacb91c4b643378cac928ed90bfb229` | Remote `main`; Go manager and implemented driver baseline. |
| [Curator Protocol](https://github.com/relux-works/curator-spec/tree/dce6643c55434464c56f0fe20064db754cd58c61) | `dce6643c55434464c56f0fe20064db754cd58c61` | Remote `main`; schemas, manager profile, conformance fixtures, and decision 0005. |
| [the alternative implementation](the alternative-implementation repository (README link)) | `7f04ae1141c9f1f39f9320e8bb0ca5ad231abf5f` (`v0.13.0`) | Remote `main`; independent Python protocol implementation. |
| [skill-grafana live feature checkout](https://gitlab.wildberries.ru/portals/agentic-infra/skills/skill-grafana/-/tree/6234e2a3ef35415b765f74cca21c8ad7846f521e) | `6234e2a3ef35415b765f74cca21c8ad7846f521e` (`feature/oparin/PMA-23845`) | Clean working-copy `HEAD` under the configured the alternative implementation skills root; declares three Python CLI surfaces distinct from installed `v1.1.0`. |
| [skill-project-management](https://github.com/relux-works/skill-project-management/tree/8dc0b71490214fe5ead6bf9cfde9574df084fd91) | `8dc0b71490214fe5ead6bf9cfde9574df084fd91` | Remote `main`; only canonical installed `agent-skill.json` with build commands. |
| [skill-ios-testing-tools](https://github.com/relux-works/skill-ios-testing-tools/tree/bd59caaf4bb712f35d4b8b73141ce28999cc13cb) | `bd59caaf4bb712f35d4b8b73141ce28999cc13cb` | Remote `main`; current SwiftPM CLI package. |
| [skill-ios-app-manager](https://github.com/relux-works/skill-ios-app-manager/tree/a8523bbffc5399af68970ba772b9c67ba7cdd5d3) | `a8523bbffc5399af68970ba772b9c67ba7cdd5d3` | Remote `main`; skill-facing Go CLI and setup path. |
| [skill-product-appraisal](https://github.com/relux-works/skill-product-appraisal/tree/01daa3436f81a6eedefb84a24d9ffbb7e337cc46) | `01daa3436f81a6eedefb84a24d9ffbb7e337cc46` | Remote `main`; skill-facing `appraise` Go CLI. |
| [skill-agent-facing-api](https://github.com/relux-works/skill-agent-facing-api/tree/656ad0a73dbffc92e732ed95c39a7ae3197f28f1) | `656ad0a73dbffc92e732ed95c39a7ae3197f28f1` | Remote `main`; source dependency used by `appraise`. |
| [relux-agents-infra](https://github.com/relux-works/alexis-agents-infra/tree/ccf0daf444aa1f6ce13119151a860855d7360837) | `ccf0daf444aa1f6ce13119151a860855d7360837` | Remote `main`; skill-facing `agents-infra` Go CLI. |
| [skill-android-testing-tools](https://github.com/relux-works/skill-android-testing-tools/tree/b79449cc3e1767680b069ae314c850a1d93c6f99) | `b79449cc3e1767680b069ae314c850a1d93c6f99` | Remote `main`; current Go/Swift surfaces and deferred Kotlin boundary. |
| [skill-currency-exchange](https://github.com/relux-works/skill-currency-exchange/tree/c29210aa6eb4cc0f64f307fa30561ac80feb6b3b) | `c29210aa6eb4cc0f64f307fa30561ac80feb6b3b` (`v2.1.1`) | Global the alternative implementation marker and remote tag agree; mixed Go `exchange` plus SwiftPM `ExchangeScraper`. |
| [telegram-telethon](https://github.com/ivanopcode/telegram-telethon/tree/b9a76b01e7ce211c1d0e707f97b231ee7b817d41) | `b9a76b01e7ce211c1d0e707f97b231ee7b817d41` | Registered `tgiv` marker and remote `main` agree; current Python/Bash CLI surface and installed-state drift case. |

The current delivery input attached to this task controls classification where
older board prose differs: Rust and Node/TypeScript are confirmed targets;
Swift and the SwiftPM-supported C family are confirmed targets; Python is a
reference implementation; Kotlin, Dart, and .NET are deferred.

## Evidence matrix — protocol implementations and manager baseline

| Surface and classification | Repository and path | Language / package manager | Build and launch entry points | Lock and integrity metadata | Recursive dependency shape | Runtime requirements | Mixed-language edges | Precompiled payload evidence and source-closure disposition |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| **Curator manager; implemented Go baseline** | [`go.mod`](https://github.com/relux-works/curator/blob/9ba552f04bacb91c4b643378cac928ed90bfb229/go.mod), [`cmd/curator/main.go`](https://github.com/relux-works/curator/blob/9ba552f04bacb91c4b643378cac928ed90bfb229/cmd/curator/main.go), [`internal/godriver`](https://github.com/relux-works/curator/tree/9ba552f04bacb91c4b643378cac928ed90bfb229/internal/godriver) | Go 1.25.5; Go modules | Source: `go install github.com/relux-works/curator/cmd/curator@…` or `go build ./cmd/curator`; launch `curator`. Hidden manager-owned worker dispatches the fixed build session. | `go.mod` + `go.sum`; one checked-in local replace to `agents/skills/skill-go-testing-tools/tuitestkit`. Manager source itself is not vendored. Build receipts bind source and toolchain identities. | `go list -m all` returned 28 module rows including the main module and local replace. The skill driver is narrower: fixed `go list -mod=vendor` then fixed `go build -mod=vendor`, network off, workspace off, cgo off, and validated recursive package inputs ([build vector](https://github.com/relux-works/curator/blob/9ba552f04bacb91c4b643378cac928ed90bfb229/internal/godriver/build.go), [graph validation](https://github.com/relux-works/curator/blob/9ba552f04bacb91c4b643378cac928ed90bfb229/internal/godriver/graph.go)). | Go 1.25.5 family for this source revision; installed observation: `curator v0.14.0-rc.3`. The built skill command is a host-native executable. | Manager implementation is Go. `go-v1` rejects cgo/C/C++/ObjC/SWIG inputs and host objects; narrowly accepted vendored assembly and `x/sys` behavior are specified by [decision 0005](https://github.com/relux-works/curator-spec/blob/dce6643c55434464c56f0fe20064db754cd58c61/decisions/0005-vendored-go-boundary-relaxation.md). | No deny-extension path occurred in the cited source-tree scan. Signed/released Curator binaries are manager/toolchain artifacts outside a skill source closure. **Implemented baseline; retain as the normative security behavior.** |
| **the alternative implementation (`csk`); independent Python protocol reference** | [`README.md`](the alternative-implementation repository (README link)), [`pyproject.toml`](the alternative-implementation repository (README link)), [`src/csk/cli.py`](the alternative-implementation repository (README link)), [`src/csk/builds/go_v1.py`](the alternative-implementation repository (README link)) | Python ≥3.11; PyPA/setuptools (`setuptools.build_meta`) | Build the distribution through the PEP 517 backend; console entry point `csk = csk.cli:main`; hidden Python worker implements the independent fixed Go boundary. | No Python dependency lock. Runtime `dependencies = []`; build requirements are ranged (`setuptools>=80`, `setuptools-scm[simple]>=8`); optional dev requirements are also ranges. Git revision/tag supplies the source identity, not a resolved package closure. | Zero declared runtime packages. Build backend has two declared requirements; dev extra has six. For skill Go builds, the manager validates the vendored Go graph independently of Curator and records protected-cache receipts. | Python ≥3.11 for the cited distribution. `go-v1` additionally requires an accepted native Go toolchain and a supported host; documentation at this revision qualifies native macOS/Windows. Installed observation is older: `csk 0.9.0`. | Python manager controls a Go compiler worker. That is a manager/toolchain boundary, not a Python package extension or a Node/Python shared adapter. | No deny-extension path occurred in the cited source tree. Packaging is not independently locked, but runtime code is pure Python with no declared third-party runtime dependency. **Reference implementation and differential oracle; do not create a new Python adapter by default.** |

The protocol itself confirms the current implemented build vocabulary: schema 6
admits `go-v1`, schema 7 adds `go-repository-v1`, and the canonical
[Go build fixture](https://github.com/relux-works/curator-spec/blob/dce6643c55434464c56f0fe20064db754cd58c61/conformance/v1/fixtures/go-build-skill/agent-skill.json)
ships `go.mod`, `vendor/modules.txt`, vendored source, and no prebuilt command.
Curator's parser accepts schemas 1–7 at the cited revision
([types.go](https://github.com/relux-works/curator/blob/9ba552f04bacb91c4b643378cac928ed90bfb229/internal/skillspec/types.go)).

## Evidence matrix — current skill-facing CLI surfaces

“Declared” means present in a machine manifest. “Repo-facing” means the skill
documentation or launcher invokes a CLI from its source repository even though
it is not yet declared as a Curator build command.

| Surface and classification | Repository and path | Language / package manager | Build and launch entry points | Lock and integrity metadata | Recursive dependency shape | Runtime requirements | Mixed-language edges | Precompiled payload evidence and source-closure disposition |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| **`task-board`, `tb-sessiond`, `task-board-tui`; declared Go build commands** | [`agent-skill.json`](https://github.com/relux-works/skill-project-management/blob/8dc0b71490214fe5ead6bf9cfde9574df084fd91/agent-skill.json); build roots [`tools/board-cli`](https://github.com/relux-works/skill-project-management/tree/8dc0b71490214fe5ead6bf9cfde9574df084fd91/tools/board-cli) and [`tools/board-tui`](https://github.com/relux-works/skill-project-management/tree/8dc0b71490214fe5ead6bf9cfde9574df084fd91/tools/board-tui) | Go 1.25.5; Go modules in two build roots | Manifest selects `go-v1` source dirs `tools/board-cli`, `tools/board-cli/cmd/tb-sessiond`, and `tools/board-tui`; Curator builds protected artifacts and publishes command shims. | Each build root has `go.mod`, `go.sum`, checked-in `vendor/modules.txt`, and vendored source. Snapshot/content hashes and Curator receipts bind admitted files and outputs. | `board-cli`: 11 vendored modules / 28 listed packages, including two same-repository modules and `agentquery`. `board-tui`: 39 vendored modules / 81 listed packages. The graph includes pure-Go assembly in `coder/websocket`, covered by protocol decision 0005. | Go 1.25.5 at build; host-native binaries at run. Installed `task-board` reports `0.24.3-3-g8dc0b71`. | No cgo. Vendored pure-Go assembly is an explicit Go-profile exception. Shared same-repository Go modules are materialized inside each vendor tree. | **Repository violation:** `tools/board-tui/task-board` is a tracked 14,501,282-byte mode-100755 Mach-O arm64 executable, blob `3cabcb35…`, inside a declared build root. `go version -m` identifies Go 1.25.5, VCS revision `5e5f095…`, `vcs.modified=true`, and older dependencies than the cited `go.mod`. It must be removed and the common binary diagnostic must reject it before this snapshot is eligible as source-only input. |
| **Installed/configured legacy `type: script` fleet: `band`; `bi-query`, `bi-catalog`, `bi-auth`; `gmr`; installed `grafana`; `sentry-api`, `sentry-cli-auth`; `wk`; `yt`, `ytx`** | Exact live/configured/installed revisions are reconciled in the table below; installed launch roots are under `$HOME/.the alternative implementation/runtime/<skill>/<commit>/`, and marker evidence is under each registered project's `.agents/skills/*/.csk-install.json`. The installed `grafana@a557671…` state is deliberately distinct from the live feature state in the next row. | Bash/CMD launchers into Python; `venv` + `pip` requirements bootstrap | the alternative implementation installs command shims pointing to commit-keyed runtime-root launchers. Launchers call `python3 run_in_skill_venv.py <entrypoint>`; missing runtimes invoke `python -m venv` and `python -m pip install -r requirements.txt`. | Install markers bind source commit, content SHA-256, file list, and commands. Python requirements are **not locked or hashed**: `keyring>=25,<26` (Band/BI/installed Grafana/Sentry), `keyring>=24` (Wiki), `youtrack-cli==0.22.2` (YouTrack), and no file for GitLab. Exact top-level YouTrack version still leaves transitive artifacts unresolved. | Transitive closure is selected online by `pip` at bootstrap time. Installed YouTrack runtime proves a substantial transitive graph including Selenium, cryptography, cffi, pydantic-core, charset-normalizer, and their dependencies. | Launchers require a compatible Python 3; the inspected YouTrack bootstrap accepts Python 3.10–3.13. Commands additionally need service credentials/network according to each skill. | Python packages include Rust/C extension edges after wheel installation (`cryptography` Rust extension, cffi, pydantic-core, charset-normalizer, mypyc output). These edges are absent from the top-level requirement text. | Source runtime roots are script text, but generated `.venv` trees contain `.so` native extensions and bundled `.exe` launchers/selenium-manager. **Existing behavior is a reference input, not an admissible source-only closure.** A future Python policy must require auditable source distributions and offline local builds or fail closed; this task does not start that adapter. |
| **`analyze`, `query`, `grafana-auth`; live source-checkout Python CLI state, not the installed Grafana tag** | Private repository [`csk-skill.json`](https://gitlab.wildberries.ru/portals/agentic-infra/skills/skill-grafana/-/blob/6234e2a3ef35415b765f74cca21c8ad7846f521e/csk-skill.json), [`SKILL.md`](https://gitlab.wildberries.ru/portals/agentic-infra/skills/skill-grafana/-/blob/6234e2a3ef35415b765f74cca21c8ad7846f521e/SKILL.md), launchers [`grafana-analyze`](https://gitlab.wildberries.ru/portals/agentic-infra/skills/skill-grafana/-/blob/6234e2a3ef35415b765f74cca21c8ad7846f521e/scripts/grafana-analyze), [`grafana-query`](https://gitlab.wildberries.ru/portals/agentic-infra/skills/skill-grafana/-/blob/6234e2a3ef35415b765f74cca21c8ad7846f521e/scripts/grafana-query), [`grafana-auth`](https://gitlab.wildberries.ru/portals/agentic-infra/skills/skill-grafana/-/blob/6234e2a3ef35415b765f74cca21c8ad7846f521e/scripts/grafana-auth), [`runtime_support.py`](https://gitlab.wildberries.ru/portals/agentic-infra/skills/skill-grafana/-/blob/6234e2a3ef35415b765f74cca21c8ad7846f521e/scripts/runtime_support.py), and [`requirements.txt`](https://gitlab.wildberries.ru/portals/agentic-infra/skills/skill-grafana/-/blob/6234e2a3ef35415b765f74cca21c8ad7846f521e/scripts/requirements.txt); live root `/Users/iv/agents/skills/skill-grafana`. | Python application plus Bash and CMD launchers; `venv` + `pip` | Manifest maps `analyze` → `scripts/grafana-analyze` → `analyze.py`, `query` → `scripts/grafana-query` → `grafana_query.py`, and `grafana-auth` → `scripts/grafana-auth` → `grafana_auth.py`; every Unix launcher execs system `python3 run_in_skill_venv.py`, which creates/uses `.venv`. `make install` copies a repo-local skill; `make skill` invokes `bootstrap_runtime.py`. | Clean Git `HEAD` `6234e2a…` binds the 41 tracked files. Relevant Git blob IDs are `004713c…` (manifest), `e836429…` (requirements), and `646d15a…` (runtime support). There is no feature-revision install marker, dependency lock, constraints file, checked-in pip report, or requirement hash; the only input is ranged `keyring>=25,<26`. | A clean macOS/Python 3.13 pip dry-run on 2026-08-11 selected five `py3-none-any` wheels: keyring 25.7.0, jaraco.classes 3.4.0, jaraco.context 6.1.2, jaraco.functools 4.6.0, and more-itertools 11.1.0. The report had artifact SHA-256 values, but it was generated evidence rather than admitted metadata. Keyring also declares platform-conditional Windows/Linux and older-Python dependencies, so this five-package graph is host-specific and can drift at the next online bootstrap. | `runtime_support.py` accepts CPython 3.10–3.13 and requires `venv`/`pip`; the first invocation may need package-index network access. Runtime reads Grafana over HTTPS, requires service credentials in a keyring backend, and may launch `open`/`xdg-open` for reports. The cited tests ran with Python 3.13. | Explicit Bash/CMD → system Python → venv Python launch chain plus OS keyring and Grafana HTTPS process/service boundaries; no application FFI or compiled mixed-language target occurred in the pinned macOS resolution. Platform-conditioned keyring dependencies remain part of the unresolved graph. | A `git archive`/`file` scan classified all 41 tracked files as source/text and found no compiled signature; the denied-extension query returned no match. After validation, the live checkout contained 31 ignored CPython 3.14 `.pyc` files totaling 373,177 bytes. They are not part of `6234e2a…`, but any frozen working-copy snapshot must exclude or reject them. Online wheel selection and an uncaptured `.venv` likewise fail the offline source-closure invariant. **Current Python ecosystem evidence only, not authorization for a new Python adapter.** |
| **`ios-app-manager`; repo-facing Go CLI** | [`tuist-starter/go.mod`](https://github.com/relux-works/skill-ios-app-manager/blob/a8523bbffc5399af68970ba772b9c67ba7cdd5d3/tuist-starter/go.mod), [`Makefile`](https://github.com/relux-works/skill-ios-app-manager/blob/a8523bbffc5399af68970ba772b9c67ba7cdd5d3/tuist-starter/Makefile), [`SKILL.md`](https://github.com/relux-works/skill-ios-app-manager/blob/a8523bbffc5399af68970ba772b9c67ba7cdd5d3/SKILL.md) | Go 1.25.5; Go modules | `make build` → `go build -o ios-app-manager ./cmd/ios-app-manager`; setup symlinks the output to `$HOME/.local/bin/ios-app-manager`; skill invokes that name. | `go.mod` + `go.sum`; no vendor tree. Build cache/module cache are local `.cache` directories but no captured offline closure is committed. | Cobra and yaml.v3 direct; mousetrap and pflag indirect (four external modules total). | Go 1.25.5 to build; runtime workflows additionally use Tuist and Xcode tools on macOS. | The CLI build is Go-only. It generates and orchestrates Swift/Tuist/Xcode project content at runtime; those are downstream tool inputs, not compiler inputs to the Go CLI. | Cited Git tree has no denied extension. Current setup builds before copying the repository and does not exclude the output name; the installed skill contains Mach-O `tuist-starter/ios-app-manager` and `ios-app-manager-ldflags`. **Migrate to a manifest build command only after vendoring/capturing dependencies, and exclude/reject pre-existing outputs from admitted source.** |
| **`appraise`; repo-facing Go CLI** | [`tools/appraise/go.mod`](https://github.com/relux-works/skill-product-appraisal/blob/01daa3436f81a6eedefb84a24d9ffbb7e337cc46/tools/appraise/go.mod), [`Makefile`](https://github.com/relux-works/skill-product-appraisal/blob/01daa3436f81a6eedefb84a24d9ffbb7e337cc46/tools/appraise/Makefile), [`product-appraisal/SKILL.md`](https://github.com/relux-works/skill-product-appraisal/blob/01daa3436f81a6eedefb84a24d9ffbb7e337cc46/product-appraisal/SKILL.md) | Go 1.25.5; Go modules | `make install` → `go build -o appraise .` and symlink to `$HOME/.local/bin/appraise`; the skill mandates `appraise` for calculations. | `go.mod` + `go.sum`; no vendor. `replace github.com/relux-works/skill-agent-facing-api/agentquery => ../../../skill-agent-facing-api/agentquery` crosses the skill repository root. | `agentquery` + Cobra direct; mousetrap/pflag indirect. The replaced `agentquery` module itself has Cobra/pflag dependencies and no vendor at [revision `656ad0a…`](https://github.com/relux-works/skill-agent-facing-api/blob/656ad0a73dbffc92e732ed95c39a7ae3197f28f1/agentquery/go.mod). | Go 1.25.5 to build; native host executable at run. `appraise` was documented but not present on `PATH` in the observed machine state. | Go-only build, but a cross-repository source edge and build order exist: `agentquery` must be admitted before `appraise`. | No denied extension in the cited tree. **Current relative replace escapes a single skill build root and cannot pass `go-v1`; vendor it or model it as an explicitly locked external repository before declaring the command.** |
| **`agents-infra`; repo-facing Go CLI** | [`tools/agents-infra/go.mod`](https://github.com/relux-works/alexis-agents-infra/blob/ccf0daf444aa1f6ce13119151a860855d7360837/tools/agents-infra/go.mod), [`scripts/setup.sh`](https://github.com/relux-works/alexis-agents-infra/blob/ccf0daf444aa1f6ce13119151a860855d7360837/scripts/setup.sh), [`SKILL.md`](https://github.com/relux-works/alexis-agents-infra/blob/ccf0daf444aa1f6ce13119151a860855d7360837/SKILL.md) | Go 1.21.0; Go modules | Setup runs `go -C tools/agents-infra build -trimpath -ldflags … -o .temp/bin/agents-infra`, copies the output to `$HOME/.local/bin`, then launches `agents-infra setup/doctor`. Project-local launchers may rebuild before invocation. | `go.mod` + `go.sum`; no vendor. Embedded version/commit/date are build inputs but the installed observation reported `dev commit=unknown build_date=unknown`. | Two external modules (`atomic`, `go-toml/v2`); no indirect module requirements are declared. | Go ≥1.21 as declared. Optional setup paths use Homebrew LLVM/lldb-mcp and PDF tools; those are external system/toolchain dependencies. | CLI build is Go-only; it orchestrates shell/PowerShell configuration and external native tools. | No denied extension in the cited source tree; setup emits a local binary under `.temp` and copies it outside the skill context. **Add a manifest and offline Go closure before Curator admission; keep optional system tools outside the package closure.** |
| **Six SwiftPM executables: `extract-screenshots`, `snapshotsdiff`, `ios-device-build`, `ios-e2e-runner`, `e2e-fake-peer`, `e2e-listener-fake-peer`; repo-facing Swift target** | [`Package.swift`](https://github.com/relux-works/skill-ios-testing-tools/blob/bd59caaf4bb712f35d4b8b73141ce28999cc13cb/Package.swift), [`README.md`](https://github.com/relux-works/skill-ios-testing-tools/blob/bd59caaf4bb712f35d4b8b73141ce28999cc13cb/README.md), skill [`SKILL.md`](https://github.com/relux-works/skill-ios-testing-tools/blob/bd59caaf4bb712f35d4b8b73141ce28999cc13cb/agents/skills/ios-testing-tools/SKILL.md) | Swift tools 6.2; Swift Package Manager | Documented launch is `swift run [--package-path <repo>] <product>` or wrapper scripts. Six executable products and three library products are declared. Current setup copies only `agents/skills/ios-testing-tools`, so the Swift package remains a source-repository dependency rather than a declared installed command. | **No `Package.resolved` is tracked at the cited revision.** `Package.swift` uses ranges `swift-nio` from 2.65.0 and Yams from 5.1.3. An observed generated `Package.resolved` v3 pinned five repositories and an `originHash`, but that local file is not authoritative source metadata. | Observed resolution: swift-nio 2.101.3, Yams 5.4.0, swift-atomics 1.3.1, swift-collections 1.6.0, swift-system 1.7.4, each with exact Git revision. | Package declares iOS 15/macOS 13; project docs require Swift 6.2+ and Xcode 26+. CLI products are macOS terminal tools and use Xcode/xcresult/device utilities according to command. | Real C edges: Yams→`CYaml`; swift-atomics→`_AtomicsShims`; swift-system→`CSystem`; swift-nio→`CNIOAtomics`, `CNIOPosix`, `CNIOSHA1`, platform shims, and `CNIOLLHTTP`. `snapshotsdiff` links AppKit/CoreGraphics. No C++, ObjC, or ObjC++ target was found in this graph. | No denied extension in the cited Git tree. A legacy installed copy contains generated `.build` Mach-O executables/dSYM payloads; those are local outputs, not tracked by the cited revision. **Confirmed adapter target, but it must capture a full resolved source graph and reject `.build`/prebuilt payloads before admission.** |
| **`android-device-telemetry`; repo-facing Go launcher** | [`scripts/android-device-telemetry`](https://github.com/relux-works/skill-android-testing-tools/blob/b79449cc3e1767680b069ae314c850a1d93c6f99/agents/skills/android-testing-tools/scripts/android-device-telemetry), [`tools/android-device-telemetry/go.mod`](https://github.com/relux-works/skill-android-testing-tools/blob/b79449cc3e1767680b069ae314c850a1d93c6f99/agents/skills/android-testing-tools/tools/android-device-telemetry/go.mod) | Go 1.21; Go modules, standard library only | Bash launcher executes `go run "$TOOL_DIR/main.go"`; skill invokes `scripts/android-device-telemetry snapshot\|monitor`. | `go.mod`; no `go.sum` is needed because there are no external modules. No manifest or build receipt. | One main package, standard library only. | Go 1.21 at every invocation plus Android `adb` and attached-device access. | Go CLI communicates with Android tools/processes but has no compiled mixed-language input. | No precompiled payload in this source subtree. **Low-complexity migration candidate: declare a Go build command and replace runtime `go run` with a Curator-built receipt-bound artifact.** |
| **Android-repository `snapshotsdiff`; repo-contained Swift utility** | [`toolkit/snapshotsdiff/Package.swift`](https://github.com/relux-works/skill-android-testing-tools/blob/b79449cc3e1767680b069ae314c850a1d93c6f99/toolkit/snapshotsdiff/Package.swift) | Swift tools 5.9; SwiftPM | `swift run snapshotsdiff`/bare `snapshotsdiff` is documented for ad-hoc two-image comparison. Latest setup does not copy `toolkit`, so this is not a declared installed command. | No external package dependencies; no `Package.resolved` needed for this graph. Source revision is the dependency identity. | Single executable target; Foundation/toolchain libraries plus AppKit and CoreGraphics frameworks. | macOS 13+, Swift 5.9-compatible toolchain. | Swift plus system frameworks; no source C-family dependency. | No denied artifact in this package tree. **Valid small Swift conformance seed, but ownership should be deduplicated with the richer iOS testing `snapshotsdiff` surface before migration.** |
| **`exchange` + `exchange-scraper`; current repo-facing mixed Go/Swift CLI** | [`SKILL.md`](https://github.com/relux-works/skill-currency-exchange/blob/c29210aa6eb4cc0f64f307fa30561ac80feb6b3b/SKILL.md), Go [`go.mod`](https://github.com/relux-works/skill-currency-exchange/blob/c29210aa6eb4cc0f64f307fa30561ac80feb6b3b/exchange/go.mod) / [`go.sum`](https://github.com/relux-works/skill-currency-exchange/blob/c29210aa6eb4cc0f64f307fa30561ac80feb6b3b/exchange/go.sum), Swift [`Package.swift`](https://github.com/relux-works/skill-currency-exchange/blob/c29210aa6eb4cc0f64f307fa30561ac80feb6b3b/exchange-scraper/Package.swift), [`build.sh`](https://github.com/relux-works/skill-currency-exchange/blob/c29210aa6eb4cc0f64f307fa30561ac80feb6b3b/scripts/build.sh), [`install.sh`](https://github.com/relux-works/skill-currency-exchange/blob/c29210aa6eb4cc0f64f307fa30561ac80feb6b3b/scripts/install.sh), and [`executor.go`](https://github.com/relux-works/skill-currency-exchange/blob/c29210aa6eb4cc0f64f307fa30561ac80feb6b3b/exchange/internal/scraper/executor.go); installed marker `/Users/iv/.the alternative implementation/global/skills/skill-currency-exchange/.csk-install.json` | Go 1.25.5 modules plus Swift tools 6.0 / SwiftPM | Both scripts build Swift release first, then Go. Install copies Swift `ExchangeScraper` as `exchange-scraper` and Go output as `exchange`, then launches `exchange ... schema()` for verification. At runtime Go uses `exec.CommandContext` to launch `exchange-scraper google\|ratam` by `PATH`. The pinned global marker ships only `SKILL.md` and the two scripts with `commands: []`; the skill nevertheless mandates `exchange`. Neither binary was on `PATH` in the observed state. | Git tag/commit and marker content hash bind the shipped three-file skill copy. Go has `go.mod` + 53-line `go.sum` (30 unique module/version pairs), no vendor, and two placeholder versions replaced by absolute `/Users/alexis/...` paths. Swift declares `swift-argument-parser` `1.5.0..<2.0.0`; **no `Package.resolved` is tracked**. An ignored local resolver file pinned 1.8.2 / `6a52f325…`, but it is generated, machine-local evidence rather than repository integrity metadata. | Go declares four direct requirements (two absolute cross-repository sources, Cobra, `x/text`) plus 20 indirect requirements; `go list -m all` fails before enumeration because the absolute replacements do not exist on this machine. The Swift root has one executable, one library, one test target, and one external package; the observed 1.8.2 checkout declares no external packages of its own but does declare two command plugins. | Go 1.25.5 build toolchain; Swift 6-compatible toolchain on macOS 13+; WebKit and network access for scraping at runtime. Observed compatible host: Go 1.25.5, Apple Swift 6.3.2, macOS 26.5 arm64. | Explicit two-product graph and build order: Swift scraper must be built/published before the Go frontend can satisfy runtime calls. The boundary is JSON over a subprocess, not FFI. Swift additionally links the macOS WebKit system framework; this package has no source C-family target. | A compiled-signature scan of the pinned Git tree found zero denied payloads. The ignored local `.build` contains a 2,764,816-byte Mach-O scraper, an XCTest bundle, and compiled `GenerateManual` / `GenerateDoccReference` plugin executables; no local Go output was present. **Best real mixed Go/Swift migration input, but currently fail closed: make both cross-repository Go edges immutable/portable, capture the Swift resolution, model build/runtime order, and exclude/reject `.build` before admission.** |
| **`tg-telethon` + auth helpers; current repo-facing Python CLI with content drift** | [`SKILL.md`](https://github.com/ivanopcode/telegram-telethon/blob/b9a76b01e7ce211c1d0e707f97b231ee7b817d41/SKILL.md), [`requirements.txt`](https://github.com/ivanopcode/telegram-telethon/blob/b9a76b01e7ce211c1d0e707f97b231ee7b817d41/requirements.txt), [`Makefile`](https://github.com/ivanopcode/telegram-telethon/blob/b9a76b01e7ce211c1d0e707f97b231ee7b817d41/Makefile), [`bootstrap.sh`](https://github.com/ivanopcode/telegram-telethon/blob/b9a76b01e7ce211c1d0e707f97b231ee7b817d41/scripts/bootstrap.sh), [`tg-telethon`](https://github.com/ivanopcode/telegram-telethon/blob/b9a76b01e7ce211c1d0e707f97b231ee7b817d41/scripts/tg-telethon), and [`telegram_telethon.py`](https://github.com/ivanopcode/telegram-telethon/blob/b9a76b01e7ce211c1d0e707f97b231ee7b817d41/scripts/telegram_telethon.py); installed marker `/Users/iv/Developer/IV/tgiv/.agents/skills/telegram-telethon/.csk-install.json` | Python plus Bash; `venv` + `pip` | `make install` / `setup.sh` launches `python3 scripts/setup_main.py`; the managed installer copies the skill and runs bootstrap. Bootstrap creates `.venv`, installs requirements, and writes `tg-telethon`, `tg-telethon-auth-login`, and `tg-telethon-auth-logout` shims. The primary launcher execs the venv interpreter on `telegram_telethon.py`. All three live shims currently point to `/Users/iv/agents/skills/telegram-telethon`, not the registered `tgiv` copy; `tg-telethon --help` exited 0. | Commit `b9a76b0…` and the `tgiv` marker's `sha256:dafcf9…` bind its declared 11-file installed subset. Source pins only `telethon==1.42.0`; there is no lock, hash mode, constraints file, or captured artifact set. The installed subset omits both the referenced `requirements.txt` and `Makefile`, so its bootstrap input is absent. | A clean 2026-08-11 pip dry-run resolved Telethon 1.42.0 → pyaes 1.6.1 (sdist/build-isolation path), rsa 4.9.1 → pyasn1 0.6.4. The live venv instead has Telethon 1.42.0, pyaes 1.6.1, rsa 4.9.1, and pyasn1 0.6.3. This proves the exact top-level pin does not bind the transitive graph or build inputs. | Source declares macOS, `python3` with no version floor, and `/usr/bin/security`; runtime needs Telegram network/API access and Keychain credentials. Observed live interpreter: CPython 3.14.4. | Pure-Python application graph plus subprocess calls to macOS `security`; Telethon supplies the MTProto/network boundary. `tg-stories` is a documentation-only external companion and is not shipped by this repository. No application native extension is required by the observed Telethon graph. | Pinned Git compiled-signature scan found zero denied payloads and `.gitignore` excludes venv/bytecode. The registered copy contains undeclared `/Users/iv/Developer/IV/tgiv/.agents/skills/telegram-telethon/scripts/__pycache__/telegram_telethon.cpython-314.pyc` (63,929 bytes, SHA-256 `dd5ddeee…`) outside the marker list; `csk status --all` reports `content-drift`. The live source checkout has 605 generated `.pyc` files (601 inside `.venv`) and pip bundles six Windows `.exe` launchers. **Current Python baseline only: repair installed source completeness and reject bytecode/venv payloads; do not turn it into a default new Python adapter deliverable.** |
| **`glab`, `sentry-cli`; declared external-system command class** | Active `skill-gitlab/csk-skill.json` and `skill-sentry/csk-skill.json` under `/Users/iv/agents/skills`; installed markers at `ecf620c…` / `85a3e37…` and global GitLab `bdce7f7…` | Externally installed native CLIs; package manager/source ownership is outside these skills | the alternative implementation `type: system` declarations require the named command on `PATH`; skill launchers invoke it directly rather than building or copying a payload. | Skill commits/markers bind only the declaration and hint, not the external executable, its package graph, signature, or provenance. Observed paths were `/opt/homebrew/bin/glab` and `/opt/homebrew/bin/sentry-cli`; observed versions were glab 1.90.0 and sentry-cli 3.4.0. | Not resolved by a Curator language adapter and therefore not part of a skill source closure. Their own transitives remain unenumerated by the current manifests. | Compatible host executables plus service network/auth requirements. | External process boundary only. | No binary is vendored by the cited skills. **Keep outside language adapters and add a separate system-dependency admission/version/provenance policy; do not misclassify these installed binaries as source-closure payloads.** |

### Observed SwiftPM resolver state (local, not repository-authoritative)

The generated root `Package.resolved` in the installed legacy copy had schema
version 3 and `originHash`
`d872e0f0cb2f9753ed1f20d3f9837580a455548f85f261932486279062a4d620`.
Every checkout HEAD matched its pin:

| Package | Version | Immutable revision | Relevant graph edge |
| --- | --- | --- | --- |
| [swift-nio](https://github.com/apple/swift-nio/blob/0b18836bd8b0162e7e17a995a3fbee20ed8f3b2b/Package.swift) | 2.101.3 | `0b18836bd8b0162e7e17a995a3fbee20ed8f3b2b` | Swift targets plus multiple `CNIO*` C targets; depends on atomics, collections, and system. |
| [Yams](https://github.com/jpsim/Yams/blob/3d6871d5b4a5cd519adf233fbb576e0a2af71c17/Package.swift) | 5.4.0 | `3d6871d5b4a5cd519adf233fbb576e0a2af71c17` | Swift `Yams` depends on C `CYaml`. |
| [swift-atomics](https://github.com/apple/swift-atomics/blob/0442cb5a3f98ab802acb777929fdb446bda11a34/Package.swift) | 1.3.1 | `0442cb5a3f98ab802acb777929fdb446bda11a34` | Swift `Atomics` depends on C `_AtomicsShims`. |
| [swift-collections](https://github.com/apple/swift-collections/blob/a0cb0954ecb21e4e31b0070e6ed5674e8556685a/Package.swift) | 1.6.0 | `a0cb0954ecb21e4e31b0070e6ed5674e8556685a` | Swift transitive collections; its checkout also contains the non-product Rust benchmark helper noted below. |
| [swift-system](https://github.com/apple/swift-system/blob/b5544ba79a70a0cb3563e75bf26dc198d6b40ed3/Package.swift) | 1.7.4 | `b5544ba79a70a0cb3563e75bf26dc198d6b40ed3` | Swift `SystemPackage` depends on C `CSystem`. |

Separately, the ignored currency source-checkout
`exchange-scraper/Package.resolved` had SHA-256
`78c62dc5ae7f18d8d8217192efabdcc1fe17b0e38395a2f1ffd941adf8e7b16b`,
schema 3, origin hash
`b4618dbc572fd852bd1df490aeb52f45e3a513c4099fb01b5bfad9b14778d885`,
and one pin: [swift-argument-parser 1.8.2](https://github.com/apple/swift-argument-parser/blob/6a52f3251125d74daf04fcbd5e6f08a75d074382/Package.swift) at
`6a52f3251125d74daf04fcbd5e6f08a75d074382`. Because Git at `c29210a…`
does not track that file, none of these values is an authoritative package
input for a clean checkout.

### Exact live/configured/installed manifest revision inventory

All eight physical manifests below are under `/Users/iv/agents/skills`. The
working-copy column is read from each clean repository `HEAD`; configured
project/global refs are resolved from `csk list --paths` and `csk global list`;
and installed state is read from project/global markers plus the matching
status commands. The union has 15
distinct commits. Equal command sets do not collapse unequal implementation
revisions; they only show that no additional command name is introduced.

| Skill | Clean working-copy authority | Configured target authority | Installed marker authority | Command reconciliation | Dependency input |
| --- | --- | --- | --- | --- | --- |
| `skill-band` | `e49922a0ce8d8123ab532d16725d35666427ba41`, `master` | revision `f2154e9c12ef3dbe515663b2f1e06ade4ae392c6` | `e49922a0ce8d8123ab532d16725d35666427ba41`; `update-available` | Both immutable manifests declare `band`; configured target is retained even though it is not installed. | Both revisions: `keyring>=25,<26` |
| `skill-bi` | `e2288fe9e617320001aa49f3d573f131b9b11890`, `master` | tag `v1.1.1` → `8f0109a2483502995e24b54c924cfce042a9ffa7` | `8f0109a2483502995e24b54c924cfce042a9ffa7`; `up-to-date` | Both revisions declare `bi-query`, `bi-catalog`, and `bi-auth`. | Both revisions: `keyring>=25,<26` |
| `skill-gitlab` | `7d2fada95433bb4c7e454cc51ea7265bac9d4a3f`, `master`; the local checkout has no `origin` URL | project tag `v1.2.0` → `ecf620c91b147119c01619182c88652b977a2158`; global tag `v1.1.0` → `bdce7f7e2c26345f6cacabd73dceb3d586e0a004` | Project `ecf620c…` and global `bdce7f7…`; both `up-to-date` | All three revisions declare script `gmr` and external-system `glab`. | No Python requirement file |
| `skill-grafana` | [`6234e2a3ef35415b765f74cca21c8ad7846f521e`](https://gitlab.wildberries.ru/portals/agentic-infra/skills/skill-grafana/-/tree/6234e2a3ef35415b765f74cca21c8ad7846f521e), `feature/oparin/PMA-23845` | tag `v1.1.0` → `a557671bce2ef7c6547293dc4b28fcc8636df51a` | `a557671bce2ef7c6547293dc4b28fcc8636df51a`; `up-to-date`; marker content `sha256:a6673e35…` | **Different surfaces:** live `HEAD` declares `analyze`, `query`, `grafana-auth`; configured/installed tag declares only `grafana`. Both states have full matrix coverage above. | Both revisions: `keyring>=25,<26` |
| `skill-sentry` | `cf13b3258d5dfe109acb7e97110ce1b7360105f0`, `feature/oparin/PMA-23641` | tag `v1.1.1` → `85a3e3795afc0c59d00188289244e787ef4d53be` | `85a3e3795afc0c59d00188289244e787ef4d53be`; `up-to-date` | Both revisions declare scripts `sentry-api`, `sentry-cli-auth` and external-system `sentry-cli`. | Both revisions: `keyring>=25,<26` |
| `skill-wiki` | `c2765ccddcd73df1faa1a313ce4066b5468992fa`, `main` | same exact revision | same exact revision; `up-to-date` | One state after deduplication; declares `wk`. | `keyring>=24` |
| `skill-wiki-memory` | `2c04440fb413bfca987492b812207ffde988fb8c`, `main` | same exact revision | same exact revision; `up-to-date` | One state after deduplication; empty command map and runtime roots. | None; explicit no-CLI state |
| `skill-youtrack` | `25bf9e85c9e61dbe17494117310037e9504ce39c`, `master` | project tag `v1.1.1` → same commit; global tag `v1.2.0` → `58eae4f26057f93c1c9ad8fe3f5ba510346365d5` | Project `25bf9e8…` and global `58eae4f…`; both `up-to-date` | Both revisions declare `yt` and `ytx`. | Both revisions: `youtrack-cli==0.22.2` |

### Exact installed marker evidence for the rework surfaces

- `/Users/iv/Developer/Wildberries/agentic infra/.agents/skills/skill-grafana/.csk-install.json`
  binds configured tag `v1.1.0`, commit
  `a557671bce2ef7c6547293dc4b28fcc8636df51a`, content
  `sha256:a6673e35f89e0c5055f5e8b4aa52d73b398fd3f18c3a5c25b903c9479239e758`,
  ten declared installed files, and only command `grafana`. The clean live
  source checkout has no feature-revision marker; its independent authority is
  Git `HEAD` `6234e2a3ef35415b765f74cca21c8ad7846f521e` plus the checked-in
  manifest and launch sources at that commit.
- `/Users/iv/.the alternative implementation/global/skills/skill-currency-exchange/.csk-install.json`
  binds commit `c29210aa6eb4cc0f64f307fa30561ac80feb6b3b`, content
  `sha256:e0f1be78e8d9f779db14c273edc6fd70693742f744ea5a182706a2cac43673ef`,
  exactly `SKILL.md`, `scripts/build.sh`, and `scripts/install.sh`, and an empty
  command list.
- `/Users/iv/Developer/IV/tgiv/.agents/skills/telegram-telethon/.csk-install.json`
  binds commit `b9a76b01e7ce211c1d0e707f97b231ee7b817d41`, content
  `sha256:dafcf911194e617af96f63bbe520062b88604895fee35a70491b0c1d4e9f5063`,
  11 declared files, and an empty command list. Its undeclared
  `scripts/__pycache__/telegram_telethon.cpython-314.pyc` has SHA-256
  `dd5ddeee6390a846638ec26ab3ca534367f5c9a61aa45468158e34f8872a5167`.

### Post-review active-root and registered-project closure ledger

This ledger is the disposition proof for the inclusion rule; “no CLI” is an
explicit scan result, not an omission.

| Scan result | Exact observed surfaces | Matrix disposition |
| --- | --- | --- |
| Curator-declared build commands | The sole active `agent-skill.json` is project-management and declares `task-board`, `tb-sessiond`, and `task-board-tui`, all `go-v1`. | Full declared-Go row. |
| the alternative implementation live manifest checkouts | The eight clean Git checkouts at the the alternative implementation skills root resolve to the exact `HEAD` revisions in the reconciliation table. Seven command sets are already represented by the legacy fleet/no-CLI rows; live Grafana `6234e2a…` uniquely adds `analyze`, `query`, and `grafana-auth`. | Full legacy-fleet row, explicit no-CLI disposition, and a dedicated full Grafana feature-revision matrix row. Physical manifest files are discovery locators only. |
| Configured and installed the alternative implementation states | Registered project markers expose Band; BI; GitLab; installed Grafana `a557671…`; Sentry; Wiki; and YouTrack. Global markers additionally expose GitLab `bdce7f7…` and YouTrack `58eae4f…` variants. All configured refs resolve to the installed commit except Band, whose configured `f2154e9…` is `update-available` over installed/live `e49922a…`. | Exact revision reconciliation retains all 15 distinct commits and deduplicates only identical state revisions. The configured Band target is retained even though it is not installed; installed Grafana remains separate from live Grafana. |
| Empty-command markers with shipped repo-facing CLIs | Global currency exchange `c29210a…`; global agents infra `513d058…`; iOS App Manager `72ee678…` and `57ba95c…`; registered Telegram `b9a76b0…`. Their `SKILL.md`/setup/launchers name commands even though marker `commands` is empty. | Currency, Agents Infra, iOS App Manager, and Telegram full matrix rows. The same rule is applied to all four. |
| Configured-but-missing project installs | iOS Testing Tools is `missing` in both `memori-fresh` and `spotify-duet`. | No installed duplicate is claimed; the authoritative current SwiftPM repository surface remains included through the active source roots and revision ledger. |
| Active empty-command marker with no shipped CLI | `skill-wiki-memory@2c04440…`; an otherwise unregistered `skill-deploy-lab@a510224…` marker was also found under the active `agentic infra` project root and contains only `SKILL.md`. | Explicit no-CLI disposition; neither creates a language-adapter surface. Any external commands mentioned in prose remain system/runtime dependencies. |
| Active non-the alternative implementation repo-facing surfaces | Product Appraisal `appraise`; iOS Testing Tools SwiftPM products; Android Testing Tools Go telemetry and Swift utility. | Full rows above. Their absence from the alternative implementation command arrays does not exclude them. |
| Library-only source dependencies | `skill-agent-facing-api/agentquery` and `skill-go-testing-tools/tuitestkit` are imported by Go CLIs but expose no independently admitted skill-facing command in this scan. | Recorded as cross-repository dependency edges, not duplicate CLI rows. |
| External system declarations | `glab` and `sentry-cli`; both were on `PATH`. | Separate external-system row. Runtime tools such as `security`, `adb`, Xcode/Tuist, and WebKit remain explicit requirements in their owning surface rows. |
| Remaining active docs-only skills | Core Data, SwiftUI, architecture diagrams, testing methodology, and skill-management guidance did not ship a language-adapter candidate command under this inclusion rule. | Explicitly outside the CLI matrix; no Node/Cargo/Dart/.NET surface is inferred from documentation or metadata alone. |

## Target and deferred classification matrix

| Ecosystem | Delivery classification | Current implementation/surface evidence | Inventory consequence |
| --- | --- | --- | --- |
| Go | Implemented baseline | Curator `go-v1`/`go-repository-v1`; project-management manifest; iOS App Manager, Appraise, Agents Infra, and currency exchange repo-facing CLIs; Android telemetry. | Use the fixed vendored/offline worker and its receipts as the security baseline. Migrate only source-clean, closed graphs; currency's absolute replacements fail this gate today. |
| Swift | Confirmed current target | Six-product iOS testing SwiftPM package, a small Android-repository utility, and the runtime-required currency scraper. | Real migration inputs exist. Currency exchange is the strongest actual mixed Go/Swift build-order case; none is a Curator Swift build command today. |
| C | Confirmed current target through SwiftPM | Actual C targets in Yams, swift-atomics, swift-system, and swift-nio. | Adapter graph must include C source, Swift→C edges, target platform, C compiler identity, and build order. |
| C++, Objective-C, Objective-C++ | Confirmed conditional targets | No current skill CLI edge found at inspected revisions. | Do not infer support from Swift/C success. Add focused fixtures and fail closed wherever SwiftPM cannot prove the complete source boundary. |
| Rust | Confirmed current target | No Cargo manifest or lock in the revision-pinned estate scan. One `.rs` benchmark helper was present only under the generated swift-collections checkout (`Benchmarks/Sources/RustBenchmarks/VecDequeBenchmarks.rs`), with no Cargo package or skill command. | New fixture/reference package required; there is no current CLI to migrate first. |
| Node/TypeScript | Confirmed current target | No Node package/lock or JS/TS source in the revision-pinned estate scan. The only installed `package.json` was SwiftUI skill-discovery metadata (`pi.skills`) with no `name`, `bin`, dependencies, or lock. | New fixture/reference package required. Do not treat metadata-only `package.json` as a Node CLI. |
| Python | Existing protocol/reference and CLI ecosystem baseline | the alternative implementation independent manager, legacy installed Python command fleet, live Grafana feature commands, and current Telegram Telethon CLI. | Compare policy and conformance behavior; no new Python adapter by default. Legacy/Grafana online `pip` bootstrap and Telegram's single top-level pin are not source-closure compliant. |
| Kotlin/JVM | Deferred | Android repository contains a Java-21 Kotlin CLI that creates a fat JAR and a committed Gradle wrapper JAR. | Preserve as backlog evidence only; do not let it gate current research or implementation. |
| Dart | Deferred | No surface found. | No current commitment. |
| .NET | Deferred | No surface found. | No current commitment. |

## Node/TypeScript versus Python — explicit correction

The assumed relationship is **protocol-level, not implementation-level**:

- the alternative implementation states that it is an “independent Python implementation” of the
  shared protocol in its
  [README](the alternative-implementation repository (README link)).
- Its [package metadata](the alternative-implementation repository (README link))
  exposes only the Python `csk` console entry point and no runtime dependency.
- A recursive tree scan at `7f04ae1…` found no `package.json`, Node lockfile,
  JavaScript, or TypeScript source.
- The newly included currency, Telegram, and live Grafana feature repositories likewise contain no
  Node manifest, lock, JavaScript, or TypeScript source at their pinned
  revisions. Telegram and Grafana add Python ecosystem surfaces, not a shared
  Node/Python implementation or package graph.
- The Curator tree at `9ba552f…` likewise contains no Node or Python build
  driver. Its accepted drivers are Go-only.

Therefore Node/TypeScript should mirror shared concepts that have independent
value—canonical manifest parsing, complete recursive closure, offline execution,
immutable source identities, binary rejection, lifecycle-script controls,
receipts, diagnostics, and conformance vectors. It should not import the alternative implementation
modules, reuse a Python virtualenv, assume `pip` semantics, or be described as a
Python implementation path. Python can serve as a behavioral cross-check where
the shared protocol already defines equivalent outcomes.

The Python behavior relevant to Node/TypeScript divides cleanly:

- **Share or mirror through the protocol/conformance layer:** canonical and
  legacy manifest-name handling, schemas 1–7, strict unknown-field rejection,
  immutable source identity, recursive dependency closure, fixed manager-owned
  build vectors, network-off build execution, artifact non-execution during
  validation, protected-cache receipts/currentness, stable diagnostics, and
  fail-closed behavior. The relevant Python entry points are
  [`skillspec.py`](the alternative-implementation repository (README link)),
  [`builds/go_v1.py`](the alternative-implementation repository (README link)), and
  [`build_repository_pipeline.py`](the alternative-implementation repository (README link)).
- **Do not share as Node implementation contract:** Python interpreter/standard
  library identity, hidden-worker launch mechanics, virtualenv or `pip` state,
  the alternative implementation filesystem paths, Python bytecode-cache handling, and
  Python-specific native-runtime proofs. Node must define equivalent security
  properties in its own ecosystem terms.

## Deferred Kotlin boundary (inventory only)

No Kotlin closure design was performed. The minimum evidence retained for
future backlog is:

- [`toolkit/extract-screenshots/build.gradle.kts`](https://github.com/relux-works/skill-android-testing-tools/blob/b79449cc3e1767680b069ae314c850a1d93c6f99/toolkit/extract-screenshots/build.gradle.kts)
  declares Kotlin/JVM, Java 21, `jvmToolchain(21)`, Clikt, and a fat-JAR task.
- [`gradle-wrapper.properties`](https://github.com/relux-works/skill-android-testing-tools/blob/b79449cc3e1767680b069ae314c850a1d93c6f99/toolkit/gradle/wrapper/gradle-wrapper.properties)
  requests Gradle 8.13 from the network and contains no distribution SHA-256.
- No dependency lock or Gradle verification metadata was found.
- `toolkit/gradle/wrapper/gradle-wrapper.jar` is a tracked 45,633-byte JVM
  bytecode payload, blob `f8e1ee3…`, and is default-deny under the delivery
  input.

This evidence explains why the existing Kotlin surface cannot be silently
folded into a current adapter. It does not authorize further Gradle/Maven work.

## Findings and anomalies

### 1. A canonical declared Go build root already contains a forbidden binary

`skill-project-management@8dc0b71` tracks
`tools/board-tui/task-board` inside the `tools/board-tui` build root declared by
`agent-skill.json`. Git metadata reports mode `100755`, size 14,501,282, and blob
`3cabcb35efb5ba6079cb33ae4754abd52bf5fee9`. Local materialization identifies a
Mach-O arm64 executable with SHA-256
`41d540604ba205b39622d0238d837e25ae5af34cb3df13eb88faddcb18daba67`.
Embedded build metadata points at a different VCS revision and dependency set
than the cited source. This is both a direct policy violation and proof that
artifact detection cannot depend on filename extensions.

### 2. Swift has a real mixed graph, but its authoritative repository has no lock

The iOS testing `Package.swift` contains only version ranges. The installed
machine happened to resolve five repositories to exact revisions, and the
checkouts expose the C targets listed in the matrix. Because `Package.resolved`
is not tracked at `bd59ca…`, another machine may select a different transitive
graph. Adapter design must capture resolution into the admitted closure before
offline build, and must bind every package revision/content snapshot rather
than trusting the generated `.build/checkouts` directory.

### 3. Most real CLI surfaces are not machine-declared

Only the project-management skill contributed an `agent-skill.json` in the
inspected active roots. iOS App Manager, Product Appraisal, Agents Infra, iOS
Testing Tools, Android telemetry, currency exchange, and Telegram are invoked
by skill prose/setup scripts, not by Curator build declarations. The live
Grafana feature checkout does have a the alternative implementation manifest, but that declaration
belongs to a different revision and command set than the installed tag. These
cases create an inventory/ownership gap before language work begins: an adapter
cannot protect a command Curator never owns, and an inventory cannot bind a
command to metadata taken from another repository state.

### 4. Existing Python script installation resolves undeclared executable input

The legacy fleet's marker hashes prove which script snapshot was installed, but
they do not bind what `pip` later places in `.venv`. The observed YouTrack
runtime includes `cryptography/hazmat/bindings/_rust.abi3.so`, cffi,
charset-normalizer, mypyc and pydantic-core `.so` files, Selenium's
`selenium-manager.exe`, and pip's bundled Windows launchers. This is expected
for ordinary Python packaging and incompatible with the task's default-deny
source-closure profile. Exact top-level requirements alone do not solve it.

### 5. Several Go CLIs are source-buildable but not offline-closed

- iOS App Manager and Agents Infra have `go.sum` but no vendor tree.
- Appraise has no vendor tree and a relative `replace` outside its repository
  root.
- Android telemetry is dependency-free but invokes `go run` on every use and
  has no manifest receipt.

These are migration inputs, not evidence for relaxing `go-v1`.

### 6. Generated local products must not be mistaken for repository payloads

The installed iOS Testing Tools copy and currency source checkout contain
`.build` Mach-O executables, and the installed iOS App Manager copy contains
locally built Mach-O outputs. Telegram's source checkout contains an ignored
venv with bytecode and pip's Windows launchers. The cited Git trees do not track
those files. Admission must scan the actual frozen source snapshot and reject
generated/precompiled payloads wherever found, while the research record must
distinguish a repository defect from dirty/generated installation state.

### 7. Currency exchange is the best real mixed Go/Swift case—and currently fails closed

The source has an explicit Swift-first/Go-second install order and a runtime
Go-to-Swift subprocess edge, so it exercises a cross-language graph more
directly than a synthetic fixture. It is not portable or offline-closed:
`go list -m all` exits 1 on two absolute replacements, the 30 checksum-bound Go
module/version pairs do not bind those replacement sources, and the Swift
version range has no tracked resolver result. This surface should become the
primary real integration vector only after both source repositories and the
Swift resolution are captured immutably.

### 8. Telegram proves both transitive dependency drift and installed artifact drift

The exact `telethon==1.42.0` requirement resolves today to pyasn1 0.6.4 while
the live venv contains 0.6.3. The registered the alternative implementation copy additionally
omits the requirement file its bootstrap reads and contains an undeclared
CPython 3.14 bytecode payload; `csk status --all` reports `content-drift`.
This is not evidence for launching a Python adapter in this cycle. It is a
high-value conformance input for dependency locking, source-distribution build
isolation, installed-subset completeness, bytecode rejection, and drift
diagnostics that Node/TypeScript should mirror only at the policy level.

### 9. Manifest discovery must preserve revision state, not just repository name

The eight live `csk-skill.json` files are not interchangeable with eight
installed markers. The deterministic union found 15 distinct commits across
working-copy, configured, and installed authority. Grafana demonstrates the
material consequence: clean feature `HEAD` `6234e2a…` exports three commands,
whereas configured/installed `a557671…` exports one differently named command.
Band demonstrates a second state: configured `f2154e9…` is not yet installed
over `e49922a…`. A closure scanner must therefore emit repository + immutable
revision + state kind + command identity, and must fail or record an unresolved
state when command-bearing working files are dirty. Repository-name
deduplication would lose real surfaces.

## Recommendations for the adapter-design handoff

1. **Fix or fail the project-management seed first.** Remove the tracked
   `tools/board-tui/task-board` binary from the skill repository, add a
   conformance case for extensionless executable detection, and require the
   common compiled-artifact diagnostic before any cache lookup or build.
2. **Create explicit command manifests before migrating repo-facing CLIs.** Add
   task-scoped manifest work for iOS Testing Tools, iOS App Manager, Appraise,
   Agents Infra, Android telemetry, currency exchange, and Telegram. A Telegram
   declaration is inventory/ownership work, not approval for a new Python
   adapter. Keep system commands separate.
3. **Make estate discovery revision-state aware.** Emit a separate record for
   each working-copy `HEAD`, configured target commit, and installed marker
   commit, then deduplicate only identical revisions. Treat dirty
   command-bearing files as unresolved evidence. Add Grafana's `6234e2a…`
   versus `a557671…` command split and Band's configured-versus-installed split
   as conformance vectors for the inventory layer.
4. **Use currency exchange as the primary real mixed Go/Swift integration
   vector.** Replace both absolute Go paths with immutable admitted sources,
   capture the Swift resolver result and plugin declarations, express
   Swift-before-Go build order plus Go-to-Swift runtime dependency, and require
   an offline rebuild with `.build`/binary rejection. Do not admit only one half
   of the two-command contract.
5. **Use Android telemetry as the smallest additional Go migration.** Its
   standard-library-only module should become a `go-v1` command and stop using
   runtime `go run`.
6. **Resolve the Appraise repository edge explicitly.** Prefer vendoring into
   the declared build root if license/provenance permits; otherwise use the
   already-specified locked external-build-repository model. Never allow the
   relative `replace` to escape the frozen root.
7. **Keep the other Swift shapes as focused conformance vectors.** Use the
   dependency-free Android `snapshotsdiff` package for a minimal executable and
   iOS Testing Tools for transitive Swift+C coverage. Require exact Git/content
   identities, offline reconstruction, plugin/build-tool rejection,
   platform/toolchain checkpoints, and `.build` exclusion.
8. **Create new Rust and Node/TypeScript reference fixtures.** No current CLI can
   serve as the first migration. Fixtures must exercise direct and transitive
   source dependencies, locks/integrity, offline builds, scripts/plugins,
   generated sources, mixed/native edges, and binary rejection without using
   the absence of estate code as permission to weaken policy.
9. **Use Python only as an independent behavioral/ecosystem reference in this
   cycle.** Compare stable protocol outcomes and diagnostics; add Telegram
   vectors for missing bootstrap inputs, transitive drift, `.pyc`, and venv
   payload rejection, plus Grafana vectors for ranged online resolution,
   platform-conditioned transitives, and ignored working-copy bytecode. Do not
   make a new Python implementation a prerequisite for Node/TypeScript.
   Preserve source-only Python/wheel research as its own policy work.
10. **Keep Kotlin, Dart, and .NET outside the active graph.** The Kotlin wrapper
   JAR and no-lock evidence belong in future backlog; they must not block current
   Rust, Node/TypeScript, or SwiftPM/C-family work.

## Fact-check record

The following direct checks produced the claims above:

- `git ls-remote <repository> refs/heads/main` for every public repository in
  the revision ledger: exit code 0 for each resolved head.
- GitHub recursive tree queries at every cited revision for package manifests,
  locks, Rust/Node sources, and denied artifact extensions: exit code 0. No
  Node source/package metadata or Cargo package matched, and no Rust source
  matched in those top-level cited trees; the Android Gradle wrapper JAR did
  match. A separate installed-checkout scan found the non-product
  swift-collections benchmark helper, and a named-output/mode query found the
  project-management Mach-O candidate.
- `file tools/board-tui/task-board`: exit code 0, Mach-O 64-bit arm64.
- `go version -m tools/board-tui/task-board`: exit code 0, reporting Go 1.25.5,
  VCS revision `5e5f095…`, `vcs.modified=true`, and the embedded module graph.
- `shasum -a 256 tools/board-tui/task-board`: exit code 0, hash recorded above.
- `csk --version`, `csk list --paths`, `csk global list`, `csk global status`,
  and `csk status --all`: exit code 0 for each read-only estate query.
- the alternative implementation marker/manifest and runtime-root reads, SwiftPM generated-lock and
  checkout-revision reads, package manifest reads, and compiled-payload `find`
  + `file` inspections: exit code 0.
- `go list -m all` in the Curator worktree: exit code 0; vendor metadata counts
  for both project-management build roots: exit code 0.

First-review rework checks were run directly as standalone processes unless a
single shell loop is explicitly described:

- `csk --version`, `csk list --paths`, `csk global status`, and both
  `csk status --all` invocations: exit code 0. The hidden-file active-root
  marker projection and manifest projection: exit code 0.
- `git ls-remote` for currency tag `v2.1.1` and Telegram `main`: exit code 0,
  returning `c29210aa…` and `b9a76b01…` respectively. Detached checkouts and
  commit reads at those exact revisions: exit code 0.
- Recursive `git ls-tree` reads and per-tracked-file compiled-signature loops
  for both pinned repositories: exit code 0 and zero compiled/denied hits.
- `go mod edit -json` and the `go.sum` shape count for currency: exit code 0,
  recording 24 requirements, 53 checksum lines, and 30 unique module/version
  pairs. `go list -m all`: **exit code 1, expected failure** because the
  absolute `agentquery` replacement path is absent; this proves the graph is
  not portable/closed.
- `git ls-files --error-unmatch exchange-scraper/Package.resolved`:
  **exit code 1, expected failure** because no Swift lock is tracked.
  `swift package describe --type json`: exit code 0, confirming tools 6.0,
  macOS 13, one executable/library/test target set, and the
  `swift-argument-parser` range.
- `swift test` at currency revision `c29210a…`: exit code 0; 13 Swift tests in
  three suites passed after resolving generated local state to
  `swift-argument-parser` 1.8.2. This green test is runtime evidence only; the
  generated resolution does not repair the missing tracked lock.
- `go test ./...` at currency revision `c29210a…`: **exit code 1, expected
  failure**. Package setup fails on the two missing absolute replacement
  directories; `internal/domain` alone passed. The result is intentionally not
  reported as a green repository gate.
- `python3 -m unittest discover -s tests -v` at Telegram revision `b9a76b0…`:
  exit code 0; three installer tests passed.
- A system-interpreter pip dry-run: **exit code 1, recoverable PEP 668
  refusal**. Creating an isolated temporary venv: exit code 0; repeating the
  dry-run there: exit code 0 and the four-package resolution recorded in the
  matrix.
- `git grep -- '--hash=' requirements.txt`: **exit code 1, expected no-match**
  proving there are no requirement hashes. `test -e` for the registered
  copy's `requirements.txt` and `Makefile`: **exit code 1 for each, expected**
  because both bootstrap/install inputs are absent. The pinned source's
  `git ls-files --error-unmatch requirements.txt`: exit code 0.
- `file`, `stat`, and `shasum -a 256` on the registered Telegram `.pyc`:
  exit code 0. Live-vendor package listing, `.pyc`/`.exe` counts, and
  `tg-telethon --help`: exit code 0.
- `command -v exchange` and `command -v exchange-scraper`: **exit code 1 for
  each, expected absent current binaries**. All three Telegram shim lookups:
  exit code 0. External `glab --version` and `sentry-cli --version`: exit code
  0.
- Node/Rust manifest/source scans in both newly included repositories:
  **exit code 1 for each, expected no-match**. They add neither a Node/Python
  shared implementation nor a Rust surface.
- `go version`, `swift --version`, `sw_vers`, `python3 --version`, and
  `command -v security`: exit code 0, producing the observed runtime/toolchain
  identities recorded above.
- The task-scoped research acceptance assertion (required sections, exact Git
  objects/paths, installed-marker/drift evidence, stale-exclusion absence, and
  Markdown table-column consistency): exit code 0.
- `task-board validate`: exit code 0, `Board is valid. No issues found.`

Second-review rework checks for deterministic revision selection and the live
Grafana surface were likewise run directly:

- Current `csk --version`, `csk list --paths`, `csk global list`, `csk global status`, and
  `csk status --all`: exit code 0 each. They reproduced four registered
  projects, four global surfaces, Band `update-available`, installed Grafana
  `a557671…` `up-to-date`, and Telegram `content-drift`.
- Hidden-file scans across the the alternative implementation root, global root, Curator root, and
  all four registered project roots: exit code 0 each, listing exactly 16
  installed markers, eight physical `csk-skill.json` files, and one
  `agent-skill.json`.
- Git top-level, `HEAD`, branch, and status reads for all eight physical
  the alternative implementation manifest repositories: exit code 0 each; all eight trees were
  clean. Configured-ref resolution and manifest extraction at every immutable
  state: exit code 0. The GitLab checkout's `git config --get
  remote.origin.url` returned **exit code 1, expected absent configuration**;
  its local Git objects still supplied the cited `HEAD` and configured/project
  and global tag commits.
- Project/global GitLab and YouTrack marker reads: exit code 0, reproducing
  global `bdce7f7…` / `58eae4f…` in addition to their project variants. The
  state union therefore contains 15 distinct commits, not eight manifest files.
- Pinned Grafana `git ls-tree` and `git archive`: exit code 0, yielding 41
  tracked files. The per-file `file` loop exited 0 and classified every member
  as source/text. The denied-extension query returned **exit code 1, expected
  no-match**; no tracked precompiled payload was found.
- Grafana lockfile query, requirement `--hash=` query, and Node/TypeScript/Rust
  manifest/source query: **exit code 1 each, expected no-match**. The feature
  revision has no lock/hash metadata and adds no Node/Python shared code or
  Rust surface.
- `python3.13 -m venv` for an isolated temporary environment: exit code 0.
  Its pip `--dry-run --ignore-installed --report ... -r
  scripts/requirements.txt`: exit code 0 and the five-package, five-wheel
  macOS resolution recorded in the matrix. Reading the generated report with
  `jq`: exit code 0.
- The initial live Grafana generated-payload scan: exit code 0, finding 17
  ignored `.pyc` files totaling 204,588 bytes. After the explicit unit-test
  gate, the repeated scan exited 0 and found 31 `.pyc` files totaling 373,177
  bytes with zero other denied-extension files. These are working-copy
  observations, not members of commit `6234e2a…`.
- `python3.13 -m unittest -q tests.test_setup_support tests.test_setup_main
  tests.test_secure_auth tests.test_grafana_auth tests.test_analyze_runtime
  tests.test_grafana_query` at Grafana `6234e2a…`: exit code 0; 26 tests passed.
- The exact live/configured/project-marker/global-marker state-union assertion
  resolved every Git ref and marker commit, deduplicated with `sort -u`, and
  exited 0 with `distinct_manifest_state_commits=15`.
- The revised task-scoped research acceptance assertion checked required
  sections, all authoritative Grafana/global/project revisions and command
  paths, dependency/runtime/artifact evidence, Node/Python and Kotlin
  boundaries, stale installed-only wording, and Markdown table-column
  consistency: exit code 0, `research_acceptance=pass`.
- Current `task-board validate`: exit code 0, `Board is valid. No issues
  found.`

No build or product code was changed by this research task. Repository-specific
expected-red closure gates are findings rather than deliverable failures. The
green research-artifact checks above validate this task's deliverable; they do
not misrepresent the expected-red currency Go closure as a passing source-repo
gate.
