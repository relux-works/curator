#!/bin/sh
# TASK-260729-osjeay — self-contained re-run of the §7.4 recipe validation.
# Revision 5 (rework cycle 4): extends the cycle-3 harness from 7 to 21 cases.
#
# Validates the `require-toolchain`, `test-linux`, `linux-package-guard`
# recipes and the §5.2 C3 source-staging contract proposed for
# TASK-260720-1pvfj5 in the final-CI execution map.
#
# NO GO IS EXECUTED. `go` is a /bin/sh stub that prints canned `version`,
# `env GOROOT`, `env GOTOOLCHAIN`, `env GOENV`, `list ./...` and `list -f`
# output, and can be told to fail mid-listing or to override the caller's
# environment the way a shim/wrapper launcher does.
# Exit codes below describe THIS HARNESS, not the Curator suite.
#
# Usage:  sh .temp/TASK-260729-osjeay/verify-recipes.sh
# Expect: the summary table ends with "ALL 21 EXPECTATIONS MET" and exit 0.

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
      GOROOT)      echo "$STUB_GOROOT"; exit 0;;
      GOTOOLCHAIN) echo "${STUB_FORCE_TOOLCHAIN:-${GOTOOLCHAIN:-auto}}"; exit 0;;
      GOENV)       echo "${STUB_FORCE_GOENV:-${GOENV:-/Users/u/Library/Application Support/go/env}}"; exit 0;;
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

# rev4 form: unchecked tar|tar pipeline under `set -e`
cat > "$WORK/stage-rev4.sh" <<'S'
#!/bin/sh
set -e
rm -rf "$DEST"; mkdir -p "$DEST"
"$WORK/badproducer.sh" | tar -xf - -C "$DEST"
echo "rev4 staging: completed, $(find "$DEST" -type f | wc -l | tr -d ' ') of 3 files present"
S
chmod +x "$WORK/stage-rev4.sh"

# corrected form: intermediate archive, separately checked statuses, then a
# complete-set assertion built from an inventory enumerated AT THE ORIGIN.
cat > "$WORK/stage-fix.sh" <<'S'
#!/bin/sh
set -u
PRODUCER="${PRODUCER:-good}"
rm -rf "$DEST"; mkdir -p "$DEST"

# 1. enumerate the intended set at the origin, before any archiving
( cd "$ORIGIN" && find . -type f -print0 | LC_ALL=C sort -z | xargs -0 $SHA256 ) > "$WORK/src-origin.sha256" \
  || { echo 'stage: origin enumeration failed'; exit 1; }
want="$(grep -c . < "$WORK/src-origin.sha256")"

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
( cd "$DEST" && $SHA256 -c --status "$WORK/src-origin.sha256" ) \
  || { echo 'stage: per-file verification failed (changed or missing)'; exit 1; }
got="$(cd "$DEST" && find . -type f | wc -l | tr -d ' ')"
[ "$got" = "$want" ] || { echo "stage: destination has $got files, origin enumerated $want"; exit 1; }
echo "stage: ok, $got of $want files verified"
S
chmod +x "$WORK/stage-fix.sh"

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

reset_stub() {
  MODE=ok; STUB_VER=go1.25.5; STUB_GOROOT="$A"
  STUB_FORCE_TOOLCHAIN=''; STUB_FORCE_GOENV=''
  PRODUCER=good; INJECT_MISSING=0; INJECT_EXTRA=0
}
export MODE STUB_VER STUB_GOROOT STUB_FORCE_TOOLCHAIN STUB_FORCE_GOENV
export WORK ORIGIN DEST SHA256 PRODUCER INJECT_MISSING INJECT_EXTRA
ORIGIN="$WORK/origin"; DEST="$WORK/dest"

echo "=== group 1: revision-3 discovery/guard forms — the defect ==="
reset_stub; MODE=partialfail
run "A rev3 \$(shell go list|grep) on a failing go list" 0 \
  make -f Makefile.rev3 bad-discovery
run "B rev3 'go list | grep -q' guard on a failing go list" 0 \
  make -f Makefile.rev3 bad-guard

echo "=== group 2: revision-4 require-toolchain — the F1 defect ==="
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

echo "=== group 5: the hosted exception must not bypass identity ==="
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

echo "-----------------------------------------------------------"
if [ "$FAIL" = 0 ]; then
  echo "ALL $PASS EXPECTATIONS MET"; exit 0
else
  echo "$FAIL of $((PASS + FAIL)) EXPECTATIONS NOT MET"; exit 1
fi
