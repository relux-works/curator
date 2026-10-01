# Review verdict — TASK-261001-3s8csu rev1: ACCEPTED

Reviewed delta e87d488b..1d3acaf8 (.research/261001_mandates-launch-context-advice.md, LOGBOOK.md). Read-only; I reran source reads myself (git/gh at the pinned commits), not the producer's tests.

## Orchestrator spot-checks
1. No blessed per-consumer trust dir: CONFIRMED. Curator-spec main 0400feab, protocol/environments.md: only XDG hits are the OpenCode ambient XDG_DATA/STATE rule (l.1296) and a state-sharing table row (l.1644). Repo-wide grep for mandate/pin store/per-consumer finds only an unrelated word in docs/security-audit-2026-09.md:199. Curator main has no such dir (managed layout managed.go:56-76, EnvRoot/ManagedParent). The claim is stated as a bounded contract finding, which is right.
2. Versions: CONFIRMED. v0.14.0 (d23dcd16) and v0.15.0-rc.2 (bd3c0f43) envfragment.go:24 = launch-env-fragment-v1; main :24 = v2. Launcher main ee66c107 (ls-remote confirms head): IdentityV3 "adds Muse and its four XDG parents"; Muse env closed to XDG_CONFIG/DATA/STATE/CACHE_HOME, each must be <home>/{config,data,state,cache}; no HOME member; Muse requires v3. Spec main schemas/v1 contains only v1 and v2 launch-env-fragment schemas (v3 is the diverged candidate d373078a, as the document says).
3. New field needs versioned change: CONFIRMED. v2 schema top-level additionalProperties:false (l.98), env minProperties/maxProperties 1 with a propertyNames enum (l.~20-25); launcher fragment.go:340-356 closedObject. Decision 0018:174-184 gives the new-token precedent. Ownership (curator-spec maintainers) matches GOVERNANCE.
4. Drafts: CONFIRMED. waggle c6182028 spec/waggle.md:3 "DRAFT v5.2 ... Not normative", README:7 "nothing here is implemented yet"; curator-trust a6a9bd03 README:13 "Nothing is implemented", spec/trust.md:3 DRAFT v0.4; trust.md:282-296 separate-agent-account direction cited accurately.
5. Wording: draft is labelled DRAFT, says "mandates' orchestrator decides" and "Advice only", separates shipped/proposed, 17 lines (<25). Launch-path claims are scoped (local path only; tracked host behaviour explicitly not established; execution.go Setpgid only, no credential change — read at process.go:39-49).
6. Hygiene: no secrets/credentials in the diff (grep clean); no employer names seen. Document is 31,029 bytes (<40 KiB).

## Non-blocking notes for the poster (no change required)
- Last draft line points at a board outcome the peer team may not be able to open; consider inlining the 3-4 key citations (managed.go:56-76, v2 schema l.98, launcher fragment.go:416-444, waggle/trust README status lines) when posting.
- Q11 asks whether Curator should pass MANDATES_MANIFEST; the draft answers "yes as discovery pointer, no fragment field exists", correctly flagged as not implemented.

Not re-run: the producer's two go test commands (scoped contract tests); accepted as recorded, source claims were verified directly instead.