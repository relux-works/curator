# Draft repository endpoint policy boundary

This is an internal implementation of unreleased `repository-transport-v1`
revision 1 (machine policy loading and attempt planning), not an enabled
fetch path. Frozen v1 and release qualification are unchanged.

`config.LoadSourcePolicy` reads the operator-owned `source-policy.json`
beside the manager configuration; `config.ParseSourcePolicy` validates
one document. Entries are keyed by exact canonical identity
(`identity.DraftCanonicalKey`, no normalization); each lists one or two
distinct closed-grammar endpoints that canonicalize to the key, an opaque
operator provider reference per endpoint (`gitcred.ValidProvider`, never
a command or path), and a closed fallback mode. `pin` selects exactly one
listed URL by exact string equality and forces no fallback.

`config.ResolveRepositoryEndpoints` plans attempts for one declaration
without network I/O: a URL without an entry attempts the declared URL
once with no policy provider; a logical identity without an entry fails
`repository_endpoint_unavailable`; an invalid or unreadable policy fails
`repository_policy_invalid` before any acquisition. A schema-2 document
is rejected closed: revision 2 ports, mirrors, and aliases need a
revision-2 reader and are a separate leaf.

This layer plans only. Failure classification, credential resolution,
and fetch attempts belong to transport application; root-input existence
is proven at acquisition time, not at load.
