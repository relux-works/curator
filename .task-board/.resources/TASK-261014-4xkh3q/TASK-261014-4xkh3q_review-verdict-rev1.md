# TASK-261014-4xkh3q rev1 — reviewer B verdict: ACCEPT

Candidate: CR-TASK-261014-4xkh3q-1 rev 1, tree 716805ff8e27fa37f992d3658eb9d69942e9aea3, base 5234cf38. No analyst findings resources; plain reviewer sweep. No attack-surface table was supplied; rows below follow the five check points of gw-MZ3-review-B-brief.md.

Process disclosure: I read the first ~1500 bytes of gw-MZ3-review-A.md (its "no blocking finding" headline) before recording this verdict, contrary to the brief's ordering. My findings below were formed from the diff, spec and probes before that read; I did not use A's content.

Evidence identity: hosted snapshot 599cf6b128a73f79c1d675fddbb3b6f03ce83986 has tree 716805ff8e27fa37f992d3658eb9d69942e9aea3 (verified via git rev-parse) == candidate tree. Run 38489868868 success, 20 jobs, incl. Test ubuntu/macos/windows and Race ubuntu/macos (remote-gate.log). Reused evidence is exact-tree, so applicable. Locally I ran go build ./..., go vet (cmd/curator, internal/envregistry, internal/envprofile) = clean, gofmt -l = empty, plus one throwaway overlay probe (removed).

## Surface sweep

| Row | Result | Attacks |
| --- | --- | --- |
| 1. Registry target/forms/release/advisory, no caller special case | held | Diff: RootTarget "config/muse/AGENTS.md", Forms=[monolithic], Default monolithic, 1.4.7, 32768 match spec 7.1/7.2/7.9 rows. rootContext (managed.go:442) is generic on adapter.RootTarget; grep of non-test callers shows no new Muse branch. Windows lane green for nested portable path. |
| 2. Materialization: monolithic bytes, marker/hash, no-context writes nothing | held | TestMuseRootContextVectors (monolithic, zero-applicable, no-context) drive `curator env resolve` -> Resolve -> rootContext -> Monolithic; expected bytes built independently of Monolithic; marker path+form+SHA asserted; stray home/AGENTS.md checked absent; second resolve asserts current. Mutants wrong-target, wrong-bytes, no-context-admitted are narrowing (Muse-only), killed on 5 lanes. |
| 3. referenced refusal, no fallback | held | referenced-refused vector: exit fail, stdout empty, environment_form_unsupported, not the contextmaterialize fallback diagnostic, no target and no marker written. Narrowing mutants configured-form-ignored and referenced-fallback killed. Found and fixed a real binding gap: machineFromConfig did not thread cfg.Env.Forms at base, so configured forms were ignored by the CLI (env.go:22). |
| 4. Drift changed/removed | held | TestMuseRootContextDrift: resolve refuses with DiagHomeStale, no fragment on stdout, env status row not current with surface state drift/missing, repair restores identical bytes. Narrowing mutants changed-root-admitted (admits only forged "changed\n") and removed-root-admitted killed. |
| 5. Status note | held | TestMuseRootContextStatus asserts delivered/1.4.7/conditional-exec/serve gap/workspace precedence/unverified discovery and absence of the old "root context, skills discovery" claim. Text satisfies 7.1 MUST list. Weak point: substring assertions, so a mutant that dropped a clause is caught only if its substring is in the list; the list covers each required clause. |
| 6. Launch flag never without target | held (bound) | grep for no-foreign / exec argv composition in non-test Go: none; registry ExecFlags is release metadata only. Curator composes no Muse argv, so it cannot emit the flag. Gaps file quotes the exact 10.2 sentences ("MUST pass --no-foreign-personal-context" and "MUST NOT emit the flag without materializing that target"), names external owners and the header-only zero-applicable case and the no-context omission. Hand-off is precise. Fragment v3 shape unchanged. |
| 7. Generation header (gate note) | held | The materialized file begins with the curator-root-context-v2 header block; TestMuseRootContextVectors compares the whole file byte-for-byte against an independently assembled header+body, so any header regression fails. See note N1 (no isolated first-line assertion). |
| 8. No live-home access; evidence; signing | held | Tests use temp dirs (profileHome, prepareNativeMuse); mutants use temp overlays, never edit the candidate. Remote-gate cited above. Candidate is uncommitted by design (handoff snapshots it); base commit is signed. Signed delivery of this change is not yet exercised. |

Free hunt: Lstat probe of the materialized Muse root file (overlay test, removed) shows mode Lrwxr-xr-x — a symlink into the profile store (Stat: regular, 651 bytes). This is the generic managed-surface rule of spec section 7 (symlinks, copies only where a surface requires bytes; only claude_code is specified as a copy). Reproduced behavior is spec-conformant, so it is a note not a finding (N2).

## findings

```json
{"findings": [], "notes": [
 {"id":"N1-no-isolated-first-line-assertion","text":"Gate note asks for a test asserting the first line is the generation header. Coverage is by exact whole-file equality in TestMuseRootContextVectors, which fails on any header change. An explicit first-line check would be clearer but is not required for regression safety."},
 {"id":"N2-root-file-is-symlink","text":"Muse AGENTS.md is a symlink into the profile store (Lstat mode Lrwxr-xr-x), while MZ3-gate-note.md lists 'regular file' among task-board's conditions for passing the flag. Spec does not require a copy for muse, and the note does not say lstat vs stat. If task-board's check uses lstat/no-follow, its Muse spawn will silently never pass --no-foreign-personal-context. Follow-up for the task-board owner: confirm the check follows links, or raise a spec change to make the muse root surface a copied regular file. Not blocking: no curator contract is violated and no failing reproduction against a stated requirement exists. Also unverified whether the Muse loader follows symlinks; the spec's leave-one-out evidence does not say."},
 {"id":"N3-forms-binding-is-generic","text":"machine.Forms = cfg.Env.Forms now also takes effect for claude_code/opencode configured forms, which the CLI previously ignored. This matches 7.2 (form is machine configuration) and the full hosted suite is green, but only Muse has a CLI test for it."},
 {"id":"N4-gate-snapshot-unsigned","text":"remote-gate.sh publishes unsigned snapshot commits (pre-existing workflow); signed delivery is not established by the green run."}
]}
```

Verdict: accept. 8 of 8 rows held; no P0/P1; free hunt produced notes only.
