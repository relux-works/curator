# TASK-260924-4mzun5 — record-dependency-directory-in-legacy-lane: revision 4 review

Verdict: **changes_requested**. Route to `to-dev` for the one P1 finding below. The separate P2 BUG note does not independently hold acceptance. No Story product code, commits, acknowledgements or integration changes were made by this reviewer. This run is not goal-bound (`task-board spawn goal` returned none).

Reviewed base `d46f2b1f41ee24276fc1d809bcb7bcef51ad5250` and candidate tree `88551d4e3b3442e6d4a11f175b853348f5054e01`, 57 changed paths. Fresh origin/main advertised and fetched that base. The current candidate was reviewed as a whole, including the additive CHANGELOG entry.

## Decision and previous findings

RECORD is correct. Curator-spec `7eaeb73fcf22cb8ce8e21325a8237accdfaf9646`, [skillfile-sources §4](https://github.com/relux-works/curator-spec/blob/7eaeb73fcf22cb8ce8e21325a8237accdfaf9646/protocol/skillfile-sources.md#4-runtime-builds-audit-and-persistent-records), explicitly requires draft marker v6 for a schema-9 installation including under Skillfile schema 1. [Core §4.4](https://github.com/relux-works/curator-spec/blob/7eaeb73fcf22cb8ce8e21325a8237accdfaf9646/protocol/core.md#44-dependencies) requires normalized directory in install/lock identity and the same package identity in audit records. The schema-9/subdirectory migration, audit identity, effective-lock digest, runtime source-v1 key and receipt-3 carrier are implemented at production entries.

Revision-2 F6-declared-ref-currentness is fixed within checked scope: permanent project/global CLI regressions pass. My independent `TestReviewR4StatusBindingAndRepair` passes both entries: initial current status, changed v1-to-v2 declaration at the same commit becomes needs-install, `--check` fails, status preserves the marker bytes, and reinstall restores currentness. A `status-commit-only` mutant fails both permanent regressions and both independent subcases.

Revision-2 F4's original unchanged-v1-to-unchanged-v2 reproduction now passes on project/global installs, and the `movedtag-commit-only` mutant fails it. However, the replacement gate does not prove that the installed declared tag stayed unchanged. The F4 invariant remains broken for ordinary repositories with multiple tags on one commit.

Revision-1 F1/F2/F3/F5 remain fixed within the current checked scope: the exact-tree full gate is green; targeted permanent runtime, effective-lock/directory, builds/receipt repair and package-cache tests pass. This review does not reuse an earlier revision's full gate as evidence for revision 4.

## P1 finding: F4 — unrelated tag defeats strict same-tag movement policy

**Severity: P1.** A normal multiple-tag history silently publishes a moved declared source despite explicit StrictTags policy, on both project and global production install entries.

At `internal/install/install.go:1557–1565`, `detectMovedTagsIn` calls `gitops.TagsPointingAt` on the old package commit and warns only when **no tag at all** still points at that commit. A tag named `archive` does not prove that the installed tag named `v1` stayed intact. The package marker still has no recoverable prior ref evidence, and the new live-tag inference substitutes unrelated repository state for it.

Reproduction through real `Project`/`Global` entries:

1. Install a schema-9 consumer declared as v1 under a schema-1 Skillfile with StrictTags enabled.
2. Add a lightweight or annotated `archive` tag to the installed commit.
3. Make a second commit and force-move **v1** to it. Leave the Skillfile declaration unchanged.
4. Reinstall with StrictTags enabled.

Expected: refusal with `moved tag for consumer: v1 old -> new`, preserving the previous installation. Actual: `Status:ok`, the new generation is installed, and no moved-tag warning is emitted. This reproduced **4/4** alias cases (project/global × lightweight/annotated). No-alias controls correctly refused **2/2**, proving the guard is reachable and the alias is the bypass.

[Core §10](https://github.com/relux-works/curator-spec/blob/7eaeb73fcf22cb8ce8e21325a8237accdfaf9646/protocol/core.md#10-install-markers) retains moved-tag refusal under strict policy. [Manager §2.1](https://github.com/relux-works/curator-spec/blob/7eaeb73fcf22cb8ce8e21325a8237accdfaf9646/profiles/manager.md#21-read-only-planning-and-source-gates) keeps moved-tag evaluation mandatory, including dry runs and cache hits. Documenting this blind spot in a comment/results does not satisfy either requirement.

Required rework: bind/recover the exact previous declared selection and ref from validated installed-generation evidence tied to the marker's lock digest, and distinguish movement of that same tag from an explicit declaration change. Preserve the closed v5/v6 carrier and schema-1 re-resolution semantics. Do not infer a previous declaration from the presence or absence of arbitrary live tags. Keep the existing changed-declaration admission and no-alias refusal controls, add the attached alias probes permanently at both production entries, and add a narrowing mutant that waives same-tag refusal only when an unrelated tag still binds the prior commit; the new tests must kill it.

`repeat-of: none` by mechanism: revision 2 compared commits unconditionally; revision 4 newly infers declaration continuity from arbitrary live-tag presence. The **F4 class recurs**, so the regression and narrowing mutant stay in this owning task. This is ordinary implementation rework; no external blocker, new research prerequisite or human-only decision was established.

## Separate non-blocking BUG note

**P2-BUG-1 — explicit change to unmoved v2 can still be falsely refused after old v1 is deleted.** Severity P2: this rejects a valid explicit operation, but is recoverable and is not an independent reason to hold landing. `repeat-of: none` for the new arbitrary-live-tag mechanism, in the recurring F4 class.

Install at v1; create v2 at another commit; delete old v1; change the Skillfile declaration to v2. v2 never moved, yet both real install entries refuse with `moved tag for consumer: v2 old -> new`. Reproduced **2/2** by `TestReviewR4ChangedTagAfterOldTagDeleted`. The same `len(bound)==0` inference causes this false positive. Keep this as a separate BUG note; fixing the validated prior-ref comparison for P1 should also preserve this explicit-declaration admission.

## Swept surfaces

| Surface | Result and evidence |
| --- | --- |
| Spec decision and v6 carrier | Held: explicit §4 decision; marker implementation/schema review and exact-tree full gate |
| Project/global runtime and GC | Held in checked scope: both permanent source-v1 runtime rows pass; project launch/reinstall/GC assertions pass |
| Effective lock and manifest binding | Held in checked scope: full-closure canonical lock implementation inspected; production directory row reconstructs and compares shared lock digest; same-commit declaration probes observe lock drift |
| Compiled providers and receipt-3 records | Held in checked scope: production receipt-3 publication and tampered-receipt repair rows pass; exact-tree multi-platform full gate green |
| Legacy moved tags and repeat installs | Broken: original explicit-change rows pass, alias bypass 4/4; P1 F4; separate P2 false refusal 2/2 |
| CLI status/currentness | Held for F6: project/global permanent rows plus independent read-only/reinstall probes pass; commit-only mutant killed |
| Audit install/CLI/sourceaudit/cache | Held in checked scope: directory/install and CLI record assertions pass; full-package equality tests pass; sourceaudit identity plumbing inspected |
| Dependency directory glob grammar | Held: production ?/[/* rows 3/3 and grammar/schema subcases 31/31; star-only mutant killed by ?/[ dependency rows |
| Vendored v1/v2 corpus preservation | Held: v1 changed paths 0; 24/24 v2 files byte-identical to pinned spec |
| Generation reads and atomic publication | Source review and exact-tree full gate; no independent concurrency attack in this round |

## Evidence identities and measured bounds

Accepted attached full-suite evidence: [run 37884289505](https://github.com/relux-works/curator/actions/runs/37884289505), success. GitHub commit metadata verifies its head `6f29efa9d38313facc925cb973b45fb1d7d808b7` has the **exact CR tree**. Test ubuntu/macOS/Windows, race ubuntu/macOS, lint, driver matrix, interop, naming and gate self-tests passed. Candidate-suite and rose-air lanes were skipped, not proven. I did not replay the full gate.

Reran myself: baseline `go build ./...` and scoped `go vet` for install/CLI/skillspec/marker/audit, exit 0; the scoped vet type-checked the two new review files. Each of `glob-star-only`, `movedtag-commit-only`, `status-commit-only` compiled locally with `go build ./...`, exit 0, then was restored byte-for-byte. Candidate diff --check passed. No local go test, compiled test execution or go run occurred.

My [hosted targeted run 37888437685](https://github.com/relux-works/curator/actions/runs/37888437685) used immutable snapshot `eb9de6731acc6073f0f5479ae736a511868dc4f6`, tree `b7d0f645a4a8fa2fc229f29f73425616e19e444a`. It differs from the CR only by two reviewer probe files and two harness files. Baseline passed **18/18 top-level tests**: install 8, CLI 5, grammar 3, audit 2. Grammar subcases passed 31/31; dependency glob rows passed 3/3. Narrowing mutants were killed **3/3**, each after successful hosted compilation. The attack job intentionally expects the product probes to fail; its green workflow result verifies the reproduced failures, not successful product behavior.

Coverage is targeted. No exhaustive conformance/mutation claim is made. Hybrid selection-index cases, mixed hash-framing/registry combinations, external build arms and independent concurrency faults were not attacked in this round. The full configured gate provides broader regression evidence, without proving these additional adversarial properties.

Attached evidence: `TASK-260924-4mzun5_review-evidence-rev4.log`, `TASK-260924-4mzun5_review-install-probes-rev4.go`, `TASK-260924-4mzun5_review-cli-probes-rev4.go`, `TASK-260924-4mzun5_review-harness-rev4.txt`, and `TASK-260924-4mzun5_review-logbook-rev4.md`. Logs neutralize fixture temporary paths. Reviewer logbook evidence is attached without editing repository LOGBOOK/CHANGELOG. The hosted scratch branch is removed after terminal evidence capture.

```verdict-findings
{
  "findings": [
    {
      "id": "F4",
      "row": "Legacy moved tags and repeat installs",
      "invariant": "StrictTags refuses movement of the same installed declared tag regardless of other tags pointing at the prior commit, while explicit declaration changes remain installable.",
      "mechanism": "internal/install/install.go:1557-1565 infers continuity of the installed declaration from the presence of ANY live tag on the old package commit; an unrelated lightweight or annotated archive tag suppresses same-v1 movement detection and permits strict publication on project/global entries.",
      "reproductions": [
        {
          "test_file": "TASK-260924-4mzun5_review-install-probes-rev4.go",
          "command": "Hosted: go test -count=1 -timeout=5m -v ./internal/install -run '^TestReviewR4SameTagMoveWithAlias$'; run 37888437685; snapshot eb9de6731acc6073f0f5479ae736a511868dc4f6",
          "expected_failure": "Four alias subcases return Status:ok and publish the moved v1; both no-alias controls refuse correctly.",
          "pinned_blobs": [
            "sha256:94bfb858d0d9153e7ce99ddd3ca0bca9ab72e2c90afdfb5c51842925c4da4d50"
          ]
        }
      ],
      "severity": "bypass",
      "category": "bypass",
      "severity_reason": "Silently publishes a moved declared source despite explicit StrictTags policy under ordinary multi-tag repository history on both production install entries.",
      "repeat-of": "none",
      "tb_r226_severity": "P1"
    }
  ],
  "bug_notes": [
    {
      "id": "P2-BUG-1",
      "row": "Legacy moved tags and repeat installs",
      "severity": "P2",
      "severity_reason": "Recoverable refusal of a valid explicit tag change; does not independently hold landing.",
      "mechanism": "The same new arbitrary-live-tag inference treats absence of every tag on the old commit as movement of the newly declared v2, even though v2 never moved.",
      "test_file": "TASK-260924-4mzun5_review-install-probes-rev4.go",
      "command": "Hosted: go test -count=1 -timeout=5m -v ./internal/install -run '^TestReviewR4ChangedTagAfterOldTagDeleted$'; run 37888437685",
      "repeat-of": "none"
    }
  ],
  "notes": [
    "F4 class recurs, but arbitrary-live-tag inference is a new mechanism compared with revision-2 commit-only comparison.",
    "No independent hybrid/registry/external-build/concurrency adversarial coverage is claimed."
  ],
  "surface_results": [
    {
      "row": "Spec decision and v6 carrier",
      "result": "held"
    },
    {
      "row": "Project/global runtime and GC",
      "result": "held"
    },
    {
      "row": "Effective lock and manifest binding",
      "result": "held"
    },
    {
      "row": "Compiled providers and receipt-3 records",
      "result": "held"
    },
    {
      "row": "Legacy moved tags and repeat installs",
      "result": "broken"
    },
    {
      "row": "CLI status/currentness",
      "result": "held"
    },
    {
      "row": "Audit install/CLI/sourceaudit/cache",
      "result": "held"
    },
    {
      "row": "Dependency directory glob grammar",
      "result": "held"
    },
    {
      "row": "Vendored v1/v2 corpus preservation",
      "result": "held"
    },
    {
      "row": "Generation reads and atomic publication",
      "result": "not-attacked",
      "detail": "Source review and exact-tree full gate; no separate concurrency adversary."
    }
  ],
  "free_hunt": [
    "Lightweight and annotated aliases on the old commit defeat strict same-tag policy; deleting the no-longer-declared old tag yields the opposite false refusal."
  ]
}
```


_Orchestrator note (2026-10-09): in the `verdict-findings` block, `findings[0].severity` was normalised from `"P1"` to `"bypass"` (the reviewer's own `category`) because the board accepts only bypass, regression, robustness or note there; the tb-R226 level P1 is kept as `tb_r226_severity` and in `severity_reason`. Nothing else was changed; the original file is kept by the orchestrator._
