# TASK-261016-2rpqii: developer handoff

Ready for review. Documentation-only candidate; no product implementation changed.
The Story worktree and index remain uncommitted. No branch was switched or rebased.

## Delivered scope

- `docs/agent-logins.md`: versioned Claude Code 2.1.299, Codex CLI 0.153.2 and Muse 1.4.8 observations; macOS/Linux evidence bounds; current isolation table link; machine-owned credential-store selector; current, rc.7 and planned routes; platform/spawn recipes; six normative specification links.
- Four focused PlantUML sources and four SVG outputs beside them, embedded in the reference.
- A visible README pointer outside the collapsed command section, exact selector command and native default, plus rendering/link-check commands in Tools.
- Environment configuration cross-link and correction of the contradictory claim that Claude's ssh file login was visible to every session.

## Identity and scope

Fetched `origin/main`, HEAD and local main all equal `205272e75ea7d40fd10e073c20ce49bbc0787cd2`; HEAD..main is zero commits.
Read public normative spec main `55094992185521dba470d80e1745d60af7bf496c`.
Built the repository CLI itself to inspect main, rather than trusting the separately installed binary.
Built CLI help has no env login or profile prepare; source dispatch/config grammar confirms the credential_store selector is absent.

Environment: macOS arm64, Go 1.25.5, PlantUML 1.2026.6, Graphviz 14.0.4, Lychee 0.24.2.
Tool versions above identify checks performed here. Historical vendor measurements were accepted from the source pack, not rerun against real logins. No login, logout, token read or credential-store mutation was performed.

## Validation commands and real exits

Each command ran directly; no gate was piped through tee. Local raw logs remain under `.temp/TASK-261016-2rpqii/` and are not public attachments because diagnostic paths identify the machine.

| Command | Exit | Evidence / bound |
|---|---:|---|
| `go build -o .temp/TASK-261016-2rpqii/curator ./cmd/curator` | 0 | `build-01.log`; built main for CLI verification. Go source/config/test identity remained unchanged throughout this docs delta. |
| `.temp/TASK-261016-2rpqii/curator env --help` | 0 | `main-env-help.log`; resolve/status/config/migrate/unmanage only. |
| `.temp/TASK-261016-2rpqii/curator profile --help` | 0 | `main-profile-help.log`; prepare absent. |
| `plantuml -tsvg -nometadata diagrams/plantuml/agent-login-*.puml` | 0 | `plantuml-render-03.log`; all 4/4 final sources rendered. SVG XML parse also passed for 4/4; metadata omitted. |
| `plantuml -tpng -o ../../.temp/TASK-261016-2rpqii/preview diagrams/plantuml/agent-login-*.puml` | 0 | Preview render for visual inspection; stores and prepare previews inspected. Two source labels were subsequently clarified and all final SVGs rerendered successfully. |
| `lychee --offline --include-fragments --no-progress README.md docs/environment-config.md docs/agent-logins.md` | 0 | `links-local-03.log`: 31 successful checks, 0 errors, 17 external exclusions; external links checked separately below. |
| `lychee --no-progress --timeout 30 docs/agent-logins.md` | 0 | `links-external-03.log`: 12/12 unique targets successful, 0 errors, 1 followed redirect. Covers every newly added external URL and the embedded assets. |
| `go test -count=1 -timeout=3m ./internal/envprofile -run '^Test(CodexIsolatedAdmission|CodexUnknownStoreSharedRefuses|CodexCredentialStoreTOMLSpellings|CodexMalformedStoreSharedRefuses|MuseDirectoriesAndIsolation|Reviewer.*)$'` | 0 | `tests-envprofile-01.log`; existing admission/refusal properties plus all four Reviewer-prefixed tests in this package. |
| `go test -count=1 -timeout=3m ./cmd/curator -run '^TestEnvConfig(ShowDefaults|SetUnsetRoundTrip|Refusals)$'` | 0 | `tests-cli-01.log`; existing CLI defaults, valid configuration roundtrip and rejection coverage. |
| `bash .github/ci/naming-gate.sh .temp/TASK-261016-2rpqii/public-scope` | 0 | `naming-gate-02.log`; exact copies of the 11 candidate files. Its employer-pattern bound is narrower than the brief's hygiene requirements; manual review checked the added prose for identifying data and local machine paths. |
| `git diff --check` | 0 | `diff-check-03.log`; no whitespace errors. New Markdown/PlantUML sources were also inspected; the named Git command does not scan untracked files. |
| `lychee --offline --no-progress .temp/TASK-261016-2rpqii/link-control.md` (positive fixture) | 0 | `links-control-positive-01.log`: 4/4 diagram links successful. |
| `lychee --offline --no-progress .temp/TASK-261016-2rpqii/link-control.md` (one retargeted link) | 2 | Expected red, `links-control-negative-01.log`: exactly 1 missing diagram, 3 accessible controls. |

Initial render exited 200 for a Salt syntax error; corrected to startsalt/endsalt and rerendered green. Initial local and online link commands each exited 2 because they ran before generated SVGs existed. Those are failed attempts, not passing evidence. Final checks ran after successful rendering. No missing or truncated log is claimed as green.

## Narrowing-control evidence

These are documentation instrument controls, not product mutation coverage.

| Mutant / control | What it narrows the gate to | Named failing check | Surviving bound |
|---|---|---|---|
| `LinkCheckMissingDiagram`: retarget only stores.svg to stores.missing.svg in the isolated four-link fixture | One missing embedded diagram while three neighboring targets remain valid | `lychee --offline --no-progress .temp/TASK-261016-2rpqii/link-control.md`, exit 2; exactly the retargeted link fails | Killed. This does not test the correctness of diagram content, external URL content or all possible Markdown syntax. |

No product gate changed; no product mutants are claimed. The existing positive/negative property tests were reused instead of adding tests that merely mirror documentation text.

## Handoff preconditions and limits

- Brief gap: no attack-surface table was supplied. No catalog-based coverage-map rows or product security coverage are claimed; this is an operator documentation leaf.
- The affected existing property suites listed above are green. All committed tests are unchanged. The four existing envprofile reviewer tests retain their original names and were run green: TestReviewerDanglingExpectedTarget, TestReviewerLiteralCodexSelector, TestReviewerRecoveryMarkerDrift, TestReviewerRecoveryPreservesRegularTemp.
- Full hosted landing gate `sh scripts/remote-gate.sh` was NOT run. That configured script creates/publishes an unsigned snapshot with a synthetic author, conflicting with the signed-human-author policy. Hosted matrix/race evidence is unverified here and must be obtained through a compliant integration path before landing; this handoff makes no full landing-gate or all-repository reviewer-test claim.
- No out-of-contract brief rows were silently excluded. Linux Muse measurement is explicitly unknown, rather than represented as measured success. Acceptance criterion 1 requires coverage of the tool, not invented platform measurements. Current behavior, rc.7 contracts and planned source delivery are distinguished throughout.
- This is the first implementation handoff, not a repeat-of rework. The delta is bounded to docs, README and the requested diagram directory. No implementation, tests, tool configuration or private source pack was added.

## Task-scoped logbook

- FINDING: accepted CIP-0017 explicitly supersedes CIP-0015's generic native Muse mapping. Managed Muse native/unset becomes effective file in rc.7; stored default stays native. The public reference records the exception and native file-login prerequisite for sharing.
- FINDING: current main still lacks env login, profile prepare and credential_store. Those capabilities are labeled rc.7; protected setup-token enrollment/delivery remains planned.
- FINDING: source material has no Linux Muse login measurement and no token-refresh write qualification. The document preserves those limits.
- CORRECTION: environment-config's ssh credential row said any session could use the file, contradicting the adjacent measured two-store isolation statement. It now says sessions using the same file fallback.
- ANOMALY: initial Salt rendering and premature link checks failed as reported above; final asset build and both link checks are green.
- INTEGRATION LIMIT: the configured remote snapshot script's signing/author behavior needs a compliant integration owner path. No remote CI result or policy exception was fabricated.

## Candidate file digests

| File | SHA-256 |
|---|---|
| `README.md` | `523974dc6ac771ed3afd1af4e7868a1c004ff470ed3a2431f65cd134fe181214` |
| `docs/environment-config.md` | `dc992319aaf518935f848b55d04e5e9a5d8a015f125ada0d752dc94753e6f975` |
| `docs/agent-logins.md` | `715d6721cc92b6adc6c8d3a5f798d4275cd28d6fcf98bf171777439718ac6358` |
| `diagrams/plantuml/agent-login-homes.puml` | `88f09cc1e94f78da6234085a48f7249b37193da932a8bb7ecd18b7d18e3ecd28` |
| `diagrams/plantuml/agent-login-homes.svg` | `d454bfdb44b37acd15d64e3c843675eb5411ec9d3b1f3aae6deaf3098ad03af7` |
| `diagrams/plantuml/agent-login-prepare.puml` | `fd08b7dcb6c19e1205537bcce98af44bd0c7d59b83839c8b8ea681b182fe8b7a` |
| `diagrams/plantuml/agent-login-prepare.svg` | `8f9bb1d2d455e5302c4fd1665838e696991504e3308ad79ba5f23581367b42d7` |
| `diagrams/plantuml/agent-login-resolution.puml` | `abd572aee619227d558d99c7a67f610bc3ce7e62f494eaa7692ca4ec3a09a934` |
| `diagrams/plantuml/agent-login-resolution.svg` | `a0f04f91ea2878544a2a7a21f18087b430d5ef80b6d4463cbf8dee6e0e8b67fe` |
| `diagrams/plantuml/agent-login-stores.puml` | `897d2c69738bd6a445de3ea087d92db7d7c27de5a58269432507f518cfc38244` |
| `diagrams/plantuml/agent-login-stores.svg` | `2b95b8958ee09d6b119bffb4247ce3b4cc6a8609538d70222f1499dc51727dde` |
