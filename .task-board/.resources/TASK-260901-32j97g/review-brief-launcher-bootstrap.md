# Review brief: curator-agent-launcher bootstrap

## Subject
Local repo `~/Developer/ReluxWorks/curator-agent-launcher`, branch main, heads `d3400fb` (skeleton) + `dae0c35` (SPEC draft). Producer notes on TASK-260901-32j97g. Remote exists but is EMPTY — nothing is pushed; do not push.

## Dimensions
1. **SPEC.md vs the normative sources**: Decision 0010 Decision 6 + Decision 10 (`~/Developer/ReluxWorks/curator-spec/decisions/0010-agent-environment-profiles.md`), `protocol/environments.md` §10 fragment + §7.3 channels + umbrella convention, and skill-agents-management SKILL.md (gh api). Check: boundary honored (no session state, plans consumed never rebuilt, curator/ax via CLI only, agents-management module as planned dependency); composition algorithm correct incl. env-merge conflict rule and fragment closed-name wins; ax ALWAYS-when-configured (config change to bypass, not a flag); system-prompt opt-in + all three warning requirements referencing (not restating) environments.md; diagnostics families sane; versioning + promotion note per Decision 6.
2. **Flag surface**: minimal and justified; pass-through after `--` verbatim; nothing that duplicates spawn-plane or curator concerns.
3. **Skeleton hygiene**: go build/vet/test green (run them); module path `github.com/relux-works/curator-agent-launcher`; go 1.23; stub imports nothing beyond stdlib; LICENSE byte-equal to curator's Apache-2.0; NOTICE shaped like curator's naming this project; README status honest (spec draft, not implemented); Makefile targets work; .gitignore sane; both commits signed.
4. **English + ecosystem prose style** for SPEC.md.

## Verdict
`review-findings-launcher-bootstrap-1.md` on TASK-260901-32j97g; blocking/major -> development; else ACCEPT explicit, leave to-review. Do not push or mark done.
