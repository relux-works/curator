# Source provenance security rework r1

Ready for original independent successor review; not accepted or landed. Status and assignee remain unchanged.

PR202: https://github.com/relux-works/curator-spec/pull/202
Base: 74a7dcc503c0fd7acd491299ab71290ae40e708e
Reviewed predecessor: ec70b25b5c7b7b23be64a05dacd6534eee71ade3
Final successor: 3d5f838717e0609fa8905c37b13b117c8d974a00
Exact-head producer COMMENT: https://github.com/relux-works/curator-spec/pull/202#pullrequestreview-5674023152

R1-01 producer disposition: both consumers discard credential-bearing OS/URL/header and HTTP protocol-parser exceptions before emitting stable failures, with no raw exception chain. Controls exercise real urllib/http.client malformed-header validation before connect, exact refusal, and generated-credential absence in helper/public diagnostics. Transport cases include BadStatusLine. Nearby valid in-memory HTTP responses succeed; real301/302/303/307/308 redirects refuse. This is transport/state evidence, not TLS, external-service or native qualification. No live credential is a fixture input or evidence value.

All36 tests in the three named modules pass, zero skips. Exact old candidate bytes fail four controls/20 intended confidentiality assertions across helper and public-output channels, zero setup/import errors. Reject-all main plants fail both valid-response status assertions. Existing SSH source/maintainer acceptance, unknown/unsigned/uncontained/key/ref refusals, fixed authority, decoy identity/bytes and source-only tag rejection remain passing. Initial checker source-literal false positive is retained as red evidence; the corrected checker compares generated credential fragments.

R1-02 producer disposition: current main/open PR inventory confirms PR200 owns Decision0033 and no other PR occupies0034. Only this source-trust decision is renamed to Decision0034, including heading/four documentation references and current PR publication text. Foreign proposal/head/semantics are preserved. Complete remote patch reverse-check against the private candidate passed; source/checker pins retained.

Original release trust SHA25612f4bfaf1bfcd954d30b0683f6b55b8ad7e19e5f60c88c95dbf0af702db82dc8 is byte-identical. Source-only trust, workflows, packaging bindings, release gate and original trust-file tests are unchanged by this correction. All three source task commits retain human signatures G and exact producing-model attribution; original objects preserved. Stable unrelated checkout/head/index/untracked evidence are unchanged. Generated fixture keys/homes/temp/cache roots removed; recorded direct fixture PIDs exited. Scratch remains below48MiB stop/64MiB hard limits.

The intermediate producer COMMENT remains historical; only final successor is handed over. Current hosted/native gates and independent security review remain pending. Historical failed main run40569031707/job122417945263 stays failed; unchanged PR provenance skip is not recovery. New actual-main provenance requires later accepted exact-head landing by the integration owner. No landing/release/closure/goal operation, full/broad local run, hosted rerun/cancel/wait or new worker.
