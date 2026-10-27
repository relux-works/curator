# Source commit provenance repair delivered

[PR202](https://github.com/relux-works/curator-spec/pull/202) landed at exact signed head `43d5831adb77f58e756cbdfa7eabffc1454922e8` over `74a7dcc503c0fd7acd491299ab71290ae40e708e`. Six human-signed appended commits verify locally and remotely; the hosted PR and default ref were verified after non-forced fast-forward landing.

[Independent security review r4](https://github.com/relux-works/curator-spec/pull/202#pullrequestreview-5675212781), produced with Codex gpt-6-astra and normal exit 0, accepts the complete twelve-path delta. All eight required PR checks pass. Actual Windows Specification job122496184268 reports1462 tests in1780.704s, OK with one existing skip.

The NEW actual main push [run40598819012](https://github.com/relux-works/curator-spec/actions/runs/40598819012), provenance job122521499257, completed SUCCESS. Its official checkout is exact43d5831adb77f58e756cbdfa7eabffc1454922e8 and its validator reports that source target has an authorized signature on origin/main. The prior failed main run40569031707 remains a historical failure; it was neither rerun nor reclassified. Other jobs of the new main run remain separate and are not declared complete here.

Source signing trust is separate from release-tag authority. Original maintainers.allowed_signers is byte-identical with SHA25612f4bfaf1bfcd954d30b0683f6b55b8ad7e19e5f60c88c95dbf0af702db82dc8. No tag, key rotation, profile adoption, installation or host activation occurred. Dependent fixture integration now resumes against fresh main with its own renewed review and native gates.
