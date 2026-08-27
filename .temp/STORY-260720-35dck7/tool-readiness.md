# Tool readiness — STORY-260720-35dck7

Checked 2026-07-20 from `/Users/iv/Developer/ReluxWorks/curator`.

- `task-board`: `/Users/iv/.local/bin/task-board`, version `0.20.1-15-g5e2c927`; status mutation, schema query, resource materialization, and planning queries succeeded.
- `rg`: version `15.1.0`; repository and artifact discovery succeeded.
- `git`: Apple Git `2.50.1`; `curator-spec` fetch, `origin/main` resolution, and read-only tree inspection succeeded.
- `jq`: version `1.8.1`; post-mutation board completeness projection succeeded.
- `java`: OpenJDK `26.0.1`; task-local PlantUML `1.2026.6` starts successfully.
- PlantUML: `-checkonly` succeeds for the existing protocol-v6 component and activity sources using `!pragma layout smetana` where needed.
- Graphviz `dot`: unavailable for rendering because `/opt/homebrew/opt/libtool/lib/libltdl.7.dylib` is missing. Do not use the Graphviz-dependent board renderer in this audit; use the verified task-local PlantUML/Smetana path.

This is a planning-only tool check. No product or specification source was changed.
