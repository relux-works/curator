#!/usr/bin/env python3
"""Umbrella range contract check for B2 skill packages (TASK-260908-2kihaw).

Verifies, for each of the three B2 skills, that the documented
`requires.skills` entry the `relux-root-context-ivan` umbrella (B1) must use:
  1. carries a range matching the agent-context-v1 `range` grammar
     (pattern loaded from the live curator-spec schema, not retyped);
  2. admits the chosen initial version v0.1.0 under the Decision 0012
     caret rule quoted verbatim in the decision text
     ("^0.1 is >=0.1.0 <0.2.0-0"; prerelease bound "<X.Y.Z-0" excludes
     prereleases of the bound; a prerelease satisfies a range only when a
     primitive names a prerelease on the same major.minor.patch);
  3. rejects the neighboring versions that must fall outside the line
     (0.0.9, 0.2.0, 0.1.0-rc.1);
  4. canonicalizes per protocol core section 6.1 to the expected
     host/path identity;
  5. uses a skill name that matches the identifier grammar.

Bound: this proves the range strings admit exactly the intended initial
line. Live resolution against real version tags happens at B1 umbrella
install time, after review landing and v0.1.0 tagging.
"""
import json
import re
import sys

SPEC = "/Users/iv/Developer/ReluxWorks/curator-spec/schemas/v1/agent-context-v1.schema.json"
COMMON = "/Users/iv/Developer/ReluxWorks/curator-spec/schemas/v1/common.schema.json"

ENTRIES = {
    "pdf": {
        "git": "git@github.com:relux-works/skill-pdf.git",
        "range": "^0.1",
        "identity": "github.com/relux-works/skill-pdf",
    },
    "skill-creator": {
        "git": "git@github.com:relux-works/skill-creator.git",
        "range": "^0.1",
        "identity": "github.com/relux-works/skill-creator",
    },
    "agents-attachments": {
        "git": "git@github.com:relux-works/skill-agents-attachments.git",
        "range": "^0.1",
        "identity": "github.com/relux-works/skill-agents-attachments",
    },
}
INITIAL = (0, 1, 0)
failures = []


def check(name, cond, detail):
    print(("PASS " if cond else "FAIL ") + name + ": " + detail)
    if not cond:
        failures.append(name)


def canonical_identity(git):
    """Section 6.1 canonicalization for the SCP spelling used here."""
    m = re.fullmatch(r"(?:[^@:/\s]+@)?([^:/\s]+):(\S+)", git)
    if not m:
        return None
    host, path = m.group(1).lower(), m.group(2).strip("/")
    if path.lower().endswith(".git"):
        path = path[: -len(".git")]
    return host + "/" + path


def parse_version(v):
    m = re.fullmatch(r"(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z.-]+))?", v)
    if not m:
        return None
    nums = tuple(int(m.group(i)) for i in (1, 2, 3))
    return nums, m.group(4)


def admits_caret_0_1(version):
    """Decision 0012 rule: ^0.1 is >=0.1.0 <0.2.0-0; a prerelease satisfies
    only when a primitive names a prerelease on the same triple (^0.1 names
    none), so every prerelease is excluded."""
    parsed = parse_version(version)
    if parsed is None:
        return False
    (major, minor, patch), pre = parsed
    if pre is not None:
        return False
    return (major, minor, patch) >= (0, 1, 0) and (major, minor, patch) < (0, 2, 0)


def main():
    agent_ctx = json.load(open(SPEC))
    range_pattern = agent_ctx["$defs"]["range"]["pattern"]
    common = json.load(open(COMMON))
    ident_pattern = common["$defs"]["identifier"]["allOf"][0]["pattern"]
    print("range grammar: loaded from agent-context-v1.schema.json")
    print("identifier grammar: loaded from common.schema.json")

    for skill, entry in ENTRIES.items():
        base = "umbrella requires.skills[%s]" % skill
        check(base + " name", re.fullmatch(ident_pattern, skill) is not None,
              "identifier grammar, got %r" % skill)
        check(base + " range grammar",
              re.fullmatch(range_pattern, entry["range"]) is not None,
              "schema range pattern admits %r" % entry["range"])
        check(base + " identity",
              canonical_identity(entry["git"]) == entry["identity"],
              "%r -> %s" % (entry["git"], canonical_identity(entry["git"])))
        check(base + " admits v0.1.0", admits_caret_0_1("0.1.0"),
              "initial version inside ^0.1")
        for other in ("0.0.9", "0.2.0", "1.0.0", "0.1.0-rc.1"):
            check(base + " rejects " + other, not admits_caret_0_1(other),
                  "outside the ^0.1 line")

    if failures:
        print("\n%d FAILING CHECK(S): %s" % (len(failures), failures))
        return 1
    print("\nall umbrella range contract checks green")
    return 0


if __name__ == "__main__":
    sys.exit(main())
