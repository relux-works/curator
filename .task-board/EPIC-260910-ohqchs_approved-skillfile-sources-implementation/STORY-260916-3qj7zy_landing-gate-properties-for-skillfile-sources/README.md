# STORY-260916-3qj7zy: landing-gate-properties-for-skillfile-sources

## Description
From skill-project-management issue #296 (the 20-revision selector/frontmatter leaf): the producer suite was green on all twenty rejected revisions, so what closes defect classes is properties run by the landing gate, not producer self-attack. Add four property suites to the gate and convert the committed reviewer example tests into class tests parametrised over the attack-surface catalog families.

## Scope
curator landing gate (test-gate lanes), Skillfile source expansion, frontmatter reader, identity hashing, error surface

## Acceptance Criteria
Boundary property: an audit hook per test process asserts no open/scandir/listdir/glob/rglob names a path outside the resolved root over generated traversal families (symlink out, cycle, hard link, case variant, NFC/NFD, swap mid-walk, permission error, second walker); identity property: for any bytes, identity equals the hash of exactly those bytes or the call refuses; oracle soundness: every accepted frontmatter document yields the reference implementation values over a generated corpus; structured-error property: each error type injected at each filesystem call site yields a structured refusal through the public entry point; committed reviewer tests re-expressed as class tests over the catalog families and green in the gate
