# Review note — TASK-260910-2t0iun installer release verification (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev3 (base 3f60f7f0, tree 18353c49, 6 paths incl. install.sh and the two CI ledgers, gate green on every lane) against
`2t0iun-brief.md`, `2t0iun-decision-1.md` (keyless cosign identity; gh attestation first; refuse when no verifier; loud
CURATOR_INSTALL_INSECURE_SKIP_VERIFY=1 opt-out) and `2t0iun-gatefix-2.md`. Verify:
1. install.sh order: gh attestation verify (repo pinned; signer workflow pinned if supported) → else cosign verify-blob with
   --certificate-identity-regexp pinned to relux-works/curator release.yml on refs/tags/v* and the GitHub Actions OIDC issuer → then the
   archive sha256 against the VERIFIED checksums.txt → install. Any failure refuses with a clear message; no silent downgrade; the opt-out
   prints a loud warning and is documented; nothing is written to the install location before verification succeeds.
2. The identity regexp cannot be satisfied by another repo/workflow/branch (anchored; dots escaped); the checksum match is exact (no
   substring/prefix match on file names).
3. Rows (local fixture, no network): valid → installs; tampered archive → refused; bad signature / wrong identity → refused; no tool →
   refused; opt-out → installs with warning. Mutants (signature check skipped; checksum check skipped; identity regexp unanchored) killed with
   real exit codes. Windows: skip registered via skip-classes.tsv + platform-cases.tsv with the exact reason.
4. README install section and SECURITY.md updated; no CHANGELOG/LOGBOOK; no stray files.
accept_cr or changes requested with file:line. No LOGBOOK.md.
