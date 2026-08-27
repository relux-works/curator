# TASK-260728-168smo probe tree

Reproduces every measurement in `TASK-260728-168smo_command-evidence.log`.
Nothing here installs, downloads, or switches anything.

- `run-evidence.sh <log> <argv...>` — runs one command standalone and appends
  its argv, real exit code, byte counts and bounded output to `<log>`.
- `shims/` — 15 logging shims that record their name and argv and exit 127.
  Point `PATH` at this directory (plus `/bin` for a real shell when running the
  control) to count `PATH` resolutions.
- `argprobe/` — the source fixtures:
  - `main.kt`          plain Kotlin source
  - `evil.kts`         script payload, marker-writing
  - `exec.kts`         script payload used for the escalation case
  - `@resp.txt`, `atfile`, `@abs`, `inject` — response-file fixtures

## The four decisive runs

    # 1. response-file expansion executes package code, exit 0
    cd argprobe && printf -- '-script\n' > inject
    kotlinc @inject exec.kts        # writes EXEC-MARKER, exit 0

    # 2. the same path in absolute or ./ form does not expand
    kotlinc "$PWD/@abs" exec2.kts   # exit 1, "source entry is not a Kotlin file"

    # 3. control: the shipped launcher resolves 4 names from PATH
    env -i HOME="$HOME" PATH="$PWD/../shims:/bin" \
      <kotlin_root>/bin/kotlinc -d out "$PWD/main.kt"

    # 4. pinned: the JDK's own java resolves 0
    env -i HOME="$HOME" PATH="$PWD/../shims" \
      <jdk_root>/bin/java \
        -cp <kotlin_root>/lib/kotlin-preloader.jar \
        org.jetbrains.kotlin.preloading.Preloader \
        -cp <kotlin_root>/lib/kotlin-compiler.jar \
        org.jetbrains.kotlin.cli.jvm.K2JVMCompiler \
        -d out3 "$PWD/main.kt"

The shims write to `../shimlog/pinned.txt`; edit that path in `shims/*` to
separate control and pinned runs.
