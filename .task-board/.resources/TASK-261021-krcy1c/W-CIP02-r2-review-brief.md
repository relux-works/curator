# Delta review: TASK-261021-krcy1c rework r2, CIP-0002 (curator-spec)

Your r1 finding F1 (no root/module entry selection for project context items) is addressed at the new head in the request, on the same branch `cip02-ready-for-acceptance`. The producer inherits schema-2 context packages and schema-4 locks, and defers bare files and native import dialects.

1. Review the delta `752307c6..<new head>` first, then re-sweep the surfaces it touches.
2. F1: inheritance must be explicit and complete: module selection, deterministic ordering and rendering, byte validation and import closure all bound into the approval digest. Two packages that differ only in root or module selection must give different approval digests: check the producer's witness, and build your own counterexample.
3. Is deferring bare repository instruction files (a single AGENTS.md with no manifest) consistent with the owner decisions and the first-wave goal? If it leaves the first slice unable to admit any real project's instructions, that is a finding: say what minimal subset would close it.
4. No regression elsewhere. Rerun make validate in a normal-.git clone (split long groups if you need to).
