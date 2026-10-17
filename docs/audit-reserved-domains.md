# Reserved domains in source audit

The deterministic HTTP(S) source detector omits
`audit.capability.network-undeclared` for the four top-level names reserved by
[RFC 2606 section 2](https://www.rfc-editor.org/rfc/rfc2606.html#section-2):
`test`, `example`, `invalid`, and `localhost`, plus their DNS descendants.
The [manager profile section 7](https://github.com/relux-works/curator-spec/blob/1059b94b4135089d842bf017070b8ab47d79a536/profiles/manager.md#7-source-audit-policy)
allows implementations to choose their deterministic detectors. Its strict
decision semantics, canary, and revocation checks still apply.

Matching ignores ASCII case and accepts one optional terminal DNS root dot.
For example, `board.example.test`, `Board.Example.TeSt.`, and
`service.localhost` qualify. Names must have nonempty ASCII DNS labels, each
at most 63 bytes, with no leading or trailing hyphen, and at most 253 bytes
overall after removing the root dot. A second trailing dot does not qualify.
The last label must match exactly: `evil.test.com`, `example.com.attacker`,
and `notest` remain audited.

The detector parses the complete matched URL authority before applying this
exemption. Ports do not change the host. In
`https://board.test@github.com/data`, the authority host is `github.com`, so
the reserved-looking username does not silence its finding. A malformed
authority does not earn the new exemption. Existing explicit declarations
and the historical loopback-address exception retain their behavior.

The separate second-level reservations `example.com`, `example.net`, and
`example.org`, including their descendants, are outside this exemption.
They remain audited, including the static canary's
`exfiltrate.example.com`. Real hosts such as `github.com` and `brew.sh`
continue to require covering declarations. A script containing both a
reserved fixture and an undeclared real host still blocks in strict mode.

This is a source-analysis convention, not a runtime network control or a
claim that reserved names cannot resolve. In particular,
[RFC 6761 section 6.2](https://www.rfc-editor.org/rfc/rfc6761.html#section-6.2)
permits test names to resolve differently on private networks. The exemption
does not alter execution policy, DNS, source allowlists, revocations, or the
static canary. Audit ruleset version 3 separates new verdicts from cached
findings produced under version 2; old records are retained.

Run `go test ./internal/audit -run '^TestReserved'` for matching, refusal,
strict-gate, canary, and cache controls. Run
`go test ./internal/envprofile -run '^TestMigratedGlobalSkillReservedFixture$'`
for the production migration path with declared real hosts, a reserved
fixture, and an undeclared neighboring host. Keep local logs under `.temp/`.
