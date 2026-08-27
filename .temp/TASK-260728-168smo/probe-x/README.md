# TASK-260728-168smo probe package — rework cycle 2

Everything needed to reproduce the measurements behind decision 0010 rev 2 on a
macOS arm64 host. Nothing here installs anything permanently; the distribution
and its dependency closure are downloaded into a scratch directory and can be
deleted afterwards (they were, on the measuring host).

## 0. Prerequisites

- macOS arm64, a JDK (measured: OpenJDK 26.0.1), `curl`, `shasum`,
  `sandbox-exec` (present on stock macOS), about 2.5 GB of free disk.
- Xcode installed. Note that this is *not* a prerequisite of the design; it is
  the prerequisite whose necessity E13/E14 measure, and whose necessity is the
  reason macOS is unsupported.

## 1. Fetch and verify

```bash
BASE=$PWD/kn ; mkdir -p "$BASE" && cd "$BASE"
curl -sSL -o kn.tar.gz \
  https://github.com/JetBrains/kotlin/releases/download/v2.4.10/kotlin-native-prebuilt-macos-aarch64-2.4.10.tar.gz
curl -sSL \
  https://github.com/JetBrains/kotlin/releases/download/v2.4.10/kotlin-native-prebuilt-macos-aarch64-2.4.10.tar.gz.sha256
# expected: 55ded039bb56a69aec9df354a92b42df9e916104e3c53d8d9852d9cc6617ed9d
shasum -a 256 kn.tar.gz
mkdir -p dist && tar -xzf kn.tar.gz -C dist
```

Then edit `vars.sh` so `KN_BASE`, `KN_DIST` and `JDK_ROOT` point at this tree,
and `. ./vars.sh`.

## 2. The four decisive runs

| Run | Command | Expected |
|---|---|---|
| hydrate | `KONAN_DATA_DIR=$KN_BASE/konan-data "$JDK_ROOT/bin/java" -ea -Xmx3G -XX:TieredStopAtLevel=1 -Dfile.encoding=UTF-8 -Dkonan.home="$KN_DIST" -cp "$KN_JAR" "$KN_MAIN" konanc -produce program -target macos_arm64 -o out/app src/main.kt` | exit 0, downloads 688 MB, produces `out/app.kexe` |
| contain | the same under `sandbox-exec -f sb-contain.sb` | **exit 2**, `Cannot run program "/bin/bash"` at `Xcode.kt:144` |
| enumerate | `./enumerate2.sh` | grows the allow-list to the seven external executables, then exit 0 |
| override | the `-Xoverride-konan-properties=…` run under `sb-override.sb` (argv in `command-evidence-cycle2.log` E14) | **exit 2**, `Cannot run program "/usr/bin/xcrun"` at `Xcode.kt:131` |

## 3. The bundle-model runs

| Run | What it shows |
|---|---|
| `sb-nonet.sb` + hydrated closure | exit 0, no download — E8 |
| aggregate SHA-256 of the data dir before/after | unchanged — E9 |
| `chmod -R a-w` on the data dir | exit 2 on `dependencies/cache/.lock` — E10 |
| symlink overlay + both roots read-only | exit 0 — E11, the model decision 0010 section 4 fixes |
| `-Xoverride-konan-properties=airplaneMode=true` with a missing dependency | exit 2, no download — E12 |

## 4. Contents

```
command-evidence-cycle2.log   the consolidated record, E1-E19
evidence/run*.log             raw stdout+stderr of every run
evidence/allowed-externals.txt the enumerated external executable set
enumerate.sh, enumerate2.sh   the iterative exec-containment enumerator
sb-*.sb                       the sandbox profiles used
src/main.kt                   the hello-world under test
vars.sh                       path variables; edit before use
```

The platform-library case of E18 is two extra imports in `src/main.kt`:

```kotlin
import platform.posix.getpid
import platform.Foundation.NSProcessInfo
```

which compile and link with no `-library` and no `.def`, and add
`/usr/lib/libresolv.9.dylib` to the artifact's dynamic dependencies.
