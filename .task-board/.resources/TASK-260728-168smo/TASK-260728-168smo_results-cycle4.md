# TASK-260728-168smo — rework cycle 4 results

Directive: review cycle 3 (`RUN-260729-a93356`) requested changes on exactly two
blockers. Both are closed. No measurement was re-run, because neither finding
disputed a measurement — both disputed what the documents *said about* the
measurements. Nothing measured changed, and nothing measured was reinterpreted.

## Verdict on the directive

Two document defects, two corrections, and both corrections move a claim **down**
to the evidence rather than moving evidence up to a claim.

## Blocker 1 — two incompatible Windows dynamic-library allow-lists

**Finding.** Reference section 10.1 fixed the measured `windows/amd64`
base-installation allow-list at `{KERNEL32.dll, msvcrt.dll}` and explicitly
excluded `ADVAPI32.dll` and `USER32.dll`. Section 11.6's K-9 row carried
`{KERNEL32.dll, msvcrt.dll, ADVAPI32.dll, USER32.dll}`. An implementer had two
closed sets and no rule to choose between them, and choosing the wider one would
admit unmeasured behaviour through the published-artifact gate.

**Closed by normalizing down to the measured set, everywhere.** Only the W14
record and the two-entry set are supported by evidence; both compile samples —
the plain program and the one importing `platform.posix` and `platform.windows`
— produced exactly those two imports and nothing else. No new A6 run was made,
so no widening is justified.

| Location | Before | After |
|---|---|---|
| reference 10.1 (normative table) | `{KERNEL32.dll, msvcrt.dll}` | unchanged, **and now declared the single normative source** |
| reference 11.6 K-9 | four entries | `KERNEL32.dll`, `msvcrt.dll` — pointer to 10.1 |
| reference 11.1 A6 verdict | "imports within the 10.1 allow-list" | names the two entries explicitly |
| reference 14, artifact gate | 4 cases, allow-list unnamed | 5 cases, fixture named as the two-entry set, with `ADVAPI32.dll` and `USER32.dll` as required **rejection** cases |
| decision 0010 section 9 | pointer to 10.1 only | states the two-entry set inline, names the excluded plausible set, restates the re-qualification rule |
| decision 0010 rejected alternatives | — | new entry rejecting the widened list, with the reason the contradiction was dangerous |

The four-entry row was a drafting error, not a second measurement. Cycle 3's own
results document already said the base-installation allow-list "is what was
MEASURED, not what is plausible" and listed `USER32`/`ADVAPI32` as *excluded*;
the obligation table simply failed to follow. Making the two named DLLs into
required negative vector cases means the corpus now fails if a future revision
re-admits them silently.

## Blocker 2 — qualification language exceeded the persisted evidence

**Finding.** Section 11.1 (reference) and section 12 (decision) both admit a
tuple only on **all** of A1–A9. Both then recorded A1–A7 as passed, assigned A8
to `TASK-260728-251p01` and made A9 conditional — and then called the tuple
qualified, filled `platforms` and `compatibility` as shipped values, and declared
the platform retirement branch closed. That is self-contradictory on its own
terms.

**Closed by the directive's preferred scope-preserving resolution.** The reviewer
offered two options: run A8/A9 now and keep the claim, or state the candidate
and defer. A8 is the allow-list classifier corpus walk and A9 is the
2.4-family conformance result; both are authored by `TASK-260728-251p01` and
neither exists yet, so producing them here would mean authoring another task's
vectors. The claim moves instead.

`windows/amd64` is now stated everywhere as an **A1–A7 host-qualified
candidate**, a term defined once in the reference header and used consistently:
host-side requirements discharged with recorded argv and real exit codes,
corpus-side requirements not; never admitted.

Concretely:

- **Registry entry** (reference 1.4, decision 0010 section 10). `platforms` and
  `compatibility` are labelled **candidate sets**. Both are covered by one
  binding rule instead of the previous rule that covered only `compatibility`:
  they ship as `{(windows, amd64)}` and `{(2, 4)}` **only** in the change where
  the section 14 corpus passes with A8 included; otherwise the entry has no
  admissible tuple and no admissible family. Every other field keeps its measured
  value — the entry is still complete in the sense decision 0007 section 1.3
  requires; nothing regressed to empty.
- **Acceptance table** (reference 11.1). A8 and A9 read **NOT DISCHARGED** with
  the owner named, replacing "owned by the vectors" and "`{(2,4)}` declared".
  The admission rule now states the candidate state explicitly so the table and
  its own rule agree.
- **11.3** retitled *Windows — A1–A7 host-qualified candidate*.
- **11.5** retitled *Retirement — both branches are open*. The platform branch
  is **re-armed**: cycle 3 disarmed it on an A1–A7 result, which its own A1–A9
  rule forbids. Both branches now resolve at one event, the schema-8 change, and
  resolve the same way.
- **Decision 0010** revision note rewritten to revision 4, stating exactly what
  revision 3 got wrong and confirming that every A1–A7 measurement, the bundle
  model, the ETW trace, `PATH` and offline evidence, artifact evidence and the
  macOS exclusion are retained unchanged.
- **Downstream obligations** now say `TASK-260728-251p01` *owns admission*, not
  merely the vectors, and that its change is what converts the candidate sets.
- **Rejected alternatives** gains "Declaring the tuple admitted on the A1–A7 host
  result", so the option is recorded as considered and refused rather than
  quietly dropped. The revision-2 alternative is re-worded to keep the real
  distinction: revision 2 had *no* values and no path to any; this record has
  measured candidate values and a fixed admission event.
- Incidental phrasing swept: "the qualified tuple" → "the candidate tuple" at
  six sites in the reference and two in the decision, "the only tuple in
  `platforms`" → the only tuple that *may enter* it, "the only claimed tuple" →
  "the only claimable tuple", capability limitation 5 in both documents.

## What did not change

Everything the reviewer passed. The Kotlin/Native selection and the JVM
deferral; the paired identifiers; `curator-kotlin-bundle-v1`, its digest and its
reproducible curation; direct `jdk\bin\java.exe` launch; the P1/P2 probes and
their grammars; the source allow-list, ordered rejection matrix and `@`-token
rule; the operation-private overlay; cache, receipt and marker identity;
local/external equivalence; the ETW process closure and its control; the
artifact and its PE class; the permanent macOS exclusion with its
union-7/sufficient-6 correction; the honest Linux deferral. No product code, no
schema, no vector, no release file, no staging, no commit, no publication, no
pin, no host install, no platform widening.

## Gates

Full transcript in `TASK-260728-168smo_gate-log-cycle4.txt`. Real exit codes:

| Gate | Exit |
|---|---|
| curator-spec `validate.py`, baseline 57c1f56 | 0 — 30 schemas, 93 vectors |
| curator-spec `validate.py`, task worktree | 1 **EXPECTED-RED**, 1 broken link, 0 from this task |
| scoped link sweep over the two authored documents | 0 — 6 links, 0 broken |
| curator-spec `unittest discover -s tools` | 0 — 8 tests |
| curator-spec `go test ./tools/...` | 0 |
| curator `go build ./...` | 0 |
| curator `go vet ./...` | 0 |
| curator `gofmt -l ./cmd ./internal` | 0 — 0 files |
| curator `go test ./...` | 0 — 31 packages |
| curator `make check` | 2 **EXPECTED-RED**, gofmt stage, 754 files listed, 0 from this task |
| `golangci-lint` | NOT RUN — not installed on this host |

Both expected-red gates are the same pre-existing ones cycle 3 attributed, and
both were re-verified this cycle rather than carried forward on assertion: the
`validate.py` failure is `docs/external-build-repositories.md` line 146 linking
to `../release/1.0.0-rc.5.json`, an untracked in-flight sibling document another
task copied into this worktree, and the same command exits 0 on the clean
baseline checkout; the `make check` failure lists 754 unformatted Go files and
**zero** of them are outside another task's `.temp/` scratch tree.

## Scope

Two Markdown documents edited, both to revision 4. No Go, schema, vector or
release file touched. No board element created, retired or re-linked. No host
touched — this cycle ran entirely on the gate host; the Windows host was not
needed because no measurement was in dispute.
