# EPIC-260713-c12fbe: dedicated-website

## Description
A dedicated website for Curator, the agent environment manager, in the
relux.works visual style. relux.works will link to it as the product page.

## Scope
- Landing: what an AEM is, what Curator manages (skills, commands, MCP
  requirements, adapters, scopes, audit registry), install commands for every
  channel (brew, install.sh, scoop, deb/rpm, go install)
- The open protocol story: the Curator Specification
  (github.com/relux-works/curator-spec) as an independently implementable
  protocol; supply chain page (signed, notarized, attested releases,
  gh attestation verify one-liner)
- Docs section rendered from the repository (README, implementation plan,
  spec cross-links)
- Design language consistent with relux.works; relux.works showcase entry
  links here
- Domain choice (for example curator.relux.works) and the install script
  endpoint (get.relux.works/curator) decided during implementation

## Acceptance Criteria
- Site live and linked from relux.works
- Install commands copy-paste correct for every channel
- Spec and repository prominently linked
- Style review against relux.works passes
