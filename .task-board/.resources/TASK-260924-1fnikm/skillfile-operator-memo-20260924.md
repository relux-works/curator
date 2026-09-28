# Operator memo 2026-09-24 — Skillfile schema 2 sources and selectors (binding acceptance for STORY-260924-iafjfs and its leaves)

Implement and release Skillfile schema 2 sources and selectors as specified in curator-spec protocol/skillfile-sources.md:
repository sources pinned by tag or revision; individual {name, from, directory} and collection
{from, directory, include, exclude} selectors. Add the same `directory` selection to `dependencies.skills` in skill manifests
(curator-spec amendment to core §4.4). Why: process playbooks ship as one repository holding several skills
(skills/orchestrator, skills/developer, ...) and may depend on role skills that live in subfolders of other repositories
(relux-works/curator-playbook spec/process-configuration.md §8).
Done when: one Skillfile entry {from: <playbook repo>, directory: "skills", include: ["*"]} installs every skill in that folder
through install/update with lock and audit; a manifest dependency carrying `directory` installs a skill from a subfolder of
another repository; conformance vectors and release notes cover both; shipped in the next curator and curator-spec releases.
Much may already be done — in any case finish it and enable it by default.

## Operator note 2026-09-24 (partial client)
cocoaskills stays on the rc.10 core suite + accepted skillfile-sources; it will not adopt rc.11/rc.12 changes. Keep skillfile-sources text and
vectors consumable independently of rc.11/rc.12 core. Review criterion for the external promotion PR (1fnikm) and every skillfile-sources change:
no draft $ref to a v1 definition absent/different at v1.0.0-rc.10; no prose dependency on post-rc.10 core/registry/manager/environments clauses;
the suite stays a separately pinnable directory (not merged into conformance/v1). Guard leaf in curator-spec (STORY-260924-360z95).
