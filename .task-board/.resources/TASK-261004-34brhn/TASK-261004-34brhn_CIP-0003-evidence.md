# CIP-0003 evidence: Claude managed-home credential modes

For **TASK-261004-34brhn — research-claude-login-transfer-modes**, within **STORY-261004-1pwxri — claude-macos-managed-home-login-modes**. Collected 2026-10-04. Read with the [CIP draft](261004_CIP-0003-claude-managed-home-credential-modes.md).

## Scope and confidence

This run inspected public sources and the installed executable, then ran bounded CLI probes with synthetic data only. No operator credential file was opened; no real Keychain secret was read or written; no account login, logout, issuance, revocation or model request was attempted. Product source and LOGBOOK.md are outside the change set. The task-specific prohibition on LOGBOOK.md overrides the generic role checklist; findings are recorded here and in the task note.

**Measured** means CLI exit/status and mocked service selection in a protected scratch HOME. **Inspected/documented** means code branches in the exact executable or vendor documentation. **Unknown** includes real account authentication, backend writes, refresh concurrency and safe sharing. Q1 remains **0/3** required disposable-account facts established by this run.

## Pinned inputs

| Input | Identity |
| --- | --- |
| Curator main | `ca1b776fb580ec0cee0173bf150daf063023aeaa`; worktree HEAD equal; initially clean. Public remote advertisement agrees, exit **0**. |
| curator-spec rc.14 | Tag object `661bead088186db70c9f57ad0b115305ec2bef35`, peeled commit `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`; local and remote agree, exit **0**. No tag-signature attestation is made. |
| Launcher main | Fresh advertised OID `2517d2753945d0a8b0c40a291e4aef883a7c16ee`, exit **0**. Local main was older (`d0920353556dd0a3cea2864c616c230915977bc7`); changed files were inspected at the fresh OID and citations updated. No branch/ref was changed. |
| Claude executable | Native macOS **2.1.287**, 236,282,368 bytes; SHA-256 `f1863213e4f55aaadc2e6ee617f934ada29930e5de4c5e5f8f9e34a0d594fdd7`; embedded build `2026-10-01T16:02:06Z`, commit `3c446a1b98aceb99a6cdee0f84a8bea42f4a8937`. `--version` independently reported 2.1.287, exit **0**. This is not a globally latest-release claim. |
| Probe platform | macOS 15.7.4, x86_64; executable invoked by its resolved version path, with auto-update disabled. |

The three `git ls-remote` processes used scratch HOME, minimal PATH, disabled global/system Git config, prompts and credential helpers. Four public raw launcher file fetches each exited **0**. Composition, plan and main-entry files differed from the older local branch and were re-inspected; execution.go was byte-identical. No product test was repeated.

## Repository source map

Paths below are relative to the named repository and pinned commit above. These are file:line citations, not citations to a changing branch name.

| ID | Citation and checked claim |
| --- | --- |
| C1 | curator `internal/config/environments.go:20`, `:799`: current config struct and closed shared/isolated parser. |
| C2 | curator `internal/envregistry/envregistry.go:240`, `:450`: Claude passthrough, 2.1.261 pin, defaults/refusals. Unknown versions currently warn and use at-or-above behavior; this does not qualify every later build. |
| C3 | curator `internal/envprofile/managed.go:319`, `:563`, `:598`: production resolution, passthrough selection and unconditional macOS isolated `keychain` marker metadata. |
| C4 | curator `internal/envfragment/envfragment.go:1`, `:54`: closed fragment contract, ordinary v2/Muse v3, no credential-source field. |
| C5 | curator `internal/envprofile/managed.go:1504`, `internal/envprofile/migrate.go:1`: credential-link preservation and explicit metadata-only migration. |
| S1 | rc.14 `protocol/environments.md:1426`, §7.4: passthrough, Linux refresh confidence, link preservation and macOS sharing refusal. `:25` separates verified tool facts from source/documentation confidence. |
| S2 | rc.14 `decisions/0017-environment-credential-modes.md:70`, `:168`: historical 2.1.273 observation and Q1–Q7. Accepted only as history, not replayed or promoted to current-build evidence. |
| S3 | rc.14 `profiles/manager.md:2847`, §12.4: manager credential boundary; nearby lines 2838–2857 cover explicit migration, marker publication and backups. |
| S4 | rc.14 `protocol/environments.md:3672`, `:3758`: knobs, both isolation-lock directions and no source-selecting lock. `schemas/v1/manager-config-v3.schema.json:1`, `launch-env-fragment-v3.schema.json:1`, and `agent-environment-marker-v3.schema.json:1` exist. |
| L1 | launcher `internal/composition/composition.go:35`, `:75`: full env excluded from JSON, literals serialized; `internal/execution/execution.go:117`, `:142`: direct/tracked split; `internal/plan/plan.go:219` and `cmd/curator-run/main.go:255`: production call chain. This product path was inspected, not executed. |

Public immutable entry points: [Curator registry](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/envregistry/envregistry.go#L240), [marker construction](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/envprofile/managed.go#L598), [rc.14 environments](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/environments.md#L1426), [Decision 0017](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/decisions/0017-environment-credential-modes.md#L168), [launcher execution](https://github.com/relux-works/curator-agent-launcher/blob/2517d2753945d0a8b0c40a291e4aef883a7c16ee/internal/execution/execution.go#L142).

## Vendor documentation

Primary Anthropic sources retrieved on the research date:

| ID | Source and locator |
| --- | --- |
| D1 | [Authentication](https://code.claude.com/docs/en/authentication): multiple accounts, storage, precedence, long-lived tokens and profiles; browser extraction lines 45–50 and 214–291. |
| D2 | [Environment variables](https://code.claude.com/docs/en/env-vars): OAuth token/provisioning variables and helper TTL. No secure-storage override entry was found on that page; this is not proof of universal lack of documentation. |
| D3 | [apiKeyHelper](https://code.claude.com/docs/en/settings-reference#apikeyhelper): shell command, output/header contract, cache/retry/trust. Public Markdown fetched by curl, exit **0**, section lines 5529–5551, page SHA-256 `76582550507f251beaa3688cfb35ad09b0241935d4a16b11b3a09077d19a7e5c`. Web rendering failed on page size; an earlier Python fetch failed HTTP 403, exit **1**, and supplied no evidence. |
| D4 | [Settings precedence](https://code.claude.com/docs/en/settings) and [CLI reference](https://code.claude.com/docs/en/cli-reference): scopes, reload, flags and setup-token entry. These identify integration test surfaces, not implemented Curator behavior. |

## Shipped executable evidence

These are zero-based byte offsets in the full executable, reproducible against the SHA-256 above. Minified names are version-specific. The evidence is the code branch, not mere string presence. No vendor corpus is attached.

| ID | Offset / locator | Inspected behavior and limit |
| --- | --- | --- |
| B1 | `187482079` (`yb`), `187482214` (`jF`), module `chunk-51mtffqy.js` | Ordinary storage follows the config home; the service has an 8-hex SHA-256 suffix. A present secure-storage override changes the directory input; explicitly empty selects default `.claude` storage and the unsuffixed service. Account selection uses validated USER/name fallback. |
| B2 | `187486784` (`x`), `187491797` (plaintext backend), `187495450` vicinity (`zn`) | Keychain read precedes file read. Strict read failures can prevent mutation; transient primary failure can skip fallback. Other failed primary writes may use plaintext; transitions may remove the prior fallback/primary. File path is `.credentials.json`, mode 0600. An alternate storage interface recognizes symlink-refusal and read-failure states. Not every Keychain error causes a file write. |
| B3 | `188284411` (`fK`) | Env OAuth produces `refreshToken:null`, `expiresAt:null`, and default inference scope unless other runtime scope metadata exists. It has no intrinsic refresh/expiry awareness. |
| B4 | `188290651` (`mK`); diagnostic near `217268526` | Ordinary user env-token 401 handling retains that source instead of adopting a saved login. SDK callback/remote-host rotation are different branches, not supplied by simple launcher injection. |
| B5 | `188247265` (TTL), `188261235` vicinity (`ens`), `188262466` (`ak`), `188263143` (`zH`) | Default cache 300,000 ms; native helper timeout 600,000 ms; stdout is trimmed/validated, and failures can include 500 characters of stderr. Background refresh failure can retain the cached value. The CIP's 30-second bridge and stderr containment are proposals. |
| B6 | `188297126` (`m8`/`gl`), `188280617` (`WQn`), `188282149` (`zQn`) | Stored-login renewal checks expiry/refresh state, obtains a storage-root lock, rereads, requests renewal and saves through a compare-and-swap-style mutation. Inference-only results are not persisted by that save path. Real refresh and cross-home coordination remain unmeasured. |
| B7 | `185907702` (`kn`/`qse`); imported by plaintext backend from `chunk-kyfj46cz.js` | Normal writer stages then renames, with conditional in-place fallback. Writer-module start `185899594`, length 12,449, SHA-256 `bf75e99a6912ca02018c584cfcacb991f5bffc982830fbd393f94d8fe9c3f406`. This cannot qualify symlink refresh safety. |

Locator reproduction reads executable bytes only: resolve `command -v claude`, read that executable, verify the full SHA-256, then inspect bounded windows at these offsets. For example, offset `188284411` starts with `function fK(){if(a.CLAUDE_CODE_OAUTH_TOKEN)`. Do not execute extracted source. The override and fallback make the historical file observation explicable; they do not establish its actual cause.

## Scratch probe setup

The scratch tree was owner-only. Its concrete path is replaced by `$SCRATCH` in this record. Every CLI invocation was a separate foreground subprocess with a 25-second limit; its wrapper returned the same exit code. No pipe/tee obscured status and no model prompt was run.

- Child env was a whitelist: scratch HOME, CLAUDE_CONFIG_DIR, TMPDIR, synthetic USER/LOGNAME, minimal PATH starting with scratch shims, locale/terminal and update/telemetry disable flags. No inherited auth, provider, profile, proxy, host or FD variables were forwarded.
- The macOS sandbox denied network, real user-home reads/writes, OS Keychain locations, managed Claude settings locations, real `/usr/bin/security` execution and Keychain daemon IPC. Only the resolved installed executable was excepted from the user-tree read denial.
- A PATH `security` shim recorded command/service names, returned not-found exit 44 by default, or supplied a synthetic scratch object. It never forwarded to the real executable. Unexpected write operations returned 99; none was requested.
- Synthetic file credentials carried `subscriptionType:max`; mock Keychain credentials carried `subscriptionType:pro`. Identity fields were null. A fake helper kept an invocation marker. Distinct labels identified selection without comparing/exporting secrets.
- Each matrix row reset only its owned scratch fixtures. P14 deliberately relaunched without recreating P13's input. Other independent rows do not establish persistence.

Normalized commands were `python3 $SCRATCH/probe.py --version`, `python3 $SCRATCH/probe.py setup-token --help`, `python3 $SCRATCH/probe.py auth status --json`, and `python3 $SCRATCH/matrix.py <case>`. Each invoked `/usr/bin/sandbox-exec -f $SCRATCH/probe.sb <resolved-claude-binary> <argv>` from an empty scratch working directory.

The synthetic account name was `credential-probe`. Ordinary service suffix `296b436d` equaled SHA-256(config-directory-string)[:8] for this scratch path; alternate storage yielded `283f76fe`; an empty override yielded no suffix. Reproduction under another scratch path should change the hash.

### Raw measured results (historical; original harness unavailable)

The status fields below are literal CLI outputs. Helper invocation is separately observed through the scratch marker.

| ID | Command/case | Real exit | Output / observation |
| --- | --- | --- | --- |
| P1 | `probe.py --version` | **0** | `2.1.287 (Claude Code)`, empty stderr. |
| P2 | `probe.py setup-token --help` | **0** | Long-lived token help, subscription required; no expiry flag shown; no issuance. |
| P3 | Empty `probe.py auth status --json` | **1** | `loggedIn:false`, `authMethod:none`, `apiProvider:firstParty`. Expected failure: no scratch credential. |
| P4 | `matrix.py token` | **0** | `loggedIn:true`, `authMethod:oauth_token`; no credential file created. |
| P5 | `matrix.py token-api` | **0** | `authMethod:api_key`, `apiKeySource:ANTHROPIC_API_KEY`. |
| P6 | `matrix.py token-helper` | **0** | `authMethod:api_key_helper`, `apiKeySource:apiKeyHelper`; helper marker **absent**. |
| P7 | `matrix.py file` | **0** | `authMethod:claude.ai`, `subscriptionType:max`; mock Keychain not-found. |
| P8 | `matrix.py keychain-file` | **0** | `authMethod:claude.ai`, `subscriptionType:pro`; both sources present, Keychain label wins. |
| P9 | `matrix.py keychain` | **0** | `authMethod:claude.ai`, `subscriptionType:pro`; no credential file. |
| P10 | `matrix.py none` | **1** | `loggedIn:false`, `authMethod:none`. Expected failure after synthetic sources were removed. |
| P11 | `matrix.py storage-override` | **1** | Expected no-login failure. Config still `$SCRATCH/config`, service uses alternate suffix `283f76fe`. |
| P12 | `matrix.py storage-empty` | **1** | Expected no-login failure. Config unchanged; services `Claude Code-credentials` and `Claude Code` unsuffixed. |
| P13 | `matrix.py linked-file` | **0** | `claude.ai` / `max` from a scratch symlink; no refresh invoked. |
| P14 | Repeat `probe.py auth status --json`, fixture unchanged after P13 | **0** | Same `claude.ai` / `max` in a new process. Repeat read, not refresh proof. |
| P15 | `matrix.py token-api-helper-bearer` | **0** | `authMethod:oauth_token` and `apiKeySource:ANTHROPIC_API_KEY`. Ambiguous summary, not HTTP precedence proof. |

All **15/15** listed child commands ran: **11 exited 0, four exited 1**. This is coverage of a small local-selection matrix, not 15 successful authentications. A dummy token can produce green status; helper selection can be reported without executing the helper.

Representative P8 stdout (path normalized; null identities unchanged):

```json
{
  "loggedIn": true,
  "authMethod": "claude.ai",
  "apiProvider": "firstParty",
  "analyticsDisabled": true,
  "projectsDirectory": "$SCRATCH/config/projects",
  "configDirectory": "$SCRATCH/config",
  "email": null,
  "orgId": null,
  "orgName": null,
  "subscriptionType": "pro"
}
```

Representative P8 shim call: `find-generic-password -a credential-probe -w -s Claude Code-credentials-296b436d`. Legacy API-key service lookups also occurred and returned no credential. No real security command, login/logout, Keychain mutation or operator credential-file operation was executed by the probe.

## Decision 0017 gate ledger

| Required Q1 fact | This run | Remaining experiment |
| --- | --- | --- |
| Real file/Keychain read order | Mocked Keychain-first; static inspection | Disposable account with both stores, including malformed, locked and denied states. |
| File/link behavior on refresh | Atomic-writer inspection and repeat symlink reads | Account-driven refresh, active-path selection, inode/link observation and concurrent homes. |
| Conditions writing a suffixed item | Service-name selection; write/fallback code | Real login/renewal across writable, locked and unavailable disposable Keychains; metadata only. |

Q7: no credential-copy consent was exercised or proposed for native logins. The CIP requests narrow authority for separate token enrollment and transient injection. Copy-at-provision, Keychain export, refresh-token cloning and shared-file copy-back remain refused.

## Artifact verification

Product suites were not run: this is a research-only candidate. Earlier attached suites/login evidence were not replayed or accepted as current-release passes.

### Temporary execution-host stall during packaging

The command interface stopped producing results, including for isolated `pwd`, `true` and echo diagnostics. Two evidence-write commands were interrupted with exit **130**; a subsequent edit/read attempt exited **1** because the evidence file had not been created. The patch tool was then used to create and extend this companion. No interrupted operation is counted as a passing check.

The artifact verifier, `python3 $SCRATCH/verify_artifacts.py`, eventually returned exit **0**. A cancellation request raced its completion; the observed result was 0, not 130. An interim blocker note's attribution of exit 130 to the verifier was incorrect and is corrected on the board. It checked 12/12 required CIP headings, 2/2 parseable JSON examples, 15/15 probe-ledger rows, whitespace/public-path patterns, local links, the 96 KiB limit, and the exact two-file research-only Git change set. The nested `git status --porcelain` exited **0**. These checks do not establish authentication or refresh behavior. `git diff --check` also ran as a standalone process and exited **0**; because new files are untracked, the verifier separately checks their whitespace.

A task-scoped draft attachment attempt was interrupted with exit **130**, without a receipt. Recovery queries the task's outcome resources before choosing add versus update. Outcome attachment and handoff receipts are recorded by the board CLI rather than pre-attested in this artifact. No interrupted operation is counted as a pass.

## R1 reproducibility appendix (reconstructed harness, 2026-10-05)

Review verdict R1 (changes requested, revision 1) found the historical P1-P15
rows above unreproducible: the original `probe.py`, `matrix.py`, `probe.sb`,
Keychain shim and `verify_artifacts.py` were described but not supplied. The
original harness is unavailable, so this appendix supplies a reconstructed,
self-contained, sanitized harness and a fresh rerun on the pinned bundle. The
P-rows stay labelled historical; the R-rows below are the reconstructed
results. No identity between the two harnesses is claimed; their key fields
are compared row by row at the end. Bounds from the historical run carry over:
synthetic fixtures only, no model prompts, no issuance or revocation, no
real-store reads, no login or logout. Q1 remains 0/3 required
disposable-account facts established.

### Pinned inputs and expected counts per script

| Script | Pinned inputs | Expected counts |
| --- | --- | --- |
| probe.py | Bundle 2.1.287, 236282368 bytes, SHA-256 f1863213e4f55aaadc2e6ee617f934ada29930e5de4c5e5f8f9e34a0d594fdd7 (verified before every run); USER credential-probe; 25 s timeout | R1 exit 0 with version string; R2 exit 0 with setup-token help; R3 exit 1 loggedIn false; R14 exit 0 repeat read |
| matrix.py | Same bundle via probe.py; 11 cases; credential fixtures carry scopes user:inference and expiry now+3600 s; file label max, mock-Keychain label pro | 11/11 PASS; child exits 8x0 plus 3x1 across R4-R13 and R15 |
| shims/security | Account credential-probe; mock dir; exits 0 found, 44 absent, 99 refused, 1 info-not-locked; never forwards | R3 two lookups both 44; R8 four lookups served (pro wins); no 99 observed |
| bin/fake-helper | Marker plus DUMMY key; installed in scratch only | Marker absent after R6 and R15 (status selects the helper without executing it) |
| probe.sb | Deny network, home writes, real security execution, credential-location reads/writes | 15/15 children launch; no sandbox abort in the recorded sequence |
| verify_artifacts.py | 12 CIP headings, min 2 parseable JSON blocks, 15 historical plus 15 rerun rows, combined size cap, two-file scope | Every check ok, exit 0 (recorded below) |

Curator, spec and launcher pins are unchanged from the historical section:
curator main ca1b776fb580ec0cee0173bf150daf063023aeaa, spec rc.14 peeled
commit 43bf0a2506d5c354a73bbc3ea4623d4653db10c7, launcher
2517d2753945d0a8b0c40a291e4aef883a7c16ee. Rerun platform: macOS 15.7.4,
x86_64, sandbox-exec present, pinned 2.1.287 bundle invoked by explicit path
(the ambient default had since moved to 2.1.289 and was not used).

### Setup recipe

```sh
export SCRATCH="$(mktemp -d /tmp/cip0003-probe-XXXXXX)"; chmod 700 "$SCRATCH"
# Save the six bodies below as probe.py, matrix.py, verify_artifacts.py,
# security-shim, fake-helper and probe.sb.template next to this shell.
cp probe.py matrix.py verify_artifacts.py "$SCRATCH/"
mkdir -p "$SCRATCH/shims" "$SCRATCH/bin" "$SCRATCH/home" \
  "$SCRATCH/config" "$SCRATCH/tmp" "$SCRATCH/cwd"
cp security-shim "$SCRATCH/shims/security"; chmod 755 "$SCRATCH/shims/security"
cp fake-helper "$SCRATCH/bin/fake-helper"; chmod 755 "$SCRATCH/bin/fake-helper"
sed -e "s|@REALHOME@|$HOME|g" probe.sb.template > "$SCRATCH/probe.sb"
export CIP0003_CLAUDE_BINARY="<resolved 2.1.287 bundle path>"
python3 "$SCRATCH/probe.py" --version  # identity gate; exit 2 on mismatch
```

`CIP0003_CLAUDE_BINARY` is host-specific (it names the operator home path) so
it stays a placeholder here; probe.py refuses to run unless the size and
SHA-256 above match. Every CLI invocation below ran as a separate foreground
process; none was piped through tee.

### probe.py (exact body)

```python
#!/usr/bin/env python3
"""CIP-0003 R1 reconstructed probe wrapper (sanitized, self-contained).

Usage (from a scratch dir that contains this file):
  python3 probe.py --version
  python3 probe.py setup-token --help
  python3 probe.py auth status --json

Runs the pinned Claude bundle under sandbox-exec with a whitelist child
environment, a 25 s timeout, and $SCRATCH-normalized output. Exits with the
child's exit code (124 on wrapper timeout, 128+SIG on signal death,
2 on wrapper misuse).
Secrets policy: prints only argv/exit/stdout/stderr with the scratch path
normalized; fixtures are synthetic DUMMY values; never reads real stores.
"""
import hashlib
import json
import os
import subprocess
import sys

TIMEOUT_S = 25
# Pinned bundle identity (measured 2026-10-05, macOS x86_64):
EXPECT_SIZE = 236282368
EXPECT_SHA256 = "f1863213e4f55aaadc2e6ee617f934ada29930e5de4c5e5f8f9e34a0d594fdd7"
EXPECT_VERSION = "2.1.287"
# Env keys matrix.py may inject via case-env.json (names only, allowlisted):
ALLOW_EXTRA_ENV = (
    "CLAUDE_CODE_OAUTH_TOKEN",
    "ANTHROPIC_API_KEY",
    "ANTHROPIC_AUTH_TOKEN",
    "CLAUDE_CODE_API_KEY_HELPER_TTL_MS",
    "CLAUDE_SECURESTORAGE_CONFIG_DIR",
)


def fail(msg, code):
    sys.stderr.write("probe.py: %s\n" % msg)
    return code


def main(argv):
    scratch = os.path.dirname(os.path.abspath(__file__))
    binary = os.environ.get("CIP0003_CLAUDE_BINARY", "")
    if not binary or not os.path.isfile(binary):
        return fail("set CIP0003_CLAUDE_BINARY to the pinned bundle path", 2)
    if not argv or argv[0] not in ("--version", "setup-token", "auth"):
        return fail("usage: probe.py [--version|setup-token --help|auth status --json]", 2)
    with open(binary, "rb") as fh:
        blob = fh.read()
    if len(blob) != EXPECT_SIZE or hashlib.sha256(blob).hexdigest() != EXPECT_SHA256:
        return fail("bundle identity mismatch (size/sha256)", 2)
    for name in ("home", "config", "tmp", "cwd"):
        d = os.path.join(scratch, name)
        if not os.path.isdir(d):
            os.makedirs(d, 0o700)
    env = {
        "HOME": scratch + "/home",
        "CLAUDE_CONFIG_DIR": scratch + "/config",
        "TMPDIR": scratch + "/tmp",
        "USER": "credential-probe",
        "LOGNAME": "credential-probe",
        "PATH": scratch + "/shims:/usr/bin:/bin",
        "LANG": "C",
        "LC_ALL": "C",
        "TERM": "dumb",
        "DISABLE_AUTOUPDATER": "1",
        "DISABLE_TELEMETRY": "1",
        "CIP0003_SCRATCH": scratch,
    }
    extra_path = os.path.join(scratch, "case-env.json")
    if os.path.exists(extra_path):
        with open(extra_path) as fh:
            extra = json.load(fh)
        for key, val in extra.items():
            if key not in ALLOW_EXTRA_ENV or not isinstance(val, str):
                return fail("case-env.json key refused: %r" % (key,), 2)
            env[key] = val
    cmd = ["/usr/bin/sandbox-exec", "-f", scratch + "/probe.sb", binary] + argv
    try:
        proc = subprocess.run(
            cmd, env=env, cwd=scratch + "/cwd", timeout=TIMEOUT_S,
            stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    except subprocess.TimeoutExpired as exc:
        out = (exc.stdout or b"").replace(scratch.encode(), b"$SCRATCH")
        err = (exc.stderr or b"").replace(scratch.encode(), b"$SCRATCH")
        sys.stdout.write(json.dumps({"argv": argv, "exit": 124, "timeout": True,
                                     "stdout": out.decode("utf-8", "replace"),
                                     "stderr": err.decode("utf-8", "replace")}))
        sys.stdout.write("\n")
        return 124
    out = proc.stdout.replace(scratch.encode(), b"$SCRATCH")
    err = proc.stderr.replace(scratch.encode(), b"$SCRATCH")
    sys.stdout.write(json.dumps({"argv": argv, "exit": proc.returncode,
                                 "timeout": False,
                                 "stdout": out.decode("utf-8", "replace"),
                                 "stderr": err.decode("utf-8", "replace")}))
    sys.stdout.write("\n")
    if proc.returncode < 0:
        return 128 - proc.returncode  # signal death surfaces as 128+SIG
    return proc.returncode


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
```

### matrix.py (exact body)

```python
#!/usr/bin/env python3
"""CIP-0003 R1 reconstructed selection matrix (sanitized, self-contained).

Usage: python3 matrix.py <case>
Cases: token token-api token-helper file keychain-file keychain none
       storage-override storage-empty linked-file token-api-helper-bearer

Each case resets only its owned scratch fixtures, writes synthetic fixtures
(DUMMY values, never real credentials), invokes probe.py as a separate
foreground process, prints a configuration snapshot plus the observed status
fields and shim calls, asserts the narrow per-case expectation, and exits 0
on assertion pass / 1 on assertion fail. The child exit is reported inside
the output, not as this process's exit.
"""
import hashlib
import json
import os
import shutil
import subprocess
import sys
import time

SCRATCH = os.path.dirname(os.path.abspath(__file__))
USER = "credential-probe"
DUMMY_TOKEN = "DUMMY-SYNTHETIC-OAUTH-TOKEN-0000"
DUMMY_API = "DUMMY-SYNTHETIC-API-KEY-0000"
DUMMY_HELPER = "DUMMY-SYNTHETIC-HELPER-KEY-0000"


def cred(access, refresh, sub):
    return {"claudeAiOauth": {
        "accessToken": access, "refreshToken": refresh,
        "expiresAt": int(time.time() * 1000) + 3600 * 1000,
        "scopes": ["user:inference"], "subscriptionType": sub}}


FILE_CRED = lambda: cred("DUMMY-SYNTHETIC-FILE-ACCESS-0000",
                         "DUMMY-SYNTHETIC-FILE-REFRESH-0000", "max")
CHAIN_CRED = lambda: cred("DUMMY-SYNTHETIC-CHAIN-ACCESS-0000",
                          "DUMMY-SYNTHETIC-CHAIN-REFRESH-0000", "pro")


def svc_suffix(config_dir):
    return hashlib.sha256(config_dir.encode()).hexdigest()[:8]


def service(config_dir):
    return "Claude Code-credentials-%s" % svc_suffix(config_dir)


def reset_owned():
    for rel in ("config/settings.json", "config/.credentials.json",
                "link-target.json", "case-env.json", "helper.marker"):
        p = os.path.join(SCRATCH, rel)
        if os.path.islink(p) or os.path.isfile(p):
            os.unlink(p)
    mock = os.path.join(SCRATCH, "mock-chain")
    if os.path.isdir(mock):
        shutil.rmtree(mock)
    open(os.path.join(SCRATCH, "shim.log"), "w").close()


def write_case_env(extra):
    with open(os.path.join(SCRATCH, "case-env.json"), "w") as fh:
        json.dump(extra, fh)


def write_mock_chain(config_dir):
    d = os.path.join(SCRATCH, "mock-chain", USER)
    os.makedirs(d, 0o700)
    with open(os.path.join(d, service(config_dir)), "w") as fh:
        json.dump(CHAIN_CRED(), fh)
    os.chmod(os.path.join(d, service(config_dir)), 0o600)


def write_file_cred():
    p = os.path.join(SCRATCH, "config", ".credentials.json")
    with open(p, "w") as fh:
        json.dump(FILE_CRED(), fh)
    os.chmod(p, 0o600)


def write_helper_settings():
    helper = os.path.join(SCRATCH, "bin", "fake-helper")
    with open(os.path.join(SCRATCH, "config", "settings.json"), "w") as fh:
        json.dump({"apiKeyHelper": helper}, fh)


def snapshot(case, extra):
    files = []
    for rel in ("config/settings.json", "config/.credentials.json",
                "link-target.json", "helper.marker"):
        p = os.path.join(SCRATCH, rel)
        if os.path.islink(p):
            files.append("%s symlink->%s" % (rel, os.readlink(p)))
        elif os.path.isfile(p):
            files.append("%s file mode=%o" % (rel, os.stat(p).st_mode & 0o777))
    mocks = []
    md = os.path.join(SCRATCH, "mock-chain", USER)
    if os.path.isdir(md):
        mocks = sorted(os.listdir(md))
    return {"case": case, "config_dir": "$SCRATCH/config",
            "service_template": "Claude Code-credentials-<suffix>",
            "suffix_scratch": svc_suffix(SCRATCH + "/config"),
            "extra_env_names": sorted(extra.keys()), "files": files,
            "mock_services": mocks}


CASES = ("token", "token-api", "token-helper", "file", "keychain-file",
         "keychain", "none", "storage-override", "storage-empty",
         "linked-file", "token-api-helper-bearer")


def setup(case):
    """Build fixtures; returns (extra_env, note)."""
    cfg = SCRATCH + "/config"
    if case == "token":
        return {"CLAUDE_CODE_OAUTH_TOKEN": DUMMY_TOKEN}, ""
    if case == "token-api":
        return {"CLAUDE_CODE_OAUTH_TOKEN": DUMMY_TOKEN,
                "ANTHROPIC_API_KEY": DUMMY_API}, ""
    if case == "token-helper":
        write_helper_settings()
        return {"CLAUDE_CODE_OAUTH_TOKEN": DUMMY_TOKEN}, ""
    if case == "file":
        write_file_cred()
        return {}, ""
    if case == "keychain-file":
        write_file_cred()
        write_mock_chain(cfg)
        return {}, ""
    if case == "keychain":
        write_mock_chain(cfg)
        return {}, ""
    if case == "none":
        return {}, ""
    if case == "storage-override":
        alt = SCRATCH + "/altstore"
        if not os.path.isdir(alt):
            os.makedirs(alt, 0o700)
        return {"CLAUDE_SECURESTORAGE_CONFIG_DIR": alt}, \
            "alt suffix=%s" % svc_suffix(alt)
    if case == "storage-empty":
        return {"CLAUDE_SECURESTORAGE_CONFIG_DIR": ""}, "unsuffixed lookup"
    if case == "linked-file":
        with open(os.path.join(SCRATCH, "link-target.json"), "w") as fh:
            json.dump(FILE_CRED(), fh)
        os.symlink("../link-target.json", cfg + "/.credentials.json")
        return {}, ""
    if case == "token-api-helper-bearer":
        write_helper_settings()
        return {"CLAUDE_CODE_OAUTH_TOKEN": DUMMY_TOKEN,
                "ANTHROPIC_API_KEY": DUMMY_API,
                "ANTHROPIC_AUTH_TOKEN": "DUMMY-SYNTHETIC-BEARER-0000"}, ""
    raise SystemExit("unknown case: %s" % case)


def check(case, exit_code, status, shimlog, note):
    """Narrow per-case assertions; returns list of failures (empty=pass)."""
    f = []
    def eq(name, got, want):
        if got != want:
            f.append("%s: got %r want %r" % (name, got, want))

    def has(path, frag):
        return any(frag in line for line in shimlog)
    if case == "token":
        eq("exit", exit_code, 0)
        eq("authMethod", status.get("authMethod"), "oauth_token")
    elif case == "token-api":
        eq("exit", exit_code, 0)
        eq("authMethod", status.get("authMethod"), "api_key")
        eq("apiKeySource", status.get("apiKeySource"), "ANTHROPIC_API_KEY")
    elif case == "token-helper":
        eq("exit", exit_code, 0)
        eq("authMethod", status.get("authMethod"), "api_key_helper")
        eq("apiKeySource", status.get("apiKeySource"), "apiKeyHelper")
        if os.path.exists(SCRATCH + "/helper.marker"):
            f.append("helper marker present (status must not execute helper)")
    elif case == "file":
        eq("exit", exit_code, 0)
        eq("authMethod", status.get("authMethod"), "claude.ai")
        eq("subscriptionType", status.get("subscriptionType"), "max")
    elif case == "keychain-file":
        eq("exit", exit_code, 0)
        eq("authMethod", status.get("authMethod"), "claude.ai")
        eq("subscriptionType", status.get("subscriptionType"), "pro")
    elif case == "keychain":
        eq("exit", exit_code, 0)
        eq("authMethod", status.get("authMethod"), "claude.ai")
        eq("subscriptionType", status.get("subscriptionType"), "pro")
        if os.path.exists(SCRATCH + "/config/.credentials.json"):
            f.append("credential file created (must not be)")
    elif case == "none":
        eq("exit", exit_code, 1)
        eq("loggedIn", status.get("loggedIn"), False)
        eq("authMethod", status.get("authMethod"), "none")
    elif case == "storage-override":
        eq("exit", exit_code, 1)
        if not has("shim", svc_suffix(SCRATCH + "/altstore")):
            f.append("alt-store suffix lookup missing from shim log")
    elif case == "storage-empty":
        eq("exit", exit_code, 1)
        if not has("shim", "-s Claude Code-credentials ") and \
           not has("shim", "-s Claude Code-credentials\n") and \
           not any(l.rstrip().endswith("-s Claude Code-credentials") for l in shimlog):
            f.append("unsuffixed service lookup missing from shim log")
    elif case == "linked-file":
        eq("exit", exit_code, 0)
        eq("authMethod", status.get("authMethod"), "claude.ai")
        eq("subscriptionType", status.get("subscriptionType"), "max")
        if not os.path.islink(SCRATCH + "/config/.credentials.json"):
            f.append("symlink fixture missing after run")
    elif case == "token-api-helper-bearer":
        eq("exit", exit_code, 0)
        eq("authMethod", status.get("authMethod"), "oauth_token")
    return f


def main(case):
    if case not in CASES:
        return "usage: matrix.py <%s>" % "|".join(CASES), 2
    reset_owned()
    extra, note = setup(case)
    write_case_env(extra)
    proc = subprocess.run(
        [sys.executable, SCRATCH + "/probe.py", "auth", "status", "--json"],
        stdout=subprocess.PIPE, stderr=subprocess.PIPE, timeout=120)
    try:
        env_out = json.loads(proc.stdout.decode())
    except ValueError:
        print(json.dumps({"case": case, "error": "probe envelope unparseable",
                          "probe_exit": proc.returncode,
                          "probe_stderr": proc.stderr.decode()[-500:]}))
        return "", 1
    try:
        status = json.loads(env_out["stdout"])
    except ValueError:
        status = {"_unparseable": env_out["stdout"][-300:],
                  "_stderr": env_out["stderr"][-300:]}
    with open(os.path.join(SCRATCH, "shim.log")) as fh:
        shimlog = [l.rstrip("\n") for l in fh]
    fails = check(case, env_out["exit"], status, shimlog, note)
    slim = {k: status.get(k) for k in (
        "loggedIn", "authMethod", "apiProvider", "apiKeySource",
        "subscriptionType", "configDirectory") if k in status}
    if "_unparseable" in status:
        slim = status
    print(json.dumps({
        "case": case, "snapshot": snapshot(case, extra), "note": note,
        "child_exit": env_out["exit"], "status": slim,
        "shim_calls": shimlog, "failures": fails,
        "verdict": "PASS" if not fails else "FAIL"}, indent=1))
    return "", 0 if not fails else 1


if __name__ == "__main__":
    msg, code = main(sys.argv[1] if len(sys.argv) > 1 else "")
    if msg:
        sys.stderr.write(msg + "\n")
    sys.exit(code)
```

### shims/security (exact body)

```sh
#!/bin/sh
# CIP-0003 R1 reconstructed Keychain shim (sanitized).
# Intercepts PATH `security` calls from the sandboxed child. Never forwards
# to the real /usr/bin/security. Serves synthetic mock objects or not-found.
# Logs argv only (command/service/account names); responses are never logged.
# Exit codes mirror macOS security: 0 found, 44 not-found, 99 refused-here.
S="$CIP0003_SCRATCH"
printf 'security %s\n' "$*" >> "$S/shim.log"
op="$1"; shift
getarg() { # $1=flag; prints following value
  flag="$1"; shift
  prev=""
  for a in "$@"; do
    if [ "$prev" = "1" ]; then printf '%s' "$a"; return; fi
    if [ "$a" = "$flag" ]; then prev="1"; fi
  done
}
case "$op" in
  find-generic-password)
    acct="$(getarg -a "$@")"; svc="$(getarg -s "$@")"
    f="$S/mock-chain/$acct/$svc"
    if [ -n "$acct" ] && [ -n "$svc" ] && [ -f "$f" ]; then cat "$f"; exit 0; fi
    exit 44
    ;;
  show-keychain-info) exit 1 ;; # not {locked}=36; no real state consulted
  -i|add-generic-password|delete-generic-password) exit 99 ;;
  *) exit 99 ;;
esac
```

### bin/fake-helper (exact body)

```sh
#!/bin/sh
# CIP-0003 R1 fake apiKeyHelper (synthetic). Records invocation, prints a
# DUMMY credential. Installed only in scratch; never in real settings.
touch "$CIP0003_SCRATCH/helper.marker"
printf '%s' 'DUMMY-SYNTHETIC-HELPER-KEY-0000'
```

### probe.sb.template (exact body)

```text
;; CIP-0003 R1 reconstructed probe sandbox (template, sanitized).
;; Placeholders @REALHOME@ are substituted at setup with $HOME (see setup
;; recipe). No secret, account, or credential identifier appears here.
;; Scope: denies network, all writes under the real home, execution of the
;; real /usr/bin/security, and reads/writes of credential-bearing locations
;; (user Keychains, managed Claude settings, system key stores). Other
;; non-credential home reads remain allowed; credential safety additionally
;; rests on the synthetic USER, suffixed service names, the PATH shim that
;; never forwards, and redirected HOME/CLAUDE_CONFIG_DIR/TMPDIR.
(version 1)
(allow default)
(deny network*)
(deny network-outbound)
(deny file-write* (subpath "@REALHOME@"))
(deny file-read* (literal "/usr/bin/security"))
(deny process-exec (literal "/usr/bin/security"))
(deny file-read* (subpath "@REALHOME@/Library/Keychains"))
(deny file-write* (subpath "@REALHOME@/Library/Keychains"))
(deny file-read* (subpath "@REALHOME@/.claude"))
(deny file-write* (subpath "@REALHOME@/.claude"))
(deny file-read* (literal "@REALHOME@/.claude.json"))
(deny file-write* (literal "@REALHOME@/.claude.json"))
(deny file-read* (subpath "/Library/Keychains"))
(deny file-write* (subpath "/Library/Keychains"))
(deny file-read* (subpath "/private/var/db/SystemKey"))
(deny file-write* (subpath "/private/var/db/SystemKey"))
```

### verify_artifacts.py (exact body)

```python
#!/usr/bin/env python3
"""CIP-0003 R1 reconstructed artifact verifier (sanitized, self-contained).

Usage: python3 verify_artifacts.py <cip-draft.md> <evidence.md>
Checks document structure only; establishes no authentication behavior.
Prints per-check counts; exits 0 iff all pass, 1 otherwise.
"""
import json
import os
import re
import subprocess
import sys

REQUIRED_HEADINGS = [
    "## Summary", "## Motivation and user stories", "## Current state",
    "## Design", "### Options considered", "### Recommendation",
    "## Security considerations", "## Compatibility and migration",
    "## Specification changes", "## Implementation plan", "## Test plan",
    "## Open questions for the operator",
]
SIZE_LIMIT = 96 * 1024
# Self-hosting: each pattern is concatenated so this source (embedded in the
# evidence file) never contains a forbidden literal itself.
FORBIDDEN = ("/User" + "s/", "/private/var/folder" + "s/", "sk-an" + "t-",
             "AKI" + "A", "xo" + "x", "gh" + "p_", "github_pa" + "t_",
             "administrato" + "r", "Mac min" + "i", "macmin" + "i",
             "minima" + "c")
EXPECT_HISTORICAL_ROWS = 15
EXPECT_RERUN_ROWS = 15


def run(cmd, cwd):
    return subprocess.run(cmd, cwd=cwd, stdout=subprocess.PIPE,
                          stderr=subprocess.PIPE, timeout=60)


def main(cip_path, ev_path):
    fails = []

    def check(name, ok_count, total, detail=""):
        nonlocal fails
        ok = ok_count == total
        print("%-28s %d/%d %s %s" % (name, ok_count, total,
                                     "ok" if ok else "FAIL", detail))
        if not ok:
            fails.append(name)

    with open(cip_path, encoding="utf-8") as fh:
        cip = fh.read()
    with open(ev_path, encoding="utf-8") as fh:
        ev = fh.read()
    both = {"cip": cip, "evidence": ev}

    heads = [h for h in REQUIRED_HEADINGS if h in cip]
    check("cip-required-headings", len(heads), len(REQUIRED_HEADINGS))
    blocks = re.findall(r"```json\n(.*?)```", cip + "\n" + ev, re.S)
    parsed = 0
    for b in blocks:
        try:
            json.loads(b)
            parsed += 1
        except ValueError:
            pass
    check("json-examples-parse", parsed, len(blocks),
          "(%d blocks)" % len(blocks))
    if len(blocks) < 2:
        fails.append("json-examples-min2")
        print("json-examples-min2          %d/2 FAIL" % len(blocks))
    else:
        print("json-examples-min2          %d/2 ok" % len(blocks))
    hist = sorted(set(re.findall(r"\| (P\d+) \|", ev)))
    check("historical-probe-rows", len(hist), EXPECT_HISTORICAL_ROWS)
    rerun = sorted(set(re.findall(r"\| (R\d+) \|", ev)))
    check("rerun-probe-rows", len(rerun), EXPECT_RERUN_ROWS)
    bad_ws = sum(1 for t in both.values() for l in t.split("\n")
                 if l != l.rstrip() or "\t" in l or "\r" in l)
    check("whitespace-clean", 0 if bad_ws else 1, 1,
          "(%d bad lines)" % bad_ws)
    ends_nl = sum(1 for t in both.values() if t.endswith("\n"))
    check("final-newlines", ends_nl, 2)
    hits = [p for p in FORBIDDEN if p in cip or p in ev]
    check("forbidden-patterns-absent", 0 if hits else 1, 1,
          "(%s)" % ",".join(hits) if hits else "")
    if "DUMMY-SYNTHETIC" not in ev:
        fails.append("synthetic-marker")
        print("synthetic-marker            0/1 FAIL")
    else:
        print("synthetic-marker            1/1 ok")
    links = re.findall(r"\]\(([^) ]+)\)", cip + ev)
    local = [l for l in links if not re.match(r"(https?://|#|mailto:)", l)]
    missing = [l for l in local if not os.path.isfile(
        os.path.join(os.path.dirname(cip_path), l))]
    check("local-links-resolve", len(local) - len(missing), len(local),
          "(%s)" % ",".join(missing) if missing else "")
    size = len(cip.encode()) + len(ev.encode())
    check("combined-size-bytes", 1 if size <= SIZE_LIMIT else 0, 1,
          "(%d<=%d)" % (size, SIZE_LIMIT))
    root = os.path.dirname(os.path.dirname(os.path.abspath(cip_path)))
    gs = run(["git", "status", "--porcelain"], root)
    changed = sorted(l for l in gs.stdout.decode().split("\n") if l.strip())
    want = sorted("?? .research/261004_CIP-0003-claude-managed-home-credential-modes%s.md"
                  % s for s in ("", "_evidence"))
    check("git-two-file-scope", 1 if changed == want and gs.returncode == 0 else 0, 1,
          "(%d paths)" % len(changed))
    dc = run(["git", "diff", "--check"], root)
    check("git-diff-check", 1 if dc.returncode == 0 else 0, 1,
          "(exit %d)" % dc.returncode)
    print("VERDICT", "PASS" if not fails else "FAIL: %s" % ",".join(fails))
    return 0 if not fails else 1


if __name__ == "__main__":
    if len(sys.argv) != 3:
        sys.stderr.write("usage: verify_artifacts.py <cip> <evidence>\n")
        sys.exit(2)
    sys.exit(main(sys.argv[1], sys.argv[2]))
```

### Run sequence (recorded 2026-10-05)

Each line ran as a separate foreground process; wrapper exits are real and
unpiped. R14 deliberately skips the fixture reset (only the shim log is
truncated) so the R13 symlink input persists into a new process.

```sh
: > "$SCRATCH/shim.log"; rm -f "$SCRATCH/case-env.json"
python3 "$SCRATCH/probe.py" --version                    # R1, wrapper exit 0
python3 "$SCRATCH/probe.py" setup-token --help          # R2, wrapper exit 0
python3 "$SCRATCH/probe.py" auth status --json          # R3, wrapper exit 1
python3 "$SCRATCH/matrix.py" token                       # R4, matrix exit 0
python3 "$SCRATCH/matrix.py" token-api                   # R5, matrix exit 0
python3 "$SCRATCH/matrix.py" token-helper                # R6, matrix exit 0
python3 "$SCRATCH/matrix.py" file                        # R7, matrix exit 0
python3 "$SCRATCH/matrix.py" keychain-file               # R8, matrix exit 0
python3 "$SCRATCH/matrix.py" keychain                    # R9, matrix exit 0
python3 "$SCRATCH/matrix.py" none                        # R10, matrix exit 0
python3 "$SCRATCH/matrix.py" storage-override            # R11, matrix exit 0
python3 "$SCRATCH/matrix.py" storage-empty               # R12, matrix exit 0
python3 "$SCRATCH/matrix.py" linked-file                 # R13, matrix exit 0
: > "$SCRATCH/shim.log"
python3 "$SCRATCH/probe.py" auth status --json          # R14, wrapper exit 0
python3 "$SCRATCH/matrix.py" token-api-helper-bearer     # R15, matrix exit 0
```

matrix.py exits 0 on assertion pass and 1 on assertion fail; the child exit
is reported inside its JSON. probe.py exits with the child exit (0, 1, 124
on wrapper timeout, 128+SIG on signal death, 2 on misuse).

### Rerun ledger (reconstructed, 2026-10-05)

| ID | Command/case | Wrapper/matrix exit | Child exit | Observation |
| --- | --- | --- | --- | --- |
| R1 | probe.py --version | 0 | 0 | 2.1.287 (Claude Code), empty stderr |
| R2 | probe.py setup-token --help | 0 | 0 | Long-lived token help, subscription required; no expiry flag; no issuance |
| R3 | probe.py auth status --json, empty | 1 | 1 | loggedIn false, none; expected-red: no fixture; two shim lookups, both 44 |
| R4 | matrix.py token | 0 | 0 | oauth_token; no credential file created |
| R5 | matrix.py token-api | 0 | 0 | api_key, source ANTHROPIC_API_KEY |
| R6 | matrix.py token-helper | 0 | 0 | api_key_helper; helper marker absent |
| R7 | matrix.py file | 0 | 0 | claude.ai, max; mock Keychain 44s |
| R8 | matrix.py keychain-file | 0 | 0 | claude.ai, pro; Keychain label wins |
| R9 | matrix.py keychain | 0 | 0 | claude.ai, pro; no credential file created |
| R10 | matrix.py none | 0 | 1 | loggedIn false, none; expected-red after removals |
| R11 | matrix.py storage-override | 0 | 1 | Expected-red; config unchanged; alt-suffix lookups da61ab1a |
| R12 | matrix.py storage-empty | 0 | 1 | Expected-red; unsuffixed lookups; config unchanged |
| R13 | matrix.py linked-file | 0 | 0 | claude.ai, max via symlink; link intact after run |
| R14 | probe.py auth status --json, repeat | 0 | 0 | Same claude.ai, max; repeat read, not refresh proof |
| R15 | matrix.py token-api-helper-bearer | 0 | 0 | oauth_token plus apiKeySource ANTHROPIC_API_KEY; ambiguous summary |

15/15 rows ran: 11 child exits 0, 4 child exits 1 (R3, R10, R11, R12, all
expected-red no-login rows). All 11 matrix assertions PASS; all 4 probe
wrappers returned their child exit. Coverage is the same small
local-selection matrix as the historical run, not 15 authentications.

### Representative R8 envelope (verbatim, path-normalized)

```json
{
 "case": "keychain-file",
 "snapshot": {
  "case": "keychain-file",
  "config_dir": "$SCRATCH/config",
  "service_template": "Claude Code-credentials-<suffix>",
  "suffix_scratch": "4184497a",
  "extra_env_names": [],
  "files": [
   "config/.credentials.json file mode=600"
  ],
  "mock_services": [
   "Claude Code-credentials-4184497a"
  ]
 },
 "note": "",
 "child_exit": 0,
 "status": {
  "loggedIn": true,
  "authMethod": "claude.ai",
  "apiProvider": "firstParty",
  "subscriptionType": "pro",
  "configDirectory": "$SCRATCH/config"
 },
 "shim_calls": [
  "security find-generic-password -a credential-probe -w -s Claude Code-4184497a",
  "security find-generic-password -a credential-probe -w -s Claude Code-credentials-4184497a",
  "security find-generic-password -a credential-probe -w -s Claude Code-credentials-4184497a",
  "security find-generic-password -a credential-probe -w -s Claude Code-credentials-4184497a"
 ],
 "failures": [],
 "verdict": "PASS"
}
```

### Bring-up corrections (reconstructed harness only)

1. The first sandbox draft excepted the bundle from a home read-denial with a
conditional clause; the platform compiler aborted (SIGABRT, exit 134) even
for a trivial command, and one shakeout version probe died with it. The
profile above (explicit credential-location denials) is the one that ran all
15 recorded rows. The shakeout death is not counted.
2. The first shim build returned the flag name instead of the flag value, so
every lookup 44'd. A direct shim test on a present object exited 44 and
exposed it; after the fix the same test exits 0 with the mock JSON, absent
44, write 99, info 1. Two pre-fix R8/R9 attempts are discarded, not counted.
3. The first R15 setup omitted the fourth input the case name requires and
observed api_key; adding the synthetic ANTHROPIC_AUTH_TOKEN reproduced the
historical oauth_token-plus-apiKeySource summary exactly. The recorded R15 is
the four-input run. R11 ran twice (dead assertion alternative removed; both
runs matrix exit 0 with the same child observation); the recorded R11 is the
second run with the final bytes above.

### Historical versus reconstructed comparison

Key fields agree on all 15 rows: same child exits (11x0, 4x1), same
authMethod per row, same subscription labels (file max, mock-Keychain pro),
same unsuffixed services on the empty override, same symlink readability and
repeat read. Service suffixes differ per scratch path by design
(reconstructed 4184497a and alt da61ab1a; historical 296b436d, alt 283f76fe,
each equal to SHA-256 of its own config path prefix 8). No wider claim is
made: both harnesses are synthetic-selection probes, and Q1 still needs its
authorized disposable-account experiment.

### Verifier run (reconstructed, 2026-10-05)

`python3 verify_artifacts.py` checks document structure only; it establishes
no authentication behavior. Final of three runs on 2026-10-05, all exit 0;
output quoted with trailing spaces stripped:

```text
cip-required-headings        12/12 ok
json-examples-parse          3/3 ok (3 blocks)
json-examples-min2          3/2 ok
historical-probe-rows        15/15 ok
rerun-probe-rows             15/15 ok
whitespace-clean             1/1 ok (0 bad lines)
final-newlines               2/2 ok
forbidden-patterns-absent    1/1 ok
synthetic-marker            1/1 ok
local-links-resolve          2/2 ok
combined-size-bytes          1/1 ok (84970<=98304)
git-two-file-scope           1/1 ok (2 paths)
git-diff-check               1/1 ok (exit 0)
VERDICT PASS
```
