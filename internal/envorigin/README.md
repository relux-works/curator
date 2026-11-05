# Provisioning-origin record codec

This package consumes the canonical standalone `provisioning-origin/1` schema
and record corpus landed on curator-spec main through PR182, commit
`6c7ce40301e254e14e55ebd29405c90fe00c79c6`:
<https://github.com/relux-works/curator-spec/pull/182>.

`schemas/provisioning-origin-v1/provisioning-origin-v1.schema.json` and
`testdata/schema-cases.json` are exact copies from that commit. Independent
source-owner review accepted this specification. That acceptance does not
qualify this implementation, certify a release or authorize live adoption.
Existing home marker schemas are unchanged.

The codec rejects malformed transport, duplicate keys, unknown members and
invalid complete records. Callers still own current binding, filesystem
identity, owner-private storage, authority refresh and recoverable lifecycle.
The parser alone supplies none of those guarantees.

Run `go test ./internal/envorigin` for the canonical record cases and transport
refusals. Store test logs under `.temp/`.
