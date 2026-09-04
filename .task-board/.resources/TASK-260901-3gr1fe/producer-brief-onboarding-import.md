# Producer brief: onboarding import (path source kind) + §9.1 ref selection

## Setup
- Base: fresh `origin/main` of `~/Developer/ReluxWorks/curator-spec` (must contain protocol/environments.md, schemas, manager §12 — i.e. 62e592a or later). Worktree `~/Developer/ReluxWorks/.worktrees/curator-spec-onboarding-import`, branch `draft/environments-onboarding-import`. Signed commits. Do not push.

## Task — two scope items, one coherent delta
### 1. Onboarding import (Decision 0010 Decision 5 step 5 + step 2/3 classification), normative
Edit `protocol/environments.md`:
- §1: promote the `path` source kind from reserved/deferred to fully normative — operator-local profile directory by absolute or project-relative path; snapshot copied into the store; keyed by the §8 content hash of its state; never network; interaction with audit (always-strict applies; source identity shape for a path profile).
- New subsection under §9 (or where lifecycle lives): the import flow — inventory reassembly of detected native context into an imported-from-native profile directory inside the machine home; normative lossless/lossy classification (lossless iff every detected surface maps onto a supported surface of the detecting adapter's revision; define the detected-surface list per adapter: root-context file, skills entries; unreadable file or unsupported surface => lossy with a named loss list); consent gate rules (lossless proceeds, lossy stops with the list); skill migration (imported skills enter the profile Skillfile with the managed-by-other-means warning and re-declare recommendation); authentication never touched. Diagnostics with stable codes for classification results and import failures. Keep Decision 0010's D5 wording authoritative; concretize, don't contradict.
- Marker/schema impact: `agent-environment-marker-v1` currently admits git/local pin branches. Extend for `path` profiles (state-hash pin + source path record). Decide explicitly, with a recorded rationale in your notes, whether in-place evolution of schema 1 is admissible (surface is pre-release, never tagged) or a v2 is required per COMPATIBILITY.md; implement the decision consistently in schema + vectors + prose. Update/add schema-cases and any affected determinism vectors; `make validate` green; generator twice byte-identical if vectors regenerate.
### 2. §9.1 ref selection gap (filed twice by reviewers)
Specify how the operator expresses the ref at `profile install <git-url>` time: propose install-level `--tag|--branch|--revision` applying to the whole repository snapshot (profiles of one repo share one commit by construction of Profilefile), default = tracking the remote default branch, resolved commit recorded as the effective pin; `--strict-tags` semantics unchanged. Align with core §6 grammar; reject mixed per-profile refs in revision 1 with a diagnostic. If you find a strictly better shape, take it and record why.

## Deliverables
Signed commit(s); `make validate` green; board resource `onboarding-import-notes.md` (decisions incl. schema-evolution rationale, loss-list definition, diagnostics added); handoff to-review.

## Do not
Touch profiles/manager.md or cli/curator.md beyond what consistency strictly requires (if §12/cli need a sentence for path/install-ref, add it minimally and note it); no CHANGELOG; no push.
