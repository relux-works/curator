#!/bin/sh
# TASK-260729-osjeay — self-contained re-run of the §7.4 recipe validation.
# Revision 6 (rework cycle 5): extends the cycle-4 harness from 21 to 41 cases.
#
# Validates the `require-toolchain`, `test-linux`, `linux-package-guard`
# recipes, the §6.0a hosted toolchain-identity step, the §5.2 C2/C3 origin
# enumeration contract, the §5.2 C3 source-staging contract, and the §5.4
# W2/W3/W9 empty-root precondition, all as proposed for TASK-260720-1pvfj5
# in the final-CI execution map.
#
# NO GO IS EXECUTED. `go` is a /bin/sh stub that prints canned `version`,
# `env GOROOT`, `env GOTOOLCHAIN`, `env GOENV`, `list ./...` and `list -f`
# output, and can be told to fail mid-listing or to override the caller's
# environment the way a shim/wrapper launcher does.
#
# NO WINDOWS HOST IS CONTACTED. `ssh`/`scp` are /bin/sh stubs that return the
# documented cmd.exe / PowerShell REPLIES and STATUSES against a simulated
# remote base directory. The control-host block under test is verbatim; the
# cmd.exe and PowerShell syntax itself stays an unexecuted producer gate.
#
# Exit codes below describe THIS HARNESS, not the Curator suite.
#
# Usage:  sh .temp/TASK-260729-osjeay/verify-recipes.sh
# Expect: the summary table ends with "ALL 41 EXPECTATIONS MET" and exit 0.

set -u
WORK="${TMPDIR:-/tmp}/osjeay-verify-recipes.$$"
rm -rf "$WORK"
mkdir -p "$WORK/rootA/bin" "$WORK/rootB/bin" "$WORK/pin"
trap 'rm -rf "$WORK"' EXIT
echo '{}' > "$WORK/pin/manifest.json"

if command -v shasum >/dev/null 2>&1; then SHA256='shasum -a 256'; else SHA256='sha256sum'; fi

# ---------------------------------------------------------------- stub `go`
# STUB_VER            what `go version` reports
# STUB_GOROOT         what `go env GOROOT` reports
# STUB_FORCE_TOOLCHAIN  when set, `go env GOTOOLCHAIN` reports this REGARDLESS of
#                     the caller's environment — models a shim/wrapper launcher
#                     (goenv shim, corporate wrapper, container entrypoint) that
#                     sets GOTOOLCHAIN itself and overrides the caller.
# STUB_FORCE_GOENV    same, for `go env GOENV`.
# MODE=partialfail    `go list ./...` prints 3 of 4 packages, then fails.
cat > "$WORK/rootA/bin/go" <<'STUB'
#!/bin/sh
case "$1" in
  version) echo "go version ${STUB_VER:-go1.25.5} darwin/arm64"; exit 0;;
  env)
    case "$2" in
      # printf, not echo: sh's echo expands \r inside a Windows-shape root and
      # would silently corrupt the value the assertion under test compares.
      GOROOT)      printf '%s\n' "$STUB_GOROOT"; exit 0;;
      GOTOOLCHAIN) printf '%s\n' "${STUB_FORCE_TOOLCHAIN:-${GOTOOLCHAIN:-auto}}"; exit 0;;
      GOENV)       printf '%s\n' "${STUB_FORCE_GOENV:-${GOENV:-/Users/u/Library/Application Support/go/env}}"; exit 0;;
    esac
    echo "stub: unhandled env $2" >&2; exit 99;;
  list)
    if [ "$2" = "-f" ]; then
      echo "github.com/relux-works/curator/cmd/curator github.com/relux-works/curator/internal/godriver"
      echo "github.com/relux-works/curator/internal/install github.com/relux-works/curator/internal/godriver"
      echo "github.com/relux-works/curator/internal/godriver fmt"
      echo "github.com/relux-works/curator/internal/adapters fmt"
      exit 0
    fi
    echo github.com/relux-works/curator/cmd/curator
    echo github.com/relux-works/curator/internal/godriver
    echo github.com/relux-works/curator/internal/install
    if [ "${MODE:-ok}" = "partialfail" ]; then
      echo "load: parse internal/scopes/gc.go: syntax error" >&2
      exit 1                       # partial output, then failure
    fi
    echo github.com/relux-works/curator/internal/adapters
    exit 0;;
  test) echo "STUB go test $*"; exit 0;;
esac
echo "stub: unhandled: $*" >&2; exit 99
STUB
chmod +x "$WORK/rootA/bin/go"
printf '#!/bin/sh\nexit 0\n' > "$WORK/rootA/bin/gofmt"
chmod +x "$WORK/rootA/bin/gofmt"

# rootB is a SECOND, complete toolchain root. It exists so the cross-root cases
# are honest: rootB/bin/gofmt really is present and executable, which is exactly
# why the revision-4 check ("some gofmt exists under the reported GOROOT")
# accepted a mismatched pairing.
cp "$WORK/rootA/bin/go" "$WORK/rootB/bin/go"
cp "$WORK/rootA/bin/gofmt" "$WORK/rootB/bin/gofmt"

# rootW models a Windows hosted-runner root: the executables carry `.exe`, and
# the PATH entry is a HARD LINK to them under a different directory and name, so
# textual comparison cannot succeed and only the `.exe` / `-ef` arms can.
mkdir -p "$WORK/rootW/bin" "$WORK/winpath" "$WORK/winpath-badfmt" \
         "$WORK/shim" "$WORK/fmtonly"
cp "$WORK/rootA/bin/go"    "$WORK/rootW/bin/go.exe"
cp "$WORK/rootA/bin/gofmt" "$WORK/rootW/bin/gofmt.exe"
ln "$WORK/rootW/bin/go.exe"    "$WORK/winpath/go"
ln "$WORK/rootW/bin/gofmt.exe" "$WORK/winpath/gofmt"
ln "$WORK/rootW/bin/go.exe"    "$WORK/winpath-badfmt/go"
ln "$WORK/rootB/bin/gofmt"     "$WORK/winpath-badfmt/gofmt"
# a cross-root formatter that PATH puts ahead of the launcher's own
ln "$WORK/rootB/bin/gofmt"     "$WORK/fmtonly/gofmt"
# a goenv-style shim: a COPY, not a link, so it is not this root's launcher
cp "$WORK/rootA/bin/go"    "$WORK/shim/go"
cp "$WORK/rootA/bin/gofmt" "$WORK/shim/gofmt"

# go.mod the recipes read for the version-drift self-check
printf 'module github.com/relux-works/curator\n\ngo 1.25.5\n' > "$WORK/go.mod"
printf 'module github.com/relux-works/curator\n\ngo 1.25.4\n' > "$WORK/go.mod.drift"

# ------------------------------------------------- recipes, verbatim from §7
cat > "$WORK/Makefile" <<'MAKEEOF'
MODULE          := github.com/relux-works/curator
GODRIVER_PKG    := $(MODULE)/internal/godriver
GODRIVER_IMPORTERS := $(MODULE)/cmd/curator $(MODULE)/internal/install
GOTESTFLAGS     := -count=1 -timeout 30m
GO_VERSION_REQUIRED := go1.25.5
PIN_ROOT        ?= $(CURATOR_CONFORMANCE_ROOT)

GOROOT_EXPECTED ?=
GO              ?= $(if $(strip $(GOROOT_EXPECTED)),$(strip $(GOROOT_EXPECTED))/bin/go,go)
GOFMT           ?= $(if $(strip $(GOROOT_EXPECTED)),$(strip $(GOROOT_EXPECTED))/bin/gofmt,gofmt)

GOENVPREFIX := GOTOOLCHAIN=local GOENV=off
ifneq ($(strip $(GOROOT_EXPECTED)),)
GOENVPREFIX := GOROOT='$(strip $(GOROOT_EXPECTED))' $(GOENVPREFIX)
endif

.PHONY: require-pin-root require-toolchain test-linux linux-package-guard

require-pin-root:
	@test -n '$(PIN_ROOT)' || { echo 'PIN_ROOT required'; exit 1; }
	@test -f '$(PIN_ROOT)/manifest.json' || { echo 'PIN_ROOT is not a conformance root'; exit 1; }

require-toolchain:
	@set -u; \
	 want="go$$(awk '/^go[ \t]/{print $$2; exit}' go.mod)"; \
	 [ "$$want" = '$(GO_VERSION_REQUIRED)' ] \
	   || { echo "toolchain: go.mod requires $$want but the Makefile pins $(GO_VERSION_REQUIRED)"; exit 1; }; \
	 if [ "$(TOOLCHAIN_ALLOW_PATH)" != "1" ]; then \
	   test -n '$(strip $(GOROOT_EXPECTED))' \
	     || { echo 'toolchain: GOROOT_EXPECTED is required on a native host (see 5.0)'; exit 1; }; \
	   case '$(strip $(GOROOT_EXPECTED))' in /*) ;; *) echo 'toolchain: GOROOT_EXPECTED must be an absolute path; got: $(GOROOT_EXPECTED)'; exit 1;; esac; \
	   case '$(GO)'    in /*) ;; *) echo 'toolchain: GO must be an absolute path; got: $(GO)'; exit 1;; esac; \
	   case '$(GOFMT)' in /*) ;; *) echo 'toolchain: GOFMT must be an absolute path; got: $(GOFMT)'; exit 1;; esac; \
	 fi; \
	 goexe="$$(command -v '$(GO)' 2>/dev/null)" || { echo 'toolchain: GO=$(GO) not found'; exit 1; }; \
	 gofmtexe="$$(command -v '$(GOFMT)' 2>/dev/null)" || { echo 'toolchain: GOFMT=$(GOFMT) not found'; exit 1; }; \
	 test -x "$$goexe"    || { echo "toolchain: $$goexe is not executable"; exit 1; }; \
	 test -x "$$gofmtexe" || { echo "toolchain: $$gofmtexe is not executable"; exit 1; }; \
	 v="$$($(GOENVPREFIX) "$$goexe" version)" || { echo 'toolchain: `go version` failed'; exit 1; }; \
	 echo "toolchain: $$v"; \
	 case "$$v" in *' $(GO_VERSION_REQUIRED) '*) ;; *) echo "toolchain: expected $(GO_VERSION_REQUIRED), got: $$v"; exit 1;; esac; \
	 r="$$($(GOENVPREFIX) "$$goexe" env GOROOT)" || { echo 'toolchain: `go env GOROOT` failed'; exit 1; }; \
	 echo "toolchain: GOROOT=$$r"; \
	 case "$$r" in /*) ;; *) echo "toolchain: reported GOROOT is not absolute: $$r"; exit 1;; esac; \
	 if [ "$(TOOLCHAIN_ALLOW_PATH)" != "1" ]; then \
	   [ "$$r" = '$(strip $(GOROOT_EXPECTED))' ] \
	     || { echo "toolchain: GOROOT drift: reported $$r != approved $(GOROOT_EXPECTED)"; exit 1; }; \
	 fi; \
	 { [ "$$goexe" = "$$r/bin/go" ] || [ "$$goexe" -ef "$$r/bin/go" ]; } \
	   || { echo "toolchain: launcher $$goexe is not $$r/bin/go"; exit 1; }; \
	 { [ "$$gofmtexe" = "$$r/bin/gofmt" ] || [ "$$gofmtexe" -ef "$$r/bin/gofmt" ]; } \
	   || { echo "toolchain: formatter $$gofmtexe is not $$r/bin/gofmt"; exit 1; }; \
	 tc="$$($(GOENVPREFIX) "$$goexe" env GOTOOLCHAIN)" || { echo 'toolchain: `go env GOTOOLCHAIN` failed'; exit 1; }; \
	 echo "toolchain: GOTOOLCHAIN=$$tc"; \
	 [ "$$tc" = "local" ] \
	   || { echo "toolchain: GOTOOLCHAIN=$$tc, not local -- an implicit toolchain download is possible"; exit 1; }; \
	 ge="$$($(GOENVPREFIX) "$$goexe" env GOENV)" || { echo 'toolchain: `go env GOENV` failed'; exit 1; }; \
	 echo "toolchain: GOENV=$$ge"; \
	 [ "$$ge" = "off" ] \
	   || { echo "toolchain: GOENV=$$ge, not off -- a per-user go env file can inject GOFLAGS/GOTOOLCHAIN"; exit 1; }; \
	 echo 'require-toolchain: ok'

test-linux: require-toolchain require-pin-root linux-package-guard
	@rows="$$($(GOENVPREFIX) $(GO) list ./...)" \
	   || { echo 'test-linux: go list ./... failed; refusing a partial package set'; exit 1; }; \
	 pkgs="$$(printf '%s\n' "$$rows" | grep -v -x '$(GODRIVER_PKG)')"; \
	 test -n "$$pkgs" || { echo 'test-linux: safe package set is empty'; exit 1; }; \
	 excluded="$$(printf '%s\n' "$$rows" | grep -c -x '$(GODRIVER_PKG)')"; \
	 [ "$$excluded" = "1" ] \
	   || { echo "test-linux: expected exactly 1 excluded package, found $$excluded"; exit 1; }; \
	 total="$$(printf '%s\n' "$$rows" | grep -c .)"; \
	 kept="$$(printf '%s\n' "$$pkgs" | grep -c .)"; \
	 [ "$$kept" = "$$((total - 1))" ] \
	   || { echo "test-linux: $$kept kept of $$total listed; exclusion is not exactly one"; exit 1; }; \
	 echo "test-linux: $$kept of $$total packages (excluded: $(GODRIVER_PKG))"; \
	 $(GOENVPREFIX) CURATOR_CONFORMANCE_ROOT='$(PIN_ROOT)' $(GO) test $(GOTESTFLAGS) $$pkgs

linux-package-guard: require-toolchain
	@rows="$$($(GOENVPREFIX) $(GO) list ./...)" \
	   || { echo 'guard: go list ./... failed; cannot validate the Linux exclusion'; exit 1; }; \
	 printf '%s\n' "$$rows" | grep -q -x '$(GODRIVER_PKG)' \
	   || { echo 'guard: $(GODRIVER_PKG) no longer exists; the exclusion is stale'; exit 1; }
	@imports="$$($(GOENVPREFIX) $(GO) list -f '{{.ImportPath}} {{join .Imports " "}} {{join .TestImports " "}} {{join .XTestImports " "}}' ./...)" \
	   || { echo 'guard: go list -f failed; cannot validate the importer set'; exit 1; }; \
	 got="$$(printf '%s\n' "$$imports" \
	          | grep '$(GODRIVER_PKG)' | awk '{print $$1}' \
	          | grep -v -x '$(GODRIVER_PKG)' | LC_ALL=C sort | tr '\n' ' ')"; \
	 want="$$(printf '%s\n' $(GODRIVER_IMPORTERS) | LC_ALL=C sort | tr '\n' ' ')"; \
	 [ "$$got" = "$$want" ] \
	   || { echo 'guard: godriver importer set drifted'; echo "  got:  $$got"; echo "  want: $$want"; exit 1; }
	@echo 'linux-package-guard: ok'
MAKEEOF

# ------------------------- the revision-3 forms, for the masking demonstration
cat > "$WORK/Makefile.rev3" <<'MAKEEOF'
GO := ./rootA/bin/go
GODRIVER_PKG := github.com/relux-works/curator/internal/godriver
BAD_PKGS = $(shell $(GO) list ./... | grep -v -x '$(GODRIVER_PKG)')

bad-discovery:
	@echo "BAD_PKGS=[$(BAD_PKGS)]"
	@echo "recipe body ran: discovery did NOT fail closed"

bad-guard:
	@$(GO) list ./... | grep -q -x '$(GODRIVER_PKG)' || { echo 'guard: missing'; exit 1; }
	@echo 'bad-guard: reported ok'
MAKEEOF

# ------------------- the revision-4 require-toolchain, for the F1 demonstration
cat > "$WORK/Makefile.rev4" <<'MAKEEOF'
GO ?= go
GOFMT ?= gofmt

.PHONY: require-toolchain
require-toolchain:
	@if [ "$(TOOLCHAIN_ALLOW_PATH)" != "1" ]; then \
	   case '$(GO)'    in /*) ;; *) echo 'GO must be an absolute path; got: $(GO)'; exit 1;; esac; \
	   case '$(GOFMT)' in /*) ;; *) echo 'GOFMT must be an absolute path; got: $(GOFMT)'; exit 1;; esac; \
	   test -x '$(GO)'    || { echo 'GO=$(GO) is not executable'; exit 1; }; \
	   test -x '$(GOFMT)' || { echo 'GOFMT=$(GOFMT) is not executable'; exit 1; }; \
	 fi
	@v="$$($(GO) version)" || { echo 'toolchain: go version failed'; exit 1; }; \
	 echo "toolchain: $$v"; \
	 case "$$v" in *' go1.25.5 '*) ;; *) echo "toolchain: expected go1.25.5, got: $$v"; exit 1;; esac
	@r="$$($(GO) env GOROOT)" || { echo 'toolchain: go env GOROOT failed'; exit 1; }; \
	 echo "toolchain: GOROOT=$$r"; \
	 test -x "$$r/bin/gofmt" || { echo "toolchain: no gofmt under reported GOROOT $$r"; exit 1; }
	@echo 'rev4 require-toolchain: ACCEPTED'
MAKEEOF

# --------------------------------- §6.0a hosted toolchain-identity step (F1)
# VERBATIM copy of the `shell: bash` step body proposed for every Go-consuming
# hosted job in §6.0a. Only the leading `#!/bin/sh` line is added so the harness
# can execute it as a file.
cat > "$WORK/ci-toolchain-identity.sh" <<'IDEOF'
#!/bin/sh
set -u
REQUIRED=go1.25.5

# 1. the pinned constant still matches go.mod
want="go$(awk '/^go[ \t]/{print $2; exit}' go.mod)" \
  || { echo 'toolchain: cannot read go.mod'; exit 1; }
[ "$want" = "$REQUIRED" ] \
  || { echo "toolchain: go.mod requires $want but this workflow pins $REQUIRED"; exit 1; }

# 2. the launcher and the formatter, as PATH actually resolves them after setup-go
goexe="$(command -v go)"     || { echo 'toolchain: no go on PATH after setup-go'; exit 1; }
fmtexe="$(command -v gofmt)" || { echo 'toolchain: no gofmt on PATH after setup-go'; exit 1; }

# 3. exact version -- not "the 1.25 family"
v="$(go version)" || { echo 'toolchain: go version failed'; exit 1; }
printf '%s\n' "toolchain: $v"
case "$v" in *" $REQUIRED "*) ;; *) echo "toolchain: expected $REQUIRED, got: $v"; exit 1;; esac

# 4. absolute reported root. git-bash on windows-latest reports
#    C:\hostedtoolcache\... here and /c/hostedtoolcache/.../go.exe from
#    `command -v`; MSYS accepts C:/..., so a separator swap makes both nameable.
r="$(go env GOROOT)" || { echo 'toolchain: go env GOROOT failed'; exit 1; }
printf '%s\n' "toolchain: GOROOT=$r"
rp="$(printf '%s' "$r" | tr '\\' '/')"
[ -n "$rp" ] || { echo 'toolchain: go env GOROOT is empty'; exit 1; }
case "$rp" in [A-Za-z]:/*|/*) ;; *) echo "toolchain: reported GOROOT is not absolute: $r"; exit 1;; esac

# 5. the launcher IS this root's go and the formatter IS this root's gofmt.
#    Textual first, then -ef (device+inode), which is what makes the assertion
#    survive /c/... vs C:/... and the .exe suffix without weakening it.
same() { [ "$1" = "$2" ] || [ "$1" = "$2.exe" ] || [ "$1" -ef "$2" ] || [ "$1" -ef "$2.exe" ]; }
same "$goexe"  "$rp/bin/go" \
  || { echo "toolchain: launcher $goexe is not $r/bin/go"; exit 1; }
same "$fmtexe" "$rp/bin/gofmt" \
  || { echo "toolchain: formatter $fmtexe is not $r/bin/gofmt"; exit 1; }

# 6. no implicit toolchain download -- READ BACK, never assume the env: block
tc="$(go env GOTOOLCHAIN)" || { echo 'toolchain: go env GOTOOLCHAIN failed'; exit 1; }
printf '%s\n' "toolchain: GOTOOLCHAIN=$tc"
[ "$tc" = local ] || { echo "toolchain: GOTOOLCHAIN=$tc, not local"; exit 1; }

# 7. no per-user go env file in the loop
ge="$(go env GOENV)" || { echo 'toolchain: go env GOENV failed'; exit 1; }
printf '%s\n' "toolchain: GOENV=$ge"
[ "$ge" = off ] || { echo "toolchain: GOENV=$ge, not off"; exit 1; }

echo 'ci-toolchain-identity: ok'
IDEOF
chmod +x "$WORK/ci-toolchain-identity.sh"

# ---------------------------------------- F3: source-staging contract fixtures
# A producer that writes a VALID but INCOMPLETE archive and then fails, which is
# exactly the case `set -e` cannot see through a pipe.
mkdir -p "$WORK/origin"
printf 'a\n' > "$WORK/origin/a.txt"
printf 'b\n' > "$WORK/origin/b.txt"
printf 'c\n' > "$WORK/origin/c.txt"
cat > "$WORK/badproducer.sh" <<'PROD'
#!/bin/sh
# valid tar stream carrying 1 of the 3 intended files, then a non-zero exit
tar -cf - -C "$ORIGIN" a.txt
exit 1
PROD
chmod +x "$WORK/badproducer.sh"

# ----------------- F2 (cycle 5): origin-enumeration failure-injection fixtures
# `find` that emits a VALID partial NUL stream and then fails — the enumeration
# analogue of badproducer.sh. `command -p` is what keeps the stub from recursing.
mkdir -p "$WORK/badfind" "$WORK/badsort" "$WORK/baddigest"
cat > "$WORK/badfind/find" <<'BF'
#!/bin/sh
command -p find . -type f -name 'a.txt' -print0
command -p find . -type f -name 'b.txt' -print0
exit 1
BF
printf '#!/bin/sh\nexit 1\n' > "$WORK/badsort/sort"
printf '#!/bin/sh\nexit 1\n' > "$WORK/baddigest/shasum"
printf '#!/bin/sh\nexit 1\n' > "$WORK/baddigest/sha256sum"
chmod +x "$WORK/badfind/find" "$WORK/badsort/sort" \
         "$WORK/baddigest/shasum" "$WORK/baddigest/sha256sum"

# rev5 form: find | sort | xargs — three producers, one observed status
cat > "$WORK/enumerate-rev5.sh" <<'E'
#!/bin/sh
set -u
( cd "$ORIGIN" && find . -type f -print0 | LC_ALL=C sort -z | xargs -0 $SHA256 ) \
  > "$WORK/rev5-origin.sha256" \
  || { echo 'stage: origin enumeration failed'; exit 1; }
echo "rev5 enumeration: completed, $(grep -c . < "$WORK/rev5-origin.sha256") of 3 files"
E
chmod +x "$WORK/enumerate-rev5.sh"

# corrected form: three MATERIALIZED stages, each status-checked on its own.
# No `pipefail` anywhere — /bin/sh does not provide it portably.
cat > "$WORK/enumerate-fix.sh" <<'E'
#!/bin/sh
set -u
OUT="$WORK/src-origin"
rm -f "$OUT.paths0" "$OUT.sorted0" "$OUT.sha256"

# stage 1: enumerate -> materialize -> check THIS producer's status
( cd "$ORIGIN" && find . -type f -print0 ) > "$OUT.paths0"
rc=$?; [ "$rc" = 0 ] || { echo "stage: origin path enumeration failed rc=$rc"; exit 1; }

# stage 2: sort the materialized stream -> materialize -> check status
LC_ALL=C sort -z < "$OUT.paths0" > "$OUT.sorted0"
rc=$?; [ "$rc" = 0 ] || { echo "stage: origin path sort failed rc=$rc"; exit 1; }

# stage 3: digest the materialized sorted stream -> materialize -> check status
( cd "$ORIGIN" && xargs -0 $SHA256 < "$OUT.sorted0" ) > "$OUT.sha256"
rc=$?; [ "$rc" = 0 ] || { echo "stage: origin digest generation failed rc=$rc"; exit 1; }

want="$(grep -c . < "$OUT.sha256")"
[ "$want" -gt 0 ] || { echo 'stage: origin enumeration produced no files'; exit 1; }
echo "stage: origin enumerated $want files"
E
chmod +x "$WORK/enumerate-fix.sh"

# rev4 form: unchecked tar|tar pipeline under `set -e`
cat > "$WORK/stage-rev4.sh" <<'S'
#!/bin/sh
set -e
rm -rf "$DEST"; mkdir -p "$DEST"
"$WORK/badproducer.sh" | tar -xf - -C "$DEST"
echo "rev4 staging: completed, $(find "$DEST" -type f | wc -l | tr -d ' ') of 3 files present"
S
chmod +x "$WORK/stage-rev4.sh"

# corrected form: three-stage origin enumeration, intermediate archive,
# separately checked statuses, then a complete-set assertion built from an
# inventory enumerated AT THE ORIGIN.
cat > "$WORK/stage-fix.sh" <<'S'
#!/bin/sh
set -u
PRODUCER="${PRODUCER:-good}"
rm -rf "$DEST"; mkdir -p "$DEST"
rm -f "$WORK/src-stage.tar"

# 1. enumerate the intended set at the origin, before any archiving.
"$WORK/enumerate-fix.sh" || exit 1
cp "$WORK/src-origin.sha256" "$WORK/src-origin-inv.sha256"
want="$(grep -c . < "$WORK/src-origin-inv.sha256")"

# 2. create the archive as its own process; check ITS status, not a pipeline's
if [ "$PRODUCER" = "bad" ]; then
  "$WORK/badproducer.sh" > "$WORK/src-stage.tar"
else
  tar -cf "$WORK/src-stage.tar" -C "$ORIGIN" .
fi
rc=$?
[ "$rc" = 0 ] || { echo "stage: source archive creation failed rc=$rc"; exit 1; }

# 3. list and extract, each status-checked
tar -tf "$WORK/src-stage.tar" > "$WORK/src-stage.list" || { echo 'stage: archive listing failed'; exit 1; }
tar -xf "$WORK/src-stage.tar" -C "$DEST" || { echo 'stage: extraction failed'; exit 1; }

[ "${INJECT_MISSING:-0}" = 1 ] && rm -f "$DEST/b.txt"
[ "${INJECT_EXTRA:-0}" = 1 ] && printf 'x\n' > "$DEST/x.txt"

# 4. complete-set assertion: -c catches changed/missing, the count catches extra
( cd "$DEST" && $SHA256 -c --status "$WORK/src-origin-inv.sha256" ) \
  || { echo 'stage: per-file verification failed (changed or missing)'; exit 1; }
got="$(cd "$DEST" && find . -type f | wc -l | tr -d ' ')"
[ "$got" = "$want" ] || { echo "stage: destination has $got files, origin enumerated $want"; exit 1; }
echo "stage: ok, $got of $want files verified"
S
chmod +x "$WORK/stage-fix.sh"

# A failing enumeration must stop BEFORE the archive exists. This wrapper exits 0
# only if staging failed AND no archive was produced.
cat > "$WORK/assert-no-archive.sh" <<'S'
#!/bin/sh
set -u
rm -f "$WORK/src-stage.tar"
env PATH="$WORK/badfind:$PATH" "$WORK/stage-fix.sh"
rc=$?
[ "$rc" != 0 ] || { echo 'assert: staging exited 0 with a failing find'; exit 1; }
[ ! -e "$WORK/src-stage.tar" ] || { echo 'assert: an archive was produced despite a failed enumeration'; exit 1; }
echo "assert: staging failed rc=$rc and produced no archive"
S
chmod +x "$WORK/assert-no-archive.sh"

# --------------------- F3 (cycle 5): the §5.4 W2/W3/W9 empty-root precondition
# `ssh` / `scp` stubs. They simulate the remote REPLY and STATUS only; the
# control-host block below is the verbatim W2/W3/W9 text from §5.4.
mkdir -p "$WORK/sshstub"
cat > "$WORK/sshstub/ssh" <<'SSHEOF'
#!/bin/sh
# models `ssh win "<cmd.exe string>"`. $FAKEWIN/base stands in for $WBASE.
last=""; for a in "$@"; do last="$a"; done
B="$FAKEWIN/base"
case "$last" in
  "echo ok") printf 'ok\r\n'; exit 0;;
  *"where tar"*)
      printf 'C:\\Windows\\System32\\tar.exe\r\n'; exit 0;;
  *"where powershell"*)
      printf 'C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe\r\n'; exit 0;;
  *"if exist"*BASE_EXISTS*)
      if [ -e "$B" ]; then printf 'BASE_EXISTS\r\n'; exit 1; fi
      printf 'BASE_ABSENT\r\n'; exit 0;;
  "mkdir "*)
      if [ -e "$B" ]; then printf 'A subdirectory or file already exists.\r\n'; exit 1; fi
      mkdir -p "$B" || exit 1
      # STALE_AFTER_MKDIR models a leftover/concurrent writer materialising a
      # file in the base between creation and use.
      [ "${STALE_AFTER_MKDIR:-0}" = 1 ] && printf 'stale\n' > "$B/leftover-source.go"
      exit 0;;
  *"Get-ChildItem"*)
      [ -d "$B" ] || { printf 'ERROR\r\n'; exit 1; }
      n=$(command -p find "$B" -mindepth 1 | command -p wc -l | tr -d ' ')
      printf '%s\r\n' "$n"; exit 0;;
  *"rmdir /s /q"*)
      # CLEAN_LEAVES_BASE models an open handle: the command reports success,
      # the directory survives.
      [ "${CLEAN_LEAVES_BASE:-0}" = 1 ] && exit 0
      rm -rf "$B"; exit 0;;
esac
printf 'stub-ssh: unhandled: %s\r\n' "$last" >&2
exit 99
SSHEOF
printf '#!/bin/sh\nprintf "" > "$FAKEWIN/transport-done"\nexit 0\n' > "$WORK/sshstub/scp"
chmod +x "$WORK/sshstub/ssh" "$WORK/sshstub/scp"

# VERBATIM §5.4 W2 + W3 control-host block (WBASE localised for the stub).
cat > "$WORK/w2-w3.sh" <<'W'
#!/bin/sh
set -u
WBASE='C:\Users\admin\AppData\Local\Temp\TASK-260720-1pvfj5'
W2LOG="$WORK/w2.txt"

# --- W2. The base must NOT pre-exist, must be created, and must be EMPTY.
ssh win "if exist \"$WBASE\" (echo BASE_EXISTS & exit /b 1) else (echo BASE_ABSENT)" > "$W2LOG" 2>&1
rc=$?
[ "$rc" = 0 ] \
  || { echo 'W2: base already exists on the remote host; run W9, confirm absence, then re-run W2'; exit 1; }
[ "$(tr -d '\r' < "$W2LOG")" = 'BASE_ABSENT' ] \
  || { echo "W2: unexpected precondition reply: $(tr -d '\r' < "$W2LOG")"; exit 1; }

ssh win "mkdir \"$WBASE\"" > "$W2LOG" 2>&1
rc=$?
[ "$rc" = 0 ] || { echo "W2: mkdir failed rc=$rc: $(tr -d '\r' < "$W2LOG")"; exit 1; }

ssh win "powershell -NoProfile -Command \"(Get-ChildItem -LiteralPath '$WBASE' -Force | Measure-Object).Count\"" > "$W2LOG" 2>&1
rc=$?
[ "$rc" = 0 ] || { echo 'W2: post-create listing failed; the base is not usable'; exit 1; }
got="$(tr -d '\r' < "$W2LOG")"
[ "$got" = '0' ] \
  || { echo "W2: base is not empty after creation ($got entries); a stale file would join the suite input"; exit 1; }
echo 'W2: base created and proved empty'

# --- W3. Only now does anything cross the wire.
scp -O "$STAGE/candidate.tar" "$STAGE/curator-src.tar" \
       "$STAGE/candidate-inventory.sha256" .scripts/verify-candidate.ps1 \
    win:AppData/Local/Temp/TASK-260720-1pvfj5/ \
  || { echo 'W3: transport failed'; exit 1; }
echo 'W3: transport complete'
W
chmod +x "$WORK/w2-w3.sh"

# VERBATIM §5.4 W9 cleanup + absence confirmation.
cat > "$WORK/w9.sh" <<'W'
#!/bin/sh
set -u
WBASE='C:\Users\admin\AppData\Local\Temp\TASK-260720-1pvfj5'
W9LOG="$WORK/w9.txt"

ssh win "rmdir /s /q \"$WBASE\"" > "$W9LOG" 2>&1
rc=$?
[ "$rc" = 0 ] || { echo "W9: cleanup command failed rc=$rc"; exit 1; }

ssh win "if exist \"$WBASE\" (echo BASE_EXISTS & exit /b 1) else (echo BASE_ABSENT)" > "$W9LOG" 2>&1
rc=$?
[ "$rc" = 0 ] || { echo 'W9: base still present after cleanup; do NOT retry the lane'; exit 1; }
[ "$(tr -d '\r' < "$W9LOG")" = 'BASE_ABSENT' ] \
  || { echo 'W9: unexpected absence reply'; exit 1; }
echo 'W9: base removed and absence confirmed'
W
chmod +x "$WORK/w9.sh"

# W2 must stop before transport. This wrapper exits 0 only if W2 failed AND the
# transport marker was never written.
cat > "$WORK/assert-no-transport.sh" <<'S'
#!/bin/sh
set -u
rm -f "$FAKEWIN/transport-done"
env PATH="$WORK/sshstub:$PATH" "$WORK/w2-w3.sh"
rc=$?
[ "$rc" != 0 ] || { echo 'assert: W2 exited 0 against a stale base'; exit 1; }
[ ! -e "$FAKEWIN/transport-done" ] || { echo 'assert: transport ran despite a failed W2'; exit 1; }
echo "assert: W2 failed rc=$rc and nothing crossed the wire"
S
chmod +x "$WORK/assert-no-transport.sh"

# --------------------------------------------------------------------- drive
cd "$WORK" || exit 1
A="$WORK/rootA"
B="$WORK/rootB"
PASS=0; FAIL=0

# run <label> <expected-exit> <command...>
# MODE / STUB_VER / STUB_FORCE_* are set explicitly by each caller, never as a
# prefix assignment: in POSIX sh a prefix assignment before a FUNCTION call
# persists in the caller's shell, so `STUB_VER=go1.25.1 run …` would leak into
# the next case. (Observed in cycle 3: it silently turned one case red.)
run() {
  label="$1"; want="$2"; shift 2
  out="$("$@" 2>&1)"; got=$?
  if [ "$got" = "$want" ]; then
    PASS=$((PASS + 1)); verdict="PASS"
  else
    FAIL=$((FAIL + 1)); verdict="FAIL"
  fi
  printf '%-62s expected=%s actual=%s  %s\n' "$label" "$want" "$got" "$verdict"
  printf '%s\n' "$out" | sed 's/^/      | /'
  echo
}

ORIGIN="$WORK/origin"; DEST="$WORK/dest"
FAKEWIN="$WORK/fakewin"; STAGE="$WORK/stage"
mkdir -p "$STAGE"

reset_stub() {
  MODE=ok; STUB_VER=go1.25.5; STUB_GOROOT="$A"
  STUB_FORCE_TOOLCHAIN=''; STUB_FORCE_GOENV=''
  PRODUCER=good; INJECT_MISSING=0; INJECT_EXTRA=0
  STALE_AFTER_MKDIR=0; CLEAN_LEAVES_BASE=0
  rm -rf "$FAKEWIN"; mkdir -p "$FAKEWIN"
}
export MODE STUB_VER STUB_GOROOT STUB_FORCE_TOOLCHAIN STUB_FORCE_GOENV
export WORK ORIGIN DEST SHA256 PRODUCER INJECT_MISSING INJECT_EXTRA
export FAKEWIN STAGE STALE_AFTER_MKDIR CLEAN_LEAVES_BASE

echo "=== group 1: revision-3 discovery/guard forms — the defect ==="
reset_stub; MODE=partialfail
run "A rev3 \$(shell go list|grep) on a failing go list" 0 \
  make -f Makefile.rev3 bad-discovery
run "B rev3 'go list | grep -q' guard on a failing go list" 0 \
  make -f Makefile.rev3 bad-guard

echo "=== group 2: revision-4 require-toolchain — the cycle-3 F1 defect ==="
reset_stub; STUB_GOROOT="$A"
run "C rev4 accepts GO from rootA paired with GOFMT from rootB" 0 \
  make -f Makefile.rev4 require-toolchain GO="$A/bin/go" GOFMT="$B/bin/gofmt"
reset_stub; STUB_GOROOT="$B"
run "D rev4 accepts a launcher whose GOROOT is not the approved root" 0 \
  make -f Makefile.rev4 require-toolchain GO="$A/bin/go" GOFMT="$A/bin/gofmt"
reset_stub; STUB_FORCE_TOOLCHAIN=auto
run "E rev4 accepts a wrapper-forced GOTOOLCHAIN=auto" 0 \
  make -f Makefile.rev4 require-toolchain GO="$A/bin/go" GOFMT="$A/bin/gofmt"

echo "=== group 3: corrected recipes — healthy must pass ==="
reset_stub
run "F corrected: healthy native host" 0 \
  make test-linux GOROOT_EXPECTED="$A" PIN_ROOT="$WORK/pin"
reset_stub
run "G corrected: healthy hosted runner (setup-go shape)" 0 \
  env PATH="$A/bin:$PATH" make linux-package-guard TOOLCHAIN_ALLOW_PATH=1

echo "=== group 4: corrected recipes — must fail closed ==="
reset_stub; MODE=partialfail
run "H go list fails mid-listing" 2 \
  make test-linux GOROOT_EXPECTED="$A" PIN_ROOT="$WORK/pin"
reset_stub
run "I native host with no GOROOT_EXPECTED" 2 \
  make require-toolchain
reset_stub
run "J relative GO on a native host" 2 \
  make require-toolchain GOROOT_EXPECTED="$A" GO=go GOFMT=gofmt
reset_stub; STUB_VER=go1.25.1
run "K launcher reports go1.25.1, go.mod needs 1.25.5" 2 \
  make require-toolchain GOROOT_EXPECTED="$A"
reset_stub; STUB_GOROOT="$B"
run "L reported GOROOT != approved GOROOT_EXPECTED" 2 \
  make require-toolchain GOROOT_EXPECTED="$A"
reset_stub
run "M GOFMT comes from a different root than the launcher" 2 \
  make require-toolchain GOROOT_EXPECTED="$A" GOFMT="$B/bin/gofmt"
reset_stub; STUB_FORCE_TOOLCHAIN=auto
run "N wrapper-forced GOTOOLCHAIN=auto" 2 \
  make require-toolchain GOROOT_EXPECTED="$A"
reset_stub; STUB_FORCE_GOENV="/Users/u/Library/Application Support/go/env"
run "O wrapper-forced user GOENV file" 2 \
  make require-toolchain GOROOT_EXPECTED="$A"
reset_stub
cp "$WORK/go.mod" "$WORK/go.mod.bak"; cp "$WORK/go.mod.drift" "$WORK/go.mod"
run "P go.mod drifts to 1.25.4 while the Makefile pins go1.25.5" 2 \
  make require-toolchain GOROOT_EXPECTED="$A"
cp "$WORK/go.mod.bak" "$WORK/go.mod"

echo "=== group 5: the hosted Make exception must not bypass identity ==="
reset_stub
run "Q hosted runner, GOFMT from a different root" 2 \
  env PATH="$A/bin:$PATH" make require-toolchain TOOLCHAIN_ALLOW_PATH=1 GOFMT="$B/bin/gofmt"

echo "=== group 6: F3 source staging ==="
reset_stub
run "R rev4 unchecked tar|tar pipeline hides a failed producer" 0 \
  "$WORK/stage-rev4.sh"
reset_stub; PRODUCER=bad
run "S corrected staging rejects the same failed producer" 1 \
  "$WORK/stage-fix.sh"
reset_stub; INJECT_MISSING=1
run "T corrected staging catches a MISSING file after extraction" 1 \
  "$WORK/stage-fix.sh"
reset_stub; INJECT_EXTRA=1
run "U corrected staging catches an EXTRA file after extraction" 1 \
  "$WORK/stage-fix.sh"

echo "=== group 7 (cycle 5, F2): origin enumeration ==="
reset_stub
run "V rev5 find|sort|xargs hides a failing find" 0 \
  env PATH="$WORK/badfind:$PATH" "$WORK/enumerate-rev5.sh"
reset_stub
run "W corrected enumeration, healthy origin" 0 \
  "$WORK/enumerate-fix.sh"
reset_stub
run "X corrected enumeration rejects a partial-then-failing find" 1 \
  env PATH="$WORK/badfind:$PATH" "$WORK/enumerate-fix.sh"
reset_stub
run "Y corrected enumeration rejects a failing sort" 1 \
  env PATH="$WORK/badsort:$PATH" "$WORK/enumerate-fix.sh"
reset_stub
run "Z corrected enumeration rejects a failing digest producer" 1 \
  env PATH="$WORK/baddigest:$PATH" "$WORK/enumerate-fix.sh"
reset_stub
run "AA a failed enumeration produces NO archive" 0 \
  "$WORK/assert-no-archive.sh"

echo "=== group 8 (cycle 5, F1): the hosted toolchain-identity step ==="
reset_stub
run "AB hosted identity step, healthy POSIX runner" 0 \
  env PATH="$A/bin:$PATH" GOTOOLCHAIN=local GOENV=off "$WORK/ci-toolchain-identity.sh"
reset_stub; STUB_VER=go1.25.1
run "AC hosted identity step rejects go1.25.1" 1 \
  env PATH="$A/bin:$PATH" GOTOOLCHAIN=local GOENV=off "$WORK/ci-toolchain-identity.sh"
reset_stub; STUB_FORCE_TOOLCHAIN=auto
run "AD hosted identity step rejects a forced GOTOOLCHAIN=auto" 1 \
  env PATH="$A/bin:$PATH" GOTOOLCHAIN=local GOENV=off "$WORK/ci-toolchain-identity.sh"
reset_stub; STUB_FORCE_GOENV="/Users/u/Library/Application Support/go/env"
run "AE hosted identity step rejects a user GOENV file" 1 \
  env PATH="$A/bin:$PATH" GOTOOLCHAIN=local GOENV=off "$WORK/ci-toolchain-identity.sh"
reset_stub
run "AF hosted identity step rejects a cross-root gofmt" 1 \
  env PATH="$WORK/fmtonly:$A/bin:$PATH" GOTOOLCHAIN=local GOENV=off "$WORK/ci-toolchain-identity.sh"
reset_stub; STUB_GOROOT="$WORK\\rootW"
run "AG identity step, Windows-shape root (backslash, .exe, hard link)" 0 \
  env PATH="$WORK/winpath:$PATH" GOTOOLCHAIN=local GOENV=off "$WORK/ci-toolchain-identity.sh"
reset_stub; STUB_GOROOT="$WORK\\rootW"
run "AH Windows-shape root, cross-root gofmt" 1 \
  env PATH="$WORK/winpath-badfmt:$PATH" GOTOOLCHAIN=local GOENV=off "$WORK/ci-toolchain-identity.sh"
reset_stub
run "AI identity step rejects a shim launcher outside GOROOT" 1 \
  env PATH="$WORK/shim:$PATH" GOTOOLCHAIN=local GOENV=off "$WORK/ci-toolchain-identity.sh"
reset_stub
cp "$WORK/go.mod" "$WORK/go.mod.bak"; cp "$WORK/go.mod.drift" "$WORK/go.mod"
run "AJ identity step rejects go.mod drift to 1.25.4" 1 \
  env PATH="$A/bin:$PATH" GOTOOLCHAIN=local GOENV=off "$WORK/ci-toolchain-identity.sh"
cp "$WORK/go.mod.bak" "$WORK/go.mod"

echo "=== group 9 (cycle 5, F3): W2 empty-root precondition (ssh/scp stubbed) ==="
reset_stub
run "AK W2 on an absent base creates it, proves empty, transports" 0 \
  env PATH="$WORK/sshstub:$PATH" "$WORK/w2-w3.sh"
reset_stub; mkdir -p "$FAKEWIN/base"; printf 'stale\n' > "$FAKEWIN/base/leftover-source.go"
run "AL W2 stops on a stale pre-existing base, before transport" 0 \
  "$WORK/assert-no-transport.sh"
reset_stub; STALE_AFTER_MKDIR=1
run "AM W2 stops when the created base is not empty, before transport" 0 \
  "$WORK/assert-no-transport.sh"
reset_stub; mkdir -p "$FAKEWIN/base"
run "AN W9 removes the base and confirms absence" 0 \
  env PATH="$WORK/sshstub:$PATH" "$WORK/w9.sh"
reset_stub; mkdir -p "$FAKEWIN/base"; CLEAN_LEAVES_BASE=1
run "AO W9 rejects a cleanup that reported success but left the base" 1 \
  env PATH="$WORK/sshstub:$PATH" "$WORK/w9.sh"

echo "-----------------------------------------------------------"
if [ "$FAIL" = 0 ]; then
  echo "ALL $PASS EXPECTATIONS MET"; exit 0
else
  echo "$FAIL of $((PASS + FAIL)) EXPECTATIONS NOT MET"; exit 1
fi
