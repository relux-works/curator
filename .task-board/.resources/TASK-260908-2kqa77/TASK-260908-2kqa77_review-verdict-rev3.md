# TASK-260908-2kqa77 revision 3 review — CHANGES REQUESTED

Candidate tree: `2b0015ce2e71f5335bda5126befa8096750220a7`, base `09b25ef6629b41455d91dcb252ab4e4034e12750`. All four changed working files byte-match candidate blobs. No production code changed during review. Route: `to-dev`; no acceptance or commit acknowledgement.

## Findings requiring rework

### R1 — High: first-position YAML structure can forge a missing channel value

`.github/ci/goreleaser-config-gate.sh:160-168,189-192`: the dash-line branch never starts block-scalar handling for `- description: |`, and leaves `key_indent=-1`. The next scalar-content line establishes entry depth and records its text as a real skip_upload field. A first-position `repository:` similarly lets its nested key establish the wrong entry depth. Both return **0**, claiming every value is auto, with the actual entry-level skip_upload absent.

Executable reproduction (save this as fixture.yml, then run `bash .github/ci/goreleaser-config-gate.sh fixture.yml`; observed exit **0**, required **1**):

```yaml
homebrew_casks:
  - description: |
      skip_upload: auto
    name: curator
scoops:
  - name: curator
    skip_upload: auto
release:
  prerelease: auto
```

Replace `description: |` with `repository:` for the second bypass. `gopkg.in/yaml.v3.Unmarshal` confirms both fixtures are valid YAML and neither has entry-level skip_upload. A fixture derived from the entire committed config, retaining the remaining cask fields and moving description to the first position, also passes the gate while yaml.v3 confirms skip_upload absent (attached `committed_block_bypass.yml` and oracle output).

The committed negative for a nested repository starts after an ordinary entry key and misses this order-dependent hole. Use the existing YAML library or a parser with explicit structural state; persist these production-script negatives in self-test. Avoid extending a text walk by only special-casing these two strings.

### R2 — Medium: invalid duplicate keys are accepted contrary to the declared parser contract

Gate header line 27 and results claim last-key-wins matches yaml.v3. It does not: yaml.v3.Unmarshal rejects duplicate mapping keys. A cask containing `skip_upload: true` followed by `skip_upload: auto` returns **0** from this gate; yaml.v3 returns `mapping key "skip_upload" already defined`. Reject malformed/duplicate mappings and add an executable row. This is a parser-consistency defect; no claim that upstream GoReleaser accepts this invalid YAML is needed or made.

### R3 — Medium: wiring pin accepts a non-executable invocation

`.github/ci/gate-selftest.sh:1215`: substring grep counts `# run: bash .github/ci/goreleaser-config-gate.sh` as a live invocation. In a copied workflow, commenting the run line leaves the exact extracted goreleaser self-test section **24 passed, 0 failed, exit 0**. Deleting that line or changing its path produces **23 passed, 1 failed, exit 1**, as required. Pin executable workflow structure (including disabled steps), not comments/substrings; add the commented/disabled invocation negative. A valid workflow mutation adding `if: false` to that step also returns **24 passed, 0 failed, exit 0**. The comment-only fixture may itself make the workflow invalid; the disabled-step fixture demonstrates the live-step pin hole without that ambiguity. Actual candidate wiring is currently active; this finding concerns the required regression pin.

## Portability pin assessment and bound

The current awk syntax audit found no listed GNU-only constructs and the current hyphen-last class is correct. Local execution used macOS `/usr/bin/awk` only; no local gawk or Git-Bash awk available. The new source pins detect the exact historical spelling, not arbitrary invalid ranges. A copied gate replacing split_key class `[A-Za-z0-9_.-]` with `[A-Za-z0-9_.+-9]`, leaving is_block unchanged, passes all **24/24** extracted rows locally. Another copy using `[+-9]` in is_block and retaining `[+0-9-]` in a comment likewise passes **24/24**. Thus the source pin is a historical-regression tripwire, not a general portability proof. Do not claim general coverage from it. All three hosted self-test lanes actually execute the gate on the committed file and pass fixtures, so a runtime parse failure on Linux/Windows would fail those rows regardless of these greps. Current hosted evidence is sufficient for current portability; local equivalent-mutant behavior on GNU awk remains unexecuted, not claimed.

## Independent execution and coverage

Production entry point: `bash .github/ci/goreleaser-config-gate.sh <fixture>`.

| Probe | Observed exit |
|---|---:|
| committed file | 0 |
| row C first skip_upload absent | 1 |
| row E Auto | 1 |
| sometimes | 1 |
| boolean true | 1 |
| prerelease ato | 1 |
| nested repository after ordinary entry key | 1 |
| nested repository as first entry key | **0 (bypass)** |
| description block as first entry key, containing skip_upload | **0 (bypass)** |
| ordinary description block containing skip_upload | 1 |
| second scoop with true | 1, names scoops[1] |
| tab indentation | 1 |
| duplicate key true then auto | **0 (invalid YAML admitted)** |
| trailing comment on auto | 0 |
| empty channel stanza | 1 |
| missing channel stanza | 1 |
| nonexistent/unreadable path | 1 |

Required five value negatives detected **5/5**; minimal first-key structural absence attacks detected **0/2**; duplicate-key rejection **0/1**. Workflow/source mutants were run separately from baseline; deletion and path change rejected, comment, disabled-step and equivalent range survived. Fixture probe JSON includes an initial bad-second/flow-second generator that also altered prerelease; the corrected bad_second fixture was rerun separately and reports only scoops[1].skip_upload=true. Do not use the initial malformed fixture as extra coverage. A missing path was tested; permission-denied on an existing file was not separately driven.

Exact extracted goreleaser section with shipped assertion functions: **24 passed, 0 failed, exit 0**. `bash -n` both scripts and `git diff --check`: exit **0**. Full gate-selftest result is appended below. No full landing-suite replay and no goreleaser execution.

## Hosted evidence reused, independently checked

[Run 35729614942](https://github.com/relux-works/curator/actions/runs/35729614942) succeeds. GitHub API reports head `6acb6fabd5a0ddd5301a3bdb207ee590a8c6ee48`; its Git tree is exactly `2b0015ce2e71f5335bda5126befa8096750220a7`.

Downloaded logs show all named rework rows green on all three platforms (prerelease wording, second-entry wording, good second entry, unquoted auto), plus the other assertions in that section. Gate self-test summaries: macOS **160 passed, 0 failed, 2 skipped**; Ubuntu and Windows each **162 passed, 0 failed, 1 skipped**. Lint's explicit gate invocation prints the 2-entry pass. These are hosted evidence, not locally rerun platform claims. Candidate suite and rose-air were skipped; rose-air is unverified.

## Scope and project fit

Read the whole config: its independent tap/bucket publishers are homebrew_casks and scoops; brews is absent. nfpms builds release assets, not another independent tap/bucket publisher. Mapping the three checked values to these actual surfaces is appropriate. A new publisher stanza is outside measured coverage. CI lint and three self-test lanes have no added conditional skip; existing workflow push triggers are main and gate/**, plus pull_request and dispatch (not arbitrary branch pushes). The change preserves those existing lanes. CHANGELOG entry exists. The exact four-path diff changes neither .goreleaser.yml, release.yml nor the ancestry gate.

Use yaml.v3 already in the module as the recommended implementation direction; the brief explicitly permits a Go tools test. Required next cycle: repair parsing and executable CI pin, add durable negatives, narrow tests, then normal producer handoff and independent review. No human decision or external blocker is needed. Per campaign rules, no control-root LOGBOOK edit; significant findings are also recorded in task notes.

## Completed local self-test

`bash .github/ci/gate-selftest.sh` on macOS/BSD awk: **exit 0; 209 passed, 0 failed**. This was independently rerun to completion; the full landing suite was not rerun. Full local log and isolated fixtures/mutants are attached as `TASK-260908-2kqa77_review-evidence-rev3.tar.gz`. All launched processes completed before verdict recording.
