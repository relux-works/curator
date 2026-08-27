# TASK-260824-3ppds1 — reviewer verdict: ACCEPTED

Independent verification of the curator-spec `1.0.0-rc.9` publication. Every
line below was re-derived by the reviewer from the remote / the published
artifacts, not read off the implementer's report.

## 1. Tag identity, type, and signature

| Check | Command | Result |
| --- | --- | --- |
| Tag exists on the remote | `git ls-remote --tags origin 'v1.0.0-rc.*'` | `b6796644…` → `refs/tags/v1.0.0-rc.9`, peeled `0ed5c691…` |
| Tag is annotated, not lightweight | `git cat-file -t v1.0.0-rc.9` | `tag` |
| Tag object hash | `git rev-parse v1.0.0-rc.9` | `b67966449220d42218bd50420e74dac673431464` |
| Target commit | `git rev-parse v1.0.0-rc.9^{commit}` | `0ed5c691e9208eea52f21db2fc05e226ce3516fd` |
| Signed by an allowlisted maintainer | `git -c gpg.format=ssh -c gpg.ssh.allowedSignersFile=maintainers.allowed_signers verify-tag v1.0.0-rc.9` | exit 0 — `Good "git" signature for oparin@me.com with ECDSA key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM` |
| Key is the allowlist entry | `cat maintainers.allowed_signers` | single entry, same ECDSA key, `oparin@me.com` |

Tag body: `object 0ed5c691…` / `type commit` / `tag v1.0.0-rc.9` /
`tagger Ivan Oparin <oparin@me.com>` / message `Curator Protocol v1.0.0-rc.9`,
followed by an `BEGIN SSH SIGNATURE` block.

**Repo tag convention:** existing remote tags are `v1.0.0-rc.1`, `.2`, `.3`,
`.5`, `.7`, `.8` — all `v<version>`. `v1.0.0-rc.9` matches. (rc.4 and rc.6 were
never tagged; that predates this task.)

## 2. Release target is the landed main

- `origin/main` = `0ed5c691e9208eea52f21db2fc05e226ce3516fd`
  ("Land schema 8 with the implementation pins that consume it (#29)").
- `git merge-base --is-ancestor 0ed5c691… origin/main` → **YES**. The tag targets
  the current tip of the protected default branch exactly, not a descendant or a
  side commit.
- All check-runs on `0ed5c691…` report `success`: Formatting, Links,
  Release target provenance, Specification (ubuntu/macos/windows-latest),
  Implementations (ubuntu/macos/windows-latest), plus the release job itself.
  Legacy commit-status count is 0, so the API's `"pending"` combined state is the
  empty-status artifact, not a real pending check.

## 3. Regeneration determinism (GOVERNANCE.md step 2) — re-run by the reviewer

Fresh detached worktree at `v1.0.0-rc.9`, go1.25.5 darwin/arm64:

| Step | Result |
| --- | --- |
| recursive SHA-256 over `conformance/v1`, committed state | `f76a2894595b2299f036a2a1d15362a5feb7a7d78017270abfd9224a7c3ff9a5` |
| `go run ./tools/generate-vectors -root .` (run 1) | exit 0 |
| tree hash after run 1 | `f76a2894…` (unchanged) |
| `git diff --exit-code -- conformance/v1 release/1.0.0-rc.{5,6,7,8,9}.json` | exit 0 |
| `go run ./tools/generate-vectors -root .` (run 2) | exit 0 |
| tree hash after run 2 | `f76a2894…` (unchanged) |
| `git diff --exit-code -- …` (second) | exit 0 |
| `git status --short` | empty |

Generation is idempotent and the committed vectors are exactly what the
generator produces. Note: the reviewer's tree-hash recipe differs from the
implementer's (`54d8e7e1…` in the results artifact), so the two digests are not
comparable — each is internally consistent across its own two runs, and the
authoritative invariant (`git diff --exit-code` clean on both runs) holds under
both. Not a discrepancy.

## 4. Release workflow

Run [32764992277](https://github.com/relux-works/curator-spec/actions/runs/32764992277):
`Release specification`, event `push`, `headBranch v1.0.0-rc.9`,
`headSha 0ed5c691…`, `conclusion success`, 18:53:46Z → 18:54:44Z. Every step of
job "Publish signed specification artifacts" is `success`.

The `Validate release input` step at the tag (`.github/workflows/release.yml`)
points `gpg.ssh.allowedSignersFile` at `$PWD/maintainers.allowed_signers` and
runs `git verify-tag "$GITHUB_REF_NAME"` — so CI, not just the local machine,
verified the tag against the maintainer allowlist. Verified in the run log:

```
Good "git" signature for oparin@me.com with ECDSA key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM
release target 0ed5c691e9208eea52f21db2fc05e226ce3516fd is a GitHub-verified commit on origin/main
validated 53 schemas and 691 vector files
Ran 134 tests in 11.942s
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	0.060s
release gate passed for 1.0.0-rc.9 at 0ed5c691e9208eea52f21db2fc05e226ce3516fd
```

The same step also re-ran `go run ./tools/generate-vectors` +
`git diff --exit-code`, i.e. a third independent clean regeneration on the
release target. The `Package normative release` step produces the checksums via
`sha256sum … > checksums.txt`.

## 5. Published artifacts — verified by independent download

Release `v1.0.0-rc.9`: `isDraft=false`, `isPrerelease=true`,
`publishedAt=2026-08-24T18:54:36Z`.

| Asset | Size | State | SHA-256 (from `checksums.txt`) |
| --- | ---: | --- | --- |
| `curator-protocol-1.0.0-rc.9.tar.gz` | 291597 | uploaded | `524f505c5f9170f15730485888db27dfa8ad48ee2939176e35a225daf3a01bd7` |
| `curator-protocol-1.0.0-rc.9.zip` | 738498 | uploaded | `dc8df7112418d636be86fc089eb7162409136d60ccc472bf41849e6e908b33cf` |
| `checksums.txt` | 199 | uploaded | — |

- `gh release download` into a clean dir, then `shasum -a 256 -c checksums.txt`
  → **exit 0**, both `OK`.
- Archive carries the normative surface: **53** JSON schemas under `schemas/v1/`,
  plus `protocol/`, `profiles/`, `cli/`, `conformance/`, `decisions/`, `docs/`,
  `release/`, `reviews/`, `maintainers.allowed_signers` and the governance docs.
- **The archive is the pinned tree, not a rebuild.** `diff -r` of every packaged
  path against the `v1.0.0-rc.9` worktree is clean for all 14 packaged
  paths (`protocol`, `profiles`, `cli`, `conformance`, `schemas`, `decisions`,
  `docs`, `release`, `reviews`, `README.md`, `CHANGELOG.md`, `GOVERNANCE.md`,
  `RELEASE.md`, `maintainers.allowed_signers`).
- Suite pin holds three ways:
  `conformance/v1/manifest.json` in the archive =
  `803918bf8672f76cf990985e51db213b826674cd5bb54fbf47731b8404b44403`,
  identical to the same file at the tag, identical to
  `release/1.0.0-rc.9.json → candidate_protocol_pin.manifest_sha256`
  (`suite_root: conformance/v1`).
- SLSA build provenance: `gh attestation verify --repo relux-works/curator-spec`
  → exit 0 for both archives. Predicate `https://slsa.dev/provenance/v1`,
  subjects `[checksums.txt, curator-protocol-1.0.0-rc.9.tar.gz,
  curator-protocol-1.0.0-rc.9.zip]`, built by `.github/workflows/release.yml` at
  `refs/tags/v1.0.0-rc.9`. Negative control (an unattested file) → exit 1 with
  HTTP 404, so the pass is real.

## 6. GOVERNANCE.md release process — all five steps

1. **CHANGELOG/version metadata/schemas/vector manifest** — `CHANGELOG.md` at the
   tag carries `## 1.0.0-rc.9 - 2026-08-23` (manifest schema 8, install marker
   schema 4, implementation coverage contract, live pin moved to
   `release/1.0.0-rc.9.json`, rc.5–rc.8 byte-frozen). `release/` holds exactly
   rc.5 … rc.9. ✅
2. **Regenerate twice, clean second run** — re-run by the reviewer, §3. ✅
3. **Required checks green on the protected default branch** — §2. ✅
4. **Annotated, cryptographically signed `v<version>` tag** — §1, and CI
   re-verified it against `maintainers.allowed_signers`. ✅
5. **GitHub release with normative schemas + conformance archive + SHA-256
   checksums** — §5. ✅

`RELEASE.md`'s mechanically enforceable candidate items are the body of
`tools/release_gate.py`, which passed both locally and in CI against
`0ed5c691…`. The Release-target-provenance items are covered by
`verify_release_merge_policy.py` / `verify_release_commit.py` (the latter
re-run in CI, output quoted in §4).

## 7. Scope discipline

No repository source changed for this task; `curator-spec` main is at
`0ed5c691…` with a clean working tree (only pre-existing untracked
`.DS_Store`, `.temp/`, `task-board.config.json`). The only new immutable object
is the signed tag. The implementer's "no new tests" reasoning is correct — the
behavior exercised is existing release tooling, and its suites (134 Python
tests, the Go generator tests) ran green against the exact release target in CI.

The implementer's recorded `PYTHONDONTWRITEBYTECODE` note (a bare local
`release_gate.py` run writes `tools/__pycache__` and then trips its own
clean-checkout check; CI sets the env at job level) is an accurate, useful
operational finding, not a defect.

## Acceptance criteria

| AC | Verdict |
| --- | --- |
| Tag `1.0.0-rc.9` — repo convention verified against existing tags — signed and pushed | ✅ `v1.0.0-rc.9`, annotated, SSH-signed by the sole `maintainers.allowed_signers` key, on the remote at `0ed5c691…` |
| Release workflow green | ✅ run 32764992277, `success`, every step green, signature re-verified in CI |
| Release artifacts published with sha256 checksums | ✅ tar.gz + zip + `checksums.txt`, `shasum -c` exit 0 on an independent download, plus verified SLSA provenance |

**Verdict: ACCEPTED.** Reviewer-archetype run — no `commit_ack` supplied. No
commit is pending for this task: nothing was committed to `curator-spec`, the
release is already published, and the tag is immutable.
