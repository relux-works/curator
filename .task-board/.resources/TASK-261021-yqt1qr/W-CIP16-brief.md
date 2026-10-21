# Brief: CIP-0016 amendment, delegated local launches (curator-spec)

**Repository:** curator-spec. **Branch:** `cip16-delegated-launches`, from current `origin/main`.

## Context

The owner accepted, on 2026-10-21, the transitional delegated launch route proposed by the curator-run lane. A registered session manager may host interactive sessions by composing against the live managed profile home until exact-plan hosting on snapshots exists. The proposed amendment text is attached as `front-door-appendix-A.md`: the curator-run FRONT-DOOR.md, Appendix A.

## Do this

1. Read GOVERNANCE.md, the CIP process notes, `cips/CIP-0016-local-hosted-launches.md`, `protocol/local-hosted-launches.md` and `cips/CIP-0014-managed-home-snapshots.md`.
2. Add the amendment to CIP-0016 as a dated revision section, following how other CIPs record accepted amendments, for example the CIP-0002 and CIP-0013 revision blocks. Record the owner acceptance date, 2026-10-21. Keep the text's substance: the division of work, the stated limits, the retained rules, per-runtime activation and the retirement condition. Edit for the repository's prose rules (plain English, no em dashes, no guillemets).
3. Add a clearly separated subsection **"Proposed conditions (pending operator confirmation)"** with two conditions:
   - (a) **Home lease.** While a hosted session runs in a profile's managed home, repair, profile switch and skill sync targeting that home refuse with a typed error and remediation, or wait. They never change the home underneath the session.
   - (b) **One profile per Werkstatt.** A Werkstatt's sessions use a profile home that no other Werkstatt or ad-hoc launch shares.

   State that these are not yet accepted.
4. If `protocol/local-hosted-launches.md` needs normative text for the delegated route, or a pointer to it, add the minimum, cross-referenced to the CIP. Do not invent new wire schemas: the request and record schemas belong to curator-run.
5. Update CHANGELOG.md and README.md only where this repository's own rules require it for an amendment (check how earlier amendments did it).
6. Self-check: links resolve, section numbers are right, no personal paths or names.

## Hand-off

Results file: the files changed, a one-paragraph summary, and open questions. Branch pushed, signed commits.
