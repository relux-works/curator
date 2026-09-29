# TASK-260910-2t0iun — rework 1 (THE ONLY CURRENT INSTRUCTION, with 2t0iun-decision-1.md)

Review rev3 = CHANGES REQUESTED (`TASK-260910-2t0iun_review-verdict-rev3.md`), F1: install.sh:69-73 passes `--signer-workflow` together with
`--cert-identity-regex`; real gh (2.88.1) rejects that ("if any flags in the group [cert-identity cert-identity-regex signer-repo
signer-workflow] are set none of the others can be"), so every install with a current gh would refuse. The test's fake gh accepts the
combination (installer_script_test.go:85-86 pins it), so the rows missed it.
1. gh branch: pass exactly one signer pin — `gh attestation verify <file> --repo "$REPO" --signer-workflow "$SIGNER_WORKFLOW"
   --cert-oidc-issuer "$OIDC_ISSUER"` when --signer-workflow is supported, else `--repo "$REPO" --cert-identity-regex …
   --cert-oidc-issuer …`. Confirm the --signer-workflow value format gh expects (check `gh attestation verify --help` on this host) and
   use it.
2. Make the fake gh in the tests model real gh: it must REJECT mutually exclusive signer flags with a non-zero exit (so the old bug fails a
   row), and assert the exact argv the script passes for each branch. Add a mutant: both flags passed → the row fails.
3. If gh is installed on this host, run the real `gh attestation verify --help` and one real invocation shape against a dummy file to show
   the flags are accepted (exit code of flag parsing, not of verification).
Set status development; update results; handoff; END YOUR TURN. No CHANGELOG/LOGBOOK edit.
