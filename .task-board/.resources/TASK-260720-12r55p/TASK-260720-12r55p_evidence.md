# TASK-260720-12r55p developer evidence

## Immutable inputs

- CocoaSkills base SHA: `870daa30aea0ed4dc5554ac5dcd0c671f8d04e09`
- curator-spec candidate SHA: `432eb2ee1fe2d6b271e37269f867c8851c325539`
- `conformance/v1/manifest.json`: `sha256:12e58b82579645ba1ccafba49d3e2dd3216005ddf37ae63c68a9fafd46773071`
- Candidate root supplied through `CURATOR_CONFORMANCE_ROOT`.
- No `.github` curator-spec ref, release pin, tag, GitHub Release, or conformance claim changed.

## Delivered coverage

- 102 selected generated schema cases across agent-skill-v6, csk-skill-v6,
  build-receipt-v1, install-marker-v2, and conformance-claim-v1/v2/v3.
- 8 build-driver positives, 77 build-driver rejections, 10 build-source cases,
  and 12 toolchain cases, all read from the candidate root.
- 18 mandatory controls, 14 identity/protocol cases, 8 package-influence
  cases, 11 capability-evidence cases, 6 deferred-capability guards, three
  failure boundaries, and the exact five-control macOS/Windows inventory.
- All 11 byte-exact `expected/build-driver/` artifacts, including cache key
  `sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b`
  and receipt hash
  `sha256:919fbbad8e6ce95532219fd952c2309d0d7026f85209650508fd6834af4020cd`.
- All 32 manager-lifecycle cases and all candidate claim-qualification rules.
- Claim-v3 remains pinned to protocol 1.0.0-rc.5; legacy rc.3 coverage remains
  in the passing full suite.

## Authoritative local gates

- Focused candidate-root conformance/build-policy pytest: exit 0, 643 passed.
- Full candidate-root pytest: exit 0, 1709 passed and 54 skipped in 305.16s.
- `python -m mypy`: exit 0, no issues in 68 source files.
- `python -m build`: exit 0; sdist and wheel built with the staged adapter.
- `python -m twine check dist/*`: exit 0; both artifacts passed.
- `git diff-tree --check -r HEAD^ HEAD`: exit 0.
- `.github` diff from the recorded base: exit 0 with no changes.
- Worktree status after commit: exit 0 with no tracked changes.

The repository has no configured standalone lint command. The applicable
whitespace/diff validation above passed.

## Commit and pull request

- Signed CocoaSkills commit:
  `b754bd7aeba87baca0c63435ddc6a14d80c29400`.
- Git signature status: `G`, signer `oparin@me.com`, key
  `SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`.
- Pull request: https://github.com/ivanopcode/cocoaskills/pull/19
- Hosted CI run: https://github.com/ivanopcode/cocoaskills/actions/runs/30686258237
- Handoff snapshot: hosted strict mypy and all Ubuntu/macOS Python 3.11-3.14
  jobs passed; four Windows jobs remained pending. `gh pr checks 19` therefore
  returned exit 8, correctly indicating pending checks. Reviewer acceptance
  must require the exact-head hosted matrix to become green.

## Non-authoritative development iterations

- The first baseline invocation used unavailable `python` from the ambient
  PATH and returned exit 127; rerunning through the task virtual environment
  returned exit 0 with 167 passing tests.
- The first new harness run returned exit 1 with nine adapter failures; the
  corrected rerun returned exit 0 with 441 passing cases.
- The first widened run returned exit 1 with one cross-vector identity
  mismatch; the corrected rerun returned exit 0 with 643 passing tests.
- Two preliminary full-suite invocations outlived detached output handles, so
  their exit codes were not captured and they are not presented as evidence.
  The separately tracked authoritative full-suite process above returned exit
  0. The hosted-check watcher later ended with exit 130; it was monitoring only
  and no hosted result is represented as green while Windows remains pending.
