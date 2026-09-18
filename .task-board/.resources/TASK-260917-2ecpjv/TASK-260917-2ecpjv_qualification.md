# Qualification — curator-spec v1.0.0-rc.12 (TASK-260917-2ecpjv)

- **Verdict: qualified**
- **Tag:** `v1.0.0-rc.12` → commit `dced9b8317e0e8af79edf2d0539b32bd22b6c85b` (matches expected)
- **Date:** 2026-09-18 (UTC). Story `STORY-260917-3w3lvj`, EPIC-260910-2hw1xb.
- **Method:** read-only. All spec-side commands ran in the disposable
  worktree `/tmp/qual-rc12` (created per the brief's worktree alternative,
  removed afterwards). No pushes, no tags, no source edits in any checkout.
  Every command below is quoted with its real exit code.

## Key aspects

1. The tag is annotated (`cat-file -t` → `tag`) and resolves to exactly the
   expected commit `dced9b8…`.
2. The tag signature verifies verbatim as
   `Good "git" signature for bot@relux.works with ED25519 key SHA256:qbALzjdB9BRgYJjDkX/p9EAPLEofB2AbskJc6Ftwhds`,
   and that fingerprint is the `bot@relux.works` entry in the repository's
   allowed-signers file (cross-checked with `ssh-keygen -lf`, not trusted
   from the verify line alone).
3. `make validate` exits 0 at the tag commit (62 schemas + 1071 vector
   files validated, 301 Python unit tests OK, `go test ./tools/...` ok) and
   `make regenerate-check` exits 0 with the tree clean afterwards.
4. The conformance manifest digest `ea9dd5a0…24ed` matches the
   `manifest_sha256` pinned inside `release/1.0.0-rc.9.json`.
5. All six wave-1 manager vector families are published, including
   `registry-client.json` with 9 `page_boundary_cases`.
6. The only expected-absent family, `environments-source-signers.json` (E1),
   is proven absent at the tag and proven to first appear in the tag's
   direct child commit `684c9f1`.
7. `SPEC_PIN` on `task-board/story/STORY-260917-3w3lvj` equals the full tag
   commit; the disposable checkout was removed; both repositories are
   unmutated.

## 1. Disposable checkout

Per the brief's worktree alternative (avoids touching the main checkout's
files and needs no new clone credentials):

```
$ git -C /Users/administrator/Developer/ReluxWorks/curator/curator-spec worktree add --detach /tmp/qual-rc12 v1.0.0-rc.12
Preparing worktree (detached HEAD dced9b8)
Updating files: 100% (1358/1358), done.
HEAD is now at dced9b8 Bind records and log pages to the committed snapshot boundary (R1/P1)
exit=0
```

(Progress lines elided; final state quoted verbatim.)

## 2. Tag resolution and signature

```
$ git -C /tmp/qual-rc12 fetch -q origin tag v1.0.0-rc.12
exit=0
```

(no output — tag already present and up to date)

```
$ git -C /tmp/qual-rc12 rev-parse v1.0.0-rc.12
3d544cd47bd935d238c1c1c5faa066b92ada281a
exit=0

$ git -C /tmp/qual-rc12 rev-parse 'v1.0.0-rc.12^{commit}'
dced9b8317e0e8af79edf2d0539b32bd22b6c85b
exit=0
```

Commit matches the expected `dced9b8317e0e8af79edf2d0539b32bd22b6c85b`. ✔

```
$ git -C /tmp/qual-rc12 cat-file -t v1.0.0-rc.12
tag
exit=0
```

Annotated tag (tag object `3d544cd…` distinct from the commit). ✔

```
$ git -C /tmp/qual-rc12 cat-file -p v1.0.0-rc.12
object dced9b8317e0e8af79edf2d0539b32bd22b6c85b
type commit
tag v1.0.0-rc.12
tagger Relux Bot <bot@relux.works> 1789729928 +0400

v1.0.0-rc.12 — the 2026-09 security audit and its wave-1/2 specifications

Sixteen commits since v1.0.0-rc.11: the 2026-09 architectural security
audit of the protocol specification with its environments/launch-plane
supplement (E1–E7) and the implementation-verification appendix, and the
first remediation revisions — S6 shell-hook trust gate for project env
files, E2 direct-only class: system modules, E4 umbrella provider
resolution from trust roots, S4 bounded MCP env passthrough and
declaration surfacing, and R1/P1 records and log pages bound to their
committed snapshot boundary (records-response-v2 / log-response-v2, with
the registry-client page_boundary_cases) — plus the repository transport
revision 2, the opt-in local and Git Skillfile sources, the claude/codex
CLI aliases and the migration and credential-mode proposals filed as
draft decisions.

Protocol version stays 1.0.0-rc.9 (release/1.0.0-rc.9.json regenerated at
each revision). This tag is the conformance suite revision the Curator
manager pins as SPEC_PIN for release qualification (TASK-260917-2ecpjv,
STORY-260917-3w3lvj on the Curator board).
-----BEGIN SSH SIGNATURE-----
U1NIU0lHAAAAAQAAADMAAAALc3NoLWVkMjU1MTkAAAAg8bvFNfTkcvVdoPhctSTT9NN6pE
0dsyPkd2eo1B0NMgMAAAADZ2l0AAAAAAAAAAZzaGE1MTIAAABTAAAAC3NzaC1lZDI1NTE5
AAAAQMOhpW8yM6SSwfe95T2ejpEX+1prM0TR30rHoOTw25GiRgCaEVileKUQgYf7beekFi
4ycFtrHY49/b424Fv2kg8=
-----END SSH SIGNATURE-----
exit=0
```

Tagger identity: `Relux Bot <bot@relux.works> 1789729928 +0400`. ✔

```
$ git -C /tmp/qual-rc12 verify-tag v1.0.0-rc.12
Good "git" signature for bot@relux.works with ED25519 key SHA256:qbALzjdB9BRgYJjDkX/p9EAPLEofB2AbskJc6Ftwhds
exit=0
```

Signature verification line quoted verbatim; exit 0. ✔

Allowed-signers cross-check (main checkout config + independent
fingerprint listing):

```
$ git -C /Users/administrator/Developer/ReluxWorks/curator/curator-spec config gpg.ssh.allowedSignersFile
/Users/administrator/developer/curator/.temp/orchestration/allowed_signers
exit=0

$ ssh-keygen -lf /Users/administrator/developer/curator/.temp/orchestration/allowed_signers
256 SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM oparin@me.com (ECDSA)
256 SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM ivan@relux.works (ECDSA)
256 SHA256:qbALzjdB9BRgYJjDkX/p9EAPLEofB2AbskJc6Ftwhds bot@relux.works (ED25519)
exit=0
```

The `verify-tag` fingerprint `SHA256:qbALzjdB…whds` equals the
`bot@relux.works` ED25519 entry in the allowed-signers file. ✔

## 3. Gates at the tag commit

Explicit checkout (the worktree already detached at the tag commit):

```
$ git -C /tmp/qual-rc12 checkout -q dced9b8317e0e8af79edf2d0539b32bd22b6c85b
exit=0

$ git -C /tmp/qual-rc12 rev-parse HEAD
dced9b8317e0e8af79edf2d0539b32bd22b6c85b
exit=0
```

All `make` runs used
`PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH"`.

```
$ make -C /tmp/qual-rc12 validate
python3 tools/validate.py
validated 62 schemas and 1071 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
.................................................................................................................................................................................................................................................
----------------------------------------------------------------------
Ran 301 tests in 456.615s

OK
go test ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	2.097s
exit=0
```

`make validate` exits 0. ✔ (Authority for the test count is the
unittest summary line `Ran 301 tests in 456.615s / OK`, quoted verbatim;
the dot row is progress output.)

```
$ make -C /tmp/qual-rc12 regenerate-check
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
exit=0
```

`make regenerate-check` exits 0 with no diff output. ✔

```
$ git -C /tmp/qual-rc12 status --short
exit=0

$ git -C /tmp/qual-rc12 diff --stat
exit=0
```

Both empty: the tree stayed clean after regeneration. ✔

For reference, the gate definitions at this commit (`Makefile`):

```
validate:
	python3 tools/validate.py
	python3 -B -m unittest discover -s tools -p 'test_*.py'
	go test ./tools/...

regenerate-check:
	go run ./tools/generate-vectors -root .
	git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
```

## 4. Manifest digest, release pin, and vector-family inventory

```
$ shasum -a 256 /tmp/qual-rc12/conformance/v1/manifest.json
ea9dd5a0030b889079cf5655517056e66a0f1d61812f8c2920b4beddc1fd24ed  /tmp/qual-rc12/conformance/v1/manifest.json
exit=0
```

Manifest sha256: `ea9dd5a0030b889079cf5655517056e66a0f1d61812f8c2920b4beddc1fd24ed`.
Structure: `{protocol_version: 1.0.0-rc.9, generator: tools/generate-vectors,
generated_at: 2026-07-13T00:00:00Z, files: list[1071]}` — the 1071 files
agree with the `validate` transcript. ✔

Full content of `release/1.0.0-rc.9.json` (1539 bytes):

```json
{
  "assurance": {
    "default_mode": "portable",
    "portable_execution_policy": "manager-worker-v1",
    "portable_policy": "portable-cli-policy-v1",
    "silent_downgrade_permitted": false,
    "skill_vendored_provider_allowed": false,
    "verified_execution_policy": "verified-provider-execution-v1",
    "verified_implementations": [],
    "verified_platform_claims": [],
    "verified_policy": "verified-provider-policy-v1",
    "verified_provider_contract": "host-execution-provider-v1"
  },
  "candidate_protocol_pin": {
    "manifest_sha256": "sha256:ea9dd5a0030b889079cf5655517056e66a0f1d61812f8c2920b4beddc1fd24ed",
    "suite_root": "conformance/v1"
  },
  "claim_v5": {
    "claim_protocol_version": "1.0.0-rc.9",
    "claims_emitted": [],
    "schema": "schemas/v1/conformance-claim-v5.schema.json"
  },
  "created_at": "2026-08-23T00:00:00Z",
  "downstream_consumption": {
    "committed_release_pin_advanced": false,
    "environment": "CURATOR_CONFORMANCE_ROOT",
    "required_manifest_sha256": "sha256:ea9dd5a0030b889079cf5655517056e66a0f1d61812f8c2920b4beddc1fd24ed"
  },
  "historical_release": {
    "immutable": true,
    "metadata_path": "release/1.0.0-rc.8.json",
    "metadata_sha256": "sha256:293f101d10665061aa049efa72141f9e3c5d608bbde300e882f6e3e095e31ede",
    "protocol_version": "1.0.0-rc.8",
    "source_commit": "f8c405aa3ad0a39d260c2ed93684e55c5a346359"
  },
  "legacy_release": "1.0.0-rc.8",
  "protocol_version": "1.0.0-rc.9",
  "source_baseline_commit": "f8c405aa3ad0a39d260c2ed93684e55c5a346359"
}
```

Both `candidate_protocol_pin.manifest_sha256` and
`downstream_consumption.required_manifest_sha256` equal
`sha256:ea9dd5…24ed`, matching the measured digest above. ✔

### Named vector families (all present)

```
$ ls -la /tmp/qual-rc12/conformance/v1/vectors/manager-config-v2.json /tmp/qual-rc12/conformance/v1/vectors/umbrella-provider-resolution.json /tmp/qual-rc12/conformance/v1/vectors/environments-env-passthrough.json /tmp/qual-rc12/conformance/v1/vectors/shell-hook-trust.json /tmp/qual-rc12/conformance/v1/vectors/registry-client.json
-rw-r--r--@ 1 administrator  wheel   9243 Sep 18 15:15 /tmp/qual-rc12/conformance/v1/vectors/environments-env-passthrough.json
-rw-r--r--@ 1 administrator  wheel  42745 Sep 18 15:24 /tmp/qual-rc12/conformance/v1/vectors/manager-config-v2.json
-rw-r--r--@ 1 administrator  wheel   6528 Sep 18 15:24 /tmp/qual-rc12/conformance/v1/vectors/registry-client.json
-rw-r--r--@ 1 administrator  wheel  19878 Sep 18 15:15 /tmp/qual-rc12/conformance/v1/vectors/shell-hook-trust.json
-rw-r--r--@ 1 administrator  wheel  16347 Sep 18 15:24 /tmp/qual-rc12/conformance/v1/vectors/umbrella-provider-resolution.json
exit=0

$ ls /tmp/qual-rc12/conformance/v1/vectors/ | grep -E '^environments'
environments-env-passthrough.json
environments.json
exit=0

$ ls -la /tmp/qual-rc12/conformance/v1/vectors/environments.json
-rw-r--r--@ 1 administrator  wheel  86074 Sep 18 15:24 /tmp/qual-rc12/conformance/v1/vectors/environments.json
exit=0
```

| Family file | Present | Structure |
|---|---|---|
| `manager-config-v2.json` | ✔ | list[48] |
| `environments.json` | ✔ | protocol 1.0.0-rc.9; `materialization_cases` 24, `header_cases` 4 |
| `environments-env-passthrough.json` | ✔ | protocol 1.0.0-rc.9; `allowlist_empty_cases` 6, `default_resolution_cases` 7, `s4_profiles` 2, `schema_cases` 4, `surfacing_cases` 6, `surfacing_order_cases` 2 |
| `umbrella-provider-resolution.json` | ✔ | protocol 1.0.0-rc.9; `cases` 14 |
| `shell-hook-trust.json` | ✔ | finding S6; `cases` 14 |
| `registry-client.json` | ✔ | `page_boundary_cases` 9 (see below) |

`environments*.json` is exactly these two files — no more, no fewer. ✔

`page_boundary_cases` (count 9):

```
top-level keys: ['key_rotation_resets_state', 'page_boundary_cases', 'pagination_rejections', 'retry_cases', 'retry_policy', 'rollback_state_cases', 'snapshot_transitions', 'state_key']
page_boundary_cases present: True
page_boundary_cases count: 9
page_boundary_cases ids: ['fresh-boundary-advances-high-water', 'equal-version-same-body-accepted', 'equal-version-different-body-rejected', 'below-high-water-rejected', 'chain-boundary-mismatch-rejected', 'missing-boundary-excluded', 'bad-signature-rejected', 'stale-and-mismatch-reports-mismatch', 'higher-and-mismatch-never-advances']
exit=0
```

### Full vector inventory at the tag (30 files)

assurance-modes, build-drivers, canonical-invalid, canonical-valid,
closures, conformance-claim-v3-qualification, context-detectors,
context-versions, environments-env-passthrough, environments,
external-repository-acquisition, external-repository-lifecycle,
go-host-execution-policy, identifiers, locale-selectors,
manager-config-v2, manager-config, manager-lifecycle, module-roots,
portable-paths, registry-behavior, registry-client, registry-resolution,
registry-service, script-host-execution-policy, shell-hook-trust,
skill-manifest-resolution, snapshot-acquisition, source-identities,
umbrella-provider-resolution (all under `conformance/v1/vectors/`,
`*.json`). ✔

### Absent families (expected)

```
$ ls /tmp/qual-rc12/conformance/v1/vectors/environments-source-signers.json
ls: /tmp/qual-rc12/conformance/v1/vectors/environments-source-signers.json: No such file or directory
exit=1
```

`environments-source-signers.json` is absent at the tag (exit 1 is the
expected absence signal, not a gate failure). It first appears in
`684c9f1` ("Verify source signers before the lock … (E1)"), which is the
tag commit's direct child:

```
$ git -C /Users/administrator/Developer/ReluxWorks/curator/curator-spec log --oneline --all -- conformance/v1/vectors/environments-source-signers.json
684c9f1 Verify source signers before the lock and confirm system-changing updates (E1)
exit=0

$ git -C /Users/administrator/Developer/ReluxWorks/curator/curator-spec merge-base --is-ancestor dced9b8317e0e8af79edf2d0539b32bd22b6c85b 684c9f1324d46b4938b2e5943f20c89e27971ec8
exit=0

$ git -C /Users/administrator/Developer/ReluxWorks/curator/curator-spec log --oneline dced9b8317e0e8af79edf2d0539b32bd22b6c85b..684c9f1324d46b4938b2e5943f20c89e27971ec8
684c9f1 Verify source signers before the lock and confirm system-changing updates (E1)
exit=0
```

So rc.12 correctly predates the E1 source-signers family; no other
wave-1 family is missing. ✔

Read-only recount straight from the tag object (no checkout involved):

```
$ git -C /Users/administrator/Developer/ReluxWorks/curator/curator-spec ls-tree -r --name-only v1.0.0-rc.12 -- conformance/v1/vectors/ | wc -l
      30
git_exit=0
```

30 files — agrees with the disposable-checkout listing above. ✔

## 5. Curator-side SPEC_PIN cross-check (read-only)

```
$ git -C /Users/administrator/Developer/ReluxWorks/curator/curator show task-board/story/STORY-260917-3w3lvj:.github/workflows/ci.yml | grep -n SPEC_PIN
53:  SPEC_PIN: dced9b8317e0e8af79edf2d0539b32bd22b6c85b
113:          ref: ${{ env.SPEC_PIN }}
255:          ref: ${{ env.SPEC_PIN }}
377:          ref: ${{ env.SPEC_PIN }}
510:          ref: ${{ env.SPEC_PIN }}
604:          ref: ${{ env.SPEC_PIN }}
exit=0
```

`SPEC_PIN` on the Story branch is the full tag commit
`dced9b8317e0e8af79edf2d0539b32bd22b6c85b`, consumed by five checkout
steps. ✔ (Exit 0 is grep's; the `show` output rendered, so the ref
resolved.)

## 6. Cleanup and no-mutation evidence

```
$ git -C /Users/administrator/Developer/ReluxWorks/curator/curator-spec worktree remove /tmp/qual-rc12
exit=0

$ ls /tmp/qual-rc12
ls: /tmp/qual-rc12: No such file or directory
ls_exit=1
```

Disposable checkout removed (exit 0; the `ls` exit 1 is the expected
gone-signal). It no longer appears in `worktree list` (grep count 0).
The unrelated `/private/tmp/spec-rc12` checkout belonging to another run
was left untouched. ✔

```
$ git -C /Users/administrator/Developer/ReluxWorks/curator/curator-spec status --short
(empty)
spec_exit=0

$ git status --short
(cwd: /Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-260917-3w3lvj/worktree)
(empty)
worktree_exit=0

$ git -C /Users/administrator/Developer/ReluxWorks/curator/curator status --short | grep -v '^\s*M \.task-board/'
?? .task-board/.resources/BUG-260918-18ny6p/18ny6p-republish-rev1.md
?? .task-board/.resources/TASK-260917-2ecpjv/
```

- curator-spec main checkout: clean.
- Story worktree (`task-board/story/STORY-260917-3w3lvj` @ `73fc8a4`):
  clean, no commits made.
- Curator control root: no source file differs; only `.task-board/`
  state written through the `task-board` CLI itself (this run's status
  change plus concurrent runs' activity, and one untracked
  `.task-board/.resources/TASK-260917-2ecpjv/` dir from this run's board
  writes). No pushes, no tags. ✔

## Fact-checking log

| Claim | Evidence |
|---|---|
| Tag → `dced9b8…` | `rev-parse v1.0.0-rc.12^{commit}`, exit 0 |
| Annotated tag | `cat-file -t` → `tag`; tag object `3d544cd…` ≠ commit |
| Tagger `Relux Bot <bot@relux.works>` | `cat-file -p` quoted verbatim |
| Signature valid, bot key | `verify-tag` line verbatim, exit 0, + `ssh-keygen -lf` fingerprint match |
| `make validate` exit 0 | Full transcript §3 (62 schemas, 1071 vectors, 301 tests OK, go test ok) |
| `make regenerate-check` exit 0, clean tree | Empty-diff run + empty `status --short`, exit 0 |
| Manifest digest `ea9dd5…24ed` | `shasum -a 256`, exit 0; echoed by both pin fields |
| Pin content | `release/1.0.0-rc.9.json` reproduced in full |
| 6/6 families present | `ls` transcripts + structure table; `page_boundary_cases` count 9 with ids |
| `environments*` exactly 2 files | `ls … \| grep -E '^environments'` → 2 lines |
| 30 vector files total | Checkout `ls` + independent `ls-tree … v1.0.0-rc.12` recount (30) |
| `environments-source-signers.json` absent, lands at `684c9f1` | `ls` exit 1 + `log --all -- <path>` → only `684c9f1` + `merge-base --is-ancestor` exit 0 + single-commit range log |
| SPEC_PIN equals tag commit | `show …:.github/workflows/ci.yml \| grep -n SPEC_PIN` line 53 |
| Disposable removed, repos unmutated | `worktree remove` exit 0, `ls` gone, clean statuses |

Corrections during drafting: the vector-file total was first written as
33 from a hand count and corrected to 30 after the independent `ls-tree`
recount — the transcript above is the authority.

No anomalies, regressions, or unexpected findings occurred, so no
logbook entry is warranted (checklist item 11's "when relevant" clause
does not trigger).

## Task questions — answered

1. *Does the annotated signed tag resolve to the expected commit?*
   Yes — `v1.0.0-rc.12^{commit}` is `dced9b8317e0e8af79edf2d0539b32bd22b6c85b`.
2. *Does the signature verify against the maintainer key?*
   Yes — `Good "git" signature for bot@relux.works … SHA256:qbALzjdB9BRgYJjDkX/p9EAPLEofB2AbskJc6Ftwhds`,
   and that key is listed in the repository's allowed-signers file.
3. *Do `make validate` and the regeneration check pass at that commit?*
   Yes — both exit 0 in the disposable checkout; tree clean afterwards.
4. *Manifest digest and pin content?*
   `ea9dd5a0030b889079cf5655517056e66a0f1d61812f8c2920b4beddc1fd24ed`;
   `release/1.0.0-rc.9.json` reproduced in full in §4; pin's
   `manifest_sha256` fields match the measured digest.
5. *Are the wave-1 manager families published?*
   Yes — manager-config-v2, environments (+ environments-env-passthrough),
   umbrella-provider-resolution, shell-hook-trust, and registry-client
   with 9 `page_boundary_cases` are all present.
6. *Was any repository mutated?* No — see §6.

**Verdict: qualified.** `v1.0.0-rc.12` at
`dced9b8317e0e8af79edf2d0539b32bd22b6c85b` satisfies every brief step.
