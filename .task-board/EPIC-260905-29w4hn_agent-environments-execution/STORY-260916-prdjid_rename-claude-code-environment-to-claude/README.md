# CLI aliases claude and codex for the environment ids (wire ids unchanged)

## Description
Operator decision 2026-09-16 (revised after review): the wire environment identifiers claude_code and codex_cli stay frozen (environments 1.1 rev 1, launch-env-fragment v1, markers, defaults.json, packages). Instead, the CLI surfaces accept the short spellings claude and codex as aliases and normalize them to the canonical ids before anything else: Curator machine CLI (curator run <env>, curator env resolve|status, profile use --env, config knobs given on the command line) and the launcher curator-run <env>. Outputs, markers, fragments and config files keep the canonical ids. Spec: one alias rule in profiles/manager.md CLI section (and the launcher SPEC), no schema or vector changes.

## Scope
(define story scope)

## Acceptance Criteria
Spec amendment accepted and landed; curator resolves both ids with claude canonical and claude_code deprecated (warning), existing managed homes keep working without re-provisioning; curator run claude works; packages validate with the new id; conformance vectors updated; all landed through signed PRs with independent review.
