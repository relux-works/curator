# Curator Documentation Refresh

Status: for execution. Owner: orchestrator session, 2026-08-27.
Language: English everywhere. Style: docs/prose-style.md (created by this
refresh; until committed, the task precondition resource carries it).
Model mandate: doc-writer runs on agy gemini-3.6-flash-high.

## Motivation

The README is 408 lines and carries a 170-line reference dump
(compiled-command status, diagnostics, repair, maintenance) inside the
pitch. docs/ holds only the build-ssh contract and a stale
implementation plan citing protocol 1.0.0-rc.2. The merged build-https
broker has no contract document at all. CocoaSkills went through this
exact refresh; this plan ports the shape and the useful documents.

## Target document set

- docs/prose-style.md (new): the English prose rules and the
  generated-text blacklist, ported from the CocoaSkills guide
  (/Users/iv/Developer/intranet/cocoaskills/docs/prose-style.md),
  Russian section dropped, examples adapted to Curator.
- README.md (restructured): definition; what Curator manages; install
  with collapsible per-platform options (Homebrew, installer script,
  Scoop, Go toolchain); quick start; a Commands section of collapsible
  groups linking docs/cli.md; the protocol section tightened;
  development kept short with a CONTRIBUTING link. The reference dumps
  move out wholesale.
- docs/compiled-commands.md (new): the current README sections
  "Compiled-command status, diagnostics, and repair" and "Maintenance
  and the build-cache grace period", restructured to the style guide,
  content preserved.
- docs/cli.md (new): full command reference, every synopsis and flag
  verified against the tree binary (go run ./cmd/curator ... --help or
  make build), one section per command group.
- docs/troubleshooting.md (new): symptom, cause, remedy entries drawn
  from the diagnostics prose and the error identifiers in internal/;
  every error string verified against the source.
- docs/build-https.md (new): the operator HTTPS credential contract,
  mirroring docs/build-ssh.md in shape: sources (git-credentials,
  keyring, token_env), scopes, precheck and candidates, env override,
  fail-closed rules. Source of truth: the merged implementation
  (internal/buildrepo, internal/buildhttps or equivalents) and the
  sibling CocoaSkills external-build-repositories.md HTTPS sections.
- docs/build-ssh.md: unchanged except the final style sweep.
- docs/implementation-plan.md: gains a two-line historical header
  (plan of record for v0.1 against rc.2; the board is the live plan);
  content untouched.

## Out of scope

The blocked board tasks compiled-skill-authoring-guide and
external-repository-authoring-and-driver-guide stay with their own
dependency chains. No Russian documents. No intranet mirror.

## Execution

Story on the curator board. Producers: agy gemini-3.6-flash-high with
the shell-only tooling note. Reviewers: claude-opus-5 with explicit
reasoning effort, the single-pass verdict note, and the time-budget
note (no full-suite runs, no clones). Delivery: orchestrator git flow,
PR into main after full per-job CI verification; exclude all
.task-board state from the docs commit (known pitfall, upstream issue
skill-project-management#18).
