# Review findings: protocol/environments.md cycle 2 (TASK-260901-ugnl0q)

Reviewer run: RUN-260901-fa73e7. Subject: worktree
`~/Developer/ReluxWorks/.worktrees/curator-spec-environments-normative`,
branch `draft/environments-protocol`, head `c3b29b1` (rework of reviewed
head `eddd509`). Inputs read in full: `review-findings-environments-1.md`,
`environments-rework-report-1.md`, both review briefs.

**Verdict: ACCEPT.** All six cycle-1 findings are genuinely resolved in the
document text at `c3b29b1`; no new finding of any severity. The two
dedicated-code branches the cycle-1 review offered (`environment_path_collision`,
`environment_surface_unreadable`) are the branches taken, each with exactly
one defining condition and one owning table row. Regression sweep of the
delta into §5/§7/§8/§13 is clean. Per the brief, the draft stays at
to-review for the orchestrator's closure; no `accept_cr` (no Change Request
section on this run).

## Git facts (verified, not accepted from the report)

- Head `c3b29b1f7f37829fd4d0c50b2023efa2feb4c615`; delta
  `eddd509..c3b29b1` is `protocol/environments.md` only, +48/−16
  (`git diff --numstat`).
- Good "git" signature for oparin@me.com, ECDSA
  `SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`, verified exit 0
  against the repo's own `maintainers.allowed_signers` (the path in the
  worktree's git config points at a deleted tmp dir; verification was run
  with `-c gpg.ssh.allowedSignersFile=$PWD/maintainers.allowed_signers`).
- Merge-base with `origin/main` is `2a861e5`, unchanged. Worktree clean.

## Finding-by-finding verification (against document text)

### M1 — managed `opencode.json` bytes — RESOLVED, attacked

§5.3 opencode bullet now pins the bytes: the CCJ-1 bytes (registry.md §1)
of the object whose single member `instructions` is the ordered
`.agent-context/modules/<profile-name>/<module-path>` list, no other
member, followed by exactly one trailing LF.

Producibility attack: CCJ-1 (read at head) is fully deterministic — sorted
keys, no insignificant whitespace, a closed escaping table, UTF-8, integers
irrelevant here — and defines a byte string with no trailing newline of its
own, so "+ exactly one LF" composes without ambiguity. I generated the
bytes two independent ways (compact `json.dumps` and a hand-rolled CCJ-1
string emitter) from the sample module list: byte-identical
(`{"instructions":[".agent-context/modules/companyA/00-base.md",...]}` +
LF). Zero-module case (§5.4) produces `{"instructions":[]}` + LF —
producible. Core §2 portable paths forbid `\`, NUL, and control characters
and preserve Unicode scalars without normalization, so path values cannot
hit an escaping ambiguity; a legal `"` in a path escapes deterministically.
The cycle-1 failure scenario (indented serialization of the same object)
now violates the rule instead of conforming with a different hash. The §13
"referenced-form layout" vector surface is producible. §5.6 hash set
unchanged and now well-defined for opencode referenced form.

### M2 — platform-path collision rule — RESOLVED, reach verified

New normative "Platform-path collisions" paragraph in the §5 opening: any
materialization or provisioning step writing two protocol paths that map
to one platform path MUST fail with `environment_path_collision` before
writing anything — explicitly the core §2 extraction rule (quote verified
at core.md: "Filesystem extraction MUST detect two protocol paths that map
to one platform path and fail before writing") extended to every
materialization and provisioning write path.

Reach across all three named write paths, in text:
1. composed-profile module trees — named in the §5 paragraph; §5.3 now says
   "collision-free in protocol-path space" and points folding cases at the
   §5 rule ("by construction" removed; the one surviving "by construction"
   at §5's part-joining is the LF argument, which is genuinely structural);
2. managed homes — §8.1's managed-home bullet names folding profile names
   below the environments root as a §5 collision failing with the code
   before writing;
3. backup paths — named in the §5 paragraph with the §8.3 pointer.

The rule is phrased in terms of the platform-path mapping, not case folding
alone, so Unicode-normalization folding (APFS/HFS+) is covered too —
matching core §2's no-normalization comparison rule. Bypass probe: a
cross-step folding write (chain changed from `[Base]` to `[base]` into the
same home) is outside the single-step collision check but cannot corrupt
silently — the §8.3 marker ledger either removes the recorded old files
first (clean) or refuses the overwrite as
`environment_surface_unmanaged_conflict`; fail-closed either way. Not a
finding.

### m1 — §7.2 code — RESOLVED

§7.2 names the error `environment_form_unsupported` inline; §7.7 has the
row "configured form not supported by the adapter". Disjoint from
`environment_form_unavailable` (§5.3/§5.7, warning + monolithic fallback);
both conditions remain separately defined, no overlap.

### m2 — §7.3 flag-spelling caveat — RESOLVED

§7.3 now carries the §7.6-style sentence: spellings recorded from vendor
documentation, exact spellings verify against the pinned tool releases
before the conformance vectors freeze. The caveat rides where the claim
rides. (Spellings remain documentation-sourced — the same open pre-freeze
item as §7.6's Xcode paths; that is what the finding required.)

### n1 — diagnostics headings — RESOLVED

`### 6.1 Diagnostics` and `### 11.1 Diagnostics` added; every section's
diagnostics table now sits in a numbered subsection, numbering consistent
with the §1.1/§2.1/§3.1 pattern for sections with no other subsections.

### n2 — surface-read failure code — RESOLVED

§8.4 separates the facts: failed marker read → `environment_marker_invalid`;
failed read of a recorded surface file → `environment_surface_unreadable`,
row non-current, currency reported as unknown, and no absence-shaped
outcome — `environment_surface_missing` included — may fire on either.
§8.5 row added. §12 already agrees ("unreadable evidence is reported as
unreadable, never as absence (section 8.4)") — cross-reference verified.

## Regression sweep of the delta (§5/§7/§8/§13)

- Full diagnostics-code sweep of the document (grep, every occurrence
  mapped to prose vs table): the three new codes each have exactly one
  owning table row (§5.7, §7.7, §8.5) and defined conditions; no orphan
  code introduced; no new duplicate. Pre-existing pattern of one condition
  appearing in two sections' tables (`environment_surface_unmanaged_conflict`
  in §5.7 and §8.5; `environment_unknown`/`environment_target_unknown`
  across §7.7/§9.6/§10.4) predates the rework, was passed by cycle 1, and
  is a per-section-relevance house pattern, not a contradiction — noted,
  not a finding.
- New codes leak into no other protocol document (grep over core.md,
  registry.md, assurance.md: zero occurrences) — environments.md remains
  the owning surface.
- §5.3's `[registry.md](registry.md) §1` link resolves; CCJ-1 is §1 there.
- §13 unchanged in scope; "referenced-form layout" now producible (above).
- No other wording in §5/§7/§8 contradicted by the inserted paragraphs;
  the reflowed lines are wrapping artifacts only.

## Validation reran at c3b29b1 (myself, not accepted from the report)

Scratch venv from `requirements-dev.txt` (ambient python3 lacks
`jsonschema`): `tools/validate.py` exit 0 — "validated 53 schemas and 691
vector files"; `python -B -m unittest discover -s tools` — 134 tests OK;
`go test ./tools/...` — ok. Matches the producer's claim; docs-only delta
confirmed independently via `--numstat`.

## Routing

ACCEPT. No blocking/major/minor/nit filed. Draft branch stays as-is at
`c3b29b1`, producer task remains at to-review; orchestrator owns closure
(including the standing pre-freeze open items: header `generated:` URL,
§7.3 flag spellings, §7.6 Xcode paths — all already tracked). DoD item
"If review does not accept…" is vacuously satisfied by acceptance.
