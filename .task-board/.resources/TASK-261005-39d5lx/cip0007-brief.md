# THE ONLY CURRENT INSTRUCTION — TASK-261005-39d5lx: design CIP-0007 for manager-provisioned CLI tools (researcher; curator-spec)
The operator asked for a design of curator-spec issue #108 once the current work is done (2026-10-04). Triage card I2 recommended option B: a registry-only lane, which needs an owner and service decision.
Read first:
- `gh issue view 108 -R relux-works/curator-spec` (the full proposal and comments);
- `cips/README.md`, `cips/TEMPLATE.md` and `cips/CIP-0001-*.md` (the process);
- the related Draft CIPs 0002–0006 for tone and depth;
- the spec sections on system commands (`protocol/core.md` §4), the skillfile schema, the lock and the install transaction.
Produce:
1. `cips/CIP-0007-<slug>.md`, Status **Draft**, in template shape. Cover:
   - problem; goals and non-goals;
   - options: (A) status quo system commands with better hints, (B) a registry-only lane (logical identity plus version constraint, resolved only from operator-trusted registries, no author-supplied URLs), (C) full provisioning with layered signed registries, trust policy and an upstream registry mirror;
   - trust model and threat analysis (author-controlled downloads, registry compromise, signature and key rotation, offline and air-gapped behaviour);
   - interaction with the skillfile schema, lock, install transaction, conformance and the CIP-0005 audit backends;
   - migration and compatibility; open questions; a **decisions needed** list for the operator (owner and service of the registry, signing keys, default policy).
2. An evidence note in `.research/` that links to it, with citations by file and § or line.
3. One index row in `cips/README.md`.
Rules:
- No normative protocol or schema edits. No CHANGELOG or LOGBOOK edits.
- Do NOT name any organisation, team, client or person (the issue text names one; never repeat it).
- No personal paths or host names: the repo is public.
- Run `python -B tools/validate.py` and record its exit code.
Then `task-board handoff TASK-261005-39d5lx --role researcher` and END YOUR TURN.
