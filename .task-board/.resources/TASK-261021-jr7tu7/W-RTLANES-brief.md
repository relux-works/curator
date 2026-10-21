# Brief: editorial clarification, which Git acquisition lanes repository-transport covers (curator-spec)

**Branch:** `repository-transport-lanes`, from `origin/main`. Editorial only: no behaviour change.

`protocol/repository-transport.md` says that it applies "to new Skillfile sources and to the existing Git acquisition lanes stated below", and revision 2 says "the existing Git acquisition lanes implemented under this protocol". It never enumerates them. A Curator bug (canonicalizing a declared ssh skill dependency to https and fetching over https) has shown why the list matters: §1 "a supported `git` URL supplies identity and a default connection hint, not a transport pin" and §2 "attempt the declared endpoint once" must clearly apply to profile closure members (manager §12.3: profile install, overlays and requirements) too.

Do this:
1. Find every Git acquisition lane in the spec: Skillfile sources, profile installs and closure members (manager §12.3), external skill build repositories (manager §11), context packages, `agent-mcp.json` packages if they are git-sourced, and anything else. Cite the sections.
2. Add a short normative "Covered lanes" list to `protocol/repository-transport.md` §1 (and point to it from revision 2 §4), naming each lane with its section, plus any lane explicitly not covered and why. It must not widen any lane's admitted grammar.
3. If a lane's own section contradicts §1 or §2 (for example by implying canonical-https fetch), list the contradiction in the results file and don't fix it silently.
4. Validator, prose rules, links.
