# TASK-260910-2t0iun review verdict — rev3 (tree 18353c49, base 3f60f7f0): CHANGES REQUESTED

## F1 (blocking): the gh path always fails with a real gh that supports --signer-workflow
install.sh:69-73 passes `--signer-workflow` together with `--cert-identity-regex`. Real gh makes these flags mutually exclusive. Reproduced on this host (gh 2.88.1, zsh, real exit code):
```
$ gh attestation verify /etc/hosts --repo relux-works/curator --signer-workflow relux-works/curator/.github/workflows/release.yml --cert-identity-regex '^x' --cert-oidc-issuer https://token.actions.githubusercontent.com
if any flags in the group [cert-identity cert-identity-regex signer-repo signer-workflow] are set none of the others can be; [cert-identity-regex signer-workflow] were all set
rc=1
```
The script sees `--signer-workflow` in the help text, takes this branch, and refuses every install for every user with a current gh. The rows cannot catch this because the test's fake gh accepts the combination; installer_script_test.go:85-86 pins the invalid argv as the expected call.
Fix: pass only one signer pin per branch, for example `--repo "$REPO" --signer-workflow "$SIGNER_WORKFLOW" --cert-oidc-issuer "$OIDC_ISSUER"`. Also check that the signer-workflow value format matches what gh expects (host/owner/repo/path or owner/repo/path) and pin the ref if gh supports that. Make the fake gh reject mutually exclusive flags the same way real gh does, and add a row for that. A useful extra check: a manual run of the gh branch against a real published release; say so if none is available.

## Checked and OK
- Order: attestation/signature, then exact awk `$2 == archive` single-match sha256, then tar. Nothing is written to the install location before verification succeeds. A missing verifier refuses. The opt-out prints a loud warning and is documented.
- The cosign identity regexp is anchored and escapes its dots (the `github.com` host dot is unescaped but only matches one character, so this is harmless). The .sig/.pem names match .goreleaser.yml `signs`. release.yml attests dist/checksums.txt.
- The Windows ledger rows match gatefix-2 (skip-classes.tsv:120, platform-cases.tsv:773).
- There are no CHANGELOG/LOGBOOK edits.

Route: to-dev.
