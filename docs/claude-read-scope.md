# Claude read scope of declared skills and selected contexts

In auto mode Claude Code asks once per configuration directory before the
first Read, Grep or Glob outside the working directories. A Curator-managed
member's skills and context modules always sit outside its working
directory. `skills/<name>` in the managed home links into the contexts
store, and selected context modules live in
`contexts/context/<package>/<pin>/`. An unattended member therefore stalled
on that prompt the first time it read its own skill reference or context
module.

Every launch runs `curator env resolve <env> --repair`. For `claude_code`,
resolve now projects a narrow Read allow set into the managed home's user
settings (`CLAUDE_CONFIG_DIR/settings.json`):

- `Read(//<home>/skills/<name>/**)` for each declared skill, together with
  `Read(//<store root>/**)` for its resolved immutable projection;
- `Read(//<store root>/**)` for each context member of the lock.

Paths are absolute literals. Gitignore pattern characters are escaped, and
each path is also rendered through the canonical location of its deepest
existing ancestor. Claude admits a read through a symlink only when the
requested path and the path it resolves to both match an allow rule. A
sibling home's `skills/<name>`, which links to the same immutable entry, is
therefore still not admitted, and neither is a retargeted link. Nothing
else is granted: no `additionalDirectories`, no Edit rule, no permission
mode, no whole home or store. Deny and ask rules from every settings source
keep their precedence over these allow rules.

## Ownership and merge

Curator owns the allow entries of the projected shape, `Read(//<literal
path>/**)`, below the managed `skills/` directory or the contexts store.
Repair keeps every other key, value and number lexeme. It also keeps every
other allow entry in its order, including an operator glob such as
`Read(//<store>/**)`. It removes owned entries the current profile no longer
projects and appends the missing ones. A satisfied file is not rewritten.
Publication rereads the exact input (including absence) immediately before
the rename, like `.claude.json`. Crash recovery captures and restores the
file with the rest of the home.

A settings document Curator cannot merge without overriding operator
policy is left byte-identical. That covers a non-object document, a
non-object `permissions`, a non-array `permissions.allow`, a link or a
non-regular file. Resolve then reports a warning that the read scope is not
projected. A home without the projection is stale for verification, so the
next launch-time repair adds it to existing members as well as to new ones.

## Tests

```sh
go test ./internal/envprofile -count=1 -v -run 'TestClaudeReadScope'
```

`claude_read_scope_test.go` drives `Resolve` with temporary homes. It
models Claude's documented allow and symlink rules for the positive case:
`skills/<declared skill>/references/firewall.md`, `SKILL.md` and a selected
context module are admitted. Negative controls cover a sibling home's link
into the same skill entry, an unselected context package, an a2a-swarm
configuration path, the wider contexts store, the managed home outside
`skills/` and a retargeted skill link. The tests do not run Claude itself.
The real Read on Claude 2.1.312 is qualified separately on the host.
