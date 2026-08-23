# TASK-260810-1veyfw reviewer verdict

Run: `RUN-260811-e84e63`

Goal: `GOAL-260811-d91bb4` revision 1

Resolved scope: `TASK-260810-1veyfw`

Verdict: **CHANGES REQUESTED**

Route: `analysis`

## Acceptance assessment

The reworked outcome fixes both omissions from the prior review and is strong
for the surfaces it now records. It still does not satisfy the claimed
closure-complete estate inclusion rule: one of the eight live
`csk-skill.json` manifests counted by the report declares three current Python
CLI surfaces that are absent from the revision ledger, matrix, exact inventory,
and closure ledger.

### Passing evidence

- The board outcome and workspace research file both have SHA-256
  `55910da5a8b08dd7acf2d61799e89b00666cfb2e9b243a527bad89cff626d7ae`.
  There is exactly one task-scoped inventory outcome resource, so the producer
  correctly updated the existing artifact rather than creating a duplicate.
- The deterministic inclusion rule now covers Curator declarations,
  repo-facing CLIs invoked by instructions or launchers, CocoaSkills
  project/global surfaces, and external system commands as a distinct class.
- `skill-currency-exchange@c29210aa6eb4cc0f64f307fa30561ac80feb6b3b`
  is fully represented. Independent Git-object checks reproduce Go 1.25.5,
  four direct plus twenty indirect requirements, two absolute replacement
  paths, 53 `go.sum` lines / 30 unique module-version pairs, Swift tools 6.0,
  macOS 13, the `swift-argument-parser` range, no tracked
  `Package.resolved`, Swift-first/Go-second build order, and the Go-to-Swift
  subprocess boundary. The installed marker binds the same commit and exactly
  three shipped files with no declared commands.
- `telegram-telethon@b9a76b01e7ce211c1d0e707f97b231ee7b817d41`
  is fully represented. Independent checks reproduce the `telethon==1.42.0`
  input, venv/pip bootstrap, three shims, absent installed `requirements.txt`
  and `Makefile`, `content-drift`, and the undeclared 63,929-byte
  `telegram_telethon.cpython-314.pyc` with SHA-256
  `dd5ddeee6390a846638ec26ab3ca534367f5c9a61aa45468158e34f8872a5167`.
  A clean isolated pip dry-run resolves pyasn1 0.6.4 while the live venv still
  contains 0.6.3.
- Go is correctly classified as the implemented Curator baseline. Pinned
  Curator and protocol Git objects reproduce `go-v1`,
  `go-repository-v1`, schemas 1-7, fixed vendored build vectors, and the
  fail-closed mixed-source fields.
- The Node/TypeScript versus Python relationship is explicitly and correctly
  resolved as protocol-level rather than implementation-level. Pinned
  CocoaSkills metadata exposes only the Python `csk` entry point, has zero
  runtime dependencies, and contains no Node manifest, lock, JavaScript, or
  TypeScript source.
- Kotlin remains deferred. The pinned Android repository evidence reproduces
  Java 21 / `jvmToolchain(21)`, Gradle 8.13 without a distribution checksum,
  fat-JAR assembly, and the tracked 45,633-byte wrapper JAR.
- Verification reproduced the reported test disposition: `task-board
  validate` is green; Telegram has 3 passing tests; the currency Swift package
  has 13 passing tests; currency `go test ./...` fails at setup on the two
  missing absolute replacement directories and is correctly reported as an
  expected-red closure finding rather than a green gate.

## Finding: the active-manifest inclusion rule is not closed

The report says to include every command declared by a CocoaSkills
`csk-skill.json` and states that its hidden-file scan covered
`/Users/iv/agents/skills`, finding eight such manifests. It then says the
post-review closure ledger accounts for every resulting command-bearing or
negative disposition.

That evidence does not agree with the current scanned source:

- `csk list --paths` identifies `/Users/iv/agents/skills` as the CocoaSkills
  skills root.
- `/Users/iv/agents/skills/skill-grafana` is a clean Git checkout at
  `6234e2a3ef35415b765f74cca21c8ad7846f521e` on
  `feature/oparin/PMA-23845` (remote
  `git@gitlab.wildberries.ru:portals/agentic-infra/skills/skill-grafana.git`).
- Its authoritative checked-in `csk-skill.json` declares three script
  commands: `analyze -> scripts/grafana-analyze`,
  `query -> scripts/grafana-query`, and
  `grafana-auth -> scripts/grafana-auth`.
- Those launchers invoke Python through `run_in_skill_venv.py`; the same tree
  contains `scripts/requirements.txt` and the complete bootstrap/runtime
  sources. This is a real current implementation surface under the report's
  stated rule, not documentation-only metadata.
- The registered project marker is a different state: tag `v1.1.0` at
  `a557671bce2ef7c6547293dc4b28fcc8636df51a` declares only `grafana`, and
  `csk status --all` reports that installed tag as up-to-date.
- The outcome records only `grafana@a557671...`. Searches of the 526-line
  artifact find no `grafana-analyze`, `grafana-query`, or `grafana-auth`.

The two revisions may legitimately be classified as an active source checkout
and an installed project surface, or the feature checkout may legitimately be
excluded in favor of configured refs. What is not reproducible is counting the
live manifest as closure evidence while silently using only the installed
marker's command set. The artifact also includes source-repository CLIs that
are not installed or on `PATH`, so installed-state-only reasoning cannot be
applied to Grafana without an explicit, consistent boundary.

## Required corrections

1. Define which revision is authoritative for every manifest discovered under
   the CocoaSkills skills root: working-copy HEAD, configured ref, installed
   marker, or an explicit combination. Make the scan and its counts operate on
   that revision rule rather than mixing working-copy file discovery with
   installed-marker command extraction.
2. If working-copy manifests are included, add
   `skill-grafana@6234e2a3ef35415b765f74cca21c8ad7846f521e` to the revision
   ledger and matrix. Record `analyze`, `query`, and `grafana-auth`, their
   launcher/bootstrap paths, Python/pip dependency input, integrity and
   transitive shape, runtime requirements, mixed-language edges, and compiled
   payload disposition; distinguish them from installed `grafana@a557671...`.
3. If the working-copy feature revision is excluded, state the deterministic
   configured-ref/installed-marker rule, stop presenting the eight physical
   manifest files as command-closure evidence, and re-run the same rule over
   every other repo-facing source so that uninstalled surfaces are still
   included or excluded consistently.
4. Re-run the closure ledger and task-scoped acceptance assertion, update the
   existing inventory outcome resource, and return the task for another
   reviewer cycle.

## Independent checks performed

- Re-ran `csk --version`, `csk list --paths`, `csk global status`, and
  `csk status --all`; all exited 0 and reproduced four registered projects,
  four global surfaces, and Telegram content drift.
- Re-ran the hidden manifest/marker scan over the stated active roots; it
  reproduced 16 installed markers, eight `csk-skill.json` files, and one
  `agent-skill.json`.
- Read pinned local Git objects for Curator, Curator Protocol, CocoaSkills,
  project management, currency exchange, and Telegram. Read the remaining
  cited repository revisions through isolated temporary bare fetches, without
  modifying their working trees.
- Verified every GitHub path citation extracted from the outcome against its
  pinned repository tree; the cited package-local Makefile paths are correct.
- Reproduced the installed SwiftPM resolver pins and C-target edges, the
  project-management extensionless Mach-O identity, the currency/Telegram
  dependency and artifact observations, and the current host toolchain
  identities.
- Re-ran the green/expected-red tests listed above and `task-board validate`.

No product code or producer artifact was modified. This is ordinary research
rework, not a stop-the-line boundary.
