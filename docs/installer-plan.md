# Installer plan and confirmation

`install.sh` shows the installation plan before creating staging files,
downloading release artifacts or creating the destination directory. Today it
installs `curator` only. Each component block lists the version, final executable
path, whether sudo is needed, its source repository, README, book and protocol
specification. Curator does not request root or run sudo.

When `CURATOR_VERSION` is unset, the installer makes a read-only GitHub
latest-release metadata GET before displaying the plan and asking for consent,
so the plan can show the real version. This request writes no files, ignores
curl configuration files, and has a 3-second connection timeout and 5-second
total timeout. The plan explicitly states this pre-consent lookup. Setting
`CURATOR_VERSION` skips it.

If metadata cannot be fetched or has no version, the complete plan still prints
with `latest (could not be resolved)` (localized in Russian). `--plan` exits 0.
Installation asks for consent as usual; a decline changes nothing. After
consent, an unresolved version is looked up again with the same timeouts. A
second failure clearly refuses installation before creating staging files,
downloading release artifacts or creating the destination directory.

The plan describes the existing verification chain: GitHub build attestation
when supported, otherwise a cosign signature, followed by the archive's SHA-256
checksum. A failed attestation still refuses rather than falling back to cosign.
The emergency `CURATOR_INSTALL_INSECURE_SKIP_VERIFY=1` escape keeps its existing
semantics and warning; the plan also says that verification is disabled.

Language selection uses the first nonempty value among `LC_ALL`, `LC_MESSAGES`
and `LANG`. Russian (`ru`, `ru_*`, `ru.*`, `ru@*`) selects Russian; English,
unknown and unset locales select English. All plan and confirmation translations
live together in the language blocks in `install.sh`. Adding a language requires
one block. Labels are padded there so UTF-8 columns stay aligned in POSIX shells.
Only a terminal on stdout gets colour; setting `NO_COLOR`, even to an empty
value, disables it.

```sh
# Preview only. Reading latest-version metadata is allowed; no release artifacts,
# staging directories or installation files are written.
curl -fsSL https://raw.githubusercontent.com/relux-works/curator/main/install.sh | sh -s -- --plan

# Automation: consent without a prompt, with normal verification still required.
curl -fsSL https://raw.githubusercontent.com/relux-works/curator/main/install.sh | sh -s -- --yes
# Alternatively: ... | env CURATOR_YES=1 sh
```

`--plan` takes precedence over `--yes` and `CURATOR_YES=1`. Without automatic
consent, the installer reads confirmation from `/dev/tty`, so `curl | sh` cannot
consume script text as an answer. Without a terminal it refuses and explains
the automation options. Blank, negative or unrecognized replies cancel with
nonzero status before any installation changes. English yes/y and the localized
affirmative reply are accepted.

The existing destination rule is unchanged: use `CURATOR_BIN_DIR` or
`/usr/local/bin` when writable; otherwise use `$HOME/.local/bin`. Resolution is
read-only before the plan; directory creation happens after confirmation and
verification. The destination used is the one shown in the plan.

## Validation

```sh
env -u GIT_DIR -u GIT_WORK_TREE -u SSH_AUTH_SOCK -u GOFLAGS go test ./internal/install -run '^TestInstallScriptSecurityRows$' -count=1 -timeout=3m
sh -n install.sh
```

The existing release-verification fixtures now explicitly consent with `--yes`.
Fixtures leave `CURATOR_VERSION` unset by default and serve metadata from a
network stub; a pinned-version control proves that it skips the lookup. Failure,
empty metadata, partial failed responses and simulated curl timeout controls
cover full-plan rendering and failure before staging/download; decline and
retry-recovery controls exercise the consent boundary. The timeout stub checks
the connection/total timeout arguments and returns curl's timeout exit code;
it does not measure live network latency.
The plan subtests use scratch homes and prefixes, fixture release assets and
verifiers, and the standard `script` utility for real PTY confirmation/colour
checks on macOS/Linux. They compare scratch-tree contents/modes and regular-file
timestamps, trace filesystem tools, and reject a deliberately planted early
`mktemp`. These tests do not perform real downloads or cryptographic verification
with release keys; they preserve the existing verification policy/argv controls.
Keep test output under task-scoped `.temp/`.
