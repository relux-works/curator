# Security

## Installed command execution

Commands launched through Curator-installed shims run as the invoking user and with that user's operating-system privileges. Portable assurance does not change the command's user identity or provide a complete operating-system sandbox. A capability declaration or audit result alone is not a privilege boundary.

Script commands are declared-only by default. Only a script command that opts in with `execution_policy: "script-worker-v1"` uses Curator's enforced script path and its manager-owned portable controls. Those controls are not a kernel sandbox and do not promise hard confinement to every declared capability. See [Protocol Core §4.1.1](https://github.com/relux-works/curator-spec/blob/23435129ebc4c29e5b7f75ec72a0aa0cd3f16065/protocol/core.md#411-portable-script-worker-v1-execution-policy).

Compiled builds default to portable `manager-worker-v1` assurance. The manager applies its portable worker controls, but they do not provide kernel-enforced confinement. Explicitly selected `execution.mode: verified` is the provider-backed enforcement path. It requires a separately installed trusted provider and fails closed rather than falling back to portable assurance when the provider or required evidence is unavailable. This Curator release ships no verified provider. See [Protocol Core §4.2.1](https://github.com/relux-works/curator-spec/blob/23435129ebc4c29e5b7f75ec72a0aa0cd3f16065/protocol/core.md#421-portable-manager-worker-v1-execution-policy), [Assurance Protocol §1](https://github.com/relux-works/curator-spec/blob/23435129ebc4c29e5b7f75ec72a0aa0cd3f16065/protocol/assurance.md#1-closed-selection), [§2](https://github.com/relux-works/curator-spec/blob/23435129ebc4c29e5b7f75ec72a0aa0cd3f16065/protocol/assurance.md#2-platform-neutral-provider-contract), and [§5](https://github.com/relux-works/curator-spec/blob/23435129ebc4c29e5b7f75ec72a0aa0cd3f16065/protocol/assurance.md#5-failure-rules).

## Release installer verification

The macOS and Linux `install.sh` installer verifies `checksums.txt` before
installing Curator. It prefers GitHub artifact attestation through
`gh attestation verify`; when that command is unavailable, it verifies the
release's `checksums.txt.sig` and `checksums.txt.pem` with cosign. Both paths
pin the signer to
`^https://github.com/relux-works/curator/\.github/workflows/release\.yml@refs/tags/v`
and the issuer to `https://token.actions.githubusercontent.com`. The GitHub
CLI path also pins the signer workflow when the installed CLI supports that
option. After the attestation or signature succeeds, the installer checks the
archive's SHA-256 against `checksums.txt` before extracting or installing it.

Missing verification tools, invalid attestations or signatures, identity or
issuer mismatches, and archive checksum failures all refuse installation. Use
the [GitHub CLI](https://cli.github.com/) or
[cosign](https://docs.sigstore.dev/cosign/system_config/installation/).

`CURATOR_INSTALL_INSECURE_SKIP_VERIFY=1` is an emergency bypass. It skips all
attestation, signature, and checksum verification and installs the downloaded
executable without integrity verification. The installer prints a warning
when this variable is set.
