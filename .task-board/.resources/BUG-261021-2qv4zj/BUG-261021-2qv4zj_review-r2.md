# BUG-261021-2qv4zj — review r2

Verdict: **changes_requested**. The previous F1 case-alias bypass is fixed; independent free hunt found blocking F2 in the non-skill profile context publication path.

Reviewed head: `9b17768c96d53e9ef7babb45e95d3b688728e815`.
Diff base: `2c89fd46443d461e25bd377fb94a0c6754282838`.
All three introduced commits (`35f8a18d50795636e342baf87abc491bf7ce1aca`, `2c01da83de4007c5d6f6d4edf3a887dc85707795`, `9b17768c96d53e9ef7babb45e95d3b688728e815`) verified with signature status `G`. Review used a detached worktree and a separate normal-.git validation clone. Board reviewer assignment and all five preconditions were read; review acknowledged in `audit-md-scope`.

## Surface sweep

| Surface | Held or broken | Attack / production entry | Evidence |
|---|---|---|---|
| Previous F1: nine root aliases | Held | Actual CLI `profile install` for SKILL.MD, References, AGENTS, ASSETS, .SKILL_TRIGGERS, Templates, Examples, Data, Scripts; strict/high refusals with publication-preservation assertions; canonical, noise and declared-host controls | All nine CLI cases passed on case-insensitive APFS; focused gates also passed |
| Skill roots and repository noise | Held | Gate/GateReadOnly and CLI with visible undeclared contact, declared contact, .research noise; every context root; scripts with/without exports; validated runtime/build roots | Focused scope suite and CLI scope suite exit 0 |
| All fixed exclusions and case neighbors | Held | CopyContext plus gates for all 19 exclusion patterns at depth, contrasting Tests/readme.md/ReadMe.MD and other case variants; repeat on sensitive APFS | Both filesystem runs exit 0; exact protocol spelling excludes, visible neighboring spelling is copied and audited |
| Unicode and mixed nested names | Held | Independent overlay test through CopyContext, GateReadOnly and Gate: references/Sub/x.md, NFC/NFD cafe filenames, sharp-s, dotted/dotless i, long-s root and uppercase extension; declared-host controls | Both filesystem runs exit 0; every published contact refused. Long-s scripts alias is published/refused on insensitive APFS, not published on sensitive APFS |
| Symlink roots and dual roots | Held within tested boundary | Independent raw CopyContext and gates with symlink directory/file roots; canonical references plus distinct References on case-sensitive APFS | Directory link gives copy/read error, file-link contact is refused; sensitive dual-root case selects only canonical root. Production selected-directory symlink rejection cases also passed |
| Canonical published bytes | Held | Independent old CopyContext extracted from base, compiled only as overlay helper; compare complete returned file list and bytes against head, with scripts on/off, nested exclusions/runtime/build roots | Byte-for-byte and list equality on both filesystem types |
| Single publication selector across subject kinds | **Broken (F2)** | Independent built CLI installs a context profile with declared context/a.md URL; compare clean control and base-detector overlay | Head publishes URL into native CLAUDE.md and AGENTS.md and saves empty ruleset-7 allowance; old detector refuses. Context materializer does not use the skill walk |
| Non-Markdown scripts and manifests | Held | Independent gates on scripts/tool.sh, scripts/tests/contact.txt, scripts/.git/contact.txt, scripts/README.txt with undeclared and declared controls; focused canonical/legacy manifest detector cases | Independent probe and focused tests exit 0; original script/manifest scan retained |
| Cache namespace and real historical allowance | Held for skill cases; F2 creates new stale allowances | Actual ruleset-6 detector/cache writer, both hash framings, read-only/writable gates, old record preservation, new refusal caching, clean and declared reuse; ruleset-5 refusal and ruleset-4 allowance coverage | Focused suites exit 0; real-6 allowances do not bypass head; F2 fixture independently persists ruleset-7 findings=[] |
| Harness controls | Held | Restored lexical matcher, every-Markdown, dropped references, scripts always/never, namespace-6/5/4 reuse and other narrowing mutants | Harness exits 0 after intended assertion failures; no skipped/non-applied plant accepted as evidence |
| CIP-0021 witness and docs | Held for stated skill scope | Lifecycle suite materializes historical notes.md witness as references/notes.md; declared-host control, schema/identity/metadata and narrowing cases; inspect vendored-corpus diff | CLI/lifecycle focused run exit 0; corpus unchanged; docs accurately describe skill walk/cache behavior, but do not cover F2 |
| Pinned build/runtime context conformance | Held | Three relevant tests using spec pin 43bf0a2506d5c354a73bbc3ea4623d4653db10c7, conformance/v1 | All three executed and passed; initial invocation used parent directory and failed for missing fixture (setup error), corrected invocation exit 0 |
| Full default suite / build / vet | Held | Normal-.git exact-head clone; no opt-in conformance env; private caches, short TMPDIR | 85/85 packages accounted for: 80 pass, 5 no-test packages; both suite commands exit 0. Build and vet exit 0. |
| Public-repo hygiene | Held | Inspect complete diff and added-line scan for personal paths, host/user identifiers and client/employer names; git diff --check | No finding; diff check exit 0 |

## F2 reproduction and causal control

Create a fresh protected fixture directory (directories 0700, files 0600) with:

- `profile/agent-context.json`: `{"schema_version":1,"name":"review-context","version":"1.0.0","context":{"modules":[{"path":"a.md"}]}}`
- `profile/context/a.md`: `Read https://api.example.net/guide` followed by LF.
- An isolated manager config with schema_version 1, private skills_root, empty projects, and audit enabled/strict/high/null; an empty schema_version-1 system config.
- Point HOME, USERPROFILE, CLAUDE_CONFIG_DIR, CODEX_HOME, XDG_CONFIG_HOME, PI_CODING_AGENT_DIR, CURATOR_CONFIG and CURATOR_SYSTEM_CONFIG only at that fixture; unset GIT_DIR/GIT_WORK_TREE. Run the exact-head binary as `curator profile install <absolute-fixture-profile>`.

| Detector / fixture | Exit | Native URL in CLAUDE.md and AGENTS.md | Network refusal |
|---|---:|---|---|
| Head / clean module | 0 | No | No |
| Head / URL module | 0 | **Yes, both** | **No** |
| Base audit.go overlay / clean module | 0 | No | No |
| Base audit.go overlay / URL module | 1 | No | Yes, audit.capability.network-undeclared |

The base control restores only `internal/audit/audit.go` using `go build -overlay`; it is not claimed to be a complete base checkout. Both binary builds exited 0. Head URL installation also created `profiles/review-context/source.json`, `profiles/review-context/lock.json` and a `verdict-null--1-7.json` record with `findings: []`. The copied context store contains the URL module. No external network call is needed for this deterministic contact detector.

Causal chain: `internal/envprofile/envprofile.go:1756` sets implicit-none for context members; its audit Subject reaches Gate at line 1791. The unconditional skill walk at `internal/audit/audit.go:871` never visits context/a.md. `internal/contextmaterialize/contextmaterialize.go:241` independently appends declared module bytes to the native root-context document. Environments section 9.1 at the pinned spec requires the unchanged strict source audit for every closure member. The rework review brief explicitly requires sweeping other materializers rather than assuming CopyContext is the sole publisher.

## Validation details

Platform: Darwin arm64, Go 1.25.5.

All Go runs used `env -u GIT_DIR -u GIT_WORK_TREE`, a private GOCACHE/GOMODCACHE/GOPATH and private HOME. Mutation helpers inherited those private caches. Normal suite TMPDIR was the short lane `.t`; sensitive-volume fixture TMPDIR was the task-local APFS mount. No repository source was edited; independent probes and base controls used scratch overlays.

- `go test -v -count=1 -timeout=15m ./internal/audit ./internal/whitelist -run 'TestMarkdownScope|TestReviewPriorMarkdownAllowance|TestNetworkContactNarrowingMutants|TestContextPathIncluded'`: exit 0.
- `go test -v -count=1 -timeout=15m ./internal/envprofile -run 'TestProfileInstallMarkdownCaseAliasesCLI|TestProfileInstallMarkdownScope|TestSkillDirectory'`: exit 0.
- Independent overlay `TestReviewer*` on insensitive and sensitive APFS: exit 0; sensitive dual-root test executed (not skipped). Additional independent non-Markdown probe: exit 0.
- `go test -count=1 -v -run TestMarkdownScope ./internal/audit` on sensitive APFS: exit 0.
- `go test -v -count=1 ./internal/whitelist ./internal/skillcheck -run TestBuildRoot` with corrected pinned conformance/v1: exit 0.
- `go build ./...`, `go vet ./...`, binary build: each exit 0.
- Full default suite split into `go test -json -count=1 -timeout=60m ./cmd/curator` (exit 0, 867.30s), then the same command over every other package from `go list ./...` (exit 0, 736.29s). This exhaustively covered all 85 listed packages: 80 passed and 5 had no test files. No package was omitted and no conformance environment was set for this default run.

The case-sensitive image was detached after its checks. Task-local clones, worktree, fixtures, binaries and all private caches are removed after handover; this report retains the review result and reproducible evidence.

```json
{
  "verdict": "changes_requested",
  "head": "9b17768c96d53e9ef7babb45e95d3b688728e815",
  "findings": [
    {
      "id": "F2",
      "severity": "P1",
      "file_line": "internal/audit/audit.go:871",
      "mechanism": "The shared skill CopyContext walk is applied unconditionally to all audit subjects. strictAuditMember also audits KindContext members through this detector with implicit-none capabilities, but their declared context/*.md modules are published by contextmaterialize, outside the skill roots. The new walk silently omits those modules, permits installation, and caches an empty ruleset-7 verdict.",
      "reproduction": "Build exact head; in a fresh isolated home use a protected path profile whose agent-context.json declares schema_version=1, name=review-context, version=1.0.0, context.modules=[{path:a.md}]; put Read https://api.example.net/guide in context/a.md. Run curator profile install <absolute-profile-path> with strict/high/null audit. It exits 0, publishes profiles/review-context/source.json and lock.json, and emits the URL into both CLAUDE.md and AGENTS.md. The clean fixture also succeeds. With only audit.go restored through a Go overlay from base 2c89fd46443d461e25bd377fb94a0c6754282838, the URL fixture exits 1 with audit.capability.network-undeclared and neither native document is published; the clean fixture still succeeds.",
      "required_fix": "Select Markdown admission by the actual subject/publication contract. Retain skill context exclusions for skills while auditing the declared prompt modules of context members (and account for other non-skill subjects). Add a real profile-install regression with clean positive control, exact network refusal and preservation assertions. Invalidate ruleset-7 allowances for affected snapshots; do not reuse the empty verdict created by this head."
    }
  ]
}
```
