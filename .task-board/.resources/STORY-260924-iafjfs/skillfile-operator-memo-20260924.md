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
