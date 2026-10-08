# rc.3 × launcher v0.2.0 compatibility smoke

Task: TASK-261004-3s6ymq — launcher-v020-rc3-compat-smoke. Research handoff, 2026-10-04.

## Verdict: GO, within the requested bounded smoke

Released curator v0.15.0-rc.3 interoperates with launcher 2517d2753945d0a8b0c40a291e4aef883a7c16ee at the exercised entry points. No tag was created. Actual native tools were absent from the isolated PATH: those launches are **bounded**, not passing provider integrations. Capture-only substitutes establish launcher argv/environment transport, not native provider behavior. No real model turns, provider installs, ax execution, or real credentials were used.

## Identity and isolation

- Host: Darwin amd64. Scratch root `/tmp/launcher-rc3-smoke` (macOS canonical spelling `/private/tmp/launcher-rc3-smoke`). All runtime HOME values were `/tmp/launcher-rc3-smoke/home`; clean environment, PATH restricted to scratch prefix plus `/usr/bin:/bin`, CI=1 for smoke rows. No inherited credential variables. Build additionally used `/usr/local/bin` for Go, scratch GOPATH/GOCACHE and GOFLAGS=-work.
- Downloaded the public [release](https://github.com/relux-works/curator/releases/tag/v0.15.0-rc.3) tarball and [checksums.txt](https://github.com/relux-works/curator/releases/download/v0.15.0-rc.3/checksums.txt) using credential-free `env -i … curl -fL`; both exits 0. SHA256 verification/extraction exited 0: `8004e8f81d054ffc6a4ac5806b9e619a47c350ba59942c02f5800b8e1406bb7c`. This verifies checksum consistency, not release signature authenticity.
- `go build -work -o /tmp/launcher-rc3-smoke/prefix/bin/curator-run ./cmd/curator-run` exited 0, WORK=/tmp/go-build3599925260. No source changes or reused suite evidence. No full unit/race suite rerun for this research-only task.
- `curator bootstrap --skills-root /tmp/launcher-rc3-smoke/skills --non-interactive` exited 0. Synthetic path profile `smoke`, version 1.0.0, system module `s.md`; profile install/use exited 0. Muse required an existing credential link target; scratch `~/.config/muse/auth.json` contains only `{}`. No operator credentials read or modified. Required source/skill reads and board operations used their normal paths; application/build state remained scratch-only.
- No LOGBOOK.md edited, as explicitly instructed. Findings persist here and in board notes instead.

## Results and bounds

- 4/4 fragment families verified: Claude, Codex and Pi v2; Muse v3 after synthetic credential seeding. Launcher production `run` calls `fragment.Resolve`, which invokes `curator env resolve <env> --profile smoke --repair --format json`; unmodified release output is used for all permission and normal transport rows.
- 4/4 native capture launches exit 0. 3/3 supported yolo captures exit 0, adding exactly one mapped flag: Claude `--dangerously-skip-permissions`, Codex `--dangerously-bypass-approvals-and-sandbox`, Muse `--yolo`. Pi yolo exits 1 with `permission_mode_unsupported`, expected because 0.84.2 has no equivalent bypass flag.
- Muse capture proves inherited HOME preserved and 4/4 XDG parents (`config`, `data`, `state`, `cache`) under its managed home. Capture providers only respond to --version or print argv/environment; they cannot invoke a model. Versions advertised: Claude 2.1.261, Codex 0.153.2, Pi 0.84.2, Muse 1.4.2.
- Headless silence selects native (`source=default-headless`), not a blanket refusal. Explicit headless yolo is permitted with verified transport. This follows [SPEC §4.6](../SPEC.md); interpreting “headless refusals” as rejection of all explicit yolo would contradict the contract. Legacy-fragment yolo refusal was exercised separately.
- Force-native system lock + explicit yolo exits 2; native under the same lock exits 0. Tracked yolo exits 1 (`permission_mode_tracked_unsupported`), with no provider output or ax fallback.
- 5/5 collision shapes refused at the launcher entry: Claude prompt, Codex prompt, Claude MCP, Codex profile, Codex MCP config. Prompt cases use unmodified release fragments. MCP cases use a clearly labeled resolver-boundary mutation adding a synthetic MCP descriptor to the real release response: this tests launcher refusal, **not released Curator MCP package installation**.
- Wrong version and malformed lock digest each exit 1 (`resolve_fragment_invalid`). Digest coverage is syntax (`wrong-digest`), not authenticity of a different well-formed SHA256. Legacy fragment with yolo exits 1 (`permission_policy_unsupported`). These three adversarial rows deliberately mutate the real resolver response; all other fields retain release output.
- Trusted install directory and explicitly configured provider directory both discover launcher and print 0.2.0 (exit 0). 4/4 managed-location probes refuse (exit 1): declared user-bin, published user-bin, global skill bin, environments root. An unpublished `.local/bin` and ordinary `.curator/bin` are not managed directories and were admitted with an outside-trust warning under permissive posture. Initial exploratory labels did not make those paths managed. No gate bypass inferred.
- README install pin @v0.2.0, supported-environments sentence including Muse, and CHANGELOG 0.2.0/Muse description: assertion process exit 0. Source: [README](../README.md), [CHANGELOG](../CHANGELOG.md), frozen at the candidate above. README’s shorthand about ~/.local/bin omits the declared/published distinction; operational advice to avoid it remains sound.
- Ancillary observation: native Muse with no executable exits 1 but its upstream diagnostic says “refusing yolo” despite the explicit native request. Recorded as a misleading missing-release diagnostic, not evidence of yolo execution. Bound accepted by the brief’s missing-native-tool rule.

Standalone Python assertion processes exited 0 for fragment revisions, captured HOME/XDG, refusal codes, collision diagnostics/no capture, and exact-once permission mappings. Each smoke command below ran as its own subprocess with its real return code; no pipeline/tee or masking. Expected negatives remain nonzero, never represented as successful commands. Harness process exit 0 alone is not used as a row result.

## Every smoke command and real exit code

`run.py` supplies the clean environment/cwd described above. Additional environment mutations and fixture construction are preserved in the attached harness. Full stdout/stderr are in the task-scoped rows.jsonl outcome.

| Row | Exit | Command |
|---|---:|---|
| curator-version | 0 | `curator --version` |
| launcher-version | 0 | `curator-run --version` |
| resolve-claude_code | 0 | `curator env resolve claude_code --profile smoke --repair --format json` |
| native-claude_code | 1 | `curator-run claude_code --profile smoke --permissions native -- --help` |
| yolo-claude_code | 1 | `curator-run claude_code --profile smoke --permissions yolo -- --help` |
| resolve-codex_cli | 0 | `curator env resolve codex_cli --profile smoke --repair --format json` |
| native-codex_cli | 1 | `curator-run codex_cli --profile smoke --permissions native -- --help` |
| yolo-codex_cli | 1 | `curator-run codex_cli --profile smoke --permissions yolo -- --help` |
| resolve-pi | 0 | `curator env resolve pi --profile smoke --repair --format json` |
| native-pi | 1 | `curator-run pi --profile smoke --permissions native -- --help` |
| yolo-pi | 1 | `curator-run pi --profile smoke --permissions yolo -- --help` |
| resolve-muse | 1 | `curator env resolve muse --profile smoke --repair --format json` |
| native-muse | 1 | `curator-run muse --profile smoke --permissions native -- --help` |
| yolo-muse | 1 | `curator-run muse --profile smoke --permissions yolo -- --help` |
| headless-default | 1 | `curator-run codex_cli --profile smoke -- --help` |
| trusted-discovery | 0 | `curator run --version` |
| resolve-muse-seeded | 0 | `curator env resolve muse --profile smoke --repair --format json` |
| native-muse-absent | 1 | `curator-run muse --profile smoke --permissions native -- --help` |
| capture-claude_code-native | 0 | `curator-run claude_code --profile smoke --permissions native -- --help` |
| capture-claude_code-yolo | 0 | `curator-run claude_code --profile smoke --permissions yolo -- --help` |
| capture-codex_cli-native | 0 | `curator-run codex_cli --profile smoke --permissions native -- --help` |
| capture-codex_cli-yolo | 0 | `curator-run codex_cli --profile smoke --permissions yolo -- --help` |
| capture-pi-native | 0 | `curator-run pi --profile smoke --permissions native -- --help` |
| capture-pi-yolo | 1 | `curator-run pi --profile smoke --permissions yolo -- --help` |
| capture-muse-native | 0 | `curator-run muse --profile smoke --permissions native -- --help` |
| capture-muse-yolo | 0 | `curator-run muse --profile smoke --permissions yolo -- --help` |
| capture-headless | 0 | `curator-run codex_cli --profile smoke -- --help` |
| locked-yolo | 2 | `curator-run codex_cli --profile smoke --yolo -- --help` |
| locked-native | 0 | `curator-run codex_cli --profile smoke --permissions native -- --help` |
| tracked-yolo | 1 | `curator-run codex_cli --profile smoke --yolo -- --help` |
| claude-prompt-collision | 1 | `curator-run claude_code --profile smoke --system-prompt append -- --append-system-prompt collision --help` |
| codex-prompt-collision | 1 | `curator-run codex_cli --profile smoke --system-prompt replace -- -c 'model_instructions_file="/tmp/collision"' --help` |
| discovery-user-bin | 0 | `/tmp/launcher-rc3-smoke/prefix/bin/curator run --version` |
| discovery-environments | 1 | `/tmp/launcher-rc3-smoke/prefix/bin/curator run --version` |
| discovery-skill-bin | 0 | `/tmp/launcher-rc3-smoke/prefix/bin/curator run --version` |
| discovery-explicit-shim | 1 | `/tmp/launcher-rc3-smoke/prefix/bin/curator run --version` |
| discovery-published-shim | 1 | `/tmp/launcher-rc3-smoke/prefix/bin/curator run --version` |
| discovery-global-skill-bin | 1 | `/tmp/launcher-rc3-smoke/prefix/bin/curator run --version` |
| mutated-version | 1 | `curator-run codex_cli --profile smoke --yolo -- --help` |
| mutated-digest | 1 | `curator-run codex_cli --profile smoke --yolo -- --help` |
| mutated-legacy | 1 | `curator-run codex_cli --profile smoke --yolo -- --help` |
| mutated-mcp-claude_code---mcp-config | 1 | `curator-run claude_code --profile smoke -- --help --mcp-config other.json` |
| mutated-mcp-codex_cli---profile | 1 | `curator-run codex_cli --profile smoke -- --help --profile other` |
| mutated-mcp-codex_cli---config=mcp_servers.other.command="other" | 1 | `curator-run codex_cli --profile smoke -- --help '--config=mcp_servers.other.command="other"'` |
| configured-trusted-provider | 0 | `/tmp/launcher-rc3-smoke/prefix/bin/curator run --version` |

## Exploratory command exits

`curator env --help` 0; `curator --help` 0; `curator-run --version` 0. `curator bootstrap --help` 2 (its help convention). Before bootstrap, `curator env resolve --help` and `curator profile install --help` each exited 1 for absent scratch global config. After bootstrap both exited 2 and printed usage. These are setup/help observations, not passing compatibility gates. Git identity/status inspection exited 0. File-search misses during source exploration were not validation gates.

## Review handoff

GO is limited to the stated compatibility smoke, with absent real tools explicitly bounded. Native provider integration and real MCP package installation are outside the measured result; no claim is made for credentials, model turns, or tracking success. All requested refusal categories were reached at the executable boundary. Findings are ready for review.
