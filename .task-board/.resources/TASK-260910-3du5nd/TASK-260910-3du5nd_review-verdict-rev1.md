# TASK-260910-3du5nd revision 1 review
VERDICT: CHANGES_REQUESTED

Candidate: CR-TASK-260910-3du5nd-1, tree a54f8afd3b38e471c89c05ffef42b6e8b9bdb745; base a68854d54725862f6f696019ef2e569f3ec29cd6.
Reviewed exact delta and worktree content. All 1327 candidate entries matched: hash-object initially reported two CRLF fixture differences caused by filtering; direct byte comparisons confirmed both equal. No candidate edits made.

## Required rework
1. High: contradictory alias/mirror admission, protocol/repository-transport.md:175-207. Lines 178-181 require mirror_of iff URL host differs and reject it on same-host URLs; lines 203-207 require it when an alias target differs. The shipped valid-alias.json, v2-alias-resolution semantic case, and operator-guide example use a canonical URL host plus a different alias target with mirror_of. They must simultaneously be admitted and refused. Define one predicate covering URL and resolved connection host, precisely scope the spurious-attestation rule, and state which actual host the attestation authorizes (line 187 currently says listed host only). Cover canonical URL + mirror alias, mirror URL + alias, and same-host alias with and without attestation. Ensure mirror resolution explicitly preserves the canonical entry-key identity rather than the mirror URL identity (line 203 currently says identity always derives from URL).

2. High: strict external SSH acquisition is not specified consistently. Revision 2 section 4 admits port-bearing listed endpoints and section 7 preserves manager section 11 mandatory boundaries. profiles/manager.md:1249 onward forbids explicit ports; section 11.3 fixes a three-argument wrapper invocation, explicitly rejects -p, and fixes an SSH command without a port. The protected record also contains no port. Thus ssh://git@example.org:2222/kit.git has no conforming strict execution path. Alias connection rewriting is similarly not reconciled with the exact validated URL/host tuple. Specify the opt-in revision-2 protected record, argument validation, resolved-host/port broker and host-key binding rules as an additive exception while retaining revision-1 behavior; or explicitly define which mappings strict lanes refuse and align the claimed scope. Add concrete strict-lane positive/refusal vectors. Do not leave implementers to invent a wrapper bypass.

3. Medium: SSH port-bound refusal lacks a distinguishing indexed vector. An in-memory mutant widened only the SSH upper port branch from 6553[0-5] to 6553[0-9], retaining the HTTPS restriction. Detection: 0/12 new indexed cases. Add SSH upper-bound refusal and positive boundary vectors (including alias integer-port bounds where relevant), so the transport-specific narrowing is detected. This is a test coverage finding, not a claim that the current schema admits overflow.

## Validation and bounds
- Independently ran the README draft conformance command: exit 0. Marker migration 25/25 fields, 2/2 arms; existing narrowed marker mutants 18/18 detected. Schema cases 114/114, negatives 90/90, wire schemas 8/8; snapshots 3/3.
- New source-policy-v2 indexed structural cases: 12/12 pass; independent SSH port probes 5/5 (:1, :65535 accepted; :0, :65536, :01 rejected).
- The draft checker explicitly reports Manager semantic execution: 0 cases. The 18 new semantic vectors are contract examples, not executed resolver evidence. Green validation does not resolve findings 1 and 2.
- First make invocation used a nonexistent control-root .venv and exited 2 with ModuleNotFoundError: jsonschema. Recovered using the producer-documented existing worktree .temp/venv; no dependency installation.
- Independent zsh command: PATH="$PWD/.temp/venv/bin:$PATH" make validate: exit 0.
  Output: validated 60 schemas and 1047 vector files; python3 -B -m unittest discover -s tools -p 'test_*.py': Ran 227 tests in 200.267s, OK; go test ./tools/...: ok github.com/relux-works/curator-spec/tools/generate-vectors (cached). Python checks executed; Go result reused its test cache.

Revision-1 substantive sections remain intact; source-policy-v1 byte-unchanged; changes are additive draft documents/schema/fixtures with appropriate index/changelog/unresolved-question updates. Producer evidence and runtime validation log were read; independent checks above were rerun by this reviewer.

Run goal queried before verdict: not goal-bound. No external blocker or human decision is required: this is specification rework. Route to to-dev and require a new producer handoff and independent review. Findings recorded in this task-scoped outcome and board notes; LOGBOOK.md intentionally not edited under campaign rules.
