# Platform release-key adoption

Machine/fleet `environments.source_signers` maps a canonical repository identity to SSH public keys (`{ "type": "ssh", "key": "<OpenSSH public key>" }`) or OpenPGP fingerprints. The key type and material identify an SSH signer; comments do not. All packages in the same repository share its source policy. The context package schema contains no signer carrier.

Curator reads the machine config through `CURATOR_CONFIG` (default `~/.curator/config.json`) and the optional fleet overlay through `CURATOR_SYSTEM_CONFIG` (default `/etc/curator/config.json`). Source policy flows through `PolicyFromConfig`, `contextresolve.verifyCandidate`, `gitManager.verifyCandidateInRepo` and `sourceSignerPosture`. Required-but-missing policy refuses; a configured empty list admits no key.

## Verify a skill release tag

From the skill repository, with its reviewed public signer file and actual release tag:

```sh
git -c gpg.ssh.allowedSignersFile=RELEASE_SIGNERS tag -v v0.1.0
```

Expected platform fingerprint: `SHA256:x3t8hOxl+PixTw1eTqV+qCqSp88Vw7AEMrmGUvs213I`. Curator-family sources (`curator`, `curator-agent-launcher`, `curator-spec`) retain `SHA256:H9oEmLBNf0JIQvCEarIqbcSaUaQCidesH6rl2+rcl0c`. Never add the agent commit key to a release signer list. A good cryptographic signature accompanied by `No principal matched` is not an admitted signer; check the exit status and expected fingerprint.

`curator env status` names a verified key as `ssh-ed25519 SHA256:...` in a source-signers row. `--json` exposes it in `source_signers[].signer`. Status checks cached objects without fetching; missing material is `unknown`, not a verified signer. A source rejection includes the source/ref and claimed signer where available. SSHSIG's version is a uint32; parsing it correctly lets rejected SSH tags report their actual claimed key as well as accepted ones.

## Existing contract and the release-only gap

Environments §1.4 currently authorizes a valid admitted **tag OR peeled commit** signature. Explicit `revision` selections use commit signatures. The source map can separate platform and Curator keys, and excludes agent commit-key-only signatures when that key is absent. It cannot express **tag signatures only**: an admitted release key signing a commit can still authorize an unsigned tag under the existing contract. This change preserves supported tag/commit/revision behavior; it does not silently replace that policy globally.

Strict source-scoped release-only admission needs an owner/spec decision defining the source policy and explicit-revision behavior. No new flag, hardcoded repository exception, key-role inference or enforcement-default flip is introduced. The pinned legacy corpus remains unchanged. Public signature fixtures prove key-specific crypto and refusal controls; they do not certify release package installation or an activated profile.
