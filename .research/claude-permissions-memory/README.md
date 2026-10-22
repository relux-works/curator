# Claude permissions and automatic-memory qualification

TASK-261022-10f1wx, Claude portion. These are separate measurements on macOS
27.0 (26A428), arm64. Nothing here changes product admission or certifies an
interactive, reload, fork, arbitrary-agent or different-platform entry.

| Runtime | Executable SHA-256 |
| --- | --- |
| Claude Code 2.1.305 | `ea3ea3c08d2a9634a4651e687fcf900d410c489bce2b00f22bd7352b91ba704f` |
| Claude Code 2.1.307 | `5b113a7752119470f1da419cb0e608af3361cae25b614216b13022093958fb2a` |

## Verdict

`qualified` describes only the bounded atom and entries in its row. Unknown
atoms refuse admission. A passing research test can establish a refusal or an
unknown; it does not turn the entire permissions or memory contract green.

| Atom | 2.1.305 | 2.1.307 | Evidence and limits |
| --- | --- | --- | --- |
| Manager deny: Bash | qualified | qualified | Real file-write positive; no file effect and native unavailable-tool result when denied; fresh, safe-origin resume, child. |
| Manager deny: Edit | qualified | qualified | Real Read then Edit positive; original file preserved when denied; same three entries. |
| Manager deny: Write | qualified | qualified | Real file creation positive; absent file when denied; same three entries. |
| Manager deny: MCP tool | qualified | qualified | Approved stdio process starts in both states; inert tool result only when granted; `mcp__*` deny wins over repository allows; same three entries. |
| Manager deny: WebFetch | unknown | unknown | Deny removes the native tool, but the allowed sibling fails external domain-safety verification. No successful fetch positive within the offline boundary. |
| Default builtin ceiling | qualified | qualified | Empty and Read-only `--tools` catalogs reject injected Bash/Edit/Write/WebFetch; Read works only in the Read-only sibling. Fresh print entry only. |
| Project permission files never apply | refused | refused | Both project and local `Read(./ordinary.txt)` denies defeat an explicit manager Read grant while profile and native project instructions still load. |
| Automatic-memory read-off | qualified | qualified | Actual repository-keyed `MEMORY.md` marker reaches model input when enabled and disappears when disabled; fresh, off-origin resume and actual child input. |
| Enabled-origin history scrubbed on resume | refused | refused | Turning the flag off leaves the previously loaded automatic-memory marker in resumed conversation input. |
| Automatic-memory write-off | unknown | unknown | No spontaneous automatic writer was activated. Fixed model-directed Writes into the memory directory succeed with the flag both on and off at all three entries. They do not certify spontaneous writes. |
| Complete native-project permissions and memory-off | refused | refused | Repository permission-source isolation fails; WebFetch and spontaneous automatic writes remain unknown. |

The measured native-project recipe keeps ordinary native project discovery:
no `--setting-sources user`, `--bare` or `--safe-mode`. Every run must contain
both the synthetic managed-home instruction marker and the native checkout
instruction marker. User settings supply `dontAsk` and the manager deny;
project and local settings both challenge it with an allow for the same atom.
Positive siblings explicitly grant the tool. Native denies remove these tools
from the catalog and return `No such tool available`; an empty
`permission_denials` list is therefore expected, not evidence of a bypass.
The MCP recipe uses explicit strict stdio configuration, and its fresh/resume
rows also use an empty builtin catalog. The child recipe adds Read and Task
to drive the native child and retains each tested atom's manager deny.

The source-isolation counterexample is intentionally different: a manager
Read grant succeeds without a repository permission rule and is blocked by
either repository rule. It is sufficient to refuse this recipe's stronger
source-isolation claim. It is not an exhaustive proof that every possible
future native configuration is impossible. The earlier
[source-exclusion record](../project-exclusion-claude/README.md) measures a
different recipe that excludes native project instructions as well.

`CLAUDE_CODE_DISABLE_AUTO_MEMORY=1` suppresses the seeded automatic-memory
index and its native guidance; it is not a filesystem Write policy. Automatic
memory in this replay is distinguished from explicit CLAUDE.md instructions.
Existing memory bytes remain unchanged in the read-only turns, but those
unchanged bytes alone do not prove automatic write-off. A login was neither
needed nor attempted. A successful WebFetch operation and a discriminating
spontaneous automatic-write activation remain follow-up research, not assumed
successes. Other runtimes are outside this split's scope.

## Tools, boundary and replay

Go 1.25.5 drives the real CLI and tools against deterministic local
Anthropic-compatible SSE responses. Python 3 supplies an inert stdio MCP
server, expected-red controls and the publication projection. Git supplies a
synthetic repository boundary. macOS `sandbox-exec` confines every vendor
invocation as `sandbox-exec -> env -i -> exact executable copy`.

The harness and offline boundary controls derive from
`../project-exclusion-claude/replay_test.go` at base
`3fd0ab21550879d7196abffe3e8456a784263970`. The boundary is rechecked here:
only the fixture's loopback provider port is reachable; writes stay in the
fixture; native configuration, Keychain paths/security services and discovery
above the fixture are denied. Synthetic keys satisfy parser requirements;
no real credential, login, external inference, launcher or installation is
used. HOME, native config, four XDG parents and temporary directories are
synthetic. Native Task children receive fixed responses and complete before
the CLI exits; they are behavioral probes, not delegated developers.

Privately copy each versioned executable under task-local `.temp/`. Never run
the updating launcher. The digest and selected version are checked before
creating any fixture. Absent opt-in skips without reading a binary. Use a
fresh absolute run root containing a `.temp/` component on each invocation.
Keep Go caches private and do not inherit Git fixture variables.

```sh
# Ordinary go test ./... omits this hidden research package.
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test \
  ./.research/claude-permissions-memory -count=1 -v

# Set private CLAUDE_QUAL_BINARY to one frozen copy, and private Go caches.
export CLAUDE_QUAL_VERSION=2.1.305  # separately replay 2.1.307
export CLAUDE_QUAL_RUN_ROOT="$PWD/.temp/qualification/replay-305-new"
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test \
  ./.research/claude-permissions-memory -count=1 -v -timeout=9m

# Fresh root. Use the host's shared suite-slot wrapper where required.
export CLAUDE_QUAL_RUN_ROOT="$PWD/.temp/qualification/mutants-305-new"
python3 .research/claude-permissions-memory/mutants.py

python3 .research/claude-permissions-memory/collect.py \
  .temp/qualification/replay-305-new .temp/qualification/observations-305.json
python3 .research/claude-permissions-memory/collect_controls.py \
  "$PWD/.temp/qualification/collector-controls-new"
```

`mutants.py` requires intended exit 1 at the named assertion, no skips or
build failure. Nineteen controls per release remove each atom's manager deny
at fresh/resume/child entry, omit the builtin ceiling, omit the memory-off
flag at fresh entry, or weaken the synthetic network/Keychains boundary.
File/MCP deny and ceiling mutants must also expose the forbidden effect.
The WebFetch mutant qualifies only the catalog checker, not a successful
fetch. The memory control checks the common off-variable gate, not an
independent child-only environment mutation. Boundary mutants run only
synthetic curl/cat/sh controls, never the vendor.

## Evidence

`observations-2.1.305.json` and `observations-2.1.307.json` contain separate
identities, sanitized argv, request digests, marker presence, native tool
errors, catalog/terminal metadata and final synthetic file effects. Digests
bind the captured requests; fresh paths and session IDs can change them on
replay even though the provider responses and expected verdicts are fixed.
Marker flags scan the complete request; the tests assert relevant native
tool results or file effects separately. Child requests are located by the
actual custom-agent system marker because native Task dispatch may interleave
parent and child requests. Resume action indices start after the origin turn.
Missing model requests, missing native init or successful terminal result,
timeouts and crashes fail instead of proving absence.

The collector publishes no full system prompt or raw debug log and rejects
ordinary or flattened home locators after mapping the known synthetic HOME
token. `collect_controls.py` pairs sanitized publication with three private-path
plants and verifies that removing only the filter exposes those same plants.
`evidence/` contains sanitized replay
logs, all control receipts and source digests. Private raw fixtures are removed
after handover; the exact-digest replay remains the reproduction path.
