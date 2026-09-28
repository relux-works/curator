# TASK-260921-3qcjsy brief (orchestrator, binding) — adopt decisions 0017 and 0018

Repository: curator-spec (this control root). Story STORY-260921-3z0fgr under EPIC-260905.
Operator decision 2026-09-21: ADOPT `decisions/0017-environment-credential-modes.md` and
`decisions/0018-curator-run-permission-interface.md` (both currently "Status: proposed — not
adopted"). Adoption means: status → adopted with a dated adoption note; every option set and
open question resolved to a recorded choice; the normative text the decision amends updated in
the same revision, or a named follow-up leaf per repository where the amendment is
implementation-side.

## Rulings

R1 0017: adopt the three review-recommended options verbatim — (1) `environments.isolation.
<profile>.<env-id> = shared|isolated` stays the canonical v1 knob, described as credential-store
sharing, not a sandbox; `credential_mode`, rename and per-run flag deferred; (2) registry-owned
strategies per environment × GOOS exactly as listed (codex_cli keyring-preferred → auth.json
file-link; pi file-link with native root `~/.pi/agent`; claude_code Linux file-link, macOS
isolated-by-default with shared refused until the experiment passes; opencode ambient only;
`keychain-shared` and copy-at-provision rejected); (3) fix-first manager repairs then an explicit
inspect → plan → apply migration under the manager lock, never silent inside `resolve --repair`,
no secret copies. Open questions 1–7: resolve each to a recorded choice using the fail-closed /
least-privilege reading where the draft gives none (e.g. Q1 macOS claude_code shared store:
unsupported until the experiment is recorded; Q2 `isolated` = bounded store separation, not
account separation — say so; Q3 Pi root precedence per option 2; Q4 Codex keyring identity:
state what the manager assumes and the probe that verifies it; Q5 marker/config fields: name
them; Q6 fleet enforcement of `isolated`: not in this adoption — follow-up; Q7 consent shape for
future copy: none authorized — copy stays refused). List every choice in a table in the
document and in results.md for operator confirmation.
R2 0018: adopt with the 2026-09-16 amendment already in the document (permission mode
configurable in the launcher global config `curator-run-defaults` or the profile config, flag
overrides, DEFAULT yolo). Open questions 1–7: (1) `--yolo` ships in the same increment as
`--permissions` as an alias of the yolo mode; (2) tracked-yolo `ax` admission precondition: the
fail-closed reading the draft describes (state it as the normative rule); (3) unknown future
native policy forms: refused (fail closed) with a versioned-capability token; (4) launcher
prints the effective native policy line and records it in the launch record — name the record;
(5) negative tests: the real `curator run` entry with a fake tool — name the rows the
implementation must add; (6)/(7) versioned-capability and minimum-version token: adopt the
draft's proposal; where a value is still unknown, record "to be fixed by the implementing
revision" as an explicit follow-up, not as an open question.
R3 Amend the normative text the decisions target where the amendment is spec-side (environments
credential modes section(s), `curator run` interface rows in cli/launcher docs, conformance
vectors if a decision names them); keep the compatibility/security-impact sections truthful;
update the decisions index/README; `make validate` (the configured gate) green. Name the
implementation follow-ups per repository (curator: envprofile repairs + migration step + tests;
curator-agent-launcher: permission interface) in results.md with one-line ACs so the
orchestrator can create the leaves.
Publish the Change Request only when the gate is green.
