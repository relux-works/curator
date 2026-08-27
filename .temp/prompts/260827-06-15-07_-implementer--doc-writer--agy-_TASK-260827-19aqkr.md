# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(TASK-260827-19aqkr, status=development)'
```

## Operational Constraints (Headless Run)

Read this before you run any gate or attach any outcome. A spawned run is
usually headless: the session ends when your turn ends.

- **Never background a long command and end your turn.** The session terminates
  with the turn and the process dies with it, while the run is still recorded as
  completed — a successful-looking run with no evidence attached.
- **A single shell call is time-bounded (~10 minutes).** Waiting out a longer
  check inside one call fails too. Split long verification into bounded
  sequential calls (package subsets, `-run` masks) and state explicitly what you
  reran yourself versus accepted from already-attached evidence.
- **Attach evidence before you end the lifecycle, not after.** "I will attach it
  when the run finishes" is unfulfillable by construction: once your turn ends
  there is no session left to attach anything.

## Your Role
# doc-writer

## Description

Writes and updates documentation — README, SKILL.md, references, CHANGELOG. Reads specs, tasks, and code to understand context. Uses domain skills for understanding only.

## Deliverable

Updated documentation files.
Final human-facing wording must say "ready for review" or "handed off to review", not "done", "complete", "finished", "final", or "готово", when the board status is `to-review`.

## Standing Orders

### Evidence Honesty Contract

1. Run each validation or gate command directly as a standalone process. Do not pipe it through `tee`; do not use a pipe chain unless `pipefail` is enabled and the gate command's real status is preserved.
2. Report the real exit code of every validation or gate command.
3. Report expected-red gates truthfully as failing: when a command is expected to fail (for example, `go test` in a package-less module), give its real non-zero exit code and a one-line expected-failure rationale; never present it as passing.
4. Check a checklist item tied to a command only after that exact command has actually run green with exit code 0. If it did not run or did not exit 0, leave the item unchecked.
5. For board reads, use compact task-specific projections. A concrete assignment does not need routine `summary()`, `plan()`, `schema()`, or `{ full }`; request scoped schema only after an unknown call.

## Status Transitions

- **start_status:** `development`
- **end_status:** `to-review` (review handoff, not accepted done)

## Constraints

Does NOT modify code. Read-only access to code, specs, tasks. Writes only documentation files.

## Skills

These role skill references are a lazy catalog, not a mandate to bulk-read every
body. Before technical work, identify the skills relevant to this task's concrete
scope and read those full skill bodies. Always read any skill explicitly required
by the task, user, or project instructions.

- **project-management**: `/Users/iv/.claude/skills/project-management/SKILL.md`

## Definition of Done

- [ ] Every build-https claim verified against internal/ with grep evidence; mirrors build-ssh shape
- [ ] Implementation matches AC
- [ ] Solution fits project architecture
- [ ] Tests green
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches
- [ ] Docs updated and consistent with current code
- [ ] No discrepancies between code and description
- [ ] Result linked as a new task-scoped outcome resource
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
## Your Task

- **ID**: TASK-260827-19aqkr
- **Title**: TASK-260827-19aqkr: build-https-contract
- **Parent**: STORY-260827-3a5efk
### Description

Create docs/build-https.md: the operator HTTPS credential contract for external build repositories, mirroring docs/build-ssh.md in shape and register. Source of truth: the merged implementation in internal/ (find the https credential broker: token sources git-credentials, keyring, token_env; scope grammar and longest-prefix; install-time precheck with detected candidates; environment override variables; fail-closed off-TTY; the broker answers only the two Git prompts for the pinned host). Cross-check against the CocoaSkills sibling doc attached as a precondition, but every claim must be verified against the Curator code, not assumed from the sibling; where the implementations differ, the Curator code wins. Cross-link with docs/build-ssh.md both ways.
### Scope

docs/build-https.md (new), one cross-link line in docs/build-ssh.md.
### Acceptance Criteria

Every credential source, scope rule, env variable, prompt behavior, and error identifier verified against internal/ with grep evidence in the outcome; shape mirrors build-ssh.md; cross-links present; style guide holds.

## Instructions

The following instructions have been attached to this task:

### TASK-260827-19aqkr_cocoaskills-prose-style.md
> Source style guide to port/apply (English rules and blacklist)

# CocoaSkills Documentation Prose Style

Binding style rules for all CocoaSkills documentation in both languages.
Synthesized from The Go Programming Language (Donovan, Kernighan), The Swift
Programming Language, and Russian engineering prose practice. Where this
guide disagrees with a source book, this guide wins.

## Voice and sentences

Write in the active voice and name the actor. The software is the usual
grammatical subject: the installer copies, the audit gate rejects, `csk
install` resolves. Use "you" for the reader's actions and choices. Reserve
the passive for rules where the agent is irrelevant: "Symlinks are not
followed during hashing."

Open with a short, flat claim; let elaboration follow. Load-bearing
sentences stay under roughly ten words. Longer sentences are additive
chains joined by commas and "and", never nested subordination.

Lead with the goal, then the mechanism: "To pin a skill to a tag, add a
`tag` field." The reader learns why before how.

State one idea per sentence. When a second idea appears, start a second
sentence.

## Paragraphs

The first sentence of a paragraph carries its claim; everything after
elaborates. A paragraph introduces at most one new concept. The shape is:
claim, explanation, consequence or example.

Give the definition before the consequences. Name the concept, define it
in one or two sentences, then use it: "A hybrid skill is declared once per
machine and activated per project. The declaration lives in..."

State the problem before the solution. First the drawback, then the fix.

Introduce every code or command block with a sentence that ends in a
colon and says what the block shows. After a non-trivial block, add one
sentence that interprets the result: what the reader should observe.

For example, show an install command:

```bash
csk install
```

The command installs declared skills into `.agents/skills/`.

## Terminology

Define a term at first use, then repeat it verbatim. Never rotate
synonyms: a skill is "installed", not sometimes "deployed" or "provisioned".
Use a pronoun only when the referent is in the same sentence; otherwise
repeat the term. In technical text, repetition is precision, not a defect.

Set every identifier, command, flag, file name, and literal in code style:
`csk install`, `Skillfile.json`, `.agents/skills/`.

When a term must appear before its full treatment, give a working
approximation and a precise pointer: "For now, treat the audit gate as a
pre-install check; the Security model section defines it."

## Punctuation and typography

Do not use em-dashes or en-dashes in prose in either language. Replace
them with a comma, a colon, parentheses, or a new sentence. This is a
deliberate deviation from the source books.

Do not use Russian guillemets. Prefer code style for identifiers; use
plain double quotes only for genuine quotations and coined readings.

Use semicolons to bind two halves of one thought; the second clause
explains or completes the first. Use parentheses for one-sentence asides
(a concrete example, a practical note). No exclamation points, no ellipses
for drama.

## Lists vs prose

Reasoning lives in prose. A list is acceptable only for genuinely parallel
enumerable items: flags, file names, supported platforms, numbered install
steps. Each bullet is a complete, grammatically parallel statement. Never
use a list to carry an argument, and never stack lists of three-item
fragments. Short enumerations (two to four items) stay inline with commas.

Numbered steps are for procedures the reader executes. Each step names the
command and its observable result.

Comparative overviews may use a structured breakdown (table or per-item
lead-ins with short points) when readers need to scan; reasoning inside each
item stays prose.

Collapsible `<details>` blocks are allowed for parallel variants (install
methods, command groups); the summary line names the variant exactly.

## Tone

State facts flatly. No marketing register: no "powerful", "seamless",
"robust", "blazingly fast", no superlatives about the project itself.
State costs and limitations as plainly as features, next to the claim they
limit: "The whitelist strips tests from context; a skill that needs its
tests at runtime must declare them as runtime roots."

Hedge only with calibrated words that carry frequency information:
"usually", "in practice", "rarely". Never hedge to soften a hard rule.

Give prescriptions directly, with the rationale attached: "Always pin a
tag or revision. Branch refs drift and break reproducibility."

Cross-reference with a precise target: "see Security model in
ARCHITECTURE.md", never "as we will see later".

## Russian rules (инженерная проза)

Субъект действия назван. Плохо: "при установке производится копирование
файлов". Хорошо: "установщик копирует файлы".

Глагол вместо отглагольного существительного: "проверяет", не "выполняет
проверку"; "создаёт", не "производит создание".

Определение до следствий, одно новое понятие за раз. Термин повторяется
там, где местоимение создаёт неоднозначность; повтор термина в техническом
тексте повышает точность.

Технический термин остаётся на английском без перевода и без кавычек
только когда у него нет точной русской замены: symlink, ref, commit,
content-hash, fetch. Тест: термин остаётся английским, если он совпадает
с идентификатором в коде, флаге или протоколе (manager-worker-v1,
prompt context) либо если русский эквивалент теряет точность или
создаёт двусмысленность. Во всех остальных случаях пишите по-русски:
предварительная проверка, а не precheck; конвейер, а не pipeline; по
умолчанию, а не дефолт. Русский текст не должен звучать тяжелее
английского: если фраза читается как перевод, перепишите её как
самостоятельное русское предложение.

Никакого канцелярита: "при необходимости", не "в случае наличия
необходимости"; "система обрабатывает", не "осуществляется выполнение
обработки".

## Blacklist

A reviewer rejects a document that contains any of the following:

- Antithesis constructions: "it's not X, it's Y", "не просто X, а Y",
  "this isn't about X". State what the thing is; do not stage a contrast
  with a strawman.
  Bad: "CocoaSkills is not just a package manager, it is a skill runner."
  Good: "CocoaSkills installs skill packages into project repositories."
- Em-dashes and en-dashes used as rhetorical glue, in either language.
  Bad: "csk — a skill manager".
  Good: "csk is a skill manager."
- Russian guillemets.
  Bad: «Skillfile».
  Good: `Skillfile`.
- Chains of triple enumerations and adjective triples ("fast, simple,
  reliable").
  Bad: "The installer is fast, simple, and reliable."
  Good: "The installer is deterministic. The same `Skillfile.json` produces the same tree."
- A closing paragraph that restates the section just written.
  Bad: "In summary, this section introduced the installation commands."
  Good: Omit the summary paragraph.
- Filler openers: "Let's dive in", "стоит отметить", "важно понимать",
  "давайте разберёмся", "in today's world".
  Bad: "Let's dive into configuration."
  Good: "Configure the installer in `Skillfile.json`."
- Bullet lists that carry reasoning instead of parallel facts.
  Bad: Bullet list explaining why content hashing prevents security attacks.
  Good: Paragraph of prose explaining content hashing and its security impact.
- Marketing adjectives applied to the project ("powerful", "seamless",
  "robust", "blazingly fast").
  Bad: "A powerful and seamless skill manager."
  Good: "A tool that manages skill packages."

## Worked contrast

Bad (slop):

> CocoaSkills is not just another package manager — it's a powerful,
> seamless solution for skill management. Let's dive into why it's a
> game-changer: reproducibility, flexibility, and simplicity.

Good:

> CocoaSkills installs skill packages from git repositories into project
> repositories. An install is reproducible: the same `Skillfile.json`
> produces the same content-hashed tree on every machine.



### TASK-260827-19aqkr_docs-refresh-spec.md
> Curator docs refresh spec

# Curator Documentation Refresh

Status: for execution. Owner: orchestrator session, 2026-08-27.
Language: English everywhere. Style: docs/prose-style.md (created by this
refresh; until committed, the task precondition resource carries it).
Model mandate: doc-writer runs on agy gemini-3.6-flash-high.

## Motivation

The README is 408 lines and carries a 170-line reference dump
(compiled-command status, diagnostics, repair, maintenance) inside the
pitch. docs/ holds only the build-ssh contract and a stale
implementation plan citing protocol 1.0.0-rc.2. The merged build-https
broker has no contract document at all. CocoaSkills went through this
exact refresh; this plan ports the shape and the useful documents.

## Target document set

- docs/prose-style.md (new): the English prose rules and the
  generated-text blacklist, ported from the CocoaSkills guide
  (/Users/iv/Developer/Wildberries/cocoaskills/docs/prose-style.md),
  Russian section dropped, examples adapted to Curator.
- README.md (restructured): definition; what Curator manages; install
  with collapsible per-platform options (Homebrew, installer script,
  Scoop, Go toolchain); quick start; a Commands section of collapsible
  groups linking docs/cli.md; the protocol section tightened;
  development kept short with a CONTRIBUTING link. The reference dumps
  move out wholesale.
- docs/compiled-commands.md (new): the current README sections
  "Compiled-command status, diagnostics, and repair" and "Maintenance
  and the build-cache grace period", restructured to the style guide,
  content preserved.
- docs/cli.md (new): full command reference, every synopsis and flag
  verified against the tree binary (go run ./cmd/curator ... --help or
  make build), one section per command group.
- docs/troubleshooting.md (new): symptom, cause, remedy entries drawn
  from the diagnostics prose and the error identifiers in internal/;
  every error string verified against the source.
- docs/build-https.md (new): the operator HTTPS credential contract,
  mirroring docs/build-ssh.md in shape: sources (git-credentials,
  keyring, token_env), scopes, precheck and candidates, env override,
  fail-closed rules. Source of truth: the merged implementation
  (internal/buildrepo, internal/buildhttps or equivalents) and the
  sibling CocoaSkills external-build-repositories.md HTTPS sections.
- docs/build-ssh.md: unchanged except the final style sweep.
- docs/implementation-plan.md: gains a two-line historical header
  (plan of record for v0.1 against rc.2; the board is the live plan);
  content untouched.

## Out of scope

The blocked board tasks compiled-skill-authoring-guide and
external-repository-authoring-and-driver-guide stay with their own
dependency chains. No Russian documents. No WB mirror.

## Execution

Story on the curator board. Producers: agy gemini-3.6-flash-high with
the shell-only tooling note. Reviewers: claude-opus-5 with explicit
reasoning effort, the single-pass verdict note, and the time-budget
note (no full-suite runs, no clones). Delivery: orchestrator git flow,
PR into main after full per-job CI verification; exclude all
.task-board state from the docs commit (known pitfall, upstream issue
skill-project-management#18).



### TASK-260827-19aqkr_tooling-note.md
> Shell-only edits, quoted heredocs, grep verification, literal outputs

# Mandatory tooling note

The agy provider's native write_to_file/artifact tool CANNOT write files
into this repository (it only accepts paths inside its own brain
directory) and a previous run on this board handed off with a complete
checklist while its file edits were silently lost.

Rules for this run:

1. Edit repository files ONLY through shell commands (cat > file,
   python3 heredoc, perl -pi -e), never through your native
   write_to_file/artifact tool.
2. Work directory is /Users/iv/Developer/Wildberries/cocoaskills. Verify
   with pwd before editing.
3. After every file edit, verify with grep/head that the change is
   actually in the file, and include that verification output in your
   outcome resource.



### TASK-260827-19aqkr_reviewer-note.md
> Single-pass review, immediate verdict, no monitors

# Reviewer execution note (mandatory)

Prior reviewer runs on this board have parked tasks in `reviewing`
without a verdict (one exited waiting for a monitor). Do NOT start
monitors, background waits, or polling loops. Complete the review in one
pass: read the diff and the task AC, run your checks synchronously, then
IMMEDIATELY hand off with exactly one verdict branch: accepted (done) or
changes requested (to-dev) with a verdict resource attached. Ending your
run while the task sits in `reviewing` is a failed review.



### TASK-260827-19aqkr_sibling-doc.md
> CocoaSkills sibling contract for cross-checking; Curator code wins on any difference

# External build repositories

CocoaSkills implements the Curator Protocol `1.0.0-rc.10` schema-8
`go-repository-v1` boundary. It builds an executable from a separately locked
Git repository while keeping the skill package unable to select credentials,
Git configuration, hooks, compiler flags, output paths, wrappers, or signing.

The accepted protocol revision is
`b8b03d597ac83d158a0eadd9d0b25d2e883de1a3`; its `conformance/v1/manifest.json`
SHA-256 is
`803918bf8672f76cf990985e51db213b826674cd5bb54fbf47731b8404b44403`.
The external-repository corpus is supplied to tests independently, so the csk
consumer imports no Curator implementation package or internal fixture value.

## Skill declaration

An `agent-skill.json` schema-7 or schema-8 declaration binds a canonical network identity,
an exact Git object ID, and optionally an exact tag:

```json
{
  "schema_version": 7,
  "capabilities": {},
  "build_repositories": {
    "golden-tools": {
      "git": "https://github.com/example/golden-tools.git",
      "locked_commit": {
        "object_format": "sha1",
        "hex": "0123456789abcdef0123456789abcdef01234567"
      },
      "tag": "v1.4.0"
    }
  },
  "commands": {
    "golden-tool": {
      "type": "build",
      "driver": "go-repository-v1",
      "repository": "golden-tools",
      "target": "golden-tool"
    }
  }
}
```

The referenced repository contains a closed `skill-build.json` descriptor:

```json
{
  "schema_version": 1,
  "targets": {
    "golden-tool": {
      "driver": "go-repository-v1",
      "build_root": ".",
      "source_dir": "cmd/golden-tool"
    }
  }
}
```

`build_root` must contain `go.mod`. Only that root is exposed to the compiler.
The output is always the manager-derived `bin/golden-tool` on macOS or
`bin/golden-tool.exe` on Windows. Arbitrary argv, environment, output, hook,
plugin, generator, credential, helper, filter, and signing fields are rejected.

## Admission and audit order

Every install reacquires the exact object or tag through an operator-selected
Git executable. CocoaSkills clears ambient Git configuration and helper state,
uses one exact refspec, proves raw object identities and the reachable graph,
rejects LFS pointers, submodules, links and special modes, materializes and
rehashes the complete snapshot, validates `skill-build.json`, and runs the
independent external audit before any protected artifact lookup or compiler
call.

Non-executable text below `vendor/` in a third-party module (such as a
dependency's README or Makefile) does not block installation: the fixed
build session runs only `go list` and `go build`, hooks and generators are
forbidden, so such text is never executed. CocoaSkills does not report such a
finding: it neither blocks the install nor appears in install output.
Executable files below `vendor/` and any critical findings still block as
before.

## Private HTTPS build repositories

A private HTTPS repository fetch authenticates through the **manager
credential broker**, the slot the manager profile describes as an OPTIONAL
member of the allowed process graph. The manager launches the broker itself:
a private wrapper beside the SSH wrapper, pinned to one host, named by both
`GIT_ASKPASS` and `core.askPass`. A repository can neither select credentials
nor divert them. The fetch goes to a single TLS-verified URL with redirects
disabled, and the broker answers only the two prompts Git asks and only for
the pinned host; any other prompt exits without printing a byte.

The config stores the token source, never a token:

```json
"build_https": {
  "gitlab.example.com/portals/infra": {"token": "git-credentials"},
  "gitlab.example.com/vendor": {"token": "keyring", "username": "oauth2"},
  "ci.example.com": {"token_env": "CI_TOKEN"}
}
```

Scopes use the `build_ssh` grammar: segment prefixes of the canonical
identity, matched on `/` boundaries, longest match wins. Three sources
exist:

| Source | What it reads | Who it suits |
| --- | --- | --- |
| `git-credentials` | the operator's existing HTTPS entry, the one their own credential helper already serves | anyone who has cloned over HTTPS once: no new secret is created |
| `keyring` | the token `csk config build-https login <scope>` stores through that same helper under a namespaced username | operators with neither SSH nor HTTPS history |
| `token_env` | an environment variable read at process entry | CI and headless runs |

The manager reads the credentials, not the broker. The read happens before
the fetch, outside its process graph, through `git credential
fill|approve|reject`. That is the one mechanism which exists identically on
macOS, Windows and Linux, speaks to whichever helper the operator already
configured (`osxkeychain`, `wincred`, `libsecret`, GCM), and needs no runtime
dependency. The operator's Git configuration selects the helper, never a
repository or a manifest. Interactive prompting is disabled
(`GIT_TERMINAL_PROMPT=0`, `GCM_INTERACTIVE=never`), so an absent credential
degrades instead of hanging the install on a dialog.

A token saved through `build-https login` lives under the username
`csk-build-https:<scope>`, separate from the operator's own entry for the
same host, so neither overwrites the other.

Manage the scopes with the subcommands:

```sh
csk config build-https add gitlab.example.com/portals/infra --token git-credentials
csk config build-https login gitlab.example.com/vendor      # hidden PAT input
csk config build-https list
csk config build-https remove gitlab.example.com/vendor     # also drops the keyring entry
```

`CSK_BUILD_HTTPS_TOKEN` (with the optional `CSK_BUILD_HTTPS_USERNAME`)
overrides every scope for one run, exactly as `CSK_BUILD_SSH_*` overrides the
SSH scopes. A token is never accepted as a flag. The unpinned override trusts
the entire closure: HTTPS basic auth transmits the token to whichever host a
manifest names, so every HTTPS build repository host in the closure can
receive it. Set `CSK_BUILD_HTTPS_HOST` to pin the override to one host; a
repository on any other host then resolves as if the override were absent.
Use the unpinned form only when every build repository host in the closure is
trusted.

As on the SSH surface, a precheck before the first fetch lists the detected
candidates on a terminal (the existing Git credentials for the host, or a new
PAT entered on the spot) and saves a choice only after an explicit scope
selection. A missing selection is not an error for HTTPS: anonymous HTTPS
stays a first-class transport, and a public repository fetches exactly as
before.

The fetch environment is deliberately clean (an empty `PATH`, a private
`HOME`), so the manager performs the helper read at its own `PATH` and
`HOME`, with the absolute path of the Git executable it already admitted. The
broker receives only the result and stays a pure answer function, identical
on every platform.

The token never lands in the config, a flag, a log, or a diagnostic: it lives
only in the environment of the fetch children. `token_value` is excluded from
`repr`; spec 11.1 forbids broker values in receipts, markers and diagnostics.

## Private SSH build repositories

An SSH build repository fetch runs in a private empty `HOME` with an empty
`PATH` and no inherited agent socket, so it never adopts the operator's
`~/.ssh/config`, ambient `GIT_SSH_COMMAND`, or a repository-selected wrapper.
Credentials therefore have to be named explicitly. Nothing is inherited
implicitly, and an SSH source with no selection fails closed with
`build_repository_ssh_credential_missing` before any launcher, snapshot, or
cache artifact is written.

| Surface | Flag | Environment variable |
| --- | --- | --- |
| Identity | `--build-ssh-identity PATH` | `CSK_BUILD_SSH_IDENTITY` |
| Agent socket | `--build-ssh-agent [SOCKET]` | `CSK_BUILD_SSH_AGENT` |
| Host keys | `--build-ssh-known-hosts PATH` | `CSK_BUILD_SSH_KNOWN_HOSTS` |

Flags win over the environment. `--build-ssh-agent` with no value, or the value
`auto`, adopts the operator's live `SSH_AUTH_SOCK`. Host keys default to the
operator home's `.ssh/known_hosts` because the fetch pins
`StrictHostKeyChecking=yes`; the file is copied into the private root, so a
fetch cannot rewrite operator state. All three accept symbolic links and are
admitted as their resolved targets.

Three selections are accepted:

```sh
# Unencrypted key on disk.
csk install --build-ssh-identity ~/.ssh/id_ed25519

# Agent holds the key, and the public key pins which agent key is offered.
# Prefer this for passphrase-protected keys.
csk install --build-ssh-agent --build-ssh-identity ~/.ssh/id_ed25519.pub

# Agent only. Every loaded key is offered in turn, so a populated agent can
# exhaust the server's MaxAuthTries budget before reaching the right one.
csk install --build-ssh-agent
```

### Persistent scoped selection

Instead of repeating flags or environment values, the operator may store the
selection in the global config, keyed by a canonical-identity prefix:

```json
"build_ssh": {
  "gitlab.example.com": {"identity": "~/.ssh/personal"},
  "gitlab.example.com/portals/infra": {
    "agent": "auto",
    "identity": "~/.ssh/work.pub"
  }
}
```

A scope is a segment prefix of the schema-7 or schema-8 canonical repository identity
(`host/path`): matching happens only on whole `/` boundaries, and the longest
matching scope wins, so a key granted to one namespace never reaches a
repository outside it. Flags win over `CSK_BUILD_SSH_*`, and both win over
every configured scope.

A scope needs at least one of `agent` or `identity`; each alone is a complete
selection:

- `{"agent": "auto"}`: agent-only. The install adopts the operator's live
  `SSH_AUTH_SOCK` at run time and the agent signs with its loaded keys in
  turn. No key file is named, so a populated agent can exhaust the server's
  `MaxAuthTries` budget before reaching the right key.
- `{"identity": "~/.ssh/key"}`: identity-file only, for an unencrypted key
  on disk (`IdentityAgent=none`).
- both: pinned-agent form. The third canonical authentication-tail form,
  RECOMMENDED, per curator-spec#22. The agent holds the private key and the
  named `.pub` pins which single key is offered.

Manage the map with:

```sh
csk config build-ssh add gitlab.example.com/portals/infra \
    --agent auto --identity ~/.ssh/work.pub
csk config build-ssh list
csk config build-ssh remove gitlab.example.com/portals/infra
```

Before any fetch, the install resolves credentials for every declared SSH
build repository. On an operator terminal an unmatched repository prompts with
a menu of **detected candidates** (the live agent socket with its loaded key
count and the `.pub` files below `~/.ssh`), so the usual answer is a single
Enter on the default "agent + pinned key" entry. Discovery only lists what
exists; nothing is ever used without the operator's explicit selection, and
nothing persists without the explicit scope choice. A non-interactive run
fails closed with `build_repository_ssh_credential_missing` and ready-to-run
`csk config build-ssh add` commands built from the same detected candidates. `csk install --dry-run`
reports which source (flags, environment, or a config scope) covered each
repository.

CocoaSkills writes a private wrapper carrying one pinned `ssh` argv and points
`GIT_SSH_COMMAND` at it. The wrapper refuses to run unless Git hands it exactly
the host and `git-upload-pack` invocation that argv was pinned to, so no
repository value can add an option, change the host, or reach another path. The
operator's own `ssh` on `PATH` is used unchanged and never has to be shadowed.

For a declared tag, the fetched tag must still terminate at `locked_commit`.
A moved tag, missing tag/object, inaccessible source, malformed raw object, or
failed audit stops without publishing a shim or marker. An untagged source may
reuse the exact protected snapshot recorded by an existing marker when the
network is unavailable; a tagged source always requires a fresh tag proof.

## Build, cache, and lifecycle

The fixed Go contract is the same `manager-worker-v1` session documented in the
main README: native toolchain, vendored modules, no network, no workspace, no
cgo, internal linking, and manager-derived output. External builds use a
receipt-v2 cache below `<csk-home>/external-builds`; schema-7 installations use
marker v3 and schema-8 installations use marker v4; both may contain local
receipt-v1 and external receipt-v2 commands together.

Project install publishes `.agents/bin/<command>`; global install publishes
`<csk-home>/global/bin/<command>`. Both managed launchers point directly at the
validated protected artifact, preserve arguments and exit status, and retain
the inherited PATH. Do not copy a compiled artifact into `scripts`, add a
hand-written wrapper, or prepend a private cache directory to PATH. Agents
already resolve project then global managed shims. For optional interactive
bare commands, use the documented `csk shell-init --install` hook.

Run ordinary lifecycle commands:

```text
csk install
csk install --dry-run
csk status
csk install                 # repair/reinstall
csk global install
csk global status
csk gc
```

A global Skillfile that mixes reachable and unreachable repositories does not
have to be installed as a whole: `csk global install --only <name>` (repeatable)
restricts the run to one declaration and its required closure, so an unselected
private repository is never cloned or fetched, and installed skills outside the
selection keep their markers, shims, and adapter entries. Combine it with the
operator SSH options above to install exactly the private build repository the
operator holds credentials for.

To uninstall, remove the skill declaration from the project or global
`Skillfile.json` and run the matching install command; reconciliation removes
the stale marker and shim transactionally.

Dry-run still acquires, proves, validates, audits, and inspects the candidate
cache, but does not compile or mutate. A corrupt receipt, artifact, or snapshot
is never patched or adopted: a mutating install quarantines it and rebuilds
from a newly proved source. Project/global marker and shim publication uses the
existing transaction engine, so a build, collision, crash, or consumer-marker
failure leaves the prior complete installation current or recoverable.

## Toolchain fingerprint deadline

Every build session hashes the complete selected GOROOT to pin the toolchain
identity, and each hashing pass is bounded by a deadline. Reading a cold Go
installation is much slower than reading a warm one, especially on Windows,
where on-access antivirus scans each file the first time it is touched. When
the pass does not finish in time the install fails closed with:

```text
csk install
app: go-v1 toolchain_timeout: toolchain fingerprint deadline exceeded
hashing the Go toolchain did not finish in time; set
CSK_GO_FINGERPRINT_TIMEOUT to a larger number of seconds (default 600, maximum
3600) on hosts where a cold GOROOT reads slowly, for example behind on-access
antivirus
```

The first line is the cross-implementation protocol string and never changes;
the remedy follows it. Operators raise the bound with
`CSK_GO_FINGERPRINT_TIMEOUT`, a number of seconds:

```bash
CSK_GO_FINGERPRINT_TIMEOUT=1800 csk install
```

The default is 600 seconds and the accepted range is 0 exclusive to 3600
seconds inclusive; a larger value is clamped to 3600 and a missing, empty, or
unparseable value falls back to the default rather than failing the install.
The deadline is a liveness bound, not a trust decision: raising it never admits
a toolchain that would otherwise be refused, and it can never be removed
entirely. Callers embedding CocoaSkills set the same bound in code through
`ToolchainConfig(fingerprint_timeout=...)` or the `timeout` argument of
`fingerprint_toolchain`, which take precedence over the environment.

Prefer raising this bound over retrying a failed install. A retry only appears
to help because the first attempt warmed the operating system cache.

## Development substitutions

`Skillfile.dev.json` schema 2 may replace one declared repository for local
development without changing the package declaration:

```json
{
  "schema_version": 2,
  "substitutions": {},
  "build_repository_substitutions": {
    "golden-skill": {
      "golden-tools": {"path": "../golden-tools"}
    }
  }
}
```

A network substitution instead declares `git` plus one typed `revision` or
`tag`. Local selection admits a narrow ordinary `.git` layout and records a
host-path-free operator-local identity. Substitution state is explicit in
receipt v2 and marker v3 and never aliases the declared source. Strict audit
refuses substituted installs. Keep `Skillfile.dev.json` ignored; csk verifies
that boundary before use.

## Platform qualification

`go-repository-v1` is supported and qualified only on native macOS and Windows
hosts. Linux support is deliberately deferred and is not implied by generic
CocoaSkills script/system-command support. Platform evidence must record the
exact OS, architecture, Python, Git, Go, csk, Curator consumer, protocol commit,
and corpus manifest used by the run.



### TASK-260827-19aqkr_finalize-note.md
> Work done; verify, evidence, CR, handoff; no rewrites

# Finalization run (previous run timed out after writing the document)

docs/build-https.md (285 lines) already exists in the story worktree
with an outcome resource attached. Do NOT rewrite it. Within 15 minutes:
1. Verify the document intact and the build-ssh cross-link present
   (grep both files).
2. Spot-check three claims against internal/ with grep (one credential
   source, one env variable, one error identifier) and append the
   literal outputs to the outcome resource.
3. Register the change request revision if the CR plane requires it,
   complete the checklist, and hand off with task-board handoff.





## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# FIRST/LAST lifecycle commands:
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(TASK-260827-19aqkr, status=analysis)'       # analyst-style work
task-board m 'set_status(TASK-260827-19aqkr, status=development)'    # implementation / testing work
task-board m 'set_status(TASK-260827-19aqkr, status=reviewing)'      # reviewer handoff
task-board m 'set_status(TASK-260827-19aqkr, status=blocked)'        # when blocked
task-board m 'set_status(TASK-260827-19aqkr, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(TASK-260827-19aqkr, item=1)'                        # check item N
task-board m 'add_checklist_item(TASK-260827-19aqkr, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(TASK-260827-19aqkr, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(TASK-260827-19aqkr, name=TASK-260827-19aqkr_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add TASK-260827-19aqkr ./path/to/file --type outcome --name TASK-260827-19aqkr_artifact.bin -d "Description"
```

## Spawn Run Control

Tracked background spawn runs expose `TASK_BOARD_RUN_ID` in the child environment.
If your work is long-running, check for operator directives at safe checkpoints:

```bash
task-board spawn status "$TASK_BOARD_RUN_ID"
task-board spawn directives "$TASK_BOARD_RUN_ID"
```

Current runtimes do not support direct inbound push into your active session.
Treat directives as cooperative checkpoint signals:
- persist your current notes/artifacts before acting on `cancel`-style requests
- only honor pause/reroute intent at a safe checkpoint
- if no directive is present, continue normally

## IMPORTANT: Saving Results

When you produce work products (research documents, design docs, screenshots, logs, archives, implementation notes), you MUST save them as outcome resources with names that include the task ID:

```bash
task-board m 'add_resource(TASK-260827-19aqkr, name=TASK-260827-19aqkr_results.md, content="...", type=outcome, description="Description")'
task-board resource add TASK-260827-19aqkr ./path/to/file --type outcome --name TASK-260827-19aqkr_artifact.bin -d "Description"
```

If you revise the same artifact later, use `task-board m 'update_resource(...)'` or `task-board resource update ...` instead of creating a silent overwrite.

If you discover important findings, decisions, anomalies, regressions, or non-obvious constraints while working, record them in `logbook` as well as on the board.

This ensures your results persist on the board and are accessible to other agents and the coordinator. Spawn completion is expected to produce at least one new task-scoped outcome artifact before the task can cleanly remain in `to-review`.

## Evidence That Counts

A passing suite means nothing unless something in it would have failed.

- Any behavior that GATES, REFUSES, VALIDATES, AUTHORIZES, or ATTESTS ships with negative tests that fail when the gate admits what it must reject. A positive test proves the gate is reachable, not that it works.
- Prove a bound by NARROWING the gate, not only by deleting it. A delete-only mutant proves the gate exists and says nothing about the class it covers.
- Prove behavior by driving the real entry point — launch, materialize, resolve, publish — and name the production call site. A helper that is unit-tested but called from nowhere promises nothing.
- Standard negative shapes to test and to look for: forged or self-minted evidence; absent evidence treated as satisfied; the check present but uncalled from production; a bypass path around the check; a capability claim that does not reproduce.
- An absence and a failure to read are different facts. A failed, partial, or malformed read is never a legitimate absence, and a fallback defined for absence must not fire on a read failure.
- Prove, or report nothing. Where a property cannot be established, report unknown instead of inferring it from a proxy signal; callers act on a plausible guess.

Shapes vocabulary and the incident record behind each rule: `references/negative-evidence.md` in the project-management skill.

## Stop-The-Line: No Forced Fits

Do not keep implementing when autonomous work starts requiring a forced fit. A forced fit is any path where the task conflicts with a platform/API constraint, product decision, UX state model, ownership boundary, or architecture, and the remaining "solution" is mostly compensating hacks.

Warning signs:
- each fix needs another flag, stub, priority rule, mock-only behavior, or special-case test
- the tests can pass only because the test harness avoids the real platform behavior
- the implementation depends on an assumption you can no longer defend
- the user-facing behavior cannot be described cleanly without contradicting the product model

When this happens, stop product-code changes before adding another workaround layer. Attach or note:
- the constraint and evidence
- the failed assumptions/attempts
- the viable options and tradeoffs
- the recommended option
- the exact human/product/architecture decision needed

Then set the board item to `blocked` and ask only for that exact decision or external input. This stop applies only to a concrete external blocker or an unresolved human-only platform/product/architecture/tradeoff/approval decision; recoverable failures and ordinary rework stay autonomous. Tests and stubs are not proof that a forced-fit design is correct; use them only after the state model and platform assumptions are valid.

## Completion Discipline

Keep working until the task reaches a terminal handoff for your role. If no objective blocker remains, do not stop while the board item is still parked in `analysis`, `development`, `testing`, or `reviewing`.

Before your final status change:
- satisfy the task acceptance criteria and relevant checklist items
- attach outcome evidence for the work you produced
- run the relevant verification commands when the task changes code, tests, docs, or config

Use `blocked` only for either a concrete external blocker you cannot resolve autonomously or an unresolved human-only platform/product/architecture/tradeoff/approval decision. Record the constraint, evidence, failed assumptions/attempts, viable alternatives and tradeoffs, recommendation, and exact human decision or external input needed. Recoverable failures and ordinary rework are not `blocked`.

Status language is literal:
- `to-review` means your role has handed work to review; it does not mean the board task is accepted or done.
- In your final response, say "ready for review" or "handed off to review" when the final board status is `to-review`.
- Do not say "done", "complete", "finished", "final", or "готово" as the overall task state unless the board status is actually `done`.

## LAST — Run For Role Handoff

When you have completed all role work and the task is ready for its role handoff, run this as your **final board command**:

```bash
task-board handoff TASK-260827-19aqkr --role doc-writer
```

## Working Directory
Board directory: `/Users/iv/Developer/ReluxWorks/curator/.task-board`

Work in the project root. Do not modify board files directly — always use the `task-board` CLI.

## Story Workspace

You are running in an isolated Git worktree for STORY-260827-3a5efk.

- Workspace path: `.temp/STORY-260827-3a5efk/worktree` (this is your working directory)
- Branch: `task-board/story/STORY-260827-3a5efk`, forked from `main`
- Authoritative board: `/Users/iv/Developer/ReluxWorks/curator/.task-board` (already exported as `TASK_BOARD_DIR`)

Make every repository change here. Do not switch, rebase, merge, or delete this branch, and do not run `task-board` against the `.task-board` copy inside this worktree — it is a checkout artifact, not the board. Integration into trunk is the orchestrator's step, not yours.
