# Review note — TASK-260924-5c0747 curator v0.15.0-rc.2 release prep (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review revision 3 (14 paths on trunk f02ba39e, gate green) against `5c0747-brief.md` and `5c0747-rework-3.md`. Bounded runs (host memory).
1. Precedent: same shape as the v0.15.0-rc.1 cut (d1bb0a4e and neighbours); deviations named and justified.
2. SPEC_PIN = curator-spec v1.0.0-rc.13 tag commit 23435129ebc4c29e5b7f75ec72a0aa0cd3f16065 everywhere (ci.yml, tests, docs, ledgers); no
   dcc7f015 left except in history text; the rc.13 protocol_version labels accepted only where 4hd81z allows.
3. Vendored conformance/skillfile-sources-v1 pinned to the rc.13 tag commit; bytes identical to the 5746367 vendoring except the new
   manifest.json/README (show the diff).
4. CHANGELOG v0.15.0-rc.2 section: every entry is VERBATIM from a "## CHANGELOG entry (for release prep)" results section of a leaf that
   landed after v0.15.0-rc.1 — spot-check at least 8 against their resources, confirm h4syhu/2as5sx/187z6x/20o9dk/cww1ov/m28s6b/1aa9wb/
   11burj/2gt5f6/2elcdc present; no entry for a leaf that has not landed (em42lw is NOT landed); nothing lost from existing sections.
5. The Windows Job-flags failure of rev1 (TestEnforcedInstallAndLaunchAtCLIEntry, flags 0x0) did not recur in rev2/rev3 — confirm from the
   validation logs and say whether it is a known flake (hwx26) or needs a leaf.
6. Results give the exact signed-tag command and the release workflow to watch. accept_cr or changes requested. No LOGBOOK.md.
