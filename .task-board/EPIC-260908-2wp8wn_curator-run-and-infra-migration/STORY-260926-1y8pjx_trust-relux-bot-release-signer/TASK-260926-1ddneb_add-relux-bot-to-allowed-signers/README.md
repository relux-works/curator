# TASK-260926-1ddneb: add-relux-bot-to-allowed-signers

## Description
Operator decision 2026-09-26: add the Relux Bot signing key as a trusted release signer. Append exactly one line to maintainers.allowed_signers: principal bot@relux.works (the email Relux Bot signs commits and tags with) followed by its public key: ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPG7xTX05HL1XaD4XLUk0/TTeqRNHbMj5HdnqNQdDTID . Keep the existing oparin@me.com line byte-identical. Update GOVERNANCE.md / RELEASE.md only where they enumerate the trusted signers or maintainers (state that the Relux Bot key is an automation signer authorized by the operator on 2026-09-26). CHANGELOG entry under Unreleased.

## Scope
(define task scope)

## Acceptance Criteria
maintainers.allowed_signers = the existing line + the bot@relux.works line; ssh-keygen -Y verify (or git verify-commit with gpg.ssh.allowedSignersFile set to the candidate file) succeeds for a Relux Bot-signed commit and still for the existing maintainer key; docs consistent; required checks green
