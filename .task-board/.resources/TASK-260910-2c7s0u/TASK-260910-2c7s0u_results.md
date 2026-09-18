# TASK-260910-2c7s0u results — service-deployment-rate-limit-docs (R4, docs only)

Story `STORY-260910-2xe3n2`, wave 4, final leaf. Docs-only change: no code,
no tests, no spec edits. Worktree
`<control-root>/.temp/STORY-260910-2xe3n2/worktree`, branch
`task-board/story/STORY-260910-2xe3n2`, left uncommitted.

Refresh run (RUN-260918, this run): the Story branch was replayed onto the
fresh trunk `bf5cac1` (R6 `fb86420` + P4 `bf5cac1` landed meanwhile) via
`task-board worktree refresh-candidate` — replayed tip `7a8627b`
(`0b19c4a` R5, `b9847a1` R8, `7a8627b` R7 on `bf5cac1`) — and the R4
documentation was re-applied byte-identical onto the new base (see
"Refresh" section). Line numbers below are post-refresh.

## Per-file changes (docs only)

- `README.md` (+116): new subsections under the existing
  `### Production transport and limits` (README.md:256) — the one deployment
  place (no `docs/DEPLOYING.md` exists; the README does not delegate):
  `#### Rate-limit model and reverse-proxy bucketing` (README.md:276),
  `#### Body-before-auth bounds and proxy mitigations` (README.md:340),
  `#### Operator checklist (proxy and load bounds)` (README.md:377).
- `SECURITY.md` (+12): one threat-model paragraph pointing to that section
  (SECURITY.md:85-96).
- `CHANGELOG.md` (+13): Unreleased `### Security` entry `R4: …`
  (CHANGELOG.md:36-48), newest-first at the top, before the landed P4 entry.
- `git status --short --untracked-files=all` lists exactly these three
  modified files; no code, test, results, logbook, or build files touched
  (`git diff`: 141 insertions, 0 deletions).

## Acceptance criteria

1. Rate-limit model in one place — met. README.md:276-338: limiter table
   (network → `network:<host>` 600/min; semaphore → 128 global slots;
   auditor → `auditor:<id>` 120/min), the shared-bucket collapse
   (README.md:286-289), the exact trust settings with defaults
   (`serve --behind-https-proxy`, default off; `serve --trusted-proxy
   <comma-separated proxy IPs>`, default unset, required with the former; no
   env equivalent — README.md:300-308), and a worked nginx snippet for the
   recommended setup (README.md:309-337).
2. Body-before-auth bounds + mitigations + checklist — met.
   README.md:340-376: 16 MiB / 15 s / 128 slots / 0.1 s acquire with
   `413`/`503 request_timeout` refusal semantics, can/cannot analysis
   (README.md:352-360), `503 overloaded` semantics (README.md:362-366),
   proxy-side mitigations (README.md:368-375); checklist README.md:377-391.
3. SECURITY.md pointer + CHANGELOG R4 — met. SECURITY.md:85-96;
   CHANGELOG.md:36-48.
4. Documented limits match the code — met, zero mismatches (table below).
   Note: the brief cites `README.md:160` for "forwarded headers from other
   sources are ignored"; that sentence now lives at README.md:261 after
   prior leaves' and trunk insertions, and is restated at README.md:305.
   Wording unchanged — a stale line number in the brief, not a code/doc
   mismatch. Likewise the `cli.py` citations moved (R6/P4 added flags);
   behaviour unchanged.
5. Suite unchanged and green; no results/logbook files in the worktree —
   met. Transcript below (219 passed on the refreshed base: R6/P4 added 40
   tests); this results file lives at
   `/tmp/TASK-260910-2c7s0u/results.md` (outside the worktree), attached as
   the task-scoped outcome resource.

## Code-match table (every documented number/setting)

Re-verified on the refreshed base (`7a8627b` + R4 docs delta).

| Documented claim | Code (refreshed worktree) |
|---|---|
| Network limiter keys `network:<client host>`, 600/min, 60 s window, `429 rate_limited` + `Retry-After` | `src/csk_registry/app.py:72` (default 600), `:121` (limiter), `:130-137` (key + 429); `src/csk_registry/limits.py:18` (60 s window) |
| Concurrency 128 slots, 0.1 s acquire, `503 overloaded` + `Retry-After: 1` | `src/csk_registry/app.py:71` (128), `:120` (semaphore), `:148-154` (0.1 s wait, 503) |
| Auditor limiter keys `auditor:<id>`, 120/min, `429`, checked after token verification | `src/csk_registry/app.py:73` (120), `:403-411` (auth then limiter), `:412-418` (429) |
| Order: network → semaphore → (route/auth) → auditor | `src/csk_registry/app.py:131` then `:148` (middleware), `:411` (inside `submit`) |
| Body cap 16 MiB; precheck 413; streaming 413; 15 s deadline → `503 request_timeout`; body read before auth | `src/csk_registry/app.py:63` (16 MiB), `:74` (15 s), `:400-402` (precheck + read), `:403-408` (auth after), `:615-622` (`_read_request_body`) |
| Trust settings are flags only: `--behind-https-proxy` (default off), `--trusted-proxy` (default unset, required with the former); wired to Uvicorn `proxy_headers` / `forwarded_allow_ips`; unlisted sources ignored | `src/csk_registry/cli.py:519-528` (flag defs), `:374-376` (requirement), `:407-408` (uvicorn wiring); no `TRUSTED_PROXY`/`BEHIND_HTTPS` env in service code (grep-verified); Uvicorn `ProxyHeadersMiddleware` honors only `X-Forwarded-For`/`X-Forwarded-Proto` from trusted hosts, and is not installed at all when `proxy_headers=False` (installed uvicorn `config.py:543-544`, `proxy_headers.py`) |
| Env overrides + defaults 600/120/128, positive integers | `src/csk_registry/app.py:798-809` (`_positive_env`); README.md:270-274 pre-existing |
| `503 overloaded` changes no state, is audited | `src/csk_registry/app.py:150-163` (returns before `call_next`), `:156-162` (audit call) |
| Storage 503s are distinct codes | `src/csk_registry/app.py:203-211` (`storage_unavailable`), `:220-229` (`not_ready`) |

The R4 text does not contradict or duplicate the R6 (passphrase key,
`KeyProvider` seam) or P4 (upstream import high-water) sections: disjoint
topics (proxy/rate-limit vs key storage vs import rollback), disjoint
insertion points. No cross-reference was warranted; none added, nothing
restated.

## Validation transcripts

Link check (`python3 /tmp/TASK-260910-2c7s0u/linkcheck.py`, exit 0):

```text
SKIP (external): README.md -> https://github.com/relux-works/curator-spec/blob/main/profiles/registry-service.md
OK: README.md -> SECURITY.md
OK: README.md -> SECURITY.md
OK: README.md -> CHANGELOG.md
OK: SECURITY.md -> README.md#production-transport-and-limits
SKIP (external): SECURITY.md -> https://github.com/relux-works/curator-spec/blob/main/profiles/registry-service.md
checked=4 failures=0
```

(The second `README.md -> SECURITY.md` is the R6 passphrase section's link;
it resolves.)

Test suite (from the worktree, `set -o pipefail`, exit 0):

```text
$ CURATOR_CONFORMANCE_ROOT=/tmp/spec-47c3c8c/conformance/v1 /tmp/csk-venv/bin/python -m pytest -q
219 passed, 2 warnings in 106.37s (0:01:46)
```

- Shell: bash; interpreter: `/tmp/csk-venv/bin/python`, Python 3.14.6.
- Conformance root: detached curator-spec worktree at pinned `47c3c8c`
  (`/tmp/spec-47c3c8c/conformance/v1`), matching
  `.github/workflows/ci.yml` `ref: 47c3c8cbd5da5e3fd6d99b0327d382c6a36494fe`
  (pin unchanged on the fresh trunk).
- No test files modified (`git status` above); warnings are pre-existing
  third-party deprecations (fastapi/starlette/anyio), unrelated.
- 219 vs the pre-refresh 179: R6/P4 added 40 tests on trunk; the R4 delta
  adds none (docs only).

Mypy strict (from the worktree, exit 0, run for completeness on a docs-only
change):

```text
$ /tmp/csk-venv/bin/python -m mypy
Success: no issues found in 15 source files
```

## Refresh (base `c7ef32c` → `bf5cac1`)

Predecessor RUN-260918-b6238d's handoff failed with
`stale-anchor: change_request_base_authority_mismatch` (R6/P4 landed on
trunk meanwhile). This run performed the sanctioned
`refresh-candidate` recovery. Every command quoted with its real exit code.

1. Obligations + pre-refresh state (exit 0):

```text
$ task-board worktree obligations
No Change Request revision is waiting on a next step without a live run.
$ git -C <worktree> status --short --untracked-files=all
M CHANGELOG.md
M README.md
M SECURITY.md
$ git -C <worktree> log --oneline -6
5f3b028 TASK-260910-3u9t1e: TASK-260910-3u9t1e: service-verify-backup-explicit-key
bd27d32 TASK-260910-2rsajv: TASK-260910-2rsajv: service-idempotency-ttl-slack
d7f424c TASK-260910-28kmef: TASK-260910-28kmef: service-recursionerror-400
c7ef32c Compare live state against the operator checkpoint at startup (R3/P2)
...
```

2. Trunk review: `git -C <control-root> diff c7ef32c bf5cac1 --
   CHANGELOG.md README.md SECURITY.md` (exit 0) shows only the R6/P4
   additions (two CHANGELOG entries; passphrase-key and import high-water
   README/SECURITY sections). No contradiction or duplication with the R4
   text (disjoint topics, disjoint anchors); no edits made for the review.

3. Replay with resolutions. `task-board worktree refresh-candidate
   TASK-260910-2c7s0u` (exit 1, expected stop) retained
   `replay-3981731794` on a `CHANGELOG.md` conflict at
   `d7f424c` (R5 entry vs trunk P4/R6 entries). Resolution-file schema is
   not documented by the template (empty `resolutions: []`), so it was
   derived empirically from the tool's strict-decoder errors:
   entry = `{"checkpoint_oid": <REBASE_HEAD full oid>, "path": <unmerged
   path>, "content": <base64 of full replacement bytes>, "sha256": <hex of
   replacement bytes>}` (`commit_oid`/`rebase_head`/… rejected with
   `unknown field`; raw text rejected with `illegal base64 data`;
   `content_sha256` rejected in favour of `sha256`). Replacement contents
   are unions, newest landing first:
   - cp1 (`d7f424c`, R5): CHANGELOG = P4, R6, R5, R3/P2… → replay advanced
     to (2/3), stopped at `bd27d32` (CHANGELOG conflict).
   - cp1+cp2 (`bd27d32`, R8): CHANGELOG = P4, R6, R8, R5, R3/P2… → replay
     advanced to (3/3), stopped at `5f3b028` (CHANGELOG **and** README
     conflicts — the brief's dry rebase mentioned only CHANGELOG).
   - cp3 (`5f3b028`, R7): CHANGELOG = P4, R6, R7, R8, R5, R3/P2…; README =
     R7 `--public-key` paragraph kept in the backup section, followed by
     the P4 `### Upstream import high-water` section (both sides kept, R7
     paragraph first since it continues the backup paragraph).
   - Final `refresh-candidate --replay-resolutions` (exit 0):

```text
{"Outcome": "refresh_advanced",
 "TrunkOID": "bf5cac1200cfa39dff0a6b0449072ff5b22f124d",
 "ReviewedTrunkOID": "bf5cac1200cfa39dff0a6b0449072ff5b22f124d",
 "BranchOID": "7a8627bc314976b459c908a11e6c3bedd0d0e225", "Detail": ""}
```

   Post-replay log (three replayed checkpoints on `bf5cac1`):

```text
7a8627b TASK-260910-3u9t1e: TASK-260910-3u9t1e: service-verify-backup-explicit-key
b9847a1 TASK-260910-2rsajv: TASK-260910-2rsajv: service-idempotency-ttl-slack
0b19c4a TASK-260910-28kmef: TASK-260910-28kmef: service-recursionerror-400
bf5cac1 Persist a per-upstream high-water and refuse rollback bundles on import (P4)
fb86420 Protect the signing key with an optional passphrase behind a key-provider seam (R6)
```

4. Worktree reconciliation (tool behaviour vs brief expectation). The
   refresh moved the branch but left all worktree files at pre-refresh
   content, so `git status` showed 12 modified + 1 deleted paths: the R4
   docs plus apparent reversions of the R6/P4 code (e.g. worktree
   `src/csk_registry/keys.py` byte-matched old tip `5f3b028`,
   `tests/test_key_passphrase.py` absent). The brief expected three
   modified doc files. Recovery, no hand commits, no rebase, no registry
   edits: the R4 delta was extracted as a patch against the old tip
   (141 added lines, pure additions), the worktree was restored to the new
   HEAD (`git checkout HEAD -- .`, status clean), and the exact R4 blocks
   were re-inserted at their anchors (CHANGELOG R4 entry newest-first
   before P4; README subsections after the limiter env fence; SECURITY
   paragraph before the profile pointer). Result: exactly the three doc
   files modified, 141 insertions, 0 deletions; wording byte-identical to
   the predecessor's. Validation transcripts above were run after this
   reconciliation.

## Deliberately out of scope

Forwarded-header trust implementation changes, limiter/semaphore/body-bound
code changes, spec edits, Docker/deploy changes — per the brief. No spec
gap found; no mismatch to report.

## Checklist mapping (task Definition of Done)

1. Deployment docs (rate-limit model, proxy bucketing, forwarded-header
   trust + worked snippet) — done, README.md:276-338.
2. Body-before-auth bounds + proxy mitigations + operator checklist — done,
   README.md:340-391.
3. SECURITY.md pointer + CHANGELOG Unreleased R4; zero code/doc mismatches —
   done, SECURITY.md:85-96, CHANGELOG.md:36-48.
4. Unchanged-green transcript in this results file — done (219 passed).
5. Docs consistent with current code — done (table above).
6. No code/description discrepancies — done (none found).
7. Result linked as task-scoped outcome — this file, attached from outside
   the worktree.
8. Logbook findings — the refresh-behaviour note (§4 above) is recorded
   here; nothing anomalous beyond the stale worktree the tool left behind,
   reconciled as described.

## Revision 2 (rework after rev1 verdict — docs only, README-only delta)

Revision 1 was rejected with three documentation corrections
(`TASK-260910-2c7s0u_review-verdict-rev1.md`, findings F1–F3); replay
fidelity, hygiene and validation passed. This run applied exactly those
corrections. No code or test changes.

Scope proof: `git diff 470a6884 --stat` (rev1 candidate tree → worktree)
touches README.md only (41 insertions, 23 deletions, three hunks);
CHANGELOG.md and SECURITY.md are byte-identical to rev1
(`git diff 470a6884 -- CHANGELOG.md SECURITY.md` empty, exit 0).
`git diff --check` exit 0. New section lines: `#### Rate-limit model…`
README.md:276, `#### Body-before-auth bounds…` README.md:340,
`#### Operator checklist…` README.md:395 (checklist body 395-409).

- F1 — attacker guarantees now match the code (README.md:340-378).
  The intro paragraph is rewritten: the 15 s bound is scoped to the body
  read (`_read_request_body`, `src/csk_registry/app.py:611-623`); the size
  guard refuses once a received chunk pushes the buffered total past the
  limit (`:616-619`), so the docs bound the accepted/buffered body, not
  inbound bandwidth; slot acquisition precedes routing and release follows
  the route (`:148-175`). The "cannot" paragraph is now stage-specific:
  without a valid token there is no auditor-limiter access and no record
  append — token verification runs after the body read and rejects unknown
  tokens first (`:402-408`, then `:411`); a request with a valid token
  reaches the auditor limiter BEFORE signature validation (`:411` vs
  `:435-436`); a record append additionally requires a valid schema
  (`validate_record`, `:419-420`) and a verifying signature (`:435-436`).
  "Changes no state" is qualified as "no registry append or persistent
  registry mutation" in both the cannot-list and the `503 overloaded`
  paragraph (README.md:372-378), since network-limiter accounting and an
  audit event DO happen on refused requests (`:131-163`).
- F2 — nginx timeout model corrected (README.md:326-327, :335, :381-393).
  `client_body_timeout` is documented as the idle gap between successive
  body reads (not a total upload deadline); `proxy_read_timeout` as
  upstream-response reads (not the incoming body); the snippet comments
  match the prose and no "≤ 15 s" claim remains. New paragraph states
  nginx request buffering defaults on (`proxy_request_buffering`), so the
  complete body is received at the edge before forwarding and slow-client
  occupancy sits primarily at the proxy in this setup. Per-client
  connection and body-size caps kept. Authority: the official nginx
  directive docs the reviewer inspected —
  https://nginx.org/en/docs/http/ngx_http_core_module.html#client_body_timeout,
  https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_request_buffering,
  https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_read_timeout
  (cited per the rework brief; not re-fetched in this run).
- F3 — literal `[REDACTED]` placeholder removed (was README.md:344).
  The sentence now reads "only then is the bearer token verified and the
  auditor limiter is checked" (`grep -c REDACTED README.md` → 0, exit 0).
  Cause note: the predecessor reproduced a redacted display string into
  the file; `app.py` never contained it (its matches are display-layer
  redaction only).

Corrections to rev1 results claims: the "zero mismatches" statements in
AC4 (§"Acceptance criteria" item 4) and §"Deliberately out of scope" are
superseded — rev1 prose contained the F1/F2/F3 inaccuracies above, all now
corrected against the cited code lines. The code-match table's numeric
claims (16 MiB `:63`, 15 s `:74`, 128 slots `:71`, 0.1 s / 503 `:148-154`)
stand; its "`503 overloaded` changes no state" row is now read with the
F1 qualification.

Validation transcripts (this run, from the worktree, bash, `set -o
pipefail`, `/tmp/csk-venv/bin/python` Python 3.14.6, conformance root
`/tmp/spec-47c3c8c/conformance/v1` at pinned `47c3c8c`):

```text
$ python3 /tmp/TASK-260910-2c7s0u/linkcheck.py
checked=4 failures=0
linkcheck exit=0
$ CURATOR_CONFORMANCE_ROOT=/tmp/spec-47c3c8c/conformance/v1 /tmp/csk-venv/bin/python -m pytest -q
219 passed, 2 warnings in 84.25s (0:01:24)
pytest exit=0
$ /tmp/csk-venv/bin/python -m mypy
Success: no issues found in 15 source files
mypy exit=0
```

No test files modified; warnings are the pre-existing third-party
deprecations. `git status --short --untracked-files=all` still lists
exactly `M CHANGELOG.md`, `M README.md`, `M SECURITY.md`; no
results/logbook/build files in the worktree. This results file lives at
`/tmp/TASK-260910-2c7s0u/results.md` (outside the worktree), re-attached
as the task-scoped outcome resource.
