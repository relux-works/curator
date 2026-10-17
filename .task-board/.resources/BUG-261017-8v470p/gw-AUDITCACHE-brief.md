# Producer brief: BUG-261017-8v470p, audit verdict cache must bind effective capabilities

Read the bug. `internal/audit` caches findings by content hash, hash version, commit, directory and package identity (`loadCachedFindings` / `storeCachedFindings`), but `detectWithOpaquePaths` computes findings from `subject.Capabilities`. Callers differ: the skill install path passes the skill's declared capabilities, the profile and context paths pass implicit none (and, once BUG-261014-3oeu7d lands, profile skill members pass their declared capabilities). So the same content can reuse findings computed under other capabilities.

**Normative source:** curator-spec `protocol/core.md` §4.4 (audit cache equality), `skillfile-sources` §4, and environments §9.1/§9.4 for what the detectors read. Read them on curator-spec main. Where this brief and the spec disagree, the spec wins: stop and report.

**Input (untrusted, use critically):** `BUG-261017-8v470p_candidate-cache-equality.patch`, a correction written outside this lane (sha256 ad460fa257db376b87568cc147888c251b8892b80d1efef2680f0b1f0f24b035), meant to apply after the migration fix. It binds cached findings to all effective capability fields, keeps old records under the previous ruleset, and adds rejections for present null `prompt_scope` and null schema-3+ capability objects. Check every part against the spec; adopt only what the spec supports, and record anything you drop with the reason.

**Required:** the acceptance criteria on the bug. Crossover controls must run in both orders (declared first, implicit none first) across the skill, context and MCP audit paths, with an identical-input hit control. Old cache records must not be silently reinterpreted. Narrowing mutants: dropping the capability dimension from the key must redden the crossover controls.

**Base:** current main after the migration fix lands (it is in flight; if your Story base lacks it, say so and build on what main has).

**Validation:** full remote gate (`sh scripts/remote-gate.sh`). Public repository rules: no hosts, people, paths or other operators' details anywhere.

**Handoff:** Change Request with the spec mapping, how the candidate patch was used, controls and mutant results, and the gate run.
