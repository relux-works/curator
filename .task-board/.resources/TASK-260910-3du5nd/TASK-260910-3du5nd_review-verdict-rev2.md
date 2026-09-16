# TASK-260910-3du5nd — independent revision 2 review

VERDICT: ACCEPT

Candidate: CR-TASK-260910-3du5nd-2, tree 5c9979fe1601e270ddd653953f7987e845ca5994, base a68854d54725862f6f696019ef2e569f3ec29cd6.

Reviewed all 25 changed files; each changed file byte-matches the candidate blob. Git diff against the candidate reports no tracked worktree delta. No candidate code or specification was modified. Compared the prior patch's blob objects: rework changes only the transport document, semantic cases, index, and the newly added SSH refusal fixture. No unrelated rework.

## Prior findings resolved

1. Alias/mirror admission: section 5 now says “mirror_of MUST be present if and only if the resolved connection host differs from the key host” (Markdown quoting omitted). It defines resolved host as alias target when present, otherwise URL host. Identity remains the entry key; every canonicalized path must match. Same-host aliases forbid attestation; mirror URLs combined with aliases are explicitly refused. Section 6 uses the same predicate before pin selection. valid-alias.json, v2-alias-resolution and v2-alias-mirror-undeclared agree with this rule.
2. Strict external acquisition: section 4 explicitly limits the external-build lane to the “port-free, alias-free subset”; section 5 separately forbids port and alias selection in that lane. Section 7 says “Revision 2 does not extend the section 11.2 URL grammar or the section 11.3 SSH wrapper” and requires build_repository_identity_invalid before network I/O, forbidding stripping the port or ignoring the alias. This agrees with unchanged profiles/manager.md 11.2/11.3. v2-external-build-port-refused and v2-external-build-alias-refused state zero attempts; v2-external-build-mirror-admitted preserves the ordinary port-free mirror lane.
3. SSH upper bound: indexed invalid-port-range-ssh.json contains SSH ports 65536 and 65539. Independently widened only the SSH regex branch from 6553[0-5] to 6553[0-9] in memory: original refuses the fixture and mutant admits it, detecting 1/1 targeted mutant. Both HTTPS and SSH independently pass six boundary probes each (1, 65535, 0, 65536, 65539, 01): 12/12.

## Scope and compatibility

Revision 1 sections 1–3 remain identical through the final scope sentence; only title, revision scoping and closing unresolved pointer change. Schema 1 is byte-unchanged. Schema 2 is additive with its new version discriminator; frozen wire schemas, manager profile and release artifacts are unchanged. Identity/allowlists remain canonical; exact pin equality includes ports; mirrors do not relax locked object checks. Two endpoints maximum, one attempt each, no generated candidates, and the fail-closed fallback table remain intact. Authentication equality, no alias chaining, operator-only resolution, private sanitized provenance and broker-held secrets are specified. README, compatibility, changelog, schema/conformance indexes and unresolved questions track the amendment.

## Independent verification and limits

- Ran the draft README's exact checker using the existing worktree venv: exit 0. Schema cases 115/115, negatives 91/91, wire schemas 8/8; snapshot vectors 3/3; existing marker narrowing mutants 18/18 detected.
- Targeted SSH mutant: 1/1 detected; port boundaries 12/12. These drive Draft202012Validator.is_valid, not a manager resolver.
- git diff --check: exit 0.
- Manager semantic execution: 0 cases. The 21 v2 semantic cases are reviewed normative examples, not executed implementation tests. No manager implementation or cross-platform conformance claim is made.
- Two preliminary review assertions were corrected: raw whole-tree hashing encountered a pre-existing CRLF fixture/filter difference (tracked Git diff is clean), and the revision-1 text comparison initially used the new closing phrase as the old delimiter. Exact changed-file blob comparisons and a corrected stable-delimiter comparison passed; neither was a candidate defect.

Full validation output is attached separately as TASK-260910-3du5nd_review-validate.log. Findings are recorded here on the board instead of LOGBOOK.md, which campaign instructions prohibit editing. No remaining blocking review findings; acceptance routes to integrating, not done.

Independent full gate: `PATH="$PWD/.temp/venv/bin:$PATH" make validate` in zsh exited 0: validated 60 schemas and 1047 vector files; 227 Python unit tests passed in 188.765 seconds; `go test ./tools/...` passed. This was rerun independently as explicitly required by the revision-2 review brief; no producer pass was substituted for this run.
