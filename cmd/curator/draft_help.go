package main

// Skillfile schema-2 source workflow help for the existing commands.
//
// Nothing here adds a command or flag. `project resolve` and `project
// refresh` answer `-h` and `--help` with the workflow help. `install` and
// `status` append the same workflow section. The bare word `help` is never
// intercepted: it resolves as a project alias or path. Every example
// below is exercised through run() by draft_diagnostics_test.go; the full
// reference with machine policy setup lives in docs/cli.md.

import (
	"flag"
	"fmt"
	"io"
)

// draftWorkflowSection is the shared workflow summary appended to
// per-command help. It describes schema-2 project verbs and the frozen-launch rule.
const draftWorkflowSection = `Skillfile schema 2 project sources:
  Schema 2 source declarations resolve with the project commands below.
  Skillfile schema 1 retains its exact meaning; no on-disk migration is implicit.

  curator project resolve [path]   freeze declared sources into Skillfile.lock.json
  curator install [path]           materialize the locked snapshot (repairs drift)
  curator status [path]            compare installed state against the lock (--check, --json)
  curator project refresh [path]   re-resolve refs, bytes, and membership explicitly
  curator install [path]           install the refreshed lock (refresh alone changes nothing live)

  Launch and status never rescan collections, advance branches, or
  replace a local snapshot: installed shims execute the frozen runtime
  until an explicit refresh plus install republishes it. A Skillfile
  edited after resolve fails source_lock_stale until refresh; run
  install again after refresh to repair drifted bytes from the lock.
`

// draftExamplesSection shows one acquisition shape per line. Each shape
// is a value of one "sources" alias or one "skills" selector; docs/cli.md
// carries the same shapes as complete manifests.
const draftExamplesSection = `Schema 2 source examples (one alias or selector value per line):
  {"path": "./agents"}              local directory, relative to Skillfile.json
  {"path": "/work/shared-agents"}   local directory, absolute
  {"git": "https://example.org/kit.git", "tag": "v1.2.0"}   Git source at a tag
  {"repository": "example.org/kit", "branch": "main"}       logical Git source (needs machine policy)
  {"name": "review", "from": "team", "directory": "skills/review"}   one skill
  {"from": "team", "directory": "skills", "include": ["review", "docs"]}   collection
  {"from": "team", "directory": "skills", "include": ["*"]}   whole directory
  {"from": "team", "directory": "skills", "include": ["*"], "exclude": ["release"]}   minus one

  Machine policy (operator-owned, beside the manager configuration):
  source-policy.json maps canonical host/path entries to one or two
  endpoints plus fallback, and admits root packages via "root_inputs".
  See "Skillfile schema 2 project sources" in docs/cli.md.
`

// projectResolveUsage is the explicit help for the two existing verbs
// that run the explicit schema-2 attempt. It describes frozen v1
// behavior first, then the default lane.
const projectResolveUsage = `curator project resolve: resolve transitive dependencies for a project closure

Usage:
  curator project resolve [path]
  curator project refresh [path]

Schema 1: prints the alias, project path, Skillfile path, and managed
skill and bin directories without modifying disk state.

` + draftWorkflowSection + `
` + draftExamplesSection

// appendDraftUsage wraps the flag defaults of one existing command with
// the schema-2 project workflow section.
func appendDraftUsage(flags *flag.FlagSet, out io.Writer, name string) {
	defaults := func() {
		_, _ = fmt.Fprintf(out, "Usage of %s:\n", name)
		flags.PrintDefaults()
	}
	flags.Usage = func() {
		defaults()
		_, _ = fmt.Fprint(out, "\n"+draftWorkflowSection)
	}
}
