# Review verdict — CR-TASK-260825-2fy132-3 (revision 3)

**Verdict: CHANGES REQUESTED → to-dev.**
Reviewer run: RUN-260825-a57cd2 (claude). Reviewed delta: `903af23..14423a9c` (4 paths: CHANGELOG.md, LOGBOOK.md, README.md, docs/build-https.md).

## Blocking finding H1 — rev3 denies the shipped interactive candidate prompt (regression of rev1 F1)

The requested rev3 change was one sentence: drop the word "private" from the
candidate-prompt sentence (rev2 verdict finding G1), change nothing else.
Instead, rev3 rewrote the Resolution section, the precedence list, the
host-narrowing sentence, and the LOGBOOK 0057 CORRECTION paragraph to state
that no install-time candidate discovery or prompt exists — inverting text the
rev1 and rev2 reviewers had verified against the delivered code.

### Why the rev3 claims are false against the accepted delivery

The producer verified against the primary checkout
`/Users/iv/Developer/ReluxWorks/curator`, which genuinely lacks the prompt.
But the orchestrator's landing map (TASK-260825-1d0eo5 notes, "verified by the
orchestrator — supersedes the earlier note") rules that tree out:

> the primary checkout holds an OLDER partial copy of the same code and must
> not be used as the source — it lacks buildhttpsprompt.go and the resolver
> wiring.

The authoritative complete code set is `.temp/STORY-260825-32bopo/worktree`.
Verified there during this review:

- `cmd/curator/main.go:538` and `:1300` — `opts.External.BuildHTTPS.Resolve = operatorBuildHTTPSResolver(cfg, opts.DryRun)`.
- `cmd/curator/main.go:1341-1349` — resolver is nil (→ anonymous continue)
  exactly when `dryRun || !terminal(stdin) || !terminal(stderr)`; otherwise it
  returns `install.InteractiveBuildHTTPSResolver(...)`.
- `internal/install/buildhttpsprompt.go:29` — `InteractiveBuildHTTPSResolver`
  exists, with `ErrBuildHTTPSAborted` abort paths (lines 75, 120, 126).
- `internal/install/buildhttps.go` — `resolveBuildHTTPS` collects every
  unmatched planned HTTPS row, calls presence-only `reader.Discover` per host,
  invokes `selection.Resolve`; a resolve error (abort) stops the run; only the
  nil-resolver path marks rows anonymous.
- TASK-260825-3kb532 ("install-precheck-and-candidates") is `done` — accepted
  at 2026-08-25T00:48Z with verdict evidence for exactly this prompt (default,
  this-run-only, abort semantics on both surfaces).

The primary checkout is a stale intermediate state (its code files predate the
story-worktree copies; the accepted 3kb532 layer never reached it or was
superseded there). Three runs of this docs task have now been misled by it.
The false rev3 claims:

1. docs/build-https.md Resolution section: "Curator does not discover
   credential candidates or prompt during installation … This is the same in
   terminal, headless, and dry-run runs."
2. docs/build-https.md precedence list reduced from four items to three
   (terminal prompt item deleted).
3. docs/build-https.md host-narrowing sentence: "may use a matching configured
   scope or remain anonymous" (dropped "be offered candidates on an operator
   terminal").
4. LOGBOOK.md 0057 "CORRECTION TO THE DOCUMENTATION BOUNDARY": "There is no
   `InteractiveBuildHTTPSResolver`, `operatorBuildHTTPSResolver`, or
   install-time candidate prompt in the delivered source." Rev2's CORRECTION
   paragraph said the opposite — and rev2's was correct.

### Required rework (rev4 = rev3 plus exactly this)

Only docs/build-https.md and LOGBOOK.md change; keep README.md and
CHANGELOG.md exactly as rev3.

**docs/build-https.md** — restore the three rev2 passages, with G1 applied:

1. Resolution section: replace the rev3 paragraph ("transport, so an uncovered
   repository is fetched anonymously. Curator does not discover … select a
   credential for a repository.") with the rev2 text, dropping only the word
   "private" from the prompt sentence:

   > No configured or environment source is not an error. HTTPS has an anonymous
   > transport, so an uncovered repository can be fetched anonymously. On an
   > operator terminal, however, an uncovered repository first opens a
   > candidate prompt before any fetch. The prompt offers an existing Git HTTPS
   > credential for that host when presence-only discovery finds one, or lets the
   > operator enter a token now. No candidate is read or used until the operator
   > selects it. The operator then chooses a scope to persist or a this-run-only
   > (`r`) choice; the latter never writes configuration or credential storage.
   > Aborting the prompt stops the run rather than silently falling back to
   > anonymous HTTPS.
   >
   > Headless, non-terminal, and dry-run runs never prompt. Their uncovered HTTPS
   > repositories continue anonymously. `list` is a presence-only check for
   > configured candidates; it does not reveal tokens or select a credential for a
   > repository.

2. Precedence: restore the four-item order (3. "On an operator terminal, the
   credential candidate prompt for an uncovered repository." 4. "Anonymous
   HTTPS when neither applies and no terminal prompt is active.").

3. Host-narrowing sentence: restore "They may use a matching configured scope,
   be offered candidates on an operator terminal, or remain anonymous."

Keep rev3's warning block as-is ("every HTTPS build repository host" — the
G1 rationale applies there too; rev3 did that part right).

**LOGBOOK.md** — in entry 0057, replace the inverted CORRECTION paragraph with
rev2's CORRECTION paragraph verbatim (it corrects the boundary bullet to
describe the shipped prompt: presence-only discovery offer, explicit selection,
persist or this-run-only, abort stops the run, anonymous only for headless /
non-terminal / dry-run). Keep rev3's SECURITY bullet (private already
dropped).

**Verification for rev4:** verify prompt claims against
`.temp/STORY-260825-32bopo/worktree` (authoritative per the landing map in
TASK-260825-1d0eo5 notes) — not the primary checkout and not your own
worktree's base. Name the tree you read in your evidence.

## What rev3 got right (keep, verified)

- The one requested G1 fix is present: "private" dropped from the warning
  block, the CHANGELOG entry, and the LOGBOOK SECURITY bullet.
- CHANGELOG additionally gained the `Spec core §12.2` citation — fine, keep.
- README delta is byte-identical to rev2 (the merged credentials bullet the
  rev2 review accepted as F3 closure); links `docs/build-ssh.md` and
  `docs/build-https.md` both resolve.
- All retained shipped-output claims still hold: the config CLI code
  (`cmdConfigBuildHTTPS*`, `formatBuildHTTPS`, list line `%s\t%s present=%t`,
  empty-list stderr `curator: no build_https scopes are configured`,
  added/replaced verbs) is byte-identical between the primary checkout and the
  authoritative story worktree, so the rev1/rev2 reviewers' byte-identical
  transcript verification carries over; format strings re-confirmed in source
  this round.
- `CaptureBuildHTTPSSelection` re-confirmed: captures
  `CURATOR_BUILD_HTTPS_TOKEN` / `CURATOR_BUILD_HTTPS_HOST` and `token_env`
  values at capture time; precedence items 1–2 match.
- Exposure warning present at every place the override is documented
  (docs page warning block, CHANGELOG entry, LOGBOOK SECURITY bullet).
- No other manager implementation named anywhere in the delta.
- `git diff --check` clean on the exact delta; `make lint` exit 0 in this
  story worktree (producer evidence, rev3 rework note); the delta touches no
  Go code.

## Process note for the orchestrator (not a rework item)

The stale primary checkout has now misled three runs of this task (rev1
producer, rev3 producer's blocker, rev3 producer's rework). Until landing,
any further spawn touching HTTPS behaviour should be pointed explicitly at the
landing map's tree list. The rev3 producer's blocker artifact
(`TASK-260825-2fy132_blocker.md`) and rework note record honest observations
of the wrong tree — the error is in tree selection, not fabrication.
